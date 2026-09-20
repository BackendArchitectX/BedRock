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
