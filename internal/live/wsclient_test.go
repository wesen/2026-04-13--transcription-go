package live

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestWSLiveClientAndReceiver(t *testing.T) {
	t.Parallel()

	upgrader := websocket.Upgrader{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/transcribe/stream" {
			http.NotFound(w, r)
			return
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("upgrade: %v", err)
			return
		}
		defer conn.Close()

		var start map[string]any
		if err := conn.ReadJSON(&start); err != nil {
			t.Errorf("read start: %v", err)
			return
		}
		if start["type"] != "start" {
			t.Errorf("expected start event, got %+v", start)
			return
		}
		if err := conn.WriteJSON(map[string]any{"type": "started", "session_id": "session-1"}); err != nil {
			t.Errorf("write started: %v", err)
			return
		}

		var audio map[string]any
		if err := conn.ReadJSON(&audio); err != nil {
			t.Errorf("read audio: %v", err)
			return
		}
		if audio["type"] != "audio" {
			t.Errorf("expected audio event, got %+v", audio)
			return
		}
		var flush map[string]any
		if err := conn.ReadJSON(&flush); err != nil {
			t.Errorf("read flush: %v", err)
			return
		}
		if flush["type"] != "flush" {
			t.Errorf("expected flush event, got %+v", flush)
			return
		}
		if err := conn.WriteJSON(map[string]any{
			"type": "partial", "session_id": "session-1", "sequence": 0,
			"words":         []map[string]any{{"word": "hello", "start": 0.0, "end": 0.5}},
			"processing_ms": 11,
		}); err != nil {
			t.Errorf("write partial: %v", err)
			return
		}
		if err := conn.WriteJSON(map[string]any{
			"type": "final_words", "session_id": "session-1", "up_to_time": 1.0,
			"words":         []map[string]any{{"word": "hello", "start": 0.0, "end": 0.5}},
			"processing_ms": 22,
		}); err != nil {
			t.Errorf("write final_words: %v", err)
			return
		}

		var stop map[string]any
		if err := conn.ReadJSON(&stop); err != nil {
			t.Errorf("read stop: %v", err)
			return
		}
		if stop["type"] != "stop" {
			t.Errorf("expected stop event, got %+v", stop)
			return
		}
		if err := conn.WriteJSON(map[string]any{"type": "stopped", "session_id": "session-1", "word_count": 1, "duration": 1.0}); err != nil {
			t.Errorf("write stopped: %v", err)
			return
		}
	}))
	defer server.Close()

	endpoint := strings.TrimPrefix(server.URL, "http://")
	client := NewWSLiveClient(endpoint)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	defer client.Close()
	if err := client.Start("session-1"); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if err := client.SendAudio(AudioChunk{Sequence: 0, Start: 0, Duration: 1.0, PCM16: []byte{0x00, 0x00}}); err != nil {
		t.Fatalf("SendAudio: %v", err)
	}
	if err := client.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}

	messages := make(chan StreamMessage)
	errCh := make(chan error, 1)
	go func() { errCh <- ReceiveResultEvents(ctx, client, messages) }()

	var gotPartial, gotFinal, gotStopped bool
	for msg := range messages {
		if msg.Event != nil && msg.Event.Type == TranscriptEventPartial {
			gotPartial = true
			if msg.Event.ProcessingMS != 11 {
				t.Fatalf("expected partial processing ms 11, got %+v", msg.Event)
			}
		}
		if msg.Event != nil && msg.Event.Type == TranscriptEventFinalWords {
			gotFinal = true
			if msg.Event.ProcessingMS != 22 || msg.Event.UpToTime != 1.0 {
				t.Fatalf("unexpected final event: %+v", msg.Event)
			}
			if err := client.Stop(); err != nil {
				t.Fatalf("Stop: %v", err)
			}
		}
		if msg.Stopped != nil {
			gotStopped = true
		}
	}
	if err := <-errCh; err != nil {
		t.Fatalf("ReceiveResultEvents: %v", err)
	}
	if !gotPartial || !gotFinal || !gotStopped {
		state, _ := json.Marshal(map[string]bool{"partial": gotPartial, "final": gotFinal, "stopped": gotStopped})
		t.Fatalf("missing expected stream events: %s", state)
	}
}
