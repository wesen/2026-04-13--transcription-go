package corpus

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// ProcessingState is the transcript lifecycle state of a video.
type ProcessingState string

const (
	StatePending      ProcessingState = "pending"
	StateReady        ProcessingState = "ready"
	StateTranscribing ProcessingState = "transcribing"
	StateComplete     ProcessingState = "complete"
	StateFailed       ProcessingState = "failed"
	StateStale        ProcessingState = "stale"
	StateUnavailable  ProcessingState = "unavailable"
)

// AttemptState is the lifecycle state of one transcription attempt.
type AttemptState string

const (
	AttemptRunning   AttemptState = "running"
	AttemptSucceeded AttemptState = "succeeded"
	AttemptFailed    AttemptState = "failed"
	AttemptAbandoned AttemptState = "abandoned"
)

// Word is a canonical timed word produced by ASR.
type Word struct {
	Text             string
	NormalizedText   string
	Start            float64
	End              float64
	Confidence       float64
	SourceChunkIndex int
	IsFiller         bool
}

// Transcription is the ASR result for one video.
type Transcription struct {
	Words           []Word
	DurationSeconds float64
	ChunkCount      int
	ProcessingTime  time.Duration
}

// Video is a stored video row needed by the runner and exports.
type Video struct {
	ID               int64
	CorpusID         int64
	SourceID         string
	PlaylistPosition int
	Title            string
	SourceURL        string
	Availability     Availability
	MediaPath        string
	AudioPath        string
	MediaSHA256      string
	AudioSHA256      string
	DurationSeconds  float64
	MetadataJSON     string
	ProcessingState  ProcessingState
	ActiveRevisionID sql.NullInt64
	LastError        sql.NullString
}

// WorkItem is a planned unit of work for the runner.
type WorkItem struct {
	Video           Video
	Reason          string // "pending", "stale", "failed", "export-only"
	NeedsTranscribe bool
	NeedsExport     bool
}

// Attempt is a started transcription attempt.
type Attempt struct {
	ID            int64
	VideoID       int64
	AttemptNumber int
	State         AttemptState
}

// Revision is a committed transcript revision.
type Revision struct {
	ID                  int64
	VideoID             int64
	SourceSHA256        string
	PipelineFingerprint string
	ModelName           string
	DurationSeconds     float64
	WordCount           int
	ChunkCount          int
}

// RevisionData is a revision plus its words and chunks for export/search.
type RevisionData struct {
	Revision Revision
	Words    []Word
	Chunks   []Chunk
}

// Chunk is a derived sentence/display fragment.
type Chunk struct {
	ID               int64
	RevisionID       int64
	Ordinal          int
	StartWordOrdinal int
	EndWordOrdinal   int
	Start            float64
	End              float64
	Text             string
	WordCount        int
	SourceType       string
	PolicyJSON       string
}

// ChunkPolicy describes how chunks were derived.
type ChunkPolicy struct {
	Schema             string   `json:"schema"`
	SplitOn            []string `json:"split_on"`
	MaxDurationSeconds float64  `json:"max_duration_seconds"`
	MaxCharacters      int      `json:"max_characters"`
	IncludeRemoved     bool     `json:"include_removed"`
}

// DefaultChunkPolicy matches output.BuildSegments defaults.
func DefaultChunkPolicy() ChunkPolicy {
	return ChunkPolicy{
		Schema:             "transcript-chunk-policy/v1",
		SplitOn:            []string{".", "?", "!"},
		MaxDurationSeconds: 15.0,
		MaxCharacters:      120,
		IncludeRemoved:     false,
	}
}

// Store is the corpus SQLite persistence layer.
type Store struct {
	db *sql.DB
}

// OpenStore opens or creates a corpus database at path and applies the schema.
func OpenStore(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create database dir: %w", err)
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open corpus db: %w", err)
	}
	if _, err := db.Exec(schemaSQL); err != nil {
		db.Close()
		return nil, fmt.Errorf("apply schema: %w", err)
	}
	return &Store{db: db}, nil
}

