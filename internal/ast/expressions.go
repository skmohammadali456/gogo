package ast

import "github.com/skmohammadali786/gogo/internal/source"

// AssignmentExpr represents simple and compound assignment expressions.
type AssignmentExpr struct {
	Span     source.Span
	Left     Expr
	Operator string
	Right    Expr
}

func (AssignmentExpr) node() {}
func (AssignmentExpr) expr() {}

// ConditionalExpr represents condition ? whenTrue : whenFalse.
type ConditionalExpr struct {
	Span      source.Span
	Condition Expr
	WhenTrue  Expr
	WhenFalse Expr
}

func (ConditionalExpr) node() {}
func (ConditionalExpr) expr() {}
