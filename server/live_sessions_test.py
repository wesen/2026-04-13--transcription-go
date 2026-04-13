import unittest

from live_decoder import LiveDecoder
from live_sessions import LiveSessionError, LiveSessionRegistry


class LiveDecoderAndSessionTests(unittest.TestCase):
    def test_decoder_flush_tracks_pts_not_cumulative_overlap(self):
        offsets = []

        def transcribe(path, chunk_start):
            offsets.append(round(chunk_start, 3))
            return [{"word": "hello", "start": round(chunk_start, 3), "end": round(chunk_start + 0.5, 3)}]

        decoder = LiveDecoder(transcribe, min_partial_seconds=0.0)
        decoder.append_audio(b"\x00\x00" * 16000, pts=0.0)
        partial = decoder.decode_partial()
        self.assertIsNotNone(partial)
        self.assertEqual(partial.up_to_time, 1.0)
        final = decoder.flush()
        self.assertEqual(final.up_to_time, 1.0)
        self.assertEqual(decoder.finalized_duration, 1.0)
        self.assertEqual(decoder.finalized_word_count, 1)

        decoder.append_audio(b"\x00\x00" * 16000, pts=0.5)
        next_final = decoder.flush()
        self.assertEqual(next_final.words[0]["start"], 0.5)
        self.assertEqual(decoder.finalized_duration, 1.5)
        self.assertEqual(offsets, [0.0, 0.0, 0.5])

    def test_session_registry_rejects_out_of_order_sequence(self):
        registry = LiveSessionRegistry(idle_timeout_seconds=300)
        session = registry.create_session(
            "session-1",
            decoder_factory=lambda: LiveDecoder(lambda path, chunk_start: [], min_partial_seconds=0.0),
            source="test",
        )

        session.append_audio(sequence=0, pts=0.0, duration=1.0, pcm16_bytes=b"\x00\x00" * 16000)
        with self.assertRaises(LiveSessionError) as ctx:
            session.append_audio(sequence=2, pts=1.0, duration=1.0, pcm16_bytes=b"\x00\x00" * 16000)
        self.assertEqual(ctx.exception.code, "invalid-sequence")

        registry.close_session("session-1")
        with self.assertRaises(LiveSessionError) as ctx2:
            registry.get_session("session-1")
        self.assertEqual(ctx2.exception.code, "session-not-found")


if __name__ == "__main__":
    unittest.main()
