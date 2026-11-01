package sweep

import (
	"go/types"
	"slices"
)

// WHY: a type an exported declaration of its package mentions (a
// parameter, a result, an exported field) is passed and received by callers
// elsewhere, so it stays exported even when no other package writes its name.
func (m *model) apiTypes() map[types.Object]bool {
	seen := map[types.Object]bool{}
	for obj := range m.decls {
		if !obj.Exported() {
			continue
		}
		switch o := obj.(type) {
		case *types.Func:
			mention(o.Signature(), o.Pkg(), seen)
		case *types.TypeName:
			fields(o, seen)
		case *types.Var, *types.Const:
			mention(o.Type(), o.Pkg(), seen)
		}
	}
	return seen
}

func fields(tn *types.TypeName, seen map[types.Object]bool) {
	st, ok := tn.Type().Underlying().(*types.Struct)
	if !ok {
		return
	}
	for f := range st.Fields() {
		if f.Exported() {
			mention(f.Type(), tn.Pkg(), seen)
		}
	}
}

func mention(t types.Type, pkg *types.Package, seen map[types.Object]bool) {
	for _, inner := range parts(t) {
		mention(inner, pkg, seen)
	}
	if n, ok := t.(*types.Named); ok && n.Obj().Pkg() == pkg {
		seen[n.Obj()] = true
	}
}

func parts(t types.Type) []types.Type {
	switch t := t.(type) {
	case *types.Named:
		return slices.Collect(t.TypeArgs().Types())
	case *types.Pointer:
		return []types.Type{t.Elem()}
	case *types.Slice:
		return []types.Type{t.Elem()}
	case *types.Array:
		return []types.Type{t.Elem()}
	case *types.Chan:
		return []types.Type{t.Elem()}
	case *types.Map:
		return []types.Type{t.Key(), t.Elem()}
	case *types.Signature:
		return append(tupleTypes(t.Params()), tupleTypes(t.Results())...)
	}
	return nil
}

func tupleTypes(t *types.Tuple) []types.Type {
	var out []types.Type
	for v := range t.Variables() {
		out = append(out, v.Type())
	}
	return out
}
