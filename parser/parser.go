package parser

import (
	"fmt"
	"strconv"

	"github.com/akristianlopez/action/ast"
	"github.com/akristianlopez/action/lexer"
	"github.com/akristianlopez/action/token"
)

type Parser struct {
	l *lexer.Lexer

	curToken  token.Token
	peekToken token.Token

	errors []ParserError

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

type ParserError struct {
	msg    string
	line   int
	column int
}

func (pe *ParserError) Message() string {
	return pe.msg
}
func (pe *ParserError) Line() int   { return pe.line }
func (pe *ParserError) Column() int { return pe.column }
func Create(message string, line, column int) *ParserError {
	return &ParserError{msg: message, line: line, column: column}
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []ParserError{},
	}

	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT_LIT, p.parseIntegerLiteral)
	p.registerPrefix(token.FLOAT_LIT, p.parseFloatLiteral)
	p.registerPrefix(token.STRING_LIT, p.parseStringLiteral)
	p.registerPrefix(token.BOOL_LIT, p.parseBooleanLiteral)
	p.registerPrefix(token.TIME_LIT, p.parseDateTimeLiteral)
	p.registerPrefix(token.DATE_LIT, p.parseDateTimeLiteral)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.SELECT, p.parseSQLSelect)
	p.registerPrefix(token.OBJECT, p.parsePrefixObjectValue)

	// Enregistrer les fonctions de fenêtrage
	p.registerPrefix(token.ROW_NUMBER, p.parseWindowFunction)
	p.registerPrefix(token.RANK, p.parseWindowFunction)
	p.registerPrefix(token.DENSE_RANK, p.parseWindowFunction)
	p.registerPrefix(token.LAG, p.parseWindowFunction)
	p.registerPrefix(token.LEAD, p.parseWindowFunction)
	p.registerPrefix(token.FIRST_VALUE, p.parseWindowFunction)
	p.registerPrefix(token.LAST_VALUE, p.parseWindowFunction)
	p.registerPrefix(token.NTILE, p.parseWindowFunction)
	p.registerPrefix(token.LBRACKET, p.parseArrayLiteral)
	p.registerPrefix(token.LENGTH, p.parseArrayFunctionCall)
	p.registerPrefix(token.APPEND, p.parseArrayFunctionCall)
	p.registerPrefix(token.PREPEND, p.parseArrayFunctionCall)
	p.registerPrefix(token.REMOVE, p.parseArrayFunctionCall)
	p.registerPrefix(token.SLICE, p.parseArrayFunctionCall)
	p.registerPrefix(token.CONTAINS, p.parseArrayFunctionCall)

	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.MOD, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	p.registerInfix(token.LTE, p.parseInfixExpression)
	p.registerInfix(token.GTE, p.parseInfixExpression)
	p.registerInfix(token.AND, p.parseInfixExpression)
	p.registerInfix(token.OR, p.parseInfixExpression)
	p.registerInfix(token.LBRACKET, p.parseIndexOrSliceExpression)
	p.registerInfix(token.IN, p.parseInExpression)
	p.registerInfix(token.AS, p.parseInfixExpression)
	p.registerInfix(token.DOT, p.parseInfixExpression)

	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()

	// Ignorer les commentaires
	for p.curToken.Type == token.COMMENT {
		p.curToken = p.peekToken
		p.peekToken = p.l.NextToken()
	}
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}

	// Vérifier que le programme commence par 'action'
	if !p.curTokenIs(token.ACTION) {
		p.errors = append(p.errors, *Create("The action must start with the word 'action'", p.curToken.Line, p.curToken.Column))
		return program
	}

	p.nextToken()

	// Lire le nom de l'action
	if !p.curTokenIs(token.STRING_LIT) {
		p.errors = append(p.errors, *Create("Attendu un nom d'action après 'action'", p.curToken.Line, p.curToken.Column))
		return program
	}
	program.ActionName = p.curToken.Literal

	p.nextToken()

	// Parser les déclarations jusqu'à 'start'
	for !p.curTokenIs(token.START) && !p.curTokenIs(token.EOF) {
		stmt, pe := p.parseStatement()
		if pe != nil {
			p.errors = append(p.errors, *pe)
		}
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}

	// Parser les instructions après 'start'
	if p.curTokenIs(token.START) {
		p.nextToken()
		for !p.curTokenIs(token.STOP) && !p.curTokenIs(token.EOF) {
			stmt, pe := p.parseStatement()
			if pe != nil {
				p.errors = append(p.errors, *pe)
			}
			if stmt != nil {
				program.Statements = append(program.Statements, stmt)
			}
			p.nextToken()
		}
	}

	return program
}

