package output

import (
	"fmt"
	"io"
	"strings"
)

// Segment is a grouping of consecutive words for subtitle display.
type Segment struct {
	Start float64
	End   float64
	Text  string
}

// BuildSegments groups words into segments based on sentence boundaries
// and max duration/length constraints.
func BuildSegments(words []Word, maxDuration float64, maxChars int) []Segment {
	if maxDuration == 0 {
		maxDuration = 15.0
	}
	if maxChars == 0 {
		maxChars = 120
	}

	splitOn := []string{".", "?", "!"}
	var segments []Segment
	var current []string
	var segStart float64

	for _, w := range words {
		if len(current) == 0 {
			segStart = w.Start
		}
		current = append(current, w.Word)

		text := strings.Join(current, " ")
		duration := w.End - segStart

		shouldSplit := false
		for _, ending := range splitOn {
			if strings.HasSuffix(w.Word, ending) && duration > 0.5 {
				shouldSplit = true
				break
			}
		}
		if duration > maxDuration || len(text) > maxChars {
			shouldSplit = true
		}

		if shouldSplit && len(current) > 0 {
			segments = append(segments, Segment{
				Start: segStart,
				End:   w.End,
				Text:  text,
			})
			current = nil
		}
	}

	if len(current) > 0 {
		lastWord := words[len(words)-1]
		segments = append(segments, Segment{
			Start: segStart,
			End:   lastWord.End,
			Text:  strings.Join(current, " "),
		})
	}

	return segments
}

func formatSRTTime(s float64) string {
	h := int(s / 3600)
	m := int((s / 60)) % 60
	sec := int(s) % 60
	ms := int((s - float64(int(s))) * 1000)
	return fmt.Sprintf("%02d:%02d:%02d,%03d", h, m, sec, ms)
}

// WriteSRT writes segments in SubRip (.srt) format.
func WriteSRT(w io.Writer, segments []Segment) error {
	for i, seg := range segments {
		fmt.Fprintf(w, "%d\n", i+1)
		fmt.Fprintf(w, "%s --> %s\n", formatSRTTime(seg.Start), formatSRTTime(seg.End))
		fmt.Fprintf(w, "%s\n\n", seg.Text)
	}
	return nil
}

// WriteVTT writes segments in WebVTT (.vtt) format.
func WriteVTT(w io.Writer, segments []Segment) error {
	fmt.Fprintf(w, "WEBVTT\n\n")
	for _, seg := range segments {
		// VTT uses . instead of , for milliseconds
		start := formatSRTTime(seg.Start)
		end := formatSRTTime(seg.End)
		fmt.Fprintf(w, "%s --> %s\n", strings.Replace(start, ",", ".", 1), strings.Replace(end, ",", ".", 1))
		fmt.Fprintf(w, "%s\n\n", seg.Text)
	}
	return nil
}

// WriteTXT writes a plain text transcript.
func WriteTXT(w io.Writer, segments []Segment) error {
	for _, seg := range segments {
		fmt.Fprintf(w, "%s\n", seg.Text)
	}
	return nil
}
