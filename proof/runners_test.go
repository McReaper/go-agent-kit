package proof_test

import (
	"strings"
	"testing"
)

// WHY: a case whose tool has no runner would have its findings silently
// unchecked; every tool a want.txt names must have one on this system.
func TestEveryCaseToolHasARunner(t *testing.T) {
	runners := map[string]bool{golangciTool: true, kitvetTool: true, guardTool: true, archTool: true, sweepTool: true, "mutesting/": true, "apidiff/": true, "race/": raceRunner}
	for _, tl := range standalone {
		runners[tl.prefix] = true
	}
	for _, c := range loadCases(t) {
		for _, w := range c.wants {
			prefix, _, _ := strings.Cut(w.tool, "/")
			if ok, known := runners[prefix+"/"]; !known || (!ok && prefix != "race") {
				t.Errorf("case %s: no runner proves %s", c.name, w.tool)
			}
		}
	}
	if !raceRunner {
		t.Log("the race detector's cases are proven on Linux CI only")
	}
}