func (p *Parser) parseStatement() (ast.Statement, *ParserError) {
	switch p.curToken.Type {
	case token.LET:
		return p.parseLetStatement()
	case token.FUNCTION:
		return p.parseFunctionStatement()
	case token.STRUCT:
		return p.parseStructStatement()
	case token.FOR:
		return p.parseForStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseLetStatement() (*ast.LetStatement, *ParserError) {
	stmt := &ast.LetStatement{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil, Create("Identifier expected", p.peekToken.Line, p.peekToken.Column)
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.peekTokenIs(token.COLON) && !p.peekTokenIs(token.ASSIGN) {
		return nil, Create("type expected", p.peekToken.Line, p.peekToken.Column)
	}

	// Vérifier s'il y a une annotation de type
	if p.peekTokenIs(token.COLON) {
		p.nextToken() // :
		p.nextToken() // type
		stmt.Type = p.parseTypeAnnotation()
	}

	if p.peekTokenIs(token.ASSIGN) {
		p.nextToken() // =
		p.nextToken() // valeur
		stmt.Value = p.parseExpression(LOWEST)
	}

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseTypeConstraints() (*ast.TypeConstraints, *ParserError) {
	constraints := &ast.TypeConstraints{}
	tok := p.curToken.Type
	var pe *ParserError

	for p.peekTokenIs(token.LPAREN) || p.peekTokenIs(token.LBRACKET) {
		p.nextToken()
		switch p.curToken.Type {
		case token.LPAREN:
			if p.peekTokenIs(token.INT_LIT) && tok != token.STRING {
				p.nextToken()
				maxDigits := &ast.IntegerLiteral{Token: p.curToken}
				val, _ := strconv.ParseInt(p.curToken.Literal, 10, 64)
				maxDigits.Value = val
				constraints.MaxDigits = maxDigits

				if p.peekTokenIs(token.COMMA) { //p.peekTokenIs(token.DOT)
					p.nextToken() // ,
					p.nextToken() // decimal places
					decimalPlaces := &ast.IntegerLiteral{Token: p.curToken}
					val, _ := strconv.ParseInt(p.curToken.Literal, 10, 64)
					decimalPlaces.Value = val
					constraints.DecimalPlaces = decimalPlaces
				}
			} else if p.peekTokenIs(token.INT_LIT) {
				p.nextToken()
				maxLength := &ast.IntegerLiteral{Token: p.curToken}
				val, _ := strconv.ParseInt(p.curToken.Literal, 10, 64)
				maxLength.Value = val
				constraints.MaxLength = maxLength
			}
			if !p.expectPeek(token.RPAREN) {
				return nil, Create("')'", p.peekToken.Line, p.peekToken.Column)
			}
		case token.LBRACKET:
			constraints.IntegerRange, pe = p.parseRangeConstraint()
			if pe != nil {
				p.errors = append(p.errors, *pe)
			}
			if !p.expectPeek(token.RBRACKET) {
				return nil, Create("']'", p.peekToken.Line, p.peekToken.Column)
			}
		}
	}

	return constraints, nil
}

func (p *Parser) parseRangeConstraint() (*ast.RangeConstraint, *ParserError) {
	rc := &ast.RangeConstraint{}

	p.nextToken()
	rc.Min = p.parseExpression(LOWEST)

	if !p.expectPeek(token.DOT) { //!p.expectPeek(token.DOT) ||
		return nil, Create("'.' expected", p.peekToken.Line, p.peekToken.Column)
	}

	p.nextToken()
	rc.Max = p.parseExpression(LOWEST)

	return rc, nil
}

func (p *Parser) parseFunctionStatement() (*ast.FunctionStatement, *ParserError) {
	stmt := &ast.FunctionStatement{Token: p.curToken}
	var pe *ParserError

	if !p.expectPeek(token.IDENT) {
		return nil, Create("Identifier expected", p.peekToken.Line, p.peekToken.Column)
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.LPAREN) {
		return nil, Create("'(' expected", p.peekToken.Line, p.peekToken.Column)
	}

	stmt.Parameters, pe = p.parseFunctionParameters()
	if pe != nil {
		p.errors = append(p.errors, *pe)
	}
	// Type de retour optionnel
	if p.peekTokenIs(token.COLON) {
		p.nextToken() // :
		p.nextToken() // type
		if !p.peekTokenIs(token.LBRACE) {
			stmt.ReturnType = p.parseTypeAnnotation()
		}
	}

	if !p.expectPeek(token.LBRACE) {
		return nil, Create("'{' expected", p.peekToken.Line, p.peekToken.Column)
	}

	stmt.Body, pe = p.parseBlockStatement()

	return stmt, pe
}

func (p *Parser) parseFunctionParameters() ([]*ast.FunctionParameter, *ParserError) {
	var params []*ast.FunctionParameter

	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return params, nil
	}

	p.nextToken()

	param := &ast.FunctionParameter{Token: p.curToken}
	param.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.COLON) {
		return nil, Create("':' expected", p.peekToken.Line, p.peekToken.Column)
	}

	p.nextToken()
	param.Type = p.parseTypeAnnotation()
	params = append(params, param)

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()

		param := &ast.FunctionParameter{Token: p.curToken}
		param.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

		if !p.expectPeek(token.COLON) {
			return nil, Create("':' expected", p.peekToken.Line, p.peekToken.Column)
		}

		p.nextToken()
		param.Type = p.parseTypeAnnotation()
		params = append(params, param)
	}

	if !p.expectPeek(token.RPAREN) {
		return nil, Create("')' expected", p.peekToken.Line, p.peekToken.Column)
	}

	return params, nil
}

func (p *Parser) parseStructStatement() (*ast.StructStatement, *ParserError) {
	stmt := &ast.StructStatement{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil, Create("Identifier expected", p.peekToken.Line, p.peekToken.Column)
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.LBRACE) {
		return nil, Create("'{' expected", p.peekToken.Line, p.peekToken.Column)
	}

	p.nextToken()

	// Parser les champs
	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		field := &ast.StructField{Token: p.curToken}
		field.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

		if !p.expectPeek(token.COLON) {
			return nil, Create("':' expected", p.peekToken.Line, p.peekToken.Column)
		}

		p.nextToken()
		field.Type = p.parseTypeAnnotation()
		stmt.Fields = append(stmt.Fields, field)

		if p.peekTokenIs(token.COMMA) {
			p.nextToken()
		}
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseForStatement() (*ast.ForStatement, *ParserError) {
	stmt := &ast.ForStatement{Token: p.curToken}
	var pe *ParserError

	//this modification is important insofar as it allows not to oblige
	//developper to put absolutely the parentheses when they wrote the
	//statement for. But the brace is the one thing obligatory to show
	//the beginning of the list of statements

	// if !p.expectPeek(token.LPAREN) {
	// 	return nil
	// }

	//check if the NextToken is '(' if true move to the next token
	if p.peekTokenIs(token.LPAREN) {
		p.nextToken()
	}
	p.nextToken()

	// Initialisation
	if !p.curTokenIs(token.SEMICOLON) {
		var stm ast.Statement
		stm, pe = p.parseStatement()
		if pe != nil {
			p.errors = append(p.errors, *pe)
		}
		if stm != nil {
			stmt.Init = stm
		}
		if !p.curTokenIs(token.SEMICOLON) {
			return nil, Create("';' expected", p.peekToken.Line, p.peekToken.Column)
		}
	}

	// if !p.expectPeek(token.SEMICOLON) {
	// 	return nil, Create("';' expected", p.peekToken.Line, p.peekToken.Column)
	// }
	// p.nextToken()
	if p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	// Condition
	if !p.curTokenIs(token.SEMICOLON) {
		stmt.Condition = p.parseExpression(LOWEST)
	}

	if !p.expectPeek(token.SEMICOLON) {
		return nil, Create("';' expected", p.peekToken.Line, p.peekToken.Column)
	}

	p.nextToken()

	// Update
	if !p.curTokenIs(token.RPAREN) {
		var tp ast.Statement
		tp, pe = p.parseStatement()
		if pe != nil {
			p.errors = append(p.errors, *pe)
		}
		stmt.Update = tp
	}

	// if !p.expectPeek(token.RPAREN) {
	// 	return nil
	// }
	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
	}

	if !p.expectPeek(token.LBRACE) {
		return nil, Create("'{' expected", p.peekToken.Line, p.peekToken.Column)
	}

	stmt.Body, pe = p.parseBlockStatement()

	return stmt, pe
}

func (p *Parser) parseReturnStatement() (*ast.ReturnStatement, *ParserError) {
	stmt := &ast.ReturnStatement{Token: p.curToken}

	p.nextToken()

	if !p.curTokenIs(token.SEMICOLON) {
		stmt.ReturnValue = p.parseExpression(LOWEST)
	}

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	// if p.curTokenIs(token.SEMICOLON) {
	// 	p.nextToken()
	// }

	return stmt, nil
}

func (p *Parser) parseBlockStatement() (*ast.BlockStatement, *ParserError) {
	block := &ast.BlockStatement{Token: p.curToken}
	// var pe *ParserError

	p.nextToken()

	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		stmt, pe := p.parseStatement()
		if pe != nil {
			p.errors = append(p.errors, *pe)
		}
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}

	return block, nil
}