// Close closes the database.
func (s *Store) Close() error { return s.db.Close() }

// IntegrityCheck runs PRAGMA integrity_check and foreign_key_check.
func (s *Store) IntegrityCheck(ctx context.Context) error {
	for _, q := range []string{"PRAGMA integrity_check", "PRAGMA foreign_key_check"} {
		row := s.db.QueryRowContext(ctx, q)
		var v string
		if err := row.Scan(&v); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				continue
			}
			return fmt.Errorf("%s: %w", q, err)
		}
		if v != "ok" && v != "" {
			return fmt.Errorf("%s: %s", q, v)
		}
	}
	return nil
}

// ImportManifest upserts the corpus and its videos. Existing active revisions
// are preserved. Unavailable items are recorded with processing_state=unavailable.
func (s *Store) ImportManifest(ctx context.Context, m *Manifest) (corpusID int64, err error) {
	manifestHash := manifestSHA256(m)
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() { rollbackOnErr(tx, err) }()

	corpusID, err = upsertCorpus(ctx, tx, m, manifestHash)
	if err != nil {
		return 0, err
	}
	for i := range m.Items {
		if err := upsertVideo(ctx, tx, corpusID, &m.Items[i]); err != nil {
			return 0, err
		}
	}
	if _, err := tx.ExecContext(ctx, `UPDATE corpora SET updated_at = CURRENT_TIMESTAMP WHERE id = ?`, corpusID); err != nil {
		return 0, err
	}
	return corpusID, tx.Commit()
}

