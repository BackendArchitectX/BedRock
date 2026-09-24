package bedrock

import "testing"

func TestRecordMutationIntentRejectsPreparedJournal(t *testing.T) {
	root := t.TempDir()
	_, path, err := PrepareRunJournal(root, "run-prepared-guard", []string{"file.txt"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := RecordMutationIntent(path, "file.txt", []byte("intent")); err == nil {
		t.Fatal("expected prepared journal to reject mutation intent")
	}
	persisted, err := loadRunJournal(path)
	if err != nil {
		t.Fatal(err)
	}
	if persisted.State != JournalPrepared {
		t.Fatalf("state = %s, want %s", persisted.State, JournalPrepared)
	}
	if len(persisted.Originals) != 1 || persisted.Originals[0].IntendedSHA256 != "" {
		t.Fatalf("prepared journal gained mutation ownership: %#v", persisted.Originals)
	}
}