func (p *Parser) parseExpressionStatement() (*ast.ExpressionStatement, *ParserError) {
	stmt := &ast.ExpressionStatement{Token: p.curToken}

	stmt.Expression = p.parseExpression(LOWEST)
	if p.peekTokenIs(token.ASSIGN) {
		p.nextToken() //read =
		p.nextToken() //read next token
		stmt.Expression = p.parseExpression(LOWEST)
	}
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt, nil
}

// Les méthodes restantes pour les expressions (parseExpression, parsePrefixExpression, etc.)
// seraient similaires à celles d'un parser standard...

const (
	_ int = iota
	LOWEST
	EQUALS
	LESSGREATER
	SUM
	PRODUCT
	PREFIX
	CALL
	INDEX
	AS
	DOT
)

func (p *Parser) parsePrefixObjectValue() ast.Expression {
	ident := &ast.Identifier{Token: p.curToken, Value: ""}
	p.nextToken()
	ident.Value = p.curToken.Literal
	return ident
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
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
		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	// tok := p.curToken
	// if p.peekToken.Type == token.DOT {
	// 	p.nextToken() // .
	// 	if p.peekToken.Type == token.IDENT {
	// 		p.nextToken() // ident
	// 		tok.Literal = fmt.Sprintf("%s.%s", tok.Literal, p.curToken.Literal)
	// 	}
	// }
	// return &ast.Identifier{Token: tok, Value: tok.Literal}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken}

	value, err := strconv.ParseInt(p.curToken.Literal, 10, 64)
	if err != nil {
		p.errors = append(p.errors, *Create(fmt.Sprintf("Impossible de parser %q comme entier", p.curToken.Literal),
			p.curToken.Line, p.curToken.Column))
		return nil
	}

	lit.Value = value
	return lit
}

func (p *Parser) parseFloatLiteral() ast.Expression {
	lit := &ast.FloatLiteral{Token: p.curToken}

	value, err := strconv.ParseFloat(p.curToken.Literal, 64)
	if err != nil {
		p.errors = append(p.errors, *Create(fmt.Sprintf("Impossible de parser %q comme flottant", p.curToken.Literal),
			p.curToken.Line, p.curToken.Column))
		return nil
	}

	lit.Value = value
	return lit
}

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseBooleanLiteral() ast.Expression {
	return &ast.BooleanLiteral{Token: p.curToken, Value: p.curToken.Literal == "true"}
}

