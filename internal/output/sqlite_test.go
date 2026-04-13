package output

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func TestWriteSQLite(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	words := []Word{
		{Word: "Hello", Start: 0, End: 0.5},
		{Word: "um", Start: 0.5, End: 1.0},
		{Word: "world", Start: 1.0, End: 1.5},
		{Word: "test.", Start: 1.5, End: 2.0},
	}

	err := WriteSQLite(words, dbPath)
	if err != nil {
		t.Fatalf("WriteSQLite failed: %v", err)
	}

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Fatal("database file not created")
	}

	// Verify the database content
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Check word count
	var wordCount int
	db.QueryRow("SELECT COUNT(*) FROM words").Scan(&wordCount)
	if wordCount != 4 {
		t.Errorf("expected 4 words, got %d", wordCount)
	}

	// Check filler count
	var fillerCount int
	db.QueryRow("SELECT COUNT(*) FROM words WHERE is_filler = 1").Scan(&fillerCount)
	if fillerCount != 1 {
		t.Errorf("expected 1 filler word, got %d", fillerCount)
	}

	// Check chunks exist
	var chunkCount int
	db.QueryRow("SELECT COUNT(*) FROM chunks").Scan(&chunkCount)
	if chunkCount == 0 {
		t.Error("expected at least 1 chunk")
	}
}
