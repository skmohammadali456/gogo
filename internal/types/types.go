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
)

func (k Kind) String() string {
	names := []string{"invalid", "string", "number", "boolean", "bigint", "bytes", "array", "map", "set", "tuple", "record", "literal"}
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
	mutable             bool
}

type Field struct {
	Name string
	Type Type
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
func (t Type) Members() []Type           { return append([]Type(nil), t.members...) }
func (t Type) Fields() []Field           { return append([]Field(nil), t.fields...) }
func (t Type) LiteralBase() (Kind, bool) { return t.base, t.kind == LiteralKind }

// Array, Map, and Set are mutable collection values by default. Tuple and
// Record are immutable value aggregates.
func Array(element Type) Type {
	return Type{kind: ArrayKind, element: copyType(element), mutable: true}
}
func Map(key, value Type) Type {
	return Type{kind: MapKind, key: copyType(key), value: copyType(value), mutable: true}
}
func Set(element Type) Type { return Type{kind: SetKind, element: copyType(element), mutable: true} }
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
func copyType(t Type) *Type      { c := t; return &c }
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
		if t.fields[i].Name != u.fields[i].Name || !t.fields[i].Type.Equal(u.fields[i].Type) {
			return false
		}
	}
	return true
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
	if t.Equal(target) {
		return true
	}
	if t.kind == LiteralKind && target.isPrimitive() {
		return t.base == target.kind
	}
	return false
}

func (t Type) String() string {
	switch t.kind {
	case LiteralKind:
		return fmt.Sprintf("%s(%q)", t.base, t.text)
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
			parts[i] = f.Name + ": " + f.Type.String()
		}
		return "record{" + strings.Join(parts, ", ") + "}"
	default:
		return t.kind.String()
	}
}
func join(ts []Type) string {
	p := make([]string, len(ts))
	for i := range ts {
		p[i] = ts[i].String()
	}
	return strings.Join(p, ", ")
}
