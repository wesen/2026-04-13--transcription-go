package live

import (
	"context"
	"errors"
	"strings"
)

var ErrNotImplemented = errors.New("live runner not implemented")

// RunnerConfig collects the stable wiring that both chunked near-live mode and
// the future streaming transport will need.
type RunnerConfig struct {
	ServerDir      string
	ChunkDir       string
	SessionID      string
	OverlapSeconds float64
	Formats        []string
	Verbose        bool
}

// LiveRunner owns the top-level orchestration for live transcription.
type LiveRunner struct {
	Config      RunnerConfig
	Accumulator TranscriptAccumulator
}

func NewRunner(cfg RunnerConfig) *LiveRunner {
	return &LiveRunner{
		Config:      cfg,
		Accumulator: NewAccumulator(),
	}
}

func (r *LiveRunner) Run(ctx context.Context) error {
	_ = ctx
	return ErrNotImplemented
}

func ParseFormats(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		out = append(out, part)
	}
	return out
}
