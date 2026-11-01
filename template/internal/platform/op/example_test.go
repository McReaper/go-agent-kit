package op_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"github.com/McReaper/go-agent-kit/template/internal/platform/op"
	"github.com/google/go-cmp/cmp"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestAnOperationWithoutAnExampleIsRefused(t *testing.T) {
	o := echo(op.ReadOnly)
	o.Example = nil
	s := mcp.NewServer(&mcp.Implementation{Name: "t", Version: "0"}, nil)
	if err := op.Register(s, &memJournal{}, o); !errors.Is(err, op.ErrNoExample) {
		t.Errorf("Register: err = %v, want ErrNoExample", err)
	}
	if _, err := op.Command(o, &memJournal{}, &bytes.Buffer{}); !errors.Is(err, op.ErrNoExample) {
		t.Errorf("Command: err = %v, want ErrNoExample", err)
	}
}

func TestAToolShowsItsExampleInItsSchema(t *testing.T) {
	s := mcp.NewServer(&mcp.Implementation{Name: "t", Version: "0"}, nil)
	o := echo(op.ReadOnly)
	o.Example = &args{Name: "groceries", Labels: []string{"home"}}
	if err := op.Register(s, &memJournal{}, o); err != nil {
		t.Fatal(err)
	}
	ct, st := mcp.NewInMemoryTransports()
	if _, err := s.Connect(t.Context(), st, nil); err != nil {
		t.Fatal(err)
	}
	cs, err := mcp.NewClient(&mcp.Implementation{Name: "c", Version: "0"}, nil).Connect(t.Context(), ct, nil)
	if err != nil {
		t.Fatal(err)
	}
	res, err := cs.ListTools(t.Context(), nil)
	if err != nil || len(res.Tools) != 1 {
		t.Fatalf("tools = %v, %v", res, err)
	}
	data, err := json.Marshal(res.Tools[0].InputSchema)
	if err != nil {
		t.Fatal(err)
	}
	var schema struct {
		Examples []map[string]any `json:"examples"`
	}
	if err := json.Unmarshal(data, &schema); err != nil {
		t.Fatal(err)
	}
	want := []map[string]any{{"name": "groceries", "dry_run": false, "labels": []any{"home"}}}
	if diff := cmp.Diff(want, schema.Examples); diff != "" {
		t.Errorf("examples (-want +got):\n%s", diff)
	}
}

// WHY: the program exits 1 when an operation refused or failed and 2 when
// the command line didn't parse; it tells them apart by this type.
func TestAnOperationsErrorIsAFailureAndABadFlagIsNot(t *testing.T) {
	_, err := runCommand(t, echo(op.ReadOnly), &memJournal{}, "--name", "fail")
	f, ok := errors.AsType[*op.Failure](err)
	if !ok || f == nil {
		t.Fatalf("operation error: err = %v (%T), want a *Failure", err, err)
	}
	if f.Op != "echo_args" || !errors.Is(f, errBoom) || f.Error() != errBoom.Error() {
		t.Errorf("operation error: %+v, want echo_args's, wrapping boom with boom's message", *f)
	}
	_, err = runCommand(t, echo(op.ReadOnly), &memJournal{}, "--no-such-flag")
	if err == nil || errors.As(err, new(*op.Failure)) {
		t.Errorf("bad flag: err = %v (%T), want an error that isn't a *Failure", err, err)
	}
}