func upsertCorpus(ctx context.Context, tx *sql.Tx, m *Manifest, manifestHash string) (int64, error) {
	var id int64
	err := tx.QueryRowContext(ctx, `SELECT id FROM corpora WHERE corpus_key = ?`, m.Corpus.ID).Scan(&id)
	if err == nil {
		if _, err := tx.ExecContext(ctx, `UPDATE corpora SET title=?, source_url=?, manifest_schema=?, manifest_sha256=?, updated_at=CURRENT_TIMESTAMP WHERE id=?`,
			m.Corpus.Title, m.Corpus.SourceURL, m.Schema, manifestHash, id); err != nil {
			return 0, err
		}
		return id, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}
	res, err := tx.ExecContext(ctx, `INSERT INTO corpora (corpus_key, title, source_url, manifest_schema, manifest_sha256) VALUES (?, ?, ?, ?, ?)`,
		m.Corpus.ID, m.Corpus.Title, m.Corpus.SourceURL, m.Schema, manifestHash)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func upsertVideo(ctx context.Context, tx *sql.Tx, corpusID int64, item *ManifestItem) error {
	metadata := ""
	if len(item.Metadata) > 0 {
		metadata = string(item.Metadata)
	}
	state := StatePending
	if !item.Availability.IsTranscribable() {
		state = StateUnavailable
	}
	row := tx.QueryRowContext(ctx, `SELECT id, processing_state, active_revision_id FROM videos WHERE corpus_id = ? AND source_id = ?`, corpusID, item.SourceID)
	var (
		existingID       int64
		existingState    string
		existingRevision sql.NullInt64
	)
	err := row.Scan(&existingID, &existingState, &existingRevision)
	if err == nil {
		// Preserve active revision; reset state for available items so planning
		// can decide stale/complete. Unavailable items stay unavailable.
		newState := state
		if state == StatePending && existingRevision.Valid {
			// Keep complete/failed/stale as-is so planning can compare; if the
			// source changed, Plan will mark it stale.
			switch ProcessingState(existingState) {
			case StateComplete, StateFailed, StateStale:
				newState = ProcessingState(existingState)
			}
		}
		if _, err := tx.ExecContext(ctx, `UPDATE videos SET
			playlist_position=?, title=?, source_url=?, availability=?,
			media_path=?, audio_path=?, media_sha256=?, audio_sha256=?,
			duration_seconds=?, metadata_json=?, processing_state=?, updated_at=CURRENT_TIMESTAMP
			WHERE id=?`,
			item.Position, item.Title, item.SourceURL, string(item.Availability),
			nullable(item.MediaPath), nullable(item.AudioPath), nullable(item.MediaSHA256), nullable(item.AudioSHA256),
			item.DurationSeconds, nullable(metadata), string(newState), existingID); err != nil {
			return err
		}
		return nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO videos
		(corpus_id, source_id, playlist_position, title, source_url, availability,
		 media_path, audio_path, media_sha256, audio_sha256, duration_seconds,
		 metadata_json, processing_state)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		corpusID, item.SourceID, item.Position, item.Title, item.SourceURL, string(item.Availability),
		nullable(item.MediaPath), nullable(item.AudioPath), nullable(item.MediaSHA256), nullable(item.AudioSHA256),
		item.DurationSeconds, nullable(metadata), string(state))
	return err
}

// Plan computes work items for the given corpus and fingerprint.
func (s *Store) Plan(ctx context.Context, corpusKey string, fp Fingerprint, retryFailed bool, sourceIDs []string) ([]WorkItem, error) {
	corpusID, err := s.corpusID(ctx, corpusKey)
	if err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT
		id, source_id, playlist_position, title, source_url, availability,
		media_path, audio_path, media_sha256, audio_sha256, duration_seconds,
		COALESCE(metadata_json, ''), processing_state, active_revision_id, last_error
		FROM videos WHERE corpus_id = ? ORDER BY playlist_position, source_id`, corpusID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	want := make(map[string]bool, len(sourceIDs))
	for _, id := range sourceIDs {
		want[id] = true
	}

	var items []WorkItem
	for rows.Next() {
		v, err := scanVideo(rows)
		if err != nil {
			return nil, err
		}
		if len(want) > 0 && !want[v.SourceID] {
			continue
		}
		item := WorkItem{Video: v}
		switch {
		case !v.Availability.IsTranscribable():
			item.Reason = "unavailable"
		case !v.ActiveRevisionID.Valid:
			if v.ProcessingState == StateFailed && !retryFailed {
				item.Reason = "failed"
			} else {
				item.Reason = "pending"
				item.NeedsTranscribe = true
			}
		default:
			rev, err := s.activeRevision(ctx, v.ActiveRevisionID.Int64)
			if err != nil {
				return nil, err
			}
			stale := rev.SourceSHA256 != effectiveHash(v.AudioSHA256, v.MediaSHA256) ||
				rev.PipelineFingerprint != fp.Hash()
			if stale {
				item.Reason = "stale"
				item.NeedsTranscribe = true
			} else {
				item.Reason = "unchanged"
				item.NeedsExport = true
			}
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// ReconcileAbandoned marks running attempts older than a heartbeat as abandoned.
// With no process coordination in Phase 1, all running attempts are abandoned.
func (s *Store) ReconcileAbandoned(ctx context.Context) (int, error) {
	res, err := s.db.ExecContext(ctx, `UPDATE transcript_attempts SET state = ?, finished_at = CURRENT_TIMESTAMP WHERE state = ?`,
		string(AttemptAbandoned), string(AttemptRunning))
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return int(n), nil
}

// BeginAttempt creates a running attempt row for the video.
func (s *Store) BeginAttempt(ctx context.Context, v Video, fp Fingerprint, endpoint string) (*Attempt, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { rollbackOnErr(tx, err) }()

	var attemptNumber int
	if err := tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(attempt_number), 0) + 1 FROM transcript_attempts WHERE video_id = ?`, v.ID).Scan(&attemptNumber); err != nil {
		return nil, err
	}
	sourceHash := effectiveHash(v.AudioSHA256, v.MediaSHA256)
	res, err := tx.ExecContext(ctx, `INSERT INTO transcript_attempts
		(video_id, attempt_number, state, source_sha256, pipeline_fingerprint, server_endpoint)
		VALUES (?, ?, ?, ?, ?, ?)`,
		v.ID, attemptNumber, string(AttemptRunning), sourceHash, fp.Hash(), nullable(endpoint))
	if err != nil {
		return nil, err
	}
	attemptID, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE videos SET processing_state = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		string(StateTranscribing), v.ID); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &Attempt{ID: attemptID, VideoID: v.ID, AttemptNumber: attemptNumber, State: AttemptRunning}, nil
}

