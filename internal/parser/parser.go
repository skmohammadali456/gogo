package parser

import (
	"github.com/skmohammadali786/gogo/internal/ast"
	"github.com/skmohammadali786/gogo/internal/diagnostics"
	"github.com/skmohammadali786/gogo/internal/grammar"
	"github.com/skmohammadali786/gogo/internal/source"
	"github.com/skmohammadali786/gogo/internal/token"
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
	if p.at(token.Semicolon) {
		p.advance()
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
	p.error("G2001", "I expected 'variable' or 'function' after 'create'.", "Write create variable name as value, or create function name(...).")
	p.recoverStatement()
	return nil
}

func (p *Parser) parseVariable(start source.Position) ast.Stmt {
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
	value := p.parseExpression(0)
	if value == nil {
		p.error("G2004", "I expected a value after as.", "Give the variable an initial value.")
		return nil
	}
	if p.at(token.Semicolon) {
		p.advance()
	}
	return ast.VariableDecl{Span: source.Span{Start: start, End: ast.SpanOf(value).End}, Name: ast.Identifier{Span: name.Span, Name: name.Text}, Value: value}
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
	if !p.at(token.RParen) {
		for {
			param := p.expect(token.Identifier, "G2013", "I expected a parameter name.", "Give each parameter a name.")
			if param.Kind == token.Invalid {
				p.recoverStatement()
				return nil
			}
			params = append(params, ast.Identifier{Span: param.Span, Name: param.Text})
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
	body, ok := p.parseBlockRequired("G2015", "I expected a function body.", "Add { ... } after the function declaration.")
	if !ok {
		return nil
	}
	return ast.FunctionDecl{Span: source.Span{Start: start, End: body.Span.End}, Name: ast.Identifier{Span: name.Span, Name: name.Text}, Parameters: params, Body: body}
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
	if p.at(token.Semicolon) {
		p.advance()
	}
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
