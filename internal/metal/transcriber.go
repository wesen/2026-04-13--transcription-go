package metal

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-go-golems/transcription-go/internal/corpus"
	"github.com/go-go-golems/transcription-go/internal/output"
)

// Engine selects which Metal-accelerated ASR backend to use.
type Engine string

const (
	EngineParakeet Engine = "parakeet" // NVIDIA Parakeet TDT 0.6B v3 — fastest, best English WER
	EngineWhisper  Engine = "whisper"  // OpenAI Whisper large-v3-turbo — widest language coverage
)

// Config configures the Metal-accelerated transcriber.
type Config struct {
	BinaryPath string // path to parakeet-cli or whisper-cli
	ModelPath  string // path to ggml model file
	Engine     Engine // parakeet or whisper
	Threads    int    // CPU threads (Metal GPU is always used)
}

// Transcriber runs ASR via whisper.cpp/parakeet-cli with Metal GPU acceleration.
// It shells out to the whisper.cpp binary, which uses Apple's Metal framework
// for GPU-accelerated inference on Apple Silicon.
type Transcriber struct {
	cfg Config
}

// NewTranscriber creates a Metal GPU transcriber.
func NewTranscriber(cfg Config) *Transcriber {
	if cfg.Threads <= 0 {
		cfg.Threads = 4
	}
	return &Transcriber{cfg: cfg}
}

// Transcribe runs the Metal-accelerated ASR binary and parses word-level output.
func (t *Transcriber) Transcribe(ctx context.Context, audioPath string, opts corpus.TranscribeOptions) (corpus.Transcription, error) {
	started := time.Now()

	// Run the CLI binary with JSON output for structured word timestamps.
	jsonPath, err := t.runCLI(ctx, audioPath)
	if err != nil {
		return corpus.Transcription{}, fmt.Errorf("run metal ASR: %w", err)
	}
	defer os.Remove(jsonPath)

	words, duration, err := t.parseOutput(jsonPath)
	if err != nil {
		return corpus.Transcription{}, fmt.Errorf("parse output: %w", err)
	}

	return corpus.Transcription{
		Words:           words,
		DurationSeconds: duration,
		ChunkCount:      1,
		ProcessingTime:  time.Since(started),
	}, nil
}

func (t *Transcriber) runCLI(ctx context.Context, audioPath string) (string, error) {
	tmpDir, err := os.MkdirTemp("", "metal-asr-*")
	if err != nil {
		return "", err
	}

	outputBase := filepath.Join(tmpDir, "transcript")
	jsonPath := outputBase + ".json"

	args := []string{
		"-m", t.cfg.ModelPath,
		"-f", audioPath,
		"-t", strconv.Itoa(t.cfg.Threads),
		"-of", outputBase,
		"-oj",   // output JSON
		"-np",   // no progress prints (parakeet only, whisper ignores)
	}

	cmd := exec.CommandContext(ctx, t.cfg.BinaryPath, args...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("cli execution: %w", err)
	}

	return jsonPath, nil
}

// whisperJSONOutput is the JSON format produced by whisper-cli -oj.
type whisperJSONOutput struct {
	Transcription struct {
		Segments []struct {
			Start float64 `json:"start"`
			End   float64 `json:"end"`
			Text  string  `json:"text"`
			Words []struct {
				Word string  `json:"word"`
				Start float64 `json:"start"`
				End   float64 `json:"end"`
			} `json:"words"`
		} `json:"segments"`
	} `json:"transcription"`
}

func (t *Transcriber) parseOutput(jsonPath string) ([]corpus.Word, float64, error) {
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return nil, 0, fmt.Errorf("read json output: %w", err)
	}

	var parsed whisperJSONOutput
	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil, 0, fmt.Errorf("unmarshal json: %w", err)
	}

	var words []corpus.Word
	var maxEnd float64

	for _, seg := range parsed.Transcription.Segments {
		if len(seg.Words) > 0 {
			// Word-level timestamps available
			for _, w := range seg.Words {
				text := strings.TrimSpace(w.Word)
				if text == "" {
					continue
				}
				ow := output.Word{Word: text, Start: w.Start, End: w.End}
				words = append(words, corpus.Word{
					Text:           text,
					NormalizedText: normalize(text),
					Start:          w.Start,
					End:            w.End,
					IsFiller:       ow.IsFiller(),
				})
				if w.End > maxEnd {
					maxEnd = w.End
				}
			}
		} else {
			// Fallback: segment-level only, split text into words at segment times
			segWords := strings.Fields(seg.Text)
			if len(segWords) == 0 {
				continue
			}
			segDuration := seg.End - seg.Start
			wordDur := segDuration / float64(len(segWords))
			for i, w := range segWords {
				wStart := seg.Start + float64(i)*wordDur
				wEnd := wStart + wordDur
				ow := output.Word{Word: w, Start: wStart, End: wEnd}
				words = append(words, corpus.Word{
					Text:           w,
					NormalizedText: normalize(w),
					Start:          wStart,
					End:            wEnd,
					IsFiller:       ow.IsFiller(),
				})
				if wEnd > maxEnd {
					maxEnd = wEnd
				}
			}
		}
	}

	if len(words) == 0 {
		return nil, 0, fmt.Errorf("no words found in output")
	}

	return words, maxEnd, nil
}

func normalize(s string) string {
	var sb strings.Builder
	for _, r := range s {
		switch r {
		case '.', ',', '?', '!', ';', ':', '"', '\'', '(', ')', '[', ']':
			continue
		}
		if r >= 'A' && r <= 'Z' {
			r += 32
		}
		sb.WriteRune(r)
	}
	return sb.String()
}
