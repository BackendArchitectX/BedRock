package bedrock

import (
	"os"
	"path/filepath"
	"testing"
)

func TestChangeSetRejectsTraversal(t *testing.T) {
	root := t.TempDir()
	set := NewChangeSet()
	err := set.Apply(root, []FileChange{{Path: "../escape.txt", Content: "nope"}}, nil)
	if err == nil {
		t.Fatal("expected traversal to be rejected")
	}
}

func TestChangeSetProtectsDirtyPath(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "owned.txt")
	if err := os.WriteFile(path, []byte("user work"), 0o644); err != nil {
		t.Fatal(err)
	}
	set := NewChangeSet()
	err := set.Apply(root, []FileChange{{Path: "owned.txt", Content: "overwrite"}}, map[string]struct{}{"owned.txt": {}})
	if err == nil {
		t.Fatal("expected dirty path to be protected")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "user work" {
		t.Fatalf("dirty file changed: %q", data)
	}
}

func TestChangeSetRollbackRestoresAndRemoves(t *testing.T) {
	root := t.TempDir()
	existing := filepath.Join(root, "existing.txt")
	if err := os.WriteFile(existing, []byte("before"), 0o600); err != nil {
		t.Fatal(err)
	}
	set := NewChangeSet()
	if err := set.Apply(root, []FileChange{
		{Path: "existing.txt", Content: "after"},
		{Path: "new.txt", Content: "new"},
	}, nil); err != nil {
		t.Fatal(err)
	}
	if err := set.Rollback(root); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(existing)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "before" {
		t.Fatalf("existing file not restored: %q", data)
	}
	if _, err := os.Stat(filepath.Join(root, "new.txt")); !os.IsNotExist(err) {
		t.Fatalf("new file still exists: %v", err)
	}
}
