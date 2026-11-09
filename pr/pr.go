// Package pr opens an agent's pull request the way Rumblestrip's check
// wants it: a short conventional title, and, on another branch than the
// default one, stacked on the pull request of that branch, so GitHub moves
// it onto the default branch once the one below merged.
package pr

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/McReaper/rumblestrip-go/internal/platform/ghapi"
)

// MaxTitle is the longest pull request title: GitHub cuts a longer one in
// its lists, and the title becomes the commit's subject on the default
// branch.
const MaxTitle = 72

var conventional = regexp.MustCompile(`^(feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert|spec)(\([a-z0-9][a-z0-9/._-]*\))?!?: \S`)

// TitleProblems says what is wrong with a pull request's title, one line
// each, the rule first; none when it is a conventional commit within
// MaxTitle.
func TitleProblems(title string) []string {
	var out []string
	if n := utf8.RuneCountInString(title); n > MaxTitle {
		out = append(out, fmt.Sprintf("title: the title has %d characters, the limit is %d: say less, the description holds the rest", n, MaxTitle))
	}
	if !conventional.MatchString(title) {
		out = append(out, "title: the title is no conventional commit: write type(scope): what changes, the type one of feat, fix, docs, refactor, perf, test, build, ci, chore, revert, style, or spec for a spec's pull request")
	}
	return out
}

// Options say what to open: the branch with the change, the one it goes
// into (empty for the default branch), and what the pull request says.
type Options struct {
	Head  string
	Base  string
	Title string
	Body  string
}

// Opened is a pull request, and the one it is stacked on, 0 when it goes
// into the default branch.
type Opened struct {
	Number    int
	URL       string
	StackedOn int
}

// WHY: GitHub's stacks answer only to this API version.
const stacksVersion = "2026-03-10"

var errNoPullRequest = errors.New("no open pull request")

// Open opens the pull request through api, in the repository gh's {owner}
// and {repo} name, and stacks it on the pull request of its base when that
// is not the default branch. A title Rumblestrip's check would refuse is
// refused before anything is opened.
func Open(ctx context.Context, api ghapi.API, o Options) (Opened, error) {
	if ps := TitleProblems(o.Title); len(ps) > 0 {
		return Opened{}, fmt.Errorf("pr: %s", strings.Join(ps, "; "))
	}
	g, err := connect(ctx, api)
	if err != nil {
		return Opened{}, err
	}
	if o.Base == "" {
		o.Base = g.defaultBranch
	}
	below := 0
	if o.Base != g.defaultBranch {
		below, err = g.pullRequestOf(ctx, o.Base)
		if err != nil {
			return Opened{}, err
		}
	}
	opened, err := g.create(ctx, o)
	if err != nil || below == 0 {
		return opened, err
	}
	if err := g.stack(ctx, below, opened.Number); err != nil {
		return opened, fmt.Errorf("pr: #%d is open, but not stacked on #%d: %w", opened.Number, below, err)
	}
	opened.StackedOn = below
	return opened, nil
}

// Stack stacks the open pull request of head on the pull request of the
// branch it goes into: for one opened apart, which the check refuses.
func Stack(ctx context.Context, api ghapi.API, head string) (Opened, error) {
	g, err := connect(ctx, api)
	if err != nil {
		return Opened{}, err
	}
	opened, base, err := g.openPullRequest(ctx, head)
	if err != nil {
		return Opened{}, err
	}
	if base == g.defaultBranch {
		return opened, fmt.Errorf("pr: #%d goes into the default branch: there is nothing to stack it on", opened.Number)
	}
	below, err := g.pullRequestOf(ctx, base)
	if err != nil {
		return opened, err
	}
	if err := g.stack(ctx, below, opened.Number); err != nil {
		return opened, fmt.Errorf("pr: #%d is not stacked on #%d: %w", opened.Number, below, err)
	}
	opened.StackedOn = below
	return opened, nil
}

// WHY: every call needs the repository's owner, for a branch's pull
// request, and its default branch: read once.
type github struct {
	api           ghapi.API
	owner         string
	defaultBranch string
}

func connect(ctx context.Context, api ghapi.API) (github, error) {
	var r struct {
		DefaultBranch string `json:"default_branch"`
		Owner         struct {
			Login string `json:"login"`
		} `json:"owner"`
	}
	g := github{api: api}
	if err := g.call(ctx, ghapi.Request{Method: "GET", Path: "repos/{owner}/{repo}"}, &r); err != nil {
		return github{}, err
	}
	g.owner, g.defaultBranch = r.Owner.Login, r.DefaultBranch
	return g, nil
}

type pullRequest struct {
	Number  int    `json:"number"`
	HTMLURL string `json:"html_url"`
	Base    struct {
		Ref string `json:"ref"`
	} `json:"base"`
}

func (g github) pullRequests(ctx context.Context, branch string) ([]pullRequest, error) {
	var prs []pullRequest
	q := url.Values{"state": {"open"}, "head": {g.owner + ":" + branch}}
	err := g.call(ctx, ghapi.Request{Method: "GET", Path: "repos/{owner}/{repo}/pulls?" + q.Encode()}, &prs)
	return prs, err
}

func (g github) pullRequestOf(ctx context.Context, branch string) (int, error) {
	prs, err := g.pullRequests(ctx, branch)
	if err != nil {
		return 0, err
	}
	if len(prs) == 0 {
		return 0, fmt.Errorf("pr: the branch %s has %w: open its pull request first, then this one on it", branch, errNoPullRequest)
	}
	return prs[0].Number, nil
}

func (g github) openPullRequest(ctx context.Context, head string) (Opened, string, error) {
	prs, err := g.pullRequests(ctx, head)
	if err != nil {
		return Opened{}, "", err
	}
	if len(prs) == 0 {
		return Opened{}, "", fmt.Errorf("pr: the branch %s has %w: open it with rumblestrip pr create", head, errNoPullRequest)
	}
	return Opened{Number: prs[0].Number, URL: prs[0].HTMLURL}, prs[0].Base.Ref, nil
}

func (g github) create(ctx context.Context, o Options) (Opened, error) {
	body, err := json.Marshal(map[string]string{"title": o.Title, "head": o.Head, "base": o.Base, "body": o.Body})
	if err != nil {
		return Opened{}, fmt.Errorf("pr: %w", err)
	}
	var created pullRequest
	err = g.call(ctx, ghapi.Request{Method: "POST", Path: "repos/{owner}/{repo}/pulls", Body: body}, &created)
	return Opened{Number: created.Number, URL: created.HTMLURL}, err
}

func (g github) stack(ctx context.Context, below, above int) error {
	body, err := json.Marshal(map[string][]int{"pull_requests": {below, above}})
	if err != nil {
		return fmt.Errorf("pr: %w", err)
	}
	_, err = g.api.Call(ctx, ghapi.Request{Method: "POST", Path: "repos/{owner}/{repo}/stacks", Version: stacksVersion, Body: body})
	return err
}

func (g github) call(ctx context.Context, r ghapi.Request, out any) error {
	data, err := g.api.Call(ctx, r)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("pr: %s %s: %w", r.Method, r.Path, err)
	}
	return nil
}
