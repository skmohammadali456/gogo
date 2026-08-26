// Package types contains GOGO's canonical, language-independent type model.
// Surface grammars resolve their spellings to this package; it deliberately has
// no dependency on tokens, AST nodes, or a particular human language.
package types

import (
	"fmt"
	"sort"
	"strings"
)

type Kind uint8

const (
	Invalid Kind = iota
	StringKind
	NumberKind
	BooleanKind
	BigIntKind
	BytesKind
	ArrayKind
	MapKind
	SetKind
	TupleKind
	RecordKind
	LiteralKind
	OptionalKind
	UnionKind
	IntersectionKind
	ResultKind
)

func (k Kind) String() string {
	names := []string{"invalid", "string", "number", "boolean", "bigint", "bytes", "array", "map", "set", "tuple", "record", "literal", "optional", "union", "intersection", "result"}
	if int(k) >= len(names) {
		return "invalid"
	}
	return names[k]
}

// Type is an immutable value type. Its unexported fields prevent callers from
// changing collection members after construction. The zero Type is Invalid.
type Type struct {
	kind                Kind
	base                Kind // the primitive kind represented by a literal
	text                string
	element, key, value *Type
	members             []Type
	fields              []Field // always sorted by Name; record ordering is not semantic
	index               *IndexSignature
	mutable             bool
}

type Field struct {
	Name     string
	Type     Type
	Optional bool
	Readonly bool
}

// IndexSignature describes values for arbitrary string property names. It is
// intentionally canonical and language-independent; the Step 11 parser does
// not expose an index-signature spelling yet.
type IndexSignature struct {
	Key, Value Type
	Readonly   bool
}

var (
	String  = Type{kind: StringKind}
	Number  = Type{kind: NumberKind}
	Boolean = Type{kind: BooleanKind}
	BigInt  = Type{kind: BigIntKind}
	Bytes   = Type{kind: BytesKind}
)

func (t Type) Kind() Kind      { return t.kind }
func (t Type) IsMutable() bool { return t.mutable }
func (t Type) Element() (Type, bool) {
	if t.element == nil {
		return Type{}, false
	}
	return *t.element, true
}
func (t Type) Key() (Type, bool) {
	if t.key == nil {
		return Type{}, false
	}
	return *t.key, true
}
func (t Type) Value() (Type, bool) {
	if t.value == nil {
		return Type{}, false
	}
	return *t.value, true
}
func (t Type) Members() []Type { return append([]Type(nil), t.members...) }
func (t Type) Fields() []Field { return append([]Field(nil), t.fields...) }
func (t Type) IndexSignature() (IndexSignature, bool) {
	if t.index == nil {
		return IndexSignature{}, false
	}
	return *t.index, true
}
func (t Type) LiteralBase() (Kind, bool) { return t.base, t.kind == LiteralKind }
func (t Type) Ok() (Type, bool)          { return t.Key() }
func (t Type) Err() (Type, bool)         { return t.Value() }

// Array, Map, and Set are mutable collection values by default. Tuple and
// Record are immutable value aggregates.
func Array(element Type) Type {
	return Type{kind: ArrayKind, element: copyType(element), mutable: true}
}
func Map(key, value Type) Type {
	return Type{kind: MapKind, key: copyType(key), value: copyType(value), mutable: true}
}
func Set(element Type) Type { return Type{kind: SetKind, element: copyType(element), mutable: true} }
func Optional(element Type) Type {
	element = Normalize(element)
	return Type{kind: OptionalKind, element: copyType(element), mutable: element.mutable}
}
func Union(members ...Type) (Type, error) { return normalizedComposite(UnionKind, members...) }
func MustUnion(members ...Type) Type {
	t, err := Union(members...)
	if err != nil {
		panic(err)
	}
	return t
}
func Intersection(members ...Type) (Type, error) {
	return normalizedComposite(IntersectionKind, members...)
}
func MustIntersection(members ...Type) Type {
	t, err := Intersection(members...)
	if err != nil {
		panic(err)
	}
	return t
}
func Result(ok, err Type) Type {
	ok = Normalize(ok)
	err = Normalize(err)
	return Type{kind: ResultKind, key: copyType(ok), value: copyType(err), mutable: ok.mutable || err.mutable}
}
func Tuple(members ...Type) Type {
	return Type{kind: TupleKind, members: append([]Type(nil), members...)}
}
func Record(fields ...Field) (Type, error) {
	out := append([]Field(nil), fields...)
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	for i, f := range out {
		if f.Name == "" {
			return Type{}, fmt.Errorf("record field name cannot be empty")
		}
		if i > 0 && out[i-1].Name == f.Name {
			return Type{}, fmt.Errorf("duplicate record field %q", f.Name)
		}
	}
	return Type{kind: RecordKind, fields: out}, nil
}

