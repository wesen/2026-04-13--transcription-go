package corpus

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func testStore(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	s, err := OpenStore(filepath.Join(dir, "corpus.db"))
	if err != nil {
		t.Fatalf("OpenStore: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

const testHash = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

func sampleManifest(audioPath string) *Manifest {
	return &Manifest{
		Schema: SchemaManifest,
		Corpus: CorpusMetadata{ID: "test-corpus", Title: "Test Corpus", SourceURL: "https://example/playlist"},
		Items: []ManifestItem{
			{
				SourceID: "vid-available", Position: 1, Title: "Available Video",
				SourceURL:    "https://example/watch?v=vid-available",
				Availability: AvailabilityAvailable, AudioPath: audioPath, AudioSHA256: testHash,
				DurationSeconds: 3.0,
			},
			{
				SourceID: "vid-members", Position: 2, Title: "Members Only",
				SourceURL:    "https://example/watch?v=vid-members",
				Availability: AvailabilityMembersOnly, DurationSeconds: 0,
			},
		},
	}
}

func TestManifestValidation(t *testing.T) {
	dir := t.TempDir()
	audio := filepath.Join(dir, "a.wav")
	if err := writeFile(audio, []byte("fake")); err != nil {
		t.Fatal(err)
	}
	m := sampleManifest(audio)
	if err := m.Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}

	bad := *m
	bad.Schema = "wrong"
	if err := bad.Validate(); err == nil {
		t.Error("expected schema error")
	}

	dup := *m
	dup.Items = append(dup.Items, ManifestItem{SourceID: "vid-available", Position: 3, Title: "Dup", Availability: AvailabilityAvailable, AudioPath: audio})
	if err := dup.Validate(); err == nil {
		t.Error("expected duplicate source_id error")
	}

	availNoAudio := *m
	availNoAudio.Items[0].AudioPath = "/nonexistent/path.wav"
	if err := availNoAudio.Validate(); err == nil {
		t.Error("expected missing audio error")
	}
}

func TestImportManifestAndPlan(t *testing.T) {
	dir := t.TempDir()
	audio := filepath.Join(dir, "a.wav")
	if err := writeFile(audio, []byte("fake")); err != nil {
		t.Fatal(err)
	}
	store := testStore(t)
	m := sampleManifest(audio)
	ctx := context.Background()

	if _, err := store.ImportManifest(ctx, m); err != nil {
		t.Fatalf("ImportManifest: %v", err)
	}

	fp := DefaultFingerprint(60)
	work, err := store.Plan(ctx, m.Corpus.ID, fp, false, nil)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	if len(work) != 2 {
		t.Fatalf("expected 2 work items, got %d", len(work))
	}
	var pending, unavailable int
	for _, w := range work {
		switch w.Reason {
		case "pending":
			pending++
		case "unavailable":
			unavailable++
		}
	}
	if pending != 1 {
		t.Errorf("expected 1 pending, got %d", pending)
	}
	if unavailable != 1 {
		t.Errorf("expected 1 unavailable, got %d", unavailable)
	}
}

type fakeTranscriber struct {
	words []Word
}

func (f *fakeTranscriber) Transcribe(ctx context.Context, audioPath string, opts TranscribeOptions) (Transcription, error) {
	return Transcription{
		Words:           f.words,
		DurationSeconds: 3.0,
		ChunkCount:      1,
	}, nil
}

func TestCommitTranscriptAtomicAndResume(t *testing.T) {
	dir := t.TempDir()
	audio := filepath.Join(dir, "a.wav")
	if err := writeFile(audio, []byte("fake")); err != nil {
		t.Fatal(err)
	}
	store := testStore(t)
	m := sampleManifest(audio)
	ctx := context.Background()
	if _, err := store.ImportManifest(ctx, m); err != nil {
		t.Fatalf("ImportManifest: %v", err)
	}
	fp := DefaultFingerprint(60)
	work, err := store.Plan(ctx, m.Corpus.ID, fp, false, []string{"vid-available"})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	if len(work) != 1 || !work[0].NeedsTranscribe {
		t.Fatalf("expected 1 pending transcribe item")
	}
	v := work[0].Video

	attempt, err := store.BeginAttempt(ctx, v, fp, "127.0.0.1:9999")
	if err != nil {
		t.Fatalf("BeginAttempt: %v", err)
	}
	result := Transcription{
		Words: []Word{
			{Text: "Hello", NormalizedText: "hello", Start: 0, End: 0.5, SourceChunkIndex: 0},
			{Text: "world.", NormalizedText: "world", Start: 0.5, End: 1.0, SourceChunkIndex: 0},
			{Text: "No", NormalizedText: "no", Start: 1.5, End: 1.8, SourceChunkIndex: 0},
			{Text: "end.", NormalizedText: "end", Start: 1.8, End: 2.3, SourceChunkIndex: 0},
		},
		DurationSeconds: 3.0,
		ChunkCount:      1,
	}
	rev, err := store.CommitTranscript(ctx, v, attempt, fp, result, DefaultChunkPolicy())
	if err != nil {
		t.Fatalf("CommitTranscript: %v", err)
	}
	if rev.WordCount != 4 {
		t.Errorf("revision word_count=%d want 4", rev.WordCount)
	}

	// Verify DB invariants.
	var wordCount, chunkCount, activeRevision, state int
	store.db.QueryRow(`SELECT COUNT(*) FROM words WHERE revision_id = ?`, rev.ID).Scan(&wordCount)
	if wordCount != 4 {
		t.Errorf("db word count=%d want 4", wordCount)
	}
	store.db.QueryRow(`SELECT COUNT(*) FROM chunks WHERE revision_id = ?`, rev.ID).Scan(&chunkCount)
	if chunkCount == 0 {
		t.Error("expected chunks")
	}
	var badChunkWordCount int
	store.db.QueryRow(`SELECT COUNT(*) FROM chunks c WHERE c.word_count != (SELECT COUNT(*) FROM chunk_words cw WHERE cw.chunk_id = c.id)`).Scan(&badChunkWordCount)
	if badChunkWordCount != 0 {
		t.Errorf("found %d chunks with mismatched word_count", badChunkWordCount)
	}
	store.db.QueryRow(`SELECT COUNT(*) FROM videos WHERE id = ? AND processing_state = 'complete' AND active_revision_id = ?`, v.ID, rev.ID).Scan(&activeRevision)
	if activeRevision != 1 {
		t.Errorf("expected video complete with active revision, got %d", activeRevision)
	}

	// Resume: re-plan with same fingerprint; item should be unchanged.
	work2, err := store.Plan(ctx, m.Corpus.ID, fp, false, []string{"vid-available"})
	if err != nil {
		t.Fatalf("Plan2: %v", err)
	}
	if len(work2) != 1 || work2[0].NeedsTranscribe {
		t.Fatalf("expected unchanged item after commit, got reason=%s needsTranscribe=%v", work2[0].Reason, work2[0].NeedsTranscribe)
	}

	// Stale: change fingerprint.
	fp2 := DefaultFingerprint(90)
	work3, err := store.Plan(ctx, m.Corpus.ID, fp2, false, []string{"vid-available"})
	if err != nil {
		t.Fatalf("Plan3: %v", err)
	}
	if len(work3) != 1 || !work3[0].NeedsTranscribe || work3[0].Reason != "stale" {
		t.Fatalf("expected stale transcribe item, got reason=%s", work3[0].Reason)
	}
	_ = state
}

func TestReconcileAbandoned(t *testing.T) {
	dir := t.TempDir()
	audio := filepath.Join(dir, "a.wav")
	if err := writeFile(audio, []byte("fake")); err != nil {
		t.Fatal(err)
	}
	store := testStore(t)
	m := sampleManifest(audio)
	ctx := context.Background()
	if _, err := store.ImportManifest(ctx, m); err != nil {
		t.Fatalf("ImportManifest: %v", err)
	}
	fp := DefaultFingerprint(60)
	work, _ := store.Plan(ctx, m.Corpus.ID, fp, false, []string{"vid-available"})
	v := work[0].Video
	attempt, _ := store.BeginAttempt(ctx, v, fp, "endpoint")
	// Simulate crash: leave attempt running, then reconcile.
	n, err := store.ReconcileAbandoned(ctx)
	if err != nil {
		t.Fatalf("ReconcileAbandoned: %v", err)
	}
	if n < 1 {
		t.Errorf("expected at least 1 abandoned, got %d", n)
	}
	var state string
	store.db.QueryRow(`SELECT state FROM transcript_attempts WHERE id = ?`, attempt.ID).Scan(&state)
	if state != string(AttemptAbandoned) {
		t.Errorf("attempt state=%s want abandoned", state)
	}
}

func TestValidateTranscriptionRejectsBadTimes(t *testing.T) {
	cases := []struct {
		name string
		t    Transcription
	}{
		{"empty", Transcription{Words: nil}},
		{"end before start", Transcription{Words: []Word{{Text: "x", NormalizedText: "x", Start: 2, End: 1}}}},
		{"negative start", Transcription{Words: []Word{{Text: "x", NormalizedText: "x", Start: -1, End: 1}}}},
		{"empty text", Transcription{Words: []Word{{Text: "  ", NormalizedText: "x", Start: 0, End: 1}}}},
	}
	for _, c := range cases {
		if err := validateTranscription(c.t); err == nil {
			t.Errorf("%s: expected error", c.name)
		}
	}
}

func TestDeriveChunksTrailingBufferNonZeroEnd(t *testing.T) {
	words := []Word{
		{Text: "Hello", Start: 0, End: 0.5},
		{Text: "world.", Start: 0.5, End: 1.0},
		{Text: "No", Start: 1.5, End: 1.8},
		{Text: "terminator", Start: 1.8, End: 2.3},
	}
	chunks := deriveChunks(words, DefaultChunkPolicy())
	if len(chunks) < 2 {
		t.Fatalf("expected at least 2 chunks, got %d", len(chunks))
	}
	last := chunks[len(chunks)-1]
	if last.End == 0 {
		t.Error("trailing chunk end is zero")
	}
	if last.End != words[len(words)-1].End {
		t.Errorf("trailing chunk end=%.3f want %.3f", last.End, words[len(words)-1].End)
	}
}

func TestSearch(t *testing.T) {
	dir := t.TempDir()
	audio := filepath.Join(dir, "a.wav")
	if err := writeFile(audio, []byte("fake")); err != nil {
		t.Fatal(err)
	}
	store := testStore(t)
	m := sampleManifest(audio)
	ctx := context.Background()
	if _, err := store.ImportManifest(ctx, m); err != nil {
		t.Fatalf("ImportManifest: %v", err)
	}
	fp := DefaultFingerprint(60)
	work, _ := store.Plan(ctx, m.Corpus.ID, fp, false, []string{"vid-available"})
	v := work[0].Video
	attempt, _ := store.BeginAttempt(ctx, v, fp, "endpoint")
	result := Transcription{
		Words: []Word{
			{Text: "category", NormalizedText: "category", Start: 0, End: 0.5},
			{Text: "theory.", NormalizedText: "theory", Start: 0.5, End: 1.0},
		},
		DurationSeconds: 1.5,
	}
	rev, err := store.CommitTranscript(ctx, v, attempt, fp, result, DefaultChunkPolicy())
	if err != nil {
		t.Fatalf("CommitTranscript: %v", err)
	}
	_ = rev

	hits, err := store.Search(ctx, m.Corpus.ID, "category", 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(hits) == 0 {
		t.Fatal("expected search hits for 'category'")
	}
	if hits[0].Video.SourceID != "vid-available" {
		t.Errorf("hit source_id=%s want vid-available", hits[0].Video.SourceID)
	}
	if hits[0].SourceURL == "" {
		t.Error("expected deep link")
	}
}

func writeFile(path string, data []byte) error {
	return os.WriteFile(path, data, 0o644)
}