// CommitTranscript atomically inserts a revision, words, chunks, links, FTS,
// marks the attempt succeeded, and switches the video's active revision.
func (s *Store) CommitTranscript(ctx context.Context, v Video, attempt *Attempt, fp Fingerprint, result Transcription, policy ChunkPolicy) (*Revision, error) {
	if err := validateTranscription(result); err != nil {
		s.recordFailure(ctx, attempt, "validation", err.Error(), 0, 0, 0)
		return nil, fmt.Errorf("validate transcription: %w", err)
	}
	policyJSON, _ := json.Marshal(policy)
	chunks := deriveChunks(result.Words, policy)
	sourceHash := effectiveHash(v.AudioSHA256, v.MediaSHA256)
	chunkSourceType := chunkSourceType

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { rollbackOnErr(tx, err) }()

	res, err := tx.ExecContext(ctx, `INSERT INTO transcript_revisions
		(video_id, attempt_id, source_sha256, pipeline_fingerprint, model_name,
		 duration_seconds, word_count, chunk_count)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		v.ID, attempt.ID, sourceHash, fp.Hash(), fp.Model,
		result.DurationSeconds, len(result.Words), len(chunks))
	if err != nil {
		return nil, err
	}
	revisionID, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	wordStmt, err := tx.PrepareContext(ctx, `INSERT INTO words
		(revision_id, ordinal, text, normalized_text, start_time, end_time,
		 confidence, source_chunk_index, is_filler, is_removed)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return nil, err
	}
	defer wordStmt.Close()

	wordIDs := make([]int64, len(result.Words))
	for i, w := range result.Words {
		isFiller := 0
		if w.IsFiller {
			isFiller = 1
		}
		wres, err := wordStmt.ExecContext(ctx, revisionID, i, w.Text, w.NormalizedText,
			w.Start, w.End, nullableConf(w.Confidence), w.SourceChunkIndex, isFiller, 0)
		if err != nil {
			return nil, fmt.Errorf("insert word %d: %w", i, err)
		}
		id, err := wres.LastInsertId()
		if err != nil {
			return nil, err
		}
		wordIDs[i] = id
	}

	chunkStmt, err := tx.PrepareContext(ctx, `INSERT INTO chunks
		(revision_id, ordinal, start_word_ordinal, end_word_ordinal, start_time, end_time,
		 text, word_count, source_type, policy_json)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return nil, err
	}
	defer chunkStmt.Close()

	linkStmt, err := tx.PrepareContext(ctx, `INSERT INTO chunk_words (chunk_id, word_id, position) VALUES (?, ?, ?)`)
	if err != nil {
		return nil, err
	}
	defer linkStmt.Close()

	for ci, c := range chunks {
		cres, err := chunkStmt.ExecContext(ctx, revisionID, c.Ordinal, c.StartWordOrdinal, c.EndWordOrdinal,
			c.Start, c.End, c.Text, len(c.WordOrdinals), chunkSourceType, string(policyJSON))
		if err != nil {
			return nil, fmt.Errorf("insert chunk %d: %w", ci, err)
		}
		chunkID, err := cres.LastInsertId()
		if err != nil {
			return nil, err
		}
		for pos, wo := range c.WordOrdinals {
			if _, err := linkStmt.ExecContext(ctx, chunkID, wordIDs[wo], pos); err != nil {
				return nil, fmt.Errorf("link chunk %d word %d: %w", ci, wo, err)
			}
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO chunk_fts (rowid, text) VALUES (?, ?)`, chunkID, c.Text); err != nil {
			return nil, fmt.Errorf("insert chunk fts %d: %w", ci, err)
		}
	}

	if _, err := tx.ExecContext(ctx, `UPDATE transcript_attempts SET state = ?, finished_at = CURRENT_TIMESTAMP,
		processing_ms = ?, chunk_count = ?, word_count = ? WHERE id = ?`,
		string(AttemptSucceeded), result.ProcessingTime.Milliseconds(), result.ChunkCount, len(result.Words), attempt.ID); err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE videos SET processing_state = ?, active_revision_id = ?, last_error = NULL, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		string(StateComplete), revisionID, v.ID); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &Revision{
		ID:                  revisionID,
		VideoID:             v.ID,
		SourceSHA256:        sourceHash,
		PipelineFingerprint: fp.Hash(),
		ModelName:           fp.Model,
		DurationSeconds:     result.DurationSeconds,
		WordCount:           len(result.Words),
		ChunkCount:          len(chunks),
	}, nil
}

