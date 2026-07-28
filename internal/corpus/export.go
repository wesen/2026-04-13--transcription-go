package corpus

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-go-golems/transcription-go/internal/output"
)

// ExportFormat is one supported per-video output format.
type ExportFormat string

const (
	ExportSRT ExportFormat = "srt"
	ExportVTT ExportFormat = "vtt"
	ExportTXT ExportFormat = "txt"
)

// Exporter generates per-video files from committed revisions.
type Exporter struct {
	store      *Store
	outputRoot string
	policy     ChunkPolicy
}

// NewExporter creates an exporter writing under outputRoot.
func NewExporter(store *Store, outputRoot string, policy ChunkPolicy) *Exporter {
	return &Exporter{store: store, outputRoot: outputRoot, policy: policy}
}

// ExportPolicyJSON returns the canonical policy JSON used for export rows.
func (e *Exporter) ExportPolicyJSON() string {
	b, _ := json.Marshal(e.policy)
	return string(b)
}

// Export writes the requested formats for one video's revision.
func (e *Exporter) Export(ctx context.Context, v Video, data *RevisionData, formats []ExportFormat) ([]string, error) {
	if err := os.MkdirAll(e.outputRoot, 0o755); err != nil {
		return nil, fmt.Errorf("create output root: %w", err)
	}
	videoDir := filepath.Join(e.outputRoot, safeDirName(v))
	if err := os.MkdirAll(videoDir, 0o755); err != nil {
		return nil, fmt.Errorf("create video dir: %w", err)
	}

	words := corpusWordsToOutput(data.Words)
	segWords := words
	if !e.policy.IncludeRemoved {
		segWords = output.FilterFillers(words)
	}
	segments := output.BuildSegments(segWords, e.policy.MaxDurationSeconds, e.policy.MaxCharacters)

	policyJSON := e.ExportPolicyJSON()
	var paths []string
	for _, f := range formats {
		var path string
		var content []byte
		switch f {
		case ExportSRT:
			path = filepath.Join(videoDir, "transcript.srt")
			c, err := renderSRT(segments)
			if err != nil {
				return paths, err
			}
			content = c
		case ExportVTT:
			path = filepath.Join(videoDir, "transcript.vtt")
			c, err := renderVTT(segments)
			if err != nil {
				return paths, err
			}
			content = c
		case ExportTXT:
			path = filepath.Join(videoDir, "transcript.txt")
			c, err := renderTXT(segments)
			if err != nil {
				return paths, err
			}
			content = c
		default:
			return paths, fmt.Errorf("unsupported export format %q", f)
		}
		if err := os.WriteFile(path, content, 0o644); err != nil {
			return paths, fmt.Errorf("write %s: %w", path, err)
		}
		hash := sha256Sum(content)
		if err := e.store.recordExport(ctx, data.Revision.ID, string(f), policyJSON, path, hash); err != nil {
			return paths, fmt.Errorf("record export %s: %w", f, err)
		}
		paths = append(paths, path)
	}
	return paths, nil
}

func corpusWordsToOutput(ws []Word) []output.Word {
	out := make([]output.Word, len(ws))
	for i, w := range ws {
		out[i] = output.Word{Word: w.Text, Start: w.Start, End: w.End}
	}
	return out
}

func renderSRT(segments []output.Segment) ([]byte, error) {
	var buf strings.Builder
	if err := output.WriteSRT(&buf, segments); err != nil {
		return nil, err
	}
	return []byte(buf.String()), nil
}

func renderVTT(segments []output.Segment) ([]byte, error) {
	var buf strings.Builder
	if err := output.WriteVTT(&buf, segments); err != nil {
		return nil, err
	}
	return []byte(buf.String()), nil
}

func renderTXT(segments []output.Segment) ([]byte, error) {
	var buf strings.Builder
	if err := output.WriteTXT(&buf, segments); err != nil {
		return nil, err
	}
	return []byte(buf.String()), nil
}

func safeDirName(v Video) string {
	name := fmt.Sprintf("%03d - %s [%s]", v.PlaylistPosition, v.Title, v.SourceID)
	for _, r := range name {
		if r == '/' || r == 0 {
			name = strings.ReplaceAll(name, string(r), "-")
		}
	}
	return name
}

func sha256Sum(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}
