"""
ASR transcription server using NVIDIA Nemotron Speech Streaming 0.6B.

Exposes a FastAPI server with:
  GET  /health             - health check
  POST /transcribe/full    - transcribe a full audio file (server handles chunking)
  POST /transcribe/chunk   - transcribe a single live chunk

The model is loaded once on startup and kept in memory.
"""

import logging
import os
import subprocess
import tempfile
import time
from contextlib import asynccontextmanager

import soundfile as sf
from fastapi import FastAPI, Form, UploadFile
from omegaconf import OmegaConf, open_dict

logger = logging.getLogger("transcription.server")

asr_model = None
time_stride = None


@asynccontextmanager
async def lifespan(app: FastAPI):
    global asr_model, time_stride

    logger.info("Loading ASR model...")
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
    logger.info("Model loaded. time_stride=%s", time_stride)

    yield


app = FastAPI(lifespan=lifespan)


@app.get("/health")
async def health():
    return {"status": "ok", "model_loaded": asr_model is not None}


@app.post("/transcribe/full")
async def transcribe_full(file: UploadFile, chunk_size: int = Form(60)):
    """
    Transcribe a full audio file.

    The server handles chunking internally to avoid OOM.
    Returns word-level timestamps as JSON.

    Args:
        file: WAV audio file (any sample rate, channels).
        chunk_size: Seconds per processing chunk (default 60).
    """
    with tempfile.NamedTemporaryFile(suffix=".wav", delete=False) as tmp:
        content = await file.read()
        tmp.write(content)
        tmp_path = tmp.name

    try:
        info = sf.info(tmp_path)
        duration = info.duration

        all_words = []
        started = time.perf_counter()
        chunk_count = len(range(0, int(duration), chunk_size))
        for chunk_index, start in enumerate(range(0, int(duration), chunk_size)):
            end = min(start + chunk_size + 2, duration)
            normalized_chunk_path = _normalize_wav(tmp_path, start=float(start), duration=end - start)
            try:
                words = _transcribe_chunk_file(normalized_chunk_path, float(start))
                all_words.extend(words)
                logger.info(
                    "full transcription chunk=%d/%d chunk_start=%.3f chunk_end=%.3f words=%d",
                    chunk_index + 1,
                    chunk_count,
                    float(start),
                    float(end),
                    len(words),
                )
            finally:
                os.remove(normalized_chunk_path)

        elapsed_ms = round((time.perf_counter() - started) * 1000)
        logger.info(
            "full transcription complete duration=%.3f chunk_count=%d word_count=%d processing_ms=%d",
            duration,
            chunk_count,
            len(all_words),
            elapsed_ms,
        )
        return {
            "words": all_words,
            "total_duration": round(duration, 3),
            "chunk_count": chunk_count,
            "word_count": len(all_words),
        }
    finally:
        os.unlink(tmp_path)


@app.post("/transcribe/chunk")
async def transcribe_chunk(
    file: UploadFile,
    session_id: str = Form(""),
    chunk_index: int = Form(0),
    chunk_start: float = Form(0.0),
    overlap_seconds: float = Form(0.0),
    is_final_chunk: bool = Form(False),
):
    """
    Transcribe a single chunk for near-live mode.

    The upload is normalized to 16kHz mono PCM16 before inference. The response
    preserves word-level timestamps in absolute timeline coordinates using
    `chunk_start` as the base offset.
    """
    with tempfile.NamedTemporaryFile(suffix=".wav", delete=False) as tmp:
        content = await file.read()
        tmp.write(content)
        tmp_path = tmp.name

    normalized_path = None
    started = time.perf_counter()
    try:
        normalized_path = _normalize_wav(tmp_path)
        info = sf.info(normalized_path)
        words = _transcribe_chunk_file(normalized_path, chunk_start)
        processing_ms = round((time.perf_counter() - started) * 1000)
        logger.info(
            "chunk transcription session_id=%s chunk_index=%d chunk_start=%.3f duration=%.3f overlap=%.3f final=%s words=%d processing_ms=%d",
            session_id or "-",
            chunk_index,
            chunk_start,
            info.duration,
            overlap_seconds,
            is_final_chunk,
            len(words),
            processing_ms,
        )
        return {
            "session_id": session_id,
            "chunk_index": chunk_index,
            "chunk_start": round(chunk_start, 3),
            "chunk_duration": round(info.duration, 3),
            "words": words,
            "partial": False,
            "processing_ms": processing_ms,
        }
    finally:
        os.unlink(tmp_path)
        if normalized_path and os.path.exists(normalized_path):
            os.unlink(normalized_path)


def _normalize_wav(input_path: str, start: float | None = None, duration: float | None = None) -> str:
    """Convert an audio file or segment into 16kHz mono PCM16 WAV."""
    with tempfile.NamedTemporaryFile(suffix=".wav", delete=False) as tmp:
        output_path = tmp.name

    cmd = ["ffmpeg", "-y", "-i", input_path]
    if start is not None:
        cmd.extend(["-ss", str(start)])
    if duration is not None:
        cmd.extend(["-t", str(duration)])
    cmd.extend([
        "-ar",
        "16000",
        "-ac",
        "1",
        "-c:a",
        "pcm_s16le",
        output_path,
    ])
    subprocess.run(cmd, capture_output=True, check=True)
    return output_path


def _transcribe_chunk_file(chunk_path: str, chunk_start: float) -> list[dict]:
    """Run ASR on a normalized chunk file and return word-level timestamps."""
    hypotheses = asr_model.transcribe([chunk_path], return_hypotheses=True, batch_size=1)
    return _extract_words(hypotheses, chunk_start)


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
