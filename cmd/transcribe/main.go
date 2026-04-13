package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-go-golems/transcription-go/internal/asr"
	"github.com/go-go-golems/transcription-go/internal/convert"
	"github.com/go-go-golems/transcription-go/internal/output"
	"github.com/go-go-golems/transcription-go/internal/server"
	"github.com/spf13/cobra"
)

func main() {
	var (
		inputPath  string
		outputDir  string
		formats    string
		noFillers  bool
		chunkSize  int
		verbose    bool
		serverDir  string
	)

	rootCmd := &cobra.Command{
		Use:   "transcribe --input audio.wav --output-dir ./out/",
		Short: "Transcribe audio using NVIDIA Nemotron ASR (via Dagger)",
		Long: `Transcribe audio files using NVIDIA Nemotron Speech Streaming 0.6B.

Uses Dagger to run a Python ASR server in a container. The model is
cached across runs. Audio conversion is done in pure Go (no ffmpeg needed).

Output formats: srt, vtt, txt, db (SQLite). Default: srt,db`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if inputPath == "" {
				return fmt.Errorf("--input is required")
			}
			if _, err := os.Stat(inputPath); err != nil {
				return fmt.Errorf("input file not found: %s", inputPath)
			}
			return run(context.Background(), inputPath, outputDir, formats, noFillers, chunkSize, verbose, serverDir)
		},
	}

	rootCmd.Flags().StringVarP(&inputPath, "input", "i", "", "Input audio file (WAV)")
	rootCmd.Flags().StringVarP(&outputDir, "output-dir", "o", "./out", "Output directory")
	rootCmd.Flags().StringVarP(&formats, "format", "f", "srt,db", "Output formats: srt,vtt,txt,db (comma-separated)")
	rootCmd.Flags().BoolVar(&noFillers, "no-fillers", false, "Remove filler words (um, uh, etc.)")
	rootCmd.Flags().IntVar(&chunkSize, "chunk-size", 60, "Seconds per transcription chunk")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output (show Dagger logs)")
	rootCmd.Flags().StringVar(&serverDir, "server-dir", "server", "Path to Python server directory")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func run(ctx context.Context, inputPath, outputDir, formats string, noFillers bool, chunkSize int, verbose bool, serverDir string) error {
	// Resolve server dir relative to binary location for go run compatibility
	if !filepath.IsAbs(serverDir) {
		exePath, err := os.Executable()
		if err == nil {
			serverDir = filepath.Join(filepath.Dir(exePath), "..", "..", serverDir)
		}
		// If running via go run, the exe is in a temp dir. Fall back to cwd.
		if _, err := os.Stat(filepath.Join(serverDir, "server.py")); err != nil {
			cwd, _ := os.Getwd()
			serverDir = filepath.Join(cwd, "server")
		}
	}

	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	// Step 1: Convert audio to 16kHz mono WAV
	convertedPath := filepath.Join(outputDir, "audio_16k_mono.wav")
	log.Printf("Converting audio: %s → %s", inputPath, convertedPath)
	if err := convert.To16kMono(inputPath, convertedPath); err != nil {
		return fmt.Errorf("convert audio: %w", err)
	}
	log.Printf("Audio converted")

	// Step 2: Start ASR server
	log.Printf("Starting ASR server...")
	svc, err := server.Start(ctx, server.Options{
		ServerDir: serverDir,
		Port:      8000,
	})
	if err != nil {
		return fmt.Errorf("start server: %w", err)
	}
	defer svc.Stop()
	log.Printf("ASR server ready at %s", svc.Endpoint())

	// Step 3: Transcribe
	log.Printf("Transcribing: %s (chunk size: %ds)", convertedPath, chunkSize)
	client := asr.NewClient(svc.Endpoint())
	result, err := client.TranscribeFull(ctx, convertedPath, chunkSize)
	if err != nil {
		return fmt.Errorf("transcribe: %w", err)
	}
	log.Printf("Transcription complete: %d words, %.1fs", result.WordCount, result.TotalDuration)

	// Convert to output.Word slice
	words := make([]output.Word, len(result.Words))
	for i, w := range result.Words {
		words[i] = output.Word{Word: w.Word, Start: w.Start, End: w.End}
	}

	// Step 4: Write outputs
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
			output.WriteSRT(w, segments)
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
			output.WriteVTT(w, segments)
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
			output.WriteTXT(w, segments)
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
