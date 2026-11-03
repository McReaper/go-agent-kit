package guard

import (
	"regexp"
	"strings"
)

var testHook = regexp.MustCompile(`^func (TestMain\(|init\(\))`)

// WHY: a TestMain or an init in a test runs before every test of its
// package and can skip them all; the one a package needs, goleak's, is
// written the one way.
func testHooks(f File) []Finding {
	var out []Finding
	for i, l := range f.Added {
		if !testHook.MatchString(l.Text) || isGoleakMain(f.Added[i:]) {
			continue
		}
		out = append(out, Finding{
			"bypass-testmain", f.Path, l.N,
			"T2: a TestMain or init in a test runs before every test of the package and can skip them all: remove it; the one TestMain allowed is func TestMain(m *testing.M) { goleak.VerifyTestMain(m) }",
		})
	}
	return out
}

func isGoleakMain(lines []Line) bool {
	return len(lines) >= 3 &&
		strings.HasPrefix(lines[0].Text, "func TestMain(m *testing.M) {") &&
		lines[1].N == lines[0].N+1 && strings.TrimSpace(lines[1].Text) == "goleak.VerifyTestMain(m)" &&
		lines[2].N == lines[0].N+2 && lines[2].Text == "}"
}
