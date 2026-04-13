from __future__ import annotations

import tempfile
import time
import wave
from dataclasses import dataclass
from typing import Callable


class LiveDecoderError(Exception):
    pass


@dataclass
class DecodeResult:
    words: list[dict]
    up_to_time: float
    chunk_duration: float
    processing_ms: int


class LiveDecoder:
    """Buffered per-session decoder used by the initial WebSocket transport.

    This is intentionally simple: audio frames are buffered as raw PCM16, the
    current buffered region can be decoded as a partial preview, and `flush()`
    finalizes the currently buffered audio span. The buffered span is anchored to
    the incoming audio PTS so replay/window overlap does not accumulate into
    artificial timestamp drift.
    """

    def __init__(
        self,
        transcribe_file: Callable[[str, float], list[dict]],
        *,
        sample_rate: int = 16000,
        channels: int = 1,
        sample_format: str = "pcm_s16le",
        min_partial_seconds: float = 1.0,
    ):
        if sample_rate <= 0:
            raise LiveDecoderError("sample_rate must be > 0")
        if channels <= 0:
            raise LiveDecoderError("channels must be > 0")
        if sample_format != "pcm_s16le":
            raise LiveDecoderError(f"unsupported sample format {sample_format!r}")

        self._transcribe_file = transcribe_file
        self.sample_rate = sample_rate
        self.channels = channels
        self.sample_format = sample_format
        self.min_partial_seconds = min_partial_seconds
        self._bytes_per_sample = 2
        self._buffer = bytearray()
        self._buffer_start_pts: float | None = None
        self._finalized_duration = 0.0
        self._finalized_word_count = 0
        self._closed = False

    @property
    def finalized_word_count(self) -> int:
        return self._finalized_word_count

    @property
    def finalized_duration(self) -> float:
        return round(self._finalized_duration, 3)

    @property
    def buffered_duration(self) -> float:
        frame_bytes = self.sample_rate * self.channels * self._bytes_per_sample
        if frame_bytes == 0:
            return 0.0
        return len(self._buffer) / frame_bytes

    def append_audio(self, pcm16_bytes: bytes, *, pts: float) -> None:
        if self._closed:
            raise LiveDecoderError("decoder is closed")
        if not pcm16_bytes:
            return
        if self._buffer_start_pts is None:
            self._buffer_start_pts = pts
        self._buffer.extend(pcm16_bytes)

    def decode_partial(self) -> DecodeResult | None:
        if self.buffered_duration < self.min_partial_seconds:
            return None
        return self._decode(finalize=False)

    def flush(self) -> DecodeResult:
        return self._decode(finalize=True)

    def close(self) -> None:
        self._buffer.clear()
        self._buffer_start_pts = None
        self._closed = True

    def _decode(self, *, finalize: bool) -> DecodeResult:
        if self._closed:
            raise LiveDecoderError("decoder is closed")

        duration = round(self.buffered_duration, 3)
        start_pts = self._buffer_start_pts if self._buffer_start_pts is not None else self._finalized_duration
        if not self._buffer:
            return DecodeResult(words=[], up_to_time=round(start_pts, 3), chunk_duration=0.0, processing_ms=0)

        started = time.perf_counter()
        path = self._write_buffer_to_wav(bytes(self._buffer))
        try:
            words = self._transcribe_file(path, start_pts)
        finally:
            try:
                import os

                os.unlink(path)
            except OSError:
                pass

        processing_ms = round((time.perf_counter() - started) * 1000)
        up_to_time = round(start_pts + duration, 3)

        if finalize:
            self._finalized_duration = max(self._finalized_duration, up_to_time)
            self._finalized_word_count += len(words)
            self._buffer.clear()
            self._buffer_start_pts = None

        return DecodeResult(
            words=words,
            up_to_time=up_to_time,
            chunk_duration=duration,
            processing_ms=processing_ms,
        )

    def _write_buffer_to_wav(self, pcm16_bytes: bytes) -> str:
        with tempfile.NamedTemporaryFile(suffix=".wav", delete=False) as tmp:
            path = tmp.name

        with wave.open(path, "wb") as wf:
            wf.setnchannels(self.channels)
            wf.setsampwidth(self._bytes_per_sample)
            wf.setframerate(self.sample_rate)
            wf.writeframes(pcm16_bytes)

        return path
