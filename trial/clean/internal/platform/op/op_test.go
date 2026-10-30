package op_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"example.com/notes/internal/platform/op"
	"github.com/google/go-cmp/cmp"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type args struct {
	Name    string `json:"name"`
	MaxSize int    `json:"max_size,omitempty"`
	DryRun  bool   `json:"dry_run"`
}

func echo(effect op.Effect) op.Op[args, args] {
	return op.Op[args, args]{
		Name:        "echo_args",
		Description: "returns its arguments",
		Effect:      effect,
		Run:         func(_ context.Context, a args) (args, error) { return a, nil },
	}
}

func TestCommandFlagsAreTheJSONNamesWithHyphens(t *testing.T) {
	want := []string{"name", "max-size", "dry-run"}
	if diff := cmp.Diff(want, op.FlagNames[args]()); diff != "" {
		t.Errorf("FlagNames (-want +got):\n%s", diff)
	}
}

func TestCommandRunsTheOperationWithItsFlags(t *testing.T) {
	var out bytes.Buffer
	cmd, err := op.Command(echo(op.ReadOnly), &out)
	if err != nil {
		t.Fatal(err)
	}
	cmd.SetArgs([]string{"--name", "x", "--max-size", "7", "--dry-run"})
	if err := cmd.ExecuteContext(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(out.String()); got != `{"name":"x","max_size":7,"dry_run":true}` {
		t.Errorf("printed %s", got)
	}
	if cmd.Use != "echo-args" {
		t.Errorf("Use = %q", cmd.Use)
	}
}

func TestRegisterRefusesAnOperationWithoutAnEffect(t *testing.T) {
	s := mcp.NewServer(&mcp.Implementation{Name: "t", Version: "0"}, nil)
	err := op.Register(s, echo(0))
	if err == nil || !strings.Contains(err.Error(), "echo_args") {
		t.Fatalf("err = %v, want a refusal naming the tool", err)
	}
}
