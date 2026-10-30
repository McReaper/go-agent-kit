// Package op binds one operation to both surfaces, an MCP tool and a CLI
// command, from a single arguments struct.
package op

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
)

// Effect is what an operation does to its environment. It has no default:
// an operation that doesn't state its effect can't be registered.
type Effect int

const (
	// ReadOnly operations change nothing.
	ReadOnly Effect = iota + 1
	// Additive operations add and never overwrite or delete.
	Additive
	// Destructive operations may overwrite or delete.
	Destructive
)

// Op is one operation: its name, its description for agents, its effect and
// its implementation. A's json tags are the MCP arguments and the CLI flags.
type Op[A, O any] struct {
	Name        string
	Description string
	Effect      Effect
	Run         func(context.Context, A) (O, error)
}

// Register adds o to s as a tool, with annotations derived from its effect.
func Register[A, O any](s *mcp.Server, o Op[A, O]) error {
	ann, err := annotations(o.Effect)
	if err != nil {
		return fmt.Errorf("tool %s: %w", o.Name, err)
	}
	tool := &mcp.Tool{Name: o.Name, Description: o.Description, Annotations: ann}
	mcp.AddTool(s, tool, func(ctx context.Context, _ *mcp.CallToolRequest, a A) (*mcp.CallToolResult, O, error) {
		out, err := o.Run(ctx, a)
		return nil, out, err
	})
	return nil
}

func annotations(e Effect) (*mcp.ToolAnnotations, error) {
	yes, no := true, false
	switch e {
	case ReadOnly:
		return &mcp.ToolAnnotations{ReadOnlyHint: true, OpenWorldHint: &no}, nil
	case Additive:
		return &mcp.ToolAnnotations{DestructiveHint: &no, OpenWorldHint: &no}, nil
	case Destructive:
		return &mcp.ToolAnnotations{DestructiveHint: &yes, OpenWorldHint: &no}, nil
	default:
		return nil, fmt.Errorf("effect %d is not ReadOnly, Additive or Destructive", e)
	}
}

// Command returns o as a CLI command whose flags are A's fields, named by
// their json tags with underscores turned into hyphens. It prints the
// result as JSON to out.
func Command[A, O any](o Op[A, O], out io.Writer) (*cobra.Command, error) {
	var a A
	cmd := &cobra.Command{
		Use:   strings.ReplaceAll(o.Name, "_", "-"),
		Short: o.Description,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			res, err := o.Run(cmd.Context(), a)
			if err != nil {
				return err
			}
			return json.NewEncoder(out).Encode(res)
		},
	}
	if err := bindFlags(cmd, reflect.ValueOf(&a).Elem()); err != nil {
		return nil, fmt.Errorf("command %s: %w", o.Name, err)
	}
	return cmd, nil
}

func bindFlags(cmd *cobra.Command, v reflect.Value) error {
	for i := range v.NumField() {
		f := v.Type().Field(i)
		name, _, _ := strings.Cut(f.Tag.Get("json"), ",")
		if name == "" || name == "-" {
			return fmt.Errorf("field %s has no json name", f.Name)
		}
		flag := strings.ReplaceAll(name, "_", "-")
		usage := f.Tag.Get("jsonschema")
		switch p := v.Field(i).Addr().Interface().(type) {
		case *string:
			cmd.Flags().StringVar(p, flag, *p, usage)
		case *int:
			cmd.Flags().IntVar(p, flag, *p, usage)
		case *bool:
			cmd.Flags().BoolVar(p, flag, *p, usage)
		default:
			return fmt.Errorf("field %s has type %s, which has no flag", f.Name, f.Type)
		}
	}
	return nil
}
