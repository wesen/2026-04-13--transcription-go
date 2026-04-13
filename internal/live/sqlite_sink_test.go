package live

import (
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/go-go-golems/transcription-go/internal/output"
	_ "modernc.org/sqlite"
)

func TestSQLiteSinkWritesCommittedWordsOnly(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "transcript.db")
	sink := NewSQLiteSink(path)

	state := TranscriptState{
		Committed: []output.Word{
			{Word: "Hello", Start: 0.0, End: 0.5},
			{Word: "world", Start: 0.5, End: 1.0},
		},
		Pending: []output.Word{{Word: "preview-only", Start: 1.1, End: 1.3}},
	}
	if err := sink.Update(state); err != nil {
		t.Fatalf("Update: %v", err)
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer db.Close()
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM words").Scan(&count); err != nil {
		t.Fatalf("count words: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2 committed words, got %d", count)
	}
}
