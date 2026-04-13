package live

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"time"
)

// MetricsSummary captures high-level performance characteristics of a live replay run.
type MetricsSummary struct {
	SessionID                 string  `json:"session_id"`
	ChunksProcessed           int     `json:"chunks_processed"`
	CommittedWords            int     `json:"committed_words"`
	EffectiveAudioSeconds     float64 `json:"effective_audio_seconds"`
	ElapsedSeconds            float64 `json:"elapsed_seconds"`
	AudioSecondsPerWallSecond float64 `json:"audio_seconds_per_wall_second"`
	AverageServerProcessingMS float64 `json:"average_server_processing_ms"`
	MaxServerProcessingMS     int     `json:"max_server_processing_ms"`
	AverageEndToEndMS         float64 `json:"average_end_to_end_ms"`
	MaxEndToEndMS             int64   `json:"max_end_to_end_ms"`
}

// MetricsCollector aggregates chunk-level performance stats.
type MetricsCollector struct {
	started                 time.Time
	chunksProcessed         int
	committedWords          int
	maxChunkEnd             float64
	serverProcessingTotalMS int
	maxServerProcessingMS   int
	endToEndTotalMS         int64
	maxEndToEndMS           int64
}

func NewMetricsCollector(started time.Time) *MetricsCollector {
	return &MetricsCollector{started: started}
}

func (m *MetricsCollector) ObserveChunk(chunk AudioChunk, committedWords int, serverProcessingMS int, endToEnd time.Duration) {
	m.chunksProcessed++
	m.committedWords = committedWords
	chunkEnd := chunk.Start + chunk.Duration
	if chunkEnd > m.maxChunkEnd {
		m.maxChunkEnd = chunkEnd
	}
	m.serverProcessingTotalMS += serverProcessingMS
	if serverProcessingMS > m.maxServerProcessingMS {
		m.maxServerProcessingMS = serverProcessingMS
	}
	endToEndMS := endToEnd.Milliseconds()
	m.endToEndTotalMS += endToEndMS
	if endToEndMS > m.maxEndToEndMS {
		m.maxEndToEndMS = endToEndMS
	}
}

func (m *MetricsCollector) Summary(sessionID string) MetricsSummary {
	elapsedSeconds := time.Since(m.started).Seconds()
	avgServer := 0.0
	avgEndToEnd := 0.0
	if m.chunksProcessed > 0 {
		avgServer = float64(m.serverProcessingTotalMS) / float64(m.chunksProcessed)
		avgEndToEnd = float64(m.endToEndTotalMS) / float64(m.chunksProcessed)
	}
	audioPerWall := 0.0
	if elapsedSeconds > 0 {
		audioPerWall = m.maxChunkEnd / elapsedSeconds
	}
	return MetricsSummary{
		SessionID:                 sessionID,
		ChunksProcessed:           m.chunksProcessed,
		CommittedWords:            m.committedWords,
		EffectiveAudioSeconds:     roundFloat(m.maxChunkEnd),
		ElapsedSeconds:            roundFloat(elapsedSeconds),
		AudioSecondsPerWallSecond: roundFloat(audioPerWall),
		AverageServerProcessingMS: roundFloat(avgServer),
		MaxServerProcessingMS:     m.maxServerProcessingMS,
		AverageEndToEndMS:         roundFloat(avgEndToEnd),
		MaxEndToEndMS:             m.maxEndToEndMS,
	}
}

func WriteMetricsSummary(outputDir string, summary MetricsSummary) error {
	if err := ensureOutputDir(outputDir); err != nil {
		return fmt.Errorf("ensure metrics dir: %w", err)
	}
	path := filepath.Join(outputDir, "live-summary.json")
	data, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal metrics summary: %w", err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("write metrics summary: %w", err)
	}
	return nil
}

func roundFloat(v float64) float64 {
	return math.Round(v*1000) / 1000
}
