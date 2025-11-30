package parser

import (
    "fmt"
    "lang/ast"
    "lang/lexer"
    "lang/token"
    "strconv"
)

type Parser struct {
    l *lexer.Lexer

    curToken  token.Token
    peekToken token.Token

    errors []string

    prefixParseFns map[token.TokenType]prefixParseFn
    infixParseFns  map[token.TokenType]infixParseFn
}

type (
    prefixParseFn func() ast.Expression
    infixParseFn  func(ast.Expression) ast.Expression
)

func New(l *lexer.Lexer) *Parser {
    p := &Parser{
        l:      l,
        errors: []string{},
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
    if !p.expectPeek(token.ACTION) {
        p.errors = append(p.errors, "Le programme doit commencer par 'action'")
        return program
    }
    
    p.nextToken()
    
    // Lire le nom de l'action
    if !p.curTokenIs(token.STRING_LIT) {
        p.errors = append(p.errors, "Attendu un nom d'action après 'action'")
        return program
    }
    program.ActionName = p.curToken.Literal
    
    p.nextToken()
    
    // Parser les déclarations jusqu'à 'start'
    for !p.curTokenIs(token.START) && !p.curTokenIs(token.EOF) {
        stmt := p.parseStatement()
        if stmt != nil {
            program.Statements = append(program.Statements, stmt)
        }
        p.nextToken()
    }
    
    // Parser les instructions après 'start'
    if p.curTokenIs(token.START) {
        p.nextToken()
        for !p.curTokenIs(token.STOP) && !p.curTokenIs(token.EOF) {
            stmt := p.parseStatement()
            if stmt != nil {
                program.Statements = append(program.Statements, stmt)
            }
            p.nextToken()
        }
    }
    
    return program
}

func (p *Parser) parseStatement() ast.Statement {
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

func (p *Parser) parseLetStatement() *ast.LetStatement {
    stmt := &ast.LetStatement{Token: p.curToken}
    
    if !p.expectPeek(token.IDENT) {
        return nil
    }
    
    stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
    
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
    
    return stmt
}

func (p *Parser) parseTypeAnnotation() *ast.TypeAnnotation {
    ta := &ast.TypeAnnotation{Token: p.curToken, Type: p.curToken.Literal}
    
    // Vérifier les contraintes
    if p.peekTokenIs(token.LPAREN) || p.peekTokenIs(token.LBRACKET) {
        ta.Constraints = p.parseTypeConstraints()
    }
    
    return ta
}

func (p *Parser) parseTypeConstraints() *ast.TypeConstraints {
    constraints := &ast.TypeConstraints{}
    
    for p.peekTokenIs(token.LPAREN) || p.peekTokenIs(token.LBRACKET) {
        p.nextToken()
        switch p.curToken.Type {
        case token.LPAREN:
            if p.peekTokenIs(token.INT_LIT) {
                p.nextToken()
                maxDigits := &ast.IntegerLiteral{Token: p.curToken}
                val, _ := strconv.ParseInt(p.curToken.Literal, 10, 64)
                maxDigits.Value = val
                constraints.MaxDigits = maxDigits
                
                if p.peekTokenIs(token.DOT) {
                    p.nextToken() // .
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
                return nil
            }
        case token.LBRACKET:
            constraints.IntegerRange = p.parseRangeConstraint()
            if !p.expectPeek(token.RBRACKET) {
                return nil
            }
        }
    }
    
    return constraints
}

func (p *Parser) parseRangeConstraint() *ast.RangeConstraint {
    rc := &ast.RangeConstraint{}
    
    p.nextToken()
    rc.Min = p.parseExpression(LOWEST)
    
    if !p.expectPeek(token.DOT) || !p.expectPeek(token.DOT) {
        return nil
    }
    
    p.nextToken()
    rc.Max = p.parseExpression(LOWEST)
    
    return rc
}

func (p *Parser) parseFunctionStatement() *ast.FunctionStatement {
    stmt := &ast.FunctionStatement{Token: p.curToken}
    
    if !p.expectPeek(token.IDENT) {
        return nil
    }
    
    stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
    
    if !p.expectPeek(token.LPAREN) {
        return nil
    }
    
    stmt.Parameters = p.parseFunctionParameters()
    
    // Type de retour optionnel
    if p.peekTokenIs(token.COLON) {
        p.nextToken() // :
        p.nextToken() // type
        stmt.ReturnType = p.parseTypeAnnotation()
    }
    
    if !p.expectPeek(token.LBRACE) {
        return nil
    }
    
    stmt.Body = p.parseBlockStatement()
    
    return stmt
}

func (p *Parser) parseFunctionParameters() []*ast.FunctionParameter {
    var params []*ast.FunctionParameter
    
    if p.peekTokenIs(token.RPAREN) {
        p.nextToken()
        return params
    }
    
    p.nextToken()
    
    param := &ast.FunctionParameter{Token: p.curToken}
    param.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
    
    if !p.expectPeek(token.COLON) {
        return nil
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
            return nil
        }
        
        p.nextToken()
        param.Type = p.parseTypeAnnotation()
        params = append(params, param)
    }
    
    if !p.expectPeek(token.RPAREN) {
        return nil
    }
    
    return params
}

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
    
    // Parser les champs
    for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
        field := &ast.StructField{Token: p.curToken}
        field.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
        
        if !p.expectPeek(token.COLON) {
            return nil
        }
        
        p.nextToken()
        field.Type = p.parseTypeAnnotation()
        stmt.Fields = append(stmt.Fields, field)
        
        if p.peekTokenIs(token.COMMA) {
            p.nextToken()
        }
        p.nextToken()
    }
    
    return stmt
}

