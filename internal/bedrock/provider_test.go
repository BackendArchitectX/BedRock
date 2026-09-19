package bedrock

import (
	"strings"
	"testing"
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