func (p *Parser) parseDateTimeLiteral() ast.Expression {
	isTime := p.curToken.Type == token.TIME_LIT
	return &ast.DateTimeLiteral{
		Token:  p.curToken,
		Value:  p.curToken.Literal,
		IsTime: isTime,
	}
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()

	exp := p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {

		return nil
	}

	return exp
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

func (p *Parser) parseSQLSelect() ast.Expression {
	selectStmt := &ast.SQLSelectStatement{Token: p.curToken}

	// Parser SELECT
	p.nextToken()
	selectStmt.Select = p.parseSelectList()

	// Parser FROM
	if !p.expectPeek(token.FROM) {
		return nil
	}
	p.nextToken()
	selectStmt.From = p.parseExpression(LOWEST)

	// Parser les JOINs optionnels
	for p.peekTokenIs(token.JOIN) ||
		(p.peekTokenIs(token.IDENT) &&
			(p.peekToken.Literal == "INNER" || p.peekToken.Literal == "LEFT" ||
				p.peekToken.Literal == "RIGHT" || p.peekToken.Literal == "FULL")) {
		p.nextToken()
		join := &ast.SQLJoin{Token: p.curToken}

		if p.curTokenIs(token.IDENT) {
			join.Type = p.curToken.Literal
			if !p.expectPeek(token.JOIN) {
				return nil
			}
			p.nextToken()
		} else {
			join.Type = "INNER"
		}

		join.Table = p.parseExpression(LOWEST)

		if !p.expectPeek(token.ON) {
			return nil
		}
		p.nextToken()
		join.On = p.parseExpression(LOWEST)

		selectStmt.Joins = append(selectStmt.Joins, join)
	}

	// Parser WHERE optionnel
	if p.peekTokenIs(token.WHERE) {
		p.nextToken()
		p.nextToken()
		selectStmt.Where = p.parseExpression(LOWEST)
	}

	return selectStmt
}

func (p *Parser) parseSQLStatement() ast.Statement {
	switch p.curToken.Type {
	case token.CREATE:
		if p.peekTokenIs(token.OBJECT) {
			return p.parseSQLCreateObject()
		} else if p.peekTokenIs(token.INDEX) {
			return p.parseSQLCreateIndex()
		}
	case token.DROP:
		if p.peekTokenIs(token.OBJECT) {
			return p.parseSQLDropObject()
		}
	case token.ALTER:
		if p.peekTokenIs(token.OBJECT) {
			return p.parseSQLAlterObject()
		}
	case token.INSERT:
		return p.parseSQLInsert()
	case token.UPDATE:
		return p.parseSQLUpdate()
	case token.DELETE:
		return p.parseSQLDelete()
	case token.TRUNCATE:
		if p.peekTokenIs(token.OBJECT) {
			return p.parseSQLTruncate()
		}
	case token.SELECT:
		return p.parseSQLSelectStatement()
	}
	return nil
}

func (p *Parser) parseSQLCreateObject() *ast.SQLCreateObjectStatement {
	stmt := &ast.SQLCreateObjectStatement{Token: p.curToken}

	// CREATE
	if !p.expectPeek(token.OBJECT) {
		return nil
	}

	// IF NOT EXISTS optionnel
	if p.peekTokenIs(token.IF) {
		p.nextToken() // IF
		p.nextToken() // NOT
		p.nextToken() // EXISTS
		stmt.IfNotExists = true
	}

	// Nom de l'objet
	if !p.expectPeek(token.IDENT) {
		return nil
	}
	stmt.ObjectName = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// (
	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	p.nextToken()

	// Colonnes et contraintes
	for !p.curTokenIs(token.RPAREN) && !p.curTokenIs(token.EOF) {
		if p.curTokenIs(token.CONSTRAINT) {
			constraint := p.parseSQLConstraint()
			if constraint != nil {
				stmt.Constraints = append(stmt.Constraints, constraint)
			}
		} else {
			column := p.parseSQLColumnDefinition()
			if column != nil {
				stmt.Columns = append(stmt.Columns, column)
			}
		}

		if p.peekTokenIs(token.COMMA) {
			p.nextToken()
		}
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseSQLColumnDefinition() *ast.SQLColumnDefinition {
	col := &ast.SQLColumnDefinition{Token: p.curToken}
	col.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// Type de données
	if !p.expectPeek(token.IDENT) {
		return nil
	}
	col.DataType = p.parseSQLDataType()

	// Contraintes de colonne
	for p.peekTokenIs(token.NOT) || p.peekTokenIs(token.UNIQUE) ||
		p.peekTokenIs(token.PRIMARY) || p.peekTokenIs(token.CHECK) ||
		p.peekTokenIs(token.DEFAULT) {
		p.nextToken()
		constraint := p.parseSQLColumnConstraint()
		if constraint != nil {
			col.Constraints = append(col.Constraints, constraint)
		}
	}

	return col
}

func (p *Parser) parseSQLDataType() *ast.SQLDataType {
	dt := &ast.SQLDataType{Token: p.curToken, Name: p.curToken.Literal}

	// Longueur/Précision optionnelle
	if p.peekTokenIs(token.LPAREN) {
		p.nextToken() // (
		p.nextToken()

		if p.curTokenIs(token.INT_LIT) {
			length := &ast.IntegerLiteral{Token: p.curToken}
			val, _ := strconv.ParseInt(p.curToken.Literal, 10, 64)
			length.Value = val

			if p.peekTokenIs(token.COMMA) {
				// NUMERIC/DECIMAL avec précision et échelle
				dt.Precision = length
				p.nextToken() // ,
				p.nextToken()
				scale := &ast.IntegerLiteral{Token: p.curToken}
				val, _ = strconv.ParseInt(p.curToken.Literal, 10, 64)
				scale.Value = val
				dt.Scale = scale
			} else {
				// VARCHAR/CHAR avec longueur
				dt.Length = length
			}
		}

		if !p.expectPeek(token.RPAREN) {
			return nil
		}
	}

	return dt
}

func (p *Parser) parseSQLColumnConstraint() *ast.SQLColumnConstraint {
	constraint := &ast.SQLColumnConstraint{Token: p.curToken}

	switch p.curToken.Type {
	case token.NOT:
		if !p.expectPeek(token.NULL) {
			return nil
		}
		constraint.Type = "NOT NULL"
	case token.UNIQUE:
		constraint.Type = "UNIQUE"
	case token.PRIMARY:
		if !p.expectPeek(token.KEY) {
			return nil
		}
		constraint.Type = "PRIMARY KEY"
	case token.DEFAULT:
		constraint.Type = "DEFAULT"
		p.nextToken()
		constraint.Expression = p.parseExpression(LOWEST)
		return constraint
	case token.CHECK:
		constraint.Type = "CHECK"
		p.nextToken()
		constraint.Expression = p.parseExpression(LOWEST)
		return constraint
	}

	return constraint
}

func (p *Parser) parseSQLConstraint() *ast.SQLConstraint {
	constraint := &ast.SQLConstraint{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}
	constraint.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	p.nextToken()

	switch p.curToken.Type {
	case token.PRIMARY:
		constraint.Type = "PRIMARY KEY"
		if !p.expectPeek(token.KEY) {
			return nil
		}
		if !p.expectPeek(token.LPAREN) {
			return nil
		}
		constraint.Columns = p.parseColumnList()
	case token.FOREIGN:
		constraint.Type = "FOREIGN KEY"
		if !p.expectPeek(token.KEY) {
			return nil
		}
		if !p.expectPeek(token.LPAREN) {
			return nil
		}
		constraint.Columns = p.parseColumnList()
		if !p.expectPeek(token.REFERENCES) {
			return nil
		}
		constraint.References = p.parseSQLReference()
	case token.UNIQUE:
		constraint.Type = "UNIQUE"
		if !p.expectPeek(token.LPAREN) {
			return nil
		}
		constraint.Columns = p.parseColumnList()
	case token.CHECK:
		constraint.Type = "CHECK"
		p.nextToken()
		constraint.Check = p.parseExpression(LOWEST)
	}

	return constraint
}

func (p *Parser) parseColumnList() []*ast.Identifier {
	var columns []*ast.Identifier

	p.nextToken()
	columns = append(columns, &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal})

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		columns = append(columns, &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal})
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return columns
}

func (p *Parser) parseSQLReference() *ast.SQLReference {
	ref := &ast.SQLReference{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}
	ref.TableName = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if p.peekTokenIs(token.LPAREN) {
		p.nextToken()
		ref.Columns = p.parseColumnList()
	}

	return ref
}

func (p *Parser) parseSQLDropObject() *ast.SQLDropObjectStatement {
	stmt := &ast.SQLDropObjectStatement{Token: p.curToken}

	if !p.expectPeek(token.OBJECT) {
		return nil
	}

	// IF EXISTS optionnel
	if p.peekTokenIs(token.IF) {
		p.nextToken() // IF
		p.nextToken() // EXISTS
		stmt.IfExists = true
	}

	// Nom de l'objet
	if !p.expectPeek(token.IDENT) {
		return nil
	}
	stmt.ObjectName = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// CASCADE optionnel
	if p.peekTokenIs(token.CASCADE) {
		p.nextToken()
		stmt.Cascade = true
	}

	return stmt
}

func (p *Parser) parseSQLAlterObject() *ast.SQLAlterObjectStatement {
	stmt := &ast.SQLAlterObjectStatement{Token: p.curToken}

	if !p.expectPeek(token.OBJECT) {
		return nil
	}

	// Nom de l'objet
	if !p.expectPeek(token.IDENT) {
		return nil
	}
	stmt.ObjectName = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// Actions
	for !p.curTokenIs(token.SEMICOLON) && !p.curTokenIs(token.EOF) {
		action := p.parseSQLAlterAction()
		if action != nil {
			stmt.Actions = append(stmt.Actions, action)
		}
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseSQLAlterAction() *ast.SQLAlterAction {
	action := &ast.SQLAlterAction{Token: p.curToken}

	switch p.curToken.Type {
	case token.ADD:
		action.Type = "ADD"
		p.nextToken()
		if p.curTokenIs(token.CONSTRAINT) {
			action.Constraint = p.parseSQLConstraint()
		} else {
			action.Column = p.parseSQLColumnDefinition()
		}
	case token.MODIFY:
		action.Type = "MODIFY"
		p.nextToken()
		action.Column = p.parseSQLColumnDefinition()
	case token.DROP:
		action.Type = "DROP"
		p.nextToken()
		if p.curTokenIs(token.CONSTRAINT) {
			p.nextToken()
			action.Constraint = &ast.SQLConstraint{
				Token: p.curToken,
				Name:  &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal},
			}
		} else {
			action.ColumnName = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		}
	}

	return action
}

func (p *Parser) parseSQLInsert() *ast.SQLInsertStatement {
	stmt := &ast.SQLInsertStatement{Token: p.curToken}

	if !p.expectPeek(token.INTO) {
		return nil
	}

	// Nom de l'objet
	if !p.expectPeek(token.IDENT) {
		return nil
	}
	stmt.ObjectName = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// Colonnes optionnelles
	if p.peekTokenIs(token.LPAREN) {
		p.nextToken()
		stmt.Columns = p.parseColumnList()
	}

	// VALUES ou SELECT
	if p.peekTokenIs(token.VALUES) {
		p.nextToken()
		stmt.Values = p.parseSQLValues()
	} else if p.peekTokenIs(token.SELECT) {
		p.nextToken()
		stmt.Select = p.parseSQLSelectStatement().(*ast.SQLSelectStatement)
	}

	return stmt
}

func (p *Parser) parseSQLValues() []*ast.SQLValues {
	var valuesList []*ast.SQLValues

	for !p.curTokenIs(token.SEMICOLON) && !p.curTokenIs(token.EOF) {
		if p.curTokenIs(token.LPAREN) {
			values := &ast.SQLValues{Token: p.curToken}
			p.nextToken()

			for !p.curTokenIs(token.RPAREN) && !p.curTokenIs(token.EOF) {
				expr := p.parseExpression(LOWEST)
				if expr != nil {
					values.Values = append(values.Values, expr)
				}
				if p.peekTokenIs(token.COMMA) {
					p.nextToken()
				}
				p.nextToken()
			}
			valuesList = append(valuesList, values)
		}
		p.nextToken()
	}

	return valuesList
}

func (p *Parser) parseSQLUpdate() *ast.SQLUpdateStatement {
	stmt := &ast.SQLUpdateStatement{Token: p.curToken}

	// Nom de l'objet
	if !p.expectPeek(token.IDENT) {
		return nil
	}
	stmt.ObjectName = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.SET) {
		return nil
	}

	p.nextToken()

	// Clauses SET
	for !p.curTokenIs(token.WHERE) && !p.curTokenIs(token.SEMICOLON) && !p.curTokenIs(token.EOF) {
		setClause := &ast.SQLSetClause{Token: p.curToken}
		setClause.Column = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

		if !p.expectPeek(token.ASSIGN) {
			return nil
		}

		p.nextToken()
		setClause.Value = p.parseExpression(LOWEST)
		stmt.Set = append(stmt.Set, setClause)

		if p.peekTokenIs(token.COMMA) {
			p.nextToken()
		}
		p.nextToken()
	}

	// WHERE optionnel
	if p.curTokenIs(token.WHERE) {
		p.nextToken()
		stmt.Where = p.parseExpression(LOWEST)
	}

	return stmt
}

func (p *Parser) parseSQLDelete() *ast.SQLDeleteStatement {
	stmt := &ast.SQLDeleteStatement{Token: p.curToken}

	if !p.expectPeek(token.FROM) {
		return nil
	}

	// Nom de l'objet
	if !p.expectPeek(token.IDENT) {
		return nil
	}
	stmt.From = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// WHERE optionnel
	if p.peekTokenIs(token.WHERE) {
		p.nextToken()
		p.nextToken()
		stmt.Where = p.parseExpression(LOWEST)
	}

	return stmt
}

func (p *Parser) parseSQLTruncate() *ast.SQLTruncateStatement {
	stmt := &ast.SQLTruncateStatement{Token: p.curToken}

	if !p.expectPeek(token.OBJECT) {
		return nil
	}

	// Nom de l'objet
	if !p.expectPeek(token.IDENT) {
		return nil
	}
	stmt.ObjectName = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	return stmt
}

func (p *Parser) parseSQLCreateIndex() *ast.SQLCreateIndexStatement {
	stmt := &ast.SQLCreateIndexStatement{Token: p.curToken}

	// UNIQUE optionnel
	if p.peekTokenIs(token.UNIQUE) {
		p.nextToken()
		stmt.Unique = true
	}

	if !p.expectPeek(token.INDEX) {
		return nil
	}

	// Nom de l'index
	if !p.expectPeek(token.IDENT) {
		return nil
	}
	stmt.IndexName = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.ON) {
		return nil
	}

	// Nom de l'objet
	if !p.expectPeek(token.IDENT) {
		return nil
	}
	stmt.ObjectName = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	stmt.Columns = p.parseColumnList()

	return stmt
}

