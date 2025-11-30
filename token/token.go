package token

type TokenType string

type Token struct {
    Type    TokenType
    Literal string
    Line    int
    Column  int
}

const (
    // Mots-clés
    ACTION     = "ACTION"
    START      = "START"
    STOP       = "STOP"
    LET        = "LET"
    FUNCTION   = "FUNCTION"
    STRUCT     = "STRUCT"
    FOR        = "FOR"
    IF         = "IF"
    ELSE       = "ELSE"
    RETURN     = "RETURN"
    
    // Types
    INTEGER    = "INTEGER"
    FLOAT      = "FLOAT"
    STRING     = "STRING"
    BOOLEAN    = "BOOLEAN"
    TIME       = "TIME"
    DATE       = "DATE"
    
    // Identifiants et littéraux
    IDENT      = "IDENT"
    INT_LIT    = "INT_LIT"
    FLOAT_LIT  = "FLOAT_LIT"
    STRING_LIT = "STRING_LIT"
    BOOL_LIT   = "BOOL_LIT"
    TIME_LIT   = "TIME_LIT"
    DATE_LIT   = "DATE_LIT"
    
    // Opérateurs
    ASSIGN     = "="
    PLUS       = "+"
    MINUS      = "-"
    ASTERISK   = "*"
    SLASH      = "/"
    MOD        = "%"
    
    // Comparaisons
    EQ         = "=="
    NOT_EQ     = "!="
    LT         = "<"
    GT         = ">"
    LTE        = "<="
    GTE        = ">="
    
    // Délimiteurs
    COMMA      = ","
    SEMICOLON  = ";"
    COLON      = ":"
    DOT        = "."
    
    LPAREN     = "("
    RPAREN     = ")"
    LBRACE     = "{"
    RBRACE     = "}"
    LBRACKET   = "["
    RBRACKET   = "]"
    
    // SQL
    SELECT     = "SELECT"
    FROM       = "FROM"
    WHERE      = "WHERE"
    JOIN       = "JOIN"
    ON         = "ON"
    OBJECT     = "OBJECT"
    AS         = "AS"
    
    // Spéciaux
    COMMENT    = "COMMENT"
    ILLEGAL    = "ILLEGAL"
    EOF        = "EOF"
)

var keywords = map[string]TokenType{
    "action":   ACTION,
    "start":    START,
    "stop":     STOP,
    "let":      LET,
    "function": FUNCTION,
    "struct":   STRUCT,
    "for":      FOR,
    "if":       IF,
    "else":     ELSE,
    "return":   RETURN,
    "integer":  INTEGER,
    "float":    FLOAT,
    "string":   STRING,
    "boolean":  BOOLEAN,
    "time":     TIME,
    "date":     DATE,
    "select":   SELECT,
    "from":     FROM,
    "where":    WHERE,
    "join":     JOIN,
    "on":       ON,
    "object":   OBJECT,
    "as":       AS,
    "true":     BOOL_LIT,
    "false":    BOOL_LIT,
}

func LookupIdent(ident string) TokenType {
    if tok, ok := keywords[ident]; ok {
        return tok
    }
    return IDENT
}