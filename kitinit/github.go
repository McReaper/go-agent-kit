package kitinit

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/McReaper/go-agent-kit/internal/platform/ghapp"
	"github.com/McReaper/go-agent-kit/internal/platform/run"
)

// Request is one call to GitHub's REST API.
type Request struct {
	Method string
	Path   string
	Body   []byte
}

// API calls GitHub's REST API as the owner.
type API interface {
	Call(ctx context.Context, r Request) ([]byte, error)
}

// Secret is one of a repository's secrets, for Actions or for Dependabot.
type Secret struct {
	Repo  string
	Name  string
	For   string
	Value []byte
}

// Owner acts on a repository as its owner: its API and its secrets.
type Owner interface {
	API
	SetSecret(ctx context.Context, s Secret) error
}

// GH calls GitHub's API through gh, as whoever gh is logged in as.
type GH struct{}

// SetSecret runs gh secret set with the value on its standard input, so
// that no command line carries it.
func (GH) SetSecret(ctx context.Context, s Secret) error {
	res, err := run.Run(ctx, run.Cmd{
		Dir:   ".",
		Argv:  []string{"gh", "secret", "set", s.Name, "--repo", s.Repo, "--app", s.For},
		Stdin: bytes.NewReader(s.Value),
	})
	if err != nil {
		return err
	}
	if res.Code != 0 {
		return fmt.Errorf("gh secret set %s --app %s: %s", s.Name, s.For, strings.TrimSpace(string(res.Stderr)))
	}
	return nil
}

// Call runs gh api with r.
func (GH) Call(ctx context.Context, r Request) (out []byte, err error) {
	argv := []string{"gh", "api", "--method", r.Method, r.Path}
	if r.Body != nil {
		f, err := os.CreateTemp("", "kit-init-*.json")
		if err != nil {
			return nil, fmt.Errorf("kit init: %w", err)
		}
		defer func() { err = errors.Join(err, os.Remove(f.Name())) }()
		_, werr := f.Write(r.Body)
		if err := errors.Join(werr, f.Close()); err != nil {
			return nil, fmt.Errorf("kit init: %w", err)
		}
		argv = append(argv, "--input", f.Name())
	}
	res, err := run.Command(ctx, ".", argv...)
	if err != nil {
		return nil, err
	}
	if res.Code != 0 {
		return nil, fmt.Errorf("gh api %s %s: %s", r.Method, r.Path, strings.TrimSpace(string(res.Stderr)+string(res.Stdout)))
	}
	return res.Stdout, nil
}

// Protect makes the kit the arbiter of repo ("owner/name"): the default
// branch's ruleset and the tags', each created or brought to the kit's;
// merged branches deleted and squash merges allowed; releases immutable.
// It returns what it did, a line each.
func Protect(ctx context.Context, api API, repo string) ([]string, error) {
	existing, err := rulesets(ctx, api, repo)
	if err != nil {
		return nil, err
	}
	var done []string
	for _, name := range []string{"ruleset.json", "ruleset-tags.json"} {
		line, err := applyRuleset(ctx, api, rulesetCall{repo, name, existing})
		if err != nil {
			return done, err
		}
		done = append(done, line)
	}
	settings := []byte(`{"delete_branch_on_merge": true, "allow_squash_merge": true}`)
	if _, err := api.Call(ctx, Request{"PATCH", "repos/" + repo, settings}); err != nil {
		return done, err
	}
	done = append(done, "merged branches deleted, squash merges allowed")
	if _, err := api.Call(ctx, Request{"PUT", "repos/" + repo + "/immutable-releases", nil}); err != nil {
		return done, err
	}
	return append(done, "releases immutable"), nil
}

func rulesets(ctx context.Context, api API, repo string) (map[string]int64, error) {
	out, err := api.Call(ctx, Request{"GET", "repos/" + repo + "/rulesets", nil})
	if err != nil {
		return nil, err
	}
	var list []struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	}
	if err := json.Unmarshal(out, &list); err != nil {
		return nil, fmt.Errorf("kit init: the rulesets of %s: %w", repo, err)
	}
	byName := map[string]int64{}
	for _, r := range list {
		byName[r.Name] = r.ID
	}
	return byName, nil
}

type rulesetCall struct {
	repo     string
	file     string
	existing map[string]int64
}

func applyRuleset(ctx context.Context, api API, c rulesetCall) (string, error) {
	body, err := files.ReadFile("files/" + c.file)
	if err != nil {
		return "", fmt.Errorf("kit init: %w", err)
	}
	var r struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(body, &r); err != nil {
		return "", fmt.Errorf("kit init: %s: %w", c.file, err)
	}
	req := Request{"POST", "repos/" + c.repo + "/rulesets", body}
	verb := "created"
	if id, ok := c.existing[r.Name]; ok {
		req = Request{"PUT", fmt.Sprintf("repos/%s/rulesets/%d", c.repo, id), body}
		verb = "brought to the kit's"
	}
	if _, err := api.Call(ctx, req); err != nil {
		return "", err
	}
	return fmt.Sprintf("ruleset %q %s", r.Name, verb), nil
}

// MergeApp is the GitHub App the kit's merge job merges as: its Client ID,
// its private key, and its installation on the repository, as GitHub
// reports it.
type MergeApp struct {
	Repo         string
	ClientID     string
	Key          []byte
	Installation ghapp.Installation
}

// GiveApp makes app the one repo's merge job merges as: its Client ID and
// key become the secrets KIT_APP_CLIENT_ID and KIT_APP_KEY, for Actions and
// for Dependabot, whose pull requests read only the latter. It refuses an
// App that may not write the repository's contents and pull requests.
func GiveApp(ctx context.Context, owner Owner, app MergeApp) ([]string, error) {
	var missing []string
	for _, p := range []string{"contents", "pull_requests"} {
		if app.Installation.Permissions[p] != "write" {
			missing = append(missing, p)
		}
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("kit init: the App may not write %s on %s: give it that in its settings, then accept it on its installation", strings.Join(missing, " and "), app.Repo)
	}
	for _, s := range []Secret{
		{app.Repo, "KIT_APP_CLIENT_ID", "actions", []byte(app.ClientID)},
		{app.Repo, "KIT_APP_KEY", "actions", app.Key},
		{app.Repo, "KIT_APP_CLIENT_ID", "dependabot", []byte(app.ClientID)},
		{app.Repo, "KIT_APP_KEY", "dependabot", app.Key},
	} {
		if err := owner.SetSecret(ctx, s); err != nil {
			return nil, err
		}
	}
	return []string{"the merge job merges as the App (the secrets KIT_APP_CLIENT_ID and KIT_APP_KEY, for Actions and Dependabot)"}, nil
}
