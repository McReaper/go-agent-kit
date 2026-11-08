package features_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/cucumber/godog"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// WHY: the spec is what the owner approves, so it runs on the program
// people run, built from cmd/notes, not on its code called in memory. A
// scenario tagged @wip is approved but not built yet: it waits.
func TestTheProgramDoesWhatTheSpecSays(t *testing.T) {
	bin := build(t)
	suite := godog.TestSuite{
		ScenarioInitializer: func(sc *godog.ScenarioContext) { steps(sc, bin) },
		Options:             &godog.Options{Format: "pretty", Paths: []string{"."}, Tags: "~@wip", Strict: true, TestingT: t},
	}
	if suite.Run() != 0 {
		t.Fatal("the program doesn't do what the spec says")
	}
}

type world struct {
	bin     string
	dataDir string
	last    result
	agent   *mcp.ClientSession
}

type result struct {
	code   int
	stdout string
	stderr string
}

func steps(sc *godog.ScenarioContext, bin string) {
	w := &world{bin: bin}
	sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		dir, err := os.MkdirTemp("", "notes-spec-")
		w.dataDir = dir
		return ctx, err
	})
	// WHY: godog hands the hook the failed step's error and reports it
	// already; the hook returns only its own, or the step's error is said
	// twice, as the hook's failure.
	sc.After(func(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
		var err error
		if w.agent != nil {
			err = w.agent.Close()
		}
		return ctx, errors.Join(err, os.RemoveAll(w.dataDir))
	})
	sc.Step(`^I add the note "([^"]*)" with the text "([^"]*)"$`, w.addNote)
	sc.Step(`^I add a note whose title is (\d+) characters long$`, w.addLongNote)
	sc.Step(`^I run "([^"]*)"$`, w.runLine)
	sc.Step(`^the list shows the note "([^"]*)"$`, w.listShows)
	sc.Step(`^it is refused with "([^"]*)"$`, w.refusedWith)
	sc.Step(`^it exits with code (\d+)$`, w.exitsWith)
	sc.Step(`^an agent adds the note "([^"]*)"$`, w.agentAdds)
	sc.Step(`^the agent's list shows the note "([^"]*)"$`, w.agentListShows)
}

func (w *world) addNote(ctx context.Context, title, body string) error {
	return w.run(ctx, "add-note", "--title="+title, "--body="+body)
}

func (w *world) addLongNote(ctx context.Context, n int) error {
	return w.run(ctx, "add-note", "--title="+strings.Repeat("x", n))
}

func (w *world) runLine(ctx context.Context, line string) error {
	return w.run(ctx, strings.Fields(line)...)
}

func (w *world) listShows(ctx context.Context, title string) error {
	if err := w.run(ctx, "list-notes"); err != nil {
		return err
	}
	return listHas([]byte(w.last.stdout), title)
}

func (w *world) refusedWith(message string) error {
	if w.last.code != 1 || !strings.Contains(w.last.stderr, message) {
		return fmt.Errorf("exit %d, stderr %q: want a refusal saying %q", w.last.code, w.last.stderr, message)
	}
	return nil
}

func (w *world) exitsWith(code int) error {
	if w.last.code != code {
		return fmt.Errorf("exit %d, want %d", w.last.code, code)
	}
	return nil
}

func (w *world) agentAdds(ctx context.Context, title string) error {
	c := mcp.NewClient(&mcp.Implementation{Name: "spec", Version: "0"}, nil)
	// WHY: the scenario's context ends before After closes the session, and
	// would kill the server before Close asks it to stop.
	s, err := c.Connect(ctx, &mcp.CommandTransport{Command: w.command(context.WithoutCancel(ctx), "mcp")}, nil)
	if err != nil {
		return err
	}
	w.agent = s
	_, err = w.call(ctx, "add_note", map[string]any{"title": title})
	return err
}

func (w *world) agentListShows(ctx context.Context, title string) error {
	data, err := w.call(ctx, "list_notes", map[string]any{})
	if err != nil {
		return err
	}
	return listHas(data, title)
}

func (w *world) call(ctx context.Context, tool string, args map[string]any) ([]byte, error) {
	res, err := w.agent.CallTool(ctx, &mcp.CallToolParams{Name: tool, Arguments: args})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", tool, err)
	}
	if res.IsError {
		return nil, fmt.Errorf("%s refused: %+v", tool, res.Content)
	}
	return json.Marshal(res.StructuredContent)
}

func listHas(data []byte, title string) error {
	var got struct {
		Notes []struct {
			Title string `json:"title"`
		} `json:"notes"`
	}
	if err := json.Unmarshal(data, &got); err != nil {
		return fmt.Errorf("the list %q: %w", data, err)
	}
	var titles []string
	for _, n := range got.Notes {
		titles = append(titles, n.Title)
	}
	if !slices.Contains(titles, title) {
		return fmt.Errorf("the list shows %q, not %q", titles, title)
	}
	return nil
}

func (w *world) command(ctx context.Context, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, w.bin, args...)
	cmd.Env = append(os.Environ(), "NOTES_DATA_DIR="+w.dataDir)
	return cmd
}

func (w *world) run(ctx context.Context, args ...string) error {
	var stdout, stderr bytes.Buffer
	cmd := w.command(ctx, args...)
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	var exit *exec.ExitError
	if err := cmd.Run(); err != nil && !errors.As(err, &exit) {
		return fmt.Errorf("notes %v: %w", args, err)
	}
	w.last = result{code: cmd.ProcessState.ExitCode(), stdout: stdout.String(), stderr: stderr.String()}
	return nil
}

func build(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "notes")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	if out, err := exec.CommandContext(t.Context(), "go", "build", "-o", bin, "../cmd/notes").CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}
	return bin
}
