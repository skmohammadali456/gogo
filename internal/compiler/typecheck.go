package compiler

import (
	"github.com/skmohammadali786/gogo/internal/ast"
	"github.com/skmohammadali786/gogo/internal/diagnostics"
	"github.com/skmohammadali786/gogo/internal/source"
	"github.com/skmohammadali786/gogo/internal/types"
)

func (s *Session) checkTypes(file ast.File) {
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
	c := checker{s: s, aliases: aliases, env: map[string]binding{}, funcs: map[string]functionSig{}}
	c.collectFunctions(file.Statements)
	c.checkStatements(file.Statements, c.env, types.Type{})
}

type binding struct {
	declared, inferred, current types.Type
	mutable                     bool
}
type functionSig struct {
	params []types.Type
	ret    types.Type
}
type checker struct {
	s       *Session
	aliases map[string]ast.TypeRef
	env     map[string]binding
	funcs   map[string]functionSig
}

func (c *checker) collectFunctions(stmts []ast.Stmt) {
	for _, st := range stmts {
		if fn, ok := st.(ast.FunctionDecl); ok {
			sig := functionSig{params: make([]types.Type, len(fn.Parameters))}
			for i, tr := range fn.ParameterTypes {
				if tr != nil {
					sig.params[i], _ = c.s.resolveAnnotation(*tr, c.aliases)
				}
			}
			if fn.ReturnType != nil {
				sig.ret, _ = c.s.resolveAnnotation(*fn.ReturnType, c.aliases)
			}
			c.funcs[fn.Name.Name] = sig
		}
	}
}
func (c *checker) checkStatements(stmts []ast.Stmt, env map[string]binding, ret types.Type) bool {
	returned := false
	for _, stmt := range stmts {
		if returned {
			continue
		}
		if c.checkStatement(stmt, env, ret) {
			returned = true
		}
	}
	return returned
}
func (c *checker) checkStatement(stmt ast.Stmt, env map[string]binding, ret types.Type) bool {
	switch n := stmt.(type) {
	case ast.TypeAliasDecl:
	case ast.VariableDecl:
		var target types.Type
		hasTarget := false
		if n.Type != nil {
			if t, ok := c.s.resolveAnnotation(*n.Type, c.aliases); ok {
				target = t
				hasTarget = true
			}
		}
		actual, known := c.infer(n.Value, target, hasTarget, env)
		if !known && !hasTarget {
			c.diag("G3007", ast.SpanOf(n.Value), "I could not infer a type for this value.", "Add an explicit annotation or simplify the expression.")
		}
		if hasTarget {
			if known && !compatibleInit(actual, target) {
				c.s.typeError("G3002", n.Value, "This value is not assignable to the declared type.", "Use a value with the declared canonical type.")
			}
		} else if known {
			target = normalizeLiteral(actual)
			hasTarget = true
		}
		if hasTarget {
			env[n.Name.Name] = binding{declared: target, inferred: actual, current: target, mutable: n.Mutable}
		}
	case ast.ExprStmt:
		c.checkAssignment(n.Expression, env)
	case ast.FunctionDecl:
		child := copyEnv(env)
		sig := c.funcs[n.Name.Name]
		for i, p := range n.Parameters {
			if i < len(sig.params) && sig.params[i].Kind() != types.Invalid {
				child[p.Name] = binding{declared: sig.params[i], current: sig.params[i]}
			} else {
				c.diag("G3007", p.Span, "I could not infer this parameter type.", "Add a parameter annotation.")
			}
		}
		c.checkStatements(n.Body.Statements, child, sig.ret)
	case ast.BlockStmt:
		c.checkStatements(n.Statements, copyEnv(env), ret)
	case ast.IfStmt:
		if !c.truthy(n.Condition, env) {
			c.diag("G3008", ast.SpanOf(n.Condition), "This condition does not have defined GOGO truthiness.", "Use Boolean, Optional, Result, or a provable union/property/discriminant check.")
		}
		thenEnv := copyEnv(env)
		elseEnv := copyEnv(env)
		c.applyNarrow(n.Condition, thenEnv, elseEnv)
		thenRet := c.checkStatements(n.Then.Statements, thenEnv, ret)
		elseRet := false
		if n.Else != nil {
			elseRet = c.checkStatements(n.Else.Statements, elseEnv, ret)
		}
		mergeEnv(env, thenEnv, elseEnv, thenRet, elseRet, n.Else != nil)
		return thenRet && n.Else != nil && elseRet
	case ast.ReturnStmt:
		actual, known := c.infer(n.Value, ret, ret.Kind() != types.Invalid, env)
		if ret.Kind() != types.Invalid && known && !compatibleInit(actual, ret) && !c.legacyUnnarrowedIdentifier(n.Value, env) {
			c.s.typeError("G3002", n.Value, "This return value is not assignable to the declared return type.", "Return a value with the declared canonical type.")
		}
		return true
	}
	return false
}

