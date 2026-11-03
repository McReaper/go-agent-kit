package kitinit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Manifest is the GitHub App the agent acts as, in the form GitHub's
// manifest flow takes. Contents and pull requests are the agent's work;
// it reads checks, actions and statuses to follow the merge check. It
// gets no workflows and no administration permission: the CI and the
// protection rules stay the owner's.
func Manifest(name, redirect string) ([]byte, error) {
	m := map[string]any{
		"name":            name,
		"url":             "https://github.com/McReaper/go-agent-kit",
		"hook_attributes": map[string]any{"url": "https://github.com/McReaper/go-agent-kit", "active": false},
		"redirect_url":    redirect,
		"public":          false,
		"default_permissions": map[string]string{
			"contents": "write", "pull_requests": "write", "metadata": "read",
			"checks": "read", "actions": "read", "statuses": "read",
		},
		"default_events": []string{},
	}
	b, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("kit init app: %w", err)
	}
	return b, nil
}

// NewAppURL is the page that creates an App from a manifest: the owner's
// account's, or an organization's when org is set.
func NewAppURL(org string) string {
	if org != "" {
		return "https://github.com/organizations/" + url.PathEscape(org) + "/settings/apps/new"
	}
	return "https://github.com/settings/apps/new"
}

// WHY: GitHub takes the manifest as a form posted from the owner's browser;
// this page posts it as soon as it loads.
const formPage = `<!doctype html>
<meta charset="utf-8"><title>kit init app</title>
<form id="f" method="post" action="{{.Action}}"><input type="hidden" name="manifest" value="{{.Manifest}}">
<p>Creating the agent's GitHub App on GitHub…</p><button>Go on</button></form>
<script>document.getElementById("f").submit()</script>
`

// Form writes the page that posts the manifest to action.
func Form(w io.Writer, action string, manifest []byte) error {
	t, err := template.New("form").Parse(formPage)
	if err != nil {
		return fmt.Errorf("kit init app: %w", err)
	}
	return t.Execute(w, struct {
		Action   string
		Manifest string
	}{action, string(manifest)})
}

// App is the App GitHub created: what kit app exec needs to act as it.
type App struct {
	ID       int64  `json:"id"`
	Slug     string `json:"slug"`
	ClientID string `json:"client_id"`
	PEM      string `json:"pem"`
}

// Exchanger trades the code GitHub's redirect carries for the new App.
type Exchanger struct {
	Client  *http.Client
	BaseURL string // https://api.github.com
}

// Exchange returns the App the code names, its private key included.
func (e Exchanger) Exchange(ctx context.Context, code string) (a App, err error) {
	u := strings.TrimSuffix(e.BaseURL, "/") + "/app-manifests/" + url.PathEscape(code) + "/conversions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, nil)
	if err != nil {
		return App{}, fmt.Errorf("kit init app: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	res, err := e.Client.Do(req)
	if err != nil {
		return App{}, fmt.Errorf("kit init app: %w", err)
	}
	defer func() { err = errors.Join(err, res.Body.Close()) }()
	if res.StatusCode != http.StatusCreated {
		body, rerr := io.ReadAll(io.LimitReader(res.Body, 4096))
		return App{}, errors.Join(fmt.Errorf("kit init app: GitHub answered %s: %s", res.Status, strings.TrimSpace(string(body))), rerr)
	}
	if err := json.NewDecoder(res.Body).Decode(&a); err != nil {
		return App{}, fmt.Errorf("kit init app: GitHub's answer: %w", err)
	}
	if a.Slug == "" || a.PEM == "" {
		return App{}, fmt.Errorf("kit init app: GitHub's answer holds no app")
	}
	return a, nil
}
