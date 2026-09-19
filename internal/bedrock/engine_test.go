package bedrock

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
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

func TestEngineRepairsAfterFailedVerification(t *testing.T) {
	root := t.TempDir()
	provider := &scriptedProvider{responses: []ProviderResponse{
		{Summary: "first attempt", Changes: []FileChange{{Path: "result.txt", Content: "bad"}}},
		{Summary: "repair", Changes: []FileChange{{Path: "result.txt", Content: "good"}}},
	}}
	engine := Engine{
		Provider:    provider,
		Verifier:    fileContentVerifier{path: "result.txt", want: "good"},
		MaxAttempts: 2,
	}
	result, err := engine.Run(context.Background(), root, "make result good")
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
	if result.Evidence.Status != "VERIFIED" {
		t.Fatalf("status=%s", result.Evidence.Status)
	}
	if provider.calls != 2 {
		t.Fatalf("provider calls=%d, want 2", provider.calls)
	}
	data, err := os.ReadFile(filepath.Join(root, "result.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "good" {
		t.Fatalf("final content=%q", data)
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
	engine := Engine{
		Provider:    provider,
		Verifier:    fileContentVerifier{path: "result.txt", want: "good"},
		MaxAttempts: 2,
	}
	result, err := engine.Run(context.Background(), root, "make result good")
	if err == nil {
		t.Fatal("expected verification failure")
	}
	if !result.Evidence.RolledBack {
		t.Fatal("expected rollback evidence")
	}
	data, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(data) != "original" {
		t.Fatalf("rollback content=%q", data)
	}
}