// Step 10 accepted un-narrowed identifier returns as a signature boundary.
// Keep that compatibility while still checking identifiers whose type has been
// refined by a local Step 13 proof.
func (c *checker) legacyUnnarrowedIdentifier(e ast.Expr, env map[string]binding) bool {
	id, ok := e.(ast.Identifier)
	if !ok {
		return false
	}
	b, ok := env[id.Name]
	return ok && b.current.Equal(b.declared)
}
func (c *checker) infer(e ast.Expr, ctx types.Type, hasCtx bool, env map[string]binding) (types.Type, bool) {
	switch x := e.(type) {
	case ast.Literal:
		t := literalType(x)
		if hasCtx && t.AssignableTo(ctx) {
			return t, true
		}
		return t, true
	case ast.Identifier:
		if x.Name == "true" || x.Name == "false" {
			return types.Literal(types.Boolean, x.Name), true
		}
		if b, ok := env[x.Name]; ok {
			return b.current, true
		}
		return types.Type{}, false
	case ast.ArrayExpr:
		if hasCtx && ctx.Kind() == types.ArrayKind {
			et, _ := ctx.Element()
			for _, it := range x.Items {
				at, ok := c.infer(it, et, true, env)
				if !ok || !at.AssignableTo(et) {
					return types.Array(et), true
				}
			}
			return types.Array(et), true
		}
		var ms []types.Type
		for _, it := range x.Items {
			t, ok := c.infer(it, types.Type{}, false, env)
			if !ok {
				return types.Type{}, false
			}
			ms = append(ms, normalizeLiteral(t))
		}
		if len(ms) == 0 {
			return types.Type{}, false
		}
		u, _ := types.Union(ms...)
		return types.Array(u), true
	case ast.ObjectExpr:
		return c.inferObject(x, ctx, hasCtx, env)
	case ast.MemberExpr:
		ot, ok := c.infer(x.Object, types.Type{}, false, env)
		if !ok {
			return types.Type{}, false
		}
		// A member is readable without a preceding proof only when every
		// possible union member has that field.  fieldOf intentionally has
		// broader use for contextual object literals and narrowing predicates.
		if f, ok := fieldOfAll(ot, x.Name); ok {
			if x.Optional {
				return types.Optional(f.Type), true
			}
			return f.Type, true
		}
		return types.Type{}, false
	case ast.BinaryExpr:
		return c.inferBinary(x, env)
	case ast.ConditionalExpr:
		tt, tok := c.infer(x.WhenTrue, ctx, hasCtx, env)
		ft, fok := c.infer(x.WhenFalse, ctx, hasCtx, env)
		if tok && fok {
			u, _ := types.Union(normalizeLiteral(tt), normalizeLiteral(ft))
			return u, true
		}
	case ast.CallExpr:
		if id, ok := x.Callee.(ast.Identifier); ok {
			if sig, exists := c.funcs[id.Name]; exists {
				for i, a := range x.Arguments {
					if i < len(sig.params) {
						at, aok := c.infer(a, sig.params[i], true, env)
						if aok && !at.AssignableTo(sig.params[i]) {
							c.s.typeError("G3002", a, "This argument is not assignable to the parameter type.", "Pass a value with the declared parameter type.")
						}
					}
				}
				return sig.ret, sig.ret.Kind() != types.Invalid
			}
		}
	}
	return expressionType(e)
}
func (c *checker) inferObject(x ast.ObjectExpr, ctx types.Type, hasCtx bool, env map[string]binding) (types.Type, bool) {
	fields := make([]types.Field, 0, len(x.Properties))
	for _, p := range x.Properties {
		var ft types.Type
		hc := false
		if hasCtx {
			if f, ok := fieldOf(ctx, p.Key); ok {
				ft = f.Type
				hc = true
			}
		}
		pt, ok := c.infer(p.Value, ft, hc, env)
		if !ok {
			return types.Type{}, false
		}
		if !hc {
			pt = normalizeLiteral(pt)
		}
		fields = append(fields, types.Field{Name: p.Key, Type: pt})
	}
	t, err := types.Object(fields...)
	return t, err == nil
}
func (c *checker) inferBinary(x ast.BinaryExpr, env map[string]binding) (types.Type, bool) {
	lt, lok := c.infer(x.Left, types.Type{}, false, env)
	rt, rok := c.infer(x.Right, types.Type{}, false, env)
	switch x.Operator {
	case "==", "===", "!=", "!==", "&&", "||", "<", "<=", ">", ">=":
		return types.Boolean, true
	case "+", "-", "*", "/", "%":
		if lok && rok && lt.AssignableTo(types.Number) && rt.AssignableTo(types.Number) {
			return types.Number, true
		}
	}
	return types.Type{}, false
}
func (c *checker) truthy(e ast.Expr, env map[string]binding) bool {
	if m, ok := e.(ast.MemberExpr); ok && propertyCondition(m, env) {
		return true
	}
	if t, ok := c.infer(e, types.Type{}, false, env); ok {
		switch t.Kind() {
		case types.BooleanKind, types.OptionalKind, types.ResultKind:
			return true
		}
	}
	return false
}

