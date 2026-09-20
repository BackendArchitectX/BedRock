package demosmoke

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestLauncherRecoversAfterVerificationFailure(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	work := filepath.Join(t.TempDir(), "workspace")
	if err := os.Mkdir(work, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(work, ".bedrock-demo-owned"), []byte("BedRock demo workspace\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	const userState = "preserve me\n"
	if err := os.WriteFile(filepath.Join(work, "user.txt"), []byte(userState), 0o600); err != nil {
		t.Fatal(err)
	}

	failed := exec.Command("go", "run", "./scripts/demo.go")
	failed.Dir = root
	failed.Env = append(os.Environ(), "BEDROCK_DEMO_DIR="+work, "BEDROCK_DEMO_VERIFY_FAIL=1")
	failedOutput, err := failed.CombinedOutput()
	if err == nil {
		t.Fatalf("launcher unexpectedly survived injected verification failure; output:\n%s", failedOutput)
	}
	if strings.Contains(string(failedOutput), "BedRock demo: READY") {
		t.Fatalf("failed verification falsely reported readiness:\n%s", failedOutput)
	}
	if _, err := os.Stat(filepath.Join(work, "repository", "result.txt")); !os.IsNotExist(err) {
		t.Fatalf("failed verification left generated result behind: stat err = %v", err)
	}
	contents, err := os.ReadFile(filepath.Join(work, "user.txt"))
	if err != nil {
		t.Fatalf("read unrelated user state after failed verification: %v", err)
	}
	if string(contents) != userState {
		t.Fatalf("failed verification modified unrelated user state: got %q", contents)
	}

	rerun := exec.Command("go", "run", "./scripts/demo.go")
	rerun.Dir = root
	rerun.Env = append(os.Environ(), "BEDROCK_DEMO_DIR="+work)
	rerunOutput, err := rerun.CombinedOutput()
	if err != nil {
		t.Fatalf("launcher did not recover after verification failure: %v\n%s", err, rerunOutput)
	}
	if !strings.Contains(string(rerunOutput), "status: VERIFIED") || !strings.Contains(string(rerunOutput), "BedRock demo: READY") {
		t.Fatalf("recovered launcher did not reach verified readiness:\n%s", rerunOutput)
	}
	contents, err = os.ReadFile(filepath.Join(work, "user.txt"))
	if err != nil {
		t.Fatalf("read unrelated user state after recovery: %v", err)
	}
	if string(contents) != userState {
		t.Fatalf("recovery modified unrelated user state: got %q", contents)
	}
	result, err := os.ReadFile(filepath.Join(work, "repository", "result.txt"))
	if err != nil {
		t.Fatalf("read recovered result: %v", err)
	}
	if strings.TrimSpace(string(result)) != "good" {
		t.Fatalf("recovered result mismatch: %q", result)
	}
}
