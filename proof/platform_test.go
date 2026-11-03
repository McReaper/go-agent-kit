package proof_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/McReaper/go-agent-kit/guard"
	"github.com/McReaper/go-agent-kit/kitinit"
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

func TestTheKitsRulesetLetsOnlyTheOwnerMergeAnApprovedCheckedChange(t *testing.T) {
	var r ruleset
	readJSON(t, "../platform/ruleset-kit.json", &r)
	var types []string
	for _, rule := range r.Rules {
		types = append(types, rule.Type)
		if rule.Type == "pull_request" {
			// WHY: the merge needs an approval of the last push; earlier
			// approvals stay on record for the merge check, which keeps a
			// protected file approved while its content is the one approved.
			for _, p := range []string{`"require_code_owner_review": true`, `"dismiss_stale_reviews_on_push": false`, `"require_last_push_approval": true`} {
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

// WHY: an app's change merges by itself once green. GitHub asks no review
// but the code owners' on the harness and keeps every approval on record;
// the kit check, which only GitHub Actions may report, holds the rest that
// is the owner's against the content they approved. No update rule, so
// GitHub's auto-merge can land the change.
func TestTheAppsRulesetMergesAGreenChangeAndHoldsWhatIsTheOwners(t *testing.T) {
	var r ruleset
	readJSON(t, "../kitinit/files/ruleset.json", &r)
	var types []string
	for _, rule := range r.Rules {
		types = append(types, rule.Type)
		var want []string
		switch rule.Type {
		case "pull_request":
			want = []string{`"required_approving_review_count": 0`, `"require_code_owner_review": true`, `"dismiss_stale_reviews_on_push": false`, `"require_last_push_approval": false`}
		case "required_status_checks":
			want = []string{`{ "context": "kit", "integration_id": 15368 }`}
		}
		for _, p := range want {
			if !strings.Contains(string(rule.Parameters), p) {
				t.Errorf("%s rule lacks %s", rule.Type, p)
			}
		}
	}
	if slices.Contains(types, "update") {
		t.Error("an update rule stops GitHub's auto-merge")
	}
	for _, want := range []string{"pull_request", "required_status_checks", "non_fast_forward", "deletion"} {
		if !slices.Contains(types, want) {
			t.Errorf("the ruleset lacks a %s rule", want)
		}
	}
}

func TestTheTagRulesetLetsOnlyTheOwnerTagARelease(t *testing.T) {
	var r ruleset
	readJSON(t, "../kitinit/files/ruleset-tags.json", &r)
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
	EnabledPlugins map[string]bool `json:"enabledPlugins"`
	Permissions    struct {
		Deny []string `json:"deny"`
	} `json:"permissions"`
	Hooks map[string][]struct {
		Matcher string `json:"matcher"`
		Hooks   []struct {
			Command string `json:"command"`
		} `json:"hooks"`
	} `json:"hooks"`
}

func initFiles(t *testing.T) map[string]string {
	t.Helper()
	fs, err := kitinit.Files(kitinit.Options{Version: kitinit.Version{Tag: "v0.2.0", Commit: strings.Repeat("ab", 20)}, Owner: "OWNER"})
	if err != nil {
		t.Fatal(err)
	}
	out := map[string]string{}
	for _, f := range fs {
		out[f.Path] = string(f.Content)
	}
	return out
}

// WHY: deny rules a plugin can't carry; the repository's settings, written
// by kit init, hold them and turn the kit's plugin on, managed settings do
// both where the agent can't undo them, and the plugin's hooks refuse what
// a deny rule's text misses (git -C . push).
func TestTheAgentSettingsDenyBypassesAndThePluginRunsTheKitHooks(t *testing.T) {
	var project, managed, plugin claudeSettings
	if err := json.Unmarshal([]byte(initFiles(t)[".claude/settings.json"]), &project); err != nil {
		t.Fatal(err)
	}
	readJSON(t, "../platform/claude/managed-settings.json", &managed)
	readJSON(t, "../plugin/hooks/hooks.json", &plugin)
	for _, want := range []string{"Bash(git commit --no-verify *)", "Bash(git push --force *)", "Bash(git push --force-with-lease * main)", "Bash(gh pr merge *)", "Edit(./.github/**)"} {
		if !slices.Contains(project.Permissions.Deny, want) {
			t.Errorf("deny lacks %s", want)
		}
	}
	// WHY: the agent rebases its own pull request while the owner hasn't
	// approved it; main's ruleset refuses any force push there anyway.
	if slices.Contains(project.Permissions.Deny, "Bash(git push --force-with-lease *)") {
		t.Error("deny refuses --force-with-lease on the agent's own branches")
	}
	if !slices.Equal(project.Permissions.Deny, managed.Permissions.Deny) || !project.EnabledPlugins["kit@go-agent-kit"] || !managed.EnabledPlugins["kit@go-agent-kit"] {
		t.Errorf("the repository's and the managed settings differ: %+v, %+v", project, managed)
	}
	for event, command := range map[string]string{"PreToolUse": "hook pre-tool", "PostToolUse": "hook post-edit", "Stop": "hook stop"} {
		groups := plugin.Hooks[event]
		if len(groups) == 0 || len(groups[0].Hooks) == 0 || !strings.Contains(groups[0].Hooks[0].Command, `"${CLAUDE_PLUGIN_ROOT}/bin/kit" `+command) {
			t.Errorf("the plugin's %s hook does not run kit %s: %+v", event, command, groups)
		}
	}
}

func TestTheAppCodeOwnersHoldTheHarnessAndLeaveTheSpecToTheCheck(t *testing.T) {
	var owned []string
	for l := range strings.Lines(initFiles(t)[".github/CODEOWNERS"]) {
		if p, ok := strings.CutSuffix(strings.TrimSpace(l), " @OWNER"); ok {
			owned = append(owned, p)
		}
	}
	var want []string
	for _, h := range guard.Harness() {
		if h == "GLOSSARY.md" {
			want = append(want, "**/"+h)
		} else {
			want = append(want, "/"+h)
		}
	}
	if !slices.Equal(owned, want) {
		t.Errorf("CODEOWNERS owns %q, the guard's harness is %q", owned, want)
	}
}

// WHY: what kit init writes is what every app's CI runs; it passes the
// workflow gates the kit holds an app's workflows to.
func TestTheCIKitInitWritesPassesTheWorkflowGates(t *testing.T) {
	withZizmor(t)
	dir := t.TempDir()
	fs, err := kitinit.Files(kitinit.Options{Version: kitinit.Version{Tag: "v0.2.0", Commit: strings.Repeat("ab", 20)}, Owner: "octo"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := kitinit.Write(dir, fs); err != nil {
		t.Fatal(err)
	}
	gitIn(t, dir, "init", "-q")
	gitIn(t, dir, "add", "-A")
	check(t, nil, actionlintFindings(t, dir))
	check(t, nil, zizmorFindings(t, dir))
}

// WHY: the plugin is how an app gets the kit's skills, rules and hooks;
// every file its manifests and hooks name must be there, the wrapper they
// call executable, and each skill named as Claude Code lists it.
func TestThePluginShipsWhatItsManifestsAndHooksName(t *testing.T) {
	var market struct {
		Plugins []struct {
			Name   string `json:"name"`
			Source string `json:"source"`
		} `json:"plugins"`
	}
	readJSON(t, "../.claude-plugin/marketplace.json", &market)
	if len(market.Plugins) != 1 || market.Plugins[0].Name != "kit" || market.Plugins[0].Source != "./plugin" {
		t.Fatalf("the marketplace: %+v", market)
	}
	var manifest struct {
		Name string `json:"name"`
	}
	readJSON(t, "../plugin/.claude-plugin/plugin.json", &manifest)
	if manifest.Name != "kit" {
		t.Errorf("the plugin is named %q", manifest.Name)
	}
	mode, err := exec.CommandContext(t.Context(), "git", "ls-files", "-s", "../plugin/bin/kit").Output()
	if err != nil || !strings.HasPrefix(string(mode), "100755 ") {
		t.Errorf("plugin/bin/kit is not executable in git: %q, %v", mode, err)
	}
	runIn(t, place{dir: ".."}, "shellcheck", "plugin/bin/kit")
	rules, err := os.ReadFile("../plugin/rules.md")
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range regexp.MustCompile("`kit:([a-z-]+)`").FindAllStringSubmatch(string(rules), -1) {
		skill, err := os.ReadFile(filepath.Join("../plugin/skills", m[1], "SKILL.md"))
		if err != nil || !strings.Contains(string(skill), "\nname: "+m[1]+"\n") {
			t.Errorf("rules.md names kit:%s, which no skill of that name holds: %v", m[1], err)
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