func (c *checker) discriminantCondition(e ast.Expr, env map[string]binding) bool {
	b, ok := e.(ast.BinaryExpr)
	if !ok {
		return false
	}
	m, _, ok := memberLiteralCheck(b.Left, b.Right)
	if !ok {
		return false
	}
	id, ok := rootIdent(m)
	if !ok {
		return false
	}
	bind, ok := env[id]
	return ok && bind.current.Kind() == types.UnionKind && len(bind.current.Members()) > 1
}

func (c *checker) applyNarrow(e ast.Expr, thenEnv, elseEnv map[string]binding) {
	switch x := e.(type) {
	case ast.Identifier:
		if b, ok := thenEnv[x.Name]; ok {
			if b.current.Kind() == types.OptionalKind {
				el, _ := b.current.Element()
				b.current = el
				thenEnv[x.Name] = b
			}
			if b.current.Kind() == types.ResultKind {
				okT, _ := b.current.Ok()
				errT, _ := b.current.Err()
				b.current = okT
				thenEnv[x.Name] = b
				eb := elseEnv[x.Name]
				eb.current = errT
				elseEnv[x.Name] = eb
			}
		}
	case ast.UnaryExpr:
		if x.Operator == "!" {
			c.applyNarrow(x.Operand, elseEnv, thenEnv)
		}
	case ast.BinaryExpr:
		c.applyBinaryNarrow(x, thenEnv, elseEnv)
	case ast.MemberExpr:
		if id, ok := rootIdent(x); ok {
			narrowProperty(thenEnv, id, x.Name, true)
			narrowProperty(elseEnv, id, x.Name, false)
		}
	}
}
func (c *checker) applyBinaryNarrow(x ast.BinaryExpr, thenEnv, elseEnv map[string]binding) {
	if x.Operator != "==" && x.Operator != "===" && x.Operator != "!=" && x.Operator != "!==" {
		return
	}
	pos := x.Operator == "==" || x.Operator == "==="
	targetEnv := thenEnv
	otherEnv := elseEnv
	if !pos {
		targetEnv, otherEnv = elseEnv, thenEnv
	}
	if m, lit, ok := memberLiteralCheck(x.Left, x.Right); ok {
		if id, ok := rootIdent(m); ok {
			narrowDiscriminant(targetEnv, id, m.Name, literalType(lit))
			narrowDiscriminantNot(otherEnv, id, m.Name, literalType(lit))
		}
	}
	if id, lit, ok := identLiteralCheck(x.Left, x.Right); ok {
		narrowUnionLiteral(targetEnv, id, literalType(lit))
		_ = otherEnv
	}
}
func (c *checker) checkAssignment(expr ast.Expr, env map[string]binding) {
	a, ok := expr.(ast.AssignmentExpr)
	if !ok {
		return
	}
	id, ok := a.Left.(ast.Identifier)
	if !ok {
		return
	}
	b, known := env[id.Name]
	if !known {
		return
	}
	if !b.mutable {
		c.s.typeError("G3003", a.Left, "This binding is immutable and cannot be assigned again.", "Declare a variable when reassignment is required.")
		return
	}
	at, ok := c.infer(a.Right, b.declared, true, env)
	if ok && !compatibleInit(at, b.declared) {
		c.s.typeError("G3002", a.Right, "This value is not assignable to the declared type.", "Use a value with the declared canonical type.")
	}
	// Assignment can replace a value narrowed by an earlier proof.  The
	// assignment's declared target is the only type known after this point.
	b.current = b.declared
	env[id.Name] = b
}
func (c *checker) diag(code string, sp source.Span, msg, hint string) {
	c.s.Diagnostics.Add(diagnostics.Diagnostic{Severity: diagnostics.Error, Code: code, Message: msg, Hint: hint, Span: sp})
}

