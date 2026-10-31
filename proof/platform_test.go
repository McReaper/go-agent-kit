package proof_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

type ruleset struct {
	Target     string `json:"target"`
	Conditions struct {
		RefName struct {
			Include []string `json:"include"`
		} `json:"ref_name"`
	} `json:"conditions"`
	BypassActors []struct {
		ActorType  string `json:"actor_type"`
		BypassMode string `json:"bypass_mode"`
	} `json:"bypass_actors"`
	Rules []struct {
		Type       string          `json:"type"`
		Parameters json.RawMessage `json:"parameters"`
	} `json:"rules"`
}

func readJSON(t *testing.T, path string, v any) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, v); err != nil {
		t.Fatalf("%s: %v", path, err)
	}
}

func TestTheRulesetLetsOnlyTheOwnerMergeAnApprovedCheckedChange(t *testing.T) {
	var r ruleset
	readJSON(t, "../platform/ruleset.json", &r)
	var types []string
	for _, rule := range r.Rules {
		types = append(types, rule.Type)
		if rule.Type == "pull_request" {
			for _, p := range []string{`"require_code_owner_review": true`, `"dismiss_stale_reviews_on_push": true`, `"require_last_push_approval": true`} {
				if !strings.Contains(string(rule.Parameters), p) {
					t.Errorf("pull_request rule lacks %s", p)
				}
			}
		}
	}
	for _, want := range []string{"update", "pull_request", "required_status_checks", "non_fast_forward", "deletion"} {
		if !slices.Contains(types, want) {
			t.Errorf("the ruleset lacks a %s rule", want)
		}
	}
	if len(r.BypassActors) != 1 || r.BypassActors[0].BypassMode != "pull_request" {
		t.Errorf("only the admin role may bypass, through a pull request: %+v", r.BypassActors)
	}
}

func TestTheTagRulesetLetsOnlyTheOwnerTagARelease(t *testing.T) {
	var r ruleset
	readJSON(t, "../platform/ruleset-tags.json", &r)
	if r.Target != "tag" || !slices.Equal(r.Conditions.RefName.Include, []string{"~ALL"}) {
		t.Errorf("the ruleset must hold every tag: target %q, include %v", r.Target, r.Conditions.RefName.Include)
	}
	var types []string
	for _, rule := range r.Rules {
		types = append(types, rule.Type)
	}
	for _, want := range []string{"creation", "update", "deletion"} {
		if !slices.Contains(types, want) {
			t.Errorf("the tag ruleset lacks a %s rule", want)
		}
	}
	if len(r.BypassActors) != 1 || r.BypassActors[0].ActorType != "RepositoryRole" || r.BypassActors[0].BypassMode != "always" {
		t.Errorf("only the admin role may tag, from its own clone: %+v", r.BypassActors)
	}
}

type claudeSettings struct {
	Permissions struct {
		Deny []string `json:"deny"`
	} `json:"permissions"`
	Hooks map[string][]struct {
		Matcher string `json:"matcher"`
		Hooks   []struct {
			Command string `json:"command"`
		} `json:"hooks"`
	} `json:"hooks"`
}

func TestTheAgentSettingsDenyBypassesAndRunTheKitHooks(t *testing.T) {
	var s claudeSettings
	readJSON(t, "../platform/claude/settings.json", &s)
	for _, want := range []string{"Bash(git commit --no-verify *)", "Bash(git push --force *)", "Bash(gh pr merge *)", "Edit(./.kit/**)", "Edit(./.github/**)"} {
		if !slices.Contains(s.Permissions.Deny, want) {
			t.Errorf("deny lacks %s", want)
		}
	}
	for event, command := range map[string]string{"PostToolUse": "kit hook post-edit", "Stop": "kit hook stop"} {
		groups := s.Hooks[event]
		if len(groups) == 0 || len(groups[0].Hooks) == 0 || !strings.HasPrefix(groups[0].Hooks[0].Command, command) {
			t.Errorf("%s hook does not run %s: %+v", event, command, groups)
		}
	}
}

func TestTheAppCodeOwnersProtectWhatJudgesTheAgent(t *testing.T) {
	data, err := os.ReadFile("../template/.github/CODEOWNERS")
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/.kit/", "/.github/", "/.claude/", "/go.mod", "*_test.go", "**/testdata/", "/AGENTS.md", "**/GLOSSARY.md", "/renovate.json"} {
		if !strings.Contains(string(data), "\n"+path+" @OWNER") {
			t.Errorf("CODEOWNERS leaves %s to the agent", path)
		}
	}
}

func hookRun(t *testing.T, input string, args ...string) (string, int) {
	t.Helper()
	cmd := exec.CommandContext(t.Context(), kitCommand(t, "kit"), append([]string{"hook"}, args...)...)
	cmd.Stdin = strings.NewReader(input)
	out, err := cmd.CombinedOutput()
	if err == nil {
		return string(out), 0
	}
	return string(out), exitCode(err)
}

func editOf(t *testing.T, file string) string {
	t.Helper()
	data, err := json.Marshal(map[string]any{"tool_name": "Edit", "tool_input": map[string]string{"file_path": filepath.ToSlash(file)}})
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// WHY: a rule the agent meets only at the end is learned late; after each
// edit, the package's findings reach it through the hook's exit code 2.
func TestThePostEditHookTellsTheAgentWhatItsEditBroke(t *testing.T) {
	dir := overlay(t, caseNamed(t, "m1-simplicity"))
	out, code := hookRun(t, editOf(t, filepath.Join(dir, "internal/notes/internal/m1case/m1.go")), "post-edit")
	if code != 2 || !strings.Contains(out, "cognitive complexity 11") || !strings.Contains(out, "nested 3 deep") {
		t.Errorf("a broken package: exit %d\n%s", code, out)
	}
	clean, err := filepath.Abs(filepath.Join(cleanDir, "internal/notes/internal/add/add.go"))
	if err != nil {
		t.Fatal(err)
	}
	if out, code := hookRun(t, editOf(t, clean), "post-edit"); code != 0 || out != "" {
		t.Errorf("a clean package: exit %d, said %q", code, out)
	}
}

func TestTheStopHookLetsTheAgentStopOnItsSecondTry(t *testing.T) {
	if out, code := hookRun(t, `{"cwd":".","stop_hook_active":true}`, "stop"); code != 0 {
		t.Errorf("stop_hook_active: exit %d\n%s", code, out)
	}
}
