package bedrock

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func TestRecordCapturedMutationIntentUsesBoundaryOriginalWithoutReread(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "owned.txt")
	if err := os.WriteFile(target, []byte("boundary-original"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, journalPath, err := PrepareRunJournal(root, "captured-boundary", nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := TransitionRunJournal(journalPath, JournalMutating); err != nil {
		t.Fatal(err)
	}

	capturedBytes := []byte("boundary-original")
	sum := sha256.Sum256(capturedBytes)
	captured := JournalOriginal{
		Path:    "owned.txt",
		Existed: true,
		Mode:    0o600,
		SHA256:  hex.EncodeToString(sum[:]),
		Content: capturedBytes,
	}
	// Simulate an external write after the boundary captured the original. The
	// durable API must not silently re-read and adopt these different bytes.
	if err := os.WriteFile(target, []byte("concurrent-change"), 0o600); err != nil {
		t.Fatal(err)
	}
	journal, err := recordCapturedMutationIntent(journalPath, captured, []byte("bedrock-write"))
	if err != nil {
		t.Fatal(err)
	}
	if len(journal.Originals) != 1 {
		t.Fatalf("original count = %d, want 1", len(journal.Originals))
	}
	if got := string(journal.Originals[0].Content); got != "boundary-original" {
		t.Fatalf("durable original = %q, want boundary capture", got)
	}
}

func TestRecordCapturedMutationIntentRequiresMutatingState(t *testing.T) {
	root := t.TempDir()
	_, journalPath, err := PrepareRunJournal(root, "captured-state", nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = recordCapturedMutationIntent(journalPath, JournalOriginal{Path: "new.txt"}, []byte("new"))
	if err == nil {
		t.Fatal("expected PREPARED journal to reject mutation intent")
	}
}

func TestRecordCapturedMutationIntentRejectsContradictoryCapture(t *testing.T) {
	root := t.TempDir()
	_, journalPath, err := PrepareRunJournal(root, "captured-invalid", nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := TransitionRunJournal(journalPath, JournalMutating); err != nil {
		t.Fatal(err)
	}

	captured := JournalOriginal{
		Path:    "owned.txt",
		Existed: true,
		Mode:    0o600,
		SHA256:  hex.EncodeToString(make([]byte, sha256.Size)),
		Content: []byte("actual-original"),
	}
	if _, err := recordCapturedMutationIntent(journalPath, captured, []byte("bedrock-write")); err == nil {
		t.Fatal("expected contradictory captured hash/content to be rejected")
	}

	journal, err := loadRunJournal(journalPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(journal.Originals) != 0 {
		t.Fatalf("journal gained %d originals after rejected capture", len(journal.Originals))
	}
}

func TestRecordCapturedMutationIntentRejectsMetadataForMissingOriginal(t *testing.T) {
	root := t.TempDir()
	_, journalPath, err := PrepareRunJournal(root, "captured-missing-invalid", nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := TransitionRunJournal(journalPath, JournalMutating); err != nil {
		t.Fatal(err)
	}

	captured := JournalOriginal{Path: "new.txt", Existed: false, Content: []byte("impossible")}
	if _, err := recordCapturedMutationIntent(journalPath, captured, []byte("bedrock-write")); err == nil {
		t.Fatal("expected missing original with content to be rejected")
	}
}
