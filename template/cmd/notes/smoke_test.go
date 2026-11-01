package main_test

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// WHY: the gates check the code, not the program people run: a crash test
// shipped a binary whose every command failed at startup while each test and
// gate passed. This test builds it and calls every tool and every command
// with the example its schema shows. A refusal is an answer (exit 1, a tool
// error); a start that fails, a command line that doesn't parse (exit 2) or
// a protocol error is not. A new operation is covered the day it's
// registered.
func TestTheBuiltProgramAnswersEveryToolAndCommandWithItsExample(t *testing.T) {
	bin := build(t)
	if r := run(t, bin, t.TempDir(), "--help"); r.code != 0 {
		t.Fatalf("notes --help exited %d:\n%s", r.code, r.stderr)
	}
	s := connect(t, bin, t.TempDir())
	res, err := s.ListTools(t.Context(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Tools) == 0 {
		t.Fatal("the program serves no tool")
	}
	dataDir := t.TempDir()
	for _, tool := range res.Tools {
		example := exampleOf(t, tool)
		if _, err := s.CallTool(t.Context(), &mcp.CallToolParams{Name: tool.Name, Arguments: example}); err != nil {
			t.Errorf("tool %s with %v: %v", tool.Name, example, err)
		}
		args := append([]string{strings.ReplaceAll(tool.Name, "_", "-")}, flagsOf(t, example)...)
		if r := run(t, bin, dataDir, args...); r.code != 0 && r.code != 1 {
			t.Errorf("notes %s exited %d:\n%s", strings.Join(args, " "), r.code, r.stderr)
		}
	}
}

func exampleOf(t *testing.T, tool *mcp.Tool) map[string]any {
	t.Helper()
	data, err := json.Marshal(tool.InputSchema)
	if err != nil {
		t.Fatal(err)
	}
	var schema struct {
		Examples []map[string]any `json:"examples"`
	}
	if err := json.Unmarshal(data, &schema); err != nil {
		t.Fatal(err)
	}
	if len(schema.Examples) == 0 {
		t.Fatalf("tool %s shows no example in its schema", tool.Name)
	}
	return schema.Examples[0]
}

// WHY: the flags come from the tool's example by the binding's own naming
// rule, so both surfaces are called with the same arguments.
func flagsOf(t *testing.T, example map[string]any) []string {
	t.Helper()
	var flags []string
	for _, k := range slices.Sorted(maps.Keys(example)) {
		name := "--" + strings.ReplaceAll(k, "_", "-")
		switch v := example[k].(type) {
		case string:
			flags = append(flags, name+"="+v)
		case float64:
			flags = append(flags, name+"="+strconv.FormatFloat(v, 'f', -1, 64))
		case bool:
			flags = append(flags, name+"="+strconv.FormatBool(v))
		case []any:
			for _, e := range v {
				flags = append(flags, fmt.Sprintf("%s=%v", name, e))
			}
		default:
			t.Fatalf("argument %s: a %T has no flag", k, v)
		}
	}
	return flags
}
