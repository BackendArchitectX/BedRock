package bedrock

import "context"

type FileContext struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type ProviderRequest struct {
	Task                 string        `json:"task"`
	Attempt              int           `json:"attempt"`
	Failure              string        `json:"failure,omitempty"`
	Files                []FileContext `json:"files"`
	ProtectedPaths       []string      `json:"protectedPaths,omitempty"`
	VerificationCommands []string      `json:"verificationCommands,omitempty"`
}

type FileChange struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type ProviderResponse struct {
	Summary string       `json:"summary"`
	Changes []FileChange `json:"changes"`
}

type Provider interface {
	Name() string
	Execute(context.Context, ProviderRequest) (ProviderResponse, error)
}

type VerificationResult struct {
	Command  string `json:"command"`
	ExitCode int    `json:"exitCode"`
	Output   string `json:"output,omitempty"`
}

type Verifier interface {
	Verify(context.Context, string) ([]VerificationResult, error)
}

type Evidence struct {
	RunID                string               `json:"runId"`
	Task                 string               `json:"task"`
	Provider             string               `json:"provider"`
	Status               string               `json:"status"`
	Attempts             int                  `json:"attempts"`
	ProviderSummaries    []string             `json:"providerSummaries,omitempty"`
	ChangedPaths         []string             `json:"changedPaths,omitempty"`
	BaselineVerification []VerificationResult `json:"baselineVerification,omitempty"`
	Verification         []VerificationResult `json:"verification,omitempty"`
	LastFailure          string               `json:"lastFailure,omitempty"`
	RolledBack           bool                 `json:"rolledBack"`
	Repository           string               `json:"repository"`
	RecordedAtUTC        string               `json:"recordedAtUtc"`
}

type RunResult struct {
	Evidence     Evidence
	EvidencePath string
}
