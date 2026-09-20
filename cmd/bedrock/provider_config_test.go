package main

import (
	"strings"
	"testing"
	"time"

	"github.com/BackendArchitectX/BedRock/internal/bedrock"
)

func TestConfiguredProviderHTTP(t *testing.T) {
	t.Setenv("BEDROCK_TEST_API_KEY", "secret")
	provider, err := configuredProvider("", nil, nil, "http://localhost:11434/v1/chat/completions", "local-model", "BEDROCK_TEST_API_KEY", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	httpProvider, ok := provider.(bedrock.OpenAICompatibleProvider)
	if !ok {
		t.Fatalf("provider type = %T", provider)
	}
	if httpProvider.APIKey != "secret" || httpProvider.Model != "local-model" {
		t.Fatalf("unexpected HTTP provider configuration: %#v", httpProvider)
	}
}

func TestConfiguredProviderRejectsAmbiguousOrMissingHTTPConfig(t *testing.T) {
	tests := []struct {
		name, bin, endpoint, model, keyEnv, want string
	}{
		{"missing provider", "", "", "", "", "one of --provider-bin or --provider-endpoint is required"},
		{"ambiguous provider", "adapter", "http://localhost/v1/chat/completions", "model", "", "mutually exclusive"},
		{"missing model", "", "http://localhost/v1/chat/completions", "", "", "--provider-model is required"},
		{"orphan model", "adapter", "", "model", "", "require --provider-endpoint"},
		{"missing key", "", "http://localhost/v1/chat/completions", "model", "BEDROCK_TEST_MISSING_KEY", "is not set"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := configuredProvider(tt.bin, nil, nil, tt.endpoint, tt.model, tt.keyEnv, time.Second)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("expected error containing %q, got %v", tt.want, err)
			}
		})
	}
}
