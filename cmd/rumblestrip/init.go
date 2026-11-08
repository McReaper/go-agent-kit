package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/McReaper/rumblestrip-go/internal/platform/ghapp"
	"github.com/McReaper/rumblestrip-go/internal/platform/gomod"
	"github.com/McReaper/rumblestrip-go/setup"
	"github.com/spf13/cobra"
)

const kitModule = "github.com/McReaper/rumblestrip-go"

type initOptions struct {
	owner      string
	ownerEmail string
	version    string
}

func initCommand(out io.Writer) *cobra.Command {
	var o initOptions
	cmd := leaf("init", "Write the files a repository keeps for Rumblestrip: its CI, code owners, Dependabot and Claude Code settings", func(cmd *cobra.Command) error {
		return runInit(cmd.Context(), out, o)
	})
	cmd.Flags().StringVar(&o.owner, "owner", "", "the owner's GitHub login, the code owner of the harness")
	cmd.Flags().StringVar(&o.version, "version", "", "the Rumblestrip release to pin, a tag (default: this binary's own version)")
	cmd.Flags().StringVar(&o.ownerEmail, "owner-email", "", "the owner's address the agent's commits name them co-author by (default: their GitHub no-reply address)")
	cmd.AddCommand(initGitHubCommand(out), initAppCommand(out))
	return cmd
}

func runInit(ctx context.Context, out io.Writer, o initOptions) error {
	tag := o.version
	if tag == "" {
		tag = ownVersion()
	}
	if tag == "" {
		return errors.New("rumblestrip init: this rumblestrip was built from a checkout: name the release to pin with --version")
	}
	commit, err := gomod.Commit(ctx, kitModule, tag)
	if err != nil {
		return fmt.Errorf("rumblestrip init: %w", err)
	}
	if o.ownerEmail == "" && o.owner != "" {
		if o.ownerEmail, err = noReply(ctx, ghapp.NewClient(), o.owner); err != nil {
			return fmt.Errorf("rumblestrip init: %w", err)
		}
	}
	fs, err := setup.Files(setup.Options{Version: setup.Version{Tag: tag, Commit: commit}, Owner: o.owner, OwnerEmail: o.ownerEmail})
	if err != nil {
		return err
	}
	r, err := setup.Write(".", fs)
	if err != nil {
		return err
	}
	_, err = io.WriteString(out, initReport(r))
	return err
}

// WHY: GitHub credits a commit to the account its address belongs to; the
// no-reply one, id+login@users.noreply.github.com, always does, whatever
// the owner made public.
func noReply(ctx context.Context, c ghapp.Client, login string) (string, error) {
	id, err := c.UserID(ctx, login)
	if err != nil {
		return "", fmt.Errorf("the owner's GitHub id: %w", err)
	}
	return ghapp.Identity{Login: login, ID: id}.Email(), nil
}

// WHY: go install and go run record the version they built; a checkout's
// build says (devel), and names no release to pin.
func ownVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Path == kitModule && strings.HasPrefix(info.Main.Version, "v") {
		return info.Main.Version
	}
	return ""
}

var pseudoVersion = regexp.MustCompile(`^v\d+\.\d+\.\d+-(?:.*\.)?\d{14}-([0-9a-f]{7})[0-9a-f]{5}$`)

// WHY: a rumblestrip pinned at a commit no release tags installs as a pseudo-version
// (v0.2.8-0.20261108130046-03876d889681), which says no more to the owner
// than its commit does.
func shortVersion(v string) string {
	if m := pseudoVersion.FindStringSubmatch(v); m != nil {
		return m[1]
	}
	return v
}

func initReport(r setup.Report) string {
	var b strings.Builder
	for _, p := range r.Written {
		fmt.Fprintf(&b, "written: %s\n", p)
	}
	for _, p := range r.Same {
		fmt.Fprintf(&b, "already Rumblestrip's: %s\n", p)
	}
	for _, p := range r.Kept {
		fmt.Fprintf(&b, "kept, it differs from Rumblestrip's: %s (delete it and run rumblestrip init again to take Rumblestrip's)\n", p)
	}
	b.WriteString("\nThen, as the owner, once: rumblestrip init app if the agent has no GitHub App yet, and rumblestrip init github --repo OWNER/REPO with its RUMBLESTRIP_APP_CLIENT_ID and RUMBLESTRIP_APP_KEY (the rulesets, the repository's settings, the App Rumblestrip merges as).\n")
	return b.String()
}

