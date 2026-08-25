package ast

import "github.com/skmohammadali786/gogo/internal/source"

type ArrayExpr struct {
	Span  source.Span
	Items []Expr
}

func (ArrayExpr) node() {}
func (ArrayExpr) expr() {}

type ObjectProperty struct {
	Span  source.Span
	Key   string
	Value Expr
}

type ObjectExpr struct {
	Span       source.Span
	Properties []ObjectProperty
}

func (ObjectExpr) node() {}
func (ObjectExpr) expr() {}

type MemberExpr struct {
	Span     source.Span
	Object   Expr
	Name     string
	Optional bool
}

func (MemberExpr) node() {}
func (MemberExpr) expr() {}

type IndexExpr struct {
	Span   source.Span
	Object Expr
	Index  Expr
}

func (IndexExpr) node() {}
func (IndexExpr) expr() {}
