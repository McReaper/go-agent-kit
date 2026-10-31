package op_test

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/McReaper/go-agent-kit/template/internal/platform/op"
	"github.com/google/go-cmp/cmp"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/pflag"
)

type args struct {
	Name    string   `json:"name"`
	MaxSize int      `json:"max_size,omitempty"`
	DryRun  bool     `json:"dry_run"`
	Tag     *string  `json:"tag,omitempty"`
	Limit   *int     `json:"limit,omitempty"`
	Labels  []string `json:"labels,omitempty"`
}

type entry struct {
	op      string
	args    any
	failure error
}

type memJournal struct {
	entries []entry
	err     error
}

func (j *memJournal) Record(op string, args any, failure error) error {
	j.entries = append(j.entries, entry{op, args, failure})
	return j.err
}

var errBoom = errors.New("boom")

func echo(effect op.Effect) op.Op[args, args] {
	return op.Op[args, args]{
		Name:        "echo_args",
		Description: "returns its arguments",
		Effect:      effect,
		Run: func(_ context.Context, a args) (args, error) {
			if a.Name == "fail" {
				return a, errBoom
			}
			return a, nil
		},
	}
}

func runCommand(t *testing.T, o op.Op[args, args], j op.Journal, flags ...string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	cmd, err := op.Command(o, j, &out)
	if err != nil {
		t.Fatal(err)
	}
	cmd.SetArgs(flags)
	cmd.SilenceErrors, cmd.SilenceUsage = true, true
	err = cmd.ExecuteContext(context.Background())
	return strings.TrimSpace(out.String()), err
}

var sameEntries = []cmp.Option{cmp.AllowUnexported(entry{}), cmp.Comparer(func(a, b error) bool { return errors.Is(a, b) || errors.Is(b, a) })}

func TestCommandFlagsAreTheJSONNamesWithHyphens(t *testing.T) {
	cmd, err := op.Command(echo(op.ReadOnly), &memJournal{}, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	cmd.Flags().VisitAll(func(f *pflag.Flag) { got = append(got, f.Name+" "+f.Value.Type()) })
	want := []string{"dry-run bool", "labels stringSlice", "limit int", "max-size int", "name string", "tag string"}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("flags (-want +got):\n%s", diff)
	}
	if cmd.Use != "echo-args" {
		t.Errorf("Use = %q", cmd.Use)
	}
}

func TestCommandRunsTheOperationWithItsFlags(t *testing.T) {
	got, err := runCommand(t, echo(op.ReadOnly), &memJournal{}, "--name", "x", "--max-size", "7", "--dry-run", "--tag", "", "--limit", "0")
	if err != nil {
		t.Fatal(err)
	}
	if got != `{"name":"x","max_size":7,"dry_run":true,"tag":"","limit":0}` {
		t.Errorf("printed %s", got)
	}
}

func TestAListFlagTakesCommasAndRepeats(t *testing.T) {
	got, err := runCommand(t, echo(op.ReadOnly), &memJournal{}, "--name", "x", "--labels", "go,cli", "--labels", "mcp")
	if err != nil {
		t.Fatal(err)
	}
	if got != `{"name":"x","dry_run":false,"labels":["go","cli","mcp"]}` {
		t.Errorf("printed %s", got)
	}
}

// WHY: an optional argument tells "not given" from its zero value; a flag
// left out stays nil, so the operation applies its default.
func TestAnOptionalFlagLeftOutIsNil(t *testing.T) {
	got, err := runCommand(t, echo(op.ReadOnly), &memJournal{}, "--name", "x")
	if err != nil {
		t.Fatal(err)
	}
	if got != `{"name":"x","dry_run":false}` {
		t.Errorf("printed %s", got)
	}
	if _, err := runCommand(t, echo(op.ReadOnly), &memJournal{}, "--limit", "many"); err == nil {
		t.Error("--limit many accepted")
	}
}

func TestAnOperationThatChangesSomethingIsJournaled(t *testing.T) {
	j := &memJournal{}
	if _, err := runCommand(t, echo(op.Additive), j, "--name", "x"); err != nil {
		t.Fatal(err)
	}
	if _, err := runCommand(t, echo(op.Destructive), j, "--name", "fail"); !errors.Is(err, errBoom) {
		t.Fatalf("err = %v, want %v", err, errBoom)
	}
	if _, err := runCommand(t, echo(op.ReadOnly), j, "--name", "y"); err != nil {
		t.Fatal(err)
	}
	want := []entry{{"echo_args", args{Name: "x"}, nil}, {"echo_args", args{Name: "fail"}, errBoom}}
	if diff := cmp.Diff(want, j.entries, sameEntries...); diff != "" {
		t.Errorf("journal (-want +got):\n%s", diff)
	}
}

