package live

import (
	"testing"

	"github.com/go-go-golems/transcription-go/internal/output"
)

func TestAccumulatorSkipsOverlappingDuplicates(t *testing.T) {
	t.Parallel()

	a := NewAccumulator()
	if err := a.ApplyFinalWords([]output.Word{
		{Word: "hello", Start: 0.00, End: 0.30},
		{Word: "world", Start: 0.31, End: 0.65},
	}); err != nil {
		t.Fatalf("ApplyFinalWords(first): %v", err)
	}
	if err := a.ApplyFinalWords([]output.Word{
		{Word: "world", Start: 0.31, End: 0.65},
		{Word: "again", Start: 0.66, End: 1.00},
	}); err != nil {
		t.Fatalf("ApplyFinalWords(second): %v", err)
	}

	got := a.CommittedWords()
	if len(got) != 3 {
		t.Fatalf("expected 3 committed words, got %d: %+v", len(got), got)
	}
	if got[2].Word != "again" {
		t.Fatalf("expected last word 'again', got %+v", got[2])
	}
}

func TestAccumulatorSkipsNearDuplicateTiming(t *testing.T) {
	t.Parallel()

	a := NewAccumulator()
	_ = a.ApplyFinalWords([]output.Word{{Word: "streaming", Start: 1.00, End: 1.40}})
	_ = a.ApplyFinalWords([]output.Word{{Word: "streaming", Start: 1.08, End: 1.43}})

	got := a.CommittedWords()
	if len(got) != 1 {
		t.Fatalf("expected 1 committed word, got %d: %+v", len(got), got)
	}
}
