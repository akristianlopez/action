package parser

import (
	"fmt"
	"strconv"
	"strings"

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

var _position, _currentPosition int
var _scr, _spk *token.Token = nil, nil

func (p *Parser) Save() {
	if p.l == nil {
		return
	}
	p.l.SaveCnt()
	_scr = &token.Token{Type: p.curToken.Type, Literal: p.curToken.Literal,
		Line: p.curToken.Line, Column: p.curToken.Column}
	_spk = &token.Token{Type: p.peekToken.Type, Literal: p.peekToken.Literal,
		Line: p.peekToken.Line, Column: p.peekToken.Column}
}
func (p *Parser) Restore() {
	if p.l == nil {
		return
	}
	p.l.RestoreCnt()
	if _scr != nil && _spk != nil {
		p.curToken = *_scr
		_scr = nil
		p.peekToken = *_spk
		_spk = nil
	}
}
func (p *Parser) Clear() {
	_scr = nil
	_spk = nil
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
	// p.registerPrefix(token.LENGTH, p.parseArrayFunctionCall)
	// p.registerPrefix(token.APPEND, p.parseArrayFunctionCall)
	// p.registerPrefix(token.PREPEND, p.parseArrayFunctionCall)
	// p.registerPrefix(token.REMOVE, p.parseArrayFunctionCall)
	p.registerPrefix(token.NULL, p.parseNullLiteral)
	p.registerPrefix(token.NOT, p.parsePrefixExpression)
	// p.registerPrefix(token.SLICE, p.parseArrayFunctionCall)
	// p.registerPrefix(token.CONTAINS, p.parseArrayFunctionCall)
	p.registerPrefix(token.DURATION_LIT, p.parseDurationLiteral)
	p.registerPrefix(token.LBRACE, p.parseStructLiteral)

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
	// p.registerInfix(token.AS, p.parseInfixExpression)
	p.registerInfix(token.IS, p.parseInfixExpression)
	p.registerInfix(token.DOT, p.parsePropertyAccess)
	p.registerInfix(token.BETWEEN, p.parseBetweenExpression)
	p.registerInfix(token.LIKE, p.parseLikeExpression)

	// p.registerInfix(token.CONCAT, p.parseInfixExpression)

	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) parseDuration(s string) *ast.Duration {
	duration := ast.Duration{Years: 0, Months: 0, Days: 0, Hours: 0, Minutes: 0, Seconds: 0, Nanos: 0}

	return &duration
}
func (p *Parser) parseDurationLiteral() ast.Expression {
	return &ast.DurationLiteral{
		Token:          p.curToken,
		Value:          p.curToken.Literal,
		ParsedDuration: p.parseDuration(p.curToken.Literal),
	}
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

func (p *Parser) ParseAction() *ast.Action {
	program := &ast.Action{}

	// Vérifier que le programme commence par 'action'
	if !p.curTokenIs(token.ACTION) {
		p.errors = append(p.errors, *Create("The action must start with the word 'action'", p.curToken.Line, p.curToken.Column))
		return program
	}

	p.nextToken() //move to name

	// Lire le nom de l'action
	if !p.curTokenIs(token.STRING_LIT) {
		p.errors = append(p.errors, *Create("Attendu un nom d'action après 'action'", p.curToken.Line, p.curToken.Column))
		return program
	}
	program.ActionName = p.curToken.Literal
	if !p.expectPeek(token.LPAREN) {
		return program
	}
	p.nextToken() //move to name

	//Parser les arguments de l'action
	for !p.curTokenIs(token.START, token.EOF, token.RPAREN) {
		field := &ast.StructField{Token: p.curToken}
		field.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

		if !p.expectPeek(token.COLON) {
			return program
		}

		p.nextToken()
		field.Type = p.parseTypeAnnotation()
		program.Paramters = append(program.Paramters, field)
		if !p.expectPeekEx(token.COMMA, token.RPAREN) {
			return program
		}
		if p.curTokenIs(token.COMMA) {
			p.nextToken()
		}
	}
	if !p.curTokenIs(token.RPAREN) {
		return program
	}
	if p.peekTokenIs(token.COLON) {
		p.nextToken() // :
		//Parser le type de retour de l'action
		p.nextToken()
		program.ReturnType = p.parseTypeAnnotation()
	} else {
		program.ReturnType = nil
	}
	p.nextToken()

	// Parser les déclarations jusqu'à 'start'
	for !p.curTokenIs(token.START) && !p.curTokenIs(token.EOF) {
		stmt, pe := p.parseStatement(false)
		if pe != nil {
			p.errors = append(p.errors, *pe)
		}
		if stmt != nil {
			if arr, ok := stmt.(*ast.LetStatements); ok {
				for _, val := range *arr {
					program.Statements = append(program.Statements, &val)
				}
			} else {
				program.Statements = append(program.Statements, stmt)
			}

		}
		p.nextToken()
	}

	// Parser les instructions après 'start'
	if p.curTokenIs(token.START) {
		p.nextToken()
		for !p.curTokenIs(token.STOP) && !p.curTokenIs(token.EOF) {
			stmt, pe := p.parseStatement(true)
			if pe != nil {
				p.errors = append(p.errors, *pe)
			}
			if stmt != nil {
				if arr, ok := stmt.(*ast.LetStatements); ok && arr != nil {
					for _, val := range *arr {
						program.Statements = append(program.Statements, &val)
					}
				} else {
					program.Statements = append(program.Statements, stmt)
				}
			}
			p.nextToken()
		}
	}

	return program
}
func (p *Parser) parseStmStartSection() (ast.Statement, *ParserError) {
	switch p.curToken.Type {
	case token.IF:
		return p.parseIfStatement()
	case token.FOR:
		return p.parseForStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	case token.CREATE:
		if p.peekTokenIs(token.OBJECT) {
			return p.parseSQLCreateObject()
		} else if p.peekTokenIs(token.INDEX, token.UNIQUE) {
			return p.parseSQLCreateIndex()
		}
	case token.DROP:
		if p.peekTokenIs(token.OBJECT) {
			return p.parseSQLDropObject()
		}
		return nil, Create("token 'object' is missing", p.peekToken.Line, p.peekToken.Column)
	case token.ALTER:
		if p.peekTokenIs(token.OBJECT) {
			return p.parseSQLAlterObject()
		}
		return nil, Create("token 'object' is missing", p.peekToken.Line, p.peekToken.Column)
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
		return nil, Create("token 'object' is missing", p.peekToken.Line, p.peekToken.Column)
	case token.SELECT, token.WITH:
		return p.parseSQLSelectStatement()
	case token.LET:
		return p.parseLetStatements()
	case token.TYPE:
		return p.parseStructStatement()
	case token.SWITCH:
		return p.parseSwitchStatement()
	case token.BREAK:
		return p.parseBreakStatement()
	case token.CONTINUE:
		return p.parseContinueStatement()
	case token.FALLTHROUGH:
		return p.parseFallthroughStatement()
	default:
		return p.parseExpressionStatement()
	}
	return nil, Create(fmt.Sprintf("'%s' is not expected", p.curToken.Type), p.curToken.Line, p.curToken.Column)
}

func (p *Parser) parseStmDeclarationSection() (ast.Statement, *ParserError) {
	switch p.curToken.Type {
	case token.LET:
		return p.parseLetStatements()
	case token.FUNCTION:
		return p.parseFunctionStatement()
	case token.TYPE:
		return p.parseStructStatement()
	}
	return nil, Create(fmt.Sprintf("'%s' is not expected", p.curToken.Type), p.curToken.Line, p.curToken.Column)
}
func (p *Parser) parseStatement(startSts bool) (ast.Statement, *ParserError) {
	if startSts {
		return p.parseStmStartSection()
	}
	return p.parseStmDeclarationSection()
}

func (p *Parser) parseLetStatements() (*ast.LetStatements, *ParserError) {
	var stms ast.LetStatements
	stms = make([]ast.LetStatement, 0)
	// stms = &stmts

	tok := p.curToken
	if !p.expectPeek(token.IDENT) {
		return nil, nil //Create("Identifier expected", p.peekToken.Line, p.peekToken.Column)
	}
	// var flag bool
	for p.curTokenIs(token.IDENT) && !p.peekTokenIs(token.EOF) &&
		!p.peekTokenIs(token.SEMICOLON) && !p.peekTokenIs(token.STOP) {
		flag := false //true
		stmt := ast.LetStatement{Token: tok}
		stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		if !p.peekTokenIs(token.COLON) && !p.peekTokenIs(token.ASSIGN) {
			return nil, Create("type expected", p.peekToken.Line, p.peekToken.Column)
		}
		// Vérifier s'il y a une annotation de type
		if p.peekTokenIs(token.COLON) {
			p.nextToken() // :
			p.nextToken() // type
			// if !p.peekTokenIs(token.ASSIGN) {
			_, cur := p.l.GetCursorPosition()
			flag = p.curToken.Type == token.IDENT // type defined by the user
			stmt.Type = p.parseTypeAnnotation()
			_, _cur := p.l.GetCursorPosition()
			flag = flag && cur != _cur
			// }
		}
		if p.peekTokenIs(token.ASSIGN) {
			p.nextToken() // =
			p.nextToken() // valeur
			// flag = true
			stmt.Value = p.parseExpression(LOWEST)
		}
		if p.peekTokenIs(token.COMMA) {
			p.nextToken() // ,
			flag = true
			if !p.expectPeek(token.IDENT) {
				return nil, nil
			}
		}
		stms = append(stms, stmt)
		if !flag {
			// flag = true
			//after reading the type if there is not any other operation that modifies the cursor
			//position, we stopped parsing to go to the next statement
			// p.nextToken()
			break
		}
	}

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return &stms, nil
}

func (p *Parser) parseLetStatement() (*ast.LetStatement, *ParserError) {
	stmt := &ast.LetStatement{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil, nil //Create("Identifier expected", p.peekToken.Line, p.peekToken.Column)
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
			} else if p.peekTokenIs(token.DURATION_LIT) {
				p.nextToken()
				maxLength := p.parseDurationLiteral()
				constraints.IntegerRange = &ast.RangeConstraint{Max: maxLength, Min: nil}
			}
			if !p.expectPeek(token.RPAREN) {
				return nil, nil
			}
		case token.LBRACKET:
			constraints.IntegerRange, pe = p.parseRangeConstraint()
			if pe != nil {
				p.errors = append(p.errors, *pe)
			}
			if !p.expectPeek(token.RBRACKET) {
				return nil, nil //Create("']'", p.peekToken.Line, p.peekToken.Column)
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
		return nil, nil //Create("',' expected", p.peekToken.Line, p.peekToken.Column)
	}
	p.nextToken()

	if !p.peekTokenIs(token.INT_LIT) && !p.peekTokenIs(token.FLOAT_LIT) &&
		!p.peekTokenIs(token.DURATION_LIT) && !p.peekTokenIs(token.DATE_LIT) &&
		!p.peekTokenIs(token.TIME_LIT) { //!p.expectPeek(token.DOT) ||
		return nil, nil //Create("'number' is missing", p.peekToken.Line, p.peekToken.Column)
	}
	p.nextToken()
	rc.Max = p.parseExpression(LOWEST)

	return rc, nil
}

func (p *Parser) parseFunctionStatement() (*ast.FunctionStatement, *ParserError) {
	stmt := &ast.FunctionStatement{Token: p.curToken}
	var pe *ParserError

	if !p.expectPeek(token.IDENT) {
		return nil, nil //Create("Identifier expected", p.peekToken.Line, p.peekToken.Column)
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.LPAREN) {
		return nil, nil //Create("'(' expected", p.peekToken.Line, p.peekToken.Column)
	}

	stmt.Parameters, pe = p.parseFunctionParameters()
	if pe != nil {
		p.errors = append(p.errors, *pe)
	}

	// Type de retour optionnel
	if p.peekTokenIs(token.COLON) {
		p.nextToken() // :
		if !p.expectPeekEx(token.IDENT, token.INTEGER, token.FLOAT,
			token.STRING, token.BOOLEAN, token.DATE, token.DATETIME,
			token.TIME, token.DURATION, token.ARRAY) {
			return nil, nil
		}
		stmt.ReturnType = p.parseTypeAnnotation()
	}

	if !p.expectPeek(token.LBRACE) {
		return nil, nil //Create("'{' expected", p.peekToken.Line, p.peekToken.Column)
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
		return nil, nil //Create("':' expected", p.peekToken.Line, p.peekToken.Column)
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
			return nil, nil //Create("':' expected", p.peekToken.Line, p.peekToken.Column)
		}

		p.nextToken()
		param.Type = p.parseTypeAnnotation()
		params = append(params, param)
	}

	if !p.expectPeek(token.RPAREN) {
		return nil, nil //Create("')' expected", p.peekToken.Line, p.peekToken.Column)
	}

	return params, nil
}

func (p *Parser) parseStructStatement() (*ast.StructStatement, *ParserError) {
	stmt := &ast.StructStatement{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil, nil //Create("Identifier expected", p.peekToken.Line, p.peekToken.Column)
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.STRUCT) {
		return nil, nil //Create("token 'struc' expected", p.peekToken.Line, p.peekToken.Column)
	}

	if !p.expectPeek(token.LBRACE) {
		return nil, nil //Create("'{' expected", p.peekToken.Line, p.peekToken.Column)
	}

	p.nextToken()

	// Parser les champs
	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		field := &ast.StructField{Token: p.curToken}
		field.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

		if !p.expectPeek(token.COLON) {
			return nil, nil //Create("':' expected", p.peekToken.Line, p.peekToken.Column)
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

func (p *Parser) parseIfStatement() (*ast.IfStatement, *ParserError) {
	stmt := &ast.IfStatement{Token: p.curToken}
	var pe *ParserError
	// var reqParen bool = false

	// if !p.expectPeek(token.LPAREN) {
	// 	return nil, nil
	// }
	// if p.peekTokenIs(token.LPAREN) {
	// 	p.nextToken()
	// 	reqParen = true
	// }
	p.nextToken()

	// Condition
	stmt.Condition = p.parseExpression(LOWEST)

	// if reqParen && !p.expectPeek(token.RPAREN) {
	// 	return nil, nil
	// }
	if !p.expectPeek(token.LBRACE) {
		return nil, nil
	}

	fl := false
	// Then
	if !p.peekTokenIs(token.RBRACE) {
		var tp *ast.BlockStatement
		tp, pe = p.parseBlockStatement()
		p.addError(pe)
		stmt.Then = tp
		fl = true
	}
	if !fl && p.peekTokenIs(token.RBRACE) {
		p.nextToken()
	}

	stmt.Else = nil
	if p.peekTokenIs(token.ELSE) {
		var tp *ast.BlockStatement
		p.nextToken() //lecture de Else
		p.nextToken()
		switch p.curToken.Type {
		case token.LBRACE:
			tp, pe = p.parseBlockStatement()
		default:
			var stm ast.Statement
			stm, pe = p.parseStatement(true)
			if stm != nil {
				tp = &ast.BlockStatement{Token: p.curToken}
				tp.Statements = append(tp.Statements, stm)
			}
		}
		p.addError(pe)
		stmt.Else = tp
	}
	return stmt, pe
}

func (p *Parser) parseForStatement() (ast.Statement, *ParserError) {
	stmt := &ast.ForStatement{Token: p.curToken}
	var pe *ParserError
	var reParen, reSecol bool = false, false

	//this modification is important insofar as it allows not to oblige
	//developper to put absolutely the parentheses when they wrote the
	//statement for. But the brace is the one thing obligatory to show
	//the beginning of the list of statements

	//check if the NextToken is '(' if true move to the next token
	p.Save()
	if p.peekTokenIs(token.LPAREN) {
		p.nextToken()
		reParen = true
	}
	if p.peekTokenIs(token.LET) { //Chef if it's a for each statement: for let x of y
		p.nextToken() // let
		if !p.expectPeek(token.IDENT) {
			return nil, nil
		}
		if p.peekTokenIs(token.OF) {
			p.Restore()
			return p.parseForEachStatement()
		}
	}
	p.Restore()

	p.Save()
	if p.peekTokenIs(token.LPAREN) {
		p.nextToken()
		reParen = true
	}
	if !p.peekTokenIs(token.SEMICOLON, token.LET) {
		p.nextToken()
		p.parseExpression(LOWEST)
		if !p.peekTokenIs(token.ASSIGN) {
			p.Restore()
			return p.parseWhileStatement()
		}
	}
	p.Restore()
	if p.peekTokenIs(token.LPAREN) {
		p.nextToken()
		reParen = true
	}
	p.nextToken()

	// Initialisation
	if !p.curTokenIs(token.SEMICOLON) {
		var stm ast.Statement
		stm, pe = p.parseStatement(true) //p.parseLetStatement()
		if pe != nil {
			p.errors = append(p.errors, *pe)
		}
		if stm != nil {
			if arr, ok := stm.(*ast.LetStatements); ok {
				stmt.Init = &(*arr)[0]
			} else {
				stmt.Init = stm
			}
			// stmt.Init = stm
		}
		if !p.curTokenIs(token.SEMICOLON) {
			p.curError(token.SEMICOLON)
			return nil, nil
		}
	}

	if p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
		reSecol = true
	}

	// Condition
	if !p.curTokenIs(token.SEMICOLON) {
		stmt.Condition = p.parseExpression(LOWEST)
	}

	if reSecol && !p.expectPeek(token.SEMICOLON) {
		return nil, nil
	}

	// Update
	if !p.peekTokenIs(token.RPAREN) && !p.peekTokenIs(token.LBRACE) {
		var tp ast.Statement
		p.nextToken()
		tp, pe = p.parseStatement(true)
		if pe != nil {
			p.errors = append(p.errors, *pe)
		}
		stmt.Update = tp
	}

	if reParen && !p.expectPeek(token.RPAREN) {
		return nil, nil
	}
	// if p.peekTokenIs(token.RPAREN) && reParen {
	// 	p.nextToken()
	// }

	if !p.expectPeek(token.LBRACE) {
		return nil, nil //Create("'{' expected", p.peekToken.Line, p.peekToken.Column)
	}

	stmt.Body, pe = p.parseBlockStatement()

	return stmt, pe
}

func (p *Parser) parseForEachStatement() (*ast.ForEachStatement, *ParserError) {
	stmt := &ast.ForEachStatement{Token: p.curToken}
	var (
		pe    *ParserError
		hasLP bool = false
	)

	if p.peekTokenIs(token.LPAREN) {
		hasLP = true
		p.nextToken()
	}
	if !p.expectPeek(token.LET) {
		return nil, nil
	}
	if !p.expectPeek(token.IDENT) {
		return nil, nil
	}
	stmt.Variable = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.OF) {
		return nil, nil
	}
	p.nextToken()
	stmt.Iterator = p.parseExpression(LOWEST)
	if hasLP && !p.expectPeek(token.RPAREN) {
		return nil, nil
	}
	if !p.expectPeek(token.LBRACE) {
		return nil, nil
	}
	stmt.Body, pe = p.parseBlockStatement()

	return stmt, pe
}

func (p *Parser) parseWhileStatement() (*ast.WhileStatement, *ParserError) {
	stmt := &ast.WhileStatement{Token: p.curToken}
	var pe *ParserError
	var reqParen bool = false
	//check if the NextToken is '(' if true move to the next token
	if p.peekTokenIs(token.LPAREN) {
		p.nextToken()
		reqParen = true
	}
	p.nextToken()

	// Condition
	stmt.Condition = p.parseExpression(LOWEST)

	if reqParen && !p.expectPeek(token.RPAREN) {
		// p.nextToken()
		return nil, nil
	}

	if !p.expectPeek(token.LBRACE) {
		return nil, nil
	}

	stmt.Body, pe = p.parseBlockStatement()

	return stmt, pe
}

func (p *Parser) parseReturnStatement() (*ast.ReturnStatement, *ParserError) {
	stmt := &ast.ReturnStatement{Token: p.curToken}

	p.nextToken()

	if !p.curTokenIs(token.SEMICOLON) {
		switch p.curToken.Type {
		case token.LBRACE:
			stmt.ReturnValue = p.parseStructLiteral()
		// case token.LBRACKET:
		// 	stmt.ReturnValue = p.parseArrayLiteral()
		default:
			stmt.ReturnValue = p.parseExpression(LOWEST)
		}
		// stmt.ReturnValue = p.parseExpression(LOWEST)
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
	// flag := len(arg) == 0
	if !p.curTokenIs(token.LBRACE) {
		p.curError(token.LBRACE)
		return nil, nil
	}
	p.nextToken()

	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		stmt, pe := p.parseStatement(true)
		if pe != nil {
			p.errors = append(p.errors, *pe)
		}
		if stmt != nil {
			if arr, ok := stmt.(*ast.LetStatements); ok {
				for _, val := range *arr {
					block.Statements = append(block.Statements, &val)
				}
			} else {
				block.Statements = append(block.Statements, stmt)
			}

		}
		p.nextToken()
	}
	if !p.curTokenIs(token.RBRACE) {
		p.curError(token.RBRACE)
		return nil, nil
	}
	return block, nil
}

func (p *Parser) parseExpressionStatement() (ast.Statement, *ParserError) {
	stmt := &ast.ExpressionStatement{Token: p.curToken}

	stmt.Expression = p.parseExpression(LOWEST)
	if p.peekTokenIs(token.ASSIGN) {
		stm := &ast.AssignmentStatement{Token: p.peekToken}
		stm.Variable = stmt.Expression
		p.nextToken() //read =
		p.nextToken() //read next token
		switch p.curToken.Type {
		case token.LBRACE:
			stm.Value = p.parseStructLiteral()
		case token.LBRACKET:
			stm.Value = p.parseArrayLiteral()
		default:
			stm.Value = p.parseExpression(LOWEST)
		}
		if p.peekTokenIs(token.SEMICOLON) {
			p.nextToken()
		}
		return stm, nil
	}
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt, nil
}

func (p *Parser) parseSwitchStatement() (*ast.SwitchStatement, *ParserError) {
	stmt := &ast.SwitchStatement{Token: p.curToken}

	// Expression du switch
	if !p.expectPeek(token.LPAREN) {
		return nil, nil //Create("'(' expected", p.peekToken.Line, p.peekToken.Column)
	}

	p.nextToken()
	stmt.Expression = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil, nil
	}

	if !p.expectPeek(token.LBRACE) {
		return nil, nil
	}

	p.nextToken()

	// Parser les cases et le default
	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.START) &&
		!p.curTokenIs(token.STOP) && !p.curTokenIs(token.EOF) {
		switch p.curToken.Type {
		case token.CASE:
			caseStmt := p.parseSwitchCase()
			if caseStmt != nil {
				stmt.Cases = append(stmt.Cases, caseStmt)
			}
		case token.DEFAULT:
			if stmt.DefaultCase != nil {
				// p.errors = append(p.errors, "Multiple default cases in switch")
				return nil, Create("Multiple default cases in switch", p.peekToken.Line, p.peekToken.Column)
			}
			stmt.DefaultCase = p.parseDefaultCase()
		default:
			// p.errors = append(p.errors, fmt.Sprintf("Unexpected token in switch: %s", p.curToken.Type))
			return nil, Create(fmt.Sprintf("Unexpected token in switch: %s", p.curToken.Type), p.peekToken.Line, p.peekToken.Column)
		}
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseSwitchCase() *ast.SwitchCase {
	if !p.curTokenIs(token.CASE) {
		p.curError(token.CASE)
		return nil
	}
	caseStmt := &ast.SwitchCase{Token: p.curToken}

	p.nextToken()

	// Parser les expressions du case (peut être multiple avec virgule)
	caseStmt.Expressions = append(caseStmt.Expressions, p.parseExpression(LOWEST))

	for p.peekTokenIs(token.COMMA) {
		p.nextToken() // ,
		p.nextToken()
		caseStmt.Expressions = append(caseStmt.Expressions, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(token.COLON) {
		return nil
	}

	p.nextToken()

	// Parser le body du case
	caseStmt.Body = &ast.BlockStatement{Token: p.curToken}

	for !p.curTokenIs(token.CASE) && !p.curTokenIs(token.DEFAULT) &&
		!p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		stmt, pe := p.parseStatement(true)
		p.addError(pe)
		if stmt != nil {
			caseStmt.Body.Statements = append(caseStmt.Body.Statements, stmt)
		}
		p.Save()
		p.nextToken()
	}
	p.Restore()
	return caseStmt
}

func (p *Parser) parseDefaultCase() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken}

	if !p.expectPeek(token.COLON) {
		return nil
	}
	p.nextToken()

	for !p.curTokenIs(token.CASE) && !p.curTokenIs(token.DEFAULT) &&
		!p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		stmt, pe := p.parseStatement(true)
		p.addError(pe)
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.Save()
		p.nextToken()
	}

	// if !p.curTokenIs(token.RBRACE) {
	// 	p.Restore()
	// }
	// p.Clear()
	p.Restore()
	return block
}

