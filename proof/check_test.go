package proof_test

import (
	"os/exec"
	"strings"
	"testing"
)

func kitRun(t *testing.T, dir string, args ...string) (string, int) {
	t.Helper()
	cmd := exec.CommandContext(t.Context(), kitCommand(t, "kit"), args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err == nil {
		return string(out), 0
	}
	if exitCode(err) != 1 {
		t.Fatalf("kit %v: %v\n%s", args, err, out)
	}
	return string(out), 1
}

// WHY: crash tests ran a local check.sh that missed a gate CI ran; kit
// check is the one command for both, so it must run every gate and fail on
// any of them.
func TestKitCheckRunsEveryGateAndFailsOnAny(t *testing.T) {
	withZizmor(t)
	gates := []string{
		"golangci-lint (linux)", "kitvet (linux)", "golangci-lint (windows)", "kitvet (windows)",
		"archtest", "sweep", "deadcode", "nilaway", "govulncheck", "go test", "guard",
		"actionlint", "zizmor", "shellcheck", "gitleaks",
	}
	out, code := kitRun(t, changeOf(t, trialCase{name: "no change"}), "check", "--base", "base", "--race=false")
	if code != 0 || !strings.Contains(out, "kit check: every gate passes") {
		t.Fatalf("a normal change: exit %d\n%s", code, out)
	}
	for _, g := range gates {
		if !strings.Contains(out, "ok   "+g+"\n") {
			t.Errorf("gate %s did not run or did not pass:\n%s", g, out)
		}
	}
	out, code = kitRun(t, changeOf(t, caseNamed(t, "m5-only-tests")), "check", "--base", "base", "--race=false")
	if code != 1 || !strings.Contains(out, "FAIL sweep\n") || !strings.Contains(out, "Oldest is reached by tests only") {
		t.Errorf("a method only tests reach: exit %d\n%s", code, out)
	}
}
