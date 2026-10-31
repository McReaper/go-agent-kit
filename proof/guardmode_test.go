package proof_test

import (
	"os/exec"
	"strings"
	"testing"
)

func guardRun(t *testing.T, dir string, args ...string) (string, int) {
	t.Helper()
	cmd := exec.CommandContext(t.Context(), kitCommand(t, "kit"), append([]string{"guard", "--base", "base"}, args...)...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err == nil {
		return string(out), 0
	}
	if exitCode(err) != 1 {
		t.Fatalf("kit guard %v: %v\n%s", args, err, stderrOf(err))
	}
	return string(out), 1
}

func caseNamed(t *testing.T, name string) trialCase {
	t.Helper()
	for _, c := range loadCases(t) {
		if c.name == name {
			return c
		}
	}
	t.Fatalf("no case %s", name)
	return trialCase{}
}

// WHY: a guard that fails locally on a changed test pushed an agent to
// bend the code to fit a wrong test (crash test r2-envtest); locally it
// reports protected changes, and only the merge check holds them.
func TestGuardHoldsProtectedChangesAtMergeAndBypassesEverywhere(t *testing.T) {
	protected := changeOf(t, caseNamed(t, "guard-protected"))
	bypass := changeOf(t, caseNamed(t, "guard-bypass"))
	cases := []struct {
		name string
		dir  string
		args []string
		code int
	}{
		{"a protected change passes locally", protected, nil, 0},
		{"and fails the merge check", protected, []string{"--ci"}, 1},
		{"until a code owner approved it", protected, []string{"--ci", "--approved"}, 0},
		{"a bypass fails locally", bypass, nil, 1},
		{"and even approved", bypass, []string{"--ci", "--approved"}, 1},
	}
	for _, c := range cases {
		out, code := guardRun(t, c.dir, c.args...)
		if code != c.code {
			t.Errorf("%s: exit %d, want %d\n%s", c.name, code, c.code, out)
		}
	}
	if out, _ := guardRun(t, protected); !strings.Contains(out, "Needs the owner's approval before merge") || strings.Contains(out, "Blocking") {
		t.Errorf("a protected change is reported for the owner, not as blocking:\n%s", out)
	}
}