func (p *Parser) parseBreakStatement() (*ast.BreakStatement, *ParserError) {
	stmt := &ast.BreakStatement{Token: p.curToken}

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseContinueStatement() (*ast.ContinueStatement, *ParserError) {
	stmt := &ast.ContinueStatement{Token: p.curToken}

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseFallthroughStatement() (*ast.FallthroughStatement, *ParserError) {
	stmt := &ast.FallthroughStatement{Token: p.curToken}

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
	AND_OR
	EQUALS
	LESSGREATER
	SUM
	PRODUCT
	PREFIX
	CALL
	INDEX
	MEMBER
)

func (p *Parser) parsePrefixObjectValue() ast.Expression {
	ident := &ast.Identifier{Token: p.curToken, Value: ""}
	p.nextToken()
	ident.Value = p.curToken.Literal
	return ident
}

func (p *Parser) parseExpression(precedence int, flag ...bool) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	// if p.curTokenIs(token.IDENT) && len(flag) > 0 {
	// 	prefix = p.parseFromIdentifier
	// }
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()
	if !p.curTokenIs(token.IDENT) && !p.curTokenIs(token.DATE_LIT) &&
		p.peekToken.Type == token.DOT {
		return leftExp
	}
	isSaved := false
	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		if p.peekTokenIs(token.NOT) {
			p.Save()
			isSaved = true
			p.nextToken()
			if !p.peekTokenIs(token.IN, token.LIKE, token.BETWEEN) {
				p.Restore()
				isSaved = false
			}
			// p.Clear()
		}
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		if isSaved {
			p.Restore()
		}
		p.nextToken()
		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) parseIdentifier() ast.Expression {
	// return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	tok := p.curToken
	// pok := p.peekToken
	if p.peekToken.Type == token.LPAREN {
		return p.parseArrayFunctionCall()
	}
	if p.peekToken.Type == token.LBRACE {
		p.Save()
		p.nextToken() // {
		p.nextToken() // next
		if p.curTokenIs(token.IDENT) && p.peekTokenIs(token.COLON) {
			p.Restore()
			return p.parseStructLiteral()
		}
		p.Restore()
	}
	return &ast.Identifier{Token: tok, Value: tok.Literal}
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
func (p *Parser) parseNullLiteral() ast.Expression {
	return &ast.NullLiteral{Token: p.curToken}
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

func (p *Parser) parsePropertyAccess(left ast.Expression) ast.Expression {
	pa := &ast.TypeMember{Token: p.curToken, Left: left}
	prefix := p.prefixParseFns[p.peekToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.peekToken.Type)
		return nil
	}
	p.nextToken()
	pa.Right = prefix()
	for !p.peekTokenIs(token.SEMICOLON) && p.peekTokenIs(token.DOT) {
		p.nextToken()
		pa.Right = p.parsePropertyAccess(pa.Right) // p.parseExpression(LOWEST)
	}
	return pa
}

func (p *Parser) parseBetweenExpression(left ast.Expression) ast.Expression {
	pa := &ast.BetweenExpression{Token: p.curToken, Base: left}
	if p.curTokenIs(token.NOT) {
		pa.Not = true
		if !p.expectPeek(token.BETWEEN) {
			return nil
		}
	}

	prefix := p.prefixParseFns[p.peekToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.peekToken.Type)
		return nil
	}
	p.nextToken()
	pa.Left = prefix()
	if !p.expectPeek(token.AND) {
		return nil
	}

	prefix = p.prefixParseFns[p.peekToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.peekToken.Type)
		return nil
	}
	p.nextToken()
	pa.Right = prefix()
	return pa
}

