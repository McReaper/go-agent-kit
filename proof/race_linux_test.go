//go:build linux

package proof_test

import (
	"regexp"
	"strings"
	"testing"
)

// WHY: the race detector needs cgo and a C compiler; the kit's CI runs
// on Linux, where both are present, so this proof is built for Linux only.
func raceFindings(t *testing.T, dir string) []finding {
	t.Helper()
	out := runIn(t, dir, []int{1}, "go", "test", "-race", "-count=1", "./...")
	return raceFrames(t, dir, string(out))
}

var raceFrame = regexp.MustCompile(`^\s+(\S+\.go):(\d+)`)

func raceFrames(t *testing.T, dir, out string) []finding {
	t.Helper()
	var findings []finding
	inRace := false
	for line := range strings.Lines(out) {
		switch {
		case strings.HasPrefix(line, "WARNING: DATA RACE"):
			inRace = true
		case strings.HasPrefix(line, "=================="):
			inRace = false
		}
		m := raceFrame.FindStringSubmatch(line)
		if !inRace || m == nil || strings.HasSuffix(m[1], "_test.go") || !strings.Contains(m[1], "/internal/") {
			continue
		}
		f := finding{tool: "race/data-race", file: relTo(t, dir, m[1]), line: atoi(t, m[2]), text: "WARNING: DATA RACE"}
		if !containsFinding(findings, f) {
			findings = append(findings, f)
		}
	}
	return findings
}

func TestRaceDetectorReportsEveryCase(t *testing.T) {
	proveCases(t, "race/", raceFindings)
}
