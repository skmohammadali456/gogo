package parser

import (
	"github.com/skmohammadali786/gogo/internal/ast"
	"github.com/skmohammadali786/gogo/internal/diagnostics"
	"github.com/skmohammadali786/gogo/internal/grammar"
	"github.com/skmohammadali786/gogo/internal/source"
	"github.com/skmohammadali786/gogo/internal/token"
	"strings"
)

type Parser struct {
	tokens      []token.Token
	pos         int
	vocabulary  grammar.Vocabulary
	diagnostics diagnostics.Bag
}

// Option configures a parser instance.
type Option func(*Parser)

// WithVocabulary selects the active surface vocabulary for this parser.
func WithVocabulary(v grammar.Vocabulary) Option {
	return func(p *Parser) { p.vocabulary = v }
}

func New(tokens []token.Token, opts ...Option) *Parser {
	p := &Parser{tokens: tokens, vocabulary: grammar.DefaultVocabulary()}
	for _, opt := range opts {
		opt(p)
	}
	return p
}
func (p *Parser) Diagnostics() []diagnostics.Diagnostic { return p.diagnostics.All() }

func (p *Parser) ParseFile() ast.File {
	var statements []ast.Stmt
	start := p.current().Span.Start
	for !p.at(token.EOF) {
		if p.at(token.RBrace) {
			p.error("G2034", "I found a closing brace without a matching block.", "Remove this } or add the block that should contain it.")
			p.advance()
			continue
		}
		before := p.pos
		if stmt := p.parseStatement(); stmt != nil {
			statements = append(statements, stmt)
		}
		if p.pos == before {
			p.recoverStatement()
			if p.pos == before && !p.at(token.EOF) {
				p.advance()
			}
		}
	}
	end := start
	if len(statements) > 0 {
		end = ast.SpanOf(statements[len(statements)-1]).End
	}
	return ast.File{Span: source.Span{Start: start, End: end}, Statements: statements}
}

func (p *Parser) parseStatement() ast.Stmt {
	if p.at(token.Invalid) {
		p.error("G2000", "I cannot parse this invalid token.", "Fix the earlier lexical error and try again.")
		p.advance()
		return nil
	}
	if p.atKeyword(grammar.KeywordCreate) {
		return p.parseCreate()
	}
	if p.atKeyword(grammar.KeywordVariable) || p.atKeyword(grammar.KeywordConstant) || p.atKeyword(grammar.KeywordFunction) {
		return p.parseConciseCreate()
	}
	if p.atKeyword(grammar.KeywordImport) {
		return p.parseImport()
	}
	if p.atKeyword(grammar.KeywordReturn) {
		return p.parseReturn()
	}
	if p.atKeyword(grammar.KeywordIf) {
		return p.parseIf()
	}
	if p.at(token.LBrace) {
		return p.parseBlock()
	}
	expr := p.parseExpression(0)
	if expr == nil {
		return nil
	}
	if p.canEndStatement(expr) {
		p.consumeTerminator()
	} else if !p.at(token.EOF) && !p.at(token.RBrace) {
		p.error("G2035", "I found extra tokens after this expression statement.", "Separate statements with semicolons or write a complete declaration/control-flow form for the selected grammar.")
		p.recoverStatement()
	}
	return ast.ExprStmt{Span: ast.SpanOf(expr), Expression: expr}
}

func (p *Parser) parseCreate() ast.Stmt {
	start := p.advance().Span.Start
	if p.atKeyword(grammar.KeywordVariable) {
		return p.parseVariable(start)
	}
	if p.atKeyword(grammar.KeywordFunction) {
		return p.parseFunction(start)
	}
	if p.atKeyword(grammar.KeywordConstant) {
		return p.parseVariable(start)
	}
	if p.atKeyword(grammar.KeywordComponent) {
		return p.parseComponent(start)
	}
	if p.atKeyword(grammar.KeywordType) {
		return p.parseTypeAlias(start)
	}
	p.error("G2001", "I expected a declaration keyword after 'create'.", "Write create variable name as value, create function name(...), or create component Name { ... }.")
	p.recoverStatement()
	return nil
}

func (p *Parser) parseConciseCreate() ast.Stmt {
	start := p.current().Span.Start
	if p.atKeyword(grammar.KeywordFunction) {
		return p.parseFunction(start)
	}
	return p.parseVariable(start)
}

