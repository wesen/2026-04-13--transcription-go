package output

import (
	"database/sql"
	"fmt"
	"strings"

	_ "modernc.org/sqlite"
)

// WriteSQLite writes word-level transcript data to a SQLite database.
// The schema matches the Python pipeline's transcript_db.py for compatibility.
func WriteSQLite(words []Word, outputPath string) error {
	db, err := sql.Open("sqlite", outputPath)
	if err != nil {
		return fmt.Errorf("open sqlite: %w", err)
	}
	defer db.Close()

	if err := initSchema(db); err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	// Insert words
	wordStmt, err := tx.Prepare(`INSERT INTO words (word, start_time, end_time, is_filler) VALUES (?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("prepare word insert: %w", err)
	}
	defer wordStmt.Close()

	for _, w := range words {
		filler := 0
		if w.IsFiller() {
			filler = 1
		}
		if _, err := wordStmt.Exec(w.Word, w.Start, w.End, filler); err != nil {
			return fmt.Errorf("insert word: %w", err)
		}
	}

	// Create default chunks
	if err := createChunks(tx, words); err != nil {
		return fmt.Errorf("create chunks: %w", err)
	}

	return tx.Commit()
}

func initSchema(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS words (
		id INTEGER PRIMARY KEY,
		word TEXT NOT NULL,
		start_time REAL NOT NULL,
		end_time REAL NOT NULL,
		duration REAL GENERATED ALWAYS AS (end_time - start_time) STORED,
		chunk_id INTEGER,
		is_filler BOOLEAN DEFAULT 0,
		is_removed BOOLEAN DEFAULT 0,
		confidence REAL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS chunks (
		id INTEGER PRIMARY KEY,
		start_time REAL NOT NULL,
		end_time REAL NOT NULL,
		duration REAL GENERATED ALWAYS AS (end_time - start_time) STORED,
		text TEXT NOT NULL,
		word_count INTEGER,
		source_type TEXT DEFAULT 'auto',
		metadata TEXT
	);

	CREATE TABLE IF NOT EXISTS chunk_words (
		chunk_id INTEGER,
		word_id INTEGER,
		position INTEGER,
		PRIMARY KEY (chunk_id, word_id),
		FOREIGN KEY (chunk_id) REFERENCES chunks(id),
		FOREIGN KEY (word_id) REFERENCES words(id)
	);

	CREATE INDEX IF NOT EXISTS idx_words_time ON words(start_time, end_time);
	CREATE INDEX IF NOT EXISTS idx_words_filler ON words(is_filler) WHERE is_filler = 1;
	CREATE INDEX IF NOT EXISTS idx_chunks_time ON chunks(start_time, end_time);
	`
	_, err := db.Exec(schema)
	return err
}

func createChunks(tx *sql.Tx, words []Word) error {
	segments := BuildSegments(words, 15.0, 120)

	for _, seg := range segments {
		res, err := tx.Exec(
			`INSERT INTO chunks (start_time, end_time, text, word_count) VALUES (?, ?, ?, ?)`,
			seg.Start, seg.End, seg.Text, len(splitWords(seg.Text)),
		)
		if err != nil {
			return err
		}
		chunkID, _ := res.LastInsertId()

		// Find words in this time range and link them in order.
		rows, err := tx.Query(
			`SELECT id FROM words WHERE start_time >= ? AND end_time <= ? ORDER BY start_time`,
			seg.Start-0.01, seg.End+0.01,
		)
		if err != nil {
			return err
		}
		pos := 0
		for rows.Next() {
			var wordID int
			if err := rows.Scan(&wordID); err != nil {
				rows.Close()
				return err
			}
			if _, err := tx.Exec(`INSERT INTO chunk_words (chunk_id, word_id, position) VALUES (?, ?, ?)`, chunkID, wordID, pos); err != nil {
				rows.Close()
				return err
			}
			if _, err := tx.Exec(`UPDATE words SET chunk_id = ? WHERE id = ?`, chunkID, wordID); err != nil {
				rows.Close()
				return err
			}
			pos++
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return err
		}
		rows.Close()

		// Correct the chunk word_count to the actual number of linked words,
		// rather than the whitespace-split length of the joined text.
		if _, err := tx.Exec(`UPDATE chunks SET word_count = ? WHERE id = ?`, pos, chunkID); err != nil {
			return err
		}
	}
	return nil
}

// splitWords splits a string on whitespace and returns the non-empty tokens.
func splitWords(s string) []string {
	return strings.Fields(s)
}
