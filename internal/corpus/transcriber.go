package corpus

import (
	"context"
	"time"

	"github.com/go-go-golems/transcription-go/internal/asr"
	"github.com/go-go-golems/transcription-go/internal/output"
)

// Transcriber transcribes one audio file into timed words.
type Transcriber interface {
	Transcribe(ctx context.Context, audioPath string, opts TranscribeOptions) (Transcription, error)
}

// TranscribeOptions configures one transcription call.
type TranscribeOptions struct {
	ChunkSizeSeconds int
}

// Service is a warm ASR service lifecycle.
type Service interface {
	Endpoint() string
	Stop() error
}

// NoService is a no-op service for backends that don't need a running service (e.g. Metal GPU).
type NoService struct{}

func (NoService) Endpoint() string { return "" }
func (NoService) Stop() error     { return nil }

// ServiceFactory starts one warm service.
type ServiceFactory interface {
	Start(ctx context.Context) (Service, error)
}

// HTTPTranscriber adapts the existing asr.Client to the corpus Transcriber.
type HTTPTranscriber struct {
	client    *asr.Client
	chunkSize int
}

// NewHTTPTranscriber wraps an ASR client.
func NewHTTPTranscriber(client *asr.Client, chunkSize int) *HTTPTranscriber {
	if chunkSize <= 0 {
		chunkSize = 60
	}
	return &HTTPTranscriber{client: client, chunkSize: chunkSize}
}

// Transcribe sends a full audio file and converts the response.
func (h *HTTPTranscriber) Transcribe(ctx context.Context, audioPath string, opts TranscribeOptions) (Transcription, error) {
	chunk := opts.ChunkSizeSeconds
	if chunk <= 0 {
		chunk = h.chunkSize
	}
	started := time.Now()
	resp, err := h.client.TranscribeFull(ctx, audioPath, chunk)
	if err != nil {
		return Transcription{}, err
	}
	words := make([]Word, len(resp.Words))
	for i, w := range resp.Words {
		ow := output.Word{Word: w.Word, Start: w.Start, End: w.End}
		words[i] = Word{
			Text:           w.Word,
			NormalizedText: normalize(w.Word),
			Start:          w.Start,
			End:            w.End,
			IsFiller:       ow.IsFiller(),
		}
	}
	return Transcription{
		Words:           words,
		DurationSeconds: resp.TotalDuration,
		ChunkCount:      resp.ChunkCount,
		ProcessingTime:  time.Since(started),
	}, nil
}

func normalize(s string) string {
	out := make([]byte, 0, len(s))
	for _, r := range s {
		switch r {
		case '.', ',', '?', '!', ';', ':', '"', '\'', '(', ')', '[', ']':
			continue
		}
		out = append(out, []byte(string(toLower(r)))...)
	}
	return string(out)
}

func toLower(r rune) rune {
	if r >= 'A' && r <= 'Z' {
		return r + 32
	}
	return r
}