func (p *Parser) parseTypeAlias(start source.Position) ast.Stmt {
	p.advance()
	name := p.expect(token.Identifier, "G3004", "I expected a type alias name.", "Give the alias a unique identifier.")
	if name.Kind == token.Invalid {
		p.recoverStatement()
		return nil
	}
	if !p.atKeyword(grammar.KeywordAs) {
		p.error("G3004", "I expected 'as' after the type alias name.", "Write create type Name as Type.")
		p.recoverStatement()
		return nil
	}
	p.advance()
	typ := p.parseType()
	p.consumeTerminator()
	return ast.TypeAliasDecl{Span: source.Span{Start: start, End: typ.Span.End}, Name: ast.Identifier{Span: name.Span, Name: name.Text}, Type: typ}
}

func (p *Parser) parseVariable(start source.Position) ast.Stmt {
	mutable := !p.atKeyword(grammar.KeywordConstant)
	p.advance()
	name := p.expect(token.Identifier, "G2002", "I expected a variable name.", "Give the variable a name.")
	if name.Kind == token.Invalid {
		p.recoverStatement()
		return nil
	}
	if !p.atKeyword(grammar.KeywordAs) {
		p.error("G2003", "I expected 'as' after the variable name.", "Write create variable name as value.")
		return nil
	}
	p.advance()
	var typ *ast.TypeRef
	if p.looksLikeTypeThenAs() {
		t := p.parseType()
		typ = &t
		p.advance() // consume initializer as
	}
	value := p.parseExpression(0)
	if value == nil {
		p.error("G2004", "I expected a value after as.", "Give the variable an initial value.")
		return nil
	}
	p.consumeTerminator()
	return ast.VariableDecl{Span: source.Span{Start: start, End: ast.SpanOf(value).End}, Name: ast.Identifier{Span: name.Span, Name: name.Text}, Type: typ, Value: value, Mutable: mutable}
}

func (p *Parser) parseFunction(start source.Position) ast.Stmt {
	p.advance()
	name := p.expect(token.Identifier, "G2011", "I expected a function name.", "Give the function a name.")
	if name.Kind == token.Invalid {
		p.recoverStatement()
		return nil
	}
	if !p.at(token.LParen) {
		p.error("G2012", "I expected '(' after the function name.", "Add the function parameter list.")
		p.recoverStatement()
		return nil
	}
	p.advance()
	var params []ast.Identifier
	var paramTypes []*ast.TypeRef
	if !p.at(token.RParen) {
		for {
			param := p.expect(token.Identifier, "G2013", "I expected a parameter name.", "Give each parameter a name.")
			if param.Kind == token.Invalid {
				p.recoverStatement()
				return nil
			}
			params = append(params, ast.Identifier{Span: param.Span, Name: param.Text})
			if p.atKeyword(grammar.KeywordAs) {
				p.advance()
				t := p.parseType()
				paramTypes = append(paramTypes, &t)
			} else {
				paramTypes = append(paramTypes, nil)
			}
			if !p.at(token.Comma) {
				break
			}
			p.advance()
			if p.at(token.RParen) {
				break
			}
		}
	}
	if !p.at(token.RParen) {
		p.error("G2014", "I expected ')' after the function parameters.", "Close the parameter list.")
		p.recoverStatement()
		return nil
	}
	p.advance()
	var returnType *ast.TypeRef
	if p.atKeyword(grammar.KeywordAs) {
		p.advance()
		t := p.parseType()
		returnType = &t
	}
	body, ok := p.parseBlockRequired("G2015", "I expected a function body.", "Add { ... } after the function declaration.")
	if !ok {
		return nil
	}
	return ast.FunctionDecl{Span: source.Span{Start: start, End: body.Span.End}, Name: ast.Identifier{Span: name.Span, Name: name.Text}, Parameters: params, ParameterTypes: paramTypes, ReturnType: returnType, Body: body}
}

