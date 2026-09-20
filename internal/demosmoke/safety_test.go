package demosmoke

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestLauncherRejectsSymlinkOwnershipMarker(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("creating symlinks is not reliably available on Windows runners")
	}

	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	work := filepath.Join(t.TempDir(), "workspace")
	if err := os.Mkdir(work, 0o700); err != nil {
		t.Fatal(err)
	}

	target := filepath.Join(t.TempDir(), "outside-marker")
	const sentinel = "BedRock demo workspace\n"
	if err := os.WriteFile(target, []byte(sentinel), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(work, ".bedrock-demo-owned")); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("go", "run", "./scripts/demo.go")
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "BEDROCK_DEMO_DIR="+work)
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("launcher unexpectedly accepted a symlink ownership marker; output:\n%s", output)
	}
	if !strings.Contains(string(output), "ownership marker is not a regular file") {
		t.Fatalf("launcher returned the wrong diagnostic:\n%s", output)
	}

	contents, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read external marker target: %v", err)
	}
	if string(contents) != sentinel {
		t.Fatalf("external marker target was modified: got %q", contents)
	}
	if _, err := os.Stat(filepath.Join(work, "bin")); !os.IsNotExist(err) {
		t.Fatalf("launcher mutated workspace before rejecting symlink marker: stat err = %v", err)
	}
}

func TestLauncherRecoversAfterGitInitFailure(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the injected git shim is POSIX-only; Windows launcher recovery is covered by rerun acceptance")
	}

	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	work := filepath.Join(t.TempDir(), "workspace")
	shimDir := t.TempDir()
	realGit, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	shim := filepath.Join(shimDir, "git")
	shimContents := "#!/bin/sh\ncase \"$*\" in\n  *' init -q') exit 73 ;;\nesac\nexec \"" + realGit + "\" \"$@\"\n"
	if err := os.WriteFile(shim, []byte(shimContents), 0o700); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("go", "run", "./scripts/demo.go")
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "PATH="+shimDir+string(os.PathListSeparator)+os.Getenv("PATH"), "BEDROCK_DEMO_DIR="+work)
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("launcher unexpectedly survived injected git init failure; output:\n%s", output)
	}
	if !strings.Contains(string(output), "git failed") {
		t.Fatalf("launcher returned the wrong failure diagnostic:\n%s", output)
	}
	marker, err := os.ReadFile(filepath.Join(work, ".bedrock-demo-owned"))
	if err != nil {
		t.Fatalf("launcher did not establish workspace ownership before init failure: %v", err)
	}
	if string(marker) != "BedRock demo workspace\n" {
		t.Fatalf("unexpected ownership marker: %q", marker)
	}

	const userState = "preserve me\n"
	if err := os.WriteFile(filepath.Join(work, "user.txt"), []byte(userState), 0o600); err != nil {
		t.Fatal(err)
	}
	rerun := exec.Command("go", "run", "./scripts/demo.go")
	rerun.Dir = root
	rerun.Env = append(os.Environ(), "BEDROCK_DEMO_DIR="+work)
	rerunOutput, err := rerun.CombinedOutput()
	if err != nil {
		t.Fatalf("launcher did not recover on rerun: %v\n%s", err, rerunOutput)
	}
	if !strings.Contains(string(rerunOutput), "BedRock demo: READY") {
		t.Fatalf("recovered launcher did not report readiness:\n%s", rerunOutput)
	}
	contents, err := os.ReadFile(filepath.Join(work, "user.txt"))
	if err != nil {
		t.Fatalf("read unrelated user state: %v", err)
	}
	if string(contents) != userState {
		t.Fatalf("rerun modified unrelated user state: got %q", contents)
	}
	result, err := os.ReadFile(filepath.Join(work, "repository", "result.txt"))
	if err != nil {
		t.Fatalf("read recovered result: %v", err)
	}
	if strings.TrimSpace(string(result)) != "good" {
		t.Fatalf("recovered result mismatch: %q", result)
	}
}
