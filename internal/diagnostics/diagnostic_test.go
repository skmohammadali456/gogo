package diagnostics

import (
	"testing"

	"github.com/skmohammadali786/gogo/internal/source"
)

func TestBagTracksErrors(t *testing.T) {
	var b Bag
	if b.HasErrors() {
		t.Fatal("empty diagnostic bag must not contain errors")
	}

	b.Add(Diagnostic{
		Severity: Error,
		Code:     "G001",
		Message:  "example error",
		Span: source.Span{
			Start: source.Position{Offset: 0, Line: 1, Column: 1},
			End:   source.Position{Offset: 1, Line: 1, Column: 2},
		},
	})
	if !b.HasErrors() {
		t.Fatal("expected diagnostic bag to contain an error")
	}
	if len(b.All()) != 1 {
		t.Fatal("expected one diagnostic")
	}
}
