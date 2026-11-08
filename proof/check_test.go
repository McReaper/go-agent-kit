package proof_test

import (
	"os/exec"
	"strings"
	"testing"
)

func kitRun(t *testing.T, dir string, args ...string) (string, int) {
	t.Helper()
	cmd := exec.CommandContext(t.Context(), kitCommand(t, "rumblestrip"), args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err == nil {
		return string(out), 0
	}
	if exitCode(err) != 1 {
		t.Fatalf("rumblestrip %v: %v\n%s", args, err, out)
	}
	return string(out), 1
}

// WHY: crash tests ran a local check.sh that missed a gate CI ran; rumblestrip
// check is the one command for both, so it must run every gate and fail on
// any of them.
func TestKitCheckRunsEveryGateAndFailsOnAny(t *testing.T) {
	withZizmor(t)
	gates := []string{
		"golangci-lint (linux)", "rumblestrip-vet (linux)", "golangci-lint (windows)", "rumblestrip-vet (windows)",
		"archtest", "sweep", "deadcode", "nilaway", "govulncheck", "go test", "spec", "guard",
		"actionlint", "zizmor", "shellcheck", "psscriptanalyzer", "gitleaks", "rumblestrip files",
	}
	out, code := kitRun(t, changeOf(t, trialCase{name: "no change"}), "check", "--base", "base", "--race=false")
	if code != 0 || !strings.Contains(out, "rumblestrip check: every gate passes") {
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