func (p *Parser) parseSQLSelect() ast.Expression {
	selectStmt, pe := p.parseSQLSelectStatement()
	if pe != nil {
		p.addError(pe)
	}

	return selectStmt
}

func (p *Parser) parseSQLCreateObject() (*ast.SQLCreateObjectStatement, *ParserError) {
	stmt := &ast.SQLCreateObjectStatement{Token: p.curToken}

	// CREATE
	if !p.expectPeek(token.OBJECT) {
		return nil, nil
	}

	// IF NOT EXISTS optionnel
	if p.peekTokenIs(token.IF) {
		p.nextToken() // IF
		if !p.expectPeek(token.NOT) {
			return nil, nil
		}
		if !p.expectPeek(token.EXISTS) {
			return nil, nil
		}
		stmt.IfNotExists = true
	}

	// Nom de l'objet
	if !p.expectPeek(token.IDENT) {
		return nil, nil
	}
	stmt.ObjectName = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// (
	if !p.expectPeek(token.LPAREN) {
		return nil, nil //Create("'(' expected", p.peekToken.Line, p.peekToken.Column)
	}

	p.nextToken()

	// Colonnes et contraintes
	for !p.curTokenIs(token.RPAREN) && !p.curTokenIs(token.EOF) {
		if p.curTokenIs(token.CONSTRAINT) {
			constraint, pe := p.parseSQLConstraint()
			p.addError(pe)
			if constraint != nil {
				stmt.Constraints = append(stmt.Constraints, constraint)
			}
		} else {
			column, pe := p.parseSQLColumnDefinition()
			p.addError(pe)
			if column != nil {
				stmt.Columns = append(stmt.Columns, column)
			}
		}

		if p.peekTokenIs(token.COMMA) {
			p.nextToken()
		}
		p.nextToken()
	}
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt, nil
}

