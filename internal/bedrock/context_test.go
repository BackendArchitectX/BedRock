package bedrock

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSnapshotExcludesLikelySecretFiles(t *testing.T) {
	root := t.TempDir()
	files := map[string]string{
		"main.go":                 "package main\n",
		".env":                    "API_KEY=secret\n",
		".env.local":              "TOKEN=secret\n",
		"credentials.json":        `{"token":"secret"}`,
		"service-account.json":    `{"private_key":"secret"}`,
		"tls/private.pem":         "secret",
		"tls/private.key":         "secret",
		".ssh/id_ed25519":         "secret",
		"config/application.yaml": "server: local\n",
	}
	for name, content := range files {
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	got, err := Snapshot(root, "inspect application configuration", 100, 1024*1024)
	if err != nil {
		t.Fatal(err)
	}
	paths := make(map[string]bool, len(got))
	for _, file := range got {
		paths[file.Path] = true
	}
	for _, safe := range []string{"main.go", "config/application.yaml"} {
		if !paths[safe] {
			t.Errorf("safe context file %q was unexpectedly excluded", safe)
		}
	}
	for _, secret := range []string{".env", ".env.local", "credentials.json", "service-account.json", "tls/private.pem", "tls/private.key", ".ssh/id_ed25519"} {
		if paths[secret] {
			t.Errorf("secret-like file %q leaked into provider context", secret)
		}
	}
}

func TestSensitiveContextPathDoesNotBlockOrdinarySource(t *testing.T) {
	for _, name := range []string{"credentials.go", "key.go", "environment.go", "monkey.go"} {
		if sensitiveContextPath(name) {
			t.Errorf("ordinary source file %q was classified as secret", name)
		}
	}
}
