package bedrock

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestPrepareRunJournalPersistsOriginalsBeforeMutation(t *testing.T) {
	root := t.TempDir()
	existing := filepath.Join(root, "existing.txt")
	original := []byte("before\n")
	if err := os.WriteFile(existing, original, 0o640); err != nil {
		t.Fatal(err)
	}

	journal, path, err := PrepareRunJournal(root, "run-test", []string{"existing.txt", "new.txt"})
	if err != nil {
		t.Fatal(err)
	}
	if journal.State != JournalPrepared || len(journal.Originals) != 2 {
		t.Fatalf("unexpected journal: %#v", journal)
	}
	if !journal.Originals[0].Existed || !bytes.Equal(journal.Originals[0].Content, original) || journal.Originals[0].SHA256 == "" {
		t.Fatalf("existing original not captured: %#v", journal.Originals[0])
	}
	if journal.Originals[1].Existed || len(journal.Originals[1].Content) != 0 {
		t.Fatalf("new path should be recorded as absent: %#v", journal.Originals[1])
	}

	persistedBytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var persisted RunJournal
	if err := json.Unmarshal(persistedBytes, &persisted); err != nil {
		t.Fatal(err)
	}
	if persisted.State != JournalPrepared || !bytes.Equal(persisted.Originals[0].Content, original) {
		t.Fatalf("persisted journal lost original state: %#v", persisted)
	}

	if err := os.WriteFile(existing, []byte("after\n"), 0o640); err != nil {
		t.Fatal(err)
	}
	persistedBytes, err = os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(persistedBytes, []byte("after")) || !bytes.Contains(persistedBytes, original) {
		t.Fatalf("journal changed with repository mutation: %s", persistedBytes)
	}
}

func TestRecordMutationIntentPreservesOriginalAndUpdatesIntent(t *testing.T) {
	root := t.TempDir()
	original := []byte("original\n")
	if err := os.WriteFile(filepath.Join(root, "file.txt"), original, 0o640); err != nil {
		t.Fatal(err)
	}
	_, path, err := PrepareRunJournal(root, "run-intent", []string{"file.txt"})
	if err != nil {
		t.Fatal(err)
	}

	first := []byte("first attempt\n")
	journal, err := RecordMutationIntent(path, "file.txt", first)
	if err != nil {
		t.Fatal(err)
	}
	firstSum := sha256.Sum256(first)
	if journal.Originals[0].IntendedSHA256 != hex.EncodeToString(firstSum[:]) {
		t.Fatalf("first intended hash = %q", journal.Originals[0].IntendedSHA256)
	}

	second := []byte("repair attempt\n")
	journal, err = RecordMutationIntent(path, "file.txt", second)
	if err != nil {
		t.Fatal(err)
	}
	secondSum := sha256.Sum256(second)
	if journal.Originals[0].IntendedSHA256 != hex.EncodeToString(secondSum[:]) {
		t.Fatalf("second intended hash = %q", journal.Originals[0].IntendedSHA256)
	}
	if !bytes.Equal(journal.Originals[0].Content, original) {
		t.Fatalf("repair attempt changed original: %q", journal.Originals[0].Content)
	}
}

func TestRecordMutationIntentRejectsUnpreparedAndCompleted(t *testing.T) {
	root := t.TempDir()
	_, path, err := PrepareRunJournal(root, "run-intent-guard", []string{"file.txt"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := RecordMutationIntent(path, "other.txt", []byte("write")); err == nil {
		t.Fatal("expected unprepared path to be rejected")
	}
	if _, err := TransitionRunJournal(path, JournalMutating); err != nil {
		t.Fatal(err)
	}
	if _, err := TransitionRunJournal(path, JournalCompleted); err != nil {
		t.Fatal(err)
	}
	if _, err := RecordMutationIntent(path, "file.txt", []byte("late")); err == nil {
		t.Fatal("expected completed journal to reject mutation intent")
	}
}

func TestTransitionRunJournalPersistsForwardState(t *testing.T) {
	root := t.TempDir()
	_, path, err := PrepareRunJournal(root, "run-transition", []string{"file.txt"})
	if err != nil {
		t.Fatal(err)
	}

	mutating, err := TransitionRunJournal(path, JournalMutating)
	if err != nil {
		t.Fatal(err)
	}
	if mutating.State != JournalMutating {
		t.Fatalf("state = %s, want %s", mutating.State, JournalMutating)
	}

	completed, err := TransitionRunJournal(path, JournalCompleted)
	if err != nil {
		t.Fatal(err)
	}
	if completed.State != JournalCompleted {
		t.Fatalf("state = %s, want %s", completed.State, JournalCompleted)
	}

	persistedBytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var persisted RunJournal
	if err := json.Unmarshal(persistedBytes, &persisted); err != nil {
		t.Fatal(err)
	}
	if persisted.State != JournalCompleted {
		t.Fatalf("persisted state = %s, want %s", persisted.State, JournalCompleted)
	}
	if _, err := TransitionRunJournal(path, JournalMutating); err == nil {
		t.Fatal("expected completed journal to reject backward transition")
	}
}

func TestPrepareRunJournalRejectsUnsafeAndDuplicatePaths(t *testing.T) {
	root := t.TempDir()
	for _, paths := range [][]string{{"../escape"}, {"same.txt", "same.txt"}, {".git/config"}} {
		if _, _, err := PrepareRunJournal(root, "run-test", paths); err == nil {
			t.Fatalf("expected paths %v to be rejected", paths)
		}
	}
}

func TestPrepareRunJournalRejectsUnsafeRunID(t *testing.T) {
	root := t.TempDir()
	for _, runID := range []string{"", ".", "..", "../escape", `..\\escape`, "nested/run"} {
		if _, _, err := PrepareRunJournal(root, runID, []string{"file.txt"}); err == nil {
			t.Fatalf("expected run id %q to be rejected", runID)
		}
	}
}