func (p *Parser) parseSQLColumnDefinition() (*ast.SQLColumnDefinition, *ParserError) {
	col := &ast.SQLColumnDefinition{Token: p.curToken}
	col.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// Type de données
	if !p.expectPeekEx(token.IDENT, token.VARCHAR, token.CHAR, token.NUMERIC, token.DATE, token.BOOLEAN,
		token.INTEGER, token.DECIMAL, token.TIMESTAMP, token.DATETIME, token.TEXT, token.JSON) {
		return nil, nil
	}
	var pi *ParserError
	col.DataType, pi = p.parseSQLDataType()
	p.addError(pi)

	// Contraintes de colonne
	for p.peekTokenIs(token.NOT, token.UNIQUE, token.PRIMARY, token.CHECK, token.DEFAULT) {
		p.nextToken()
		constraint, pe := p.parseSQLColumnConstraint()
		p.addError(pe)
		if constraint != nil {
			col.Constraints = append(col.Constraints, constraint)
		}
	}

	return col, nil
}

func (p *Parser) parseSQLDataType() (*ast.SQLDataType, *ParserError) {
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
			return nil, nil //Create("')' expected", p.peekToken.Line, p.peekToken.Column)
		}
	}

	return dt, nil
}

func (p *Parser) parseSQLColumnConstraint() (*ast.SQLColumnConstraint, *ParserError) {
	constraint := &ast.SQLColumnConstraint{Token: p.curToken}

	switch p.curToken.Type {
	case token.NOT:
		if !p.expectPeek(token.NULL) {
			return nil, nil //Create("token 'null' expected", p.peekToken.Line, p.peekToken.Column)
		}
		constraint.Type = "NOT NULL"
	case token.UNIQUE:
		constraint.Type = "UNIQUE"
	case token.PRIMARY:
		if !p.expectPeek(token.KEY) {
			return nil, nil //Create("token 'key' expected", p.peekToken.Line, p.peekToken.Column)
		}
		constraint.Type = "PRIMARY KEY"
	case token.DEFAULT:
		constraint.Type = "DEFAULT"
		p.nextToken()
		constraint.Expression = p.parseExpression(LOWEST)
		return constraint, nil
	case token.CHECK:
		constraint.Type = "CHECK"
		p.nextToken()
		constraint.Expression = p.parseExpression(LOWEST)
		return constraint, nil
	}

	return constraint, nil
}

