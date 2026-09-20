package bedrock

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestOpenAICompatibleProviderExecute(t *testing.T) {
	var authorization string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"{\"summary\":\"done\",\"changes\":[{\"path\":\"result.txt\",\"content\":\"good\\n\"}]}"}}]}`))
	}))
	defer server.Close()

	provider := OpenAICompatibleProvider{Endpoint: server.URL + "/v1/chat/completions", Model: "local-model", APIKey: "secret"}
	response, err := provider.Execute(context.Background(), ProviderRequest{Task: "write result"})
	if err != nil {
		t.Fatal(err)
	}
	if authorization != "Bearer secret" {
		t.Fatalf("authorization = %q", authorization)
	}
	if response.Summary != "done" || len(response.Changes) != 1 || response.Changes[0].Path != "result.txt" {
		t.Fatalf("unexpected response: %#v", response)
	}
}

func TestOpenAICompatibleProviderRedactsCredentialFromHTTPError(t *testing.T) {
	const secret = "top-secret-token"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("rejected " + secret))
	}))
	defer server.Close()

	provider := OpenAICompatibleProvider{Endpoint: server.URL, Model: "local-model", APIKey: secret}
	_, err := provider.Execute(context.Background(), ProviderRequest{Task: "x"})
	if err == nil || strings.Contains(err.Error(), secret) || !strings.Contains(err.Error(), "[REDACTED]") {
		t.Fatalf("expected redacted HTTP error, got %v", err)
	}
}

func TestOpenAICompatibleProviderPreservesCallerCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	provider := OpenAICompatibleProvider{Endpoint: server.URL, Model: "local-model", Timeout: time.Minute}
	_, err := provider.Execute(ctx, ProviderRequest{Task: "x"})
	if !errors.Is(err, context.Canceled) || !strings.Contains(err.Error(), "provider canceled") {
		t.Fatalf("expected caller cancellation, got %v", err)
	}
}

func TestOpenAICompatibleProviderRejectsUnsafeConfiguration(t *testing.T) {
	tests := []OpenAICompatibleProvider{
		{Endpoint: "file:///tmp/model", Model: "model"},
		{Endpoint: "http://user:pass@localhost/v1/chat/completions", Model: "model"},
		{Endpoint: "http://localhost/v1/chat/completions"},
	}
	for _, provider := range tests {
		if _, err := provider.Execute(context.Background(), ProviderRequest{Task: "x"}); err == nil {
			t.Fatalf("expected configuration error for %#v", provider)
		}
	}
}

func TestOpenAICompatibleProviderBoundsResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(strings.Repeat("x", 64)))
	}))
	defer server.Close()
	provider := OpenAICompatibleProvider{Endpoint: server.URL, Model: "model", MaxResponse: 16}
	_, err := provider.Execute(context.Background(), ProviderRequest{Task: "x"})
	if err == nil || !strings.Contains(err.Error(), "exceeded 16 bytes") {
		t.Fatalf("expected bounded response error, got %v", err)
	}
}
