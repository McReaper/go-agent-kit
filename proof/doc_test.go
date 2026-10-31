package proof_test

import (
	"os"
	"path"
	"regexp"
	"slices"
	"strings"
	"testing"
)

func readDoc(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile("../docs/kit.md")
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

var backticked = regexp.MustCompile("`([a-z0-9][a-z0-9-]*)`")

func TestTheDocCitesEveryCaseAndOnlyCasesThatExist(t *testing.T) {
	doc := readDoc(t)
	cited := map[string]bool{}
	for _, m := range backticked.FindAllStringSubmatch(doc, -1) {
		cited[m[1]] = true
	}
	entries, err := os.ReadDir(casesDir)
	if err != nil {
		t.Fatal(err)
	}
	var cases []string
	for _, e := range entries {
		if e.IsDir() {
			cases = append(cases, e.Name())
			if !cited[e.Name()] {
				t.Errorf("case %s proves a rule docs/kit.md never names", e.Name())
			}
		}
	}
	caseLike := regexp.MustCompile(`^(m\d+|e\d+|s\d*|t\d|guard|ci|sh|fmt)-[a-z0-9-]+$`)
	for name := range cited {
		if caseLike.MatchString(name) && !slices.Contains(cases, name) {
			t.Errorf("docs/kit.md cites case %s, which trial/cases doesn't hold", name)
		}
	}
}

func TestTheDocNamesEveryPinnedTool(t *testing.T) {
	doc := readDoc(t)
	data, err := os.ReadFile("../tools.txt")
	if err != nil {
		t.Fatal(err)
	}
	for line := range strings.Lines(string(data)) {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		name, _, _ := strings.Cut(strings.Fields(line)[0], "@")
		name = path.Base(name)
		if name == "v8" {
			name = "gitleaks"
		}
		if !strings.Contains(doc, name) {
			t.Errorf("tools.txt pins %s, which docs/kit.md never names", name)
		}
	}
}

func TestTheDocsProofTestsExist(t *testing.T) {
	doc := readDoc(t)
	tests := regexp.MustCompile("`(Test[A-Za-z]+)`").FindAllStringSubmatch(doc, -1)
	src := ""
	files, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		data, err := os.ReadFile(f.Name())
		if err != nil {
			t.Fatal(err)
		}
		src += string(data)
	}
	for _, m := range tests {
		if !strings.Contains(src, "func "+m[1]+"(") {
			t.Errorf("docs/kit.md cites %s, which no proof test defines", m[1])
		}
	}
}