// Object is the Step 11 structural object constructor. Record remains an alias
// for compatibility with Step 10; both have exactly one representation.
func Object(fields ...Field) (Type, error) { return Record(fields...) }
func MustObject(fields ...Field) Type      { return MustRecord(fields...) }
func RecordWithIndex(index IndexSignature, fields ...Field) (Type, error) {
	t, err := Record(fields...)
	if err != nil {
		return Type{}, err
	}
	if !index.Key.Equal(String) {
		return Type{}, fmt.Errorf("object index key must be string")
	}
	for _, f := range t.fields {
		if !f.Type.AssignableTo(index.Value) {
			return Type{}, fmt.Errorf("object field %q is incompatible with index signature", f.Name)
		}
	}
	t.index = &IndexSignature{Key: index.Key, Value: index.Value, Readonly: index.Readonly}
	return t, nil
}
func MustRecordWithIndex(index IndexSignature, fields ...Field) Type {
	t, err := RecordWithIndex(index, fields...)
	if err != nil {
		panic(err)
	}
	return t
}
func MustRecord(fields ...Field) Type {
	t, err := Record(fields...)
	if err != nil {
		panic(err)
	}
	return t
}
func Literal(base Type, text string) Type {
	if !base.isPrimitive() {
		return Type{}
	}
	return Type{kind: LiteralKind, base: base.kind, text: text}
}
func copyType(t Type) *Type { c := t; return &c }

func Normalize(t Type) Type {
	switch t.kind {
	case ArrayKind:
		e := Normalize(*t.element)
		t.element = copyType(e)
	case MapKind:
		k, v := Normalize(*t.key), Normalize(*t.value)
		t.key, t.value = copyType(k), copyType(v)
	case SetKind, OptionalKind:
		e := Normalize(*t.element)
		t.element = copyType(e)
		if t.kind == OptionalKind {
			t.mutable = e.mutable
		}
	case TupleKind:
		for i := range t.members {
			t.members[i] = Normalize(t.members[i])
		}
	case RecordKind:
		for i := range t.fields {
			t.fields[i].Type = Normalize(t.fields[i].Type)
		}
	case ResultKind:
		ok, er := Normalize(*t.key), Normalize(*t.value)
		t.key, t.value = copyType(ok), copyType(er)
		t.mutable = ok.mutable || er.mutable
	case UnionKind, IntersectionKind:
		n, _ := normalizedComposite(t.kind, t.members...)
		return n
	}
	return t
}

func normalizedComposite(kind Kind, members ...Type) (Type, error) {
	var flat []Type
	for _, m := range members {
		m = Normalize(m)
		if m.kind == Invalid {
			return Type{}, fmt.Errorf("%s member cannot be invalid", kind)
		}
		if m.kind == kind {
			flat = append(flat, m.members...)
		} else {
			flat = append(flat, m)
		}
	}
	if len(flat) == 0 {
		return Type{}, fmt.Errorf("%s requires at least one member", kind)
	}
	sort.SliceStable(flat, func(i, j int) bool { return flat[i].String() < flat[j].String() })
	uniq := flat[:0]
	for _, m := range flat {
		if len(uniq) == 0 || !m.Equal(uniq[len(uniq)-1]) {
			uniq = append(uniq, m)
		}
	}
	uniq = simplifyCompositeMembers(kind, uniq)
	if len(uniq) == 1 {
		return uniq[0], nil
	}
	if kind == IntersectionKind {
		if o, ok := combineObjectIntersection(uniq); ok {
			return o, nil
		}
	}
	mut := false
	for _, m := range uniq {
		mut = mut || m.mutable
	}
	return Type{kind: kind, members: append([]Type(nil), uniq...), mutable: mut}, nil
}

