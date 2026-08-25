package ast

import (
	"testing"

	"github.com/skmohammadali786/gogo/internal/source"
)

func TestCollectionNodesExposeSpans(t *testing.T) {
	span := source.Span{Start: source.Position{Offset: 0, Line: 1, Column: 1}, End: source.Position{Offset: 3, Line: 1, Column: 4}}
	array := ArrayExpr{Span: span}
	object := ObjectExpr{Span: span}
	member := MemberExpr{Span: span}
	index := IndexExpr{Span: span}
	if SpanOf(array) != span || SpanOf(object) != span || SpanOf(member) != span || SpanOf(index) != span {
		t.Fatal("collection AST nodes must preserve their spans")
	}
}
