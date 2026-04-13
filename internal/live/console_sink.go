package live

import (
	"log"
	"strings"

	"github.com/go-go-golems/transcription-go/internal/output"
)

// ConsoleSink prints newly committed transcript words incrementally.
type ConsoleSink struct {
	printed int
}

func NewConsoleSink() *ConsoleSink {
	return &ConsoleSink{}
}

func (s *ConsoleSink) WriteCommitted(words []output.Word) {
	if len(words) <= s.printed {
		return
	}
	newWords := words[s.printed:]
	parts := make([]string, 0, len(newWords))
	for _, w := range newWords {
		parts = append(parts, w.Word)
	}
	log.Printf("Committed +%d words: %s", len(newWords), strings.Join(parts, " "))
	s.printed = len(words)
}
