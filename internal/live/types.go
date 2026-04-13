package live

import "github.com/go-go-golems/transcription-go/internal/output"

// TranscriptEventType identifies the kind of transcript update arriving from a
// live transport.
type TranscriptEventType string

const (
	TranscriptEventPartial    TranscriptEventType = "partial"
	TranscriptEventFinalWords TranscriptEventType = "final_words"
)

// TranscriptEvent is the transport-neutral representation of a live transcript
// update. It is intentionally shaped like the planned WebSocket events so the
// accumulator can stay stable as the transport evolves.
type TranscriptEvent struct {
	Type      TranscriptEventType
	SessionID string
	Sequence  int
	UpToTime  float64
	Words     []output.Word
}

// TranscriptState is the accumulator snapshot consumed by live sinks.
type TranscriptState struct {
	Committed     []output.Word
	Pending       []output.Word
	LastFinalTime float64
}
