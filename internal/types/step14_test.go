package types

import "testing"

func TestStep14EnumCanonicalIdentityAndPayload(t *testing.T) {
	e := MustEnum("State", EnumVariant{Name: "Ready"}, EnumVariant{Name: "Loading", Payload: String})
	vs := e.Variants()
	if len(vs) != 2 || !vs[0].AssignableTo(e) || vs[0].Equal(e) {
		t.Fatal("variant must be distinct and assignable to parent")
	}
	if _, err := Enum("State", EnumVariant{Name: "A"}, EnumVariant{Name: "A"}); err == nil {
		t.Fatal("duplicate accepted")
	}
	if _, err := EnumValue(vs[0], nil); err == nil {
		t.Fatal("payload variant accepted without payload")
	}
	p := StringValue("wait")
	if _, err := EnumValue(vs[0], &p); err != nil {
		t.Fatal(err)
	}
}
