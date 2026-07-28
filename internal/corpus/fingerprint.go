package corpus

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

// SchemaFingerprint is the pipeline fingerprint schema version.
const SchemaFingerprint = "transcription-pipeline-fingerprint/v1"

// AudioContract describes the normalized audio input contract.
const AudioContract = "wav-pcm-s16le-mono-16000/v1"

// WordSchema describes the returned word timestamp shape.
const WordSchema = "word-timestamps/v1"

// ModelName is the Nemotron model used by the reference service.
const ModelName = "nvidia/nemotron-speech-streaming-en-0.6b"

// Fingerprint captures the settings that change transcript output.
type Fingerprint struct {
	Schema                string              `json:"schema"`
	Model                 string              `json:"model"`
	ModelRevision         string              `json:"model_revision"`
	ServerRequirementsSHA string              `json:"server_requirements_sha256,omitempty"`
	Decoding              DecodingFingerprint `json:"decoding"`
	ChunkSizeSeconds      int                 `json:"chunk_size_seconds"`
	ChunkOverlapSeconds   int                 `json:"chunk_overlap_seconds"`
	AudioContract         string              `json:"audio_contract"`
	WordSchema            string              `json:"word_schema"`
}

// DecodingFingerprint records Nemotron decoding configuration.
type DecodingFingerprint struct {
	PreserveAlignments bool     `json:"preserve_alignments"`
	ComputeTimestamps  bool     `json:"compute_timestamps"`
	SegmentSeparators  []string `json:"segment_separators"`
	WordSeparator      string   `json:"word_separator"`
}

// DefaultFingerprint returns the fingerprint matching the current server.py
// configuration and the default 60s/2s chunking policy.
func DefaultFingerprint(chunkSize int) Fingerprint {
	if chunkSize <= 0 {
		chunkSize = 60
	}
	return Fingerprint{
		Schema:        SchemaFingerprint,
		Model:         ModelName,
		ModelRevision: "unpinned",
		Decoding: DecodingFingerprint{
			PreserveAlignments: true,
			ComputeTimestamps:  true,
			SegmentSeparators:  []string{},
			WordSeparator:      " ",
		},
		ChunkSizeSeconds:    chunkSize,
		ChunkOverlapSeconds: 2,
		AudioContract:       AudioContract,
		WordSchema:          WordSchema,
	}
}

// String returns the canonical JSON used to compute the fingerprint hash.
func (f Fingerprint) String() string {
	b, _ := json.Marshal(f.canonical())
	return string(b)
}

// Hash returns the SHA-256 hex digest of the canonical fingerprint JSON.
func (f Fingerprint) Hash() string {
	b := f.canonical()
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// ShortHash returns the first 16 hex characters of Hash.
func (f Fingerprint) ShortHash() string {
	h := f.Hash()
	if len(h) < 16 {
		return h
	}
	return h[:16]
}

func (f Fingerprint) canonical() []byte {
	// Marshal with sorted keys by using json.Marshal of a map-free struct and
	// then re-marshalling into a canonical form. For this application, struct
	// field order is stable and sufficient.
	out := map[string]any{
		"schema":                     f.Schema,
		"model":                      f.Model,
		"model_revision":             f.ModelRevision,
		"server_requirements_sha256": f.ServerRequirementsSHA,
		"decoding": map[string]any{
			"preserve_alignments": f.Decoding.PreserveAlignments,
			"compute_timestamps":  f.Decoding.ComputeTimestamps,
			"segment_separators":  f.Decoding.SegmentSeparators,
			"word_separator":      f.Decoding.WordSeparator,
		},
		"chunk_size_seconds":    f.ChunkSizeSeconds,
		"chunk_overlap_seconds": f.ChunkOverlapSeconds,
		"audio_contract":        f.AudioContract,
		"word_schema":           f.WordSchema,
	}
	b, err := json.Marshal(out)
	if err != nil {
		// canonicalize returns a stable fallback if marshalling ever fails.
		return []byte(fmt.Sprintf("%v", out))
	}
	return b
}
