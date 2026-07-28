package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/go-go-golems/transcription-go/internal/asr"
	"github.com/go-go-golems/transcription-go/internal/corpus"
	"github.com/go-go-golems/transcription-go/internal/server"
	"github.com/spf13/cobra"
)

type corpusOptions struct {
	manifestPath string
	databasePath string
	outputDir    string
	formats      string
	chunkSize    int
	serverDir    string
	sourceIDs    []string
	retryFailed  bool
	dryRun       bool
	failFast     bool
	verbose      bool
}

func defaultCorpusOptions() corpusOptions {
	return corpusOptions{
		formats:   "srt,vtt,txt",
		chunkSize: 60,
		serverDir: "server",
	}
}

func newCorpusCmd() *cobra.Command {
	opts := defaultCorpusOptions()

	cmd := &cobra.Command{
		Use:   "corpus",
		Short: "Run a resumable playlist video corpus transcription pipeline",
	}
	cmd.PersistentFlags().StringVar(&opts.manifestPath, "manifest", opts.manifestPath, "Path to normalized corpus manifest JSON")
	cmd.PersistentFlags().StringVar(&opts.databasePath, "database", opts.databasePath, "Path to corpus SQLite database")
	cmd.PersistentFlags().IntVar(&opts.chunkSize, "chunk-size", opts.chunkSize, "Seconds per ASR chunk")
	cmd.PersistentFlags().StringVar(&opts.serverDir, "server-dir", opts.serverDir, "Path to Python server directory")
	cmd.PersistentFlags().BoolVarP(&opts.verbose, "verbose", "v", opts.verbose, "Verbose output")

	runCmd := &cobra.Command{
		Use:   "run",
		Short: "Import manifest and transcribe pending videos",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCorpusRun(commandContext(cmd), opts)
		},
	}
	runCmd.Flags().StringVar(&opts.outputDir, "output-dir", opts.outputDir, "Directory for per-video exports")
	runCmd.Flags().StringVar(&opts.formats, "format", opts.formats, "Export formats: srt,vtt,txt (comma-separated)")
	runCmd.Flags().StringSliceVar(&opts.sourceIDs, "source-id", opts.sourceIDs, "Only process these source IDs (repeatable)")
	runCmd.Flags().BoolVar(&opts.retryFailed, "retry-failed", opts.retryFailed, "Retry failed videos")
	runCmd.Flags().BoolVar(&opts.dryRun, "dry-run", opts.dryRun, "Plan only; do not start the ASR service or transcribe")
	runCmd.Flags().BoolVar(&opts.failFast, "fail-fast", opts.failFast, "Stop on first failure")

	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Print corpus state counts",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCorpusStatus(commandContext(cmd), opts)
		},
	}

	searchCmd := &cobra.Command{
		Use:   "search",
		Short: "Search chunk text across the corpus",
		RunE: func(cmd *cobra.Command, args []string) error {
			query, _ := cmd.Flags().GetString("query")
			limit, _ := cmd.Flags().GetInt("limit")
			return runCorpusSearch(commandContext(cmd), opts, query, limit)
		},
	}
	searchCmd.Flags().StringP("query", "q", "", "Search query")
	searchCmd.Flags().IntP("limit", "n", 20, "Maximum results")
	_ = searchCmd.MarkFlagRequired("query")

	exportCmd := &cobra.Command{
		Use:   "export",
		Short: "Regenerate exports from committed transcripts without ASR",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCorpusExport(commandContext(cmd), opts)
		},
	}
	exportCmd.Flags().StringVar(&opts.outputDir, "output-dir", opts.outputDir, "Directory for per-video exports")
	exportCmd.Flags().StringVar(&opts.formats, "format", opts.formats, "Export formats: srt,vtt,txt (comma-separated)")
	exportCmd.Flags().StringSliceVar(&opts.sourceIDs, "source-id", opts.sourceIDs, "Only export these source IDs (repeatable)")

	cmd.AddCommand(runCmd, statusCmd, searchCmd, exportCmd)
	return cmd
}

func runCorpusRun(ctx context.Context, opts corpusOptions) error {
	manifest, err := corpus.LoadManifest(opts.manifestPath)
	if err != nil {
		return err
	}
	store, err := corpus.OpenStore(resolveCorpusDB(opts, opts.manifestPath))
	if err != nil {
		return err
	}
	defer store.Close()

	fingerprint := corpus.DefaultFingerprint(opts.chunkSize)
	policy := corpus.DefaultChunkPolicy()
	var exporter *corpus.Exporter
	if opts.outputDir != "" {
		exporter = corpus.NewExporter(store, opts.outputDir, policy)
	}

	var factory corpus.ServiceFactory
	var transcriberFactory func(endpoint string) corpus.Transcriber
	if !opts.dryRun {
		resolvedServerDir, err := server.ResolveServerDir(opts.serverDir)
		if err != nil {
			return err
		}
		factory = &daggerServiceFactory{serverDir: resolvedServerDir}
		transcriberFactory = asrTranscriberFactory(opts.chunkSize)
	}

	cfg := corpus.RunConfig{
		Manifest:           manifest,
		Store:              store,
		ServiceFactory:     factory,
		TranscriberFactory: transcriberFactory,
		Exporter:           exporter,
		Fingerprint:        fingerprint,
		Formats:            parseExportFormats(opts.formats),
		SourceIDs:          opts.sourceIDs,
		RetryFailed:        opts.retryFailed,
		DryRun:             opts.dryRun,
		FailFast:           opts.failFast,
		Verbose:            opts.verbose,
	}

	summary, err := corpus.Run(ctx, cfg)
	if err != nil {
		return err
	}
	if summary != nil {
		fmt.Println(summary.String())
	}
	return nil
}

