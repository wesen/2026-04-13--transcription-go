package corpus

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
)

// RunConfig configures one corpus run.
type RunConfig struct {
	Manifest           *Manifest
	Store              *Store
	ServiceFactory     ServiceFactory
	TranscriberFactory func(endpoint string) Transcriber // nil when DryRun is true
	Exporter           *Exporter
	Fingerprint        Fingerprint
	Formats            []ExportFormat
	SourceIDs          []string
	RetryFailed        bool
	DryRun             bool
	FailFast           bool
	Verbose            bool
}

// PlanSummary is the dry-run planning report.
type PlanSummary struct {
	CorpusKey      string
	Total          int
	Unavailable    int
	Pending        int
	Stale          int
	Failed         int
	Unchanged      int
	ExportOnly     int
	WillTranscribe int
	WillStart      bool
}

// Run executes the corpus pipeline. In dry-run mode it returns a PlanSummary
// as error value nil and does not start the service.
func Run(ctx context.Context, cfg RunConfig) (*PlanSummary, error) {
	if cfg.Manifest == nil {
		return nil, fmt.Errorf("manifest is required")
	}
	if cfg.Store == nil {
		return nil, fmt.Errorf("store is required")
	}

	if _, err := cfg.Store.ImportManifest(ctx, cfg.Manifest); err != nil {
		return nil, fmt.Errorf("import manifest: %w", err)
	}
	if _, err := cfg.Store.ReconcileAbandoned(ctx); err != nil {
		return nil, fmt.Errorf("reconcile abandoned: %w", err)
	}

	work, err := cfg.Store.Plan(ctx, cfg.Manifest.Corpus.ID, cfg.Fingerprint, cfg.RetryFailed, cfg.SourceIDs)
	if err != nil {
		return nil, fmt.Errorf("plan: %w", err)
	}

	summary := summarize(cfg.Manifest.Corpus.ID, work)
	summary.WillStart = !cfg.DryRun && summary.WillTranscribe > 0

	if cfg.Verbose || cfg.DryRun {
		log.Print(summary.String())
	}
	if cfg.DryRun {
		return summary, nil
	}

	if summary.WillTranscribe == 0 {
		// Only export-only work remains (or nothing). No service needed.
		if err := runExportOnly(ctx, cfg, work); err != nil {
			return summary, err
		}
		return summary, nil
	}

	if cfg.ServiceFactory == nil && cfg.TranscriberFactory == nil {
		return summary, fmt.Errorf("either a service factory or a transcriber factory is required for a real run")
	}

	var service Service
	var transcriber Transcriber

	if cfg.ServiceFactory != nil {
		log.Printf("Starting ASR service...")
		s, err := cfg.ServiceFactory.Start(ctx)
		if err != nil {
			return summary, fmt.Errorf("start service: %w", err)
		}
		defer func() {
			if err := s.Stop(); err != nil {
				log.Printf("service stop: %v", err)
			}
		}()
		log.Printf("ASR service ready at %s", s.Endpoint())
		service = s
	}

	if cfg.TranscriberFactory != nil {
		endpoint := ""
		if service != nil {
			endpoint = service.Endpoint()
		}
		transcriber = cfg.TranscriberFactory(endpoint)
	}

	var failures []error
	for _, item := range work {
		if err := ctx.Err(); err != nil {
			return summary, err
		}
		if !item.NeedsTranscribe {
			if item.NeedsExport && cfg.Exporter != nil {
				if err := exportExisting(ctx, cfg, item.Video); err != nil {
					failures = append(failures, err)
					if cfg.FailFast {
						return summary, errors.Join(failures...)
					}
				}
			}
			continue
		}
		if err := processOne(ctx, cfg, transcriber, item.Video, service); err != nil {
			failures = append(failures, fmt.Errorf("%s: %w", item.Video.SourceID, err))
			if cfg.FailFast {
				return summary, errors.Join(failures...)
			}
		}
	}
	return summary, errors.Join(failures...)
}

