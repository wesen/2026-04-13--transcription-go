package live

import "github.com/go-go-golems/transcription-go/internal/output"

// TranscriptAccumulator tracks pending and committed words as live results arrive.
type TranscriptAccumulator interface {
	ApplyFinalWords(words []output.Word) error
	CommittedWords() []output.Word
	PendingWords() []output.Word
	Reset()
}

// Accumulator is a placeholder implementation that will gain overlap handling
// and partial/final semantics in later tasks.
type Accumulator struct {
	committed []output.Word
	pending   []output.Word
}

func NewAccumulator() *Accumulator {
	return &Accumulator{}
}

func (a *Accumulator) ApplyFinalWords(words []output.Word) error {
	a.committed = append(a.committed, words...)
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
}
