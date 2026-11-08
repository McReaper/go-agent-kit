package run_test

import (
	"strings"
	"testing"

	"github.com/McReaper/go-agent-kit/internal/platform/run"
)

func TestACommandReadsWhatItIsGivenOnItsStandardInput(t *testing.T) {
	out, err := run.Run(t.Context(), run.Cmd{Dir: ".", Argv: []string{"cat"}, Stdin: strings.NewReader("the key")})
	if err != nil || out.Code != 0 || string(out.Stdout) != "the key" {
		t.Fatalf("cat = %q, exit %d, %v; want what it was given", out.Stdout, out.Code, err)
	}
}
