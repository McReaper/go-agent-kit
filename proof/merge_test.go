package proof_test

import (
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/McReaper/rumblestrip-go/setup"
)

// WHY: a pull request of a stack takes no auto-merge, and GitHub rebases
// the next one onto the default branch only once the one below merged: the
// Rumblestrip's merge job merges through the asynchronous merge, and says why when
// the ruleset refuses. The commit is the pull request's title and
// description, with the owner as co-author and Rumblestrip's release, never
// the agent's commits' messages; Rumblestrip's comment says where it went.
func TestTheMergeJobMergesThroughTheAsynchronousMergeAndSaysWhyGitHubRefused(t *testing.T) {
	script := stepRun(t, "../.github/workflows/rumblestrip-check.yml", "Merge")
	for _, c := range []struct {
		name, state, result string
		code                int
		merges              bool
		says, status        string
	}{
		{"green", "open false h1", "merged", 0, true, "merged", "✅ Merged into `main` as m1. Next in the stack: #9."},
		{"refused", "open false h1", "failed Repository rule violations found", 1, true, "::error title=GitHub refused the merge::Repository rule violations found", "❌ GitHub refused the merge: Repository rule violations found"},
		{"merged by an earlier run", "closed true h1", "", 0, false, "already merged", ""},
		{"pushed to since", "open false h2", "", 0, false, "its own run merges it", ""},
	} {
		t.Run(c.name, func(t *testing.T) {
			tmp := t.TempDir()
			fake := mergeFake + `sleep() { :; }
`
			out, code, log := runStep(t, fake+script, "PR=7", "HEAD=h1", "STATE="+c.state, "RESULT="+c.result,
				"OWNER=octo", "RELEASE=v0.2.8", "DEFAULT=main", "RUNNER_TEMP="+tmp, "GITHUB_STEP_SUMMARY="+filepath.Join(tmp, "summary.md"))
			if code != c.code || !strings.Contains(out, c.says) {
				t.Errorf("exit %d, want %d; output lacks %q:\n%s", code, c.code, c.says, out)
			}
			if merged := strings.Contains(log, " -X PUT repos/o/r/pulls/7/merge-async "); merged != c.merges {
				t.Errorf("merged %v, want %v:\n%s", merged, c.merges, log)
			}
			if !c.merges {
				return
			}
			if strings.Contains(log, "2026-03-10 repos/o/r/pulls/7 --jq .merge_commit_sha") {
				t.Error("the merge commit is read with the API version that has no merge_commit_sha")
			}
			wantMessage := "Why it changes.\n\nCo-authored-by: octo <1+octo@users.noreply.github.com>\nBuilt-with: Rumblestrip v0.2.8 <https://github.com/McReaper/rumblestrip-go>"
			for file, want := range map[string]string{
				"title":   "feat(notes): a short title is refused (#7)",
				"message": wantMessage,
				"comment": "<!-- rumblestrip -->\n<!-- rumblestrip status -->\n" + c.status + "\n<!-- /rumblestrip status -->\n\n## For you\n",
			} {
				if got := readFile(t, filepath.Join(tmp, file)); got != want {
					t.Errorf("%s = %q, want %q", file, got, want)
				}
			}
		})
	}
}

