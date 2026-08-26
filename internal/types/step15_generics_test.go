package types

import "testing"

func TestStep15GenericParameterIdentityAndInstantiationEquality(t *testing.T) {
	a := TypeParam("func a/T", "T")
	b := TypeParam("func b/T", "T")
	if a.Equal(b) {
		t.Fatal("type parameters from different declarations must not compare equal")
	}
	boxNumber1 := GenericInstance("Box", Number)
	boxNumber2 := GenericInstance("Box", Number)
	boxString := GenericInstance("Box", String)
	if !boxNumber1.Equal(boxNumber2) {
		t.Fatal("equivalent generic instances should compare equal")
	}
	if boxNumber1.Equal(boxString) {
		t.Fatal("different type arguments must not compare equal")
	}
}

func TestStep15GenericParameterStructuralNesting(t *testing.T) {
	tp := TypeParam("type Nest/T", "T")
	obj := MustObject(Field{Name: "value", Type: Result(Optional(Array(tp)), String)})
	if obj.Kind() != RecordKind {
		t.Fatal("generic parameter should nest in ordinary canonical object types")
	}
}
