package kitinit_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/McReaper/go-agent-kit/kitinit"
)

const commit = "0123456789abcdef0123456789abcdef01234567"

func files(t *testing.T) map[string]string {
	t.Helper()
	fs, err := kitinit.Files(kitinit.Options{Version: kitinit.Version{Tag: "v0.2.0", Commit: commit}, Owner: "octo"})
	if err != nil {
		t.Fatal(err)
	}
	out := map[string]string{}
	for _, f := range fs {
		out[f.Path] = string(f.Content)
	}
	return out
}

func TestTheFilesPinTheKitNameTheOwnerAndTurnThePluginOn(t *testing.T) {
	got := files(t)
	if want := "uses: McReaper/go-agent-kit/.github/workflows/kit-check.yml@" + commit + " # v0.2.0\n"; !strings.Contains(got[".github/workflows/kit.yml"], want) {
		t.Errorf("kit.yml lacks %q:\n%s", want, got[".github/workflows/kit.yml"])
	}
	for _, line := range []string{"/.github/ @octo\n", "/AGENTS.md @octo\n", "**/GLOSSARY.md @octo\n"} {
		if !strings.Contains(got[".github/CODEOWNERS"], line) {
			t.Errorf("CODEOWNERS lacks %q:\n%s", line, got[".github/CODEOWNERS"])
		}
	}
	var settings struct {
		EnabledPlugins map[string]bool `json:"enabledPlugins"`
		Permissions    struct {
			Deny []string `json:"deny"`
		} `json:"permissions"`
	}
	if err := json.Unmarshal([]byte(got[".claude/settings.json"]), &settings); err != nil || !settings.EnabledPlugins["kit@go-agent-kit"] || !slices.Contains(settings.Permissions.Deny, "Bash(gh pr merge *)") {
		t.Errorf("settings.json: %+v, %v", settings, err)
	}
	for _, eco := range []string{"package-ecosystem: github-actions", "package-ecosystem: gomod"} {
		if !strings.Contains(got[".github/dependabot.yml"], eco) {
			t.Errorf("dependabot.yml lacks %s", eco)
		}
	}
}

func TestTheFilesNeedTheOwnerAndTheKitsFullCommit(t *testing.T) {
	for _, o := range []kitinit.Options{
		{Version: kitinit.Version{Tag: "v0.2.0", Commit: commit}},
		{Version: kitinit.Version{Tag: "v0.2.0", Commit: "0123456"}, Owner: "octo"},
	} {
		if _, err := kitinit.Files(o); err == nil {
			t.Errorf("Files(%+v) wrote files", o)
		}
	}
}

func TestWriteKeepsAFileTheRepositoryHasAndWritesTheOthersOnce(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".github"), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".github", "CODEOWNERS"), []byte("* @octo @hubot\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	fs, err := kitinit.Files(kitinit.Options{Version: kitinit.Version{Tag: "v0.2.0", Commit: commit}, Owner: "octo"})
	if err != nil {
		t.Fatal(err)
	}
	first, err := kitinit.Write(dir, fs)
	if err != nil {
		t.Fatal(err)
	}
	want := kitinit.Report{Written: []string{".github/workflows/kit.yml", ".github/dependabot.yml", ".claude/settings.json"}, Kept: []string{".github/CODEOWNERS"}}
	if diff := cmp.Diff(want, first); diff != "" {
		t.Errorf("first run (-want +got):\n%s", diff)
	}
	again, err := kitinit.Write(dir, fs)
	if err != nil || len(again.Written) != 0 || len(again.Same) != 3 {
		t.Errorf("second run: %+v, %v", again, err)
	}
}

type fakeAPI struct {
	calls []string
	reply map[string]string
}

func (f *fakeAPI) Call(_ context.Context, r kitinit.Request) ([]byte, error) {
	f.calls = append(f.calls, r.Method+" "+r.Path)
	if r.Method == "PATCH" && !strings.Contains(string(r.Body), `"delete_branch_on_merge": true`) {
		return nil, errors.New("the deletion of merged branches not asked")
	}
	return []byte(f.reply[r.Method+" "+r.Path]), nil
}

func TestProtectBringsAnExistingRulesetToTheKitsAndCreatesTheOther(t *testing.T) {
	var tags struct {
		Name string `json:"name"`
	}
	data, err := os.ReadFile("files/ruleset-tags.json")
	if err != nil || json.Unmarshal(data, &tags) != nil {
		t.Fatalf("the tag ruleset: %v", err)
	}
	api := &fakeAPI{reply: map[string]string{"GET repos/o/r/rulesets": `[{"id": 7, "name": "` + tags.Name + `"}]`}}
	if _, err := kitinit.Protect(t.Context(), api, "o/r"); err != nil {
		t.Fatal(err)
	}
	want := []string{"GET repos/o/r/rulesets", "POST repos/o/r/rulesets", "PUT repos/o/r/rulesets/7", "PATCH repos/o/r", "PUT repos/o/r/immutable-releases"}
	if diff := cmp.Diff(want, api.calls); diff != "" {
		t.Errorf("calls (-want +got):\n%s", diff)
	}
}

type fakeOwner struct {
	fakeAPI
	secrets []kitinit.Secret
}