func (p *Parser) parseSQLConstraint() (*ast.SQLConstraint, *ParserError) {
	constraint := &ast.SQLConstraint{Token: p.curToken}
	var pe *ParserError = nil

	if !p.expectPeek(token.IDENT) {
		return nil, nil //Create("'identifier' expected", p.peekToken.Line, p.peekToken.Column)
	}
	constraint.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	p.nextToken()

	switch p.curToken.Type {
	case token.PRIMARY:
		constraint.Type = "PRIMARY KEY"
		if !p.expectPeek(token.KEY) {
			return nil, nil //Create("token 'key' expected", p.peekToken.Line, p.peekToken.Column)
		}
		if !p.expectPeek(token.LPAREN) {
			return nil, nil //Create("'(' expected", p.peekToken.Line, p.peekToken.Column)
		}
		constraint.Columns, pe = p.parseColumnList()
		p.addError(pe)
	case token.FOREIGN:
		constraint.Type = "FOREIGN KEY"
		if !p.expectPeek(token.KEY) {
			return nil, nil //Create("token 'key' expected", p.peekToken.Line, p.peekToken.Column)
		}
		if !p.expectPeek(token.LPAREN) {
			return nil, nil //Create("'(' expected", p.peekToken.Line, p.peekToken.Column)
		}
		constraint.Columns, pe = p.parseColumnList()
		p.addError(pe)
		if !p.expectPeek(token.REFERENCES) {
			return nil, nil //Create("token 'references' expected", p.peekToken.Line, p.peekToken.Column)
		}
		constraint.References, pe = p.parseSQLReference()
		p.addError(pe)
	case token.UNIQUE:
		constraint.Type = "UNIQUE"
		if !p.expectPeek(token.LPAREN) {
			return nil, nil //Create("token 'unique' expected", p.peekToken.Line, p.peekToken.Column)
		}
		constraint.Columns, pe = p.parseColumnList()
		p.addError(pe)
	case token.CHECK:
		constraint.Type = "CHECK"
		p.nextToken()
		constraint.Check = p.parseExpression(LOWEST)
	}

	return constraint, pe
}

func (p *Parser) parseColumnList() ([]*ast.Identifier, *ParserError) {
	var columns []*ast.Identifier

	p.nextToken()
	columns = append(columns, &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal})

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		columns = append(columns, &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal})
	}

	if !p.expectPeek(token.RPAREN) {
		return nil, nil //Create("')' expected", p.peekToken.Line, p.peekToken.Column)
	}

	return columns, nil
}

