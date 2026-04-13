package live

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-go-golems/transcription-go/internal/output"
)

func TestSubtitleSinkWritesSRT(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	sink := NewSubtitleSink(dir, "srt")

	words := []output.Word{
		{Word: "Hello", Start: 0.0, End: 0.5},
		{Word: "world.", Start: 0.5, End: 1.0},
	}
	if err := sink.Update(words); err != nil {
		t.Fatalf("Update: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "transcript.srt"))
	if err != nil {
		t.Fatalf("read transcript.srt: %v", err)
	}
	if !strings.Contains(string(data), "Hello world.") {
		t.Fatalf("unexpected srt content: %s", string(data))
	}
}
