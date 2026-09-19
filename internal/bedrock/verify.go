package bedrock

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
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
		cancel()

		exitCode := 0
		if err != nil {
			exitCode = -1
			if exitErr, ok := err.(*exec.ExitError); ok {
				exitCode = exitErr.ExitCode()
			}
		}
		result := VerificationResult{Command: command, ExitCode: exitCode, Output: output.String()}
		results = append(results, result)
		if err != nil {
			if runCtx.Err() != nil {
				return results, fmt.Errorf("verification command %q timed out: %w", command, runCtx.Err())
			}
			return results, fmt.Errorf("verification command %q failed with exit code %d", command, exitCode)
		}
	}
	return results, nil
}

func shellCommand(ctx context.Context, command string) *exec.Cmd {
	if runtime.GOOS == "windows" {
		return exec.CommandContext(ctx, "cmd.exe", "/S", "/C", command)
	}
	return exec.CommandContext(ctx, "sh", "-c", command)
}
