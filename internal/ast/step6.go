package ast

import (
	"github.com/skmohammadali786/gogo/internal/source"
	"github.com/skmohammadali786/gogo/internal/types"
)

type TypeRef struct {
	Span  source.Span
	Name  string
	Array bool
	// Union and Intersection retain infix type expressions while keeping one
	// semantic AST node and one compiler ResolveType path.
	Union        []TypeRef
	Intersection []TypeRef
	// Arguments and Fields retain surface annotation structure until the
	// compiler resolves it through the canonical types package.
	Arguments []TypeRef
	Fields    []TypeFieldRef
	// Canonical is used only by compiler-created declarations (not parser
	// output) to retain named session types in the one resolver path.
	Canonical *types.Type
}

type TypeFieldRef struct {
	Span     source.Span
	Name     string
	Type     TypeRef
	Optional bool
	Readonly bool
}

// TypeAliasDecl binds a source name to an already canonical type; it does not
// introduce a second type identity.
type TypeAliasDecl struct {
	Span source.Span
	Name Identifier
	Type TypeRef
}

func (TypeAliasDecl) node() {}
func (TypeAliasDecl) stmt() {}

// EnumDecl and InterfaceDecl retain declarations in the one semantic AST;
// their members resolve through the existing canonical type resolver.
type EnumDecl struct {
	Span     source.Span
	Name     Identifier
	Variants []EnumVariantDecl
}
type EnumVariantDecl struct {
	Span    source.Span
	Name    Identifier
	Payload *TypeRef
}

func (EnumDecl) node() {}
func (EnumDecl) stmt() {}

type InterfaceDecl struct {
	Span       source.Span
	Name       Identifier
	Extends    []Identifier
	Properties []TypeFieldRef
}

func (InterfaceDecl) node() {}
func (InterfaceDecl) stmt() {}

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
