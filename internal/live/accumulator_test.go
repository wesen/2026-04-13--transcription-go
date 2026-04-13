package live

import (
	"strings"
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

func TestAccumulatorRepeatedPartialRevisionReplacesPending(t *testing.T) {
	t.Parallel()

	a := NewAccumulator()
	err := a.ApplyEvent(TranscriptEvent{
		Type:     TranscriptEventPartial,
		Sequence: 0,
		Words: []output.Word{
			{Word: "hello", Start: 0.00, End: 0.30},
			{Word: "wor", Start: 0.31, End: 0.50},
		},
	})
	if err != nil {
		t.Fatalf("ApplyEvent(partial 1): %v", err)
	}
	err = a.ApplyEvent(TranscriptEvent{
		Type:     TranscriptEventPartial,
		Sequence: 0,
		Words: []output.Word{
			{Word: "hello", Start: 0.00, End: 0.30},
			{Word: "world", Start: 0.31, End: 0.65},
		},
	})
	if err != nil {
		t.Fatalf("ApplyEvent(partial 2): %v", err)
	}

	state := a.State()
	if len(state.Committed) != 0 {
		t.Fatalf("expected no committed words yet, got %+v", state.Committed)
	}
	if got := joinTestWords(state.Pending); got != "hello world" {
		t.Fatalf("expected pending revision 'hello world', got %q", got)
	}
}

func TestAccumulatorPartialToFinalPromotion(t *testing.T) {
	t.Parallel()

	a := NewAccumulator()
	if err := a.ApplyEvent(TranscriptEvent{
		Type:     TranscriptEventPartial,
		Sequence: 0,
		Words: []output.Word{
			{Word: "hello", Start: 0.00, End: 0.30},
			{Word: "world", Start: 0.31, End: 0.65},
		},
	}); err != nil {
		t.Fatalf("ApplyEvent(partial): %v", err)
	}
	if err := a.ApplyEvent(TranscriptEvent{
		Type:     TranscriptEventFinalWords,
		Sequence: 0,
		UpToTime: 0.65,
		Words: []output.Word{
			{Word: "hello", Start: 0.00, End: 0.30},
			{Word: "world", Start: 0.31, End: 0.65},
		},
	}); err != nil {
		t.Fatalf("ApplyEvent(final): %v", err)
	}

	state := a.State()
	if got := joinTestWords(state.Committed); got != "hello world" {
		t.Fatalf("expected committed 'hello world', got %q", got)
	}
	if len(state.Pending) != 0 {
		t.Fatalf("expected pending to be cleared, got %+v", state.Pending)
	}
	if state.LastFinalTime < 0.65 {
		t.Fatalf("expected last final time >= 0.65, got %.3f", state.LastFinalTime)
	}
}

func TestAccumulatorPartialSkipsAlreadyCommittedOverlap(t *testing.T) {
	t.Parallel()

	a := NewAccumulator()
	if err := a.ApplyEvent(TranscriptEvent{
		Type:     TranscriptEventFinalWords,
		Sequence: 0,
		UpToTime: 0.65,
		Words: []output.Word{
			{Word: "hello", Start: 0.00, End: 0.30},
			{Word: "world", Start: 0.31, End: 0.65},
		},
	}); err != nil {
		t.Fatalf("ApplyEvent(final 0): %v", err)
	}
	if err := a.ApplyEvent(TranscriptEvent{
		Type:     TranscriptEventPartial,
		Sequence: 1,
		Words: []output.Word{
			{Word: "world", Start: 0.31, End: 0.65},
			{Word: "again", Start: 0.66, End: 1.00},
		},
	}); err != nil {
		t.Fatalf("ApplyEvent(partial 1): %v", err)
	}

	if got := joinTestWords(a.PendingWords()); got != "again" {
		t.Fatalf("expected only non-overlapping pending word, got %q", got)
	}
}

func TestAccumulatorRejectsOutOfOrderSequence(t *testing.T) {
	t.Parallel()

	a := NewAccumulator()
	if err := a.ApplyEvent(TranscriptEvent{
		Type:     TranscriptEventFinalWords,
		Sequence: 1,
		UpToTime: 1.00,
		Words:    []output.Word{{Word: "one", Start: 0.00, End: 1.00}},
	}); err != nil {
		t.Fatalf("ApplyEvent(final 1): %v", err)
	}

	err := a.ApplyEvent(TranscriptEvent{
		Type:     TranscriptEventPartial,
		Sequence: 0,
		Words:    []output.Word{{Word: "late", Start: 0.20, End: 0.40}},
	})
	if err == nil {
		t.Fatal("expected out-of-order sequence error, got nil")
	}
	if !strings.Contains(err.Error(), "out-of-order") {
		t.Fatalf("expected out-of-order error, got %v", err)
	}
}

func joinTestWords(words []output.Word) string {
	parts := make([]string, 0, len(words))
	for _, w := range words {
		parts = append(parts, w.Word)
	}
	return strings.Join(parts, " ")
}