func copyEnv(in map[string]binding) map[string]binding {
	out := make(map[string]binding, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
func mergeEnv(base, thenEnv, elseEnv map[string]binding, thenRet, elseRet, hasElse bool) {
	for name, b := range base {
		if thenRet && hasElse && !elseRet {
			b.current = elseEnv[name].current
		} else if elseRet && !thenRet {
			b.current = thenEnv[name].current
		} else {
			b.current = b.declared
		}
		base[name] = b
	}
}
func fieldOf(t types.Type, name string) (types.Field, bool) {
	t = types.Normalize(t)
	if t.Kind() == types.UnionKind {
		var found []types.Type
		opt := false
		ro := false
		for _, m := range t.Members() {
			if f, ok := fieldOf(m, name); ok {
				found = append(found, f.Type)
				opt = opt || f.Optional
				ro = ro || f.Readonly
			}
		}
		if len(found) == 0 {
			return types.Field{}, false
		}
		u, _ := types.Union(found...)
		return types.Field{Name: name, Type: u, Optional: opt, Readonly: ro}, true
	}
	if t.Kind() == types.IntersectionKind {
		for _, m := range t.Members() {
			if f, ok := fieldOf(m, name); ok {
				return f, true
			}
		}
	}
	if t.Kind() != types.RecordKind {
		return types.Field{}, false
	}
	for _, f := range t.Fields() {
		if f.Name == name {
			return f, true
		}
	}
	return types.Field{}, false
}

// fieldOfAll returns a field only when it is safe to read on every possible
// union member.  This deliberately differs from fieldOf, which is also used
// to discover contextual fields and to select variants during a proof.
func fieldOfAll(t types.Type, name string) (types.Field, bool) {
	t = types.Normalize(t)
	if t.Kind() != types.UnionKind {
		return fieldOf(t, name)
	}
	var found []types.Type
	optional, readonly := false, false
	for _, member := range t.Members() {
		field, ok := fieldOf(member, name)
		if !ok {
			return types.Field{}, false
		}
		found = append(found, field.Type)
		optional = optional || field.Optional
		readonly = readonly || field.Readonly
	}
	joined, err := types.Union(found...)
	if err != nil {
		return types.Field{}, false
	}
	return types.Field{Name: name, Type: joined, Optional: optional, Readonly: readonly}, true
}

func propertyCondition(m ast.MemberExpr, env map[string]binding) bool {
	name, ok := rootIdent(m)
	if !ok {
		return false
	}
	binding, ok := env[name]
	if !ok || binding.current.Kind() != types.UnionKind {
		return false
	}
	has, missing := false, false
	for _, member := range binding.current.Members() {
		if _, ok := fieldOf(member, m.Name); ok {
			has = true
		} else {
			missing = true
		}
	}
	return has && missing
}
func rootIdent(m ast.MemberExpr) (string, bool) {
	switch o := m.Object.(type) {
	case ast.Identifier:
		return o.Name, true
	case ast.MemberExpr:
		return rootIdent(o)
	}
	return "", false
}
func memberLiteralCheck(a, b ast.Expr) (ast.MemberExpr, ast.Literal, bool) {
	if m, ok := a.(ast.MemberExpr); ok {
		if l, ok := b.(ast.Literal); ok {
			return m, l, true
		}
	}
	if m, ok := b.(ast.MemberExpr); ok {
		if l, ok := a.(ast.Literal); ok {
			return m, l, true
		}
	}
	return ast.MemberExpr{}, ast.Literal{}, false
}
func identLiteralCheck(a, b ast.Expr) (string, ast.Literal, bool) {
	if id, ok := a.(ast.Identifier); ok {
		if l, ok := b.(ast.Literal); ok {
			return id.Name, l, true
		}
	}
	if id, ok := b.(ast.Identifier); ok {
		if l, ok := a.(ast.Literal); ok {
			return id.Name, l, true
		}
	}
	return "", ast.Literal{}, false
}
func narrowUnionLiteral(env map[string]binding, name string, lit types.Type) {
	b, ok := env[name]
	if !ok || b.current.Kind() != types.UnionKind {
		return
	}
	var keep []types.Type
	for _, m := range b.current.Members() {
		if lit.AssignableTo(m) || m.AssignableTo(lit) {
			keep = append(keep, m)
		}
	}
	if len(keep) > 0 {
		u, _ := types.Union(keep...)
		b.current = u
		env[name] = b
	}
}
func narrowDiscriminant(env map[string]binding, name, prop string, lit types.Type) {
	b, ok := env[name]
	if !ok || b.current.Kind() != types.UnionKind {
		return
	}
	var keep []types.Type
	for _, m := range b.current.Members() {
		if f, ok := fieldOf(m, prop); ok {
			if lit.Kind() == types.Invalid || lit.AssignableTo(f.Type) || f.Type.AssignableTo(lit) {
				keep = append(keep, m)
			}
		}
	}
	if len(keep) > 0 {
		u, _ := types.Union(keep...)
		b.current = u
		env[name] = b
	}
}
func narrowProperty(env map[string]binding, name, prop string, exists bool) {
	b, ok := env[name]
	if !ok || b.current.Kind() != types.UnionKind {
		return
	}
	var keep []types.Type
	for _, m := range b.current.Members() {
		_, has := fieldOf(m, prop)
		if has == exists {
			keep = append(keep, m)
		}
	}
	if len(keep) > 0 {
		u, _ := types.Union(keep...)
		b.current = u
		env[name] = b
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
		first = normalizeLiteral(first)
		var members []types.Type
		members = append(members, first)
		for _, item := range x.Items[1:] {
			next, known := expressionType(item)
			if !known {
				return types.Type{}, false
			}
			members = append(members, normalizeLiteral(next))
		}
		u, _ := types.Union(members...)
		return types.Array(u), true
	case ast.ObjectExpr:
		fields := make([]types.Field, len(x.Properties))
		for i, p := range x.Properties {
			pt, ok := expressionType(p.Value)
			if !ok {
				return types.Type{}, false
			}
			fields[i] = types.Field{Name: p.Key, Type: normalizeLiteral(pt)}
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

func compatibleInit(actual, target types.Type) bool {
	if actual.AssignableTo(target) {
		return true
	}
	if target.Kind() == types.OptionalKind {
		el, _ := target.Element()
		return actual.AssignableTo(el)
	}
	if target.Kind() == types.ResultKind {
		okT, _ := target.Ok()
		errT, _ := target.Err()
		return actual.AssignableTo(okT) || actual.AssignableTo(errT)
	}
	return false
}

func narrowDiscriminantNot(env map[string]binding, name, prop string, lit types.Type) {
	b, ok := env[name]
	if !ok || b.current.Kind() != types.UnionKind {
		return
	}
	var keep []types.Type
	for _, m := range b.current.Members() {
		if f, ok := fieldOf(m, prop); ok {
			if !(lit.AssignableTo(f.Type) || f.Type.AssignableTo(lit)) {
				keep = append(keep, m)
			}
		} else {
			// A missing property cannot equal the checked literal, so it belongs
			// to the false branch rather than being incorrectly discarded.
			keep = append(keep, m)
		}
	}
	if len(keep) > 0 {
		u, _ := types.Union(keep...)
		b.current = u
		env[name] = b
	}
}
