package bedrock

import (
	"context"
	"strings"
	"testing"
)

func TestSensitiveEnvironmentKey(t *testing.T) {
	for _, key := range []string{"OPENAI_API_KEY", "GITHUB_TOKEN", "DB_PASSWORD", "CLIENT_SECRET", "AWS_SECRET_ACCESS_KEY"} {
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
