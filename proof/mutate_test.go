package proof_test

import (
	"regexp"
	"strconv"
	"strings"
	"testing"
)

var livedLine = regexp.MustCompile(`^([^:]+):(\d+): lived: [^:]+: (.*)$`)

func kitMutateFindings(t *testing.T, out string) []finding {
	t.Helper()
	var findings []finding
	for line := range strings.Lines(out) {
		m := livedLine.FindStringSubmatch(strings.TrimRight(line, "\r\n"))
		if m == nil {
			continue
		}
		n, err := strconv.Atoi(m[2])
		if err != nil {
			t.Fatal(err)
		}
		findings = append(findings, finding{tool: "mutesting/lived", file: m[1], line: n, text: m[3]})
	}
	return findings
}

// WHY: mutation is too slow for every package on every change; kit mutate
// scores only the lines a change touched, so a weak test of new code fails
// the change while old code's survivors don't.
func TestKitMutateFailsAChangeItsTestsDontPin(t *testing.T) {
	c := caseNamed(t, "m6-weak-test")
	out, code := kitRun(t, changeOf(t, c), "mutate", "--base", "base")
	if code != 1 || !strings.Contains(out, "killed 4 of the 7 mutants") {
		t.Errorf("a weak test: exit %d\n%s", code, out)
	}
	check(t, []trialCase{c}, "mutesting/", kitMutateFindings(t, out))
	out, code = kitRun(t, changeOf(t, trialCase{name: "no change"}), "mutate", "--base", "base")
	if code != 0 || !strings.Contains(out, "killed 0 of the 0 mutants") {
		t.Errorf("no change: exit %d\n%s", code, out)
	}
	out, code = kitRun(t, changeOf(t, c), "mutate", "--base", "base", "--min", "0.5")
	if code != 0 {
		t.Errorf("a weak test under a lower bar: exit %d\n%s", code, out)
	}
}
