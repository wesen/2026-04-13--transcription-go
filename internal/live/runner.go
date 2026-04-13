package live

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-go-golems/transcription-go/internal/asr"
	"github.com/go-go-golems/transcription-go/internal/convert"
	"github.com/go-go-golems/transcription-go/internal/output"
	"github.com/go-go-golems/transcription-go/internal/server"
)

// RunnerConfig collects the stable wiring that both chunked near-live mode and
// the future streaming transport will need.
type RunnerConfig struct {
	ServerDir      string
	InputPath      string
	OutputDir      string
	SessionID      string
	ChunkDuration  float64
	OverlapSeconds float64
	ReplaySpeed    float64
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
	if r.Config.InputPath == "" {
		return fmt.Errorf("live replay input is required")
	}
	if _, err := os.Stat(r.Config.InputPath); err != nil {
		return fmt.Errorf("live replay input not found: %s", r.Config.InputPath)
	}
	if r.Config.ChunkDuration <= 0 {
		return fmt.Errorf("chunk duration must be > 0")
	}
	if err := ensureOutputDir(r.Config.OutputDir); err != nil {
		return fmt.Errorf("create live output dir: %w", err)
	}

	sessionID := r.Config.SessionID
	if strings.TrimSpace(sessionID) == "" {
		sessionID = fmt.Sprintf("live-%s", time.Now().Format("20060102-150405"))
	}

	tempDir, err := os.MkdirTemp("", "transcription-live-*")
	if err != nil {
		return fmt.Errorf("create live temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	convertedPath := filepath.Join(tempDir, "input_16k_mono.wav")
	log.Printf("Preparing replay input: %s -> %s", r.Config.InputPath, convertedPath)
	if err := convert.To16kMono(r.Config.InputPath, convertedPath); err != nil {
		return fmt.Errorf("convert replay input: %w", err)
	}

	log.Printf("Starting ASR server for live replay...")
	svc, err := server.StartDefault(ctx, r.Config.ServerDir)
	if err != nil {
		return fmt.Errorf("start server: %w", err)
	}
	defer svc.Stop()
	log.Printf("ASR server ready at %s", svc.Endpoint())

	sinks, err := buildSinks(r.Config.OutputDir, r.Config.Formats)
	if err != nil {
		return err
	}

	client := asr.NewClient(svc.Endpoint())
	source := NewReplaySource(ReplaySourceConfig{
		InputPath:      convertedPath,
		TempDir:        filepath.Join(tempDir, "chunks"),
		SessionID:      sessionID,
		ChunkDuration:  r.Config.ChunkDuration,
		OverlapSeconds: r.Config.OverlapSeconds,
		ReplaySpeed:    r.Config.ReplaySpeed,
	})

	chunks := make(chan AudioChunk)
	sourceErrCh := make(chan error, 1)
	go func() { sourceErrCh <- source.Run(ctx, chunks) }()

	started := time.Now()
	metrics := NewMetricsCollector(started)
	processedChunks := 0
	for chunk := range chunks {
		processedChunks++
		resp, err := client.TranscribeChunk(ctx, chunk.WAVPath, asr.ChunkRequest{
			SessionID:      chunk.SessionID,
			ChunkIndex:     chunk.Sequence,
			ChunkStart:     chunk.Start,
			OverlapSeconds: r.Config.OverlapSeconds,
			IsFinalChunk:   false,
		})
		_ = os.Remove(chunk.WAVPath)
		if err != nil {
			return fmt.Errorf("transcribe live chunk %d: %w", chunk.Sequence, err)
		}

		words := make([]output.Word, len(resp.Words))
		for i, w := range resp.Words {
			words[i] = output.Word{Word: w.Word, Start: w.Start, End: w.End}
		}
		before := len(r.Accumulator.CommittedWords())
		if err := r.Accumulator.ApplyEvent(TranscriptEvent{
			Type:      TranscriptEventFinalWords,
			SessionID: chunk.SessionID,
			Sequence:  chunk.Sequence,
			UpToTime:  maxWordEnd(words),
			Words:     words,
		}); err != nil {
			return fmt.Errorf("accumulate chunk %d: %w", chunk.Sequence, err)
		}
		state := r.Accumulator.State()
		for _, sink := range sinks {
			if err := sink.Update(state); err != nil {
				return fmt.Errorf("update sink after chunk %d: %w", chunk.Sequence, err)
			}
		}
		latency := time.Since(chunk.EmittedAt).Round(time.Millisecond)
		metrics.ObserveChunk(chunk, len(state.Committed), resp.ProcessingMS, latency)
		if err := WriteMetricsSummary(r.Config.OutputDir, metrics.Summary(sessionID)); err != nil {
			return fmt.Errorf("write metrics summary after chunk %d: %w", chunk.Sequence, err)
		}
		log.Printf(
			"Processed live chunk seq=%d start=%.3fs duration=%.3fs server_words=%d committed_total=%d committed_added=%d processing_ms=%d end_to_end=%s",
			chunk.Sequence,
			chunk.Start,
			chunk.Duration,
			len(resp.Words),
			len(state.Committed),
			len(state.Committed)-before,
			resp.ProcessingMS,
			latency,
		)
	}

	if err := <-sourceErrCh; err != nil {
		return err
	}

	state := r.Accumulator.State()
	summary := metrics.Summary(sessionID)
	if err := WriteMetricsSummary(r.Config.OutputDir, summary); err != nil {
		return err
	}
	log.Printf(
		"Live replay complete: session=%s chunks_processed=%d committed_words=%d output_dir=%s elapsed=%s avg_server_ms=%.1f avg_end_to_end_ms=%.1f audio_per_wall=%.2fx",
		sessionID,
		processedChunks,
		len(state.Committed),
		r.Config.OutputDir,
		time.Since(started).Round(time.Second),
		summary.AverageServerProcessingMS,
		summary.AverageEndToEndMS,
		summary.AudioSecondsPerWallSecond,
	)
	return nil
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
