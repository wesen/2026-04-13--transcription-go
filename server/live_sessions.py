from __future__ import annotations

import time
from dataclasses import dataclass, field
from typing import Callable

from live_decoder import DecodeResult, LiveDecoder, LiveDecoderError


class LiveSessionError(Exception):
    def __init__(self, code: str, message: str):
        super().__init__(message)
        self.code = code
        self.message = message


@dataclass
class LiveSession:
    session_id: str
    decoder: LiveDecoder
    source: str = "unknown"
    created_at: float = field(default_factory=time.time)
    last_activity_at: float = field(default_factory=time.time)
    expected_sequence: int = 0
    stopped: bool = False

    def touch(self) -> None:
        self.last_activity_at = time.time()

    def append_audio(self, *, sequence: int, pts: float, duration: float, pcm16_bytes: bytes) -> DecodeResult | None:
        if self.stopped:
            raise LiveSessionError("session-stopped", f"session {self.session_id} is already stopped")
        if sequence != self.expected_sequence:
            raise LiveSessionError(
                "invalid-sequence",
                f"expected sequence {self.expected_sequence}, got {sequence}",
            )
        if pts < 0:
            raise LiveSessionError("invalid-pts", f"pts must be >= 0, got {pts}")
        if duration < 0:
            raise LiveSessionError("invalid-duration", f"duration must be >= 0, got {duration}")
        try:
            self.decoder.append_audio(pcm16_bytes, pts=pts)
        except LiveDecoderError as exc:
            raise LiveSessionError("decoder-error", str(exc)) from exc
        self.expected_sequence += 1
        self.touch()
        try:
            return self.decoder.decode_partial()
        except LiveDecoderError as exc:
            raise LiveSessionError("decode-error", str(exc)) from exc

    def flush(self) -> DecodeResult:
        if self.stopped:
            raise LiveSessionError("session-stopped", f"session {self.session_id} is already stopped")
        self.touch()
        try:
            return self.decoder.flush()
        except LiveDecoderError as exc:
            raise LiveSessionError("decode-error", str(exc)) from exc

    def stop(self) -> DecodeResult:
        result = self.flush()
        self.stopped = True
        self.touch()
        return result

    def close(self) -> None:
        self.stopped = True
        self.touch()
        self.decoder.close()

    @property
    def word_count(self) -> int:
        return self.decoder.finalized_word_count

    @property
    def duration(self) -> float:
        return self.decoder.finalized_duration


class LiveSessionRegistry:
    def __init__(self, *, idle_timeout_seconds: float = 300.0):
        self.idle_timeout_seconds = idle_timeout_seconds
        self._sessions: dict[str, LiveSession] = {}

    def create_session(
        self,
        session_id: str,
        *,
        decoder_factory: Callable[[], LiveDecoder],
        source: str = "unknown",
    ) -> LiveSession:
        self.cleanup_expired()
        session_id = session_id.strip()
        if not session_id:
            raise LiveSessionError("invalid-session-id", "session_id is required")
        if session_id in self._sessions:
            raise LiveSessionError("session-already-exists", f"session {session_id} already exists")
        session = LiveSession(session_id=session_id, decoder=decoder_factory(), source=source)
        self._sessions[session_id] = session
        return session

    def get_session(self, session_id: str) -> LiveSession:
        self.cleanup_expired()
        try:
            session = self._sessions[session_id]
        except KeyError as exc:
            raise LiveSessionError("session-not-found", f"session {session_id} not found") from exc
        session.touch()
        return session

    def close_session(self, session_id: str, *, reason: str = "closed") -> None:
        session = self._sessions.pop(session_id, None)
        if session is None:
            return
        session.close()

    def close_all(self) -> None:
        for session_id in list(self._sessions.keys()):
            self.close_session(session_id, reason="shutdown")

    def cleanup_expired(self) -> list[str]:
        if self.idle_timeout_seconds <= 0:
            return []
        now = time.time()
        expired = [
            session_id
            for session_id, session in self._sessions.items()
            if now - session.last_activity_at > self.idle_timeout_seconds
        ]
        for session_id in expired:
            self.close_session(session_id, reason="idle-timeout")
        return expired
