package op_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/McReaper/go-agent-kit/template/internal/platform/op"
)

type listArgs struct {
	Tags []string `json:"tags,omitempty" jsonschema:"tags"`
}

func TestCommandSliceFlagRepeatsAndSplitsOnCommas(t *testing.T) {
	var out bytes.Buffer
	o := op.Op[listArgs, listArgs]{
		Name:        "echo_tags",
		Description: "returns its tags",
		Effect:      op.ReadOnly,
		Run:         func(_ context.Context, a listArgs) (listArgs, error) { return a, nil },
	}
	cmd, err := op.Command(o, &memJournal{}, &out)
	if err != nil {
		t.Fatal(err)
	}
	cmd.SetArgs([]string{"--tags", "a,b", "--tags", "c"})
	cmd.SilenceErrors, cmd.SilenceUsage = true, true
	if err := cmd.ExecuteContext(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(out.String()); got != `{"tags":["a","b","c"]}` {
		t.Errorf("printed %s", got)
	}
}
