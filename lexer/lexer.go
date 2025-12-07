package lexer

import (
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/akristianlopez/action/token"
)

type Lexer struct {
	input        string
	position     int
	readPosition int
	ch           rune
	line         int
	column       int
}

var cnt *Lexer = nil

func (l *Lexer) SaveCnt() {
	cnt = &Lexer{
		position:     l.position,
		readPosition: l.readPosition,
		ch:           l.ch,
		line:         l.line,
		column:       l.column,
	}
}
func (l *Lexer) RestoreCnt() {
	if cnt != nil {
		l.position = cnt.position
		l.readPosition = cnt.readPosition
		l.ch = cnt.ch
		l.line = cnt.line
		l.column = cnt.column
		cnt = nil
	}
}

func (l *Lexer) GetCursorPosition() (int, int) {
	return l.position, l.readPosition
}

func (l *Lexer) SetCursorPosition(pos, cur int) {
	l.position = pos
	l.readPosition = cur
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = rune(l.input[l.readPosition-1])
	}

}

func New(input string) *Lexer {
	l := &Lexer{input: input, line: 1, column: 0}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch, _ = utf8.DecodeRuneInString(l.input[l.readPosition:])
	}

	l.position = l.readPosition
	if l.ch == '\n' {
		l.line++
		l.column = 0
	} else {
		l.column++
	}
	l.readPosition += utf8.RuneLen(l.ch)
}

func (l *Lexer) skipComment() {
	// Vérifier les commentaires
	if l.ch == '(' && l.peekChar() == '*' {
		l.readChar()
		l.readChar()
		l.readComment()
		// return l.readComment()
	}
	l.skipWhitespace()
}

func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	l.skipWhitespace()
	l.skipComment()

	// Vérifier les littéraux date/time
	if l.ch == '#' {
		return l.readDateTimeLiteral()
	}

	switch l.ch {
	case '=':
		if l.peekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.EQ, Literal: "==", Line: l.line, Column: l.column}
		} else {
			tok = newToken(token.ASSIGN, l.ch, l.line, l.column)
		}
	case '+':
		tok = newToken(token.PLUS, l.ch, l.line, l.column)
	case '-':
		tok = newToken(token.MINUS, l.ch, l.line, l.column)
	case '*':
		tok = newToken(token.ASTERISK, l.ch, l.line, l.column)
	case '/':
		tok = newToken(token.SLASH, l.ch, l.line, l.column)
	case '%':
		tok = newToken(token.MOD, l.ch, l.line, l.column)
	case '<':
		if l.peekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.LTE, Literal: "<=", Line: l.line, Column: l.column}
		} else if l.peekChar() == '-' {
			tok = newToken(token.LAR, l.ch, l.line, l.column)
		} else {
			tok = newToken(token.LT, l.ch, l.line, l.column)
		}
	case '>':
		if l.peekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.GTE, Literal: ">=", Line: l.line, Column: l.column}
		} else {
			tok = newToken(token.GT, l.ch, l.line, l.column)
		}
	case '!':
		if l.peekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.NOT_EQ, Literal: "!=", Line: l.line, Column: l.column}
		}
	case ',':
		tok = newToken(token.COMMA, l.ch, l.line, l.column)
	case ';':
		tok = newToken(token.SEMICOLON, l.ch, l.line, l.column)
	case ':':
		tok = newToken(token.COLON, l.ch, l.line, l.column)
	case '.':
		tok = newToken(token.DOT, l.ch, l.line, l.column)
	case '(':
		tok = newToken(token.LPAREN, l.ch, l.line, l.column)
	case ')':
		tok = newToken(token.RPAREN, l.ch, l.line, l.column)
	case '{':
		tok = newToken(token.LBRACE, l.ch, l.line, l.column)
	case '}':
		tok = newToken(token.RBRACE, l.ch, l.line, l.column)
	case '[':
		tok = newToken(token.LBRACKET, l.ch, l.line, l.column)
	case ']':
		tok = newToken(token.RBRACKET, l.ch, l.line, l.column)
	case '|':
		if l.peekChar() == '|' {
			l.readChar()
			tok = token.Token{Type: token.CONCAT, Literal: "||", Line: l.line, Column: l.column}
		}
	case '"':
		tok.Type = token.STRING_LIT
		tok.Literal = l.readString()
		tok.Line = l.line
		tok.Column = l.column
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
		tok.Line = l.line
		tok.Column = l.column
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = token.LookupIdent(strings.ToLower(tok.Literal))
			tok.Line = l.line
			tok.Column = l.column
			return tok
		} else if isDigit(l.ch) {
			return l.readNumber()
		} else {
			tok = newToken(token.ILLEGAL, l.ch, l.line, l.column)
		}
	}

	l.readChar()
	return tok
}

func (l *Lexer) readComment() token.Token {
	position := l.position
	line := l.line
	column := l.column

	for {
		l.readChar()
		if l.ch == '*' && l.peekChar() == ')' {
			l.readChar()
			l.readChar()
			break
		}
		if l.ch == 0 {
			break
		}
	}

	return token.Token{
		Type:    token.COMMENT,
		Literal: l.input[position:l.position],
		Line:    line,
		Column:  column,
	}
}

func (l *Lexer) readDateTimeLiteral() token.Token {
	position := l.position
	line := l.line
	column := l.column

	l.readChar() // consommer le '#'

	for l.ch != '#' && l.ch != 0 {
		l.readChar()
	}

	if l.ch == '#' {
		l.readChar() // consommer le '#' de fermeture
	}

	literal := l.input[position:l.position]

	// Déterminer si c'est une date ou un time
	if isTimeLiteral(literal) {
		return token.Token{Type: token.TIME_LIT, Literal: literal, Line: line, Column: column}
	}
	return token.Token{Type: token.DATE_LIT, Literal: literal, Line: line, Column: column}
}

func isTimeLiteral(literal string) bool {
	if len(literal) < 3 {
		return false
	}
	// Vérifier les formats de temps courants
	timeStr := literal[1 : len(literal)-1] // enlever les #
	_, err := time.Parse("15:04:05", timeStr)
	if err == nil {
		return true
	}
	_, err = time.Parse("15:04", timeStr)
	return err == nil
}

func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) || isDigit(l.ch) || l.ch == '_' {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readNumber() token.Token {
	line := l.line
	column := l.column
	position := l.position
	var tokType token.TokenType = token.INT_LIT
	// tokType = token.INT_LIT

	for isDigit(l.ch) {
		l.readChar()
	}

	if l.ch == '.' && isDigit(l.peekChar()) {
		tokType = token.FLOAT_LIT
		l.readChar()
		for isDigit(l.ch) {
			l.readChar()
		}
	}

	return token.Token{
		Type:    tokType,
		Literal: l.input[position:l.position],
		Line:    line,
		Column:  column,
	}
}

func (l *Lexer) readString() string {
	position := l.position + 1
	for {
		l.readChar()
		if l.ch == '"' || l.ch == 0 {
			break
		}
	}
	return l.input[position:l.position]
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *Lexer) peekChar() rune {
	if l.readPosition >= len(l.input) {
		return 0
	}
	ch, _ := utf8.DecodeRuneInString(l.input[l.readPosition:])
	return ch
}

func isLetter(ch rune) bool {
	return unicode.IsLetter(ch) || ch == '_'
}

func isDigit(ch rune) bool {
	return unicode.IsDigit(ch)
}

func newToken(tokenType token.TokenType, ch rune, line, column int) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch), Line: line, Column: column}
}