func (p *Parser) parseIf() ast.Stmt {
	start := p.advance().Span.Start
	condition := p.parseExpression(0)
	if condition == nil {
		p.error("G2016", "I expected a condition after 'if'.", "Add an expression describing when the block should run.")
		return nil
	}
	then, ok := p.parseBlockRequired("G2017", "I expected a block after the if condition.", "Add { ... } for the conditional body.")
	if !ok {
		return nil
	}
	var elseBlock *ast.BlockStmt
	if p.atKeyword(grammar.KeywordElse) {
		p.advance()
		block, ok := p.parseBlockRequired("G2018", "I expected a block after 'else'.", "Add { ... } for the else body.")
		if !ok {
			return nil
		}
		elseBlock = &block
	}
	end := then.Span.End
	if elseBlock != nil {
		end = elseBlock.Span.End
	}
	return ast.IfStmt{Span: source.Span{Start: start, End: end}, Condition: condition, Then: then, Else: elseBlock}
}

func (p *Parser) parseReturn() ast.Stmt {
	start := p.advance().Span.Start
	if p.at(token.Semicolon) || p.at(token.RBrace) || p.at(token.EOF) {
		p.error("G2005", "I expected a value after 'return'.", "Return a value or use a grammar form that explicitly permits an empty return.")
		return nil
	}
	value := p.parseExpression(0)
	if value == nil {
		p.error("G2005", "I expected a value after 'return'.", "Return a value or use a grammar form that explicitly permits an empty return.")
		return nil
	}
	p.consumeTerminator()
	return ast.ReturnStmt{Span: source.Span{Start: start, End: ast.SpanOf(value).End}, Value: value}
}

func (p *Parser) parseBlockRequired(code, message, hint string) (ast.BlockStmt, bool) {
	if !p.at(token.LBrace) {
		p.error(code, message, hint)
		return ast.BlockStmt{}, false
	}
	return p.parseBlock().(ast.BlockStmt), true
}

func (p *Parser) parseBlock() ast.Stmt {
	start := p.advance().Span.Start
	var statements []ast.Stmt
	for !p.at(token.RBrace) && !p.at(token.EOF) {
		before := p.pos
		if stmt := p.parseStatement(); stmt != nil {
			statements = append(statements, stmt)
		}
		if before == p.pos {
			p.recoverStatement()
			if p.pos == before && !p.at(token.EOF) && !p.at(token.RBrace) {
				p.advance()
			}
		}
	}
	end := p.current().Span.End
	if p.at(token.RBrace) {
		end = p.advance().Span.End
	} else {
		p.error("G2006", "This block is missing its closing brace.", "Add } to close the block.")
	}
	return ast.BlockStmt{Span: source.Span{Start: start, End: end}, Statements: statements}
}

func (p *Parser) parseExpression(minPrec int) ast.Expr {
	left := p.parsePrefix()
	if left == nil {
		return nil
	}
	for {
		op := p.current()
		prec := precedence(op.Kind)
		if prec < minPrec {
			break
		}
		p.advance()
		rightMin := prec + 1
		if isRightAssociative(op.Kind) {
			rightMin = prec
		}
		right := p.parseExpression(rightMin)
		if right == nil {
			p.error("G2007", "I expected an expression after this operator.", "Add a value after the operator.")
			return left
		}
		span := source.Span{Start: ast.SpanOf(left).Start, End: ast.SpanOf(right).End}
		if isAssignmentOperator(op.Kind) {
			if !isAssignable(left) {
				p.error("G2028", "The left side of an assignment must be a variable or writable property.", "Assign to an identifier, object member, or indexed element.")
			}
			left = ast.AssignmentExpr{Span: span, Left: left, Operator: op.Text, Right: right}
		} else {
			left = ast.BinaryExpr{Span: span, Left: left, Operator: op.Text, Right: right}
		}
	}
	if minPrec <= 0 && p.at(token.Question) {
		start := ast.SpanOf(left).Start
		p.advance()
		whenTrue := p.parseExpression(0)
		if whenTrue == nil {
			p.error("G2029", "I expected an expression after '?'.", "Add the value to use when the condition is true.")
			return left
		}
		if !p.at(token.Colon) {
			p.error("G2030", "I expected ':' in this conditional expression.", "Write condition ? whenTrue : whenFalse.")
			return left
		}
		p.advance()
		whenFalse := p.parseExpression(0)
		if whenFalse == nil {
			p.error("G2031", "I expected an expression after ':'.", "Add the value to use when the condition is false.")
			return left
		}
		return ast.ConditionalExpr{Span: source.Span{Start: start, End: ast.SpanOf(whenFalse).End}, Condition: left, WhenTrue: whenTrue, WhenFalse: whenFalse}
	}
	return left
}

