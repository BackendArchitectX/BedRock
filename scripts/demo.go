//go:build ignore

package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const demoMarker = "BedRock demo workspace\n"

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "BedRock demo: "+format+"\n", args...)
	os.Exit(1)
}

func main() {
	if err := run(); err != nil {
		fail("%v", err)
	}
}

func run() error {
	if _, err := exec.LookPath("go"); err != nil {
		return fmt.Errorf("Go 1.22+ is required and was not found on PATH")
	}
	if err := requireSupportedGo(); err != nil {
		return err
	}
	if _, err := exec.LookPath("git"); err != nil {
		return fmt.Errorf("Git is required and was not found on PATH")
	}

	root, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("resolve checkout: %w", err)
	}
	root, err = filepath.Abs(root)
	if err != nil {
		return fmt.Errorf("resolve checkout: %w", err)
	}
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		return fmt.Errorf("run this command from the BedRock checkout root: %w", err)
	}

	work := os.Getenv("BEDROCK_DEMO_DIR")
	if work == "" {
		work = filepath.Join(os.TempDir(), "bedrock-demo")
	}
	work, err = filepath.Abs(work)
	if err != nil {
		return fmt.Errorf("resolve demo directory: %w", err)
	}
	if filepath.Dir(work) == work {
		return fmt.Errorf("refusing to use filesystem root %s as BEDROCK_DEMO_DIR", work)
	}
	if err := rejectSymlinkComponents(work); err != nil {
		return err
	}
	// Check the containing-workspace direction first so equality retains the
	// established checkout-protection diagnostic. A strict child still falls
	// through to the inside-checkout guard below.
	if containsPath(work, root) {
		return fmt.Errorf("refusing to use %s as BEDROCK_DEMO_DIR because it contains the BedRock checkout %s", work, root)
	}
	if containsPath(root, work) {
		return fmt.Errorf("refusing to use %s as BEDROCK_DEMO_DIR because it is inside the BedRock checkout %s", work, root)
	}
	marker := filepath.Join(work, ".bedrock-demo-owned")
	newWorkspace := false
	if info, statErr := os.Stat(work); statErr == nil {
		if !info.IsDir() {
			return fmt.Errorf("refusing to reuse %s because it is not a directory", work)
		}
		markerInfo, markerErr := os.Lstat(marker)
		if markerErr != nil {
			return fmt.Errorf("refusing to reuse %s because it is not marked as a BedRock demo directory; choose an empty BEDROCK_DEMO_DIR or remove it yourself", work)
		}
		if !markerInfo.Mode().IsRegular() {
			return fmt.Errorf("refusing to reuse %s because its BedRock ownership marker is not a regular file", work)
		}
		markerContents, markerErr := os.ReadFile(marker)
		if markerErr != nil {
			return fmt.Errorf("inspect BedRock ownership marker: %w", markerErr)
		}
		if string(markerContents) != demoMarker {
			return fmt.Errorf("refusing to reuse %s because its BedRock ownership marker is invalid", work)
		}
	} else if os.IsNotExist(statErr) {
		newWorkspace = true
	} else {
		return fmt.Errorf("inspect demo directory: %w", statErr)
	}
	if err := os.MkdirAll(work, 0o700); err != nil {
		return fmt.Errorf("create demo directory: %w", err)
	}
	if newWorkspace {
		markerFile, err := os.OpenFile(marker, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
		if err != nil {
			return fmt.Errorf("mark demo directory: %w", err)
		}
		if _, err := markerFile.WriteString(demoMarker); err != nil {
			_ = markerFile.Close()
			return fmt.Errorf("mark demo directory: %w", err)
		}
		if err := markerFile.Close(); err != nil {
			return fmt.Errorf("mark demo directory: %w", err)
		}
	}

	binDir := filepath.Join(work, "bin")
	repo := filepath.Join(work, "repository")
	for _, path := range []string{binDir, repo} {
		if err := os.RemoveAll(path); err != nil {
			return fmt.Errorf("clean %s: %w", path, err)
		}
		if err := os.MkdirAll(path, 0o700); err != nil {
			return fmt.Errorf("create %s: %w", path, err)
		}
	}
	if err := command(root, "git", "-C", repo, "init", "-q"); err != nil {
		return err
	}

	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}
	bedrock := filepath.Join(binDir, "bedrock"+ext)
	provider := filepath.Join(binDir, "fake-provider"+ext)
	fmt.Println("BedRock demo: building CLI and deterministic provider...")
	if err := command(root, "go", "build", "-o", bedrock, "./cmd/bedrock"); err != nil {
		return err
	}
	if err := command(root, "go", "build", "-o", provider, "./cmd/bedrock/testdata/fakeprovider"); err != nil {
		return err
	}

	fmt.Println("BedRock demo: starting verified orchestration run...")
	verify := "test \"$(cat result.txt)\" = good"
	if runtime.GOOS == "windows" {
		verify = "for /f %i in (result.txt) do @if \"%i\"==\"good\" (exit /b 0) else (exit /b 1)"
	}
	if err := command(root, bedrock, "run", "--repo", repo, "--task", "write deterministic result", "--provider-bin", provider, "--verify", verify); err != nil {
		return err
	}
	result, err := os.ReadFile(filepath.Join(repo, "result.txt"))
	if err != nil {
		return fmt.Errorf("read verified result: %w", err)
	}
	if string(bytes.TrimSpace(result)) != "good" {
		return fmt.Errorf("verification output did not match the expected result")
	}
	fmt.Println("BedRock demo: READY")
	fmt.Printf("Workspace: %s\nResult: %s\n", repo, filepath.Join(repo, "result.txt"))
	return nil
}

func rejectSymlinkComponents(path string) error {
	current := path
	for {
		info, err := os.Lstat(current)
		if err == nil {
			if info.Mode()&os.ModeSymlink != 0 {
				if current == path {
					return fmt.Errorf("refusing to use symlink %s as BEDROCK_DEMO_DIR", path)
				}
				return fmt.Errorf("refusing to use %s as BEDROCK_DEMO_DIR because path component %s is a symlink", path, current)
			}
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("inspect demo directory path component %s: %w", current, err)
		}
		parent := filepath.Dir(current)
		if parent == current {
			return nil
		}
		current = parent
	}
}

func containsPath(parent, child string) bool {
	rel, err := filepath.Rel(parent, child)
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}

func requireSupportedGo() error {
	output, err := exec.Command("go", "env", "GOVERSION").Output()
	if err != nil {
		return fmt.Errorf("determine Go version: %w", err)
	}
	version := strings.TrimSpace(string(output))
	var major, minor int
	if _, err := fmt.Sscanf(strings.TrimPrefix(version, "go"), "%d.%d", &major, &minor); err != nil {
		return fmt.Errorf("could not parse Go version %q; Go 1.22+ is required", version)
	}
	if major < 1 || (major == 1 && minor < 22) {
		return fmt.Errorf("Go 1.22+ is required; found %s", version)
	}
	return nil
}

func command(dir, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s failed: %w", name, err)
	}
	return nil
}