// // Mettre à jour parseSQLSelect pour supporter les clauses avancées
// func (p *Parser) parseSQLSelectStatement() ast.Statement {
// 	selectStmt := &ast.SQLSelectStatement{Token: p.curToken}

// 	// DISTINCT optionnel
// 	if p.peekTokenIs(token.DISTINCT) {
// 		p.nextToken()
// 		selectStmt.Distinct = true
// 	}

// 	p.nextToken()
// 	selectStmt.Select = p.parseSelectList()

// 	// FROM
// 	if !p.expectPeek(token.FROM) {
// 		return nil
// 	}
// 	p.nextToken()
// 	selectStmt.From = p.parseExpression(LOWEST)

// 	// JOINs optionnels
// 	for p.peekTokenIs(token.JOIN) ||
// 		(p.peekTokenIs(token.IDENT) &&
// 			(p.peekToken.Literal == "INNER" || p.peekToken.Literal == "LEFT" ||
// 				p.peekToken.Literal == "RIGHT" || p.peekToken.Literal == "FULL")) {
// 		p.nextToken()
// 		join := &ast.SQLJoin{Token: p.curToken}

// 		if p.curTokenIs(token.IDENT) {
// 			join.Type = p.curToken.Literal
// 			if !p.expectPeek(token.JOIN) {
// 				return nil
// 			}
// 			p.nextToken()
// 		} else {
// 			join.Type = "INNER"
// 		}

// 		join.Table = p.parseExpression(LOWEST)

// 		if !p.expectPeek(token.ON) {
// 			return nil
// 		}
// 		p.nextToken()
// 		join.On = p.parseExpression(LOWEST)

// 		selectStmt.Joins = append(selectStmt.Joins, join)
// 	}

// 	// WHERE optionnel
// 	if p.peekTokenIs(token.WHERE) {
// 		p.nextToken()
// 		p.nextToken()
// 		selectStmt.Where = p.parseExpression(LOWEST)
// 	}

