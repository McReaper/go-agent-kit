package pr_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/McReaper/rumblestrip-go/internal/platform/ghapi"
	"github.com/McReaper/rumblestrip-go/pr"
)

type fakeAPI struct {
	calls []string
	reply map[string]string
	fail  map[string]error
}

func (f *fakeAPI) Call(_ context.Context, r ghapi.Request) ([]byte, error) {
	call := r.Method + " " + r.Path
	if r.Version != "" {
		call += " @" + r.Version
	}
	if r.Body != nil {
		call += " " + string(r.Body)
	}
	f.calls = append(f.calls, call)
	if err := f.fail[r.Method+" "+r.Path]; err != nil {
		return nil, err
	}
	return []byte(f.reply[r.Method+" "+r.Path]), nil
}

func newAPI() *fakeAPI {
	return &fakeAPI{reply: map[string]string{
		"GET repos/{owner}/{repo}":                                       `{"default_branch": "main", "owner": {"login": "octo"}}`,
		"GET repos/{owner}/{repo}/pulls?head=octo%3Aspec%2Fx&state=open": `[{"number": 4}]`,
		"POST repos/{owner}/{repo}/pulls":                                `{"number": 5, "html_url": "https://github.com/octo/app/pull/5"}`,
		"POST repos/{owner}/{repo}/stacks":                               `{}`,
	}}
}

func TestAPullRequestIntoTheDefaultBranchIsOpenedAlone(t *testing.T) {
	api := newAPI()
	got, err := pr.Open(t.Context(), api, pr.Options{Head: "spec/x", Title: "spec(notes): a blank title is refused", Body: "Why."})
	if err != nil || got != (pr.Opened{Number: 5, URL: "https://github.com/octo/app/pull/5"}) {
		t.Fatalf("Open = %+v, %v", got, err)
	}
	want := []string{
		"GET repos/{owner}/{repo}",
		`POST repos/{owner}/{repo}/pulls {"base":"main","body":"Why.","head":"spec/x","title":"spec(notes): a blank title is refused"}`,
	}
	if diff := cmp.Diff(want, api.calls); diff != "" {
		t.Errorf("calls (-want +got):\n%s", diff)
	}
}

// WHY: GitHub moves a pull request onto the default branch once the one
// below it merged only when both are a stack; opened apart, the agent
// forgets the second step, and the check waits for it, then fails.
func TestAPullRequestOnAnotherBranchIsStackedOnThatBranchsPullRequest(t *testing.T) {
	api := newAPI()
	got, err := pr.Open(t.Context(), api, pr.Options{Head: "feat/x", Base: "spec/x", Title: "feat(notes): a blank title is refused"})
	if err != nil || got.Number != 5 || got.StackedOn != 4 {
		t.Fatalf("Open = %+v, %v", got, err)
	}
	if last := api.calls[len(api.calls)-1]; last != `POST repos/{owner}/{repo}/stacks @2026-03-10 {"pull_requests":[4,5]}` {
		t.Errorf("the last call is %q, want the stack of #4 and #5", last)
	}
}

func TestABaseWithNoPullRequestOpensNothing(t *testing.T) {
	api := newAPI()
	api.reply["GET repos/{owner}/{repo}/pulls?head=octo%3Aspec%2Fx&state=open"] = `[]`
	_, err := pr.Open(t.Context(), api, pr.Options{Head: "feat/x", Base: "spec/x", Title: "feat(notes): a blank title is refused"})
	if err == nil || !strings.Contains(err.Error(), "the branch spec/x has no open pull request: open its pull request first") {
		t.Errorf("err = %v", err)
	}
	for _, c := range api.calls {
		if strings.HasPrefix(c, "POST") {
			t.Errorf("it opened something: %s", c)
		}
	}
}

func TestAStackGitHubRefusesSaysThePullRequestIsOpen(t *testing.T) {
	api := newAPI()
	api.fail = map[string]error{"POST repos/{owner}/{repo}/stacks": errors.New("HTTP 422")}
	got, err := pr.Open(t.Context(), api, pr.Options{Head: "feat/x", Base: "spec/x", Title: "feat(notes): a blank title is refused"})
	if err == nil || got.Number != 5 || !strings.Contains(err.Error(), "#5 is open, but not stacked on #4: HTTP 422") {
		t.Errorf("Open = %+v, %v", got, err)
	}
}

// WHY: a pull request opened with gh apart from its stack is the one the
// check refuses; stacking it afterwards is one command, which the check's
// message names.
func TestAPullRequestOpenedApartIsStackedOnItsBasesPullRequest(t *testing.T) {
	api := newAPI()
	api.reply["GET repos/{owner}/{repo}/pulls?head=octo%3Afeat%2Fx&state=open"] = `[{"number": 5, "html_url": "u", "base": {"ref": "spec/x"}}]`
	got, err := pr.Stack(t.Context(), api, "feat/x")
	if err != nil || got.Number != 5 || got.StackedOn != 4 {
		t.Fatalf("Stack = %+v, %v", got, err)
	}
	if last := api.calls[len(api.calls)-1]; last != `POST repos/{owner}/{repo}/stacks @2026-03-10 {"pull_requests":[4,5]}` {
		t.Errorf("the last call is %q", last)
	}
	api.reply["GET repos/{owner}/{repo}/pulls?head=octo%3Afeat%2Fx&state=open"] = `[{"number": 5, "html_url": "u", "base": {"ref": "main"}}]`
	if _, err := pr.Stack(t.Context(), api, "feat/x"); err == nil || !strings.Contains(err.Error(), "goes into the default branch") {
		t.Errorf("a pull request into main: %v", err)
	}
}

func TestATitleTheCheckWouldRefuseOpensNothing(t *testing.T) {
	for _, title := range []string{"Refuse blank titles", "feat(notes): " + strings.Repeat("a", 60)} {
		api := newAPI()
		if _, err := pr.Open(t.Context(), api, pr.Options{Head: "feat/x", Title: title}); err == nil || len(api.calls) != 0 {
			t.Errorf("%q: err %v, calls %v", title, err, api.calls)
		}
	}
	if got := pr.TitleProblems("fix!: the list starts at the newest note"); len(got) != 0 {
		t.Errorf("a breaking fix's title: %v", got)
	}
}