func TestARunWhoseJournalFailsReportsIt(t *testing.T) {
	_, err := runCommand(t, echo(op.Additive), &memJournal{err: errors.New("disk full")}, "--name", "x")
	if err == nil || !strings.Contains(err.Error(), "echo_args ran, but its journal entry failed: disk full") {
		t.Errorf("err = %v", err)
	}
}

func TestRegisterRefusesAnOperationWithoutAnEffect(t *testing.T) {
	s := mcp.NewServer(&mcp.Implementation{Name: "t", Version: "0"}, nil)
	err := op.Register(s, &memJournal{}, echo(0))
	if err == nil || !strings.Contains(err.Error(), "echo_args") {
		t.Fatalf("err = %v, want a refusal naming the tool", err)
	}
	if _, err := op.Command(echo(0), &memJournal{}, &bytes.Buffer{}); err == nil {
		t.Error("Command accepted an operation without an effect")
	}
}

func TestAnOperationBoundWithoutAJournalIsRefused(t *testing.T) {
	s := mcp.NewServer(&mcp.Implementation{Name: "t", Version: "0"}, nil)
	if err := op.Register(s, nil, echo(op.ReadOnly)); !errors.Is(err, op.ErrNoJournal) {
		t.Errorf("Register: err = %v, want ErrNoJournal", err)
	}
	if _, err := op.Command(echo(op.ReadOnly), nil, &bytes.Buffer{}); !errors.Is(err, op.ErrNoJournal) {
		t.Errorf("Command: err = %v, want ErrNoJournal", err)
	}
	if err := op.Register(s, nil, echo(0)); err == nil || !strings.Contains(err.Error(), "effect 0") {
		t.Errorf("Register without an effect or a journal: err = %v, want the effect named first", err)
	}
	if _, err := op.Command(echo(0), nil, &bytes.Buffer{}); err == nil || !strings.Contains(err.Error(), "effect 0") {
		t.Errorf("Command without an effect or a journal: err = %v, want the effect named first", err)
	}
}

type unnamed struct {
	Name string
}

type skipped struct {
	Name string `json:"-"`
}

type unsupported struct {
	Ratio float64 `json:"ratio"`
}

func refusal[A any](t *testing.T) error {
	t.Helper()
	o := op.Op[A, A]{Name: "x", Description: "x", Effect: op.ReadOnly, Run: func(_ context.Context, a A) (A, error) { return a, nil }}
	_, err := op.Command(o, &memJournal{}, &bytes.Buffer{})
	return err
}

// WHY: every field of Args is an MCP argument and a flag; one the binding
// can't name or type would exist on one surface only.
func TestAFieldWithoutAFlagIsRefused(t *testing.T) {
	cases := []struct {
		err  error
		want string
	}{
		{refusal[unnamed](t), "field Name has no json name"},
		{refusal[skipped](t), "field Name has no json name"},
		{refusal[unsupported](t), "field Ratio has type float64, which has no flag"},
	}
	for _, c := range cases {
		if c.err == nil || !strings.Contains(c.err.Error(), c.want) {
			t.Errorf("err = %v, want %q", c.err, c.want)
		}
	}
}

func TestTheHelpShowsEveryFlagOptionalOnesWithoutADefault(t *testing.T) {
	cmd, err := op.Command(echo(op.ReadOnly), &memJournal{}, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	usage := cmd.Flags().FlagUsages()
	for _, want := range []string{"--limit int", "--tag string", "--max-size int"} {
		if !strings.Contains(usage, want) {
			t.Errorf("usage lacks %q:\n%s", want, usage)
		}
	}
}

func TestAToolRunIsJournaledLikeACommand(t *testing.T) {
	j := &memJournal{}
	s := mcp.NewServer(&mcp.Implementation{Name: "t", Version: "0"}, nil)
	if err := op.Register(s, j, echo(op.Additive)); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	st, ct := mcp.NewInMemoryTransports()
	ss, err := s.Connect(ctx, st, nil)
	if err != nil {
		t.Fatal(err)
	}
	cs, err := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "0"}, nil).Connect(ctx, ct, nil)
	if err != nil {
		t.Fatal(err)
	}
	res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "echo_args", Arguments: map[string]any{"name": "x", "dry_run": true}})
	if err != nil {
		t.Fatal(err)
	}
	if err := errors.Join(cs.Close(), ss.Wait()); err != nil || res.IsError {
		t.Fatalf("echo_args = %+v, %v", res, err)
	}
	if diff := cmp.Diff([]entry{{"echo_args", args{Name: "x", DryRun: true}, nil}}, j.entries, sameEntries...); diff != "" {
		t.Errorf("journal (-want +got):\n%s", diff)
	}
}