func (p *Parser) parsePrefix() ast.Expr {
	t := p.current()
	var expr ast.Expr
	switch t.Kind {
	case token.Identifier:
		p.advance()
		expr = ast.Identifier{Span: t.Span, Name: t.Text}
	case token.Number, token.String:
		p.advance()
		expr = ast.Literal{Span: t.Span, Text: t.Text}
	case token.Bang, token.Plus, token.Minus, token.Tilde, token.PlusPlus, token.MinusMinus:
		p.advance()
		operand := p.parseExpression(11)
		if operand == nil {
			p.error("G2008", "I expected an expression after this unary operator.", "Add a value after the operator.")
			return nil
		}
		expr = ast.UnaryExpr{Span: source.Span{Start: t.Span.Start, End: ast.SpanOf(operand).End}, Operator: t.Text, Operand: operand}
	case token.LParen:
		p.advance()
		expr = p.parseExpression(0)
		if expr == nil {
			p.error("G2009", "I expected an expression inside these parentheses.", "Add an expression between ( and ).")
			return nil
		}
		p.expect(token.RParen, "G2010", "This parenthesized expression is missing a closing parenthesis.", "Add ) to close it.")
	case token.LBracket:
		expr = p.parseArray()
	case token.LBrace:
		expr = p.parseObject()
	default:
		return nil
	}
	return p.parsePostfix(expr)
}

func (p *Parser) parsePostfix(expr ast.Expr) ast.Expr {
	for {
		switch {
		case p.at(token.LParen):
			expr = p.parseCall(expr)
		case p.at(token.Dot), p.at(token.QuestionDot):
			start := ast.SpanOf(expr).Start
			optional := p.at(token.QuestionDot)
			p.advance()
			name := p.expect(token.Identifier, "G2020", "I expected a member name after property access.", "Add a property or method name.")
			if name.Kind == token.Invalid {
				return expr
			}
			expr = ast.MemberExpr{Span: source.Span{Start: start, End: name.Span.End}, Object: expr, Name: name.Text, Optional: optional}
		case p.at(token.LBracket):
			start := ast.SpanOf(expr).Start
			p.advance()
			index := p.parseExpression(0)
			if index == nil {
				p.error("G2021", "I expected an index expression.", "Add a value between [ and ].")
				return expr
			}
			close := p.expect(token.RBracket, "G2022", "This index expression is missing a closing bracket.", "Add ] to close the index.")
			end := ast.SpanOf(index).End
			if close.Kind != token.Invalid {
				end = close.Span.End
			}
			expr = ast.IndexExpr{Span: source.Span{Start: start, End: end}, Object: expr, Index: index}
		default:
			return expr
		}
	}
}

func (p *Parser) parseCall(callee ast.Expr) ast.Expr {
	start := ast.SpanOf(callee).Start
	p.advance()
	var args []ast.Expr
	if !p.at(token.RParen) {
		for {
			expr := p.parseExpression(0)
			if expr == nil {
				p.error("G2032", "I expected a function argument.", "Add an expression or close the argument list.")
				break
			}
			args = append(args, expr)
			if !p.at(token.Comma) {
				break
			}
			p.advance()
			if p.at(token.RParen) {
				break
			}
		}
	}
	end := p.current().Span.End
	if p.at(token.RParen) {
		end = p.advance().Span.End
	} else {
		p.error("G2010", "This function call is missing a closing parenthesis.", "Add ) after the arguments.")
	}
	return ast.CallExpr{Span: source.Span{Start: start, End: end}, Callee: callee, Arguments: args}
}

func (p *Parser) parseArray() ast.Expr {
	start := p.advance().Span.Start
	var items []ast.Expr
	for !p.at(token.RBracket) && !p.at(token.EOF) {
		expr := p.parseExpression(0)
		if expr == nil {
			p.error("G2033", "I expected an array element.", "Add a value or close the array.")
			break
		}
		items = append(items, expr)
		if !p.at(token.Comma) {
			break
		}
		p.advance()
	}
	end := p.current().Span.End
	if p.at(token.RBracket) {
		end = p.advance().Span.End
	} else {
		p.error("G2023", "This array is missing a closing bracket.", "Add ] to close the array.")
	}
	return ast.ArrayExpr{Span: source.Span{Start: start, End: end}, Items: items}
}

