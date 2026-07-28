package corpus

// schemaSQL is the canonical corpus database schema. It is applied once when
// a store is opened on a fresh or existing database.
const schemaSQL = `
PRAGMA foreign_keys = ON;
PRAGMA journal_mode = WAL;

CREATE TABLE IF NOT EXISTS corpora (
    id INTEGER PRIMARY KEY,
    corpus_key TEXT NOT NULL UNIQUE,
    title TEXT NOT NULL,
    source_url TEXT,
    manifest_schema TEXT NOT NULL,
    manifest_sha256 TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS videos (
    id INTEGER PRIMARY KEY,
    corpus_id INTEGER NOT NULL REFERENCES corpora(id),
    source_id TEXT NOT NULL,
    playlist_position INTEGER,
    title TEXT NOT NULL,
    source_url TEXT,
    availability TEXT NOT NULL CHECK (
      availability IN ('available','members_only','private','deleted','missing','unknown')
    ),
    media_path TEXT,
    audio_path TEXT,
    media_sha256 TEXT,
    audio_sha256 TEXT,
    duration_seconds REAL,
    metadata_json TEXT,
    processing_state TEXT NOT NULL DEFAULT 'pending' CHECK (
      processing_state IN ('pending','ready','transcribing','complete','failed','stale','unavailable')
    ),
    active_revision_id INTEGER,
    last_error TEXT,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(corpus_id, source_id),
    UNIQUE(corpus_id, playlist_position)
);

CREATE INDEX IF NOT EXISTS videos_corpus_state ON videos(corpus_id, processing_state);

CREATE TABLE IF NOT EXISTS transcript_attempts (
    id INTEGER PRIMARY KEY,
    video_id INTEGER NOT NULL REFERENCES videos(id),
    attempt_number INTEGER NOT NULL,
    state TEXT NOT NULL CHECK (state IN ('running','succeeded','failed','abandoned')),
    source_sha256 TEXT NOT NULL,
    pipeline_fingerprint TEXT NOT NULL,
    started_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    finished_at TEXT,
    processing_ms INTEGER,
    server_endpoint TEXT,
    chunk_count INTEGER,
    word_count INTEGER,
    error_class TEXT,
    error_message TEXT,
    log_path TEXT,
    UNIQUE(video_id, attempt_number)
);

CREATE TABLE IF NOT EXISTS transcript_revisions (
    id INTEGER PRIMARY KEY,
    video_id INTEGER NOT NULL REFERENCES videos(id),
    attempt_id INTEGER NOT NULL UNIQUE REFERENCES transcript_attempts(id),
    source_sha256 TEXT NOT NULL,
    pipeline_fingerprint TEXT NOT NULL,
    model_name TEXT NOT NULL,
    duration_seconds REAL NOT NULL,
    word_count INTEGER NOT NULL,
    chunk_count INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(video_id, source_sha256, pipeline_fingerprint)
);

CREATE TABLE IF NOT EXISTS words (
    id INTEGER PRIMARY KEY,
    revision_id INTEGER NOT NULL REFERENCES transcript_revisions(id) ON DELETE CASCADE,
    ordinal INTEGER NOT NULL,
    text TEXT NOT NULL,
    normalized_text TEXT NOT NULL,
    start_time REAL NOT NULL CHECK (start_time >= 0),
    end_time REAL NOT NULL CHECK (end_time >= start_time),
    confidence REAL,
    source_chunk_index INTEGER,
    is_filler INTEGER NOT NULL DEFAULT 0,
    is_removed INTEGER NOT NULL DEFAULT 0,
    metadata_json TEXT,
    UNIQUE(revision_id, ordinal)
);

CREATE INDEX IF NOT EXISTS words_revision_time ON words(revision_id, start_time, end_time);
CREATE INDEX IF NOT EXISTS words_normalized ON words(normalized_text);

CREATE TABLE IF NOT EXISTS chunks (
    id INTEGER PRIMARY KEY,
    revision_id INTEGER NOT NULL REFERENCES transcript_revisions(id) ON DELETE CASCADE,
    ordinal INTEGER NOT NULL,
    start_word_ordinal INTEGER NOT NULL,
    end_word_ordinal INTEGER NOT NULL,
    start_time REAL NOT NULL,
    end_time REAL NOT NULL,
    text TEXT NOT NULL,
    word_count INTEGER NOT NULL,
    source_type TEXT NOT NULL,
    policy_json TEXT NOT NULL,
    UNIQUE(revision_id, source_type, ordinal)
);

CREATE TABLE IF NOT EXISTS chunk_words (
    chunk_id INTEGER NOT NULL REFERENCES chunks(id) ON DELETE CASCADE,
    word_id INTEGER NOT NULL REFERENCES words(id) ON DELETE CASCADE,
    position INTEGER NOT NULL,
    PRIMARY KEY(chunk_id, position),
    UNIQUE(chunk_id, word_id)
);

CREATE TABLE IF NOT EXISTS exports (
    id INTEGER PRIMARY KEY,
    revision_id INTEGER NOT NULL REFERENCES transcript_revisions(id) ON DELETE CASCADE,
    format TEXT NOT NULL,
    policy_json TEXT NOT NULL,
    output_path TEXT NOT NULL,
    content_sha256 TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(revision_id, format, policy_json)
);

CREATE VIRTUAL TABLE IF NOT EXISTS chunk_fts USING fts5(
    text,
    content='chunks',
    content_rowid='id',
    tokenize='unicode61'
);
`
