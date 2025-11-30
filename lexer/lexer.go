package lexer

import (
	"strings"

	"github.com/akristianlopez/action/token"
)

type Lexer struct {
	input        string
	position     int  // position courante
	readPosition int  // prochaine position
	ch           byte // caractère courant
	line         int
	column       int
}

func New(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition += 1
}

func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	l.skipWhitespace()
	line := l.line
	column := l.column

	switch l.ch {
	case '=':
		if l.peekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.EQ, Literal: "==", Line: line, Column: column}
		} else {
			tok = newToken(token.ASSIGN, l.ch, line, column)
		}
	case '+':
		tok = newToken(token.PLUS, l.ch, line, column)
	case '-':
		tok = newToken(token.MINUS, l.ch, line, column)
	case '!':
		if l.peekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.NOT_EQ, Literal: "!=", Line: line, Column: column}
		} else {
			tok = newToken(token.BANG, l.ch, line, column)
		}
	case '/':
		tok = newToken(token.SLASH, l.ch, line, column)
	case '*':
		tok = newToken(token.ASTERISK, l.ch, line, column)
	case '<':
		tok = newToken(token.LT, l.ch, line, column)
	case '>':
		tok = newToken(token.GT, l.ch, line, column)
	case ';':
		tok = newToken(token.SEMICOLON, l.ch, line, column)
	case '(':
		tok = newToken(token.LPAREN, l.ch, line, column)
	case ')':
		tok = newToken(token.RPAREN, l.ch, line, column)
	case ',':
		tok = newToken(token.COMMA, l.ch, line, column)
	case '{':
		tok = newToken(token.LBRACE, l.ch, line, column)
	case '}':
		tok = newToken(token.RBRACE, l.ch, line, column)
	case ':':
		tok = newToken(token.COLON, l.ch, line, column)
	case '.':
		tok = newToken(token.DOT, l.ch, line, column)
	case '"':
		tok.Type = token.STRING
		tok.Literal = l.readString()
		tok.Line = line
		tok.Column = column
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
		tok.Line = line
		tok.Column = column
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = token.LookupIdent(tok.Literal)
			tok.Line = line
			tok.Column = column
			return tok
		} else if isDigit(l.ch) {
			tok.Type = token.INT
			tok.Literal = l.readNumber()
			tok.Line = line
			tok.Column = column
			return tok
		} else {
			tok = newToken(token.ILLEGAL, l.ch, line, column)
		}
	}

	l.readChar()
	return tok
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

// Ajouter la fonction pour lire les strings
func (l *Lexer) readString() string {
	var out strings.Builder
	l.readChar() // sauter le premier guillemet

	for l.ch != '"' {
		if l.ch == 0 {
			return "" // erreur: string non fermée
		}
		if l.ch == '\\' {
			l.readChar()
			switch l.ch {
			case 'n':
				out.WriteByte('\n')
			case 't':
				out.WriteByte('\t')
			case '"':
				out.WriteByte('"')
			case '\\':
				out.WriteByte('\\')
			default:
				out.WriteByte('\\')
				out.WriteByte(l.ch)
			}
		} else {
			out.WriteByte(l.ch)
		}
		l.readChar()
	}

	return out.String()
}

func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readNumber() string {
	position := l.position
	for isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func newToken(tokenType token.TokenType, ch byte, line int, col int) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch), Line: line, Column: col}
}
