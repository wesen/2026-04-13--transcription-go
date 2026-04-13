// Package asr provides HTTP clients for the transcription ASR server.
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
	"strconv"
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

// ChunkRequest describes metadata sent with POST /transcribe/chunk.
type ChunkRequest struct {
	SessionID      string
	ChunkIndex     int
	ChunkStart     float64
	OverlapSeconds float64
	IsFinalChunk   bool
}

// ChunkResponse is the JSON response from POST /transcribe/chunk.
type ChunkResponse struct {
	SessionID     string  `json:"session_id"`
	ChunkIndex    int     `json:"chunk_index"`
	ChunkStart    float64 `json:"chunk_start"`
	ChunkDuration float64 `json:"chunk_duration"`
	Words         []Word  `json:"words"`
	Partial       bool    `json:"partial"`
	ProcessingMS  int     `json:"processing_ms"`
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
	resp, err := c.uploadAudio(ctx, "/transcribe/full", audioPath, func(writer *multipart.Writer) error {
		return writer.WriteField("chunk_size", strconv.Itoa(chunkSize))
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result TranscribeResponse
	if err := decodeJSONResponse(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// TranscribeChunk sends a single chunk to the live/near-live chunk endpoint.
func (c *Client) TranscribeChunk(ctx context.Context, audioPath string, req ChunkRequest) (*ChunkResponse, error) {
	resp, err := c.uploadAudio(ctx, "/transcribe/chunk", audioPath, func(writer *multipart.Writer) error {
		if err := writer.WriteField("session_id", req.SessionID); err != nil {
			return err
		}
		if err := writer.WriteField("chunk_index", strconv.Itoa(req.ChunkIndex)); err != nil {
			return err
		}
		if err := writer.WriteField("chunk_start", strconv.FormatFloat(req.ChunkStart, 'f', -1, 64)); err != nil {
			return err
		}
		if err := writer.WriteField("overlap_seconds", strconv.FormatFloat(req.OverlapSeconds, 'f', -1, 64)); err != nil {
			return err
		}
		if err := writer.WriteField("is_final_chunk", strconv.FormatBool(req.IsFinalChunk)); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result ChunkResponse
	if err := decodeJSONResponse(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) uploadAudio(ctx context.Context, endpointPath, audioPath string, writeFields func(*multipart.Writer) error) (*http.Response, error) {
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
	if writeFields != nil {
		if err := writeFields(writer); err != nil {
			return nil, fmt.Errorf("write form fields: %w", err)
		}
	}
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("close multipart writer: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("http://%s%s", c.endpoint, endpointPath), &buf)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	return resp, nil
}

func decodeJSONResponse(resp *http.Response, out any) error {
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server error %d: %s", resp.StatusCode, string(body))
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}