func (f *fakeOwner) SetSecret(_ context.Context, s kitinit.Secret) error {
	f.secrets = append(f.secrets, s)
	return nil
}

func TestGiveAppStoresTheAppsClientIDAndKeyForActionsAndDependabot(t *testing.T) {
	owner := &fakeOwner{}
	app := kitinit.MergeApp{Repo: "o/r", ClientID: "Iv1", Key: []byte("the key")}
	app.Installation.Permissions = map[string]string{"contents": "write", "pull_requests": "write", "metadata": "read"}
	if _, err := kitinit.GiveApp(t.Context(), owner, app); err != nil {
		t.Fatal(err)
	}
	want := []kitinit.Secret{
		{Repo: "o/r", Name: "KIT_APP_CLIENT_ID", For: "actions", Value: []byte("Iv1")},
		{Repo: "o/r", Name: "KIT_APP_KEY", For: "actions", Value: []byte("the key")},
		{Repo: "o/r", Name: "KIT_APP_CLIENT_ID", For: "dependabot", Value: []byte("Iv1")},
		{Repo: "o/r", Name: "KIT_APP_KEY", For: "dependabot", Value: []byte("the key")},
	}
	if diff := cmp.Diff(want, owner.secrets); diff != "" {
		t.Errorf("secrets (-want +got):\n%s", diff)
	}
}

func TestGiveAppRefusesAnAppThatMayNotMergeAndSetsNoSecret(t *testing.T) {
	for _, c := range []struct {
		name  string
		perms map[string]string
		says  string
	}{
		{"reads the contents", map[string]string{"contents": "read", "pull_requests": "write"}, "may not write contents on o/r"},
		{"no pull requests", map[string]string{"contents": "write"}, "may not write pull_requests on o/r"},
	} {
		t.Run(c.name, func(t *testing.T) {
			owner := &fakeOwner{}
			app := kitinit.MergeApp{Repo: "o/r", ClientID: "Iv1", Key: []byte("the key")}
			app.Installation.Permissions = c.perms
			_, err := kitinit.GiveApp(t.Context(), owner, app)
			if err == nil || !strings.Contains(err.Error(), c.says) {
				t.Errorf("err = %v, want one saying %q", err, c.says)
			}
			if len(owner.secrets) != 0 {
				t.Errorf("secrets set: %d, want none", len(owner.secrets))
			}
		})
	}
}

func TestTheAppsManifestAsksNoWorkflowsAndNoAdministration(t *testing.T) {
	m, err := kitinit.Manifest("notes-agent", "http://127.0.0.1:9/created")
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		Permissions map[string]string `json:"default_permissions"`
		Redirect    string            `json:"redirect_url"`
	}
	if err := json.Unmarshal(m, &got); err != nil {
		t.Fatal(err)
	}
	_, workflows := got.Permissions["workflows"]
	_, admin := got.Permissions["administration"]
	if workflows || admin || got.Permissions["contents"] != "write" || got.Redirect != "http://127.0.0.1:9/created" {
		t.Errorf("manifest: %s", m)
	}
}

func TestTheFormPostsTheManifestToGitHubsNewAppPage(t *testing.T) {
	var b strings.Builder
	if err := kitinit.Form(&b, kitinit.NewAppURL("acme"), []byte(`{"name":"a<b"}`)); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`action="https://github.com/organizations/acme/settings/apps/new"`, `value="{&#34;name&#34;:&#34;a&lt;b&#34;}"`} {
		if !strings.Contains(b.String(), want) {
			t.Errorf("the form lacks %s:\n%s", want, b.String())
		}
	}
}

func TestExchangeTradesTheCodeForTheAppAndItsKey(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/app-manifests/abc/conversions" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusCreated)
		if _, err := w.Write([]byte(`{"id": 42, "slug": "notes-agent", "client_id": "Iv1", "pem": "KEY"}`)); err != nil {
			t.Error(err)
		}
	}))
	defer srv.Close()
	ex := kitinit.Exchanger{Client: srv.Client(), BaseURL: srv.URL}
	app, err := ex.Exchange(t.Context(), "abc")
	if err != nil || app != (kitinit.App{ID: 42, Slug: "notes-agent", ClientID: "Iv1", PEM: "KEY"}) {
		t.Errorf("Exchange = %+v, %v", app, err)
	}
	if _, err := ex.Exchange(t.Context(), "expired"); err == nil || !strings.Contains(err.Error(), "404") {
		t.Errorf("an expired code: %v", err)
	}
}

func TestTheWorkflowForAPinIsTheOneKitInitWritesForIt(t *testing.T) {
	written := []byte(files(t)[".github/workflows/kit.yml"])
	if got, err := kitinit.WorkflowFor(written); err != nil || string(got) != string(written) {
		t.Errorf("WorkflowFor(what kit init wrote) = %s, %v", got, err)
	}
	trimmed := []byte(strings.Replace(string(written), "  pull_request_review:\n    types: [submitted]\n", "", 1))
	if got, err := kitinit.WorkflowFor(trimmed); err != nil || string(got) == string(trimmed) {
		t.Errorf("a kit.yml without the review trigger passes: %v", err)
	}
	if _, err := kitinit.WorkflowFor([]byte("name: kit\n")); err == nil {
		t.Error("a kit.yml with no pin passes")
	}
}
