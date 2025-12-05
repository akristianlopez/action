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
	ACTION   = "ACTION"
	START    = "START"
	STOP     = "STOP"
	LET      = "LET"
	FUNCTION = "FUNCTION"
	STRUCT   = "STRUCT"
	TYPE     = "TYPE"
	FOR      = "FOR"
	IF       = "IF"
	ELSE     = "ELSE"
	RETURN   = "RETURN"
	WHILE    = "WHILE"
	FOREACH  = "FOREACH"

	// Types
	INTEGER = "INTEGER"
	FLOAT   = "FLOAT"
	STRING  = "STRING"
	BOOLEAN = "BOOLEAN"
	TIME    = "TIME"
	DATE    = "DATE"

	// Identifiants et littéraux
	IDENT      = "IDENT"
	INT_LIT    = "INT_LIT"
	FLOAT_LIT  = "FLOAT_LIT"
	STRING_LIT = "STRING_LIT"
	BOOL_LIT   = "BOOL_LIT"
	TIME_LIT   = "TIME_LIT"
	DATE_LIT   = "DATE_LIT"

	// Opérateurs
	ASSIGN   = "="
	PLUS     = "+"
	MINUS    = "-"
	ASTERISK = "*"
	SLASH    = "/"
	MOD      = "%"

	// Comparaisons
	EQ     = "=="
	NOT_EQ = "!="
	LT     = "<"
	GT     = ">"
	LTE    = "<="
	GTE    = ">="

	// Délimiteurs
	COMMA     = ","
	SEMICOLON = ";"
	COLON     = ":"
	DOT       = "."

	LPAREN   = "("
	RPAREN   = ")"
	LBRACE   = "{"
	RBRACE   = "}"
	LBRACKET = "["
	RBRACKET = "]"

	// SQL
	SELECT = "SELECT"
	FROM   = "FROM"
	WHERE  = "WHERE"
	JOIN   = "JOIN"
	ON     = "ON"
	OBJECT = "OBJECT"
	AS     = "AS"

	// Spéciaux
	COMMENT = "COMMENT"
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	// SQL DDL
	CREATE     = "CREATE"
	DROP       = "DROP"
	ALTER      = "ALTER"
	TABLE      = "TABLE" // Gardé pour compatibilité, mais on utilisera OBJECT
	ADD        = "ADD"
	MODIFY     = "MODIFY"
	CONSTRAINT = "CONSTRAINT"
	PRIMARY    = "PRIMARY"
	KEY        = "KEY"
	FOREIGN    = "FOREIGN"
	REFERENCES = "REFERENCES"
	UNIQUE     = "UNIQUE"
	CHECK      = "CHECK"
	DEFAULT    = "DEFAULT"
	INDEX      = "INDEX"

	// SQL DML
	INSERT   = "INSERT"
	INTO     = "INTO"
	VALUES   = "VALUES"
	UPDATE   = "UPDATE"
	SET      = "SET"
	DELETE   = "DELETE"
	TRUNCATE = "TRUNCATE"

	// Clauses SQL supplémentaires
	ORDER    = "ORDER"
	BY       = "BY"
	GROUP    = "GROUP"
	HAVING   = "HAVING"
	LIMIT    = "LIMIT"
	OFFSET   = "OFFSET"
	DISTINCT = "DISTINCT"
	UNION    = "UNION"
	ALL      = "ALL"
	IN       = "IN"
	EXISTS   = "EXISTS"
	LIKE     = "LIKE"
	BETWEEN  = "BETWEEN"
	IS       = "IS"
	NULL     = "NULL"
	NOT      = "NOT"
	AND      = "AND"
	OR       = "OR"
	ASC      = "ASC"
	DESC     = "DESC"

	// Types SQL
	VARCHAR   = "VARCHAR"
	CHAR      = "CHAR"
	NUMERIC   = "NUMERIC"
	DECIMAL   = "DECIMAL"
	TIMESTAMP = "TIMESTAMP"
	DATETIME  = "DATETIME"

	// SQL récursif
	WITH        = "WITH"
	RECURSIVE   = "RECURSIVE"
	WINDOW      = "WINDOW"
	OVER        = "OVER"
	PARTITION   = "PARTITION"
	ROW         = "ROW"
	ROWS        = "ROWS"
	RANGE       = "RANGE"
	PRECEDING   = "PRECEDING"
	FOLLOWING   = "FOLLOWING"
	CURRENT     = "CURRENT"
	UNBOUNDED   = "UNBOUNDED"
	LAG         = "LAG"
	LEAD        = "LEAD"
	FIRST_VALUE = "FIRST_VALUE"
	LAST_VALUE  = "LAST_VALUE"
	RANK        = "RANK"
	DENSE_RANK  = "DENSE_RANK"
	ROW_NUMBER  = "ROW_NUMBER"
	NTILE       = "NTILE"
	CONNECT     = "CONNECT"
	PRIOR       = "PRIOR"
	NOCYCLE     = "NOCYCLE"
	SIBLINGS    = "SIBLINGS"
	CASCADE     = "CASCADE"

	// Tableaux
	ARRAY = "ARRAY"
	OF    = "OF"

	// Opérateurs de tableau
	CONCAT = "||"
	NOT_IN = "NOT IN"

	// Fonctions de tableau
	LENGTH   = "LENGTH"
	APPEND   = "APPEND"
	PREPEND  = "PREPEND"
	REMOVE   = "REMOVE"
	SLICE    = "SLICE"
	CONTAINS = "CONTAINS"

	SWITCH      = "SWITCH"
	CASE        = "CASE"
	BREAK       = "BREAK"
	FALLTHROUGH = "FALLTHROUGH"
)

