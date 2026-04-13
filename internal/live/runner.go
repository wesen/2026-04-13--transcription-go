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

const (
	TransportChunk = "chunk"
	TransportWS    = "ws"
)

// RunnerConfig collects the stable wiring that both chunked near-live mode and
// the future streaming transport will need.
type RunnerConfig struct {
	ServerDir      string
	InputPath      string
	OutputDir      string
	SessionID      string
	Transport      string
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

	transport := strings.TrimSpace(r.Config.Transport)
	if transport == "" {
		transport = TransportWS
	}
	if transport != TransportChunk && transport != TransportWS {
		return fmt.Errorf("unsupported live transport %q", transport)
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

	source := NewReplaySource(ReplaySourceConfig{
		InputPath:      convertedPath,
		TempDir:        filepath.Join(tempDir, "chunks"),
		SessionID:      sessionID,
		ChunkDuration:  r.Config.ChunkDuration,
		OverlapSeconds: r.Config.OverlapSeconds,
		ReplaySpeed:    r.Config.ReplaySpeed,
	})

	switch transport {
	case TransportChunk:
		return r.runChunkTransport(ctx, sessionID, svc.Endpoint(), source, sinks)
	case TransportWS:
		return r.runWSTransport(ctx, sessionID, svc.Endpoint(), source, sinks)
	default:
		return fmt.Errorf("unsupported live transport %q", transport)
	}
}

func (r *LiveRunner) runChunkTransport(ctx context.Context, sessionID, endpoint string, source AudioSource, sinks []Sink) error {
	client := asr.NewClient(endpoint)
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

		before := len(r.Accumulator.CommittedWords())
		event := TranscriptEvent{
			Type:         TranscriptEventFinalWords,
			SessionID:    chunk.SessionID,
			Sequence:     chunk.Sequence,
			UpToTime:     resp.ChunkStart + resp.ChunkDuration,
			Words:        respWordsToOutput(resp.Words),
			ProcessingMS: resp.ProcessingMS,
		}
		if err := r.Accumulator.ApplyEvent(event); err != nil {
			return fmt.Errorf("accumulate chunk %d: %w", chunk.Sequence, err)
		}
		state := r.Accumulator.State()
		if err := updateSinksForEvent(sinks, state, event.Type); err != nil {
			return fmt.Errorf("update sink after chunk %d: %w", chunk.Sequence, err)
		}
		latency := time.Since(chunk.EmittedAt).Round(time.Millisecond)
		metrics.ObserveChunk(chunk, len(state.Committed), resp.ProcessingMS, latency)
		if err := WriteMetricsSummary(r.Config.OutputDir, metrics.Summary(sessionID)); err != nil {
			return fmt.Errorf("write metrics summary after chunk %d: %w", chunk.Sequence, err)
		}
		log.Printf(
			"Processed live chunk transport=%s seq=%d start=%.3fs duration=%.3fs server_words=%d committed_total=%d committed_added=%d processing_ms=%d end_to_end=%s",
			TransportChunk,
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
	return r.finishRun(sessionID, started, metrics, processedChunks)
}

func (r *LiveRunner) runWSTransport(ctx context.Context, sessionID, endpoint string, source AudioSource, sinks []Sink) error {
	client := NewWSLiveClient(endpoint)
	if err := client.Connect(ctx); err != nil {
		return fmt.Errorf("connect ws client: %w", err)
	}
	defer client.Close()
	if err := client.Start(sessionID); err != nil {
		return fmt.Errorf("start ws session: %w", err)
	}

	chunks := make(chan AudioChunk)
	sourceErrCh := make(chan error, 1)
	go func() { sourceErrCh <- source.Run(ctx, chunks) }()

	finalizedCh := make(chan AudioChunk, 32)
	finalizedAckCh := make(chan struct{}, 1)
	senderErrCh := make(chan error, 1)
	go func() {
		senderErrCh <- SendAudioFrames(ctx, client, chunks, finalizedCh, finalizedAckCh, StreamSenderConfig{FlushEveryChunk: true})
	}()

	messages := make(chan StreamMessage)
	receiverErrCh := make(chan error, 1)
	go func() { receiverErrCh <- ReceiveResultEvents(ctx, client, messages) }()

	started := time.Now()
	metrics := NewMetricsCollector(started)
	processedChunks := 0
	var finalizedQueue []AudioChunk

	for {
		select {
		case chunk, ok := <-finalizedCh:
			if !ok {
				finalizedCh = nil
				continue
			}
			finalizedQueue = append(finalizedQueue, chunk)
		case msg, ok := <-messages:
			if !ok {
				messages = nil
				continue
			}
			if msg.Event != nil {
				before := len(r.Accumulator.CommittedWords())
				var finalizedChunk AudioChunk
				if msg.Event.Type == TranscriptEventFinalWords && len(finalizedQueue) > 0 {
					finalizedChunk = finalizedQueue[0]
					msg.Event.Sequence = finalizedChunk.Sequence
				}
				if err := r.Accumulator.ApplyEvent(*msg.Event); err != nil {
					return fmt.Errorf("accumulate ws event %q: %w", msg.Event.Type, err)
				}
				state := r.Accumulator.State()
				if err := updateSinksForEvent(sinks, state, msg.Event.Type); err != nil {
					return fmt.Errorf("update sinks after ws event %q: %w", msg.Event.Type, err)
				}
				if msg.Event.Type == TranscriptEventFinalWords {
					processedChunks++
					if len(finalizedQueue) > 0 {
						chunk := finalizedQueue[0]
						finalizedQueue = finalizedQueue[1:]
						latency := time.Since(chunk.EmittedAt).Round(time.Millisecond)
						metrics.ObserveChunk(chunk, len(state.Committed), msg.Event.ProcessingMS, latency)
						if err := WriteMetricsSummary(r.Config.OutputDir, metrics.Summary(sessionID)); err != nil {
							return fmt.Errorf("write metrics summary after ws final event: %w", err)
						}
						log.Printf(
							"Processed live chunk transport=%s seq=%d start=%.3fs duration=%.3fs server_words=%d committed_total=%d committed_added=%d processing_ms=%d end_to_end=%s",
							TransportWS,
							chunk.Sequence,
							chunk.Start,
							chunk.Duration,
							len(msg.Event.Words),
							len(state.Committed),
							len(state.Committed)-before,
							msg.Event.ProcessingMS,
							latency,
						)
						select {
						case finalizedAckCh <- struct{}{}:
						default:
						}
					}
				}
			}
			if msg.Stopped != nil {
				if err := <-sourceErrCh; err != nil {
					return err
				}
				if err := <-senderErrCh; err != nil {
					return err
				}
				if err := <-receiverErrCh; err != nil {
					return err
				}
				return r.finishRun(sessionID, started, metrics, processedChunks)
			}
		case err := <-sourceErrCh:
			if err != nil {
				return err
			}
			sourceErrCh = nil
		case err := <-senderErrCh:
			if err != nil {
				return err
			}
			senderErrCh = nil
		case err := <-receiverErrCh:
			if err != nil {
				return err
			}
			receiverErrCh = nil
		case <-ctx.Done():
			return ctx.Err()
		}

		if messages == nil && receiverErrCh == nil {
			break
		}
	}
	return r.finishRun(sessionID, started, metrics, processedChunks)
}

func (r *LiveRunner) finishRun(sessionID string, started time.Time, metrics *MetricsCollector, processedChunks int) error {
	state := r.Accumulator.State()
	summary := metrics.Summary(sessionID)
	if err := WriteMetricsSummary(r.Config.OutputDir, summary); err != nil {
		return err
	}
	log.Printf(
		"Live replay complete: session=%s transport=%s chunks_processed=%d committed_words=%d output_dir=%s elapsed=%s avg_server_ms=%.1f avg_end_to_end_ms=%.1f audio_per_wall=%.2fx",
		sessionID,
		r.effectiveTransport(),
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

func (r *LiveRunner) effectiveTransport() string {
	transport := strings.TrimSpace(r.Config.Transport)
	if transport == "" {
		return TransportWS
	}
	return transport
}

func updateSinksForEvent(sinks []Sink, state TranscriptState, eventType TranscriptEventType) error {
	for _, sink := range sinks {
		if eventType == TranscriptEventPartial {
			if _, ok := sink.(*ConsoleSink); !ok {
				continue
			}
		}
		if err := sink.Update(state); err != nil {
			return err
		}
	}
	return nil
}

func respWordsToOutput(words []asr.Word) []output.Word {
	converted := make([]output.Word, len(words))
	for i, w := range words {
		converted[i] = output.Word{Word: w.Word, Start: w.Start, End: w.End}
	}
	return converted
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
