package setup

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/McReaper/rumblestrip-go/internal/platform/ghapi"
	"github.com/McReaper/rumblestrip-go/internal/platform/ghapp"
	"github.com/McReaper/rumblestrip-go/internal/platform/run"
)

// Request is one call to GitHub's REST API.
type Request = ghapi.Request

// API calls GitHub's REST API, here as the owner.
type API = ghapi.API

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

// GH calls GitHub's API through gh, as whoever gh is logged in as, and sets
// a repository's secrets.
type GH struct {
	ghapi.GH
}

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

// Protect makes Rumblestrip the arbiter of repo ("owner/name"): the default
// branch's ruleset and the tags', each created, or updated to Rumblestrip's;
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
	if _, err := api.Call(ctx, Request{Method: "PATCH", Path: "repos/" + repo, Body: settings}); err != nil {
		return done, err
	}
	done = append(done, "merged branches deleted, squash merges allowed")
	if _, err := api.Call(ctx, Request{Method: "PUT", Path: "repos/" + repo + "/immutable-releases"}); err != nil {
		return done, err
	}
	return append(done, "releases immutable"), nil
}

func rulesets(ctx context.Context, api API, repo string) (map[string]int64, error) {
	out, err := api.Call(ctx, Request{Method: "GET", Path: "repos/" + repo + "/rulesets"})
	if err != nil {
		return nil, err
	}
	var list []struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	}
	if err := json.Unmarshal(out, &list); err != nil {
		return nil, fmt.Errorf("rumblestrip init: the rulesets of %s: %w", repo, err)
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
		return "", fmt.Errorf("rumblestrip init: %w", err)
	}
	var r struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(body, &r); err != nil {
		return "", fmt.Errorf("rumblestrip init: %s: %w", c.file, err)
	}
	req := Request{Method: "POST", Path: "repos/" + c.repo + "/rulesets", Body: body}
	verb := "created"
	if id, ok := c.existing[r.Name]; ok {
		req = Request{Method: "PUT", Path: fmt.Sprintf("repos/%s/rulesets/%d", c.repo, id), Body: body}
		verb = "updated to Rumblestrip's"
	}
	if _, err := api.Call(ctx, req); err != nil {
		return "", err
	}
	return fmt.Sprintf("ruleset %q %s", r.Name, verb), nil
}

// MergeApp is the GitHub App Rumblestrip's merge job merges as: its Client ID,
// its private key, and its installation on the repository, as GitHub
// reports it.
type MergeApp struct {
	Repo         string
	ClientID     string
	Key          []byte
	Installation ghapp.Installation
}

// GiveApp makes app the one repo's merge job merges as: its Client ID and
// key become the secrets RUMBLESTRIP_APP_CLIENT_ID and RUMBLESTRIP_APP_KEY, for Actions and
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
		return nil, fmt.Errorf("rumblestrip init: the App may not write %s on %s: give it that in its settings, then accept it on its installation", strings.Join(missing, " and "), app.Repo)
	}
	for _, s := range []Secret{
		{app.Repo, "RUMBLESTRIP_APP_CLIENT_ID", "actions", []byte(app.ClientID)},
		{app.Repo, "RUMBLESTRIP_APP_KEY", "actions", app.Key},
		{app.Repo, "RUMBLESTRIP_APP_CLIENT_ID", "dependabot", []byte(app.ClientID)},
		{app.Repo, "RUMBLESTRIP_APP_KEY", "dependabot", app.Key},
	} {
		if err := owner.SetSecret(ctx, s); err != nil {
			return nil, err
		}
	}
	return []string{"the merge job merges as the App (the secrets RUMBLESTRIP_APP_CLIENT_ID and RUMBLESTRIP_APP_KEY, for Actions and Dependabot)"}, nil
}
