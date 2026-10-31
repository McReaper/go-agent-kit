package sweep

import (
	"cmp"
	"fmt"
	"go/ast"
	"go/constant"
	"go/types"
	"slices"
	"strings"

	"golang.org/x/tools/go/packages"
)

// WHY: only a module's internal packages are closed to other modules; a
// library's exported API exists for callers this module never sees.
func isInternal(path string) bool {
	return strings.Contains("/"+path+"/", "/internal/")
}

func (m *model) onlyTests() []Finding {
	var out []Finding
	for obj, id := range m.decls {
		if !isInternal(obj.Pkg().Path()) || m.exempt(obj) || m.reached(obj) {
			continue
		}
		out = append(out, m.finding("only-tests", id, fmt.Sprintf(
			"M5: %s is reached by tests only, or by nothing: delete it, and its tests with it (the owner approves the test change)", qualified(obj))))
	}
	return sorted(out)
}

// WHY: a module's contract and the platform export for other packages by
// design; only a module's inside, under its own internal directory, has
// names nothing outside its package should need.
func isModuleInside(path string) bool {
	return strings.Count("/"+path+"/", "/internal/") >= 2
}

func (m *model) overExported() []Finding {
	inAPI := m.apiTypes()
	var out []Finding
	for obj, id := range m.decls {
		if !obj.Exported() || isMethod(obj) || !isModuleInside(obj.Pkg().Path()) || inAPI[obj] || !m.reached(obj) || m.usedElsewhere(obj) {
			continue
		}
		out = append(out, m.finding("over-exported", id, fmt.Sprintf(
			"M5: %s is exported but only its own package uses it: unexport it, so nothing outside comes to depend on it", qualified(obj))))
	}
	return sorted(out)
}

func (m *model) duplicateMessages() []Finding {
	var out []Finding
	for msg, at := range m.messages {
		if len(at) < 2 {
			continue
		}
		for _, p := range at {
			out = append(out, Finding{"duplicate-log-message", p.Filename, p.Line, fmt.Sprintf(
				"M7: the log message %q is written in %d places: give each one its own message, so a log line points to one place", msg, len(at))})
		}
	}
	return sorted(out)
}

func (m *model) reached(obj types.Object) bool {
	return slices.ContainsFunc(m.uses[obj], func(u use) bool { return !u.test })
}

func (m *model) usedElsewhere(obj types.Object) bool {
	return slices.ContainsFunc(m.uses[obj], func(u use) bool { return u.pkg != obj.Pkg().Path() })
}

func (m *model) exempt(obj types.Object) bool {
	name := obj.Name()
	switch {
	case name == "main" || name == "init" || name == "_":
		return true
	case isMethod(obj):
		return m.ifaceSeen[name] || isStdInterfaceMethod(name)
	}
	return false
}

func isStdInterfaceMethod(name string) bool {
	switch name {
	case "Error", "String", "GoString", "Format", "Unwrap", "Is", "As",
		"MarshalJSON", "UnmarshalJSON", "MarshalText", "UnmarshalText", "MarshalBinary", "UnmarshalBinary",
		"Read", "Write", "Close", "Seek", "ReadFrom", "WriteTo", "ServeHTTP", "LogValue",
		"Len", "Less", "Swap", "Scan", "Value", "Set", "Type":
		return true
	}
	return false
}

func (m *model) finding(rule string, id *ast.Ident, msg string) Finding {
	p := m.fset.Position(id.Pos())
	return Finding{Rule: rule, File: p.Filename, Line: p.Line, Message: msg}
}

func qualified(obj types.Object) string {
	if fn, ok := obj.(*types.Func); ok && fn.Signature().Recv() != nil {
		return fmt.Sprintf("%s.%s", types.TypeString(fn.Signature().Recv().Type(), types.RelativeTo(obj.Pkg())), obj.Name())
	}
	return obj.Pkg().Name() + "." + obj.Name()
}

func sorted(fs []Finding) []Finding {
	slices.SortFunc(fs, func(a, b Finding) int {
		return cmp.Or(cmp.Compare(a.File, b.File), cmp.Compare(a.Line, b.Line), cmp.Compare(a.Rule, b.Rule))
	})
	return fs
}

func (m *model) collectMessages(p *packages.Package, f *ast.File) {
	ast.Inspect(f, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			if msg, ok := logMessage(p.TypesInfo, call); ok {
				m.messages[msg] = append(m.messages[msg], p.Fset.Position(call.Pos()))
			}
		}
		return true
	})
}

func logMessage(info *types.Info, call *ast.CallExpr) (string, bool) {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return "", false
	}
	fn, ok := info.Uses[sel.Sel].(*types.Func)
	if !ok || fn.Pkg() == nil || fn.Pkg().Path() != "log/slog" {
		return "", false
	}
	i := messageIndex(fn.Name())
	if i < 0 || i >= len(call.Args) {
		return "", false
	}
	tv := info.Types[call.Args[i]]
	if tv.Value == nil || tv.Value.Kind() != constant.String {
		return "", false
	}
	return constant.StringVal(tv.Value), true
}

func messageIndex(name string) int {
	switch name {
	case "Debug", "Info", "Warn", "Error":
		return 0
	case "DebugContext", "InfoContext", "WarnContext", "ErrorContext":
		return 1
	case "Log", "LogAttrs":
		return 2
	}
	return -1
}
