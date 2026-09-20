package bedrock

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type Engine struct {
	Provider        Provider
	Verifier        Verifier
	MaxAttempts     int
	ContextMaxFiles int
	ContextMaxBytes int
}

func (e Engine) Run(ctx context.Context, root, task string) (RunResult, error) {
	if e.Provider == nil {
		return RunResult{}, errors.New("provider is required")
	}
	if strings.TrimSpace(task) == "" {
		return RunResult{}, errors.New("task is required")
	}
	root, err := filepath.Abs(root)
	if err != nil {
		return RunResult{}, err
	}
	info, err := os.Stat(root)
	if err != nil {
		return RunResult{}, fmt.Errorf("repository root: %w", err)
	}
	if !info.IsDir() {
		return RunResult{}, fmt.Errorf("repository root is not a directory: %s", root)
	}

	maxAttempts := e.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 2
	}

	protected, err := DirtyPaths(root)
	if err != nil {
		return RunResult{}, err
	}
	protectedPaths := make([]string, 0, len(protected))
	for path := range protected {
		protectedPaths = append(protectedPaths, path)
	}
	sort.Strings(protectedPaths)

	var verificationCommands []string
	switch verifier := e.Verifier.(type) {
	case ShellVerifier:
		verificationCommands = append([]string(nil), verifier.Commands...)
	case *ShellVerifier:
		if verifier != nil {
			verificationCommands = append([]string(nil), verifier.Commands...)
		}
	}

	runID := "run-" + time.Now().UTC().Format("20060102T150405.000000000Z")
	evidence := Evidence{
		RunID:         runID,
		Task:          task,
		Provider:      e.Provider.Name(),
		Status:        "FAILED",
		Repository:    root,
		RecordedAtUTC: time.Now().UTC().Format(time.RFC3339Nano),
	}
	changes := NewChangeSet()
	var failure string

	finish := func(runErr error) (RunResult, error) {
		evidence.ChangedPaths = changes.ChangedPaths()
		evidence.LastFailure = failure
		path, saveErr := SaveEvidence(root, evidence)
		if saveErr != nil {
			if runErr != nil {
				return RunResult{Evidence: evidence}, fmt.Errorf("%v; save evidence: %w", runErr, saveErr)
			}
			return RunResult{Evidence: evidence}, fmt.Errorf("save evidence: %w", saveErr)
		}
		return RunResult{Evidence: evidence, EvidencePath: path}, runErr
	}

	rollback := func(cause error) error {
		rollbackErr := changes.Rollback(root)
		if rollbackErr != nil {
			evidence.RolledBack = false
			combined := fmt.Errorf("%w; rollback failed: %v", cause, rollbackErr)
			failure = combined.Error()
			return combined
		}
		evidence.RolledBack = true
		return cause
	}

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		evidence.Attempts = attempt
		files, err := Snapshot(root, task, e.ContextMaxFiles, e.ContextMaxBytes)
		if err != nil {
			failure = err.Error()
			return finish(rollback(err))
		}

		response, err := e.Provider.Execute(ctx, ProviderRequest{
			Task:                 task,
			Attempt:              attempt,
			Failure:              failure,
			Files:                files,
			ProtectedPaths:       protectedPaths,
			VerificationCommands: verificationCommands,
		})
		if err != nil {
			failure = err.Error()
			return finish(rollback(err))
		}

		if err := changes.ValidateOwnWrites(root); err != nil {
			failure = err.Error()
			return finish(rollback(err))
		}
		currentProtected, err := DirtyPaths(root)
		if err != nil {
			failure = err.Error()
			return finish(rollback(err))
		}
		if err := changes.ExcludeOwnWrites(root, currentProtected); err != nil {
			failure = err.Error()
			return finish(rollback(err))
		}
		if err := changes.Apply(root, response.Changes, currentProtected); err != nil {
			failure = err.Error()
			return finish(rollback(err))
		}

		if e.Verifier == nil {
			evidence.Status = "UNVERIFIED"
			return finish(nil)
		}
		results, verifyErr := e.Verifier.Verify(ctx, root)
		evidence.Verification = results
		if verifyErr == nil {
			verifyErr = verificationResultsError(results)
		}
		if verifyErr == nil {
			if len(results) == 0 {
				evidence.Status = "UNVERIFIED"
			} else {
				evidence.Status = "VERIFIED"
			}
			return finish(nil)
		}

		failure = verificationFailure(verifyErr, results)
		if attempt == maxAttempts {
			return finish(rollback(verifyErr))
		}
	}
	return finish(errors.New("attempt loop ended unexpectedly"))
}

func verificationResultsError(results []VerificationResult) error {
	for _, result := range results {
		if result.ExitCode != 0 {
			return fmt.Errorf("verification command %q reported exit code %d without an error", result.Command, result.ExitCode)
		}
	}
	return nil
}

func verificationFailure(err error, results []VerificationResult) string {
	var b strings.Builder
	b.WriteString(err.Error())
	for _, result := range results {
		if result.ExitCode == 0 {
			continue
		}
		b.WriteString("\ncommand: ")
		b.WriteString(result.Command)
		b.WriteString("\noutput:\n")
		b.WriteString(result.Output)
	}
	return b.String()
}

func SaveEvidence(root string, evidence Evidence) (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256([]byte(root))
	repoKey := hex.EncodeToString(sum[:8])
	dir := filepath.Join(cacheDir, "bedrock", "runs", repoKey)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	data, err := json.MarshalIndent(evidence, "", "  ")
	if err != nil {
		return "", err
	}
	path := filepath.Join(dir, evidence.RunID+".json")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return "", err
	}
	return path, nil
}
