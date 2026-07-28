package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-go-golems/transcription-go/internal/asr"
	"github.com/go-go-golems/transcription-go/internal/convert"
	"github.com/go-go-golems/transcription-go/internal/output"
	"github.com/go-go-golems/transcription-go/internal/server"
	"github.com/spf13/cobra"
)

type batchOptions struct {
	inputPath string
	outputDir string
	formats   string
	noFillers bool
	chunkSize int
	verbose   bool
	serverDir string
}

func defaultBatchOptions() batchOptions {
	return batchOptions{
		outputDir: "./out",
		formats:   "srt,db",
		chunkSize: 60,
		serverDir: "server",
	}
}

func addBatchFlags(cmd *cobra.Command, opts *batchOptions) {
	cmd.Flags().StringVarP(&opts.inputPath, "input", "i", opts.inputPath, "Input audio file (WAV)")
	cmd.Flags().StringVarP(&opts.outputDir, "output-dir", "o", opts.outputDir, "Output directory")
	cmd.Flags().StringVarP(&opts.formats, "format", "f", opts.formats, "Output formats: srt,vtt,txt,db (comma-separated)")
	cmd.Flags().BoolVar(&opts.noFillers, "no-fillers", opts.noFillers, "Remove filler words (um, uh, etc.)")
	cmd.Flags().IntVar(&opts.chunkSize, "chunk-size", opts.chunkSize, "Seconds per transcription chunk")
	cmd.Flags().BoolVarP(&opts.verbose, "verbose", "v", opts.verbose, "Verbose output (show Dagger logs)")
	cmd.Flags().StringVar(&opts.serverDir, "server-dir", opts.serverDir, "Path to Python server directory")
}

func runBatch(ctx context.Context, opts batchOptions) error {
	if opts.inputPath == "" {
		return fmt.Errorf("--input is required")
	}
	if _, err := os.Stat(opts.inputPath); err != nil {
		return fmt.Errorf("input file not found: %s", opts.inputPath)
	}

	resolvedServerDir, err := server.ResolveServerDir(opts.serverDir)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(opts.outputDir, 0o755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	convertedPath := filepath.Join(opts.outputDir, "audio_16k_mono.wav")
	log.Printf("Converting audio: %s → %s", opts.inputPath, convertedPath)
	if err := convert.To16kMono(opts.inputPath, convertedPath); err != nil {
		return fmt.Errorf("convert audio: %w", err)
	}
	log.Printf("Audio converted")

	log.Printf("Starting ASR server...")
	svc, err := server.StartDefault(ctx, resolvedServerDir)
	if err != nil {
		return fmt.Errorf("start server: %w", err)
	}
	defer svc.Stop()
	log.Printf("ASR server ready at %s", svc.Endpoint())

	durationSec, estimatedChunks := estimateConvertedWAV(convertedPath, opts.chunkSize)
	if durationSec > 0 {
		log.Printf("Transcribing: %s (chunk size: %ds, duration: %.1fs, estimated chunks: %d)", convertedPath, opts.chunkSize, durationSec, estimatedChunks)
	} else {
		log.Printf("Transcribing: %s (chunk size: %ds)", convertedPath, opts.chunkSize)
	}

	client := asr.NewClient(svc.Endpoint())
	started := time.Now()
	stopHeartbeat := make(chan struct{})
	go logHeartbeat(started, estimatedChunks, stopHeartbeat)

	result, err := client.TranscribeFull(ctx, convertedPath, opts.chunkSize)
	close(stopHeartbeat)
	if err != nil {
		return fmt.Errorf("transcribe: %w", err)
	}
	elapsed := time.Since(started).Round(time.Second)
	log.Printf("Transcription complete: %d words, %.1fs audio, %d chunks, elapsed=%s", result.WordCount, result.TotalDuration, result.ChunkCount, elapsed)

	words := make([]output.Word, len(result.Words))
	for i, w := range result.Words {
		words[i] = output.Word{Word: w.Word, Start: w.Start, End: w.End}
	}

	return writeBatchOutputs(words, opts.outputDir, opts.formats, opts.noFillers)
}

func logHeartbeat(started time.Time, estimatedChunks int, stop <-chan struct{}) {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			elapsed := time.Since(started).Round(time.Second)
			if estimatedChunks > 0 {
				log.Printf("Still transcribing... elapsed=%s estimated_chunks=%d", elapsed, estimatedChunks)
			} else {
				log.Printf("Still transcribing... elapsed=%s", elapsed)
			}
		}
	}
}

func writeBatchOutputs(words []output.Word, outputDir, formats string, noFillers bool) error {
	formatList := strings.Split(formats, ",")
	for _, f := range formatList {
		f = strings.TrimSpace(f)
		switch f {
		case "srt":
			segWords := words
			if noFillers {
				segWords = output.FilterFillers(words)
			}
			segments := output.BuildSegments(segWords, 15.0, 120)
			path := filepath.Join(outputDir, "transcript.srt")
			w, err := os.Create(path)
			if err != nil {
				return err
			}
			if err := output.WriteSRT(w, segments); err != nil {
				w.Close()
				return fmt.Errorf("write srt: %w", err)
			}
			w.Close()
			log.Printf("Written: %s (%d segments)", path, len(segments))

		case "vtt":
			segWords := words
			if noFillers {
				segWords = output.FilterFillers(words)
			}
			segments := output.BuildSegments(segWords, 15.0, 120)
			path := filepath.Join(outputDir, "transcript.vtt")
			w, err := os.Create(path)
			if err != nil {
				return err
			}
			if err := output.WriteVTT(w, segments); err != nil {
				w.Close()
				return fmt.Errorf("write vtt: %w", err)
			}
			w.Close()
			log.Printf("Written: %s (%d segments)", path, len(segments))

		case "txt":
			segWords := words
			if noFillers {
				segWords = output.FilterFillers(words)
			}
			segments := output.BuildSegments(segWords, 15.0, 120)
			path := filepath.Join(outputDir, "transcript.txt")
			w, err := os.Create(path)
			if err != nil {
				return err
			}
			if err := output.WriteTXT(w, segments); err != nil {
				w.Close()
				return fmt.Errorf("write txt: %w", err)
			}
			w.Close()
			log.Printf("Written: %s", path)

		case "db":
			dbPath := filepath.Join(outputDir, "transcript.db")
			if err := output.WriteSQLite(words, dbPath); err != nil {
				return fmt.Errorf("write sqlite: %w", err)
			}
			log.Printf("Written: %s (%d words)", dbPath, len(words))

		default:
			log.Printf("Unknown format: %s (skipping)", f)
		}
	}

	log.Printf("Done!")
	return nil
}

func estimateConvertedWAV(path string, chunkSize int) (durationSec float64, chunkCount int) {
	st, err := os.Stat(path)
	if err != nil {
		return 0, 0
	}
	const wavHeaderBytes = 44
	const bytesPerSecond = 16000 * 2
	dataBytes := st.Size() - wavHeaderBytes
	if dataBytes <= 0 {
		return 0, 0
	}
	durationSec = float64(dataBytes) / float64(bytesPerSecond)
	if chunkSize > 0 {
		chunkCount = int(math.Ceil(durationSec / float64(chunkSize)))
	}
	return durationSec, chunkCount
}
