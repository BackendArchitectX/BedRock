package bedrock

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSnapshotExcludesLikelySecretFiles(t *testing.T) {
	root := t.TempDir()
	files := map[string]string{
		"main.go":                              "package main\n",
		".env":                                 "API_KEY=secret\n",
		".env.local":                           "TOKEN=secret\n",
		".envrc":                               "export TOKEN=secret\n",
		".git-credentials":                     "https://user:secret@example.com\n",
		"credentials":                          "aws_secret_access_key=secret\n",
		"credentials.json":                     `{"token":"secret"}`,
		"application_default_credentials.json": `{"private_key":"secret"}`,
		"service-account.json":                 `{"private_key":"secret"}`,
		"deploy/secrets.yaml":                  "token: secret\n",
		"infra/terraform.tfstate":              `{"secret":"value"}`,
		"infra/terraform.tfstate.backup":       `{"secret":"old-value"}`,
		"tls/private.pem":                      "secret",
		"tls/private.key":                      "secret",
		".ssh/id_ed25519":                      "secret",
		".aws/credentials":                     "aws_secret_access_key=secret\n",
		".azure/accessTokens.json":             `{"token":"secret"}`,
		".docker/config.json":                  `{"auths":{"registry":{"auth":"secret"}}}`,
		".kube/config":                         "token: secret\n",
		".gnupg/private-keys-v1.d/key":         "secret",
		"config/application.yaml":              "server: local\n",
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
	for _, secret := range []string{
		".env", ".env.local", ".envrc", ".git-credentials", "credentials", "credentials.json",
		"application_default_credentials.json", "service-account.json", "deploy/secrets.yaml",
		"infra/terraform.tfstate", "infra/terraform.tfstate.backup", "tls/private.pem", "tls/private.key",
		".ssh/id_ed25519", ".aws/credentials", ".azure/accessTokens.json", ".docker/config.json",
		".kube/config", ".gnupg/private-keys-v1.d/key",
	} {
		if paths[secret] {
			t.Errorf("secret-like file %q leaked into provider context", secret)
		}
	}
}

func TestSensitiveContextPathDoesNotBlockOrdinarySource(t *testing.T) {
	for _, name := range []string{"credentials.go", "key.go", "environment.go", "monkey.go", "secrets.go", "terraform.tf"} {
		if sensitiveContextPath(name) {
			t.Errorf("ordinary source file %q was classified as secret", name)
		}
	}
}
