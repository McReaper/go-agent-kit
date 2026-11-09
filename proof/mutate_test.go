package proof_test

import (
	"path/filepath"
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

// WHY: mutation is too slow for every package on every change; rumblestrip mutate
// scores only the lines a change touched, so a weak test of new code fails
// the change while old code's survivors don't.
// WHY: on GitHub the owner reads mutation like the other checks: each
// change the tests missed, at its line, in the run's summary.
func TestKitMutateWritesItsLinesAndItsSectionOfTheSummary(t *testing.T) {
	summaryDir := t.TempDir()
	out, code := kitRun(t, changeOf(t, caseNamed(t, "m6-weak-test")), "mutate", "--base", "base", "--github", "--summary", summaryDir)
	if code != 1 || !regexp.MustCompile(`(?m)^::error file=[^,]+\.go,line=\d+,title=Mutation::the tests still pass when a change `).MatchString(out) {
		t.Errorf("exit %d; no annotation of a lived mutant:\n%s", code, out)
	}
	if got := readFile(t, filepath.Join(summaryDir, "mutate.row")); got != "Mutation\tThe tests catch the changes made to the lines the pull request added: 80% of them at least.\n" {
		t.Errorf("mutate.row = %q", got)
	}
	section := readFile(t, filepath.Join(summaryDir, "mutate.md"))
	if !strings.HasPrefix(section, "### ❌ Mutation\n\nIt wants: the tests catch the changes made to the lines the pull request added: 80% of them at least. It found 3 problems") ||
		!strings.Contains(section, "| the tests still pass when a change ") {
		t.Errorf("mutate.md:\n%s", section)
	}
}

func TestKitMutateFailsAChangeItsTestsDontPin(t *testing.T) {
	c := caseNamed(t, "m6-weak-test")
	out, code := kitRun(t, changeOf(t, c), "mutate", "--base", "base")
	if code != 1 || !strings.Contains(out, "killed 4 of the 7 mutants") {
		t.Errorf("a weak test: exit %d\n%s", code, out)
	}
	check(t, wantsOf("mutesting/", c), kitMutateFindings(t, out))
	out, code = kitRun(t, changeOf(t, trialCase{name: "no change"}), "mutate", "--base", "base")
	if code != 0 || !strings.Contains(out, "killed 0 of the 0 mutants") {
		t.Errorf("no change: exit %d\n%s", code, out)
	}
	out, code = kitRun(t, changeOf(t, c), "mutate", "--base", "base", "--min", "0.5")
	if code != 0 {
		t.Errorf("a weak test under a lower bar: exit %d\n%s", code, out)
	}
}
