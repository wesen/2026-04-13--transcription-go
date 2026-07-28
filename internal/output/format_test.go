package output

import (
	"bytes"
	"database/sql"
	"path/filepath"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

func TestBuildSegmentsTrailingBufferHasNonZeroEnd(t *testing.T) {
	words := []Word{
		{Word: "Hello", Start: 0, End: 0.5},
		{Word: "world.", Start: 0.5, End: 1.0},
		{Word: "No", Start: 1.5, End: 1.8},
		{Word: "terminator", Start: 1.8, End: 2.3},
	}

	segments := BuildSegments(words, 15.0, 120)
	if len(segments) < 2 {
		t.Fatalf("expected at least 2 segments, got %d", len(segments))
	}
	last := segments[len(segments)-1]
	if last.End == 0 {
		t.Errorf("trailing segment end is zero; want last word end %.3f, got 0", words[len(words)-1].End)
	}
	if last.End != words[len(words)-1].End {
		t.Errorf("trailing segment end = %.3f, want %.3f", last.End, words[len(words)-1].End)
	}
}

func TestWriteSQLiteChunkWordCount(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	words := []Word{
		{Word: "Hello", Start: 0, End: 0.5},
		{Word: "world.", Start: 0.5, End: 1.0},
		{Word: "Another", Start: 1.5, End: 1.8},
		{Word: "sentence.", Start: 1.8, End: 2.3},
	}

	if err := WriteSQLite(words, dbPath); err != nil {
		t.Fatalf("WriteSQLite failed: %v", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	var zeroCountChunks int
	if err := db.QueryRow(`SELECT COUNT(*) FROM chunks WHERE word_count = 0`).Scan(&zeroCountChunks); err != nil {
		t.Fatal(err)
	}
	if zeroCountChunks != 0 {
		t.Errorf("found %d chunks with word_count=0; expected every chunk to link real words", zeroCountChunks)
	}

	var unmatchedLinks int
	if err := db.QueryRow(`
		SELECT COUNT(*) FROM chunks c
		WHERE c.word_count != (SELECT COUNT(*) FROM chunk_words cw WHERE cw.chunk_id = c.id)
	`).Scan(&unmatchedLinks); err != nil {
		t.Fatal(err)
	}
	if unmatchedLinks != 0 {
		t.Errorf("found %d chunks whose word_count does not match linked chunk_words rows", unmatchedLinks)
	}
}

func TestFilterFillers(t *testing.T) {
	words := []Word{
		{Word: "Hello", Start: 0, End: 1},
		{Word: "um", Start: 1, End: 1.5},
		{Word: "world", Start: 1.5, End: 2},
		{Word: "uh", Start: 2, End: 2.3},
		{Word: "test", Start: 2.3, End: 3},
	}

	filtered := FilterFillers(words)
	if len(filtered) != 3 {
		t.Errorf("expected 3 words, got %d", len(filtered))
	}
	if filtered[0].Word != "Hello" || filtered[1].Word != "world" || filtered[2].Word != "test" {
		t.Errorf("unexpected words: %v", filtered)
	}
}

func TestIsFiller(t *testing.T) {
	tests := []struct {
		word string
		want bool
	}{
		{"um", true},
		{"Um", true},
		{"um.", true},
		{"UH?", true},
		{"hello", false},
		{"world", false},
		{"like", true},
		{"er,", true},
	}
	for _, tt := range tests {
		w := Word{Word: tt.word}
		if got := w.IsFiller(); got != tt.want {
			t.Errorf("IsFiller(%q) = %v, want %v", tt.word, got, tt.want)
		}
	}
}

func TestBuildSegments(t *testing.T) {
	words := []Word{
		{Word: "Hello", Start: 0, End: 0.5},
		{Word: "world.", Start: 0.5, End: 1.0},
		{Word: "This", Start: 1.5, End: 2.0},
		{Word: "is", Start: 2.0, End: 2.5},
		{Word: "a", Start: 2.5, End: 2.7},
		{Word: "test.", Start: 2.7, End: 3.2},
	}

	segments := BuildSegments(words, 15.0, 120)
	if len(segments) < 2 {
		t.Errorf("expected at least 2 segments, got %d", len(segments))
	}

	// First segment should split on the period after "world."
	if segments[0].Text != "Hello world." {
		t.Errorf("first segment text = %q, want %q", segments[0].Text, "Hello world.")
	}
}

func TestWriteSRT(t *testing.T) {
	segments := []Segment{
		{Start: 0, End: 1.5, Text: "Hello world."},
		{Start: 2.0, End: 3.5, Text: "This is a test."},
	}

	var buf bytes.Buffer
	err := WriteSRT(&buf, segments)
	if err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "1\n") {
		t.Error("SRT should contain index 1")
	}
	if !strings.Contains(output, "00:00:00,000 --> 00:00:01,500") {
		t.Error("SRT should contain first timestamp")
	}
	if !strings.Contains(output, "Hello world.") {
		t.Error("SRT should contain first text")
	}
}

func TestWriteVTT(t *testing.T) {
	segments := []Segment{
		{Start: 0, End: 1.5, Text: "Hello world."},
	}

	var buf bytes.Buffer
	err := WriteVTT(&buf, segments)
	if err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "WEBVTT") {
		t.Error("VTT should start with WEBVTT header")
	}
	if !strings.Contains(output, "00:00:00.000 --> 00:00:01.500") {
		t.Error("VTT should use . for milliseconds")
	}
}

func TestWriteTXT(t *testing.T) {
	segments := []Segment{
		{Start: 0, End: 1.5, Text: "Hello world."},
		{Start: 2.0, End: 3.5, Text: "This is a test."},
	}

	var buf bytes.Buffer
	err := WriteTXT(&buf, segments)
	if err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "Hello world.\nThis is a test.") {
		t.Errorf("TXT output = %q", output)
	}
}
