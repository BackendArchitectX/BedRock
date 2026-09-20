package bedrock

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestSensitiveEnvironmentKey(t *testing.T) {
	for _, key := range []string{"OPENAI_API_KEY", "GITHUB_TOKEN", "DB_PASSWORD", "DATABASE_URL", "CLIENT_SECRET"} {
		if !sensitiveEnvironmentKey(key) {
			t.Fatalf("expected %s to be sensitive", key)
		}
	}
	for _, key := range []string{"PATH", "HOME", "GOPATH", "LOG_LEVEL"} {
		if sensitiveEnvironmentKey(key) {
			t.Fatalf("did not expect %s to be sensitive", key)
		}
	}
}

func TestShellVerifierRedactsSensitiveEnvironmentValues(t *testing.T) {
	secret := "bedrock-test-secret-9f1a"
	t.Setenv("BEDROCK_TEST_API_KEY", secret)

	command := fmt.Sprintf("printf '%s'; exit 7", secret)
	if runtime.GOOS == "windows" {
		command = fmt.Sprintf("echo %s & exit /b 7", secret)
	}

	results, err := (ShellVerifier{Commands: []string{command}}).Verify(context.Background(), t.TempDir())
	if err == nil {
		t.Fatal("expected verification failure")
	}
	if len(results) != 1 {
		t.Fatalf("results=%d, want 1", len(results))
	}
	for label, text := range map[string]string{
		"command evidence": results[0].Command,
		"output evidence":  results[0].Output,
		"returned error":   err.Error(),
	} {
		if strings.Contains(text, secret) {
			t.Fatalf("secret leaked in %s: %q", label, text)
		}
		if !strings.Contains(text, "[REDACTED]") {
			t.Fatalf("redaction marker missing from %s: %q", label, text)
		}
	}
}

func TestShellVerifierPreservesCancellationCause(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	command := "sleep 30"
	if shellCommand(context.Background(), "").Path == "cmd.exe" {
		command = "ping -n 30 127.0.0.1 >NUL"
	}

	done := make(chan error, 1)
	go func() {
		_, err := (ShellVerifier{Commands: []string{command}, Timeout: time.Minute}).Verify(ctx, t.TempDir())
		done <- err
	}()

	time.Sleep(100 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("verify error=%v, want context cancellation", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("verifier did not stop after cancellation")
	}
}

func TestSensitiveEnvironmentValues(t *testing.T) {
	secret := "bedrock-test-secret-43b7"
	t.Setenv("BEDROCK_TEST_SECRET", secret)
	values := sensitiveEnvironmentValues()
	found := false
	for _, value := range values {
		if value == secret {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected sensitive environment value to be collected for redaction")
	}
}

func TestShellVerifierReturnsSuccessfulEvidence(t *testing.T) {
	command := "printf ok"
	if runtime.GOOS == "windows" {
		command = "echo ok"
	}
	results, err := (ShellVerifier{Commands: []string{command}}).Verify(context.Background(), t.TempDir())
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if len(results) != 1 || results[0].ExitCode != 0 {
		t.Fatalf("unexpected results: %#v", results)
	}
	if !strings.Contains(results[0].Output, "ok") {
		t.Fatalf("output=%q, want ok", results[0].Output)
	}
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
