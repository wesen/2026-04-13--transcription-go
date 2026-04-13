package live

import (
	"fmt"
	"os"
	"path/filepath"
)

// Sink consumes transcript state and updates a live output artifact. Durable
// artifacts should derive from committed words only.
type Sink interface {
	Update(state TranscriptState) error
}

func buildSinks(outputDir string, formats []string) ([]Sink, error) {
	sinks := make([]Sink, 0, len(formats))
	for _, format := range formats {
		switch format {
		case "", "console":
			sinks = append(sinks, NewConsoleSink())
		case "srt", "vtt", "txt":
			if outputDir == "" {
				return nil, fmt.Errorf("live output dir is required for format %q", format)
			}
			sinks = append(sinks, NewSubtitleSink(outputDir, format))
		case "db":
			if outputDir == "" {
				return nil, fmt.Errorf("live output dir is required for format %q", format)
			}
			sinks = append(sinks, NewSQLiteSink(filepath.Join(outputDir, "transcript.db")))
		default:
			return nil, fmt.Errorf("unsupported live format %q", format)
		}
	}
	if len(sinks) == 0 {
		sinks = append(sinks, NewConsoleSink())
	}
	return sinks, nil
}

func ensureOutputDir(path string) error {
	if path == "" {
		return nil
	}
	return os.MkdirAll(path, 0o755)
}
