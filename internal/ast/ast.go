package ast

import "github.com/skmohammadali786/gogo/internal/source"

type Node interface{ node() }
type Expr interface{ Node; expr() }
type Stmt interface{ Node; stmt() }

type File struct { Span source.Span; Statements []Stmt }
func (File) node() {}

type Identifier struct { Span source.Span; Name string }
func (Identifier) node() {}; func (Identifier) expr() {}

type Literal struct { Span source.Span; Text string }
func (Literal) node() {}; func (Literal) expr() {}

type UnaryExpr struct { Span source.Span; Operator string; Operand Expr }
func (UnaryExpr) node() {}; func (UnaryExpr) expr() {}

type BinaryExpr struct { Span source.Span; Left Expr; Operator string; Right Expr }
func (BinaryExpr) node() {}; func (BinaryExpr) expr() {}

type CallExpr struct { Span source.Span; Callee Expr; Arguments []Expr }
func (CallExpr) node() {}; func (CallExpr) expr() {}

type BlockStmt struct { Span source.Span; Statements []Stmt }
func (BlockStmt) node() {}; func (BlockStmt) stmt() {}

type ExprStmt struct { Span source.Span; Expression Expr }
func (ExprStmt) node() {}; func (ExprStmt) stmt() {}

type VariableDecl struct { Span source.Span; Name Identifier; Value Expr }
func (VariableDecl) node() {}; func (VariableDecl) stmt() {}

type ReturnStmt struct { Span source.Span; Value Expr }
func (ReturnStmt) node() {}; func (ReturnStmt) stmt() {}

func SpanOf(n Node) source.Span { switch v := n.(type) { case File: return v.Span; case Identifier: return v.Span; case Literal: return v.Span; case UnaryExpr: return v.Span; case BinaryExpr: return v.Span; case CallExpr: return v.Span; case BlockStmt: return v.Span; case ExprStmt: return v.Span; case VariableDecl: return v.Span; case ReturnStmt: return v.Span; default: return source.Span{} } }
