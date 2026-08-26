package compiler

import (
	"fmt"
	"strings"

	"github.com/skmohammadali786/gogo/internal/ast"
	"github.com/skmohammadali786/gogo/internal/types"
)

// ResolveType is the sole conversion path from a surface AST type annotation
// to a canonical GOGO type. Names are intentionally normalized here, not in a
// grammar or parser, so all vocabularies share one semantic identity.
func ResolveType(ref ast.TypeRef) (types.Type, error) {
	name := strings.ToLower(ref.Name)
	var t types.Type
	switch name {
	case "string", "text":
		t = types.String
	case "number":
		t = types.Number
	case "boolean", "bool":
		t = types.Boolean
	case "bigint":
		t = types.BigInt
	case "bytes":
		t = types.Bytes
	case "array":
		if len(ref.Arguments) != 1 {
			return types.Type{}, fmt.Errorf("array requires exactly one element type")
		}
		e, err := ResolveType(ref.Arguments[0])
		if err != nil {
			return types.Type{}, err
		}
		t = types.Array(e)
	case "map":
		if len(ref.Arguments) != 2 {
			return types.Type{}, fmt.Errorf("map requires key and value types")
		}
		k, err := ResolveType(ref.Arguments[0])
		if err != nil {
			return types.Type{}, err
		}
		v, err := ResolveType(ref.Arguments[1])
		if err != nil {
			return types.Type{}, err
		}
		t = types.Map(k, v)
	case "set":
		if len(ref.Arguments) != 1 {
			return types.Type{}, fmt.Errorf("set requires exactly one element type")
		}
		e, err := ResolveType(ref.Arguments[0])
		if err != nil {
			return types.Type{}, err
		}
		t = types.Set(e)
	case "tuple":
		items := make([]types.Type, len(ref.Arguments))
		for i := range ref.Arguments {
			var err error
			items[i], err = ResolveType(ref.Arguments[i])
			if err != nil {
				return types.Type{}, err
			}
		}
		t = types.Tuple(items...)
	case "record", "object":
		fields := make([]types.Field, len(ref.Fields))
		for i, f := range ref.Fields {
			ft, err := ResolveType(f.Type)
			if err != nil {
				return types.Type{}, err
			}
			fields[i] = types.Field{Name: f.Name, Type: ft}
		}
		var err error
		t, err = types.Record(fields...)
		if err != nil {
			return types.Type{}, err
		}
	default:
		return types.Type{}, fmt.Errorf("unknown canonical type %q", ref.Name)
	}
	if ref.Array {
		t = types.Array(t)
	}
	return t, nil
}

func literalType(l ast.Literal) types.Type {
	if len(l.Text) > 0 && (l.Text[0] == '"' || l.Text[0] == '\'') {
		return types.Literal(types.String, l.Text)
	}
	if strings.HasSuffix(l.Text, "n") {
		return types.Literal(types.BigInt, l.Text)
	}
	return types.Literal(types.Number, l.Text)
}