func runCorpusStatus(ctx context.Context, opts corpusOptions) error {
	manifest, err := corpus.LoadManifest(opts.manifestPath)
	if err != nil {
		return err
	}
	store, err := corpus.OpenStore(resolveCorpusDB(opts, opts.manifestPath))
	if err != nil {
		return err
	}
	defer store.Close()
	counts, err := store.Status(ctx, manifest.Corpus.ID)
	if err != nil {
		return err
	}
	fmt.Printf("corpus=%s total=%d available=%d unavailable=%d pending=%d transcribing=%d complete=%d failed=%d stale=%d\n",
		manifest.Corpus.ID, counts.Total, counts.Available, counts.Unavailable,
		counts.Pending, counts.Transcribing, counts.Complete, counts.Failed, counts.Stale)
	return nil
}

func runCorpusSearch(ctx context.Context, opts corpusOptions, query string, limit int) error {
	manifest, err := corpus.LoadManifest(opts.manifestPath)
	if err != nil {
		return err
	}
	store, err := corpus.OpenStore(resolveCorpusDB(opts, opts.manifestPath))
	if err != nil {
		return err
	}
	defer store.Close()
	hits, err := store.Search(ctx, manifest.Corpus.ID, query, limit)
	if err != nil {
		return err
	}
	for _, h := range hits {
		fmt.Printf("%03d %s [%s] %.3f-%.3s %s\n    %s\n    %s\n",
			h.Video.PlaylistPosition, h.Video.SourceID, h.Video.Title,
			h.Chunk.Start, fmt.Sprintf("%.3f", h.Chunk.End), h.SourceURL, h.Chunk.Text, h.Video.MediaPath)
	}
	if len(hits) == 0 {
		fmt.Println("no results")
	}
	return nil
}

func runCorpusExport(ctx context.Context, opts corpusOptions) error {
	manifest, err := corpus.LoadManifest(opts.manifestPath)
	if err != nil {
		return err
	}
	store, err := corpus.OpenStore(resolveCorpusDB(opts, opts.manifestPath))
	if err != nil {
		return err
	}
	defer store.Close()
	policy := corpus.DefaultChunkPolicy()
	if opts.outputDir == "" {
		return fmt.Errorf("--output-dir is required for export")
	}
	exporter := corpus.NewExporter(store, opts.outputDir, policy)

	fingerprint := corpus.DefaultFingerprint(opts.chunkSize)
	work, err := store.Plan(ctx, manifest.Corpus.ID, fingerprint, true, opts.sourceIDs)
	if err != nil {
		return err
	}
	formats := parseExportFormats(opts.formats)
	var failures []error
	for _, item := range work {
		if !item.Video.ActiveRevisionID.Valid {
			continue
		}
		data, err := store.LoadRevision(ctx, item.Video.ActiveRevisionID.Int64)
		if err != nil {
			failures = append(failures, fmt.Errorf("%s: %w", item.Video.SourceID, err))
			continue
		}
		paths, err := exporter.Export(ctx, item.Video, data, formats)
		if err != nil {
			failures = append(failures, fmt.Errorf("%s: %w", item.Video.SourceID, err))
			continue
		}
		log.Printf("exported %s: %s", item.Video.SourceID, strings.Join(paths, ", "))
	}
	return joinErrors(failures)
}

// daggerServiceFactory adapts internal/server.StartDefault to corpus.ServiceFactory.
type daggerServiceFactory struct {
	serverDir string
}

func (d *daggerServiceFactory) Start(ctx context.Context) (corpus.Service, error) {
	svc, err := server.StartDefault(ctx, d.serverDir)
	if err != nil {
		return nil, err
	}
	return &daggerService{svc: svc}, nil
}

type daggerService struct {
	svc *server.ASRServer
}

func (d *daggerService) Endpoint() string { return d.svc.Endpoint() }
func (d *daggerService) Stop() error {
	d.svc.Stop()
	return nil
}

func resolveCorpusDB(opts corpusOptions, manifestPath string) string {
	if opts.databasePath != "" {
		return opts.databasePath
	}
	if manifestPath == "" {
		return "corpus.db"
	}
	// Default next to the manifest.
	dir := dirOf(manifestPath)
	return dir + "/corpus.db"
}

func parseExportFormats(s string) []corpus.ExportFormat {
	var out []corpus.ExportFormat
	for _, f := range strings.Split(s, ",") {
		f = strings.TrimSpace(f)
		if f == "" {
			continue
		}
		out = append(out, corpus.ExportFormat(f))
	}
	return out
}

func joinErrors(errs []error) error {
	if len(errs) == 0 {
		return nil
	}
	out := make([]string, 0, len(errs))
	for _, e := range errs {
		out = append(out, e.Error())
	}
	return fmt.Errorf("%d export failures: %s", len(errs), strings.Join(out, "; "))
}

func dirOf(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			if i == 0 {
				return "/"
			}
			return path[:i]
		}
	}
	return "."
}

// asrTranscriberFactory wires a real HTTP transcriber once a service is running.
func asrTranscriberFactory(chunkSize int) func(endpoint string) corpus.Transcriber {
	return func(endpoint string) corpus.Transcriber {
		return corpus.NewHTTPTranscriber(asr.NewClient(endpoint), chunkSize)
	}
}
