package bedrock

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateOwnWritesRejectsTrackedFileRestoredToOriginal(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "tracked.txt")
	if err := os.WriteFile(path, []byte("original"), 0o644); err != nil {
		t.Fatal(err)
	}
	set := NewChangeSet()
	if err := set.Apply(root, []FileChange{{Path: "tracked.txt", Content: "bedrock edit"}}, nil); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("original"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := set.ValidateOwnWrites(root)
	if err == nil || !strings.Contains(err.Error(), "changed after BedRock wrote it") {
		t.Fatalf("expected ownership conflict, got %v", err)
	}
	data, readErr := os.ReadFile(path)
	if readErr != nil || string(data) != "original" {
		t.Fatalf("external state changed: %q err=%v", data, readErr)
	}
}
