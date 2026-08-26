package ast

import "github.com/skmohammadali786/gogo/internal/source"

type Node interface{ node() }
type Expr interface {
	Node
	expr()
}
type Stmt interface {
	Node
	stmt()
}

type File struct {
	Span       source.Span
	Statements []Stmt
}

func (File) node() {}

type Identifier struct {
	Span source.Span
	Name string
}

func (Identifier) node() {}
func (Identifier) expr() {}

type Literal struct {
	Span source.Span
	Text string
}

func (Literal) node() {}
func (Literal) expr() {}

type UnaryExpr struct {
	Span     source.Span
	Operator string
	Operand  Expr
}

func (UnaryExpr) node() {}
func (UnaryExpr) expr() {}

type BinaryExpr struct {
	Span     source.Span
	Left     Expr
	Operator string
	Right    Expr
}

func (BinaryExpr) node() {}
func (BinaryExpr) expr() {}

type CallExpr struct {
	Span          source.Span
	Callee        Expr
	TypeArguments []TypeRef
	Arguments     []Expr
}

func (CallExpr) node() {}
func (CallExpr) expr() {}

type FunctionDecl struct {
	Span           source.Span
	Name           Identifier
	TypeParams     []GenericParam
	Parameters     []Identifier
	ParameterTypes []*TypeRef
	ReturnType     *TypeRef
	Body           BlockStmt
}

func (FunctionDecl) node() {}
func (FunctionDecl) stmt() {}

type IfStmt struct {
	Span      source.Span
	Condition Expr
	Then      BlockStmt
	Else      *BlockStmt
}

func (IfStmt) node() {}
func (IfStmt) stmt() {}

type BlockStmt struct {
	Span       source.Span
	Statements []Stmt
}

func (BlockStmt) node() {}
func (BlockStmt) stmt() {}

type ExprStmt struct {
	Span       source.Span
	Expression Expr
}

func (ExprStmt) node() {}
func (ExprStmt) stmt() {}

type VariableDecl struct {
	Span  source.Span
	Name  Identifier
	Type  *TypeRef
	Value Expr
	// Mutable distinguishes `variable` from the existing `constant` surface
	// declaration without introducing a second declaration AST.
	Mutable bool
}

func (VariableDecl) node() {}
func (VariableDecl) stmt() {}

type ReturnStmt struct {
	Span  source.Span
	Value Expr
}

func (ReturnStmt) node() {}
func (ReturnStmt) stmt() {}

func SpanOf(n Node) source.Span {
	switch v := n.(type) {
	case File:
		return v.Span
	case Identifier:
		return v.Span
	case Literal:
		return v.Span
	case UnaryExpr:
		return v.Span
	case BinaryExpr:
		return v.Span
	case AssignmentExpr:
		return v.Span
	case ConditionalExpr:
		return v.Span
	case CallExpr:
		return v.Span
	case FunctionDecl:
		return v.Span
	case IfStmt:
		return v.Span
	case BlockStmt:
		return v.Span
	case ExprStmt:
		return v.Span
	case VariableDecl:
		return v.Span
	case ReturnStmt:
		return v.Span
	case ImportDecl:
		return v.Span
	case ComponentDecl:
		return v.Span
	case TypeAliasDecl:
		return v.Span
	case EnumDecl:
		return v.Span
	case InterfaceDecl:
		return v.Span
	case ArrayExpr:
		return v.Span
	case ObjectExpr:
		return v.Span
	case MemberExpr:
		return v.Span
	case IndexExpr:
		return v.Span
	default:
		return source.Span{}
	}
}
