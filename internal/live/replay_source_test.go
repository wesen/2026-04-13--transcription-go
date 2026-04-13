package live

import (
	"context"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
)

func TestReplaySourceChunksWAVWithOverlap(t *testing.T) {
	t.Parallel()

	inputPath := filepath.Join(t.TempDir(), "input.wav")
	writeTestWAV(t, inputPath, []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13}, 4, 1, 16)
	tempDir := filepath.Join(t.TempDir(), "chunks")

	source := NewReplaySource(ReplaySourceConfig{
		InputPath:      inputPath,
		TempDir:        tempDir,
		SessionID:      "session-test",
		ChunkDuration:  1.5,
		OverlapSeconds: 0.5,
		ReplaySpeed:    0,
	})

	ch := make(chan AudioChunk)
	errCh := make(chan error, 1)
	go func() { errCh <- source.Run(context.Background(), ch) }()

	var got []AudioChunk
	for chunk := range ch {
		got = append(got, chunk)
	}
	if err := <-errCh; err != nil {
		t.Fatalf("replay source returned error: %v", err)
	}

	if len(got) != 3 {
		t.Fatalf("expected 3 chunks, got %d", len(got))
	}
	assertApprox(t, got[0].Start, 0.0)
	assertApprox(t, got[0].Duration, 1.5)
	assertApprox(t, got[1].Start, 1.0)
	assertApprox(t, got[1].Duration, 1.5)
	assertApprox(t, got[2].Start, 2.0)
	assertApprox(t, got[2].Duration, 1.25)

	for _, chunk := range got {
		if chunk.SessionID != "session-test" {
			t.Fatalf("unexpected session id: %+v", chunk)
		}
		if _, err := os.Stat(chunk.WAVPath); err != nil {
			t.Fatalf("chunk wav missing: %s (%v)", chunk.WAVPath, err)
		}
		_ = os.Remove(chunk.WAVPath)
	}
}

func writeTestWAV(t *testing.T, path string, samples []int, sampleRate, numChans, bitDepth int) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create test wav: %v", err)
	}
	defer f.Close()

	enc := wav.NewEncoder(f, sampleRate, bitDepth, numChans, 1)
	buf := &audio.IntBuffer{
		Data:   samples,
		Format: &audio.Format{NumChannels: numChans, SampleRate: sampleRate},
	}
	if err := enc.Write(buf); err != nil {
		t.Fatalf("write test wav: %v", err)
	}
	if err := enc.Close(); err != nil {
		t.Fatalf("close test wav: %v", err)
	}
}

func assertApprox(t *testing.T, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 0.0001 {
		t.Fatalf("got %.4f want %.4f", got, want)
	}
}
