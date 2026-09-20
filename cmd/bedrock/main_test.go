package main

import (
	"strings"
	"testing"
	"time"
)

func TestConfiguredProviderRejectsCommandOptionsWithHTTPEndpoint(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		args []string
		env  []string
	}{
		{name: "argument", args: []string{"--unsafe"}},
		{name: "environment", env: []string{"TOKEN"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := configuredProvider("", tc.args, tc.env, "http://127.0.0.1:11434/v1/chat/completions", "local-model", "", time.Second)
			if err == nil {
				t.Fatal("configuredProvider accepted command-provider options with HTTP endpoint")
			}
			if !strings.Contains(err.Error(), "require --provider-bin") {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