type initGitHubOptions struct {
	repo     string
	clientID string
	key      string
}

func initGitHubCommand(out io.Writer) *cobra.Command {
	o := initGitHubOptions{clientID: os.Getenv("RUMBLESTRIP_APP_CLIENT_ID"), key: os.Getenv("RUMBLESTRIP_APP_KEY")}
	cmd := leaf("github", "As the owner: the rulesets that make Rumblestrip's checks the arbiter, merged branches deleted, immutable releases, the App Rumblestrip merges as", func(cmd *cobra.Command) error {
		return runInitGitHub(cmd.Context(), out, o)
	})
	cmd.Flags().StringVar(&o.repo, "repo", "", "the repository, OWNER/REPO")
	cmd.Flags().StringVar(&o.clientID, "client-id", o.clientID, "the Client ID of the GitHub App Rumblestrip merges as, a new one from rumblestrip init app or one you have (RUMBLESTRIP_APP_CLIENT_ID)")
	cmd.Flags().StringVar(&o.key, "key", o.key, "that App's private key file (RUMBLESTRIP_APP_KEY)")
	return cmd
}

// WHY: the App is checked before anything changes, so a repository is
// never left with Rumblestrip's rules and no App to merge as.
func runInitGitHub(ctx context.Context, out io.Writer, o initGitHubOptions) error {
	app, err := mergeApp(ctx, o)
	if err != nil {
		return err
	}
	lines, err := setup.Protect(ctx, setup.GH{}, o.repo)
	if err == nil {
		var more []string
		more, err = setup.GiveApp(ctx, setup.GH{}, app)
		lines = append(lines, more...)
	}
	for _, l := range lines {
		if _, werr := fmt.Fprintln(out, l); werr != nil {
			return errors.Join(err, werr)
		}
	}
	return err
}

func mergeApp(ctx context.Context, o initGitHubOptions) (setup.MergeApp, error) {
	if o.repo == "" {
		return setup.MergeApp{}, errors.New("rumblestrip init github: name the repository with --repo OWNER/REPO")
	}
	if o.clientID == "" || o.key == "" {
		return setup.MergeApp{}, errors.New("rumblestrip init github: Rumblestrip merges as a GitHub App: give its Client ID and key, --client-id and --key or RUMBLESTRIP_APP_CLIENT_ID and RUMBLESTRIP_APP_KEY; rumblestrip init app creates one")
	}
	jwt, err := (&appOptions{clientID: o.clientID, key: o.key}).jwt()
	if err != nil {
		return setup.MergeApp{}, fmt.Errorf("rumblestrip init github: %w", err)
	}
	inst, err := ghapp.NewClient().RepoInstallation(ctx, jwt, o.repo)
	if err != nil {
		return setup.MergeApp{}, fmt.Errorf("rumblestrip init github: is the App installed on %s? %w", o.repo, err)
	}
	key, err := ghapp.KeyFile(o.key)
	if err != nil {
		return setup.MergeApp{}, fmt.Errorf("rumblestrip init github: %w", err)
	}
	return setup.MergeApp{Repo: o.repo, ClientID: o.clientID, Key: key, Installation: inst}, nil
}

type initAppOptions struct {
	name string
	org  string
}

func initAppCommand(out io.Writer) *cobra.Command {
	var o initAppOptions
	cmd := leaf("app", "As the owner: create the agent's GitHub App from Rumblestrip's manifest, through your browser", func(cmd *cobra.Command) error {
		return runInitApp(cmd.Context(), out, o)
	})
	flags := cmd.Flags()
	flags.StringVar(&o.name, "name", "", "the App's name, unique on GitHub")
	flags.StringVar(&o.org, "org", "", "create it for this organization rather than your account")
	return cmd
}