var keywords = map[string]TokenType{
	"action":   ACTION,
	"start":    START,
	"stop":     STOP,
	"let":      LET,
	"function": FUNCTION,
	"struct":   STRUCT,
	"type":     TYPE,
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
	"while":    WHILE,
	"foreach":  FOREACH,

	// SQL DDL
	"create":     CREATE,
	"drop":       DROP,
	"alter":      ALTER,
	"table":      TABLE,
	"add":        ADD,
	"modify":     MODIFY,
	"constraint": CONSTRAINT,
	"primary":    PRIMARY,
	"key":        KEY,
	"foreign":    FOREIGN,
	"references": REFERENCES,
	"unique":     UNIQUE,
	"check":      CHECK,
	"default":    DEFAULT,
	"index":      INDEX,

	// SQL DML
	"insert":   INSERT,
	"into":     INTO,
	"values":   VALUES,
	"update":   UPDATE,
	"set":      SET,
	"delete":   DELETE,
	"truncate": TRUNCATE,

	// Clauses supplémentaires
	"order":    ORDER,
	"by":       BY,
	"group":    GROUP,
	"having":   HAVING,
	"limit":    LIMIT,
	"offset":   OFFSET,
	"distinct": DISTINCT,
	"union":    UNION,
	"all":      ALL,
	"in":       IN,
	"exists":   EXISTS,
	"like":     LIKE,
	"between":  BETWEEN,
	"is":       IS,
	"null":     NULL,
	"not":      NOT,
	"and":      AND,
	"or":       OR,
	"asc":      ASC,
	"desc":     DESC,
	"connect":  CONNECT,
	"prior":    PRIOR,
	"nocycle":  NOCYCLE,
	"siblings": SIBLINGS,
	"cascade":  CASCADE,

	// Types SQL
	"varchar":   VARCHAR,
	"char":      CHAR,
	"numeric":   NUMERIC,
	"decimal":   DECIMAL,
	"timestamp": TIMESTAMP,
	"datetime":  DATETIME,

	// SQL récursif et analytique
	"with":        WITH,
	"recursive":   RECURSIVE,
	"window":      WINDOW,
	"over":        OVER,
	"partition":   PARTITION,
	"row":         ROW,
	"rows":        ROWS,
	"range":       RANGE,
	"preceding":   PRECEDING,
	"following":   FOLLOWING,
	"current":     CURRENT,
	"unbounded":   UNBOUNDED,
	"lag":         LAG,
	"lead":        LEAD,
	"first_value": FIRST_VALUE,
	"last_value":  LAST_VALUE,
	"rank":        RANK,
	"dense_rank":  DENSE_RANK,
	"row_number":  ROW_NUMBER,
	"ntile":       NTILE,
	"array":       ARRAY,
	"of":          OF,
	"length":      LENGTH,
	"append":      APPEND,
	"prepend":     PREPEND,
	"remove":      REMOVE,
	"slice":       SLICE,
	"contains":    CONTAINS,

	// Switch statement
	"switch":      SWITCH,
	"case":        CASE,
	"break":       BREAK,
	"fallthrough": FALLTHROUGH,
}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
