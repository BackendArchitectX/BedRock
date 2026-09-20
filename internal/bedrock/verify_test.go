package bedrock

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestSensitiveEnvironmentKey(t *testing.T) {
	for _, key := range []string{
		"OPENAI_API_KEY", "GITHUB_TOKEN", "DB_PASSWORD", "DB_PASS", "CLIENT_SECRET",
		"AWS_SECRET_ACCESS_KEY", "DATABASE_URL", "DATABASE_URI", "REDIS_CONNECTION_STRING",
	} {
		if !sensitiveEnvironmentKey(key) {
			t.Fatalf("expected %q to be sensitive", key)
		}
	}
	for _, key := range []string{"PATH", "HOME", "LANG", "GOPATH"} {
		if sensitiveEnvironmentKey(key) {
			t.Fatalf("did not expect %q to be sensitive", key)
		}
	}
}

func TestShellVerifierRedactsSecretEnvironmentOutput(t *testing.T) {
	const secret = "bedrock-verifier-secret-9f42"
	t.Setenv("BEDROCK_TEST_API_KEY", secret)

	var command string
	if shellCommand(context.Background(), "").Path == "cmd.exe" {
		command = "echo %BEDROCK_TEST_API_KEY%"
	} else {
		command = "printf '%s' \"$BEDROCK_TEST_API_KEY\""
	}
	results, err := (ShellVerifier{Commands: []string{command}}).Verify(context.Background(), t.TempDir())
	if err != nil {
		t.Fatalf("verify failed: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("results=%d, want 1", len(results))
	}
	if strings.Contains(results[0].Output, secret) {
		t.Fatalf("secret leaked in verification output: %q", results[0].Output)
	}
	if !strings.Contains(results[0].Output, "[REDACTED]") {
		t.Fatalf("redaction marker missing: %q", results[0].Output)
	}
}

func TestShellVerifierRedactsSecretFromCommandAndError(t *testing.T) {
	const secret = "bedrock-command-secret-7a31"
	t.Setenv("BEDROCK_TEST_TOKEN", secret)

	command := "echo " + secret
	if shellCommand(context.Background(), "").Path == "cmd.exe" {
		command += " && exit /b 1"
	} else {
		command += "; false"
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