func simplifyCompositeMembers(kind Kind, members []Type) []Type {
	keep := make([]bool, len(members))
	for i := range keep {
		keep[i] = true
	}
	for i, m := range members {
		for j, other := range members {
			if i == j || !keep[i] {
				continue
			}
			switch kind {
			case UnionKind:
				if m.AssignableTo(other) && !other.AssignableTo(m) {
					keep[i] = false
				}
			case IntersectionKind:
				if other.AssignableTo(m) && !m.AssignableTo(other) {
					keep[i] = false
				}
			}
		}
	}
	out := members[:0]
	for i, m := range members {
		if keep[i] {
			out = append(out, m)
		}
	}
	return out
}

func combineObjectIntersection(ms []Type) (Type, bool) {
	fields := map[string]Field{}
	var index *IndexSignature
	allObj := true
	for _, m := range ms {
		if m.kind != RecordKind {
			allObj = false
			break
		}
		if m.index != nil {
			if index == nil {
				c := *m.index
				index = &c
			} else {
				if !index.Key.Equal(m.index.Key) || !index.Value.Equal(m.index.Value) {
					return Type{kind: IntersectionKind, members: ms}, false
				}
				index.Readonly = index.Readonly || m.index.Readonly
			}
		}
		for _, f := range m.fields {
			if old, ok := fields[f.Name]; ok {
				if !old.Type.Equal(f.Type) {
					return Type{kind: IntersectionKind, members: ms}, false
				}
				f.Optional = old.Optional && f.Optional
				f.Readonly = old.Readonly || f.Readonly
			}
			fields[f.Name] = f
		}
	}
	if !allObj {
		return Type{}, false
	}
	out := make([]Field, 0, len(fields))
	for _, f := range fields {
		out = append(out, f)
	}
	var (
		t   Type
		err error
	)
	if index != nil {
		t, err = RecordWithIndex(*index, out...)
	} else {
		t, err = Object(out...)
	}
	if err != nil {
		return Type{kind: IntersectionKind, members: ms}, false
	}
	return t, true
}

func (t Type) isPrimitive() bool { return t.kind >= StringKind && t.kind <= BytesKind }

// Equal is deterministic structural equality. Literal types compare their
// primitive base and exact canonical literal text; records are name-sorted.
func (t Type) Equal(u Type) bool {
	if t.kind != u.kind || t.mutable != u.mutable || t.base != u.base || t.text != u.text {
		return false
	}
	if !samePtr(t.element, u.element) || !samePtr(t.key, u.key) || !samePtr(t.value, u.value) || len(t.members) != len(u.members) || len(t.fields) != len(u.fields) {
		return false
	}
	for i := range t.members {
		if !t.members[i].Equal(u.members[i]) {
			return false
		}
	}
	for i := range t.fields {
		if t.fields[i].Name != u.fields[i].Name || t.fields[i].Optional != u.fields[i].Optional || t.fields[i].Readonly != u.fields[i].Readonly || !t.fields[i].Type.Equal(u.fields[i].Type) {
			return false
		}
	}
	if t.index == nil || u.index == nil {
		return t.index == nil && u.index == nil
	}
	return t.index.Readonly == u.index.Readonly && t.index.Key.Equal(u.index.Key) && t.index.Value.Equal(u.index.Value)
}
func samePtr(a, b *Type) bool {
	if a == nil || b == nil {
		return a == b
	}
	return a.Equal(*b)
}

