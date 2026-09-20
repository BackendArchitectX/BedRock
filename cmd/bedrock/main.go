package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/BackendArchitectX/BedRock/internal/bedrock"
)

type stringList []string

func (s *stringList) String() string { return strings.Join(*s, ",") }
func (s *stringList) Set(value string) error {
	*s = append(*s, value)
	return nil
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	switch os.Args[1] {
	case "run":
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()
		if err := run(ctx, os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, "bedrock:", err)
			os.Exit(1)
		}
	case "help", "-h", "--help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n", os.Args[1])
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "BedRock - local-first AI software-engineering orchestration")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, "  bedrock run --repo . --task \"...\" --provider-bin <adapter> [--provider-arg <arg>] [--provider-env <NAME>] [--verify <command>]")
	fmt.Fprintln(os.Stderr, "  bedrock run --repo . --task \"...\" --provider-endpoint <url> --provider-model <model> [--provider-api-key-env <NAME>] [--verify <command>]")
}

func run(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	repo := fs.String("repo", ".", "repository root")
	task := fs.String("task", "", "engineering task")
	providerBin := fs.String("provider-bin", "", "provider adapter executable")
	providerEndpoint := fs.String("provider-endpoint", "", "OpenAI-compatible chat-completions endpoint")
	providerModel := fs.String("provider-model", "", "model name for --provider-endpoint")
	providerAPIKeyEnv := fs.String("provider-api-key-env", "", "environment variable containing the HTTP provider API key")
	maxAttempts := fs.Int("max-attempts", 2, "maximum implementation/repair attempts")
	contextFiles := fs.Int("context-files", 80, "maximum context files")
	contextBytes := fs.Int("context-bytes", 200*1024, "maximum context bytes")
	providerTimeout := fs.Duration("provider-timeout", 2*time.Minute, "provider invocation timeout")
	verifyTimeout := fs.Duration("verify-timeout", 3*time.Minute, "timeout per verification command")
	var providerArgs stringList
	var providerEnv stringList
	var verifyCommands stringList
	fs.Var(&providerArgs, "provider-arg", "provider adapter argument; repeatable")
	fs.Var(&providerEnv, "provider-env", "environment variable name to pass to the provider; repeatable")
	fs.Var(&verifyCommands, "verify", "explicit verification shell command; repeatable")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*task) == "" {
		return fmt.Errorf("--task is required")
	}

	provider, err := configuredProvider(*providerBin, providerArgs, providerEnv, *providerEndpoint, *providerModel, *providerAPIKeyEnv, *providerTimeout)
	if err != nil {
		return err
	}

	engine := bedrock.Engine{
		Provider: provider,
		Verifier: bedrock.ShellVerifier{
			Commands: verifyCommands,
			Timeout:  *verifyTimeout,
		},
		MaxAttempts:     *maxAttempts,
		ContextMaxFiles: *contextFiles,
		ContextMaxBytes: *contextBytes,
	}

	result, err := engine.Run(ctx, *repo, *task)
	fmt.Printf("run: %s\n", result.Evidence.RunID)
	fmt.Printf("status: %s\n", result.Evidence.Status)
	fmt.Printf("provider: %s\n", result.Evidence.Provider)
	if len(result.Evidence.ChangedPaths) > 0 {
		fmt.Printf("changed: %s\n", strings.Join(result.Evidence.ChangedPaths, ", "))
	}
	if result.EvidencePath != "" {
		fmt.Printf("evidence: %s\n", result.EvidencePath)
	}
	return err
}

func configuredProvider(bin string, args, env []string, endpoint, model, apiKeyEnv string, timeout time.Duration) (bedrock.Provider, error) {
	bin = strings.TrimSpace(bin)
	endpoint = strings.TrimSpace(endpoint)
	model = strings.TrimSpace(model)
	apiKeyEnv = strings.TrimSpace(apiKeyEnv)
	if bin != "" && endpoint != "" {
		return nil, fmt.Errorf("--provider-bin and --provider-endpoint are mutually exclusive")
	}
	if endpoint != "" {
		if len(args) != 0 || len(env) != 0 {
			return nil, fmt.Errorf("--provider-arg and --provider-env require --provider-bin")
		}
		if model == "" {
			return nil, fmt.Errorf("--provider-model is required with --provider-endpoint")
		}
		var apiKey string
		if apiKeyEnv != "" {
			var ok bool
			apiKey, ok = os.LookupEnv(apiKeyEnv)
			if !ok || apiKey == "" {
				return nil, fmt.Errorf("provider API key environment variable %q is not set", apiKeyEnv)
			}
		}
		return bedrock.OpenAICompatibleProvider{Endpoint: endpoint, Model: model, APIKey: apiKey, Timeout: timeout}, nil
	}
	if model != "" || apiKeyEnv != "" {
		return nil, fmt.Errorf("--provider-model and --provider-api-key-env require --provider-endpoint")
	}
	if bin == "" {
		return nil, fmt.Errorf("one of --provider-bin or --provider-endpoint is required")
	}
	return bedrock.CommandProvider{Bin: bin, Args: args, Timeout: timeout, EnvAllow: env}, nil
}
