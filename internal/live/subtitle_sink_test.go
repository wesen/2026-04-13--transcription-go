package live

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-go-golems/transcription-go/internal/output"
)

func TestSubtitleSinkWritesSRTFromCommittedWordsOnly(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	sink := NewSubtitleSink(dir, "srt")

	state := TranscriptState{
		Committed: []output.Word{
			{Word: "Hello", Start: 0.0, End: 0.5},
			{Word: "world.", Start: 0.5, End: 1.0},
		},
		Pending: []output.Word{{Word: "preview-only", Start: 1.0, End: 1.2}},
	}
	if err := sink.Update(state); err != nil {
		t.Fatalf("Update: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "transcript.srt"))
	if err != nil {
		t.Fatalf("read transcript.srt: %v", err)
	}
	text := string(data)
	if !strings.Contains(text, "Hello world.") {
		t.Fatalf("unexpected srt content: %s", text)
	}
	if strings.Contains(text, "preview-only") {
		t.Fatalf("pending preview should not be written: %s", text)
	}
}
