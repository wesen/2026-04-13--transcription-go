package live

import (
	"math"
	"strings"

	"github.com/go-go-golems/transcription-go/internal/output"
)

const duplicateTimingTolerance = 0.35
const monotonicEndTolerance = 0.05

// TranscriptAccumulator tracks pending and committed words as live results arrive.
type TranscriptAccumulator interface {
	ApplyFinalWords(words []output.Word) error
	CommittedWords() []output.Word
	PendingWords() []output.Word
	Reset()
}

// Accumulator keeps committed transcript words and applies simple overlap-aware
// dedupe suitable for Phase 1 chunked near-live mode.
type Accumulator struct {
	committed     []output.Word
	pending       []output.Word
	lastFinalTime float64
}

func NewAccumulator() *Accumulator {
	return &Accumulator{}
}

func (a *Accumulator) ApplyFinalWords(words []output.Word) error {
	for _, w := range words {
		if !a.shouldAppend(w) {
			continue
		}
		a.committed = append(a.committed, w)
		if w.End > a.lastFinalTime {
			a.lastFinalTime = w.End
		}
	}
	a.pending = nil
	return nil
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
}

func (a *Accumulator) shouldAppend(w output.Word) bool {
	if strings.TrimSpace(w.Word) == "" {
		return false
	}
	if len(a.committed) == 0 {
		return true
	}
	if w.End <= a.lastFinalTime+monotonicEndTolerance {
		return false
	}
	for i := len(a.committed) - 1; i >= 0; i-- {
		prev := a.committed[i]
		if prev.End < w.Start-1.0 {
			break
		}
		if sameWord(prev.Word, w.Word) && math.Abs(prev.Start-w.Start) <= duplicateTimingTolerance && math.Abs(prev.End-w.End) <= duplicateTimingTolerance {
			return false
		}
	}
	return true
}

func sameWord(a, b string) bool {
	normalize := func(s string) string {
		return strings.ToLower(strings.TrimSpace(strings.Trim(s, ".,!?;:\"'")))
	}
	return normalize(a) == normalize(b)
}
