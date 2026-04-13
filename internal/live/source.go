package live

import (
	"context"
	"time"
)

// AudioChunk is the transport-neutral unit emitted by live audio sources.
type AudioChunk struct {
	SessionID string
	Sequence  int
	Start     float64
	Duration  float64
	WAVPath   string
	PCM16     []byte
	EmittedAt time.Time
}

// AudioSource produces ordered chunks for a live transcription session.
type AudioSource interface {
	Run(ctx context.Context, out chan<- AudioChunk) error
}
