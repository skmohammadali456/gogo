package parser

import (
	"github.com/skmohammadali786/gogo/internal/ast"
	"github.com/skmohammadali786/gogo/internal/diagnostics"
	"github.com/skmohammadali786/gogo/internal/source"
	"github.com/skmohammadali786/gogo/internal/token"
)

type Parser struct { tokens []token.Token; pos int; diagnostics diagnostics.Bag }
func New(tokens []token.Token) *Parser { return &Parser{tokens: tokens} }
func (p *Parser) Diagnostics() []diagnostics.Diagnostic { return p.diagnostics.All() }
func (p *Parser) ParseFile() ast.File { var statements []ast.Stmt; start := p.current().Span.Start; for !p.at(token.EOF) { before := p.pos; if s := p.parseStatement(); s != nil { statements = append(statements, s) }; if p.pos == before { p.recoverStatement() } }; end := source.Position{}; if len(statements)>0 { end = ast.SpanOf(statements[len(statements)-1]).End }; return ast.File{Span: source.Span{Start:start, End:end}, Statements: statements} }

func (p *Parser) parseStatement() ast.Stmt {
	if p.atIdentifier("create") { return p.parseCreate() }
	if p.atIdentifier("return") { return p.parseReturn() }
	if p.at(token.LBrace) { return p.parseBlock() }
	e := p.parseExpression(0); if e == nil { return nil }; return ast.ExprStmt{Span: ast.SpanOf(e), Expression:e}
}
func (p *Parser) parseCreate() ast.Stmt {
	start := p.advance().Span.Start
	if !p.atIdentifier("variable") { p.error("G2001", "I expected 'variable' after 'create'.", "Write create variable name as value."); p.recoverStatement(); return nil }
	p.advance()
	name := p.expect(token.Identifier, "G2002", "I expected a variable name.", "Give the variable a name.")
	if name.Kind == token.Invalid { p.recoverStatement(); return nil }
	if p.atIdentifier("as") { p.advance() } else { p.error("G2003", "I expected 'as' after the variable name.", "Write create variable name as value."); return nil }
	value := p.parseExpression(0); if value == nil { p.error("G2004", "I expected a value for this variable.", "Give the variable an initial value."); return nil }
	return ast.VariableDecl{Span: source.Span{Start:start, End:ast.SpanOf(value).End}, Name:ast.Identifier{Span:name.Span, Name:name.Text}, Value:value}
}
func (p *Parser) parseReturn() ast.Stmt { start:=p.advance().Span.Start; value:=p.parseExpression(0); if value==nil { p.error("G2005", "I expected a value after 'return'.", "Return a value or use a grammar form that explicitly permits an empty return."); return nil }; return ast.ReturnStmt{Span:source.Span{Start:start, End:ast.SpanOf(value).End}, Value:value} }
func (p *Parser) parseBlock() ast.Stmt { start:=p.advance().Span.Start; var ss []ast.Stmt; for !p.at(token.RBrace) && !p.at(token.EOF) { before:=p.pos; if s:=p.parseStatement(); s!=nil { ss=append(ss,s) }; if before==p.pos { p.recoverStatement() } }; end:=p.current().Span.End; if p.at(token.RBrace) { end=p.advance().Span.End } else { p.error("G2006", "This block is missing its closing brace.", "Add } to close the block.") }; return ast.BlockStmt{Span:source.Span{Start:start,End:end},Statements:ss} }

func (p *Parser) parseExpression(minPrec int) ast.Expr { left:=p.parsePrefix(); if left==nil{return nil}; for { t:=p.current(); prec:=precedence(t.Kind); if prec<minPrec {break}; op:=p.advance(); right:=p.parseExpression(prec+1); if right==nil { p.error("G2007", "I expected an expression after this operator.", "Add a value after the operator."); return left }; left=ast.BinaryExpr{Span:source.Span{Start:ast.SpanOf(left).Start,End:ast.SpanOf(right).End},Left:left,Operator:op.Text,Right:right} }; return left }
func (p *Parser) parsePrefix() ast.Expr { t:=p.current(); switch t.Kind { case token.Identifier: p.advance(); var e ast.Expr=ast.Identifier{Span:t.Span,Name:t.Text}; if p.at(token.LParen){ e=p.parseCall(e) }; return e; case token.Number,token.String: p.advance(); return ast.Literal{Span:t.Span,Text:t.Text}; case token.Bang,token.Plus,token.Minus,token.Tilde: p.advance(); operand:=p.parseExpression(12); if operand==nil {p.error("G2008","I expected an expression after this unary operator.","Add a value after the operator.");return nil}; return ast.UnaryExpr{Span:source.Span{Start:t.Span.Start,End:ast.SpanOf(operand).End},Operator:t.Text,Operand:operand}; case token.LParen: p.advance(); e:=p.parseExpression(0); if !p.at(token.RParen){p.error("G2009","This parenthesized expression is missing a closing parenthesis.","Add ) to close it.")} else {p.advance()}; return e; default:return nil} }
func (p *Parser) parseCall(callee ast.Expr) ast.Expr { start:=ast.SpanOf(callee).Start; p.advance(); var args []ast.Expr; if !p.at(token.RParen){for { e:=p.parseExpression(0); if e==nil {break}; args=append(args,e); if !p.at(token.Comma){break}; p.advance() }}; end:=p.current().Span.End; if p.at(token.RParen){end=p.advance().Span.End}else{p.error("G2010","This function call is missing a closing parenthesis.","Add ) after the arguments.")}; return ast.CallExpr{Span:source.Span{Start:start,End:end},Callee:callee,Arguments:args} }

func precedence(k token.Kind) int { switch k { case token.OrOr,token.QuestionQuestion:return 1; case token.AndAnd:return 2; case token.Or:return 3; case token.Caret:return 4; case token.And:return 5; case token.EqualEqual,token.EqualEqualEqual,token.BangEqual,token.BangEqualEqual:return 6; case token.Less,token.LessEqual,token.Greater,token.GreaterEqual,token.ShiftLeft,token.ShiftRight,token.UnsignedShiftRight:return 7; case token.Plus,token.Minus:return 8; case token.Star,token.Slash,token.Percent:return 9; case token.StarStar:return 10; default:return -1} }
func (p *Parser) current() token.Token { if p.pos>=len(p.tokens){return token.New(token.EOF,"",source.Span{})}; return p.tokens[p.pos] }
func (p *Parser) advance() token.Token { t:=p.current(); if p.pos<len(p.tokens){p.pos++}; return t }
func (p *Parser) at(k token.Kind) bool{return p.current().Kind==k}
func (p *Parser) atIdentifier(s string) bool{return p.at(token.Identifier)&&p.current().Text==s}
func (p *Parser) expect(k token.Kind,code,msg,hint string) token.Token {if p.at(k){return p.advance()}; p.error(code,msg,hint); return token.New(token.Invalid,"",p.current().Span)}
func (p *Parser) error(code,msg,hint string){p.diagnostics.Add(diagnostics.Diagnostic{Severity:diagnostics.Error,Code:code,Message:msg,Hint:hint,Span:p.current().Span})}
func (p *Parser) recoverStatement(){for !p.at(token.EOF)&&!p.at(token.Semicolon)&&!p.at(token.RBrace){p.advance()}; if p.at(token.Semicolon){p.advance()}}
