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
	if t.Equal(target) {
		return true
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
func join(ts []Type) string {
	p := make([]string, len(ts))
	for i := range ts {
		p[i] = ts[i].String()
	}
	return strings.Join(p, ", ")
}
