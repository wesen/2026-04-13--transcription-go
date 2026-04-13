package live

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-go-golems/transcription-go/internal/output"
	"github.com/gorilla/websocket"
)

// WSLiveClient speaks the session-oriented WebSocket streaming API.
type WSLiveClient struct {
	endpoint string
	conn     *websocket.Conn
	mu       sync.Mutex
}

func NewWSLiveClient(endpoint string) *WSLiveClient {
	return &WSLiveClient{endpoint: endpoint}
}

func (c *WSLiveClient) Connect(ctx context.Context) error {
	dialer := websocket.Dialer{Proxy: http.ProxyFromEnvironment}
	conn, _, err := dialer.DialContext(ctx, fmt.Sprintf("ws://%s/transcribe/stream", c.endpoint), nil)
	if err != nil {
		return fmt.Errorf("dial websocket: %w", err)
	}
	c.conn = conn
	return nil
}

func (c *WSLiveClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn == nil {
		return nil
	}
	err := c.conn.Close()
	c.conn = nil
	return err
}

func (c *WSLiveClient) Start(sessionID string) error {
	return c.writeJSON(wsClientStartEvent{
		Type:       "start",
		SessionID:  sessionID,
		SampleRate: 16000,
		Channels:   1,
		Format:     "pcm_s16le",
		Source:     "replay_wav",
	})
}

func (c *WSLiveClient) SendAudio(chunk AudioChunk) error {
	return c.writeJSON(wsClientAudioEvent{
		Type:        "audio",
		Sequence:    chunk.Sequence,
		PTS:         chunk.Start,
		Duration:    chunk.Duration,
		PCM16Base64: base64.StdEncoding.EncodeToString(chunk.PCM16),
	})
}

func (c *WSLiveClient) Flush() error {
	return c.writeJSON(map[string]any{"type": "flush"})
}

func (c *WSLiveClient) Stop() error {
	return c.writeJSON(map[string]any{"type": "stop"})
}

func (c *WSLiveClient) ReadServerMessage() (*WSServerMessage, error) {
	if c.conn == nil {
		return nil, fmt.Errorf("websocket not connected")
	}
	var msg WSServerMessage
	if err := c.conn.ReadJSON(&msg); err != nil {
		return nil, fmt.Errorf("read websocket message: %w", err)
	}
	return &msg, nil
}

func (c *WSLiveClient) writeJSON(v any) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn == nil {
		return fmt.Errorf("websocket not connected")
	}
	_ = c.conn.SetWriteDeadline(time.Now().Add(30 * time.Second))
	if err := c.conn.WriteJSON(v); err != nil {
		return fmt.Errorf("write websocket message: %w", err)
	}
	return nil
}

type wsClientStartEvent struct {
	Type       string `json:"type"`
	SessionID  string `json:"session_id"`
	SampleRate int    `json:"sample_rate"`
	Channels   int    `json:"channels"`
	Format     string `json:"format"`
	Source     string `json:"source"`
}

type wsClientAudioEvent struct {
	Type        string  `json:"type"`
	Sequence    int     `json:"sequence"`
	PTS         float64 `json:"pts"`
	Duration    float64 `json:"duration"`
	PCM16Base64 string  `json:"pcm16_base64"`
}

// WSServerMessage is the raw JSON envelope received from the WS API.
type WSServerMessage struct {
	Type         string        `json:"type"`
	SessionID    string        `json:"session_id"`
	Sequence     int           `json:"sequence"`
	UpToTime     float64       `json:"up_to_time"`
	Words        []output.Word `json:"words"`
	Text         string        `json:"text"`
	WordCount    int           `json:"word_count"`
	Duration     float64       `json:"duration"`
	Code         string        `json:"code"`
	Message      string        `json:"message"`
	ProcessingMS int           `json:"processing_ms"`
}

func (m *WSServerMessage) String() string {
	data, _ := json.Marshal(m)
	return string(data)
}
