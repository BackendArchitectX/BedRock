package bedrock

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type scriptedProvider struct {
	responses []ProviderResponse
	calls     int
}

func (p *scriptedProvider) Name() string { return "scripted-test" }
func (p *scriptedProvider) Execute(_ context.Context, _ ProviderRequest) (ProviderResponse, error) {
	if p.calls >= len(p.responses) {
		return ProviderResponse{}, errors.New("no scripted response")
	}
	response := p.responses[p.calls]
	p.calls++
	return response, nil
}

type mutatingProvider struct {
	root string
}

func (p mutatingProvider) Name() string { return "mutating-test" }
func (p mutatingProvider) Execute(_ context.Context, _ ProviderRequest) (ProviderResponse, error) {
	if err := os.WriteFile(filepath.Join(p.root, "result.txt"), []byte("external edit"), 0o644); err != nil {
		return ProviderResponse{}, err
	}
	return ProviderResponse{Changes: []FileChange{{Path: "result.txt", Content: "provider edit"}}}, nil
}

type blockingProvider struct {
	started chan struct{}
}

func (p blockingProvider) Name() string { return "blocking-test" }
func (p blockingProvider) Execute(ctx context.Context, _ ProviderRequest) (ProviderResponse, error) {
	close(p.started)
	<-ctx.Done()
	return ProviderResponse{}, ctx.Err()
}

type fileContentVerifier struct {
	path string
	want string
}

func (v fileContentVerifier) Verify(_ context.Context, root string) ([]VerificationResult, error) {
	data, err := os.ReadFile(filepath.Join(root, v.path))
	if err != nil {
		return []VerificationResult{{Command: "content-check", ExitCode: 1, Output: err.Error()}}, err
	}
	if string(data) != v.want {
		return []VerificationResult{{Command: "content-check", ExitCode: 1, Output: "unexpected content"}}, errors.New("content verification failed")
	}
	return []VerificationResult{{Command: "content-check", ExitCode: 0, Output: "ok"}}, nil
}

type inconsistentVerifier struct{}

func (inconsistentVerifier) Verify(_ context.Context, _ string) ([]VerificationResult, error) {
	return []VerificationResult{{Command: "broken-contract", ExitCode: 1, Output: "failed"}}, nil
}

type concurrentEditVerifier struct {
	path string
}

func (v concurrentEditVerifier) Verify(_ context.Context, root string) ([]VerificationResult, error) {
	if err := os.WriteFile(filepath.Join(root, v.path), []byte("user edit during verification"), 0o644); err != nil {
		return nil, err
	}
	return []VerificationResult{{Command: "failing-check", ExitCode: 1, Output: "failed"}}, errors.New("verification failed")
}

func TestEngineRepairsAfterFailedVerification(t *testing.T) {
	root := t.TempDir()
	provider := &scriptedProvider{responses: []ProviderResponse{
		{Summary: "first attempt", Changes: []FileChange{{Path: "result.txt", Content: "bad"}}},
		{Summary: "repair", Changes: []FileChange{{Path: "result.txt", Content: "good"}}},
	}}
	engine := Engine{Provider: provider, Verifier: fileContentVerifier{path: "result.txt", want: "good"}, MaxAttempts: 2}
	result, err := engine.Run(context.Background(), root, "make result good")
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
	if result.Evidence.Status != "VERIFIED" || provider.calls != 2 {
		t.Fatalf("status=%s provider calls=%d", result.Evidence.Status, provider.calls)
	}
	data, err := os.ReadFile(filepath.Join(root, "result.txt"))
	if err != nil || string(data) != "good" {
		t.Fatalf("final content=%q err=%v", data, err)
	}
}

func TestEngineRollsBackWhenVerificationNeverPasses(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "result.txt")
	if err := os.WriteFile(path, []byte("original"), 0o644); err != nil {
		t.Fatal(err)
	}
	provider := &scriptedProvider{responses: []ProviderResponse{
		{Changes: []FileChange{{Path: "result.txt", Content: "bad-1"}}},
		{Changes: []FileChange{{Path: "result.txt", Content: "bad-2"}}},
	}}
	engine := Engine{Provider: provider, Verifier: fileContentVerifier{path: "result.txt", want: "good"}, MaxAttempts: 2}
	result, err := engine.Run(context.Background(), root, "make result good")
	if err == nil || !result.Evidence.RolledBack {
		t.Fatalf("expected failed verified rollback, err=%v evidence=%+v", err, result.Evidence)
	}
	data, readErr := os.ReadFile(path)
	if readErr != nil || string(data) != "original" {
		t.Fatalf("rollback content=%q err=%v", data, readErr)
	}
}