// FailAttempt marks an attempt failed and records the error on the video.
func (s *Store) FailAttempt(ctx context.Context, attempt *Attempt, errKind string, errMsg string, processingMs int64, chunkCount, wordCount int) error {
	return s.recordFailure(ctx, attempt, errKind, errMsg, processingMs, chunkCount, wordCount)
}

func (s *Store) recordFailure(ctx context.Context, attempt *Attempt, errKind, errMsg string, processingMs int64, chunkCount, wordCount int) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { rollbackOnErr(tx, err) }()
	if _, err := tx.ExecContext(ctx, `UPDATE transcript_attempts SET state = ?, finished_at = CURRENT_TIMESTAMP,
		processing_ms = ?, chunk_count = ?, word_count = ?, error_class = ?, error_message = ? WHERE id = ?`,
		string(AttemptFailed), processingMs, chunkCount, wordCount, errKind, errMsg, attempt.ID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE videos SET processing_state = ?, last_error = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		string(StateFailed), errMsg, attempt.VideoID); err != nil {
		return err
	}
	return tx.Commit()
}

// LoadRevision returns a revision's words and chunks.
func (s *Store) LoadRevision(ctx context.Context, revisionID int64) (*RevisionData, error) {
	rev, err := s.revisionByID(ctx, revisionID)
	if err != nil {
		return nil, err
	}
	words, err := s.revisionWords(ctx, revisionID)
	if err != nil {
		return nil, err
	}
	chunks, err := s.revisionChunks(ctx, revisionID)
	if err != nil {
		return nil, err
	}
	return &RevisionData{Revision: *rev, Words: words, Chunks: chunks}, nil
}

// VideoBySourceID returns a stored video.
func (s *Store) VideoBySourceID(ctx context.Context, corpusKey, sourceID string) (*Video, error) {
	corpusID, err := s.corpusID(ctx, corpusKey)
	if err != nil {
		return nil, err
	}
	row := s.db.QueryRowContext(ctx, `SELECT
		id, source_id, playlist_position, title, source_url, availability,
		media_path, audio_path, media_sha256, audio_sha256, duration_seconds,
		COALESCE(metadata_json, ''), processing_state, active_revision_id, last_error
		FROM videos WHERE corpus_id = ? AND source_id = ?`, corpusID, sourceID)
	return scanVideoRow(row)
}

// StatusCounts summarizes corpus state.
type StatusCounts struct {
	Total        int
	Available    int
	Unavailable  int
	Pending      int
	Transcribing int
	Complete     int
	Failed       int
	Stale        int
}