// 	// GROUP BY optionnel
// 	if p.peekTokenIs(token.GROUP) {
// 		p.nextToken() // GROUP
// 		if !p.expectPeek(token.BY) {
// 			return nil
// 		}
// 		p.nextToken()
// 		selectStmt.GroupBy = p.parseExpressionList(token.HAVING, token.ORDER, token.LIMIT)
// 	}

// 	// HAVING optionnel
// 	if p.peekTokenIs(token.HAVING) {
// 		p.nextToken()
// 		p.nextToken()
// 		selectStmt.Having = p.parseExpression(LOWEST)
// 	}

// 	// ORDER BY optionnel
// 	if p.peekTokenIs(token.ORDER) {
// 		p.nextToken() // ORDER
// 		if !p.expectPeek(token.BY) {
// 			return nil
// 		}
// 		p.nextToken()
// 		selectStmt.OrderBy = p.parseOrderByList()
// 	}

// 	// LIMIT optionnel
// 	if p.peekTokenIs(token.LIMIT) {
// 		p.nextToken()
// 		p.nextToken()
// 		selectStmt.Limit = p.parseExpression(LOWEST)
// 	}

// 	// OFFSET optionnel
// 	if p.peekTokenIs(token.OFFSET) {
// 		p.nextToken()
// 		p.nextToken()
// 		selectStmt.Offset = p.parseExpression(LOWEST)
// 	}

// 	// UNION optionnel
// 	if p.peekTokenIs(token.UNION) {
// 		p.nextToken()
// 		if p.peekTokenIs(token.ALL) {
// 			p.nextToken()
// 			selectStmt.UnionAll = true
// 		}
// 		p.nextToken()
// 		selectStmt.Union = p.parseSQLSelectStatement().(*ast.SQLSelectStatement)
// 	}

// 	return selectStmt
// }

func (p *Parser) parseOrderByList() []*ast.SQLOrderBy {
	var orderByList []*ast.SQLOrderBy

	orderBy := &ast.SQLOrderBy{}
	orderBy.Expression = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.ASC) || p.peekTokenIs(token.DESC) {
		p.nextToken()
		orderBy.Direction = p.curToken.Literal
	}

	orderByList = append(orderByList, orderBy)

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()

		orderBy := &ast.SQLOrderBy{}
		orderBy.Expression = p.parseExpression(LOWEST)

		if p.peekTokenIs(token.ASC) || p.peekTokenIs(token.DESC) {
			p.nextToken()
			orderBy.Direction = p.curToken.Literal
		}

		orderByList = append(orderByList, orderBy)
	}

	return orderByList
}

func (p *Parser) parseExpressionList(stopTokens ...token.TokenType) []ast.Expression {
	var expressions []ast.Expression

	expressions = append(expressions, p.parseExpression(LOWEST))

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		expressions = append(expressions, p.parseExpression(LOWEST))
	}

	return expressions
}

func (p *Parser) parseSelectList() []ast.Expression {
	var expressions []ast.Expression

	if p.curTokenIs(token.ASTERISK) {
		expressions = append(expressions, &ast.Identifier{
			Token: p.curToken,
			Value: "*",
		})
		p.nextToken()
		return expressions
	}

	expressions = append(expressions, p.parseExpression(LOWEST))
	if p.curToken.Type == token.IDENT && p.peekToken.Type == token.DOT {
		tok := p.curToken
		tok.Literal = tok.Literal + "."
		p.nextToken() //lecture .
		p.nextToken() //
		tok.Literal = tok.Literal + "." + p.curToken.Literal
		exp := p.parseExpression(LOWEST)
		expressions[len(expressions)-1] = exp
	}

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		expressions = append(expressions, p.parseExpression(LOWEST))
	}

	return expressions
}

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
	}
	p.peekError(t)
	return false
}

func (p *Parser) Errors() []ParserError {
	return p.errors
}

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("Attendu %s, got %s", t, p.peekToken.Type)
	p.errors = append(p.errors, *Create(msg, p.peekToken.Line, p.peekToken.Column))
}

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("Aucune fonction prefix parse pour %s", t)
	p.errors = append(p.errors, *Create(msg, p.peekToken.Line, p.peekToken.Column))
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

var precedences = map[token.TokenType]int{
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.LTE:      LESSGREATER,
	token.GTE:      LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.SLASH:    PRODUCT,
	token.ASTERISK: PRODUCT,
	token.MOD:      PRODUCT,
	token.LBRACKET: INDEX,
	token.CONCAT:   SUM,
	token.IN:       EQUALS,
	token.NOT:      EQUALS,
	token.AS:       AS,
	token.DOT:      DOT,
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func (p *Parser) parseSQLWithStatement() *ast.SQLWithStatement {
	stmt := &ast.SQLWithStatement{Token: p.curToken}

	// RECURSIVE optionnel
	if p.peekTokenIs(token.RECURSIVE) {
		p.nextToken()
		stmt.Recursive = true
	}

	// CTEs
	stmt.CTEs = p.parseCTEList()

	// Requête principale
	if !p.expectPeek(token.SELECT) {
		return nil
	}
	p.nextToken()
	stmt.Select = p.parseSQLSelectStatement().(*ast.SQLSelectStatement)

	return stmt
}

func (p *Parser) parseCTEList() []*ast.SQLCommonTableExpression {
	var ctes []*ast.SQLCommonTableExpression

	cte := p.parseCommonTableExpression()
	if cte != nil {
		ctes = append(ctes, cte)
	}

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		cte = p.parseCommonTableExpression()
		if cte != nil {
			ctes = append(ctes, cte)
		}
	}

	return ctes
}