func (p *Parser) parseSQLReference() (*ast.SQLReference, *ParserError) {
	ref := &ast.SQLReference{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil, nil //Create("'identifier' expected", p.peekToken.Line, p.peekToken.Column)
	}
	ref.TableName = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	var pe *ParserError = nil

	if p.peekTokenIs(token.LPAREN) {
		p.nextToken()
		ref.Columns, pe = p.parseColumnList()
	}

	return ref, pe
}

func (p *Parser) parseSQLDropObject() (*ast.SQLDropObjectStatement, *ParserError) {
	stmt := &ast.SQLDropObjectStatement{Token: p.curToken}

	if !p.expectPeek(token.OBJECT) {
		return nil, nil //Create("'object' expected", p.peekToken.Line, p.peekToken.Column)
	}

	// IF EXISTS optionnel
	if p.peekTokenIs(token.IF) {
		p.nextToken() // IF
		if !p.expectPeek(token.EXISTS) {
			return nil, nil
		}
		stmt.IfExists = true
	}

	// Nom de l'objet
	if !p.expectPeek(token.IDENT) {
		return nil, nil
	}
	stmt.ObjectName = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// CASCADE optionnel
	if p.peekTokenIs(token.CASCADE) {
		p.nextToken()
		stmt.Cascade = true
	}
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt, nil
}

func (p *Parser) parseSQLAlterObject() (*ast.SQLAlterObjectStatement, *ParserError) {
	stmt := &ast.SQLAlterObjectStatement{Token: p.curToken}

	if !p.expectPeek(token.OBJECT) {
		return nil, nil //Create("'object' expected", p.peekToken.Line, p.peekToken.Column)
	}

	// Nom de l'objet
	if !p.expectPeek(token.IDENT) {
		return nil, nil //Create("'identifier' expected", p.peekToken.Line, p.peekToken.Column)
	}
	stmt.ObjectName = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// Actions
	for !p.curTokenIs(token.SEMICOLON) && !p.curTokenIs(token.EOF) &&
		p.expectPeekEx(token.ADD, token.MODIFY, token.DROP) {
		action, pe := p.parseSQLAlterAction()
		if action != nil {
			stmt.Actions = append(stmt.Actions, action)
		}
		p.addError(pe)
		if p.peekTokenIs(token.COMMA) && !p.curTokenIs(token.SEMICOLON) {
			p.nextToken()
		}
	}

	return stmt, nil
}

func (p *Parser) parseSQLAlterAction() (*ast.SQLAlterAction, *ParserError) {
	action := &ast.SQLAlterAction{Token: p.curToken}
	var pe *ParserError = nil

	switch p.curToken.Type {
	case token.ADD:
		action.Type = "ADD"
		if !p.expectPeekEx(token.CONSTRAINT, token.COLUMN) {
			return nil, nil
		}
		if p.curTokenIs(token.CONSTRAINT) {
			action.Constraint, pe = p.parseSQLConstraint()
		} else {
			p.nextToken()
			action.Column, pe = p.parseSQLColumnDefinition()
		}
	case token.MODIFY:
		action.Type = "MODIFY"
		p.nextToken()
		action.Column, pe = p.parseSQLColumnDefinition()
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
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return action, pe
}

func (p *Parser) parseSQLInsert() (*ast.SQLInsertStatement, *ParserError) {
	stmt := &ast.SQLInsertStatement{Token: p.curToken}

	if !p.expectPeek(token.INTO) {
		return nil, nil //Create("token 'INTO' expected", p.peekToken.Line, p.peekToken.Column)
	}

	// Nom de l'objet
	if !p.expectPeek(token.IDENT) {
		return nil, nil //Create("'identifier' expected", p.peekToken.Line, p.peekToken.Column)
	}
	stmt.ObjectName = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// Colonnes optionnelles
	var pe *ParserError = nil
	if p.peekTokenIs(token.LPAREN) {
		p.nextToken()
		stmt.Columns, pe = p.parseColumnList()
	}

	// VALUES ou SELECT
	if p.peekTokenIs(token.VALUES) {
		p.nextToken()
		stmt.Values = p.parseSQLValues()
	} else if p.peekTokenIs(token.SELECT) {
		p.nextToken()
		stmt.Select, pe = p.parseSQLSelectStatement() //.(*ast.SQLSelectStatement)
	}
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt, pe
}

func (p *Parser) parseSQLValues() []*ast.SQLValues {
	var valuesList []*ast.SQLValues

	if p.curTokenIs(token.VALUES) {
		p.nextToken()
	}
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
		if !p.peekTokenIs(token.COMMA) {
			break
		}
		p.nextToken()
		if p.peekTokenIs(token.LPAREN) {
			p.nextToken()
		}
	}

	return valuesList
}