// WHY: the merge job reads pull request #7 piece by piece, whose
// description ends with Rumblestrip's line; the fake records the commit's title
// and message and the comment the job writes, in $RUNNER_TEMP, for the
// proof to read.
const mergeFake = `gh() {
  echo "$*" >> "$LOG"
  for a in "$@"; do
    case $a in
    commit_title=*) printf '%s' "${a#commit_title=}" > "$RUNNER_TEMP/title" ;;
    commit_message=*) printf '%s' "${a#commit_message=}" > "$RUNNER_TEMP/message" ;;
    body=@*) cp "${a#body=@}" "$RUNNER_TEMP/comment" ;;
    esac
  done
  case "$*" in
  *merge-async/u1*) echo "$RESULT" ;;
  *merge-async*) echo u1 ;;
  *"pulls/7 --jq .title") echo "feat(notes): a short title is refused" ;;
  *"pulls/7 --jq .body"*) printf 'Why it changes.\r\n\r\n\360\237\247\260 Built with [Rumblestrip](https://github.com/McReaper/rumblestrip-go)\n' ;;
  *"pulls/7 --jq .head.ref") echo feat/x ;;
  *"pulls/7 --jq .merge_commit_sha"*) echo m1 ;;
  *"pulls/7 --jq"*) echo "$STATE" ;;
  *users/octo*) echo "1+octo@users.noreply.github.com" ;;
  *base=feat/x*) echo "#9" ;;
  *issues/7/comments*) echo 55 ;;
  *"comments/55 --jq .body") printf '%s\n' "<!-- rumblestrip -->" "<!-- rumblestrip status -->" "⏳ The checks are running." "<!-- /rumblestrip status -->" "" "## For you" ;;
  esac
}
`

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// WHY: GitHub moves a pull request onto the default branch once the one
// below it merged only when both are a stack; outside one it stays on a
// merged branch and never merges. The check says so, with the command.
func TestAPullRequestOnAnotherBranchIsInAStackOrTheCheckSaysHow(t *testing.T) {
	script := stepRun(t, "../.github/workflows/rumblestrip-check.yml", "Is a pull request on another branch than the default one in a stack?")
	fake := `gh() { echo "$*" >> "$LOG"; echo "$STACK"; }
sleep() { :; }
`
	out, code, _ := runStep(t, fake+script, "PR=9", "BASE_PR=spec/x", "STACK=3")
	if code != 0 {
		t.Errorf("in a stack: exit %d\n%s", code, out)
	}
	out, code, _ = runStep(t, fake+script, "PR=9", "BASE_PR=spec/x", "STACK=")
	if code != 1 || !strings.Contains(out, "gh api -X POST repos/o/r/stacks") || !strings.Contains(out, "pull_requests[]=9") {
		t.Errorf("in no stack: exit %d\n%s", code, out)
	}
}

