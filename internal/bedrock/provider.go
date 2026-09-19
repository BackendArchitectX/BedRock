package bedrock

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"time"
)

const defaultProviderOutputLimit = 4 * 1024 * 1024

type CommandProvider struct {
	Bin       string
	Args      []string
	Timeout   time.Duration
	MaxOutput int
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
	var stdout, stderr cappedBuffer
	stdout.limit = maxOutput
	stderr.limit = 64 * 1024
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if runCtx.Err() != nil {
			return ProviderResponse{}, fmt.Errorf("provider timed out: %w", runCtx.Err())
		}
		return ProviderResponse{}, fmt.Errorf("provider failed: %w: %s", err, stderr.String())
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
