package compiler

import (
	"github.com/skmohammadali786/gogo/internal/ast"
	"github.com/skmohammadali786/gogo/internal/diagnostics"
	"github.com/skmohammadali786/gogo/internal/types"
)

// checkTypes is intentionally limited to Step 10 declarations and signatures.
// Name resolution, inference, control-flow analysis, and function checking are
// later roadmap work; this verifier never pretends those features exist.
func (s *Session) checkTypes(file ast.File) {
	bindings := make(map[string]binding)
	aliases := make(map[string]ast.TypeRef)
	for _, stmt := range file.Statements {
		if a, ok := stmt.(ast.TypeAliasDecl); ok {
			if _, exists := aliases[a.Name.Name]; exists {
				s.Diagnostics.Add(diagnostics.Diagnostic{Severity: diagnostics.Error, Code: "G3004", Message: "This type alias is declared more than once.", Hint: "Use a unique type alias name.", Span: a.Name.Span})
				continue
			}
			aliases[a.Name.Name] = a.Type
		}
	}
	for _, stmt := range file.Statements {
		s.checkStatementTypes(stmt, bindings, aliases)
	}
}

type binding struct {
	typ     types.Type
	mutable bool
}

func (s *Session) checkStatementTypes(stmt ast.Stmt, bindings map[string]binding, aliases map[string]ast.TypeRef) {
	switch n := stmt.(type) {
	case ast.VariableDecl:
		var target types.Type
		if n.Type != nil {
			var ok bool
			target, ok = s.resolveAnnotation(*n.Type, aliases)
			if !ok {
				return
			}
		} else if inferred, known := expressionType(n.Value); known {
			target = normalizeLiteral(inferred)
		}
		if actual, known := expressionType(n.Value); known && target.Kind() != types.Invalid && !actual.AssignableTo(target) {
			s.typeError("G3002", n.Value, "This value is not assignable to the declared type.", "Use a value with the declared canonical type.")
		}
		if target.Kind() != types.Invalid {
			bindings[n.Name.Name] = binding{typ: target, mutable: n.Mutable}
		}
	case ast.ExprStmt:
		s.checkAssignment(n.Expression, bindings)
	case ast.FunctionDecl:
		for _, t := range n.ParameterTypes {
			if t != nil {
				s.resolveAnnotation(*t, aliases)
			}
		}
		if n.ReturnType != nil {
			s.resolveAnnotation(*n.ReturnType, aliases)
		}
	case ast.BlockStmt:
		for _, child := range n.Statements {
			s.checkStatementTypes(child, bindings, aliases)
		}
	}
}
func (s *Session) checkAssignment(expr ast.Expr, bindings map[string]binding) {
	assign, ok := expr.(ast.AssignmentExpr)
	if !ok {
		return
	}
	name, ok := assign.Left.(ast.Identifier)
	if !ok {
		return
	}
	b, known := bindings[name.Name]
	if !known {
		return
	}
	if !b.mutable {
		s.typeError("G3003", assign.Left, "This binding is immutable and cannot be assigned again.", "Declare a variable when reassignment is required.")
		return
	}
	if actual, known := expressionType(assign.Right); known && !actual.AssignableTo(b.typ) {
		s.typeError("G3002", assign.Right, "This value is not assignable to the declared type.", "Use a value with the declared canonical type.")
	}
}
func (s *Session) resolveAnnotation(ref ast.TypeRef, aliases map[string]ast.TypeRef) (types.Type, bool) {
	t, err := resolveType(ref, aliases, map[string]bool{})
	if err != nil {
		s.Diagnostics.Add(diagnostics.Diagnostic{Severity: diagnostics.Error, Code: "G3001", Message: "This type annotation is not a supported canonical GOGO type.", Hint: err.Error(), Span: ref.Span})
		return types.Type{}, false
	}
	return t, true
}
func (s *Session) typeError(code string, expr ast.Expr, message, hint string) {
	s.Diagnostics.Add(diagnostics.Diagnostic{Severity: diagnostics.Error, Code: code, Message: message, Hint: hint, Span: ast.SpanOf(expr)})
}
func expressionType(e ast.Expr) (types.Type, bool) {
	switch x := e.(type) {
	case ast.Literal:
		return literalType(x), true
	case ast.ArrayExpr:
		if len(x.Items) == 0 {
			return types.Type{}, false
		}
		first, ok := expressionType(x.Items[0])
		if !ok {
			return types.Type{}, false
		}
		base, _ := first.LiteralBase()
		if first.Kind() == types.LiteralKind {
			first = primitive(base)
		}
		for _, item := range x.Items[1:] {
			next, known := expressionType(item)
			if !known || !next.AssignableTo(first) {
				return types.Type{}, false
			}
		}
		return types.Array(first), true
	case ast.ObjectExpr:
		fields := make([]types.Field, len(x.Properties))
		for i, p := range x.Properties {
			pt, ok := expressionType(p.Value)
			if !ok {
				return types.Type{}, false
			}
			if b, is := pt.LiteralBase(); is {
				pt = primitive(b)
			}
			fields[i] = types.Field{Name: p.Key, Type: pt}
		}
		t, err := types.Record(fields...)
		return t, err == nil
	}
	return types.Type{}, false
}
func primitive(k types.Kind) types.Type {
	switch k {
	case types.StringKind:
		return types.String
	case types.NumberKind:
		return types.Number
	case types.BooleanKind:
		return types.Boolean
	case types.BigIntKind:
		return types.BigInt
	case types.BytesKind:
		return types.Bytes
	}
	return types.Type{}
}
func normalizeLiteral(t types.Type) types.Type {
	if base, ok := t.LiteralBase(); ok {
		return primitive(base)
	}
	return t
}