// WHY: after a merge by the workflow's own token GitHub deletes no branch,
// and moves a stack's next pull request onto the default branch without
// rebasing it or running its check: the merge job merges as an App, and
// says how to give it one rather than merging as the workflow.
func TestTheMergeJobMergesAsTheRepositorysAppOrSaysHowToGiveItOne(t *testing.T) {
	script := stepRun(t, "../.github/workflows/rumblestrip-check.yml", "Has the repository an App to merge as?")
	if out, code, _ := runStep(t, script, "HAS_APP=true"); code != 0 {
		t.Errorf("with an App: exit %d\n%s", code, out)
	}
	out, code, _ := runStep(t, script, "HAS_APP=false")
	if code != 1 || !strings.Contains(out, "rumblestrip init github --repo o/r") {
		t.Errorf("with no App: exit %d, want 1 and the command\n%s", code, out)
	}
	merge := jobs(t, "../.github/workflows/rumblestrip-check.yml")["merge"]
	for _, want := range []string{
		"    permissions: {}\n",
		"uses: actions/create-github-app-token@",
		"client-id: ${{ secrets.RUMBLESTRIP_APP_CLIENT_ID }}\n",
		"private-key: ${{ secrets.RUMBLESTRIP_APP_KEY }}\n",
		"GH_TOKEN: ${{ steps.app.outputs.token }}\n",
	} {
		if !strings.Contains(merge, want) {
			t.Errorf("rumblestrip-check.yml's merge job lacks %q", want)
		}
	}
	if strings.Contains(merge, "github.token") {
		t.Errorf("rumblestrip-check.yml's merge job uses the workflow's own token")
	}
	dir := t.TempDir()
	fs, err := setup.Files(setup.Options{Version: setup.Version{Tag: "v0.2.0", Commit: strings.Repeat("ab", 20)}, Owner: "octo", OwnerEmail: "1+octo@users.noreply.github.com"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := setup.Write(dir, fs); err != nil {
		t.Fatal(err)
	}
	rumblestrip := jobs(t, filepath.Join(dir, ".github", "workflows", "rumblestrip.yml"))["pr"]
	for _, want := range []string{
		"      RUMBLESTRIP_APP_CLIENT_ID: ${{ secrets.RUMBLESTRIP_APP_CLIENT_ID }}\n",
		"      RUMBLESTRIP_APP_KEY: ${{ secrets.RUMBLESTRIP_APP_KEY }}\n",
	} {
		if !strings.Contains(rumblestrip, want) {
			t.Errorf("rumblestrip.yml's rumblestrip job does not pass %q", strings.TrimSpace(want))
		}
	}
	if strings.Contains(rumblestrip, "inherit") || strings.Contains(rumblestrip, "contents: write") {
		t.Errorf("rumblestrip.yml's rumblestrip job passes every secret or lets the workflow write the code:\n%s", rumblestrip)
	}
}

// WHY: the owner's approval status is the review sheet's job's to post,
// and Check and Merge run after it: the approval re-runs that job, in the
// run of the approved commit, once the run finished; and nothing when the
// status already says approved.
func TestAnApprovalReRunsTheReviewSheetOfTheApprovedCommitsRun(t *testing.T) {
	script := stepRun(t, "../.github/workflows/rumblestrip-approval.yml", "Re-run the review sheet of the approved commit")
	fake := `gh() {
  echo "$*" >> "$LOG"
  case "$*" in
  *commits/h1/status*) echo "$STATE" ;;
  *"actions/runs?head_sha=h1"*) echo 11 ;;
  *runs/11/jobs*) echo 101 ;;
  esac
}
`
	for state, want := range map[string][]string{
		"pending": {"run watch 11 --repo o/r", "run rerun --job 101 --repo o/r"},
		"success": nil,
	} {
		out, code, log := runStep(t, fake+script, "HEAD=h1", "STATE="+state, "WORKFLOW=.github/workflows/rumblestrip.yml")
		if code != 0 {
			t.Fatalf("%s: exit %d\n%s", state, code, out)
		}
		if got := gestures(log); strings.Join(got, "\n") != strings.Join(want, "\n") {
			t.Errorf("%s: gestures %q, want %q", state, got, want)
		}
		if state == "pending" && !strings.Contains(log, `select(.path == ".github/workflows/rumblestrip.yml")`) {
			t.Errorf("the runs listed are not rumblestrip.yml's only:\n%s", log)
		}
	}
}

// WHY: Check is the owner's line in the checks list and the comment's
// status: it names the gates that failed, with their jobs, and says what
// happens next; with every gate green, it passes. Its summary is the run's
// one: the table of every check, then each failing gate's section.
func TestTheCheckJobNamesTheGatesThatFailedAndSaysWhatComesNext(t *testing.T) {
	script := stepRun(t, "../.github/workflows/rumblestrip-check.yml", "Did every gate pass?")
	fake := `gh() {
  echo "$*" >> "$LOG"
  for a in "$@"; do
    case $a in body=@*) cp "${a#body=@}" "$RUNNER_TEMP/comment" ;; esac
  done
  case "$*" in
  *runs/5/jobs*) printf '%b' "$JOBS" ;;
  *issues/7/comments*) echo 55 ;;
  *"comments/55 --jq .body") printf '%s\n' "<!-- rumblestrip -->" "<!-- rumblestrip status -->" "⏳ The checks are running." "<!-- /rumblestrip status -->" "" "## For you" ;;
  esac
}
`
	green := `success\tRumblestrip version\thttps://x/0\nsuccess\tTests\thttps://x/1\nsuccess\tSpec\thttps://x/2\n`
	for _, c := range []struct {
		name, jobs, needs, approval, base string
		code                              int
		status                            string
	}{
		{"green", green, "success success success", "none", "main", 0, "✅ Every check passes. Rumblestrip merges it now."},
		{"waiting for the owner", green, "success success success", "waiting", "main", 0, "✅ Every check passes. **Waiting for you:** read what is below, then approve the pull request; Rumblestrip merges it then."},
		{"stacked", green, "success success success", "none", "spec/x", 0, "✅ Every check passes. It merges once the pull request below it in the stack has."},
		{
			"a gate fails", `success\tRumblestrip version\thttps://x/0\nfailure\tTests\thttps://x/1\nsuccess\tSpec\thttps://x/2\n`, "success failure success", "none", "main", 1,
			"❌ 1 check fails: [Tests](https://x/1). The agent fixes them; the checks run again on its next push.",
		},
		{
			"two fail", `success\tRumblestrip version\thttps://x/0\nfailure\tTests\thttps://x/1\ncancelled\tSpec\thttps://x/2\n`, "success failure success", "waiting", "main", 1,
			"❌ 2 checks fail: [Tests](https://x/1), [Spec](https://x/2). The agent fixes them; the checks run again on its next push. Meanwhile, what is yours to approve is below.",
		},
		{
			"a job before failed unseen", green, "success success failure", "", "main", 1,
			"❌ A job of Rumblestrip did not finish: re-run the failed jobs. That is Rumblestrip's or the runner's fault, not the change's.",
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			tmp := t.TempDir()
			section := "### ❌ Tests\n\nIt wants: every test passes.\n"
			files := map[string]string{"go-test.row": "Tests\tEvery test passes.\n", "spec.row": "Spec\tEach built scenario passes.\n"}
			if c.code == 1 {
				files["go-test.md"] = section
			}
			if err := os.MkdirAll(filepath.Join(tmp, "rumblestrip-summary"), 0o750); err != nil {
				t.Fatal(err)
			}
			for name, content := range files {
				if err := os.WriteFile(filepath.Join(tmp, "rumblestrip-summary", name), []byte(content), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			out, code, _ := runStep(t, fake+script, "PR=7", "RUN=5", "JOBS="+c.jobs, "NEEDS="+c.needs, "APPROVAL="+c.approval, "BASE="+c.base, "DEFAULT=main",
				"RUNNER_TEMP="+tmp, "GITHUB_STEP_SUMMARY="+filepath.Join(tmp, "summary.md"))
			if code != c.code {
				t.Errorf("exit %d, want %d\n%s", code, c.code, out)
			}
			want := "<!-- rumblestrip -->\n<!-- rumblestrip status -->\n" + c.status + "\n<!-- /rumblestrip status -->\n\n## For you\n"
			if got := readFile(t, filepath.Join(tmp, "comment")); got != want {
				t.Errorf("comment = %q, want %q", got, want)
			}
			summary := readFile(t, filepath.Join(tmp, "summary.md"))
			if !strings.HasPrefix(summary, c.status+"\n") || strings.Contains(summary, section) != (c.code == 1) {
				t.Errorf("summary:\n%s", summary)
			}
			for _, row := range []string{
				"| ✅ | [Rumblestrip version](https://x/0) | Which commit of Rumblestrip runs: the one this repository pins. |\n",
				"| [Tests](https://x/1) | Every test passes. |\n",
				"| [Spec](https://x/2) | Each built scenario passes. |\n",
			} {
				if !strings.Contains(summary, row) {
					t.Errorf("the summary's table lacks %q:\n%s", row, summary)
				}
			}
		})
	}
}

func runStep(t *testing.T, script string, env ...string) (string, int, string) {
	t.Helper()
	dir := t.TempDir()
	path, log := filepath.Join(dir, "step.sh"), filepath.Join(dir, "log")
	if err := os.WriteFile(path, []byte(script), 0o600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.CommandContext(t.Context(), "sh", "-c", `. "$1"`, "sh", path)
	cmd.Env = append(append(os.Environ(), "LOG="+log, "REPO=o/r", "GH_TOKEN=x", "RUNNER_TEMP="+dir, "GITHUB_STEP_SUMMARY="+filepath.Join(dir, "summary.md")), env...)
	out, err := cmd.CombinedOutput()
	code := 0
	if exit, ok := errors.AsType[*exec.ExitError](err); ok {
		code = exit.ExitCode()
	} else if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(log)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		t.Fatal(err)
	}
	return string(out), code, string(data)
}

func gestures(log string) []string {
	var out []string
	for line := range strings.Lines(log) {
		if strings.Contains(line, " -X PUT ") || strings.HasPrefix(line, "run ") {
			out = append(out, strings.TrimSpace(line))
		}
	}
	return out
}

func stepRun(t *testing.T, path, step string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	_, rest, ok := strings.Cut(string(data), "- name: "+step+"\n")
	if !ok {
		t.Fatalf("%s has no step %q", path, step)
	}
	head, rest, ok := strings.Cut(rest, "run: |\n")
	if !ok || strings.Contains(head, "- name: ") {
		t.Fatalf("step %q of %s has no run block", step, path)
	}
	indent := head[strings.LastIndex(head, "\n")+1:] + "  "
	var b strings.Builder
	for line := range strings.Lines(rest) {
		body, ok := strings.CutPrefix(line, indent)
		if !ok && strings.TrimSpace(line) != "" {
			break
		}
		b.WriteString(body)
	}
	return b.String()
}