// WHY: GitHub's manifest flow ends in a redirect to a page of the owner's
// machine carrying a one-time code; a local server serves the form that
// starts it and takes the code back, so the private key goes from GitHub
// to the owner's disk and nowhere else.
func runInitApp(ctx context.Context, out io.Writer, o initAppOptions) error {
	if o.name == "" {
		return errors.New("rumblestrip init app: name the App with --name")
	}
	l, err := (&net.ListenConfig{}).Listen(ctx, "tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("rumblestrip init app: %w", err)
	}
	base := "http://" + l.Addr().String()
	manifest, err := setup.Manifest(o.name, base+"/created")
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "Open %s in your browser, and create the App on GitHub.\n", base); err != nil {
		return errors.Join(err, l.Close())
	}
	app, err := serveUntilCreated(ctx, l, appHandler(manifest, o.org))
	if app.Slug == "" {
		return errors.Join(errors.New("rumblestrip init app: no App was created"), err)
	}
	return errors.Join(err, saveApp(out, app))
}

// WHY: the server lives until GitHub's redirect brings the App back, the
// owner gives up (the context), or a quarter of an hour passes.
func serveUntilCreated(ctx context.Context, l net.Listener, h func(chan<- setup.App) http.Handler) (setup.App, error) {
	created := make(chan setup.App, 1)
	srv := &http.Server{Handler: h(created), ReadHeaderTimeout: 10 * time.Second}
	var serveErr error
	var wg sync.WaitGroup
	wg.Go(func() {
		if err := srv.Serve(l); !errors.Is(err, http.ErrServerClosed) {
			serveErr = err
		}
	})
	var app setup.App
	select {
	case app = <-created:
	case <-ctx.Done():
	case <-time.After(15 * time.Minute):
	}
	err := srv.Shutdown(context.WithoutCancel(ctx))
	wg.Wait()
	return app, errors.Join(err, serveErr)
}

func appHandler(manifest []byte, org string) func(chan<- setup.App) http.Handler {
	return func(created chan<- setup.App) http.Handler {
		return appMux(manifest, org, created)
	}
}

func appMux(manifest []byte, org string, created chan<- setup.App) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, _ *http.Request) {
		if err := setup.Form(w, setup.NewAppURL(org), manifest); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
	mux.HandleFunc("GET /created", func(w http.ResponseWriter, r *http.Request) {
		ex := setup.Exchanger{Client: &http.Client{Timeout: 30 * time.Second}, BaseURL: "https://api.github.com"}
		app, err := ex.Exchange(r.Context(), r.URL.Query().Get("code"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		select {
		case created <- app:
		default:
		}
		if _, err := io.WriteString(w, "The App is created: go back to your terminal.\n"); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
	return mux
}

// WHY: the key goes beside the user's other configuration, readable by them
// alone, never into a repository.
func saveApp(out io.Writer, app setup.App) error {
	dir, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("rumblestrip init app: %w", err)
	}
	dir = filepath.Join(dir, "rumblestrip-go")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("rumblestrip init app: %w", err)
	}
	key := filepath.Join(dir, app.Slug+".private-key.pem")
	if err := os.WriteFile(key, []byte(app.PEM), 0o600); err != nil {
		return fmt.Errorf("rumblestrip init app: %w", err)
	}
	_, err = fmt.Fprintf(out, `The App %s is created (id %d). Its private key: %s

Install it on the repositories the agent works in:
  https://github.com/apps/%s/installations/new

The agent acts as it with rumblestrip app exec, and Rumblestrip merges as it once
rumblestrip init github --repo OWNER/REPO ran, given:
  RUMBLESTRIP_APP_CLIENT_ID=%s
  RUMBLESTRIP_APP_KEY=%s
  RUMBLESTRIP_APP_INSTALLATION=<the installation's id: rumblestrip app installations lists it>
`, app.Slug, app.ID, key, app.Slug, app.ClientID, key)
	return err
}
