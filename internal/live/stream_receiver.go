package live

import (
	"context"
	"fmt"
)

// StreamMessage is the transport-neutral output of the WS receiver loop.
type StreamMessage struct {
	Event   *TranscriptEvent
	Stopped *StreamStopped
}

// StreamStopped carries the terminal server-side stop event payload.
type StreamStopped struct {
	SessionID string
	WordCount int
	Duration  float64
}

// ReceiveResultEvents reads WS server messages and converts them into
// transport-neutral transcript events plus a terminal stopped message.
func ReceiveResultEvents(ctx context.Context, client *WSLiveClient, out chan<- StreamMessage) error {
	defer close(out)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		msg, err := client.ReadServerMessage()
		if err != nil {
			return err
		}
		switch msg.Type {
		case "started":
			continue
		case "partial":
			out <- StreamMessage{Event: &TranscriptEvent{
				Type:         TranscriptEventPartial,
				SessionID:    msg.SessionID,
				Sequence:     msg.Sequence,
				Words:        msg.Words,
				ProcessingMS: msg.ProcessingMS,
			}}
		case "final_words":
			out <- StreamMessage{Event: &TranscriptEvent{
				Type:         TranscriptEventFinalWords,
				SessionID:    msg.SessionID,
				UpToTime:     msg.UpToTime,
				Words:        msg.Words,
				ProcessingMS: msg.ProcessingMS,
			}}
		case "stopped":
			out <- StreamMessage{Stopped: &StreamStopped{
				SessionID: msg.SessionID,
				WordCount: msg.WordCount,
				Duration:  msg.Duration,
			}}
			return nil
		case "error":
			return fmt.Errorf("ws server error %s: %s", msg.Code, msg.Message)
		default:
			return fmt.Errorf("unsupported ws server message type %q", msg.Type)
		}
	}
}
