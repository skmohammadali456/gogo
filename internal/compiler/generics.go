package compiler

import (
	"fmt"

	"github.com/skmohammadali786/gogo/internal/ast"
	"github.com/skmohammadali786/gogo/internal/types"
)

type genericDecl struct {
	owner  string
	params []ast.GenericParam
	body   ast.TypeRef
}

type genericEnv map[string]types.Type

func typeParamEnv(owner string, ps []ast.GenericParam) genericEnv {
	env := genericEnv{}
	for _, p := range ps {
		env[p.Name] = types.TypeParam(owner+"/"+p.Name, p.Name)
	}
	return env
}

func substituteType(t types.Type, sub map[string]types.Type, depth int) (types.Type, error) {
	if depth > 64 {
		return types.Type{}, fmt.Errorf("generic instantiation is too deeply recursive")
	}
	switch t.Kind() {
	case types.TypeParamKind:
		if v, ok := sub[t.TypeParamID()]; ok {
			return v, nil
		}
		return t, nil
	case types.ArrayKind:
		e, _ := t.Element()
		ne, err := substituteType(e, sub, depth+1)
		if err != nil {
			return types.Type{}, err
		}
		return types.Array(ne), nil
	case types.SetKind:
		e, _ := t.Element()
		ne, err := substituteType(e, sub, depth+1)
		if err != nil {
			return types.Type{}, err
		}
		return types.Set(ne), nil
	case types.OptionalKind:
		e, _ := t.Element()
		ne, err := substituteType(e, sub, depth+1)
		if err != nil {
			return types.Type{}, err
		}
		return types.Optional(ne), nil
	case types.MapKind:
		k, _ := t.Key()
		v, _ := t.Value()
		nk, err := substituteType(k, sub, depth+1)
		if err != nil {
			return types.Type{}, err
		}
		nv, err := substituteType(v, sub, depth+1)
		if err != nil {
			return types.Type{}, err
		}
		return types.Map(nk, nv), nil
	case types.ResultKind:
		ok, _ := t.Ok()
		er, _ := t.Err()
		nok, err := substituteType(ok, sub, depth+1)
		if err != nil {
			return types.Type{}, err
		}
		ner, err := substituteType(er, sub, depth+1)
		if err != nil {
			return types.Type{}, err
		}
		return types.Result(nok, ner), nil
	case types.TupleKind:
		ms := t.Members()
		for i := range ms {
			var err error
			ms[i], err = substituteType(ms[i], sub, depth+1)
			if err != nil {
				return types.Type{}, err
			}
		}
		return types.Tuple(ms...), nil
	case types.RecordKind:
		fs := t.Fields()
		for i := range fs {
			var err error
			fs[i].Type, err = substituteType(fs[i].Type, sub, depth+1)
			if err != nil {
				return types.Type{}, err
			}
		}
		return types.Object(fs...)
	case types.UnionKind, types.IntersectionKind:
		ms := t.Members()
		for i := range ms {
			var err error
			ms[i], err = substituteType(ms[i], sub, depth+1)
			if err != nil {
				return types.Type{}, err
			}
		}
		if t.Kind() == types.UnionKind {
			return types.Union(ms...)
		}
		return types.Intersection(ms...)
	case types.EnumKind:
		vs := t.Variants()
		ev := make([]types.EnumVariant, len(vs))
		for i, v := range vs {
			if p, ok := v.VariantPayload(); ok {
				np, err := substituteType(p, sub, depth+1)
				if err != nil {
					return types.Type{}, err
				}
				ev[i] = types.EnumVariant{Name: v.VariantName(), Payload: np}
			} else {
				ev[i] = types.EnumVariant{Name: v.VariantName()}
			}
		}
		return types.Enum(t.EnumName(), ev...)
	case types.GenericInstanceKind:
		args := t.TypeArguments()
		for i := range args {
			var err error
			args[i], err = substituteType(args[i], sub, depth+1)
			if err != nil {
				return types.Type{}, err
			}
		}
		return types.GenericInstance(t.TypeParamName(), args...), nil
	}
	return t, nil
}

func bindInference(pattern, actual types.Type, out map[string]types.Type) bool {
	pattern, actual = types.Normalize(pattern), types.Normalize(actual)
	if pattern.Kind() == types.TypeParamKind {
		if old, ok := out[pattern.TypeParamID()]; ok {
			actual = normalizeLiteral(actual)
			return actual.AssignableTo(old) || old.AssignableTo(actual)
		}
		out[pattern.TypeParamID()] = types.Normalize(normalizeLiteral(actual))
		return true
	}
	if pattern.Kind() != actual.Kind() {
		if actual.AssignableTo(pattern) {
			return true
		}
		return false
	}
	switch pattern.Kind() {
	case types.ArrayKind, types.SetKind, types.OptionalKind:
		pe, _ := pattern.Element()
		ae, _ := actual.Element()
		return bindInference(pe, ae, out)
	case types.MapKind:
		pk, _ := pattern.Key()
		pv, _ := pattern.Value()
		ak, _ := actual.Key()
		av, _ := actual.Value()
		return bindInference(pk, ak, out) && bindInference(pv, av, out)
	case types.ResultKind:
		po, _ := pattern.Ok()
		pe, _ := pattern.Err()
		ao, _ := actual.Ok()
		ae, _ := actual.Err()
		return bindInference(po, ao, out) && bindInference(pe, ae, out)
	case types.TupleKind:
		pm := pattern.Members()
		am := actual.Members()
		if len(pm) != len(am) {
			return false
		}
		for i := range pm {
			if !bindInference(pm[i], am[i], out) {
				return false
			}
		}
		return true
	case types.RecordKind:
		for _, pf := range pattern.Fields() {
			af, ok := fieldOf(actual, pf.Name)
			if !ok || !bindInference(pf.Type, af.Type, out) {
				return false
			}
		}
		return true
	}
	return true
}
