package main

import (
	"context"

	"github.com/go-go-golems/transcription-go/internal/live"
	"github.com/go-go-golems/transcription-go/internal/server"
	"github.com/spf13/cobra"
)

type liveOptions struct {
	inputPath      string
	outputDir      string
	sessionID      string
	transport      string
	chunkDuration  float64
	overlapSeconds float64
	replaySpeed    float64
	formats        string
	verbose        bool
	serverDir      string
}

func defaultLiveOptions() liveOptions {
	return liveOptions{
		outputDir:     "./out-live",
		transport:     live.TransportWS,
		chunkDuration: 2.0,
		replaySpeed:   1.0,
		formats:       "console",
		serverDir:     "server",
	}
}

func addLiveFlags(cmd *cobra.Command, opts *liveOptions) {
	cmd.Flags().StringVarP(&opts.inputPath, "input", "i", opts.inputPath, "Input WAV file to replay as a simulated live source")
	cmd.Flags().StringVarP(&opts.outputDir, "output-dir", "o", opts.outputDir, "Output directory for live transcript artifacts")
	cmd.Flags().StringVar(&opts.sessionID, "session-id", opts.sessionID, "Stable session identifier for the live transcription run")
	cmd.Flags().StringVar(&opts.transport, "transport", opts.transport, "Live transport: ws (default) or chunk (debug/fallback)")
	cmd.Flags().Float64Var(&opts.chunkDuration, "chunk-duration", opts.chunkDuration, "Replay chunk duration in seconds")
	cmd.Flags().Float64Var(&opts.overlapSeconds, "overlap-seconds", opts.overlapSeconds, "Expected chunk overlap in seconds")
	cmd.Flags().Float64Var(&opts.replaySpeed, "replay-speed", opts.replaySpeed, "Replay speed multiplier (1.0 = real time, 2.0 = 2x faster, 0 = no pacing)")
	cmd.Flags().StringVar(&opts.formats, "live-format", opts.formats, "Live output formats: console,srt,vtt,txt,db (comma-separated)")
	cmd.Flags().BoolVarP(&opts.verbose, "verbose", "v", opts.verbose, "Verbose output (show Dagger logs)")
	cmd.Flags().StringVar(&opts.serverDir, "server-dir", opts.serverDir, "Path to Python server directory")
}

func runLive(ctx context.Context, opts liveOptions) error {
	resolvedServerDir, err := server.ResolveServerDir(opts.serverDir)
	if err != nil {
		return err
	}

	runner := live.NewRunner(live.RunnerConfig{
		ServerDir:      resolvedServerDir,
		InputPath:      opts.inputPath,
		OutputDir:      opts.outputDir,
		SessionID:      opts.sessionID,
		Transport:      opts.transport,
		ChunkDuration:  opts.chunkDuration,
		OverlapSeconds: opts.overlapSeconds,
		ReplaySpeed:    opts.replaySpeed,
		Formats:        live.ParseFormats(opts.formats),
		Verbose:        opts.verbose,
	})

	return runner.Run(ctx)
}