func (p *Parser) parseSQLUpdate() (*ast.SQLUpdateStatement, *ParserError) {
	stmt := &ast.SQLUpdateStatement{Token: p.curToken}

	// Nom de l'objet
	if !p.expectPeek(token.IDENT) {
		return nil, nil
	}
	stmt.ObjectName = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.SET) {
		return nil, nil
	}

	p.nextToken()

	// Clauses SET
	for !p.curTokenIs(token.WHERE) && !p.curTokenIs(token.SEMICOLON) &&
		!p.curTokenIs(token.STOP) && !p.curTokenIs(token.EOF) {
		setClause := &ast.SQLSetClause{Token: p.curToken}
		setClause.Column = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

		if !p.expectPeek(token.ASSIGN) {
			return nil, nil //Create("'=' expected", p.peekToken.Line, p.peekToken.Column)
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
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt, nil
}

func (p *Parser) parseSQLDelete() (*ast.SQLDeleteStatement, *ParserError) {
	stmt := &ast.SQLDeleteStatement{Token: p.curToken}

	if !p.expectPeek(token.FROM) {
		return nil, nil
	}

	// Nom de l'objet
	if !p.expectPeek(token.IDENT) {
		return nil, nil
	}
	stmt.From = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// WHERE optionnel
	if p.peekTokenIs(token.WHERE) {
		p.nextToken()
		p.nextToken()
		stmt.Where = p.parseExpression(LOWEST)
	}
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt, nil
}

func (p *Parser) parseSQLTruncate() (*ast.SQLTruncateStatement, *ParserError) {
	stmt := &ast.SQLTruncateStatement{Token: p.curToken}

	if !p.expectPeek(token.OBJECT) {
		return nil, nil //Create("'object' expected", p.peekToken.Line, p.peekToken.Column)
	}

	// Nom de l'objet
	if !p.expectPeek(token.IDENT) {
		return nil, nil //Create("'identifier' expected", p.peekToken.Line, p.peekToken.Column)
	}
	stmt.ObjectName = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseSQLCreateIndex() (*ast.SQLCreateIndexStatement, *ParserError) {
	stmt := &ast.SQLCreateIndexStatement{Token: p.curToken}

	// UNIQUE optionnel
	if p.peekTokenIs(token.UNIQUE) {
		p.nextToken()
		stmt.Unique = true
	}

	if !p.expectPeek(token.INDEX) {
		return nil, nil //Create("token 'index' expected", p.peekToken.Line, p.peekToken.Column)
	}

	// Nom de l'index
	if !p.expectPeek(token.IDENT) {
		return nil, nil //Create("'identifier' expected", p.peekToken.Line, p.peekToken.Column)
	}
	stmt.IndexName = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.ON) {
		return nil, nil //Create("token 'on' expected", p.peekToken.Line, p.peekToken.Column)
	}

	// Nom de l'objet
	if !p.expectPeek(token.IDENT) {
		return nil, nil //Create("'identifier' expected", p.peekToken.Line, p.peekToken.Column)
	}
	stmt.ObjectName = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.LPAREN) {
		return nil, nil //Create("'(' expected", p.peekToken.Line, p.peekToken.Column)
	}

	var pe *ParserError
	stmt.Columns, pe = p.parseColumnList()
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt, pe
}

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
		p.addError(Create(fmt.Sprintf("'%s' is not expected here.", p.curToken.Literal), p.curToken.Line, p.curToken.Column))
		return nil
		// expressions = append(expressions, &ast.SelectArgs{
		// 	Expr:    &ast.Identifier{Token: p.curToken, Value: "*"},
		// 	NewName: nil,
		// })
		// p.nextToken()
		// return expressions
	}
	arg := &ast.SelectArgs{}
	arg.Expr = p.parseExpression(LOWEST)
	if p.peekTokenIs(token.AS) {
		p.nextToken()
		if !p.expectPeekEx(token.IDENT, token.STRING_LIT, token.INT_LIT) {
			return nil
		}
		p.nextToken() //move to the new name
		arg.NewName = p.parseIdentifier().(*ast.Identifier)
	}
	expressions = append(expressions, arg)
	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		arg = &ast.SelectArgs{Expr: p.parseExpression(LOWEST), NewName: nil}
		if p.peekTokenIs(token.AS) {
			p.nextToken() //as
			p.nextToken() //move to the new name
			arg.NewName = p.parseIdentifier().(*ast.Identifier)
		}
		expressions = append(expressions, arg)

		// expressions = append(expressions, p.parseExpression(LOWEST))
	}

	return expressions
}
func (p *Parser) contains(tab []token.TokenType, element token.TokenType) bool {
	for _, v := range tab {
		if v == element {
			return true
		}
	}
	return false
}
func (p *Parser) curTokenIs(t ...token.TokenType) bool {
	return p.contains(t, p.curToken.Type)
}

func (p *Parser) peekTokenIs(t ...token.TokenType) bool {
	return p.contains(t, p.peekToken.Type)
}

func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}

func (p *Parser) expectPeekEx(t ...token.TokenType) bool {
	var v token.TokenType
	for _, v = range t {
		if p.peekTokenIs(v) {
			p.nextToken()
			return true
		}
	}
	msg := fmt.Sprintf("Expected %s, got %s", t, p.peekToken.Type)
	p.errors = append(p.errors, *Create(msg, p.peekToken.Line, p.peekToken.Column))
	return false
}

func (p *Parser) Errors() []ParserError {
	return p.errors
}

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("Expected %s, got %s", t, p.peekToken.Type)
	p.errors = append(p.errors, *Create(msg, p.peekToken.Line, p.peekToken.Column))
}

func (p *Parser) curError(t token.TokenType) {
	msg := fmt.Sprintf("Expected %s, got %s", t, p.curToken.Type)
	p.errors = append(p.errors, *Create(msg, p.curToken.Line, p.curToken.Column))
}
func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("Unexpected here %s", t)
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
	token.AND:      AND_OR,
	token.OR:       AND_OR,
	token.SLASH:    PRODUCT,
	token.ASTERISK: PRODUCT,
	token.MOD:      PRODUCT,
	token.LBRACKET: INDEX,
	// token.CONCAT:   SUM,
	token.IN:      LESSGREATER,
	token.NOT:     LESSGREATER,
	token.IS:      EQUALS,
	token.LIKE:    LESSGREATER,
	token.BETWEEN: LESSGREATER,
	// token.AS:       EQUALS,
	token.DOT: MEMBER,
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func (p *Parser) parseSQLWithStatement() (*ast.SQLWithStatement, *ParserError) {
	stmt := &ast.SQLWithStatement{Token: p.curToken}
	var pe *ParserError = nil

	// RECURSIVE optionnel
	if p.peekTokenIs(token.RECURSIVE) {
		p.nextToken()
		stmt.Recursive = true
	}

	// CTEs
	stmt.CTEs, pe = p.parseCTEList()
	p.addError(pe)

	// Requête principale
	if !p.expectPeek(token.SELECT) {
		return nil, nil //Create("token 'select' expected", p.peekToken.Line, p.peekToken.Column)
	}
	// p.nextToken()
	stmt.Select, pe = p.parseSQLSelectStatement() //.(*ast.SQLSelectStatement)

	return stmt, pe
}

func (p *Parser) parseCTEList() ([]*ast.SQLCommonTableExpression, *ParserError) {
	var ctes []*ast.SQLCommonTableExpression
	var pe *ParserError = nil
	cte, pe := p.parseCommonTableExpression()
	if cte == nil {
		return nil, pe
	}
	ctes = append(ctes, cte)

	if pe != nil {
		p.errors = append(p.errors, *pe)
	}

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		cte, pe = p.parseCommonTableExpression()
		if pe != nil {
			p.errors = append(p.errors, *pe)
		}
		if cte != nil {
			ctes = append(ctes, cte)
		}
	}

	return ctes, pe
}

