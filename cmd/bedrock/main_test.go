package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
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

func TestRunWithBuiltInHTTPProvider(t *testing.T) {
	const (
		apiKey = "integration-secret"
		model  = "local-model"
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer "+apiKey {
			t.Errorf("Authorization = %q", got)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		var request struct {
			Model    string `json:"model"`
			Messages []struct {
				Content string `json:"content"`
			} `json:"messages"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Errorf("decode request: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if request.Model != model {
			t.Errorf("model = %q", request.Model)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if len(request.Messages) != 2 {
			t.Errorf("messages = %d", len(request.Messages))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		var providerRequest struct {
			VerificationCommands []string `json:"verificationCommands"`
		}
		if err := json.Unmarshal([]byte(request.Messages[1].Content), &providerRequest); err != nil {
			t.Errorf("decode provider context: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if len(providerRequest.VerificationCommands) != 1 || providerRequest.VerificationCommands[0] != "git diff --check" {
			t.Errorf("verification commands = %#v", providerRequest.VerificationCommands)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"{\"summary\":\"done\",\"changes\":[{\"path\":\"result.txt\",\"content\":\"good\\n\"}]}"}}]}`))
	}))
	defer server.Close()

	repo := t.TempDir()
	git := exec.Command("git", "init", "--quiet", repo)
	if output, err := git.CombinedOutput(); err != nil {
		t.Fatalf("git init: %v: %s", err, output)
	}
	t.Setenv("BEDROCK_TEST_API_KEY", apiKey)

	err := run(context.Background(), []string{
		"--repo", repo,
		"--task", "write result",
		"--provider-endpoint", server.URL + "/v1/chat/completions",
		"--provider-model", model,
		"--provider-api-key-env", "BEDROCK_TEST_API_KEY",
		"--max-attempts", "1",
		"--verify", "git diff --check",
	})
	if err != nil {
		t.Fatal(err)
	}
	result, err := os.ReadFile(filepath.Join(repo, "result.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(result) != "good\n" {
		t.Fatalf("result.txt = %q", result)
	}
}
