package live

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-go-golems/transcription-go/internal/output"
)

// SubtitleSink rewrites a full transcript artifact from committed words.
type SubtitleSink struct {
	outputDir string
	format    string
}

func NewSubtitleSink(outputDir, format string) *SubtitleSink {
	return &SubtitleSink{outputDir: outputDir, format: format}
}

func (s *SubtitleSink) Update(words []output.Word) error {
	if err := ensureOutputDir(s.outputDir); err != nil {
		return fmt.Errorf("ensure output dir: %w", err)
	}
	segments := output.BuildSegments(words, 15.0, 120)
	path := filepath.Join(s.outputDir, "transcript."+s.format)
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}
	defer f.Close()

	switch s.format {
	case "srt":
		return output.WriteSRT(f, segments)
	case "vtt":
		return output.WriteVTT(f, segments)
	case "txt":
		return output.WriteTXT(f, segments)
	default:
		return fmt.Errorf("unsupported subtitle format %q", s.format)
	}
}
