package live

import (
	"fmt"
	"math"
	"strings"

	"github.com/go-go-golems/transcription-go/internal/output"
)

const duplicateTimingTolerance = 0.35
const monotonicEndTolerance = 0.05

// TranscriptAccumulator tracks pending and committed words as live results arrive.
type TranscriptAccumulator interface {
	ApplyEvent(event TranscriptEvent) error
	ApplyPartialWords(words []output.Word) error
	ApplyFinalWords(words []output.Word) error
	State() TranscriptState
	CommittedWords() []output.Word
	PendingWords() []output.Word
	Reset()
}

// Accumulator keeps committed transcript words and explicit pending preview
// words. Phase 1 chunk mode currently feeds only final events, but the state
// model matches the planned WebSocket path.
type Accumulator struct {
	committed         []output.Word
	pending           []output.Word
	lastFinalTime     float64
	haveSequence      bool
	lastSequence      int
	sequenceFinalized bool
}

func NewAccumulator() *Accumulator {
	return &Accumulator{lastSequence: -1}
}

func (a *Accumulator) ApplyEvent(event TranscriptEvent) error {
	if err := a.validateSequence(event); err != nil {
		return err
	}

	switch event.Type {
	case TranscriptEventPartial:
		a.pending = a.filterIncoming(event.Words, a.committed)
		return nil
	case TranscriptEventFinalWords:
		for _, w := range a.filterIncoming(event.Words, a.committed) {
			a.committed = append(a.committed, w)
			if w.End > a.lastFinalTime {
				a.lastFinalTime = w.End
			}
		}
		if event.UpToTime > a.lastFinalTime {
			a.lastFinalTime = event.UpToTime
		}
		a.pending = a.filterIncoming(a.pending, a.committed)
		return nil
	default:
		return fmt.Errorf("unsupported transcript event type %q", event.Type)
	}
}

func (a *Accumulator) ApplyPartialWords(words []output.Word) error {
	return a.ApplyEvent(TranscriptEvent{Type: TranscriptEventPartial, Sequence: -1, Words: words})
}

func (a *Accumulator) ApplyFinalWords(words []output.Word) error {
	return a.ApplyEvent(TranscriptEvent{Type: TranscriptEventFinalWords, Sequence: -1, UpToTime: maxWordEnd(words), Words: words})
}

func (a *Accumulator) State() TranscriptState {
	return TranscriptState{
		Committed:     append([]output.Word(nil), a.committed...),
		Pending:       append([]output.Word(nil), a.pending...),
		LastFinalTime: a.lastFinalTime,
	}
}

func (a *Accumulator) CommittedWords() []output.Word {
	return append([]output.Word(nil), a.committed...)
}

func (a *Accumulator) PendingWords() []output.Word {
	return append([]output.Word(nil), a.pending...)
}

func (a *Accumulator) Reset() {
	a.committed = nil
	a.pending = nil
	a.lastFinalTime = 0
	a.haveSequence = false
	a.lastSequence = -1
	a.sequenceFinalized = false
}

func (a *Accumulator) validateSequence(event TranscriptEvent) error {
	if event.Sequence < 0 {
		return nil
	}
	if !a.haveSequence {
		a.haveSequence = true
		a.lastSequence = event.Sequence
		a.sequenceFinalized = event.Type == TranscriptEventFinalWords
		return nil
	}
	if event.Sequence < a.lastSequence {
		return fmt.Errorf("out-of-order transcript event sequence: got %d after %d", event.Sequence, a.lastSequence)
	}
	if event.Sequence > a.lastSequence {
		a.lastSequence = event.Sequence
		a.sequenceFinalized = event.Type == TranscriptEventFinalWords
		return nil
	}
	if a.sequenceFinalized {
		return fmt.Errorf("transcript sequence %d already finalized", event.Sequence)
	}
	if event.Type == TranscriptEventFinalWords {
		a.sequenceFinalized = true
	}
	return nil
}

func (a *Accumulator) filterIncoming(words []output.Word, existing []output.Word) []output.Word {
	filtered := make([]output.Word, 0, len(words))
	seen := append([]output.Word(nil), existing...)
	for _, w := range words {
		if !a.shouldAppend(w, seen) {
			continue
		}
		filtered = append(filtered, w)
		seen = append(seen, w)
	}
	return filtered
}

func (a *Accumulator) shouldAppend(w output.Word, existing []output.Word) bool {
	if strings.TrimSpace(w.Word) == "" {
		return false
	}
	if w.End <= a.lastFinalTime+monotonicEndTolerance {
		return false
	}
	for i := len(existing) - 1; i >= 0; i-- {
		prev := existing[i]
		if prev.End < w.Start-1.0 {
			break
		}
		if sameWord(prev.Word, w.Word) && math.Abs(prev.Start-w.Start) <= duplicateTimingTolerance && math.Abs(prev.End-w.End) <= duplicateTimingTolerance {
			return false
		}
	}
	return true
}

func maxWordEnd(words []output.Word) float64 {
	var maxEnd float64
	for _, w := range words {
		if w.End > maxEnd {
			maxEnd = w.End
		}
	}
	return maxEnd
}

func sameWord(a, b string) bool {
	normalize := func(s string) string {
		return strings.ToLower(strings.TrimSpace(strings.Trim(s, ".,!?;:\"'")))
	}
	return normalize(a) == normalize(b)
}