func (p *Parser) parseObject() ast.Expr {
	start := p.advance().Span.Start
	var properties []ast.ObjectProperty
	for !p.at(token.RBrace) && !p.at(token.EOF) {
		key := p.current()
		if key.Kind != token.Identifier && key.Kind != token.String {
			p.error("G2024", "I expected an object property name.", "Use an identifier or string as the property name.")
			p.recoverObject()
			if p.at(token.RBrace) || p.at(token.EOF) {
				break
			}
			continue
		}
		p.advance()
		if !p.at(token.Colon) {
			p.error("G2025", "I expected ':' after the object property name.", "Write key: value.")
			p.recoverObject()
			if p.at(token.RBrace) || p.at(token.EOF) {
				break
			}
			continue
		}
		p.advance()
		value := p.parseExpression(0)
		if value == nil {
			p.error("G2026", "I expected an object property value.", "Give the property a value.")
			p.recoverObject()
			if p.at(token.RBrace) || p.at(token.EOF) {
				break
			}
			continue
		}
		properties = append(properties, ast.ObjectProperty{Span: source.Span{Start: key.Span.Start, End: ast.SpanOf(value).End}, Key: key.Text, Value: value})
		if !p.at(token.Comma) {
			break
		}
		p.advance()
	}
	end := p.current().Span.End
	if p.at(token.RBrace) {
		end = p.advance().Span.End
	} else {
		p.error("G2027", "This object is missing a closing brace.", "Add } to close the object.")
	}
	return ast.ObjectExpr{Span: source.Span{Start: start, End: end}, Properties: properties}
}

func (p *Parser) recoverObject() {
	for !p.at(token.EOF) && !p.at(token.Comma) && !p.at(token.RBrace) {
		p.advance()
	}
	if p.at(token.Comma) {
		p.advance()
	}
}

func precedence(k token.Kind) int {
	switch k {
	case token.Equal, token.PlusEqual, token.MinusEqual, token.StarEqual, token.StarStarEqual, token.SlashEqual, token.PercentEqual, token.AndEqual, token.OrEqual, token.CaretEqual, token.ShiftLeftEqual, token.ShiftRightEqual, token.UnsignedShiftRightEqual:
		return 0
	case token.OrOr, token.QuestionQuestion:
		return 1
	case token.AndAnd:
		return 2
	case token.Or:
		return 3
	case token.Caret:
		return 4
	case token.And:
		return 5
	case token.EqualEqual, token.EqualEqualEqual, token.BangEqual, token.BangEqualEqual:
		return 6
	case token.Less, token.LessEqual, token.Greater, token.GreaterEqual, token.ShiftLeft, token.ShiftRight, token.UnsignedShiftRight:
		return 7
	case token.Plus, token.Minus:
		return 8
	case token.Star, token.Slash, token.Percent:
		return 9
	case token.StarStar:
		return 10
	default:
		return -1
	}
}

func isRightAssociative(k token.Kind) bool { return isAssignmentOperator(k) || k == token.StarStar }

func isAssignmentOperator(k token.Kind) bool {
	switch k {
	case token.Equal, token.PlusEqual, token.MinusEqual, token.StarEqual, token.StarStarEqual, token.SlashEqual, token.PercentEqual, token.AndEqual, token.OrEqual, token.CaretEqual, token.ShiftLeftEqual, token.ShiftRightEqual, token.UnsignedShiftRightEqual:
		return true
	default:
		return false
	}
}

func isAssignable(expr ast.Expr) bool {
	switch expr.(type) {
	case ast.Identifier, ast.MemberExpr, ast.IndexExpr:
		return true
	default:
		return false
	}
}

func (p *Parser) current() token.Token {
	if p.pos >= len(p.tokens) {
		return token.New(token.EOF, "", source.Span{})
	}
	return p.tokens[p.pos]
}

func (p *Parser) advance() token.Token {
	t := p.current()
	if p.pos < len(p.tokens) {
		p.pos++
	}
	return t
}

func (p *Parser) at(k token.Kind) bool { return p.current().Kind == k }
func (p *Parser) atKeyword(k grammar.Keyword) bool {
	if !p.at(token.Identifier) {
		return false
	}
	actual, ok := p.vocabulary.Lookup(p.current().Text)
	return ok && actual == k
}