// Status returns counts for a corpus.
func (s *Store) Status(ctx context.Context, corpusKey string) (*StatusCounts, error) {
	corpusID, err := s.corpusID(ctx, corpusKey)
	if err != nil {
		return nil, err
	}
	c := &StatusCounts{}
	type counter struct {
		state string
		count *int
	}
	counters := []counter{
		{"pending", &c.Pending}, {"transcribing", &c.Transcribing},
		{"complete", &c.Complete}, {"failed", &c.Failed}, {"stale", &c.Stale},
	}
	for _, ct := range counters {
		if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM videos WHERE corpus_id = ? AND processing_state = ?`,
			corpusID, ct.state).Scan(ct.count); err != nil {
			return nil, err
		}
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM videos WHERE corpus_id = ?`, corpusID).Scan(&c.Total); err != nil {
		return nil, err
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM videos WHERE corpus_id = ? AND availability = 'available'`, corpusID).Scan(&c.Available); err != nil {
		return nil, err
	}
	c.Unavailable = c.Total - c.Available
	return c, nil
}

func (s *Store) corpusID(ctx context.Context, corpusKey string) (int64, error) {
	var id int64
	if err := s.db.QueryRowContext(ctx, `SELECT id FROM corpora WHERE corpus_key = ?`, corpusKey).Scan(&id); err != nil {
		return 0, fmt.Errorf("corpus %q not found: %w", corpusKey, err)
	}
	return id, nil
}

func (s *Store) activeRevision(ctx context.Context, id int64) (*Revision, error) {
	return s.revisionByID(ctx, id)
}

func (s *Store) revisionByID(ctx context.Context, id int64) (*Revision, error) {
	var r Revision
	err := s.db.QueryRowContext(ctx, `SELECT id, video_id, source_sha256, pipeline_fingerprint, model_name,
		duration_seconds, word_count, chunk_count FROM transcript_revisions WHERE id = ?`, id).
		Scan(&r.ID, &r.VideoID, &r.SourceSHA256, &r.PipelineFingerprint, &r.ModelName,
			&r.DurationSeconds, &r.WordCount, &r.ChunkCount)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *Store) revisionWords(ctx context.Context, revisionID int64) ([]Word, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT text, normalized_text, start_time, end_time,
		confidence, source_chunk_index, is_filler FROM words WHERE revision_id = ? ORDER BY ordinal`, revisionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ws []Word
	for rows.Next() {
		var w Word
		var conf sql.NullFloat64
		var ci sql.NullInt64
		var filler int
		if err := rows.Scan(&w.Text, &w.NormalizedText, &w.Start, &w.End, &conf, &ci, &filler); err != nil {
			return nil, err
		}
		if conf.Valid {
			w.Confidence = conf.Float64
		}
		if ci.Valid {
			w.SourceChunkIndex = int(ci.Int64)
		}
		w.IsFiller = filler == 1
		ws = append(ws, w)
	}
	return ws, rows.Err()
}

func (s *Store) revisionChunks(ctx context.Context, revisionID int64) ([]Chunk, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, revision_id, ordinal, start_word_ordinal, end_word_ordinal,
		start_time, end_time, text, word_count, source_type, policy_json
		FROM chunks WHERE revision_id = ? ORDER BY ordinal`, revisionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cs []Chunk
	for rows.Next() {
		var c Chunk
		if err := rows.Scan(&c.ID, &c.RevisionID, &c.Ordinal, &c.StartWordOrdinal, &c.EndWordOrdinal,
			&c.Start, &c.End, &c.Text, &c.WordCount, &c.SourceType, &c.PolicyJSON); err != nil {
			return nil, err
		}
		cs = append(cs, c)
	}
	return cs, rows.Err()
}

// recordExport inserts an export row, replacing any prior row for the same
// revision/format/policy.
func (s *Store) recordExport(ctx context.Context, revisionID int64, format, policyJSON, path, hash string) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO exports (revision_id, format, policy_json, output_path, content_sha256)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(revision_id, format, policy_json) DO UPDATE SET
		output_path = excluded.output_path, content_sha256 = excluded.content_sha256,
		created_at = CURRENT_TIMESTAMP`,
		revisionID, format, policyJSON, path, hash)
	return err
}

// Search returns chunk search hits across a corpus.
type SearchHit struct {
	Video     Video
	Chunk     Chunk
	SourceURL string
}

