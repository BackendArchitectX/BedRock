package bedrock

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"
)

const defaultProviderOutputLimit = 4 * 1024 * 1024

type CommandProvider struct {
	Bin       string
	Args      []string
	Timeout   time.Duration
	MaxOutput int
	EnvAllow  []string
}

func (p CommandProvider) Name() string {
	return "command:" + p.Bin
}

func (p CommandProvider) Execute(ctx context.Context, req ProviderRequest) (ProviderResponse, error) {
	if p.Bin == "" {
		return ProviderResponse{}, errors.New("provider executable is required")
	}
	timeout := p.Timeout
	if timeout <= 0 {
		timeout = 2 * time.Minute
	}
	maxOutput := p.MaxOutput
	if maxOutput <= 0 {
		maxOutput = defaultProviderOutputLimit
	}

	payload, err := json.Marshal(req)
	if err != nil {
		return ProviderResponse{}, fmt.Errorf("encode provider request: %w", err)
	}

	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(runCtx, p.Bin, p.Args...)
	cmd.Stdin = bytes.NewReader(payload)
	cmd.Env, secretValues := providerEnvironment(p.EnvAllow)
	var stdout, stderr cappedBuffer
	stdout.limit = maxOutput
	stderr.limit = 64 * 1024
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if runCtx.Err() != nil {
			return ProviderResponse{}, fmt.Errorf("provider timed out: %w", runCtx.Err())
		}
		return ProviderResponse{}, fmt.Errorf("provider failed: %w: %s", err, redactValues(stderr.String(), secretValues))
	}
	if stdout.truncated {
		return ProviderResponse{}, fmt.Errorf("provider response exceeded %d bytes", maxOutput)
	}

	var response ProviderResponse
	dec := json.NewDecoder(bytes.NewReader(stdout.Bytes()))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&response); err != nil {
		return ProviderResponse{}, fmt.Errorf("decode provider response: %w", err)
	}
	var extra any
	if err := dec.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return ProviderResponse{}, errors.New("provider returned multiple JSON values")
		}
		return ProviderResponse{}, fmt.Errorf("provider returned trailing data: %w", err)
	}
	return response, nil
}

func providerEnvironment(allow []string) ([]string, []string) {
	allowed := make(map[string]struct{}, len(allow))
	for _, name := range allow {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		allowed[strings.ToUpper(name)] = struct{}{}
	}

	var env []string
	var secretValues []string
	for _, entry := range os.Environ() {
		key, value, ok := strings.Cut(entry, "=")
		if !ok {
			continue
		}
		upper := strings.ToUpper(key)
		_, explicitlyAllowed := allowed[upper]
		if !safeProviderEnvironmentKey(upper) && !explicitlyAllowed {
			continue
		}
		env = append(env, entry)
		if explicitlyAllowed && value != "" {
			secretValues = append(secretValues, value)
		}
	}
	sort.Strings(env)
	return env, secretValues
}

func safeProviderEnvironmentKey(key string) bool {
	switch key {
	case "PATH", "PATHEXT", "SYSTEMROOT", "COMSPEC", "WINDIR", "TEMP", "TMP", "TMPDIR", "LANG", "LC_ALL", "LC_CTYPE":
		return true
	default:
		return false
	}
}

func redactValues(text string, values []string) string {
	for _, value := range values {
		if value == "" {
			continue
		}
		text = strings.ReplaceAll(text, value, "[REDACTED]")
	}
	return text
}

type cappedBuffer struct {
	buf       bytes.Buffer
	limit     int
	truncated bool
}

func (b *cappedBuffer) Write(p []byte) (int, error) {
	original := len(p)
	if b.limit <= 0 {
		b.truncated = true
		return original, nil
	}
	remaining := b.limit - b.buf.Len()
	if remaining <= 0 {
		b.truncated = true
		return original, nil
	}
	if len(p) > remaining {
		_, _ = b.buf.Write(p[:remaining])
		b.truncated = true
		return original, nil
	}
	_, _ = b.buf.Write(p)
	return original, nil
}

func (b *cappedBuffer) Bytes() []byte { return b.buf.Bytes() }
func (b *cappedBuffer) String() string { return b.buf.String() }
