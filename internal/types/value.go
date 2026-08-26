package types

import "fmt"

// Value is an immutable canonical runtime value boundary. Its fields are
// private: callers can only construct a value through a type-checked
// constructor, so an invalid type/value pair cannot be represented.
type Value struct {
	typ  Type
	data any
}

func (v Value) Type() Type { return v.typ }
func (v Value) Data() any  { return cloneData(v.data) }

func cloneData(data any) any {
	switch value := data.(type) {
	case []byte:
		return append([]byte(nil), value...)
	case []Value:
		return append([]Value(nil), value...)
	case []MapEntry:
		return append([]MapEntry(nil), value...)
	case map[string]Value:
		out := make(map[string]Value, len(value))
		for k, v := range value {
			out[k] = v
		}
		return out
	default:
		return value
	}
}

func StringValue(value string) Value { return Value{typ: String, data: value} }
func NumberValue(text string) Value  { return Value{typ: Number, data: text} }
func BooleanValue(value bool) Value  { return Value{typ: Boolean, data: value} }
func BigIntValue(text string) Value  { return Value{typ: BigInt, data: text} }
func BytesValue(value []byte) Value  { return Value{typ: Bytes, data: append([]byte(nil), value...)} }

func ArrayValue(t Type, values []Value) (Value, error) {
	element, ok := t.Element()
	if !ok || t.Kind() != ArrayKind {
		return Value{}, fmt.Errorf("array value requires an array type")
	}
	for _, v := range values {
		if !v.typ.AssignableTo(element) {
			return Value{}, fmt.Errorf("array element is not assignable to %s", element)
		}
	}
	return Value{typ: t, data: append([]Value(nil), values...)}, nil
}
func SetValue(t Type, values []Value) (Value, error) {
	element, ok := t.Element()
	if !ok || t.Kind() != SetKind {
		return Value{}, fmt.Errorf("set value requires a set type")
	}
	for _, v := range values {
		if !v.typ.AssignableTo(element) {
			return Value{}, fmt.Errorf("set element is not assignable to %s", element)
		}
	}
	return Value{typ: t, data: append([]Value(nil), values...)}, nil
}
func TupleValue(t Type, values []Value) (Value, error) {
	members := t.Members()
	if t.Kind() != TupleKind || len(members) != len(values) {
		return Value{}, fmt.Errorf("tuple value does not match tuple type")
	}
	for i, v := range values {
		if !v.typ.AssignableTo(members[i]) {
			return Value{}, fmt.Errorf("tuple value at position %d is incompatible", i)
		}
	}
	return Value{typ: t, data: append([]Value(nil), values...)}, nil
}

type MapEntry struct{ Key, Value Value }

func MapValue(t Type, entries []MapEntry) (Value, error) {
	key, keyOK := t.Key()
	value, valueOK := t.Value()
	if t.Kind() != MapKind || !keyOK || !valueOK {
		return Value{}, fmt.Errorf("map value requires a map type")
	}
	for _, e := range entries {
		if !e.Key.typ.AssignableTo(key) || !e.Value.typ.AssignableTo(value) {
			return Value{}, fmt.Errorf("map entry is incompatible with map type")
		}
	}
	return Value{typ: t, data: append([]MapEntry(nil), entries...)}, nil
}
func RecordValue(t Type, fields map[string]Value) (Value, error) {
	if t.Kind() != RecordKind {
		return Value{}, fmt.Errorf("record value requires a record type")
	}
	expected := t.Fields()
	byName := make(map[string]Field, len(expected))
	for _, f := range expected {
		byName[f.Name] = f
	}
	copyFields := make(map[string]Value, len(fields))
	for name, v := range fields {
		f, known := byName[name]
		if known {
			if !v.typ.AssignableTo(f.Type) {
				return Value{}, fmt.Errorf("record field %q is incompatible", name)
			}
		} else if index, ok := t.IndexSignature(); !ok || !v.typ.AssignableTo(index.Value) {
			return Value{}, fmt.Errorf("record field %q is not allowed", name)
		}
		copyFields[name] = v
	}
	for _, f := range expected {
		if _, ok := fields[f.Name]; !ok && !f.Optional {
			return Value{}, fmt.Errorf("record field %q is required", f.Name)
		}
	}
	return Value{typ: t, data: copyFields}, nil
}
