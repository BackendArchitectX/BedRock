package bedrock

import (
	"bytes"
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

func TestPrepareRunJournalRejectsUnsafeAndDuplicatePaths(t *testing.T) {
	root := t.TempDir()
	for _, paths := range [][]string{{"../escape"}, {"same.txt", "same.txt"}, {".git/config"}} {
		if _, _, err := PrepareRunJournal(root, "run-test", paths); err == nil {
			t.Fatalf("expected paths %v to be rejected", paths)
		}
	}
}
