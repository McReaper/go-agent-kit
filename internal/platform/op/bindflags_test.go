package op_test

import (
	"bytes"
	"context"
	"testing"

	"example.com/notes/internal/platform/op"
	"github.com/google/go-cmp/cmp"
)

type listArgs struct {
	Tags []string `json:"tags,omitempty"`
}

func TestCommandRepeatedFlagsMakeAList(t *testing.T) {
	var out bytes.Buffer
	o := op.Op[listArgs, listArgs]{
		Name:        "tag_args",
		Description: "returns its arguments",
		Effect:      op.Additive,
		Run:         func(_ context.Context, a listArgs) (listArgs, error) { return a, nil },
	}
	cmd, err := op.Command(o, &out)
	if err != nil {
		t.Fatal(err)
	}
	cmd.SetArgs([]string{"--tags", "a", "--tags", "b"})
	if err := cmd.ExecuteContext(context.Background()); err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(`{"tags":["a","b"]}`+"\n", out.String()); diff != "" {
		t.Errorf("printed (-want +got):\n%s", diff)
	}
}