func (p *Parser) expect(k token.Kind, code, message, hint string) token.Token {
	if p.at(k) {
		return p.advance()
	}
	p.error(code, message, hint)
	return token.New(token.Invalid, "", p.current().Span)
}

func (p *Parser) error(code, message, hint string) {
	p.diagnostics.Add(diagnostics.Diagnostic{Severity: diagnostics.Error, Code: code, Message: message, Hint: hint, Span: p.current().Span})
}

func (p *Parser) recoverStatement() {
	for !p.at(token.EOF) && !p.at(token.Semicolon) && !p.at(token.RBrace) {
		p.advance()
	}
	if p.at(token.Semicolon) {
		p.advance()
	}
}

func (p *Parser) parseImport() ast.Stmt {
	start := p.advance().Span.Start
	path := p.expect(token.String, "G2036", "I expected a quoted module path after import.", "Write import \"module\" or import \"module\" as alias.")
	if path.Kind == token.Invalid {
		p.recoverStatement()
		return nil
	}
	var alias *ast.Identifier
	if p.atKeyword(grammar.KeywordAs) {
		p.advance()
		name := p.expect(token.Identifier, "G2037", "I expected an import alias after as.", "Add an identifier for the imported module alias.")
		if name.Kind == token.Invalid {
			p.recoverStatement()
			return nil
		}
		alias = &ast.Identifier{Span: name.Span, Name: name.Text}
	}
	end := path.Span.End
	if alias != nil {
		end = alias.Span.End
	}
	p.consumeTerminator()
	return ast.ImportDecl{Span: source.Span{Start: start, End: end}, Path: path.Text, Alias: alias}
}

func (p *Parser) parseComponent(start source.Position) ast.Stmt {
	p.advance()
	name := p.expect(token.Identifier, "G2038", "I expected a component name.", "Give the component a name.")
	if name.Kind == token.Invalid {
		p.recoverStatement()
		return nil
	}
	var props []ast.ComponentProp
	if p.at(token.LParen) {
		p.advance()
		for !p.at(token.RParen) && !p.at(token.EOF) {
			prop := p.expect(token.Identifier, "G2039", "I expected a component property name.", "Write properties as name as value.")
			if prop.Kind == token.Invalid {
				p.recoverStatement()
				return nil
			}
			if !p.atKeyword(grammar.KeywordAs) {
				p.error("G2040", "I expected as after the component property name.", "Write properties as name as value.")
				p.recoverStatement()
				return nil
			}
			p.advance()
			value := p.parseExpression(0)
			if value == nil {
				p.error("G2041", "I expected a component property value.", "Add an expression after as.")
				p.recoverStatement()
				return nil
			}
			props = append(props, ast.ComponentProp{Span: source.Span{Start: prop.Span.Start, End: ast.SpanOf(value).End}, Name: prop.Text, Value: value})
			if !p.at(token.Comma) {
				break
			}
			p.advance()
		}
		p.expect(token.RParen, "G2042", "This component property list is missing a closing parenthesis.", "Add ) after the properties.")
	}
	body, ok := p.parseBlockRequired("G2043", "I expected a component body.", "Add { ... } after the component declaration.")
	if !ok {
		return nil
	}
	return ast.ComponentDecl{Span: source.Span{Start: start, End: body.Span.End}, Name: ast.Identifier{Span: name.Span, Name: name.Text}, Properties: props, Body: body}
}

func (p *Parser) canEndStatement(expr ast.Expr) bool {
	if p.at(token.Semicolon) || p.at(token.EOF) || p.at(token.RBrace) {
		return true
	}
	return ast.SpanOf(expr).End.Line < p.current().Span.Start.Line
}

func (p *Parser) consumeTerminator() {
	if p.at(token.Semicolon) {
		p.advance()
	}
}

func (p *Parser) looksLikeTypeThenAs() bool {
	if p.current().Kind != token.Identifier {
		return false
	}
	depth := 0
	for i := p.pos + 1; i < len(p.tokens); i++ {
		t := p.tokens[i]
		switch t.Kind {
		case token.Less, token.LBrace:
			depth++
		case token.Greater, token.RBrace:
			if depth > 0 {
				depth--
			}
		}
		if depth == 0 && t.Kind == token.Identifier {
			kw, ok := p.vocabulary.Lookup(t.Text)
			return ok && kw == grammar.KeywordAs
		}
		if depth == 0 && (t.Kind == token.Semicolon || t.Kind == token.EOF) {
			return false
		}
	}
	return false
}

