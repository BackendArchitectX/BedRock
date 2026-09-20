package bedrock

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type retryMutatingProvider struct {
	root  string
	calls int
}

func (p *retryMutatingProvider) Name() string { return "retry-mutating-test" }

func (p *retryMutatingProvider) Execute(_ context.Context, _ ProviderRequest) (ProviderResponse, error) {
	p.calls++
	if p.calls == 1 {
		return ProviderResponse{Changes: []FileChange{{Path: "result.txt", Content: "bad"}}}, nil
	}
	if err := os.WriteFile(filepath.Join(p.root, "result.txt"), []byte("user edit between attempts"), 0o644); err != nil {
		return ProviderResponse{}, err
	}
	return ProviderResponse{Changes: []FileChange{{Path: "result.txt", Content: "good"}}}, nil
}

func TestEnginePreservesTargetChangedBetweenRepairAttempts(t *testing.T) {
	root := t.TempDir()
	runGit(t, root, "init", "-q")
	path := filepath.Join(root, "result.txt")
	if err := os.WriteFile(path, []byte("original"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, root, "add", "result.txt")
	runGit(t, root, "-c", "user.name=BedRock Test", "-c", "user.email=bedrock@example.invalid", "commit", "-qm", "fixture")

	provider := &retryMutatingProvider{root: root}
	engine := Engine{Provider: provider, Verifier: fileContentVerifier{path: "result.txt", want: "good"}, MaxAttempts: 2}
	result, err := engine.Run(context.Background(), root, "make result good")
	if err == nil || !strings.Contains(err.Error(), "changed after BedRock wrote it") {
		t.Fatalf("expected ownership conflict, got %v", err)
	}
	if !strings.Contains(err.Error(), "rollback failed") {
		t.Fatalf("expected rollback conflict to be reported, got %v", err)
	}
	if result.Evidence.RolledBack {
		t.Fatal("rollback conflict was incorrectly recorded as successful")
	}
	data, readErr := os.ReadFile(path)
	if readErr != nil || string(data) != "user edit between attempts" {
		t.Fatalf("user edit was overwritten: %q err=%v", data, readErr)
	}
}