func (p *Parser) parseCommonTableExpression() (*ast.SQLCommonTableExpression, *ParserError) {
	p.nextToken() //nom de l'objet temporaire
	cte := &ast.SQLCommonTableExpression{Token: p.curToken}
	cte.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	var pe *ParserError = nil

	// Colonnes optionnelles
	if p.peekTokenIs(token.LPAREN) {
		p.nextToken()
		cte.Columns, pe = p.parseColumnList()
	}
	p.addError(pe)
	if !p.expectPeek(token.AS) {
		return nil, nil
	}

	if !p.expectPeek(token.LPAREN) {
		return nil, nil
	}

	p.nextToken()

	if p.curTokenIs(token.SELECT) {
		cte.Query, pe = p.parseSQLSelectStatement() //.(*ast.SQLSelectStatement)
	}

	if !p.expectPeek(token.RPAREN) {
		return nil, nil //Create("')' expected", p.peekToken.Line, p.peekToken.Column)
	}

	return cte, pe
}
func (p *Parser) addError(pe *ParserError) {
	if pe != nil {
		p.errors = append(p.errors, *pe)
	}
}
func (p *Parser) parseRecursiveCTE() (*ast.SQLRecursiveCTE, *ParserError) {
	cte := &ast.SQLRecursiveCTE{Token: p.curToken}
	cte.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	var pe *ParserError = nil
	// Colonnes optionnelles
	if p.peekTokenIs(token.LPAREN) {
		p.nextToken()
		cte.Columns, pe = p.parseColumnList()
	}

	if !p.expectPeek(token.AS) {
		return nil, nil //Create("token 'as' expected", p.peekToken.Line, p.peekToken.Column)
	}

	if !p.expectPeek(token.LPAREN) {
		return nil, nil //Create("'(' expected", p.peekToken.Line, p.peekToken.Column)
	}

	p.nextToken()

	// Partie anchor
	if p.curTokenIs(token.SELECT) {
		cte.Anchor, pe = p.parseSQLSelectStatement() //.(*ast.SQLSelectStatement)
	}
	p.addError(pe)

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
			cte.Recursive, pe = p.parseSQLSelectStatement() //.(*ast.SQLSelectStatement)
		}
	}

	if !p.expectPeek(token.RPAREN) {
		return nil, nil //Create("')' expected", p.peekToken.Line, p.peekToken.Column)
	}

	return cte, pe
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
func (p *Parser) parseSQLSelectStatement() (*ast.SQLSelectStatement, *ParserError) {
	// Vérifier d'abord s'il y a une clause WITH
	// var pe *ParserError=nil

	if p.curTokenIs(token.WITH) {
		// var selectStmt *ast.SQLSelectStatement
		withStmt, pe := p.parseSQLWithStatement()
		if withStmt == nil {
			return nil, pe
		}
		p.addError(pe)
		selectStmt := withStmt.Select
		if selectStmt == nil {
			return nil, pe
		}
		selectStmt.With = withStmt
		return selectStmt, nil //return withStmt //
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
		return nil, nil
	}
	p.nextToken()
	from := ast.FromIdentifier{Token: p.curToken}

	from.Value = p.parseExpression(LOWEST, true)
	if p.peekTokenIs(token.IDENT) {
		p.nextToken()
		from.NewName = p.parseIdentifier()
	}
	selectStmt.From = &from
	// JOINs optionnels
	for p.peekTokenIs(token.JOIN) ||
		(p.peekTokenIs(token.IDENT) &&
			(strings.ToUpper(p.peekToken.Literal) == "INNER" ||
				strings.ToUpper(p.peekToken.Literal) == "LEFT" ||
				// strings.ToUpper(p.peekToken.Literal) == "RIGHT" ||
				strings.ToUpper(p.peekToken.Literal) == "FULL")) {
		p.nextToken()
		join := &ast.SQLJoin{Token: p.curToken}

		if p.curTokenIs(token.IDENT) {
			join.Type = strings.ToUpper(p.curToken.Literal)
			if !p.expectPeek(token.JOIN) {
				return nil, nil //Create("token 'join' expected", p.peekToken.Line, p.peekToken.Column)
			}
			p.nextToken()
		} else {
			join.Type = "INNER"
		}
		frm := &ast.FromIdentifier{Token: p.curToken}
		frm.Value = p.parseExpression(LOWEST)
		if p.peekTokenIs(token.IDENT) {
			p.nextToken()
			frm.NewName = p.parseIdentifier()
		}
		join.Table = frm
		// join.Table = p.parseExpression(LOWEST, true)
		if !p.expectPeek(token.ON) {
			return nil, nil
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
			return nil, nil //Create("'group' expected", p.peekToken.Line, p.peekToken.Column)
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
			return nil, nil //Create("'order' expected", p.peekToken.Line, p.peekToken.Column)
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
	var pe *ParserError = nil
	if p.peekTokenIs(token.UNION) {
		p.nextToken()
		if p.peekTokenIs(token.ALL) {
			p.nextToken()
			selectStmt.UnionAll = true
		}
		p.nextToken()
		selectStmt.Union, pe = p.parseSQLSelectStatement() //.(*ast.SQLSelectStatement)
	}
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return selectStmt, pe
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

	if p.curTokenIs(token.LBRACE) {
		array.Elements = append(array.Elements, p.parseStructLiteral())
	} else {
		array.Elements = append(array.Elements, p.parseExpression(LOWEST))
	}

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		if p.peekTokenIs(token.RBRACKET) {
			break
		}
		p.nextToken()
		if p.curTokenIs(token.LBRACE) {
			array.Elements = append(array.Elements, p.parseStructLiteral())
		} else {
			array.Elements = append(array.Elements, p.parseExpression(LOWEST))
		}
		// array.Elements = append(array.Elements, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(token.RBRACKET) {
		return nil
	}

	return array
}
func (p *Parser) parseStructLiteral() ast.Expression {
	strt := &ast.StructLiteral{Token: p.curToken, Name: nil}
	if p.curTokenIs(token.IDENT) {
		strt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		if !p.expectPeek(token.LBRACE) {
			return nil
		}
	}
	if p.peekTokenIs(token.RBRACE) {
		p.nextToken()
		// p.nextToken()
		return strt
	}
	p.nextToken() //

	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		field := &ast.StructFieldLit{Token: p.curToken}
		field.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

		if !p.expectPeek(token.COLON) {
			return nil
		}
		p.nextToken()
		field.Value = p.parseExpression(LOWEST)

		strt.Fields = append(strt.Fields, *field)

		if p.peekTokenIs(token.COMMA) {
			p.nextToken()
		}
		p.nextToken()
	}

	if !p.curTokenIs(token.RBRACE) {
		p.curError(token.RBRACE)
		return nil
	}

	return strt
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
	exp := &ast.SliceExpression{Token: p.curToken, Left: left, Start: nil, End: nil}

	p.nextToken()

	if !p.curTokenIs(token.COLON) {
		exp.Start = p.parseExpression(LOWEST)
		if !p.expectPeek(token.COLON) {
			return nil
		}
	}

	if p.curTokenIs(token.COLON) && !p.peekTokenIs(token.RBRACKET) {
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

func (p *Parser) parseLikeExpression(left ast.Expression) ast.Expression {
	exp := &ast.LikeExpression{Token: p.curToken, Left: left}

	// Vérifier NOT IN
	if p.curTokenIs(token.NOT) {
		exp.Not = true
		if !p.expectPeek(token.LIKE) {
			return nil
		}
	}

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
	if p.curTokenIs(token.RPAREN) {
		call.Arguments = nil
		call.Array = nil
		return call
	}
	//Check if the function has asterisk as unique argument
	if p.curTokenIs(token.ASTERISK) && p.peekTokenIs(token.RPAREN) {
		call.Array = &ast.Identifier{Token: p.curToken, Value: "*"}
		call.Arguments = nil
		p.nextToken()
		return call
	}
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

	if p.curTokenIs(token.OBJECT) {
		if !p.expectPeek(token.IDENT) {
			return nil
		}
		ta.Type = p.curToken.Literal
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
	p.Save()
	// tok := p.curToken

	p.nextToken()

	// Vérifier si c'est un slice (contient :)
	isSlice := false

	// Avancer pour vérifier
	for !p.curTokenIs(token.RBRACKET) && !p.curTokenIs(token.EOF) {
		// if p.curTokenIs(token.COLON) {
		// 	isSlice = true
		// 	break
		// }
		// p.nextToken()
		if p.curTokenIs(token.COLON) {
			isSlice = true
			break
		}
		p.nextToken()
	}

	p.Restore()

	if isSlice {
		return p.parseSliceExpression(left)
	}
	return p.parseIndexExpression(left)
}