// AssignableTo is deliberately distinct from equality: a literal is
// assignable to its primitive base, but is not equal to it. Collections are
// invariant, tuples require equal arity, and records require exactly the same
// named fields in this Step 10 closed-record model.
func (t Type) AssignableTo(target Type) bool {
	t, target = Normalize(t), Normalize(target)
	if t.Equal(target) {
		return true
	}
	if target.kind == OptionalKind || t.kind == OptionalKind {
		return target.kind == OptionalKind && t.kind == OptionalKind && t.element.AssignableTo(*target.element)
	}
	if target.kind == UnionKind {
		for _, m := range target.members {
			if t.AssignableTo(m) {
				return true
			}
		}
		return false
	}
	if t.kind == UnionKind {
		for _, m := range t.members {
			if !m.AssignableTo(target) {
				return false
			}
		}
		return true
	}
	if target.kind == IntersectionKind {
		for _, m := range target.members {
			if !t.AssignableTo(m) {
				return false
			}
		}
		return true
	}
	if t.kind == IntersectionKind {
		for _, m := range t.members {
			if m.AssignableTo(target) {
				return true
			}
		}
		return false
	}
	if t.kind == ResultKind && target.kind == ResultKind {
		return t.key.AssignableTo(*target.key) && t.value.AssignableTo(*target.value)
	}
	if t.kind == LiteralKind && target.isPrimitive() {
		return t.base == target.kind
	}
	if t.kind == RecordKind && target.kind == RecordKind {
		return assignObject(t, target)
	}
	return false
}

// assignObject is width-subtyping for objects: every target required property
// must be present in source; source optional cannot satisfy target required.
// A readonly source cannot satisfy a mutable target, while mutable source may
// satisfy readonly target. Extra source properties are allowed, except that a
// target index signature constrains them. Index signatures must be provided by
// the source as well because arbitrary reads cannot be proven otherwise.
func assignObject(source, target Type) bool {
	byName := make(map[string]Field, len(source.fields))
	for _, f := range source.fields {
		byName[f.Name] = f
	}
	for _, want := range target.fields {
		got, ok := byName[want.Name]
		if !ok {
			if want.Optional {
				continue
			}
			return false
		}
		if !want.Optional && got.Optional {
			return false
		}
		if got.Readonly && !want.Readonly {
			return false
		}
		if !got.Type.AssignableTo(want.Type) {
			return false
		}
	}
	if target.index != nil {
		if source.index == nil || (source.index.Readonly && !target.index.Readonly) || !source.index.Key.AssignableTo(target.index.Key) || !source.index.Value.AssignableTo(target.index.Value) {
			return false
		}
		for _, f := range source.fields {
			if !f.Type.AssignableTo(target.index.Value) {
				return false
			}
		}
	}
	return true
}

func (t Type) String() string {
	switch t.kind {
	case LiteralKind:
		return fmt.Sprintf("%s(%q)", t.base, t.text)
	case OptionalKind:
		return "optional<" + t.element.String() + ">"
	case UnionKind:
		return strings.Join(typeStrings(t.members), " | ")
	case IntersectionKind:
		return strings.Join(typeStrings(t.members), " & ")
	case ResultKind:
		return "result<" + t.key.String() + ", " + t.value.String() + ">"
	case ArrayKind:
		return "array<" + t.element.String() + ">"
	case SetKind:
		return "set<" + t.element.String() + ">"
	case MapKind:
		return "map<" + t.key.String() + ", " + t.value.String() + ">"
	case TupleKind:
		return "tuple<" + join(t.members) + ">"
	case RecordKind:
		parts := make([]string, len(t.fields))
		for i, f := range t.fields {
			prefix := ""
			if f.Readonly {
				prefix = "readonly "
			}
			suffix := ""
			if f.Optional {
				suffix = "?"
			}
			parts[i] = prefix + f.Name + suffix + ": " + f.Type.String()
		}
		body := "record{" + strings.Join(parts, ", ") + "}"
		if t.index != nil {
			body = "record{[string]: " + t.index.Value.String() + ", " + strings.Join(parts, ", ") + "}"
		}
		return body
	default:
		return t.kind.String()
	}
}
func join(ts []Type) string { return strings.Join(typeStrings(ts), ", ") }
func typeStrings(ts []Type) []string {
	p := make([]string, len(ts))
	for i := range ts {
		p[i] = ts[i].String()
	}
	return p
}
