package asr

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTranscribeChunk(t *testing.T) {
	t.Parallel()

	var seen struct {
		SessionID      string
		ChunkIndex     string
		ChunkStart     string
		OverlapSeconds string
		IsFinalChunk   string
		FileName       string
		FileContents   string
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/transcribe/chunk" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Fatalf("parse multipart form: %v", err)
		}

		seen.SessionID = r.FormValue("session_id")
		seen.ChunkIndex = r.FormValue("chunk_index")
		seen.ChunkStart = r.FormValue("chunk_start")
		seen.OverlapSeconds = r.FormValue("overlap_seconds")
		seen.IsFinalChunk = r.FormValue("is_final_chunk")

		file, header, err := r.FormFile("file")
		if err != nil {
			t.Fatalf("form file: %v", err)
		}
		defer file.Close()
		seen.FileName = header.Filename
		body, err := io.ReadAll(file)
		if err != nil {
			t.Fatalf("read uploaded file: %v", err)
		}
		seen.FileContents = string(body)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(ChunkResponse{
			SessionID:     seen.SessionID,
			ChunkIndex:    7,
			ChunkStart:    12.5,
			ChunkDuration: 1.5,
			Words:         []Word{{Word: "hello", Start: 12.5, End: 12.8}},
			Partial:       false,
			ProcessingMS:  187,
		})
	}))
	defer ts.Close()

	chunkPath := writeTempChunk(t, "chunk-0007.wav", "fake wav payload")
	client := testClient(ts.URL, ts.Client())

	resp, err := client.TranscribeChunk(context.Background(), chunkPath, ChunkRequest{
		SessionID:      "session-123",
		ChunkIndex:     7,
		ChunkStart:     12.5,
		OverlapSeconds: 0.25,
		IsFinalChunk:   true,
	})
	if err != nil {
		t.Fatalf("TranscribeChunk returned error: %v", err)
	}

	if seen.SessionID != "session-123" || seen.ChunkIndex != "7" || seen.ChunkStart != "12.5" || seen.OverlapSeconds != "0.25" || seen.IsFinalChunk != "true" {
		t.Fatalf("unexpected form fields: %+v", seen)
	}
	if seen.FileName != filepath.Base(chunkPath) {
		t.Fatalf("unexpected uploaded file name: %s", seen.FileName)
	}
	if seen.FileContents != "fake wav payload" {
		t.Fatalf("unexpected uploaded payload: %q", seen.FileContents)
	}
	if resp.SessionID != "session-123" || resp.ChunkIndex != 7 || resp.ChunkDuration != 1.5 || len(resp.Words) != 1 {
		t.Fatalf("unexpected chunk response: %+v", resp)
	}
}

func TestTranscribeFull(t *testing.T) {
	t.Parallel()

	var seenChunkSize string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/transcribe/full" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Fatalf("parse multipart form: %v", err)
		}
		seenChunkSize = r.FormValue("chunk_size")
		if _, _, err := r.FormFile("file"); err != nil {
			t.Fatalf("missing file upload: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(TranscribeResponse{
			Words:         []Word{{Word: "test", Start: 0.0, End: 0.3}},
			TotalDuration: 2.0,
			ChunkCount:    1,
			WordCount:     1,
		})
	}))
	defer ts.Close()

	audioPath := writeTempChunk(t, "audio.wav", "fake wav payload")
	client := testClient(ts.URL, ts.Client())

	resp, err := client.TranscribeFull(context.Background(), audioPath, 42)
	if err != nil {
		t.Fatalf("TranscribeFull returned error: %v", err)
	}
	if seenChunkSize != "42" {
		t.Fatalf("unexpected chunk_size field: %s", seenChunkSize)
	}
	if resp.WordCount != 1 || len(resp.Words) != 1 {
		t.Fatalf("unexpected full response: %+v", resp)
	}
}

func testClient(serverURL string, httpClient *http.Client) *Client {
	endpoint := strings.TrimPrefix(serverURL, "http://")
	endpoint = strings.TrimPrefix(endpoint, "https://")
	return &Client{endpoint: endpoint, client: httpClient}
}

func writeTempChunk(t *testing.T, name, contents string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("write temp chunk: %v", err)
	}
	return path
}
