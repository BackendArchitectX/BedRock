package bedrock

import (
	"context"
	"errors"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestProviderEnvironmentOnlyPassesSafeAndExplicitVariables(t *testing.T) {
	t.Setenv("BEDROCK_ALLOWED_SECRET", "super-secret-value")
	t.Setenv("BEDROCK_BLOCKED_SECRET", "must-not-leak")

	env, secrets := providerEnvironment([]string{"BEDROCK_ALLOWED_SECRET"})

	joined := strings.Join(env, "\n")
	if !strings.Contains(joined, "BEDROCK_ALLOWED_SECRET=super-secret-value") {
		t.Fatal("explicitly allowed provider environment variable was not passed")
	}
	if strings.Contains(joined, "BEDROCK_BLOCKED_SECRET=must-not-leak") {
		t.Fatal("unapproved provider environment variable leaked to provider")
	}
	if len(secrets) != 1 || secrets[0] != "super-secret-value" {
		t.Fatalf("secret values=%q, want explicit allowed value", secrets)
	}
}

func TestRedactValuesRemovesExplicitProviderSecrets(t *testing.T) {
	got := redactValues("provider failed with token super-secret-value", []string{"super-secret-value"})
	if strings.Contains(got, "super-secret-value") {
		t.Fatalf("secret remained in output: %q", got)
	}
	if !strings.Contains(got, "[REDACTED]") {
		t.Fatalf("redaction marker missing: %q", got)
	}
}

func TestCommandProviderDistinguishesCallerCancellation(t *testing.T) {
	var bin string
	var args []string
	if runtime.GOOS == "windows" {
		bin = os.Getenv("COMSPEC")
		if bin == "" {
			bin = "cmd.exe"
		}
		args = []string{"/S", "/C", "ping -n 30 127.0.0.1 >NUL"}
	} else {
		bin = "sh"
		args = []string{"-c", "sleep 30"}
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := (CommandProvider{Bin: bin, Args: args, Timeout: time.Minute}).Execute(ctx, ProviderRequest{})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error=%v, want context.Canceled", err)
	}
	if err == nil || !strings.Contains(err.Error(), "provider canceled") {
		t.Fatalf("error=%v, want cancellation diagnostic", err)
	}
	if strings.Contains(err.Error(), "timed out") {
		t.Fatalf("caller cancellation mislabeled as timeout: %v", err)
	}
}
