package live

import (
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"time"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
)

// ReplaySourceConfig configures WAV-backed simulated live replay.
type ReplaySourceConfig struct {
	InputPath      string
	TempDir        string
	SessionID      string
	ChunkDuration  float64
	OverlapSeconds float64
	ReplaySpeed    float64
}

// ReplaySource emits timed chunks from a prerecorded WAV file.
type ReplaySource struct {
	cfg ReplaySourceConfig
}

func NewReplaySource(cfg ReplaySourceConfig) *ReplaySource {
	return &ReplaySource{cfg: cfg}
}

func (s *ReplaySource) Run(ctx context.Context, out chan<- AudioChunk) error {
	defer close(out)

	if s.cfg.InputPath == "" {
		return fmt.Errorf("replay source input path is required")
	}
	if s.cfg.TempDir == "" {
		return fmt.Errorf("replay source temp dir is required")
	}
	if s.cfg.ChunkDuration <= 0 {
		return fmt.Errorf("chunk duration must be > 0")
	}
	if s.cfg.OverlapSeconds < 0 {
		return fmt.Errorf("overlap seconds must be >= 0")
	}

	f, err := os.Open(s.cfg.InputPath)
	if err != nil {
		return fmt.Errorf("open replay input: %w", err)
	}
	defer f.Close()

	dec := wav.NewDecoder(f)
	if !dec.IsValidFile() {
		return fmt.Errorf("not a valid WAV file: %s", s.cfg.InputPath)
	}
	if err := dec.FwdToPCM(); err != nil {
		return fmt.Errorf("seek to PCM data: %w", err)
	}

	sampleRate := int(dec.SampleRate)
	numChans := int(dec.NumChans)
	bitDepth := int(dec.BitDepth)
	if sampleRate <= 0 || numChans <= 0 || bitDepth <= 0 {
		return fmt.Errorf("invalid WAV format: rate=%d channels=%d bitDepth=%d", sampleRate, numChans, bitDepth)
	}

	chunkFrames := int(math.Round(s.cfg.ChunkDuration * float64(sampleRate)))
	overlapFrames := int(math.Round(s.cfg.OverlapSeconds * float64(sampleRate)))
	if overlapFrames >= chunkFrames {
		return fmt.Errorf("overlap %.3fs must be smaller than chunk duration %.3fs", s.cfg.OverlapSeconds, s.cfg.ChunkDuration)
	}
	chunkSamples := chunkFrames * numChans
	overlapSamples := overlapFrames * numChans
	stepFrames := chunkFrames - overlapFrames
	stepSamples := stepFrames * numChans

	current, err := readSamples(dec, chunkSamples, sampleRate, numChans)
	if err != nil {
		return err
	}
	if len(current) == 0 {
		return nil
	}

	sequence := 0
	startSeconds := 0.0
	for {
		durationSeconds := float64(len(current)/numChans) / float64(sampleRate)
		chunkPath, err := writeChunkWAV(s.cfg.TempDir, sequence, current, sampleRate, numChans, bitDepth)
		if err != nil {
			return err
		}
		chunk := AudioChunk{
			SessionID: s.cfg.SessionID,
			Sequence:  sequence,
			Start:     startSeconds,
			Duration:  durationSeconds,
			WAVPath:   chunkPath,
			EmittedAt: time.Now(),
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case out <- chunk:
		}

		newSamples, err := readSamples(dec, stepSamples, sampleRate, numChans)
		if err != nil {
			return err
		}
		if len(newSamples) == 0 {
			return nil
		}

		tail := current
		if overlapSamples > 0 {
			if len(tail) > overlapSamples {
				tail = tail[len(tail)-overlapSamples:]
			}
		} else {
			tail = nil
		}
		next := make([]int, 0, len(tail)+len(newSamples))
		next = append(next, tail...)
		next = append(next, newSamples...)
		current = next
		sequence++
		startSeconds += float64(stepFrames) / float64(sampleRate)

		if err := sleepForReplayStep(ctx, float64(stepFrames)/float64(sampleRate), s.cfg.ReplaySpeed); err != nil {
			return err
		}
	}
}

func readSamples(dec *wav.Decoder, wantSamples, sampleRate, numChans int) ([]int, error) {
	_ = sampleRate
	if wantSamples <= 0 {
		return nil, nil
	}
	out := make([]int, 0, wantSamples)
	for len(out) < wantSamples {
		remaining := wantSamples - len(out)
		buf := &audio.IntBuffer{
			Format: &audio.Format{NumChannels: numChans, SampleRate: sampleRate},
			Data:   make([]int, remaining),
		}
		n, err := dec.PCMBuffer(buf)
		if err != nil {
			return nil, fmt.Errorf("read replay PCM: %w", err)
		}
		if n == 0 {
			break
		}
		out = append(out, buf.Data[:n]...)
	}
	return out, nil
}

func writeChunkWAV(dir string, sequence int, samples []int, sampleRate, numChans, bitDepth int) (string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create replay temp dir: %w", err)
	}
	path := filepath.Join(dir, fmt.Sprintf("chunk-%06d.wav", sequence))
	f, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("create chunk wav: %w", err)
	}
	defer f.Close()

	enc := wav.NewEncoder(f, sampleRate, bitDepth, numChans, 1)
	buf := &audio.IntBuffer{
		Data:   append([]int(nil), samples...),
		Format: &audio.Format{NumChannels: numChans, SampleRate: sampleRate},
	}
	if err := enc.Write(buf); err != nil {
		return "", fmt.Errorf("write chunk wav: %w", err)
	}
	if err := enc.Close(); err != nil {
		return "", fmt.Errorf("close chunk wav encoder: %w", err)
	}
	return path, nil
}

func sleepForReplayStep(ctx context.Context, stepSeconds, replaySpeed float64) error {
	if replaySpeed <= 0 {
		return nil
	}
	wait := time.Duration((stepSeconds / replaySpeed) * float64(time.Second))
	if wait <= 0 {
		return nil
	}
	timer := time.NewTimer(wait)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
