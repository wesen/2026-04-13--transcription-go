package live

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestMetricsCollectorSummary(t *testing.T) {
	t.Parallel()
	started := time.Now().Add(-10 * time.Second)
	m := NewMetricsCollector(started)
	m.ObserveChunk(AudioChunk{Start: 0.0, Duration: 5.0}, 10, 1200, 1500*time.Millisecond)
	m.ObserveChunk(AudioChunk{Start: 4.5, Duration: 5.0}, 20, 1800, 2500*time.Millisecond)

	summary := m.Summary("session-1")
	if summary.ChunksProcessed != 2 {
		t.Fatalf("expected 2 chunks, got %+v", summary)
	}
	if summary.CommittedWords != 20 {
		t.Fatalf("expected 20 committed words, got %+v", summary)
	}
	if summary.MaxServerProcessingMS != 1800 {
		t.Fatalf("expected max server processing 1800, got %+v", summary)
	}
	if summary.MaxEndToEndMS != 2500 {
		t.Fatalf("expected max end-to-end 2500, got %+v", summary)
	}
	if summary.EffectiveAudioSeconds != 9.5 {
		t.Fatalf("expected effective audio seconds 9.5, got %+v", summary)
	}
}

func TestWriteMetricsSummary(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	summary := MetricsSummary{SessionID: "abc", ChunksProcessed: 3}
	if err := WriteMetricsSummary(dir, summary); err != nil {
		t.Fatalf("WriteMetricsSummary: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "live-summary.json"))
	if err != nil {
		t.Fatalf("read live-summary.json: %v", err)
	}
	var decoded MetricsSummary
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal live-summary.json: %v", err)
	}
	if decoded.SessionID != "abc" || decoded.ChunksProcessed != 3 {
		t.Fatalf("unexpected decoded summary: %+v", decoded)
	}
}