// Search runs an FTS5 query over chunk text for a corpus.
func (s *Store) Search(ctx context.Context, corpusKey, query string, limit int) ([]SearchHit, error) {
	if limit <= 0 {
		limit = 20
	}
	corpusID, err := s.corpusID(ctx, corpusKey)
	if err != nil {
		return nil, err
	}
	ftsQuery := sanitizeFTS(query)
	rows, err := s.db.QueryContext(ctx, `
		SELECT c.id, c.revision_id, c.ordinal, c.start_word_ordinal, c.end_word_ordinal,
		       c.start_time, c.end_time, c.text, c.word_count, c.source_type, c.policy_json,
		       v.id, v.source_id, v.playlist_position, v.title, v.source_url, v.availability,
		       v.media_path, v.audio_path
		FROM chunk_fts f
		JOIN chunks c ON c.id = f.rowid
		JOIN transcript_revisions r ON r.id = c.revision_id
		JOIN videos v ON v.id = r.video_id
		WHERE v.corpus_id = ? AND chunk_fts MATCH ?
		ORDER BY rank
		LIMIT ?`, corpusID, ftsQuery, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var hits []SearchHit
	for rows.Next() {
		var c Chunk
		var v Video
		var mediaPath, audioPath sql.NullString
		if err := rows.Scan(&c.ID, &c.RevisionID, &c.Ordinal, &c.StartWordOrdinal, &c.EndWordOrdinal,
			&c.Start, &c.End, &c.Text, &c.WordCount, &c.SourceType, &c.PolicyJSON,
			&v.ID, &v.SourceID, &v.PlaylistPosition, &v.Title, &v.SourceURL, &v.Availability,
			&mediaPath, &audioPath); err != nil {
			return nil, err
		}
		v.MediaPath = mediaPath.String
		v.AudioPath = audioPath.String
		hits = append(hits, SearchHit{Video: v, Chunk: c, SourceURL: deepLink(v.SourceURL, c.Start)})
	}
	return hits, rows.Err()
}

// --- helpers ---

func manifestSHA256(m *Manifest) string {
	b, _ := json.Marshal(m)
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func effectiveHash(audioHash, mediaHash string) string {
	if audioHash != "" {
		return audioHash
	}
	if mediaHash != "" {
		return mediaHash
	}
	return ""
}

func validateTranscription(result Transcription) error {
	if len(result.Words) == 0 {
		return errors.New("transcription has no words")
	}
	for i, w := range result.Words {
		if strings.TrimSpace(w.Text) == "" {
			return fmt.Errorf("word %d has empty text", i)
		}
		if w.Start < 0 {
			return fmt.Errorf("word %d has negative start %.3f", i, w.Start)
		}
		if w.End < w.Start {
			return fmt.Errorf("word %d end %.3f before start %.3f", i, w.End, w.Start)
		}
		if i > 0 && w.Start < result.Words[i-1].Start-0.75 {
			return fmt.Errorf("word %d start %.3f significantly before previous word start %.3f", i, w.Start, result.Words[i-1].Start)
		}
	}
	return nil
}

type chunkPlan struct {
	Ordinal          int
	StartWordOrdinal int
	EndWordOrdinal   int
	Start            float64
	End              float64
	Text             string
	WordOrdinals     []int
}

func deriveChunks(words []Word, policy ChunkPolicy) []chunkPlan {
	if len(words) == 0 {
		return nil
	}
	maxDur := policy.MaxDurationSeconds
	if maxDur <= 0 {
		maxDur = 15.0
	}
	maxChars := policy.MaxCharacters
	if maxChars <= 0 {
		maxChars = 120
	}
	splitOn := policy.SplitOn
	if len(splitOn) == 0 {
		splitOn = []string{".", "?", "!"}
	}

	var out []chunkPlan
	var current []int
	var segStart float64

	flush := func(lastIdx int) {
		if len(current) == 0 {
			return
		}
		parts := make([]string, 0, len(current))
		for _, idx := range current {
			parts = append(parts, words[idx].Text)
		}
		out = append(out, chunkPlan{
			Ordinal:          len(out),
			StartWordOrdinal: current[0],
			EndWordOrdinal:   current[len(current)-1],
			Start:            segStart,
			End:              words[lastIdx].End,
			Text:             strings.Join(parts, " "),
			WordOrdinals:     append([]int(nil), current...),
		})
		current = nil
	}

	for i, w := range words {
		if len(current) == 0 {
			segStart = w.Start
		}
		current = append(current, i)
		text := strings.Join(func() []string {
			ps := make([]string, 0, len(current))
			for _, idx := range current {
				ps = append(ps, words[idx].Text)
			}
			return ps
		}(), " ")
		dur := w.End - segStart
		split := false
		for _, ending := range splitOn {
			if strings.HasSuffix(w.Text, ending) && dur > 0.5 {
				split = true
				break
			}
		}
		if dur > maxDur || len(text) > maxChars {
			split = true
		}
		if split {
			flush(i)
		}
	}
	if len(current) > 0 {
		flush(len(words) - 1)
	}
	return out
}

func deepLink(sourceURL string, start float64) string {
	if sourceURL == "" {
		return ""
	}
	secs := int(start)
	if secs <= 0 {
		return sourceURL
	}
	sep := "&"
	if !strings.Contains(sourceURL, "?") {
		sep = "?"
	}
	return fmt.Sprintf("%s%st=%d", sourceURL, sep, secs)
}

func sanitizeFTS(q string) string {
	q = strings.TrimSpace(q)
	if q == "" {
		return ""
	}
	terms := strings.Fields(q)
	for i, t := range terms {
		if !strings.HasPrefix(t, "\"") {
			terms[i] = "\"" + strings.ReplaceAll(t, "\"", "") + "\""
		}
	}
	return strings.Join(terms, " ")
}

func nullable(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func nullableConf(f float64) any {
	if f == 0 {
		return nil
	}
	return f
}

func rollbackOnErr(tx *sql.Tx, err error) {
	if err != nil {
		_ = tx.Rollback()
	}
}

// scanVideo scans a video from an open rows iterator.
func scanVideo(rows *sql.Rows) (Video, error) {
	var v Video
	var mediaPath, audioPath, mediaHash, audioHash, metadata, lastError sql.NullString
	var activeRevision sql.NullInt64
	if err := rows.Scan(&v.ID, &v.SourceID, &v.PlaylistPosition, &v.Title, &v.SourceURL, &v.Availability,
		&mediaPath, &audioPath, &mediaHash, &audioHash, &v.DurationSeconds,
		&metadata, &v.ProcessingState, &activeRevision, &lastError); err != nil {
		return v, err
	}
	v.MediaPath = mediaPath.String
	v.AudioPath = audioPath.String
	v.MediaSHA256 = mediaHash.String
	v.AudioSHA256 = audioHash.String
	v.MetadataJSON = metadata.String
	v.ActiveRevisionID = activeRevision
	v.LastError = lastError
	return v, nil
}

func scanVideoRow(row *sql.Row) (*Video, error) {
	var v Video
	var mediaPath, audioPath, mediaHash, audioHash, metadata, lastError sql.NullString
	var activeRevision sql.NullInt64
	err := row.Scan(&v.ID, &v.SourceID, &v.PlaylistPosition, &v.Title, &v.SourceURL, &v.Availability,
		&mediaPath, &audioPath, &mediaHash, &audioHash, &v.DurationSeconds,
		&metadata, &v.ProcessingState, &activeRevision, &lastError)
	if err != nil {
		return nil, err
	}
	v.MediaPath = mediaPath.String
	v.AudioPath = audioPath.String
	v.MediaSHA256 = mediaHash.String
	v.AudioSHA256 = audioHash.String
	v.MetadataJSON = metadata.String
	v.ActiveRevisionID = activeRevision
	v.LastError = lastError
	return &v, nil
}
