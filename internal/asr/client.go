// Package asr provides an HTTP client for the transcription ASR server.
package asr

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

// Word represents a transcribed word with timing.
type Word struct {
	Word  string  `json:"word"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
}

// TranscribeResponse is the JSON response from POST /transcribe/full.
type TranscribeResponse struct {
	Words         []Word  `json:"words"`
	TotalDuration float64 `json:"total_duration"`
	ChunkCount    int     `json:"chunk_count"`
	WordCount     int     `json:"word_count"`
}

// Client is an HTTP client for the ASR server.
type Client struct {
	endpoint string
	client   *http.Client
}

// NewClient creates a new ASR client pointing at the given host:port endpoint.
func NewClient(endpoint string) *Client {
	return &Client{
		endpoint: endpoint,
		client:   &http.Client{Timeout: 0}, // no timeout — transcription can take minutes
	}
}

// Health checks the server health endpoint.
func (c *Client) Health(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("http://%s/health", c.endpoint), nil)
	if err != nil {
		return err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed: %d", resp.StatusCode)
	}
	return nil
}

// TranscribeFull sends a full audio file to the server for transcription.
// The server handles chunking internally.
func (c *Client) TranscribeFull(ctx context.Context, audioPath string, chunkSize int) (*TranscribeResponse, error) {
	f, err := os.Open(audioPath)
	if err != nil {
		return nil, fmt.Errorf("open audio: %w", err)
	}
	defer f.Close()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("file", filepath.Base(audioPath))
	if err != nil {
		return nil, fmt.Errorf("create form file: %w", err)
	}
	if _, err := io.Copy(part, f); err != nil {
		return nil, fmt.Errorf("copy audio to form: %w", err)
	}

	if err := writer.WriteField("chunk_size", fmt.Sprintf("%d", chunkSize)); err != nil {
		return nil, fmt.Errorf("write chunk_size field: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("close multipart writer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("http://%s/transcribe/full", c.endpoint), &buf)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server error %d: %s", resp.StatusCode, string(body))
	}

	var result TranscribeResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}
