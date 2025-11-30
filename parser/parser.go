package parser

import (
	"fmt"
	"strconv"

	// "strings"

	"github.com/akristianlopez/action/ast"
	"github.com/akristianlopez/action/lexer"
	"github.com/akristianlopez/action/token"
)

type Parser struct {
	l *lexer.Lexer

	curToken  token.Token
	peekToken token.Token

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn

	errors []string
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}

	// Initialisation des tables de parsing
	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.STRING, p.parseStringLiteral)
	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.TRUE, p.parseBoolean)
	p.registerPrefix(token.FALSE, p.parseBoolean)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(token.IF, p.parseIfExpression)
	p.registerPrefix(token.FUNCTION, p.parseFunctionLiteral)
	p.registerPrefix(token.STRUCT, p.parseStructLiteral)
	p.registerPrefix(token.NEW, p.parseNewExpression)

	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	p.registerInfix(token.LPAREN, p.parseCallExpression)
	p.registerInfix(token.DOT, p.parseMemberExpression)

	// Lire deux tokens pour initialiser curToken et peekToken
	p.nextToken()
	p.nextToken()

	return p
}

// Parser pour les expressions de membre (obj.champ)
func (p *Parser) parseMemberExpression(left ast.Expression) ast.Expression {
	exp := &ast.MemberExpression{Token: p.curToken, Object: left}

	p.nextToken()
	if !p.curTokenIs(token.IDENT) {
		p.errors = append(p.errors, "expected property name after '.'")
		return nil
	}

	exp.Property = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	return exp
}

// Parser pour les strings
func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) ParseProgram() *ast.ActionStatement {
	program := &ast.ActionStatement{}

	// Parser le mot-clé 'action'
	if !p.curTokenIs(token.ACTION) {
		p.errors = append(p.errors, "expected 'action' at beginning of program")
		return nil
	}
	program.Token = p.curToken

	// Parser le nom de l'action (string)
	p.nextToken()
	if !p.curTokenIs(token.STRING) {
		p.errors = append(p.errors, "expected string after 'action'")
		return nil
	}
	program.Name = &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}

	// Parser les déclarations
	p.nextToken()
	program.Declarations = p.parseDeclarationBlock()

	// Parser 'start'
	if !p.curTokenIs(token.START) {
		p.errors = append(p.errors, "expected 'start' after declarations")
		return nil
	}
	program.Start = p.curToken

	// Parser le corps du programme
	p.nextToken()
	program.Body = p.parseBlockStatement()

	// Parser 'stop'
	if !p.curTokenIs(token.STOP) {
		p.errors = append(p.errors, "expected 'stop' at end of program")
		return nil
	}
	program.Stop = p.curToken

	return program
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.LET:
		return p.parseLetStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	p.nextToken()

	stmt.Value = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}

	p.nextToken()

	stmt.ReturnValue = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}

	stmt.Expression = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// Logique de parsing des expressions avec precedence
type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

const (
	_ int = iota
	LOWEST
	EQUALS      // ==
	LESSGREATER // > or <
	SUM         // +
	PRODUCT     // *
	PREFIX      // -X or !X
	CALL        // myFunction(X)
)

var precedences = map[token.TokenType]int{
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.SLASH:    PRODUCT,
	token.ASTERISK: PRODUCT,
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type] //A verifier
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()

	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		p.nextToken()

		leftExp = infix(leftExp) //A verifier
	}

	return leftExp
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken}

	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value
	return lit
}

func (p *Parser) parseBoolean() ast.Expression {
	return &ast.Boolean{Token: p.curToken, Value: p.curTokenIs(token.TRUE)}
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}

	p.nextToken()

	expression.Right = p.parseExpression(PREFIX)

	return expression
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()

	exp := p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return exp
}

func (p *Parser) parseIfExpression() ast.Expression {
	expression := &ast.IfExpression{Token: p.curToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	p.nextToken()
	expression.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	expression.Consequence = p.parseBlockStatement()

	if p.peekTokenIs(token.ELSE) {
		p.nextToken()

		if !p.expectPeek(token.LBRACE) {
			return nil
		}

		expression.Alternative = p.parseBlockStatement()
	}

	return expression
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken}
	block.Statements = []ast.Statement{}

	p.nextToken()

	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}

	return block
}

// Fonctions utilitaires
func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		p.peekError(t)
		return false
	}
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead",
		t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

// Tables de parsing
func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, msg)
}

// func (p *Parser) parseFunctionLiteral() *ast.FunctionLiteral {
// 	lit := &ast.FunctionLiteral{Token: p.curToken}

// 	if !p.expectPeek(token.LPAREN) {
// 		return nil
// 	}

// 	lit.Parameters = p.parseFunctionParameters()

// 	if !p.expectPeek(token.LBRACE) {
// 		return nil
// 	}