func (p *Parser) parseForStatement() *ast.ForStatement {
    stmt := &ast.ForStatement{Token: p.curToken}
    
    if !p.expectPeek(token.LPAREN) {
        return nil
    }
    
    p.nextToken()
    
    // Initialisation
    if !p.curTokenIs(token.SEMICOLON) {
        stmt.Init = p.parseStatement()
    }
    
    if !p.expectPeek(token.SEMICOLON) {
        return nil
    }
    
    p.nextToken()
    
    // Condition
    if !p.curTokenIs(token.SEMICOLON) {
        stmt.Condition = p.parseExpression(LOWEST)
    }
    
    if !p.expectPeek(token.SEMICOLON) {
        return nil
    }
    
    p.nextToken()
    
    // Update
    if !p.curTokenIs(token.RPAREN) {
        stmt.Update = p.parseStatement()
    }
    
    if !p.expectPeek(token.RPAREN) {
        return nil
    }
    
    if !p.expectPeek(token.LBRACE) {
        return nil
    }
    
    stmt.Body = p.parseBlockStatement()
    
    return stmt
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
    stmt := &ast.ReturnStatement{Token: p.curToken}
    
    p.nextToken()
    
    if !p.curTokenIs(token.SEMICOLON) {
        stmt.ReturnValue = p.parseExpression(LOWEST)
    }
    
    if p.curTokenIs(token.SEMICOLON) {
        p.nextToken()
    }
    
    return stmt
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
    block := &ast.BlockStatement{Token: p.curToken}
    
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

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
    stmt := &ast.ExpressionStatement{Token: p.curToken}
    
    stmt.Expression = p.parseExpression(LOWEST)
    
    if p.peekTokenIs(token.SEMICOLON) {
        p.nextToken()
    }
    
    return stmt
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
)

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
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
    lit := &ast.IntegerLiteral{Token: p.curToken}
    
    value, err := strconv.ParseInt(p.curToken.Literal, 10, 64)
    if err != nil {
        p.errors = append(p.errors, fmt.Sprintf("Impossible de parser %q comme entier", p.curToken.Literal))
        return nil
    }
    
    lit.Value = value
    return lit
}

func (p *Parser) parseFloatLiteral() ast.Expression {
    lit := &ast.FloatLiteral{Token: p.curToken}
    
    value, err := strconv.ParseFloat(p.curToken.Literal, 64)
    if err != nil {
        p.errors = append(p.errors, fmt.Sprintf("Impossible de parser %q comme flottant", p.curToken.Literal))
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

func (p *Parser) Errors() []string {
    return p.errors
}

func (p *Parser) peekError(t token.TokenType) {
    msg := fmt.Sprintf("Attendu %s, got %s", t, p.peekToken.Type)
    p.errors = append(p.errors, msg)
}

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
    msg := fmt.Sprintf("Aucune fonction prefix parse pour %s", t)
    p.errors = append(p.errors, msg)
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
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
    p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
    p.infixParseFns[tokenType] = fn
}