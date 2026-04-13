"""
ASR transcription server using NVIDIA Nemotron Speech Streaming 0.6B.

Exposes a FastAPI server with:
  GET  /health          - health check
  POST /transcribe/full - transcribe a full audio file (server handles chunking)

The model is loaded once on startup and kept in memory.
"""

import os
import subprocess
import tempfile
from contextlib import asynccontextmanager

import soundfile as sf
from fastapi import FastAPI, UploadFile
from omegaconf import OmegaConf, open_dict

asr_model = None
time_stride = None


@asynccontextmanager
async def lifespan(app: FastAPI):
    global asr_model, time_stride

    print("Loading ASR model...")
    import nemo.collections.asr as nemo_asr

    asr_model = nemo_asr.models.ASRModel.from_pretrained(
        "nvidia/nemotron-speech-streaming-en-0.6b"
    )
    asr_model = asr_model.cpu().eval()

    decoding_cfg = asr_model.cfg.decoding
    with open_dict(decoding_cfg):
        decoding_cfg.preserve_alignments = True
        decoding_cfg.compute_timestamps = True
        decoding_cfg.segment_separators = []
        decoding_cfg.word_separator = " "
        asr_model.change_decoding_strategy(decoding_cfg)

    time_stride = 8 * asr_model.cfg.preprocessor.window_stride
    print(f"Model loaded. time_stride={time_stride}")

    yield


app = FastAPI(lifespan=lifespan)


@app.get("/health")
async def health():
    return {"status": "ok", "model_loaded": asr_model is not None}


@app.post("/transcribe/full")
async def transcribe_full(file: UploadFile, chunk_size: int = 60):
    """
    Transcribe a full audio file.

    The server handles chunking internally to avoid OOM.
    Returns word-level timestamps as JSON.

    Args:
        file: WAV audio file (any sample rate, channels).
        chunk_size: Seconds per processing chunk (default 60).
    """
    # Save uploaded audio to temp file
    with tempfile.NamedTemporaryFile(suffix=".wav", delete=False) as tmp:
        content = await file.read()
        tmp.write(content)
        tmp_path = tmp.name

    try:
        info = sf.info(tmp_path)
        duration = info.duration

        all_words = []
        for start in range(0, int(duration), chunk_size):
            end = min(start + chunk_size + 2, duration)
            chunk_tmp = tempfile.mktemp(suffix=".wav")
            subprocess.run(
                [
                    "ffmpeg", "-y", "-i", tmp_path,
                    "-ss", str(start), "-t", str(end - start),
                    "-ar", "16000", "-ac", "1", "-c:a", "pcm_s16le",
                    chunk_tmp,
                ],
                capture_output=True,
                check=True,
            )

            hypotheses = asr_model.transcribe(
                [chunk_tmp], return_hypotheses=True, batch_size=1
            )
            all_words.extend(_extract_words(hypotheses, start))
            os.remove(chunk_tmp)

        return {
            "words": all_words,
            "total_duration": round(duration, 3),
            "chunk_count": len(range(0, int(duration), chunk_size)),
            "word_count": len(all_words),
        }
    finally:
        os.unlink(tmp_path)


def _extract_words(hypotheses, chunk_start: float) -> list[dict]:
    """Extract word-level timestamps from ASR hypothesis."""
    words = []
    if not hypotheses or not hypotheses[0]:
        return words

    hyp = hypotheses[0]
    if not hasattr(hyp, "timestamp") or not hyp.timestamp:
        return words

    if "word" in hyp.timestamp:
        for stamp in hyp.timestamp["word"]:
            word = stamp.get("word", "") or stamp.get("char", "")
            word_start = stamp.get("start_offset", 0) * time_stride + chunk_start
            word_end = stamp.get("end_offset", 0) * time_stride + chunk_start
            if word.strip():
                words.append(
                    {
                        "word": word.strip(),
                        "start": round(word_start, 3),
                        "end": round(word_end, 3),
                    }
                )
    elif "char" in hyp.timestamp:
        # Group chars into words
        chars = []
        for stamp in hyp.timestamp["char"]:
            char = stamp.get("char", "")
            c_start = stamp.get("start_offset", 0) * time_stride + chunk_start
            c_end = stamp.get("end_offset", 0) * time_stride + chunk_start
            chars.append((char, c_start, c_end))

        current_word = []
        for char, c_start, c_end in chars:
            if char == " ":
                if current_word:
                    w = "".join([c[0] for c in current_word])
                    ws = current_word[0][1]
                    we = current_word[-1][2]
                    words.append({"word": w, "start": round(ws, 3), "end": round(we, 3)})
                    current_word = []
            else:
                current_word.append((char, c_start, c_end))

        if current_word:
            w = "".join([c[0] for c in current_word])
            ws = current_word[0][1]
            we = current_word[-1][2]
            words.append({"word": w, "start": round(ws, 3), "end": round(we, 3)})

    return words
