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
	"log"
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

	// MaxChunkSeconds is the maximum audio chunk length sent to the ASR binary
	// in a single invocation. Audio longer than this is split with ffmpeg,
	// transcribed in segments, and word timestamps are merged with offsets.
	// Parakeet's Metal encoder runs out of GPU memory on audio longer than
	// ~5000s, so we chunk at 3600s (1 hour) by default to stay safely below.
	MaxChunkSeconds int
}

// DefaultMaxChunkSeconds is the default chunk size for Metal ASR.
const DefaultMaxChunkSeconds = 3600

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
	if cfg.MaxChunkSeconds <= 0 {
		cfg.MaxChunkSeconds = DefaultMaxChunkSeconds
	}
	return &Transcriber{cfg: cfg}
}

// Transcribe runs the Metal-accelerated ASR binary and parses word-level output.
// For audio longer than MaxChunkSeconds, it splits the file with ffmpeg,
// transcribes each chunk, and merges word timestamps with time offsets.
func (t *Transcriber) Transcribe(ctx context.Context, audioPath string, opts corpus.TranscribeOptions) (corpus.Transcription, error) {
	started := time.Now()

	// Get audio duration to decide whether chunking is needed.
	duration, err := getAudioDuration(audioPath)
	if err != nil {
		return corpus.Transcription{}, fmt.Errorf("get audio duration: %w", err)
	}

	var allWords []corpus.Word
	var totalDuration float64
	chunkCount := 1

	if duration <= float64(t.cfg.MaxChunkSeconds) {
		// Single-shot transcription — no chunking needed.
		words, dur, err := t.transcribeFile(ctx, audioPath)
		if err != nil {
			return corpus.Transcription{}, fmt.Errorf("run metal ASR: %w", err)
		}
		allWords = words
		totalDuration = dur
	} else {
		// Chunk the audio with ffmpeg, transcribe each chunk, merge.
		chunks, cleanup, err := t.splitAudio(ctx, audioPath, duration)
		if err != nil {
			return corpus.Transcription{}, fmt.Errorf("split audio: %w", err)
		}
		defer cleanup()

		chunkCount = len(chunks)
		for i, chunk := range chunks {
			if err := ctx.Err(); err != nil {
				return corpus.Transcription{}, err
			}
			log.Printf("[metal] Transcribing chunk %d/%d (offset %.1fs, %.1fs)",
				i+1, chunkCount, chunk.offset, chunk.duration)

			words, _, err := t.transcribeFile(ctx, chunk.path)
			if err != nil {
				return corpus.Transcription{}, fmt.Errorf("run metal ASR on chunk %d: %w", i+1, err)
			}

			// Offset word timestamps by the chunk's start time.
			for _, w := range words {
				w.Start += chunk.offset
				w.End += chunk.offset
				w.SourceChunkIndex = i
				allWords = append(allWords, w)
			}
		}
		totalDuration = duration
	}

	if len(allWords) == 0 {
		return corpus.Transcription{}, fmt.Errorf("no words found in transcription")
	}

	return corpus.Transcription{
		Words:           allWords,
		DurationSeconds: totalDuration,
		ChunkCount:      chunkCount,
		ProcessingTime:  time.Since(started),
	}, nil
}

// audioChunk represents a split audio segment.
type audioChunk struct {
	path     string
	offset   float64 // start time in seconds within the original audio
	duration float64 // chunk duration in seconds
}

// splitAudio uses ffmpeg to split the audio file into chunks of MaxChunkSeconds.
// Returns the chunk file paths and a cleanup function.
func (t *Transcriber) splitAudio(ctx context.Context, audioPath string, duration float64) ([]audioChunk, func(), error) {
	tmpDir, err := os.MkdirTemp("", "metal-chunks-*")
	if err != nil {
		return nil, nil, err
	}
	cleanup := func() { os.RemoveAll(tmpDir) }

	maxChunk := float64(t.cfg.MaxChunkSeconds)
	var chunks []audioChunk

	numChunks := int(duration/maxChunk) + 1
	for i := 0; i < numChunks; i++ {
		offset := float64(i) * maxChunk
		if offset >= duration {
			break
		}
		chunkDur := maxChunk
		if offset+chunkDur > duration {
			chunkDur = duration - offset
		}

		chunkPath := filepath.Join(tmpDir, fmt.Sprintf("chunk_%06d.wav", i))
		// Use ffmpeg to extract the segment. -ar 16000 -ac 1 ensures 16kHz mono WAV.
		cmd := exec.CommandContext(ctx, "ffmpeg",
			"-ss", fmt.Sprintf("%.3f", offset),
			"-i", audioPath,
			"-t", fmt.Sprintf("%.3f", chunkDur),
			"-ar", "16000",
			"-ac", "1",
			"-y",
			chunkPath,
		)
		cmd.Stdout = nil
		cmd.Stderr = nil
		if err := cmd.Run(); err != nil {
			cleanup()
			return nil, nil, fmt.Errorf("ffmpeg chunk %d: %w", i, err)
		}

		chunks = append(chunks, audioChunk{
			path:     chunkPath,
			offset:   offset,
			duration: chunkDur,
		})
	}

	return chunks, cleanup, nil
}

// transcribeFile runs the CLI binary on a single audio file and parses the JSON output.
func (t *Transcriber) transcribeFile(ctx context.Context, audioPath string) ([]corpus.Word, float64, error) {
	jsonPath, err := t.runCLI(ctx, audioPath)
	if err != nil {
		return nil, 0, err
	}
	defer os.Remove(jsonPath)

	return t.parseOutput(jsonPath)
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
		"-oj", // output JSON
		"-np", // no progress prints (parakeet only, whisper ignores)
	}

	cmd := exec.CommandContext(ctx, t.cfg.BinaryPath, args...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("cli execution: %w", err)
	}

	return jsonPath, nil
}

// getAudioDuration returns the duration of an audio file in seconds using ffprobe.
func getAudioDuration(audioPath string) (float64, error) {
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		audioPath,
	)
	out, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("ffprobe: %w", err)
	}
	var dur float64
	if _, err := fmt.Sscanf(strings.TrimSpace(string(out)), "%f", &dur); err != nil {
		return 0, fmt.Errorf("parse duration: %w", err)
	}
	return dur, nil
}

// whisperJSONOutput is the JSON format produced by parakeet-cli -oj and whisper-cli -oj.
type whisperJSONOutput struct {
	Transcription struct {
		Segments []struct {
			Start  float64 `json:"start"`
			End    float64 `json:"end"`
			Text   string  `json:"text"`
			Words  []struct {
				Word  string  `json:"word"`
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
			// Fallback: segment-level only, split text into words at segment times.
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
