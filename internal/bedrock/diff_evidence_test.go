package bedrock

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestDiffHashIsStableAndContentSensitive(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "result.txt")
	if err := os.WriteFile(path, []byte("before"), 0o644); err != nil {
		t.Fatal(err)
	}

	first := NewChangeSet()
	if err := first.Apply(root, []FileChange{{Path: "result.txt", Content: "after"}}, nil); err != nil {
		t.Fatal(err)
	}
	firstHash := first.DiffHash()
	if firstHash == "" {
		t.Fatal("changed content produced empty diff hash")
	}
	if err := first.Rollback(root); err != nil {
		t.Fatal(err)
	}

	second := NewChangeSet()
	if err := second.Apply(root, []FileChange{{Path: "result.txt", Content: "after"}}, nil); err != nil {
		t.Fatal(err)
	}
	if second.DiffHash() != firstHash {
		t.Fatalf("same mutation produced unstable hashes: %q != %q", second.DiffHash(), firstHash)
	}
	if err := second.Rollback(root); err != nil {
		t.Fatal(err)
	}

	third := NewChangeSet()
	if err := third.Apply(root, []FileChange{{Path: "result.txt", Content: "different"}}, nil); err != nil {
		t.Fatal(err)
	}
	if third.DiffHash() == firstHash {
		t.Fatal("different mutation produced the same diff hash")
	}
}

func TestEnginePersistsDiffHashWithoutRawContent(t *testing.T) {
	root := t.TempDir()
	provider := &scriptedProvider{responses: []ProviderResponse{{
		Summary: "write result",
		Changes: []FileChange{{Path: "result.txt", Content: "sensitive-patch-content"}},
	}}}
	engine := Engine{Provider: provider, MaxAttempts: 1}
	result, err := engine.Run(context.Background(), root, "write result")
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
	if result.Evidence.DiffHash == "" {
		t.Fatal("evidence omitted diff hash for a changed path")
	}
	if len(result.Evidence.ChangedPaths) != 1 || result.Evidence.ChangedPaths[0] != "result.txt" {
		t.Fatalf("changed paths=%v", result.Evidence.ChangedPaths)
	}
	data, err := os.ReadFile(result.EvidencePath)
	if err != nil {
		t.Fatal(err)
	}
	if containsBytes(data, []byte("sensitive-patch-content")) {
		t.Fatal("persisted evidence leaked raw changed content")
	}
}

func containsBytes(haystack, needle []byte) bool {
	if len(needle) == 0 {
		return true
	}
	for i := 0; i+len(needle) <= len(haystack); i++ {
		match := true
		for j := range needle {
			if haystack[i+j] != needle[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
