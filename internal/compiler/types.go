package compiler

import (
	"fmt"
	"strings"

	"github.com/skmohammadali786/gogo/internal/ast"
	"github.com/skmohammadali786/gogo/internal/types"
)

// ResolveType resolves a standalone annotation. Session resolution additionally
// supplies locally declared aliases, but both paths construct the same types.Type.
func ResolveType(ref ast.TypeRef) (types.Type, error) { return resolveType(ref, nil, nil) }

func resolveType(ref ast.TypeRef, aliases map[string]ast.TypeRef, resolving map[string]bool) (types.Type, error) {
	if len(ref.Union) > 0 {
		members := make([]types.Type, len(ref.Union))
		for i := range ref.Union {
			var err error
			members[i], err = resolveType(ref.Union[i], aliases, resolving)
			if err != nil {
				return types.Type{}, err
			}
		}
		return types.Union(members...)
	}
	if len(ref.Intersection) > 0 {
		members := make([]types.Type, len(ref.Intersection))
		for i := range ref.Intersection {
			var err error
			members[i], err = resolveType(ref.Intersection[i], aliases, resolving)
			if err != nil {
				return types.Type{}, err
			}
		}
		return types.Intersection(members...)
	}
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
	case "optional":
		if len(ref.Arguments) != 1 {
			return types.Type{}, fmt.Errorf("optional requires exactly one type")
		}
		e, err := resolveType(ref.Arguments[0], aliases, resolving)
		if err != nil {
			return types.Type{}, err
		}
		t = types.Optional(e)
	case "result":
		if len(ref.Arguments) != 2 {
			return types.Type{}, fmt.Errorf("result requires ok and err types")
		}
		okt, err := resolveType(ref.Arguments[0], aliases, resolving)
		if err != nil {
			return types.Type{}, err
		}
		errt, err := resolveType(ref.Arguments[1], aliases, resolving)
		if err != nil {
			return types.Type{}, err
		}
		t = types.Result(okt, errt)
	case "array":
		if len(ref.Arguments) != 1 {
			return types.Type{}, fmt.Errorf("array requires exactly one element type")
		}
		e, err := resolveType(ref.Arguments[0], aliases, resolving)
		if err != nil {
			return types.Type{}, err
		}
		t = types.Array(e)
	case "map":
		if len(ref.Arguments) != 2 {
			return types.Type{}, fmt.Errorf("map requires key and value types")
		}
		k, err := resolveType(ref.Arguments[0], aliases, resolving)
		if err != nil {
			return types.Type{}, err
		}
		v, err := resolveType(ref.Arguments[1], aliases, resolving)
		if err != nil {
			return types.Type{}, err
		}
		t = types.Map(k, v)
	case "set":
		if len(ref.Arguments) != 1 {
			return types.Type{}, fmt.Errorf("set requires exactly one element type")
		}
		e, err := resolveType(ref.Arguments[0], aliases, resolving)
		if err != nil {
			return types.Type{}, err
		}
		t = types.Set(e)
	case "tuple":
		items := make([]types.Type, len(ref.Arguments))
		for i := range ref.Arguments {
			var err error
			items[i], err = resolveType(ref.Arguments[i], aliases, resolving)
			if err != nil {
				return types.Type{}, err
			}
		}
		t = types.Tuple(items...)
	case "record", "object":
		fields := make([]types.Field, len(ref.Fields))
		for i, f := range ref.Fields {
			ft, err := resolveType(f.Type, aliases, resolving)
			if err != nil {
				return types.Type{}, err
			}
			fields[i] = types.Field{Name: f.Name, Type: ft, Optional: f.Optional, Readonly: f.Readonly}
		}
		var err error
		t, err = types.Object(fields...)
		if err != nil {
			return types.Type{}, err
		}
	default:
		if len(ref.Name) > 0 && (ref.Name[0] == '"' || ref.Name[0] == '\'') {
			t = types.Literal(types.String, ref.Name)
			break
		}
		if len(ref.Name) > 0 && ref.Name[0] >= '0' && ref.Name[0] <= '9' {
			t = types.Literal(types.Number, ref.Name)
			break
		}
		if aliases == nil {
			return types.Type{}, fmt.Errorf("unknown canonical type %q", ref.Name)
		}
		alias, ok := aliases[ref.Name]
		if !ok {
			return types.Type{}, fmt.Errorf("unresolved type alias %q", ref.Name)
		}
		if resolving[ref.Name] {
			return types.Type{}, fmt.Errorf("cyclic type alias %q", ref.Name)
		}
		resolving[ref.Name] = true
		var err error
		t, err = resolveType(alias, aliases, resolving)
		delete(resolving, ref.Name)
		if err != nil {
			return types.Type{}, err
		}
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
