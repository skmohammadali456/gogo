package types

import "testing"

func TestStructuralEqualityAndAssignability(t *testing.T) {
	a := Array(String)
	b := Array(String)
	if !a.Equal(b) || a.Equal(Array(Number)) {
		t.Fatal("array equality must be structural")
	}
	nested := Map(String, Tuple(Array(String), Set(Boolean)))
	if !nested.Equal(Map(String, Tuple(Array(String), Set(Boolean)))) {
		t.Fatal("nested equality")
	}
	left := MustRecord(Field{Name: "name", Type: String}, Field{Name: "age", Type: Number})
	right := MustRecord(Field{Name: "age", Type: Number}, Field{Name: "name", Type: String})
	if !left.Equal(right) {
		t.Fatal("record order must not matter")
	}
	for _, primitive := range []Type{String, Number, Boolean, BigInt, Bytes} {
		literal := Literal(primitive, "value")
		if literal.Equal(primitive) || !literal.AssignableTo(primitive) {
			t.Fatalf("literal relation for %s", primitive)
		}
	}
	lit := Literal(String, "Alex")
	if lit.Equal(String) || !lit.AssignableTo(String) || lit.AssignableTo(Number) {
		t.Fatal("literal rules")
	}
	if Tuple(String).AssignableTo(Tuple(String, Number)) || Map(String, Number).AssignableTo(Map(String, String)) {
		t.Fatal("collections are invariant")
	}
}

func TestRecordAndMutability(t *testing.T) {
	if _, err := Record(Field{Name: "x", Type: String}, Field{Name: "x", Type: Number}); err == nil {
		t.Fatal("duplicate fields accepted")
	}
	if !Array(String).IsMutable() || Tuple(String).IsMutable() || RecordMust().IsMutable() {
		t.Fatal("collection mutability")
	}
}
func RecordMust() Type { return MustRecord() }

func TestRuntimeValuesRejectIncompatibleShapes(t *testing.T) {
	array := Array(String)
	if _, err := ArrayValue(array, []Value{NumberValue("1")}); err == nil {
		t.Fatal("incompatible array value accepted")
	}
	if _, err := TupleValue(Tuple(String, Number), []Value{StringValue("a")}); err == nil {
		t.Fatal("short tuple accepted")
	}
	record := MustRecord(Field{Name: "name", Type: String})
	if _, err := RecordValue(record, map[string]Value{"extra": StringValue("x")}); err == nil {
		t.Fatal("wrong record fields accepted")
	}
	bytes := BytesValue([]byte{1})
	copy := bytes.Data().([]byte)
	copy[0] = 2
	if bytes.Data().([]byte)[0] != 1 {
		t.Fatal("runtime value leaked mutable bytes")
	}
}

func TestStep11ObjectPropertiesAndIndexCompatibility(t *testing.T) {
	wide := MustObject(Field{Name: "name", Type: String}, Field{Name: "meta", Type: MustObject(Field{Name: "enabled", Type: Boolean})}, Field{Name: "tag", Type: String, Optional: true})
	narrow := MustObject(Field{Name: "name", Type: String})
	if !wide.AssignableTo(narrow) || narrow.AssignableTo(wide) {
		t.Fatal("required and optional structural compatibility")
	}
	if MustObject(Field{Name: "name", Type: String, Optional: true}).AssignableTo(narrow) {
		t.Fatal("optional source satisfied required target")
	}
	ro := MustObject(Field{Name: "name", Type: String, Readonly: true})
	if !narrow.AssignableTo(ro) || ro.AssignableTo(narrow) {
		t.Fatal("readonly variance")
	}
	if wide.Equal(MustObject(Field{Name: "name", Type: String}, Field{Name: "meta", Type: MustObject(Field{Name: "enabled", Type: Boolean})}, Field{Name: "tag", Type: String, Optional: true, Readonly: true})) {
		t.Fatal("property flags participate in equality")
	}
	indexed, err := RecordWithIndex(IndexSignature{Key: String, Value: String}, Field{Name: "name", Type: String})
	if err != nil {
		t.Fatal(err)
	}
	if MustObject(Field{Name: "name", Type: String}).AssignableTo(indexed) || indexed.AssignableTo(MustObject(Field{Name: "name", Type: Number})) {
		t.Fatal("index compatibility")
	}
	if _, err := RecordWithIndex(IndexSignature{Key: String, Value: Number}, Field{Name: "name", Type: String}); err == nil {
		t.Fatal("explicit property violates index")
	}
}

func TestStep11ObjectRuntimeOptionalAndExtra(t *testing.T) {
	typ := MustObject(Field{Name: "name", Type: String}, Field{Name: "tag", Type: String, Optional: true})
	if _, err := RecordValue(typ, map[string]Value{"name": StringValue("a")}); err != nil {
		t.Fatal(err)
	}
	if _, err := RecordValue(typ, map[string]Value{}); err == nil {
		t.Fatal("missing required accepted")
	}
	if _, err := RecordValue(typ, map[string]Value{"name": StringValue("a"), "extra": StringValue("x")}); err == nil {
		t.Fatal("extra closed property accepted")
	}
}
