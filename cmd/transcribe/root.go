package main

import (
	"context"

	"github.com/spf13/cobra"
)

func newRootCmd() *cobra.Command {
	rootBatch := defaultBatchOptions()
	batchOpts := defaultBatchOptions()
	liveOpts := defaultLiveOptions()

	rootCmd := &cobra.Command{
		Use:   "transcribe",
		Short: "Transcribe audio using NVIDIA Nemotron ASR (via Dagger)",
		Long: `Transcribe audio using NVIDIA Nemotron Speech Streaming 0.6B.

Uses Dagger to run a Python ASR server in a container. The model is
cached across runs. Audio conversion is done in pure Go (no ffmpeg needed).

Use the default/root invocation or the explicit 'batch' subcommand for the
existing batch pipeline. Use 'live' for the work-in-progress live pipeline.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBatch(commandContext(cmd), rootBatch)
		},
	}
	addBatchFlags(rootCmd, &rootBatch)

	batchCmd := &cobra.Command{
		Use:   "batch",
		Short: "Run the batch transcription pipeline",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBatch(commandContext(cmd), batchOpts)
		},
	}
	addBatchFlags(batchCmd, &batchOpts)

	liveCmd := &cobra.Command{
		Use:   "live",
		Short: "Run the live transcription pipeline (work in progress)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLive(commandContext(cmd), liveOpts)
		},
	}
	addLiveFlags(liveCmd, &liveOpts)

	rootCmd.AddCommand(batchCmd, liveCmd)
	return rootCmd
}

func commandContext(cmd *cobra.Command) context.Context {
	if cmd == nil || cmd.Context() == nil {
		return context.Background()
	}
	return cmd.Context()
}