// 	lit.Body = p.parseBlockStatement()

// 	return lit
// }

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curToken, Function: function}
	exp.Arguments = p.parseExpressionList(token.RPAREN)
	return exp
}

func (p *Parser) parseExpressionList(end token.TokenType) []ast.Expression {
	list := []ast.Expression{}

	if p.peekTokenIs(end) {
		p.nextToken()
		return list
	}

	p.nextToken()
	list = append(list, p.parseExpression(LOWEST))

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		list = append(list, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(end) {
		return nil
	}

	return list
}

func (p *Parser) parseStructField() *ast.StructField {
	if !p.curTokenIs(token.IDENT) {
		p.errors = append(p.errors, "expected field name")
		return nil
	}

	field := &ast.StructField{
		Token: p.curToken,
		Name:  &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal},
	}

	if !p.expectPeek(token.COLON) {
		return nil
	}

	p.nextToken()
	if !p.curTokenIs(token.IDENT) {
		p.errors = append(p.errors, "expected field type")
		return nil
	}

	field.Type = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	return field
}

// Nouvelle fonction pour parser le bloc declaration
func (p *Parser) parseDeclarationBlock() *ast.DeclarationBlock {
	block := &ast.DeclarationBlock{
		Token:     p.curToken, // token DECLARATION
		Structs:   []*ast.StructStatement{},
		Functions: []*ast.FunctionLiteral{},
		Variables: []*ast.LetStatement{},
	}

	p.nextToken() // avancer après 'declaration'

	// Parser jusqu'au 'end' ou 'start'
	for !p.curTokenIs(token.START) && !p.curTokenIs(token.EOF) {
		switch p.curToken.Type {
		case token.STRUCT:
			stmt := p.parseStructStatement()
			if stmt != nil {
				block.Structs = append(block.Structs, stmt)
			}
		case token.FUNCTION:
			fn := p.parseFunctionLiteral()
			if fn != nil {
				block.Functions = append(block.Functions, fn)
			}
		case token.LET:
			letStmt := p.parseLetStatement()
			if letStmt != nil {
				block.Variables = append(block.Variables, letStmt)
			}
		default:
			if p.curTokenIs(token.START) {
				break // sortir de la boucle si on rencontre 'start'
			}
			p.errors = append(p.errors, "unexpected token in declaration block: "+p.curToken.Literal)
			p.nextToken()
		}
		p.nextToken()
	}

	return block
}

// Modifier parseFunctionLiteral pour qu'il soit indenté dans le bloc
func (p *Parser) parseFunctionLiteral() *ast.FunctionLiteral {
	lit := &ast.FunctionLiteral{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	lit.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	lit.Parameters = p.parseFunctionParameters()

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	lit.Body = p.parseBlockStatement()

	return lit
}

// Modifier parseStructStatement pour qu'il soit indenté dans le bloc
func (p *Parser) parseStructStatement() *ast.StructStatement {
	stmt := &ast.StructStatement{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	p.nextToken()
	stmt.Fields = []*ast.StructField{}

	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		field := p.parseStructField()
		if field != nil {
			stmt.Fields = append(stmt.Fields, field)
		}
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseFunctionParameters() []*ast.FunctionParameter {
	parameters := []*ast.FunctionParameter{}

	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return parameters
	}

	p.nextToken()

	param := &ast.FunctionParameter{
		Token: p.curToken,
		Name:  &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal},
	}

	if !p.expectPeek(token.COLON) {
		return nil
	}

	p.nextToken()
	if !p.curTokenIs(token.IDENT) {
		p.errors = append(p.errors, "expected parameter type")
		return nil
	}

	param.Type = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	parameters = append(parameters, param)

	for p.peekTokenIs(token.COMMA) {
		p.nextToken() // comma
		p.nextToken() // next param

		param := &ast.FunctionParameter{
			Token: p.curToken,
			Name:  &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal},
		}

		if !p.expectPeek(token.COLON) {
			return nil
		}

		p.nextToken()
		if !p.curTokenIs(token.IDENT) {
			p.errors = append(p.errors, "expected parameter type")
			return nil
		}

		param.Type = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		parameters = append(parameters, param)
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return parameters
}

// // Modifier parseLetStatement pour qu'il soit indenté dans le bloc
// func (p *Parser) parseLetStatement() *ast.LetStatement {
// 	stmt := &ast.LetStatement{Token: p.curToken}

// 	if !p.expectPeek(token.IDENT) {
// 		return nil
// 	}

// 	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

// 	if !p.expectPeek(token.ASSIGN) {
// 		return nil
// 	}

// 	p.nextToken()

// 	stmt.Value = p.parseExpression(LOWEST)

// 	if p.peekTokenIs(token.SEMICOLON) {
// 		p.nextToken()
// 	}

// 	return stmt
// }