func processOne(ctx context.Context, cfg RunConfig, transcriber Transcriber, v Video, service Service) error {
	log.Printf("Transcribing: %s (%s)", v.SourceID, v.Title)
	endpoint := ""
	if service != nil {
		endpoint = service.Endpoint()
	}
	attempt, err := cfg.Store.BeginAttempt(ctx, v, cfg.Fingerprint, endpoint)
	if err != nil {
		return fmt.Errorf("begin attempt: %w", err)
	}
	result, err := transcriber.Transcribe(ctx, v.AudioPath, TranscribeOptions{ChunkSizeSeconds: cfg.Fingerprint.ChunkSizeSeconds})
	if err != nil {
		_ = cfg.Store.FailAttempt(ctx, attempt, "asr", err.Error(), 0, 0, 0)
		return fmt.Errorf("transcribe: %w", err)
	}
	policy := DefaultChunkPolicy()
	if cfg.Exporter != nil {
		policy = cfg.Exporter.policy
	}
	rev, err := cfg.Store.CommitTranscript(ctx, v, attempt, cfg.Fingerprint, result, policy)
	if err != nil {
		_ = cfg.Store.FailAttempt(ctx, attempt, "commit", err.Error(), result.ProcessingTime.Milliseconds(), result.ChunkCount, len(result.Words))
		return fmt.Errorf("commit: %w", err)
	}
	log.Printf("Committed revision %d: %d words, %d chunks, %.1fs", rev.ID, rev.WordCount, rev.ChunkCount, rev.DurationSeconds)

	if cfg.Exporter != nil && len(cfg.Formats) > 0 {
		data, err := cfg.Store.LoadRevision(ctx, rev.ID)
		if err != nil {
			return fmt.Errorf("load revision for export: %w", err)
		}
		paths, err := cfg.Exporter.Export(ctx, v, data, cfg.Formats)
		if err != nil {
			return fmt.Errorf("export: %w", err)
		}
		if cfg.Verbose {
			log.Printf("Exported: %s", strings.Join(paths, ", "))
		}
	}
	return nil
}

func runExportOnly(ctx context.Context, cfg RunConfig, work []WorkItem) error {
	if cfg.Exporter == nil {
		return nil
	}
	var failures []error
	for _, item := range work {
		if !item.NeedsExport || !item.Video.ActiveRevisionID.Valid {
			continue
		}
		if err := exportExisting(ctx, cfg, item.Video); err != nil {
			failures = append(failures, fmt.Errorf("%s: %w", item.Video.SourceID, err))
			if cfg.FailFast {
				return errors.Join(failures...)
			}
		}
	}
	return errors.Join(failures...)
}

func exportExisting(ctx context.Context, cfg RunConfig, v Video) error {
	if !v.ActiveRevisionID.Valid {
		return nil
	}
	data, err := cfg.Store.LoadRevision(ctx, v.ActiveRevisionID.Int64)
	if err != nil {
		return fmt.Errorf("load revision: %w", err)
	}
	paths, err := cfg.Exporter.Export(ctx, v, data, cfg.Formats)
	if err != nil {
		return fmt.Errorf("export: %w", err)
	}
	if cfg.Verbose {
		log.Printf("Re-exported %s: %s", v.SourceID, strings.Join(paths, ", "))
	}
	return nil
}

func summarize(corpusKey string, work []WorkItem) *PlanSummary {
	s := &PlanSummary{CorpusKey: corpusKey, Total: len(work)}
	for _, item := range work {
		switch item.Reason {
		case "unavailable":
			s.Unavailable++
		case "pending":
			s.Pending++
			s.WillTranscribe++
		case "stale":
			s.Stale++
			s.WillTranscribe++
		case "failed":
			s.Failed++
		case "unchanged":
			s.Unchanged++
			s.ExportOnly++
		case "export-only":
			s.ExportOnly++
		}
	}
	return s
}

func (s *PlanSummary) String() string {
	return fmt.Sprintf(
		"corpus=%s total=%d available=%d unavailable=%d pending=%d stale=%d failed=%d unchanged=%d export_only=%d will_transcribe=%d service_start=%v",
		s.CorpusKey, s.Total, s.Total-s.Unavailable, s.Unavailable,
		s.Pending, s.Stale, s.Failed, s.Unchanged, s.ExportOnly, s.WillTranscribe, s.WillStart,
	)
}
