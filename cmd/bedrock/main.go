package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
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
		if err := run(os.Args[2:]); err != nil {
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
}

func run(args []string) error {
	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	repo := fs.String("repo", ".", "repository root")
	task := fs.String("task", "", "engineering task")
	providerBin := fs.String("provider-bin", "", "provider adapter executable")
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
	if strings.TrimSpace(*providerBin) == "" {
		return fmt.Errorf("--provider-bin is required")
	}

	engine := bedrock.Engine{
		Provider: bedrock.CommandProvider{
			Bin:      *providerBin,
			Args:     providerArgs,
			Timeout:  *providerTimeout,
			EnvAllow: providerEnv,
		},
		Verifier: bedrock.ShellVerifier{
			Commands: verifyCommands,
			Timeout:  *verifyTimeout,
		},
		MaxAttempts:     *maxAttempts,
		ContextMaxFiles: *contextFiles,
		ContextMaxBytes: *contextBytes,
	}

	result, err := engine.Run(context.Background(), *repo, *task)
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