func TestEngineRejectsNonzeroVerificationResultWithoutError(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "result.txt")
	if err := os.WriteFile(path, []byte("original"), 0o644); err != nil {
		t.Fatal(err)
	}
	provider := &scriptedProvider{responses: []ProviderResponse{{Changes: []FileChange{{Path: "result.txt", Content: "provider-change"}}}}}
	engine := Engine{Provider: provider, Verifier: inconsistentVerifier{}, MaxAttempts: 1}
	result, err := engine.Run(context.Background(), root, "change result")
	if err == nil || result.Evidence.Status == "VERIFIED" || !result.Evidence.RolledBack {
		t.Fatalf("unexpected result err=%v evidence=%+v", err, result.Evidence)
	}
	data, readErr := os.ReadFile(path)
	if readErr != nil || string(data) != "original" {
		t.Fatalf("rollback content=%q err=%v", data, readErr)
	}
}

func TestEngineReportsRollbackConflictWithoutOverwritingConcurrentEdit(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "result.txt")
	if err := os.WriteFile(path, []byte("original"), 0o644); err != nil {
		t.Fatal(err)
	}
	provider := &scriptedProvider{responses: []ProviderResponse{{Changes: []FileChange{{Path: "result.txt", Content: "provider-change"}}}}}
	engine := Engine{Provider: provider, Verifier: concurrentEditVerifier{path: "result.txt"}, MaxAttempts: 1}
	result, err := engine.Run(context.Background(), root, "change result")
	if err == nil || !strings.Contains(err.Error(), "rollback failed") {
		t.Fatalf("expected rollback conflict in run error, got %v", err)
	}
	if result.Evidence.RolledBack {
		t.Fatal("rollback conflict was incorrectly recorded as successful rollback")
	}
	if !strings.Contains(result.Evidence.LastFailure, "rollback failed") || !strings.Contains(result.Evidence.LastFailure, "changed after BedRock wrote it") {
		t.Fatalf("rollback conflict missing from evidence: %q", result.Evidence.LastFailure)
	}
	data, readErr := os.ReadFile(path)
	if readErr != nil || string(data) != "user edit during verification" {
		t.Fatalf("concurrent edit=%q err=%v", data, readErr)
	}
}

func TestEngineRejectsTargetChangedDuringProviderExecution(t *testing.T) {
	root := t.TempDir()
	runGit(t, root, "init", "-q")
	path := filepath.Join(root, "result.txt")
	if err := os.WriteFile(path, []byte("original"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, root, "add", "result.txt")
	runGit(t, root, "-c", "user.name=BedRock Test", "-c", "user.email=bedrock@example.invalid", "commit", "-qm", "fixture")

	engine := Engine{Provider: mutatingProvider{root: root}, MaxAttempts: 1}
	result, err := engine.Run(context.Background(), root, "change result")
	if err == nil || !strings.Contains(err.Error(), "refusing to overwrite pre-existing dirty path") {
		t.Fatalf("expected concurrent mutation rejection, got %v", err)
	}
	if !result.Evidence.RolledBack {
		t.Fatal("expected no-op rollback to complete")
	}
	data, readErr := os.ReadFile(path)
	if readErr != nil || string(data) != "external edit" {
		t.Fatalf("external edit was overwritten: %q err=%v", data, readErr)
	}
}

func TestEnginePropagatesCancellationToProvider(t *testing.T) {
	root := t.TempDir()
	started := make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())
	engine := Engine{Provider: blockingProvider{started: started}, MaxAttempts: 1}

	done := make(chan error, 1)
	go func() {
		_, err := engine.Run(ctx, root, "wait for cancellation")
		done <- err
	}()

	select {
	case <-started:
		cancel()
	case <-time.After(2 * time.Second):
		t.Fatal("provider did not start")
	}

	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("run error=%v, want context cancellation", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("engine did not stop after cancellation")
	}
}