func (p *Parser) parseType() ast.TypeRef { return p.parseUnionType() }

func (p *Parser) parseUnionType() ast.TypeRef {
	left := p.parseIntersectionType()
	for p.at(token.Or) {
		p.advance()
		right := p.parseIntersectionType()
		members := append([]ast.TypeRef{}, left.Union...)
		if len(members) == 0 {
			members = append(members, left)
		}
		if len(right.Union) > 0 {
			members = append(members, right.Union...)
		} else {
			members = append(members, right)
		}
		left = ast.TypeRef{Span: source.Span{Start: left.Span.Start, End: right.Span.End}, Union: members}
	}
	return left
}

func (p *Parser) parseIntersectionType() ast.TypeRef {
	left := p.parsePrimaryType()
	for p.at(token.And) {
		p.advance()
		right := p.parsePrimaryType()
		members := append([]ast.TypeRef{}, left.Intersection...)
		if len(members) == 0 {
			members = append(members, left)
		}
		if len(right.Intersection) > 0 {
			members = append(members, right.Intersection...)
		} else {
			members = append(members, right)
		}
		left = ast.TypeRef{Span: source.Span{Start: left.Span.Start, End: right.Span.End}, Intersection: members}
	}
	return left
}

func (p *Parser) parsePrimaryType() ast.TypeRef {
	if p.at(token.LParen) {
		start := p.advance().Span.Start
		t := p.parseType()
		close := p.expect(token.RParen, "G2044", "I expected ')' to close this type expression.", "Close the grouped type with ).")
		t.Span.Start = start
		if close.Kind != token.Invalid {
			t.Span.End = close.Span.End
		}
		return t
	}
	if p.at(token.String) || p.at(token.Number) {
		lit := p.advance()
		return ast.TypeRef{Span: lit.Span, Name: lit.Text}
	}
	name := p.expect(token.Identifier, "G2044", "I expected a type name.", "Use a type name such as String, Number, Boolean, Object, Optional, Result, or a user-defined type.")
	if name.Kind == token.Invalid {
		return ast.TypeRef{Span: name.Span}
	}
	t := ast.TypeRef{Span: name.Span, Name: name.Text}
	if p.at(token.Less) {
		p.advance()
		for !p.at(token.Greater) && !p.at(token.EOF) {
			arg := p.parseType()
			t.Arguments = append(t.Arguments, arg)
			if !p.at(token.Comma) {
				break
			}
			p.advance()
		}
		close := p.expect(token.Greater, "G2044", "I expected '>' to close type arguments.", "Close the type argument list with >.")
		if close.Kind != token.Invalid {
			t.Span.End = close.Span.End
		}
	}
	if strings.EqualFold(t.Name, "Record") || strings.EqualFold(t.Name, "Object") {
		if p.at(token.LBrace) {
			p.advance()
			for !p.at(token.RBrace) && !p.at(token.EOF) {
				readonly := false
				if p.atKeyword(grammar.KeywordReadonly) {
					readonly = true
					p.advance()
				}
				field := p.expect(token.Identifier, "G2044", "I expected a record field name.", "Use name: Type in a record type.")
				optional := false
				if p.at(token.Question) {
					optional = true
					p.advance()
				}
				p.expect(token.Colon, "G2044", "I expected ':' after a record field name.", "Write name: Type.")
				ft := p.parseType()
				t.Fields = append(t.Fields, ast.TypeFieldRef{Span: source.Span{Start: field.Span.Start, End: ft.Span.End}, Name: field.Text, Type: ft, Optional: optional, Readonly: readonly})
				if !p.at(token.Comma) {
					break
				}
				p.advance()
			}
			close := p.expect(token.RBrace, "G2044", "I expected '}' to close record type.", "Close the record type with }.")
			if close.Kind != token.Invalid {
				t.Span.End = close.Span.End
			}
		}
	}
	if p.at(token.LBracket) && p.pos+1 < len(p.tokens) && p.tokens[p.pos+1].Kind == token.RBracket {
		p.advance()
		close := p.advance()
		t.Array = true
		t.Span.End = close.Span.End
	}
	return t
}
