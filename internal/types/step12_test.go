package types

import "testing"

func TestStep12OptionalSemantics(t *testing.T) {
	opt := Optional(String)
	if opt.Equal(String) || String.AssignableTo(opt) || opt.AssignableTo(String) {
		t.Fatalf("optional relationship broken")
	}
	if !Optional(Optional(String)).ElementOrPanic().Equal(Optional(String)) {
		t.Fatalf("nested optional not explicit")
	}
	if Optional(String).Equal(Optional(Number)) {
		t.Fatalf("different optional payloads equal")
	}
}

func TestStep12UnionNormalizationEqualityAssignability(t *testing.T) {
	a := MustUnion(String, Number, String)
	b := MustUnion(Number, String)
	if !a.Equal(b) || a.String() != "number | string" {
		t.Fatalf("bad union normalization: %s vs %s", a, b)
	}
	if !String.AssignableTo(a) || !Literal(String, "\"ok\"").AssignableTo(a) {
		t.Fatalf("concrete value should satisfy union")
	}
	if a.AssignableTo(String) {
		t.Fatalf("union should not satisfy target unless every member does")
	}
	if !MustUnion(Literal(String, "\"a\""), Literal(String, "\"b\"")).AssignableTo(String) {
		t.Fatalf("literal union should satisfy primitive when all members do")
	}
}

func TestStep12IntersectionObjectsAndConflicts(t *testing.T) {
	user := MustObject(Field{Name: "name", Type: String})
	serial := MustObject(Field{Name: "id", Type: String, Readonly: true})
	both := MustIntersection(user, serial)
	want := MustObject(Field{Name: "name", Type: String}, Field{Name: "id", Type: String, Readonly: true})
	if !both.Equal(want) || !want.AssignableTo(user) || !want.AssignableTo(serial) {
		t.Fatalf("object intersection did not combine structurally: %s", both)
	}
	conflict := MustIntersection(MustObject(Field{Name: "id", Type: String}), MustObject(Field{Name: "id", Type: Number}))
	if MustObject(Field{Name: "id", Type: String}).AssignableTo(conflict) {
		t.Fatalf("conflicting intersection should be unassignable")
	}
	if !MustIntersection(String, String).Equal(String) {
		t.Fatalf("duplicate intersection should collapse")
	}
}

func TestStep12ResultAndRuntimeValues(t *testing.T) {
	err := MustObject(Field{Name: "message", Type: String})
	res := Result(String, err)
	if !res.Equal(Result(String, err)) || res.Equal(Result(Number, err)) {
		t.Fatalf("result equality broken")
	}
	if !Result(Literal(String, "\"ok\""), err).AssignableTo(res) {
		t.Fatalf("result assignability broken")
	}
	if _, e := OkValue(res, StringValue("done")); e != nil {
		t.Fatal(e)
	}
	errVal, _ := RecordValue(err, map[string]Value{"message": StringValue("bad")})
	if _, e := ErrValue(res, errVal); e != nil {
		t.Fatal(e)
	}
	if _, e := OkValue(res, NumberValue("1")); e == nil {
		t.Fatalf("invalid ok payload accepted")
	}
	opt := Optional(Array(MustUnion(String, Number)))
	arr, _ := ArrayValue(Array(MustUnion(String, Number)), []Value{StringValue("a"), NumberValue("1")})
	if _, e := OptionalPresent(opt, arr); e != nil {
		t.Fatal(e)
	}
	abs, _ := OptionalAbsent(opt)
	if abs.Data().(OptionalData).Present {
		t.Fatalf("absent optional became present")
	}
}

func TestStep12ObjectPropertiesWithUnionIntersectionResult(t *testing.T) {
	state := MustObject(
		Field{Name: "value", Type: MustUnion(String, Number), Optional: true},
		Field{Name: "both", Type: MustIntersection(MustObject(Field{Name: "a", Type: String}), MustObject(Field{Name: "b", Type: Number}))},
		Field{Name: "load", Type: Result(String, MustUnion(String, MustObject(Field{Name: "message", Type: String})))},
	)
	if state.Kind() != RecordKind {
		t.Fatalf("bad state")
	}
}

func (t Type) ElementOrPanic() Type {
	e, ok := t.Element()
	if !ok {
		panic("no element")
	}
	return e
}

func TestStep12NormalizationRemovesRedundantLiteralMembers(t *testing.T) {
	if !MustUnion(String, Literal(String, "\"ready\"")).Equal(String) {
		t.Fatalf("string | string literal should normalize to string")
	}
	if !MustIntersection(String, Literal(String, "\"ready\"")).Equal(Literal(String, "\"ready\"")) {
		t.Fatalf("string & string literal should normalize to the literal")
	}
}

func TestStep12RuntimeIntersectionRequiresEveryObjectMember(t *testing.T) {
	left := MustObject(Field{Name: "name", Type: String})
	right := MustObject(Field{Name: "id", Type: Number})
	intersection := Type{kind: IntersectionKind, members: []Type{left, right}}
	partial, err := RecordValue(left, map[string]Value{"name": StringValue("Ada")})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := IntersectionValue(intersection, partial); err == nil {
		t.Fatalf("intersection runtime value accepted payload satisfying only one member")
	}
	completeType := MustObject(Field{Name: "name", Type: String}, Field{Name: "id", Type: Number})
	complete, err := RecordValue(completeType, map[string]Value{"name": StringValue("Ada"), "id": NumberValue("1")})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := IntersectionValue(intersection, complete); err != nil {
		t.Fatalf("intersection runtime value rejected complete payload: %v", err)
	}
}

func TestStep12IntersectionPreservesCompatibleIndexSignatures(t *testing.T) {
	indexed := MustRecordWithIndex(IndexSignature{Key: String, Value: String, Readonly: true})
	named := MustObject(Field{Name: "name", Type: String})
	combined := MustIntersection(indexed, named)
	idx, ok := combined.IndexSignature()
	if !ok || !idx.Readonly || !idx.Value.Equal(String) {
		t.Fatalf("intersection lost compatible index signature: %s", combined)
	}
	conflict := MustIntersection(indexed, MustObject(Field{Name: "count", Type: Number}))
	if MustObject(Field{Name: "count", Type: Number}).AssignableTo(conflict) {
		t.Fatalf("index signature conflict should not be assignable")
	}
}
