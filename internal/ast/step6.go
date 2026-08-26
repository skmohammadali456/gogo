package ast

import "github.com/skmohammadali786/gogo/internal/source"

type TypeRef struct {
	Span  source.Span
	Name  string
	Array bool
	// Arguments and Fields retain surface annotation structure until the
	// compiler resolves it through the canonical types package.
	Arguments []TypeRef
	Fields    []TypeFieldRef
}

type TypeFieldRef struct {
	Span source.Span
	Name string
	Type TypeRef
}

func (TypeRef) node() {}

type ImportDecl struct {
	Span  source.Span
	Path  string
	Alias *Identifier
}

func (ImportDecl) node() {}
func (ImportDecl) stmt() {}

type ComponentProp struct {
	Span  source.Span
	Name  string
	Value Expr
}

type ComponentDecl struct {
	Span       source.Span
	Name       Identifier
	Properties []ComponentProp
	Body       BlockStmt
}

func (ComponentDecl) node() {}
func (ComponentDecl) stmt() {}
