// Package output provides transcript formatters: SQLite, SRT, VTT, TXT, and filler word filtering.
package output

import (
	"strings"
)

// FillerWords is the set of words treated as fillers.
var FillerWords = map[string]bool{
	"um": true, "uh": true, "er": true, "ah": true,
	"like": true, "you know": true, "sort of": true, "kind of": true,
}

// Word is a transcript word with timing.
type Word struct {
	Word  string
	Start float64
	End   float64
}

// IsFiller reports whether the word is a filler.
func (w Word) IsFiller() bool {
	clean := strings.ToLower(strings.Trim(w.Word, ".,!?;:\"'"))
	return FillerWords[clean]
}

// FilterFillers returns a new slice with filler words removed.
func FilterFillers(words []Word) []Word {
	result := make([]Word, 0, len(words))
	for _, w := range words {
		if !w.IsFiller() {
			result = append(result, w)
		}
	}
	return result
}
