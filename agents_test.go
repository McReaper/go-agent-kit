package agents_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var (
	backticked = regexp.MustCompile("`([^`]+)`")
	skillName  = regexp.MustCompile("the `([a-z0-9-]+)` skill")
)

// WHY: an agent follows AGENTS.md to the letter; a path it names that no
// longer exists sends it looking, or making one in the wrong place.
func TestEveryPathAgentsMDNamesExists(t *testing.T) {
	doc, err := os.ReadFile("AGENTS.md")
	if err != nil {
		t.Fatal(err)
	}
	checked := 0
	for _, m := range backticked.FindAllStringSubmatch(string(doc), -1) {
		p := m[1]
		if !strings.Contains(p, "/") || strings.ContainsAny(p, " <*") || strings.HasPrefix(p, "//") {
			continue
		}
		checked++
		if _, err := os.Stat(filepath.FromSlash(strings.TrimSuffix(p, "/"))); err != nil {
			t.Errorf("AGENTS.md names %s: %v", p, err)
		}
	}
	if checked == 0 {
		t.Fatal("AGENTS.md names no path: the test reads nothing")
	}
}

func TestEverySkillAgentsMDNamesExists(t *testing.T) {
	doc, err := os.ReadFile("AGENTS.md")
	if err != nil {
		t.Fatal(err)
	}
	skills := skillName.FindAllStringSubmatch(string(doc), -1)
	if len(skills) == 0 {
		t.Fatal("AGENTS.md names no skill: the test that checks them reads nothing")
	}
	for _, m := range skills {
		if _, err := os.Stat(filepath.Join(".claude", "skills", m[1], "SKILL.md")); err != nil {
			t.Errorf("AGENTS.md names the %s skill: %v", m[1], err)
		}
	}
}
