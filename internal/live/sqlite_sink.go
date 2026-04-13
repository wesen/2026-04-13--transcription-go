package live

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-go-golems/transcription-go/internal/output"
)

// SQLiteSink rewrites a transcript database from the committed transcript state.
type SQLiteSink struct {
	path string
}

func NewSQLiteSink(path string) *SQLiteSink {
	return &SQLiteSink{path: path}
}

func (s *SQLiteSink) Update(words []output.Word) error {
	if err := ensureOutputDir(filepath.Dir(s.path)); err != nil {
		return fmt.Errorf("ensure sqlite dir: %w", err)
	}
	_ = os.Remove(s.path)
	if err := output.WriteSQLite(words, s.path); err != nil {
		return fmt.Errorf("write sqlite: %w", err)
	}
	return nil
}