func (p *Parser) parseCommonTableExpression() *ast.SQLCommonTableExpression {
	cte := &ast.SQLCommonTableExpression{Token: p.curToken}
	cte.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// Colonnes optionnelles
	if p.peekTokenIs(token.LPAREN) {
		p.nextToken()
		cte.Columns = p.parseColumnList()
	}

	if !p.expectPeek(token.AS) {
		return nil
	}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	p.nextToken()

	if p.curTokenIs(token.SELECT) {
		cte.Query = p.parseSQLSelectStatement().(*ast.SQLSelectStatement)
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return cte
}

func (p *Parser) parseRecursiveCTE() *ast.SQLRecursiveCTE {
	cte := &ast.SQLRecursiveCTE{Token: p.curToken}
	cte.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// Colonnes optionnelles
	if p.peekTokenIs(token.LPAREN) {
		p.nextToken()
		cte.Columns = p.parseColumnList()
	}

	if !p.expectPeek(token.AS) {
		return nil
	}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	p.nextToken()

	// Partie anchor
	if p.curTokenIs(token.SELECT) {
		cte.Anchor = p.parseSQLSelectStatement().(*ast.SQLSelectStatement)
	}

	// UNION ou UNION ALL
	if p.peekTokenIs(token.UNION) {
		p.nextToken()
		if p.peekTokenIs(token.ALL) {
			p.nextToken()
			cte.UnionAll = true
		}

		// Partie récursive
		p.nextToken()
		if p.curTokenIs(token.SELECT) {
			cte.Recursive = p.parseSQLSelectStatement().(*ast.SQLSelectStatement)
		}
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return cte
}

func (p *Parser) parseWindowFunction() ast.Expression {
	function := &ast.SQLWindowFunction{Token: p.curToken, Name: p.curToken.Literal}

	if p.peekTokenIs(token.LPAREN) {
		p.nextToken()
		if !p.curTokenIs(token.RPAREN) {
			function.Arguments = p.parseExpressionList(token.RPAREN)
		}
		if !p.expectPeek(token.RPAREN) {
			return nil
		}
	}

	if p.peekTokenIs(token.OVER) {
		p.nextToken()
		function.Over = p.parseWindowClause()
	}

	return function
}

func (p *Parser) parseWindowClause() *ast.SQLWindowClause {
	clause := &ast.SQLWindowClause{Token: p.curToken}

	if !p.expectPeek(token.LPAREN) {
		// Peut être un nom de fenêtre prédéfini
		if p.curTokenIs(token.IDENT) {
			clause.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
			return clause
		}
		return nil
	}

	p.nextToken()

	// PARTITION BY optionnel
	if p.curTokenIs(token.PARTITION) {
		if !p.expectPeek(token.BY) {
			return nil
		}
		p.nextToken()
		clause.Partition = p.parseExpressionList(token.ORDER, token.ROWS, token.RANGE)
	}

	// ORDER BY optionnel
	if p.curTokenIs(token.ORDER) {
		if !p.expectPeek(token.BY) {
			return nil
		}
		p.nextToken()
		clause.OrderBy = p.parseOrderByList()
	}

	// Frame optionnel
	if p.curTokenIs(token.ROWS) || p.curTokenIs(token.RANGE) {
		clause.Frame = p.parseWindowFrame()
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return clause
}

func (p *Parser) parseWindowFrame() *ast.SQLWindowFrame {
	frame := &ast.SQLWindowFrame{Token: p.curToken, Type: p.curToken.Literal}

	if !p.expectPeek(token.BETWEEN) {
		return nil
	}

	p.nextToken()
	frame.Start = p.parseWindowFrameBound()

	if !p.expectPeek(token.AND) {
		return nil
	}

	p.nextToken()

	if p.curTokenIs(token.CURRENT) {
		frame.End = &ast.SQLWindowFrameBound{Token: p.curToken, Type: "ROW"}
		p.nextToken() // ROW
	} else {
		frame.End = p.parseWindowFrameBound()
	}

	return frame
}

func (p *Parser) parseWindowFrameBound() *ast.SQLWindowFrameBound {
	bound := &ast.SQLWindowFrameBound{Token: p.curToken}

	if p.curTokenIs(token.UNBOUNDED) {
		bound.Unbounded = true
		p.nextToken()
		bound.Type = p.curToken.Literal // PRECEDING ou FOLLOWING
	} else if p.curTokenIs(token.CURRENT) {
		bound.Type = "ROW"
		p.nextToken() // ROW
	} else {
		// Expression numérique
		bound.Value = p.parseExpression(LOWEST)
		p.nextToken()
		bound.Type = p.curToken.Literal // PRECEDING ou FOLLOWING
	}

	return bound
}

func (p *Parser) parseHierarchicalQuery() *ast.SQLHierarchicalQuery {
	hierarchical := &ast.SQLHierarchicalQuery{Token: p.curToken}

	// START WITH optionnel
	if p.curTokenIs(token.START) {
		if !p.expectPeek(token.WITH) {
			return nil
		}
		p.nextToken()
		hierarchical.StartWith = p.parseExpression(LOWEST)
	}

	// CONNECT BY
	if p.curTokenIs(token.CONNECT) {
		if !p.expectPeek(token.BY) {
			return nil
		}
		p.nextToken()

		// PRIOR optionnel
		if p.curTokenIs(token.PRIOR) {
			hierarchical.Prior = true
			p.nextToken()
		}

		hierarchical.ConnectBy = p.parseExpression(LOWEST)
	}

	// NOCYCLE optionnel
	if p.peekTokenIs(token.NOCYCLE) {
		p.nextToken()
		hierarchical.Nocycle = true
	}

	// ORDER SIBLINGS BY optionnel
	if p.peekTokenIs(token.ORDER) {
		p.nextToken()
		if p.peekTokenIs(token.SIBLINGS) {
			p.nextToken()
			hierarchical.OrderSiblings = true
		}
	}

	return hierarchical
}

// Mettre à jour parseSQLSelectStatement pour inclure les fonctionnalités récursives
func (p *Parser) parseSQLSelectStatement() ast.Statement {
	// Vérifier d'abord s'il y a une clause WITH
	if p.curTokenIs(token.WITH) {
		// var selectStmt *ast.SQLSelectStatement
		withStmt := p.parseSQLWithStatement()
		selectStmt := withStmt.Select
		selectStmt.With = withStmt
		return selectStmt //return withStmt //
	}

	selectStmt := &ast.SQLSelectStatement{Token: p.curToken}

	// DISTINCT optionnel
	if p.peekTokenIs(token.DISTINCT) {
		p.nextToken()
		selectStmt.Distinct = true
	}

	p.nextToken()
	selectStmt.Select = p.parseSelectList()

	// FROM
	if !p.expectPeek(token.FROM) {
		return nil
	}
	p.nextToken()
	selectStmt.From = p.parseExpression(LOWEST)

	// JOINs optionnels
	for p.peekTokenIs(token.JOIN) ||
		(p.peekTokenIs(token.IDENT) &&
			(p.peekToken.Literal == "INNER" || p.peekToken.Literal == "LEFT" ||
				p.peekToken.Literal == "RIGHT" || p.peekToken.Literal == "FULL")) {
		p.nextToken()
		join := &ast.SQLJoin{Token: p.curToken}

		if p.curTokenIs(token.IDENT) {
			join.Type = p.curToken.Literal
			if !p.expectPeek(token.JOIN) {
				return nil
			}
			p.nextToken()
		} else {
			join.Type = "INNER"
		}

		join.Table = p.parseExpression(LOWEST)

		if !p.expectPeek(token.ON) {
			return nil
		}
		p.nextToken()
		join.On = p.parseExpression(LOWEST)

		selectStmt.Joins = append(selectStmt.Joins, join)
	}

	// WHERE optionnel
	if p.peekTokenIs(token.WHERE) {
		p.nextToken()
		p.nextToken()
		selectStmt.Where = p.parseExpression(LOWEST)
	}

	// Clause hiérarchique CONNECT BY optionnelle
	if p.peekTokenIs(token.CONNECT) || p.peekTokenIs(token.START) {
		p.nextToken()
		selectStmt.Hierarchical = p.parseHierarchicalQuery()
	}

	// GROUP BY optionnel
	if p.peekTokenIs(token.GROUP) {
		p.nextToken() // GROUP
		if !p.expectPeek(token.BY) {
			return nil
		}
		p.nextToken()
		selectStmt.GroupBy = p.parseExpressionList(token.HAVING, token.ORDER, token.LIMIT, token.WINDOW)
	}

	// HAVING optionnel
	if p.peekTokenIs(token.HAVING) {
		p.nextToken()
		p.nextToken()
		selectStmt.Having = p.parseExpression(LOWEST)
	}

	// WINDOW optionnel (définitions de fenêtres nommées)
	if p.peekTokenIs(token.WINDOW) {
		p.nextToken()
		selectStmt.WindowClauses = p.parseWindowDefinitions()
	}

	// ORDER BY optionnel
	if p.peekTokenIs(token.ORDER) {
		p.nextToken() // ORDER
		if !p.expectPeek(token.BY) {
			return nil
		}
		p.nextToken()
		selectStmt.OrderBy = p.parseOrderByList()
	}

	// LIMIT optionnel
	if p.peekTokenIs(token.LIMIT) {
		p.nextToken()
		p.nextToken()
		selectStmt.Limit = p.parseExpression(LOWEST)
	}

	// OFFSET optionnel
	if p.peekTokenIs(token.OFFSET) {
		p.nextToken()
		p.nextToken()
		selectStmt.Offset = p.parseExpression(LOWEST)
	}

	// UNION optionnel
	if p.peekTokenIs(token.UNION) {
		p.nextToken()
		if p.peekTokenIs(token.ALL) {
			p.nextToken()
			selectStmt.UnionAll = true
		}
		p.nextToken()
		selectStmt.Union = p.parseSQLSelectStatement().(*ast.SQLSelectStatement)
	}

	return selectStmt
}

func (p *Parser) parseWindowDefinitions() []*ast.SQLWindowClause {
	var windows []*ast.SQLWindowClause

	window := p.parseWindowClause()
	if window != nil {
		windows = append(windows, window)
	}

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		window = p.parseWindowClause()
		if window != nil {
			windows = append(windows, window)
		}
	}

	return windows
}

func (p *Parser) parseArrayLiteral() ast.Expression {
	array := &ast.ArrayLiteral{Token: p.curToken}

	if p.peekTokenIs(token.RBRACKET) {
		p.nextToken()
		return array
	}

	p.nextToken()
	array.Elements = append(array.Elements, p.parseExpression(LOWEST))

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		array.Elements = append(array.Elements, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(token.RBRACKET) {
		return nil
	}

	return array
}

func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	exp := &ast.IndexExpression{Token: p.curToken, Left: left}

	p.nextToken()
	exp.Index = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RBRACKET) {
		return nil
	}

	return exp
}

func (p *Parser) parseSliceExpression(left ast.Expression) ast.Expression {
	exp := &ast.SliceExpression{Token: p.curToken, Left: left}

	p.nextToken()

	if !p.curTokenIs(token.COLON) {
		exp.Start = p.parseExpression(LOWEST)
	}

	if p.curTokenIs(token.COLON) {
		p.nextToken()

		if !p.curTokenIs(token.RBRACKET) {
			exp.End = p.parseExpression(LOWEST)
		}
	}

	if !p.expectPeek(token.RBRACKET) {
		return nil
	}

	return exp
}

func (p *Parser) parseArrayType() *ast.ArrayType {
	arrayType := &ast.ArrayType{Token: p.curToken}

	// Taille optionnelle [n]
	if p.peekTokenIs(token.LBRACKET) {
		p.nextToken() // [
		p.nextToken()
		arrayType.Size = p.parseIntegerLiteral().(*ast.IntegerLiteral)
		if !p.expectPeek(token.RBRACKET) {
			return nil
		}
	}

	if !p.expectPeek(token.OF) {
		return nil
	}

	p.nextToken()
	arrayType.ElementType = p.parseTypeAnnotation()

	return arrayType
}

func (p *Parser) parseInExpression(left ast.Expression) ast.Expression {
	exp := &ast.InExpression{Token: p.curToken, Left: left}

	// Vérifier NOT IN
	if p.curTokenIs(token.NOT) {
		exp.Not = true
		if !p.expectPeek(token.IN) {
			return nil
		}
	}

	p.nextToken()
	exp.Right = p.parseExpression(LOWEST)

	return exp
}

func (p *Parser) parseAsExpression(left ast.Expression) ast.Expression {
	exp := &ast.InExpression{Token: p.curToken, Left: left}
	p.nextToken()
	exp.Right = p.parseExpression(LOWEST)
	return exp
}

func (p *Parser) parseArrayFunctionCall() ast.Expression {
	call := &ast.ArrayFunctionCall{Token: p.curToken}
	call.Function = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	p.nextToken()
	call.Array = p.parseExpression(LOWEST)

	// Arguments optionnels
	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		call.Arguments = append(call.Arguments, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return call
}

func (p *Parser) parseTypeAnnotation() *ast.TypeAnnotation {
	ta := &ast.TypeAnnotation{Token: p.curToken}

	if p.curTokenIs(token.ARRAY) {
		ta.ArrayType = p.parseArrayType()
		return ta
	}

	ta.Type = p.curToken.Literal
	var pe *ParserError

	// Vérifier les contraintes
	if p.peekTokenIs(token.LPAREN) || p.peekTokenIs(token.LBRACKET) {
		ta.Constraints, pe = p.parseTypeConstraints()
		if pe != nil {
			p.errors = append(p.errors, *pe)
		}
	}

	return ta
}

// Nouvelle méthode pour gérer à la fois l'index et le slice
func (p *Parser) parseIndexOrSliceExpression(left ast.Expression) ast.Expression {
	// Sauvegarder la position pour vérifier si c'est un slice
	_, currentPosition := p.l.GetCursorPosition()

	p.nextToken()

	// Vérifier si c'est un slice (contient :)
	isSlice := false
	// tempPosition := p.l.position
	// tempToken := p.curToken

	// Avancer pour vérifier
	for !p.curTokenIs(token.RBRACKET) && !p.curTokenIs(token.EOF) {
		if p.curTokenIs(token.COLON) {
			isSlice = true
			break
		}
		p.nextToken()
	}

	p.l.SetCursorPosition(currentPosition, currentPosition+1)

	p.curToken = token.Token{Type: token.LBRACKET, Literal: "["}

	if isSlice {
		return p.parseSliceExpression(left)
	}
	return p.parseIndexExpression(left)
}
