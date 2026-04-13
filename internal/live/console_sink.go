package live

import (
	"log"
	"strings"

	"github.com/go-go-golems/transcription-go/internal/output"
)

// ConsoleSink prints newly committed transcript words incrementally and can
// optionally surface preview text when pending words change.
type ConsoleSink struct {
	printed     int
	lastPreview string
}

func NewConsoleSink() *ConsoleSink {
	return &ConsoleSink{}
}

func (s *ConsoleSink) Update(state TranscriptState) error {
	if len(state.Committed) > s.printed {
		newWords := state.Committed[s.printed:]
		log.Printf("Committed +%d words: %s", len(newWords), joinOutputWords(newWords))
		s.printed = len(state.Committed)
	}

	preview := joinOutputWords(state.Pending)
	if preview != "" && preview != s.lastPreview {
		log.Printf("Preview: %s", preview)
	}
	s.lastPreview = preview
	return nil
}

func joinOutputWords(words []output.Word) string {
	parts := make([]string, 0, len(words))
	for _, w := range words {
		parts = append(parts, w.Word)
	}
	return strings.Join(parts, " ")
}
