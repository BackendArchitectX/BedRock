package bedrock

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestParsePorcelainV1ZPreservesUnusualPaths(t *testing.T) {
	dirty, err := parsePorcelainV1Z([]byte(" M path with spaces.txt\x00??  leading-and-trailing .txt \x00"))
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"path with spaces.txt", " leading-and-trailing .txt "} {
		if _, ok := dirty[path]; !ok {
			t.Fatalf("dirty path %q was not preserved: %#v", path, dirty)
		}
	}
}

func TestParsePorcelainV1ZProtectsBothRenamePaths(t *testing.T) {
	dirty, err := parsePorcelainV1Z([]byte("R  new name.txt\x00old name.txt\x00"))
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"new name.txt", "old name.txt"} {
		if _, ok := dirty[path]; !ok {
			t.Fatalf("rename path %q was not protected: %#v", path, dirty)
		}
	}
}

func TestParsePorcelainV1ZRejectsTruncatedRename(t *testing.T) {
	if _, err := parsePorcelainV1Z([]byte("R  new.txt\x00")); err == nil {
		t.Fatal("expected truncated rename record to fail closed")
	}
}

func TestDirtyPathsProtectsNestedRepositoryRoot(t *testing.T) {
	repo := t.TempDir()
	runGit(t, repo, "init", "-q")
	nested := filepath.Join(repo, "nested")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(nested, "owned.txt")
	if err := os.WriteFile(path, []byte("user work"), 0o644); err != nil {
		t.Fatal(err)
	}

	dirty, err := DirtyPaths(nested)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := dirty["nested/owned.txt"]; !ok {
		t.Fatalf("nested dirty path was not protected: %#v", dirty)
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, output)
	}
}

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
