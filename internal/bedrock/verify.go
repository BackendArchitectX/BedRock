package bedrock

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

type ShellVerifier struct {
	Commands  []string
	Timeout   time.Duration
	MaxOutput int
}

func (v ShellVerifier) Verify(ctx context.Context, root string) ([]VerificationResult, error) {
	if len(v.Commands) == 0 {
		return nil, nil
	}
	timeout := v.Timeout
	if timeout <= 0 {
		timeout = 3 * time.Minute
	}
	maxOutput := v.MaxOutput
	if maxOutput <= 0 {
		maxOutput = 256 * 1024
	}
	secretValues := sensitiveEnvironmentValues()

	results := make([]VerificationResult, 0, len(v.Commands))
	for _, command := range v.Commands {
		runCtx, cancel := context.WithTimeout(ctx, timeout)
		cmd := shellCommand(runCtx, command)
		cmd.Dir = root
		var output cappedBuffer
		output.limit = maxOutput
		cmd.Stdout = &output
		cmd.Stderr = &output
		err := cmd.Run()
		ctxErr := runCtx.Err()
		cancel()

		exitCode := 0
		if err != nil {
			exitCode = -1
			if exitErr, ok := err.(*exec.ExitError); ok {
				exitCode = exitErr.ExitCode()
			}
		}
		result := VerificationResult{Command: command, ExitCode: exitCode, Output: redactValues(output.String(), secretValues)}
		results = append(results, result)
		if err != nil {
			if errors.Is(ctxErr, context.DeadlineExceeded) {
				return results, fmt.Errorf("verification command %q timed out: %w", command, ctxErr)
			}
			return results, fmt.Errorf("verification command %q failed with exit code %d", command, exitCode)
		}
	}
	return results, nil
}

func sensitiveEnvironmentValues() []string {
	var values []string
	for _, entry := range os.Environ() {
		key, value, ok := strings.Cut(entry, "=")
		if !ok || value == "" || !sensitiveEnvironmentKey(strings.ToUpper(key)) {
			continue
		}
		values = append(values, value)
	}
	return values
}

func sensitiveEnvironmentKey(key string) bool {
	for _, marker := range []string{"TOKEN", "SECRET", "PASSWORD", "PASSWD", "API_KEY", "APIKEY", "PRIVATE_KEY", "ACCESS_KEY", "CREDENTIAL"} {
		if strings.Contains(key, marker) {
			return true
		}
	}
	return false
}

func shellCommand(ctx context.Context, command string) *exec.Cmd {
	if runtime.GOOS == "windows" {
		return exec.CommandContext(ctx, "cmd.exe", "/S", "/C", command)
	}
	return exec.CommandContext(ctx, "sh", "-c", command)
}
