// Package corpus implements a resumable, corpus-wide video transcription
// pipeline around the existing Go/Dagger/Nemotron transcription tool.
//
// The package owns manifest validation, corpus SQLite persistence, pipeline
// fingerprinting, resume planning, and a sequential warm-service runner. It
// treats raw word timestamps as canonical evidence and chunks/exports/search
// as derived projections that can be regenerated without ASR.
package corpus

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// chunkSourceType is the default chunk derivation source label.
const chunkSourceType = "punctuation-v1"

// SchemaManifest is the manifest schema version accepted by this package.
const SchemaManifest = "transcription-video-corpus/v1"

// Availability describes whether a source video can be transcribed.
type Availability string

const (
	AvailabilityAvailable   Availability = "available"
	AvailabilityMembersOnly Availability = "members_only"
	AvailabilityPrivate     Availability = "private"
	AvailabilityDeleted     Availability = "deleted"
	AvailabilityMissing     Availability = "missing"
	AvailabilityUnknown     Availability = "unknown"
)

// IsTranscribable reports whether an available audio file is expected.
func (a Availability) IsTranscribable() bool {
	return a == AvailabilityAvailable
}

// Manifest is the normalized, versioned corpus input.
type Manifest struct {
	Schema string         `json:"schema"`
	Corpus CorpusMetadata `json:"corpus"`
	Items  []ManifestItem `json:"items"`
}

// CorpusMetadata identifies the corpus.
type CorpusMetadata struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	SourceURL string `json:"source_url"`
}

// ManifestItem is one expected playlist entry, available or not.
type ManifestItem struct {
	SourceID        string          `json:"source_id"`
	Position        int             `json:"position"`
	Title           string          `json:"title"`
	SourceURL       string          `json:"source_url"`
	MediaPath       string          `json:"media_path,omitempty"`
	AudioPath       string          `json:"audio_path,omitempty"`
	MediaSHA256     string          `json:"media_sha256,omitempty"`
	AudioSHA256     string          `json:"audio_sha256,omitempty"`
	DurationSeconds float64         `json:"duration_seconds,omitempty"`
	Availability    Availability    `json:"availability"`
	Metadata        json.RawMessage `json:"metadata,omitempty"`
}

// LoadManifest reads and validates a manifest JSON file.
func LoadManifest(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read manifest: %w", err)
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse manifest: %w", err)
	}
	if err := m.Validate(); err != nil {
		return nil, err
	}
	return &m, nil
}

// Validate checks schema version, corpus identity, and item invariants.
func (m *Manifest) Validate() error {
	if m.Schema != SchemaManifest {
		return fmt.Errorf("unsupported manifest schema %q (want %q)", m.Schema, SchemaManifest)
	}
	if strings.TrimSpace(m.Corpus.ID) == "" {
		return fmt.Errorf("manifest corpus.id is required")
	}
	if strings.TrimSpace(m.Corpus.Title) == "" {
		return fmt.Errorf("manifest corpus.title is required")
	}
	if len(m.Items) == 0 {
		return fmt.Errorf("manifest has no items")
	}
	seenIDs := make(map[string]int, len(m.Items))
	seenPositions := make(map[int]int, len(m.Items))
	for i := range m.Items {
		item := &m.Items[i]
		if strings.TrimSpace(item.SourceID) == "" {
			return fmt.Errorf("item %d has empty source_id", i)
		}
		if item.Availability == "" {
			return fmt.Errorf("item %s has empty availability", item.SourceID)
		}
		if _, dup := seenIDs[item.SourceID]; dup {
			return fmt.Errorf("duplicate source_id %q", item.SourceID)
		}
		seenIDs[item.SourceID] = i
		if item.Position > 0 {
			if _, dup := seenPositions[item.Position]; dup {
				return fmt.Errorf("duplicate playlist position %d", item.Position)
			}
			seenPositions[item.Position] = i
		}
		if item.Availability.IsTranscribable() {
			if strings.TrimSpace(item.AudioPath) == "" {
				return fmt.Errorf("available item %s has no audio_path", item.SourceID)
			}
			if _, err := os.Stat(item.AudioPath); err != nil {
				return fmt.Errorf("available item %s audio_path: %w", item.SourceID, err)
			}
		}
		if item.MediaSHA256 != "" && !looksLikeSHA256(item.MediaSHA256) {
			return fmt.Errorf("item %s has malformed media_sha256", item.SourceID)
		}
		if item.AudioSHA256 != "" && !looksLikeSHA256(item.AudioSHA256) {
			return fmt.Errorf("item %s has malformed audio_sha256", item.SourceID)
		}
		if item.DurationSeconds < 0 {
			return fmt.Errorf("item %s has negative duration_seconds", item.SourceID)
		}
	}
	return nil
}

// FilterBySourceID returns items matching one of the given source IDs, or all
// items when no filter is supplied.
func (m *Manifest) FilterBySourceID(ids []string) []ManifestItem {
	if len(ids) == 0 {
		return m.Items
	}
	want := make(map[string]bool, len(ids))
	for _, id := range ids {
		want[id] = true
	}
	out := make([]ManifestItem, 0, len(m.Items))
	for _, item := range m.Items {
		if want[item.SourceID] {
			out = append(out, item)
		}
	}
	return out
}

func looksLikeSHA256(s string) bool {
	if len(s) != 64 {
		return false
	}
	for _, r := range s {
		switch {
		case r >= '0' && r <= '9':
		case r >= 'a' && r <= 'f':
		case r >= 'A' && r <= 'F':
		default:
			return false
		}
	}
	return true
}
