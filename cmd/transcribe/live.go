package main

import (
	"context"
	"fmt"

	"github.com/go-go-golems/transcription-go/internal/live"
	"github.com/go-go-golems/transcription-go/internal/server"
	"github.com/spf13/cobra"
)

type liveOptions struct {
	chunkDir       string
	sessionID      string
	overlapSeconds float64
	formats        string
	verbose        bool
	serverDir      string
}

func defaultLiveOptions() liveOptions {
	return liveOptions{
		formats:   "console",
		serverDir: "server",
	}
}

func addLiveFlags(cmd *cobra.Command, opts *liveOptions) {
	cmd.Flags().StringVar(&opts.chunkDir, "chunk-dir", opts.chunkDir, "Directory containing live chunk WAV files")
	cmd.Flags().StringVar(&opts.sessionID, "session-id", opts.sessionID, "Stable session identifier for the live transcription run")
	cmd.Flags().Float64Var(&opts.overlapSeconds, "overlap-seconds", opts.overlapSeconds, "Expected chunk overlap in seconds")
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
		ChunkDir:       opts.chunkDir,
		SessionID:      opts.sessionID,
		OverlapSeconds: opts.overlapSeconds,
		Formats:        live.ParseFormats(opts.formats),
		Verbose:        opts.verbose,
	})

	if err := runner.Run(ctx); err != nil {
		if err == live.ErrNotImplemented {
			return fmt.Errorf("live transcription is scaffolded but not implemented yet")
		}
		return err
	}
	return nil
}
