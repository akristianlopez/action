package ast

import (
	"fmt"
	"strings"

	"github.com/akristianlopez/action/token"
)

// Node interface de base pour tous les nœuds AST
type Node interface {
	TokenLiteral() string
	String() string
}

// Statement interface pour les instructions
type Statement interface {
	Node
	statementNode()
}

// Expression interface pour les expressions
type Expression interface {
	Node
	expressionNode()
	Line() int
	Column() int
}

// Program - le programme racine
type Program struct {
	ActionName string
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

func (p *Program) String() string {
	var out string
	for _, s := range p.Statements {
		out += s.String()
	}
	return out
}

// Let TypeMember - membre de type (pour les types composés)
type TypeMember struct {
	Token token.Token
	Left  Expression
	Right Expression
}

func (tm *TypeMember) statementNode()       {}
func (tm *TypeMember) TokenLiteral() string { return tm.Token.Literal }
func (tm *TypeMember) String() string {
	return tm.Left.String() + "." + tm.Right.String()
}

// LetStatement - déclaration de variable
type LetStatement struct {
	Token token.Token
	Name  *Identifier
	Type  *TypeAnnotation
	Value Expression
}

func (ls *LetStatement) statementNode()       {}
func (ls *LetStatement) TokenLiteral() string { return ls.Token.Literal }
func (ls *LetStatement) String() string {
	var out string
	out += ls.TokenLiteral() + " "
	out += ls.Name.String()
	if ls.Type != nil {
		out += " : " + ls.Type.String()
	}
	if ls.Value != nil {
		out += " = " + ls.Value.String()
	}
	out += ";"
	return out
}

// Type liste de declarations exemple Let a:type1, b:type2
type LetStatements []LetStatement

func (lm *LetStatements) statementNode()       {}
func (lm *LetStatements) TokenLiteral() string { return "" }
func (lm *LetStatements) String() string {
	var out string = ""
	for _, ls := range *lm {
		out += ls.TokenLiteral() + " "
		out += ls.Name.String()
		if ls.Type != nil {
			out += " : " + ls.Type.String()
		}
		if ls.Value != nil {
			out += " = " + ls.Value.String()
		}
		out += ", "
	}
	return strings.TrimRight(out, ", ")
	// return fmt.Sprintf("Let %s", out)
}

// TypeConstraints - contraintes de type
type TypeConstraints struct {
	MaxDigits     *IntegerLiteral
	IntegerRange  *RangeConstraint
	DecimalPlaces *IntegerLiteral
	MaxLength     *IntegerLiteral
}

func (tc *TypeConstraints) String() string {
	var out string
	if tc.MaxDigits != nil {
		out += "(" + tc.MaxDigits.String() + ")"
	}
	if tc.IntegerRange != nil {
		out += "[" + tc.IntegerRange.String() + "]"
	}
	if tc.DecimalPlaces != nil {
		out += "(" + tc.MaxDigits.String() + "." + tc.DecimalPlaces.String() + ")"
	}
	if tc.MaxLength != nil {
		out += "(" + tc.MaxLength.String() + ")"
	}
	return out
}

// RangeConstraint - contrainte de plage
type RangeConstraint struct {
	Min Expression
	Max Expression
}

func (rc *RangeConstraint) String() string {
	return rc.Min.String() + ".." + rc.Max.String()
}

// Identifier - identifiant
type Identifier struct {
	Token token.Token
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) Line() int            { return i.Token.Line }
func (i *Identifier) Column() int          { return i.Token.Column }
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string       { return i.Value }

// IntegerLiteral - littéral entier
type IntegerLiteral struct {
	Token token.Token
	Value int64
}

func (il *IntegerLiteral) expressionNode()      {}
func (i *IntegerLiteral) Line() int             { return i.Token.Line }
func (i *IntegerLiteral) Column() int           { return i.Token.Column }
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) String() string       { return il.Token.Literal }

// FloatLiteral - littéral flottant
type FloatLiteral struct {
	Token token.Token
	Value float64
}

func (fl *FloatLiteral) expressionNode()      {}
func (i *FloatLiteral) Line() int             { return i.Token.Line }
func (i *FloatLiteral) Column() int           { return i.Token.Column }
func (fl *FloatLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FloatLiteral) String() string       { return fl.Token.Literal }

// StringLiteral - littéral chaîne
type StringLiteral struct {
	Token token.Token
	Value string
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) Line() int            { return sl.Token.Line }
func (sl *StringLiteral) Column() int          { return sl.Token.Column }
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }
func (sl *StringLiteral) String() string       { return "\"" + sl.Value + "\"" }

// BooleanLiteral - littéral booléen
type BooleanLiteral struct {
	Token token.Token
	Value bool
}

func (b *BooleanLiteral) expressionNode()      {}
func (b *BooleanLiteral) Line() int            { return b.Token.Line }
func (b *BooleanLiteral) Column() int          { return b.Token.Column }
func (b *BooleanLiteral) TokenLiteral() string { return b.Token.Literal }
func (b *BooleanLiteral) String() string       { return b.Token.Literal }

// DateTimeLiteral - littéral date/time
type DateTimeLiteral struct {
	Token  token.Token
	Value  string
	IsTime bool
}

func (dt *DateTimeLiteral) expressionNode()      {}
func (dt *DateTimeLiteral) Line() int            { return dt.Token.Line }
func (dt *DateTimeLiteral) Column() int          { return dt.Token.Column }
func (dt *DateTimeLiteral) TokenLiteral() string { return dt.Token.Literal }
func (dt *DateTimeLiteral) String() string       { return dt.Token.Literal }

// ExpressionStatement - instruction d'expression
type ExpressionStatement struct {
	Token      token.Token
	Expression Expression
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

// PrefixExpression - expression préfixe
type PrefixExpression struct {
	Token    token.Token
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) Line() int            { return pe.Token.Line }
func (pe *PrefixExpression) Column() int          { return pe.Token.Column }
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }
func (pe *PrefixExpression) String() string {
	return "(" + pe.Operator + pe.Right.String() + ")"
}

// InfixExpression - expression infixe
type InfixExpression struct {
	Token    token.Token
	Left     Expression
	Operator string
	Right    Expression
}

func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) Line() int            { return ie.Token.Line }
func (ie *InfixExpression) Column() int          { return ie.Token.Column }
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *InfixExpression) String() string {
	return "(" + ie.Left.String() + " " + ie.Operator + " " + ie.Right.String() + ")"
}

// BlockStatement - bloc d'instructions
type BlockStatement struct {
	Token      token.Token
	Statements []Statement
}

func (bs *BlockStatement) statementNode()       {}
func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BlockStatement) String() string {
	var out string
	for _, s := range bs.Statements {
		out += s.String()
	}
	return out
}

// ForStatement - boucle for
type ForStatement struct {
	Token     token.Token
	Init      Statement
	Condition Expression
	Update    Statement
	Body      *BlockStatement
}

func (fs *ForStatement) statementNode()       {}
func (fs *ForStatement) TokenLiteral() string { return fs.Token.Literal }
func (fs *ForStatement) String() string {
	var out string
	out += "for "
	if fs.Init != nil {
		out += fs.Init.String()
	}
	out += "; "
	if fs.Condition != nil {
		out += fs.Condition.String()
	}
	out += "; "
	if fs.Update != nil {
		out += fs.Update.String()
	}
	out += " " + fs.Body.String()
	return out
}

// ForStatement - If
type IfStatement struct {
	Token     token.Token
	Condition Expression
	Then      *BlockStatement
	Else      *BlockStatement
}

func (is *IfStatement) statementNode()       {}
func (is *IfStatement) TokenLiteral() string { return is.Token.Literal }
func (is *IfStatement) String() string {
	var out string
	out += "if("
	if is.Condition != nil {
		out += is.Condition.String()
	}
	out += ")\n"
	if is.Then != nil {
		out += is.Then.String()
	}
	if is.Else != nil {
		out += " Else\n " + is.Else.String()
	}
	return out
}

// ForStatement - While
type WhileStatement struct {
	Token     token.Token
	Condition Expression
	Body      Statement
}

func (ws *WhileStatement) statementNode()       {}
func (ws *WhileStatement) TokenLiteral() string { return ws.Token.Literal }
func (ws *WhileStatement) String() string {
	var out string
	out += "While ("
	if ws.Condition != nil {
		out += ws.Condition.String()
	}
	out += ") "
	out += " " + ws.Body.String()
	return out
}

// ForStatement - ForEach
type ForEachStatement struct {
	Token    token.Token
	Variable *Identifier
	Iterator Expression
	Body     Statement
}

func (fe *ForEachStatement) statementNode()       {}
func (fe *ForEachStatement) TokenLiteral() string { return fe.Token.Literal }
func (fe *ForEachStatement) String() string {
	var out string
	out += "ForEach "
	if fe.Variable != nil {
		out += fe.Variable.String()
	}
	out += fmt.Sprintf(" IN (%s)", fe.Iterator.String())
	out += " " + fe.Body.String()
	return out
}

// FunctionStatement - déclaration de fonction
type FunctionStatement struct {
	Token      token.Token
	Name       *Identifier
	Parameters []*FunctionParameter
	ReturnType *TypeAnnotation
	Body       *BlockStatement
}

func (fs *FunctionStatement) statementNode()       {}
func (fs *FunctionStatement) TokenLiteral() string { return fs.Token.Literal }
func (fs *FunctionStatement) String() string {
	var out string
	out += "function " + fs.Name.String() + "("
	for i, p := range fs.Parameters {
		if i > 0 {
			out += ", "
		}
		out += p.String()
	}
	out += ")"
	if fs.ReturnType != nil {
		out += " : " + fs.ReturnType.String()
	}
	out += " " + fs.Body.String()
	return out
}

// FunctionParameter - paramètre de fonction
type FunctionParameter struct {
	Token token.Token
	Name  *Identifier
	Type  *TypeAnnotation
}

func (fp *FunctionParameter) String() string {
	return fp.Name.String() + " : " + fp.Type.String()
}
func (fp *FunctionParameter) TokenLiteral() string {
	return fp.Token.Literal
}

// StructStatement - déclaration de structure
type StructStatement struct {
	Token  token.Token
	Name   *Identifier
	Fields []*StructField
}

func (ss *StructStatement) statementNode()       {}
func (ss *StructStatement) TokenLiteral() string { return ss.Token.Literal }
func (ss *StructStatement) String() string {
	var out string
	out += "struct " + ss.Name.String() + " { "
	for i, f := range ss.Fields {
		if i > 0 {
			out += ", "
		}
		out += f.String()
	}
	out += " }"
	return out
}

// StructField - champ de structure
type StructField struct {
	Token token.Token
	Name  *Identifier
	Type  *TypeAnnotation
}

func (sf *StructField) String() string {
	return sf.Name.String() + " : " + sf.Type.String()
}

// SQLCreateObjectStatement - CREATE OBJECT
type SQLCreateObjectStatement struct {
	Token       token.Token
	ObjectName  *Identifier
	Columns     []*SQLColumnDefinition
	Constraints []*SQLConstraint
	IfNotExists bool
}

func (sc *SQLCreateObjectStatement) statementNode()       {}
func (sc *SQLCreateObjectStatement) TokenLiteral() string { return sc.Token.Literal }
func (sc *SQLCreateObjectStatement) String() string {
	var out string
	out += "CREATE OBJECT "
	if sc.IfNotExists {
		out += "IF NOT EXISTS "
	}
	out += sc.ObjectName.String() + " ("
	for i, col := range sc.Columns {
		if i > 0 {
			out += ", "
		}
		out += col.String()
	}
	for _, constraint := range sc.Constraints {
		out += ", " + constraint.String()
	}
	out += ")"
	return out
}

// SQLColumnDefinition - Définition de colonne
type SQLColumnDefinition struct {
	Token       token.Token
	Name        *Identifier
	DataType    *SQLDataType
	Constraints []*SQLColumnConstraint
}

func (sc *SQLColumnDefinition) String() string {
	out := sc.Name.String() + " " + sc.DataType.String()
	for _, constraint := range sc.Constraints {
		out += " " + constraint.String()
	}
	return out
}

// SQLDataType - Type de données SQL
type SQLDataType struct {
	Token     token.Token
	Name      string
	Length    *IntegerLiteral
	Precision *IntegerLiteral
	Scale     *IntegerLiteral
}

func (sd *SQLDataType) String() string {
	out := sd.Name
	if sd.Length != nil {
		out += "(" + sd.Length.String() + ")"
	} else if sd.Precision != nil && sd.Scale != nil {
		out += "(" + sd.Precision.String() + "," + sd.Scale.String() + ")"
	} else if sd.Precision != nil {
		out += "(" + sd.Precision.String() + ")"
	}
	return out
}

// SQLColumnConstraint - Contrainte de colonne
type SQLColumnConstraint struct {
	Token      token.Token
	Type       string // NOT NULL, UNIQUE, etc.
	Expression Expression
}

func (scc *SQLColumnConstraint) String() string {
	out := scc.Type
	if scc.Expression != nil {
		out += " " + scc.Expression.String()
	}
	return out
}

// SQLConstraint - Contrainte de table
type SQLConstraint struct {
	Token      token.Token
	Name       *Identifier
	Type       string // PRIMARY KEY, FOREIGN KEY, etc.
	Columns    []*Identifier
	References *SQLReference
	Check      Expression
}

func (sc *SQLConstraint) String() string {
	out := "CONSTRAINT " + sc.Name.String() + " " + sc.Type
	if len(sc.Columns) > 0 {
		out += " ("
		for i, col := range sc.Columns {
			if i > 0 {
				out += ", "
			}
			out += col.String()
		}
		out += ")"
	}
	if sc.References != nil {
		out += " " + sc.References.String()
	}
	if sc.Check != nil {
		out += " CHECK " + sc.Check.String()
	}
	return out
}

// SQLReference - Référence FOREIGN KEY
type SQLReference struct {
	Token     token.Token
	TableName *Identifier
	Columns   []*Identifier
}

func (sr *SQLReference) String() string {
	out := "REFERENCES " + sr.TableName.String()
	if len(sr.Columns) > 0 {
		out += " ("
		for i, col := range sr.Columns {
			if i > 0 {
				out += ", "
			}
			out += col.String()
		}
		out += ")"
	}
	return out
}

// SQLDropObjectStatement - DROP OBJECT
type SQLDropObjectStatement struct {
	Token      token.Token
	ObjectName *Identifier
	IfExists   bool
	Cascade    bool
}

func (sd *SQLDropObjectStatement) statementNode()       {}
func (sd *SQLDropObjectStatement) TokenLiteral() string { return sd.Token.Literal }
func (sd *SQLDropObjectStatement) String() string {
	out := "DROP OBJECT "
	if sd.IfExists {
		out += "IF EXISTS "
	}
	out += sd.ObjectName.String()
	if sd.Cascade {
		out += " CASCADE"
	}
	return out
}

// SQLAlterObjectStatement - ALTER OBJECT
type SQLAlterObjectStatement struct {
	Token      token.Token
	ObjectName *Identifier
	Actions    []*SQLAlterAction
}

func (sa *SQLAlterObjectStatement) statementNode()       {}
func (sa *SQLAlterObjectStatement) TokenLiteral() string { return sa.Token.Literal }
func (sa *SQLAlterObjectStatement) String() string {
	out := "ALTER OBJECT " + sa.ObjectName.String()
	for i, action := range sa.Actions {
		if i > 0 {
			out += ", "
		}
		out += " " + action.String()
	}
	return out
}

// SQLAlterAction - Action ALTER
type SQLAlterAction struct {
	Token      token.Token
	Type       string // ADD, MODIFY, DROP
	Column     *SQLColumnDefinition
	Constraint *SQLConstraint
	ColumnName *Identifier
}

func (sa *SQLAlterAction) String() string {
	out := sa.Type + " "
	if sa.Column != nil {
		out += sa.Column.String()
	} else if sa.Constraint != nil {
		out += sa.Constraint.String()
	} else if sa.ColumnName != nil {
		out += sa.ColumnName.String()
	}
	return out
}

// SQLInsertStatement - INSERT
type SQLInsertStatement struct {
	Token      token.Token
	ObjectName *Identifier
	Columns    []*Identifier
	Values     []*SQLValues
	Select     *SQLSelectStatement
}

func (si *SQLInsertStatement) statementNode()       {}
func (si *SQLInsertStatement) TokenLiteral() string { return si.Token.Literal }
func (si *SQLInsertStatement) String() string {
	out := "INSERT INTO " + si.ObjectName.String()
	if len(si.Columns) > 0 {
		out += " ("
		for i, col := range si.Columns {
			if i > 0 {
				out += ", "
			}
			out += col.String()
		}
		out += ")"
	}
	if si.Select != nil {
		out += " " + si.Select.String()
	} else {
		out += " VALUES"
		for i, values := range si.Values {
			if i > 0 {
				out += ","
			}
			out += " " + values.String()
		}
	}
	return out
}
func (si *SQLInsertStatement) expressionNode() {}
func (si *SQLInsertStatement) Line() int       { return si.Token.Line }
func (si *SQLInsertStatement) Column() int     { return si.Token.Column }

// SQLValues - Valeurs pour INSERT
type SQLValues struct {
	Token  token.Token
	Values []Expression
}

func (sv *SQLValues) String() string {
	out := "("
	for i, val := range sv.Values {
		if i > 0 {
			out += ", "
		}
		out += val.String()
	}
	out += ")"
	return out
}

// SQLUpdateStatement - UPDATE
type SQLUpdateStatement struct {
	Token      token.Token
	ObjectName *Identifier
	Set        []*SQLSetClause
	Where      Expression
}

func (su *SQLUpdateStatement) statementNode()       {}
func (su *SQLUpdateStatement) TokenLiteral() string { return su.Token.Literal }
func (su *SQLUpdateStatement) String() string {
	out := "UPDATE " + su.ObjectName.String() + " SET "
	for i, set := range su.Set {
		if i > 0 {
			out += ", "
		}
		out += set.String()
	}
	if su.Where != nil {
		out += " WHERE " + su.Where.String()
	}
	return out
}
func (su *SQLUpdateStatement) expressionNode() {}
func (su *SQLUpdateStatement) Line() int       { return su.Token.Line }
func (su *SQLUpdateStatement) Column() int     { return su.Token.Column }

// SQLSetClause - Clause SET pour UPDATE
type SQLSetClause struct {
	Token  token.Token
	Column *Identifier
	Value  Expression
}

func (ss *SQLSetClause) String() string {
	return ss.Column.String() + " = " + ss.Value.String()
}

// SQLDeleteStatement - DELETE
type SQLDeleteStatement struct {
	Token token.Token
	From  *Identifier
	Where Expression
}

func (sd *SQLDeleteStatement) statementNode()       {}
func (sd *SQLDeleteStatement) TokenLiteral() string { return sd.Token.Literal }
func (sd *SQLDeleteStatement) String() string {
	out := "DELETE FROM " + sd.From.String()
	if sd.Where != nil {
		out += " WHERE " + sd.Where.String()
	}
	return out
}
func (sd *SQLDeleteStatement) expressionNode() {}
func (sd *SQLDeleteStatement) Line() int       { return sd.Token.Line }
func (sd *SQLDeleteStatement) Column() int     { return sd.Token.Column }

// SQLTruncateStatement - TRUNCATE
type SQLTruncateStatement struct {
	Token      token.Token
	ObjectName *Identifier
}

func (st *SQLTruncateStatement) statementNode()       {}
func (st *SQLTruncateStatement) TokenLiteral() string { return st.Token.Literal }
func (st *SQLTruncateStatement) String() string {
	return "TRUNCATE OBJECT " + st.ObjectName.String()
}
func (st *SQLTruncateStatement) expressionNode() {}
func (st *SQLTruncateStatement) Line() int       { return st.Token.Line }
func (st *SQLTruncateStatement) Column() int     { return st.Token.Column }

// SQLCreateIndexStatement - CREATE INDEX
type SQLCreateIndexStatement struct {
	Token      token.Token
	IndexName  *Identifier
	ObjectName *Identifier
	Columns    []*Identifier
	Unique     bool
}

func (si *SQLCreateIndexStatement) statementNode()       {}
func (si *SQLCreateIndexStatement) TokenLiteral() string { return si.Token.Literal }
func (si *SQLCreateIndexStatement) String() string {
	out := "CREATE "
	if si.Unique {
		out += "UNIQUE "
	}
	out += "INDEX " + si.IndexName.String() + " ON " + si.ObjectName.String() + " ("
	for i, col := range si.Columns {
		if i > 0 {
			out += ", "
		}
		out += col.String()
	}
	out += ")"
	return out
}

// SQLOrderBy - Clause ORDER BY
type SQLOrderBy struct {
	Expression Expression
	Direction  string // ASC ou DESC
}

func (so *SQLOrderBy) String() string {
	out := so.Expression.String()
	if so.Direction != "" {
		out += " " + so.Direction
	}
	return out
}

// SQLJoin - clause JOIN
type SQLJoin struct {
	Token token.Token
	Type  string // INNER, LEFT, etc.
	Table Expression
	On    Expression
}

func (sj *SQLJoin) String() string {
	return sj.Type + " JOIN " + sj.Table.String() + " ON " + sj.On.String()
}

// ReturnStatement - instruction return
type ReturnStatement struct {
	Token       token.Token
	ReturnValue Expression
}

func (rs *ReturnStatement) statementNode()       {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }
func (rs *ReturnStatement) String() string {
	var out string
	out += "return"
	if rs.ReturnValue != nil {
		out += " " + rs.ReturnValue.String()
	}
	out += ";"
	return out
}

// SQLWithStatement - Clause WITH pour les CTE
type SQLWithStatement struct {
	Token     token.Token
	Recursive bool
	CTEs      []*SQLCommonTableExpression
	Select    *SQLSelectStatement
}

func (sw *SQLWithStatement) statementNode()       {}
func (sw *SQLWithStatement) TokenLiteral() string { return sw.Token.Literal }
func (sw *SQLWithStatement) String() string {
	out := "WITH "
	if sw.Recursive {
		out += "RECURSIVE "
	}
	for i, cte := range sw.CTEs {
		if i > 0 {
			out += ", "
		}
		out += cte.String()
	}
	out += " " + sw.Select.String()
	return out
}
func (sw *SQLWithStatement) expressionNode() {}
func (sw *SQLWithStatement) Line() int       { return sw.Token.Line }
func (sw *SQLWithStatement) Column() int     { return sw.Token.Column }

// SQLCommonTableExpression - CTE
type SQLCommonTableExpression struct {
	Token   token.Token
	Name    *Identifier
	Columns []*Identifier
	Query   *SQLSelectStatement
}

func (cte *SQLCommonTableExpression) String() string {
	out := cte.Name.String()
	if len(cte.Columns) > 0 {
		out += " ("
		for i, col := range cte.Columns {
			if i > 0 {
				out += ", "
			}
			out += col.String()
		}
		out += ")"
	}
	out += " AS (" + cte.Query.String() + ")"
	return out
}

// SQLWindowFunction - Fonction de fenêtrage
type SQLWindowFunction struct {
	Token     token.Token
	Name      string
	Arguments []Expression
	Over      *SQLWindowClause
}

func (sw *SQLWindowFunction) expressionNode()      {}
func (sw *SQLWindowFunction) TokenLiteral() string { return sw.Token.Literal }
func (sw *SQLWindowFunction) String() string {
	out := sw.Name + "("
	for i, arg := range sw.Arguments {
		if i > 0 {
			out += ", "
		}
		out += arg.String()
	}
	out += ")"
	if sw.Over != nil {
		out += " OVER " + sw.Over.String()
	}
	return out
}
func (sw *SQLWindowFunction) Line() int   { return sw.Token.Line }
func (sw *SQLWindowFunction) Column() int { return sw.Token.Column }

// SQLWindowClause - Clause OVER
type SQLWindowClause struct {
	Token     token.Token
	Name      *Identifier
	Partition []Expression
	OrderBy   []*SQLOrderBy
	Frame     *SQLWindowFrame
}

func (sw *SQLWindowClause) String() string {
	out := "("
	if sw.Name != nil {
		out += sw.Name.String()
	}
	if len(sw.Partition) > 0 {
		out += " PARTITION BY "
		for i, expr := range sw.Partition {
			if i > 0 {
				out += ", "
			}
			out += expr.String()
		}
	}
	if len(sw.OrderBy) > 0 {
		out += " ORDER BY "
		for i, ob := range sw.OrderBy {
			if i > 0 {
				out += ", "
			}
			out += ob.String()
		}
	}
	if sw.Frame != nil {
		out += " " + sw.Frame.String()
	}
	out += ")"
	return out
}
func (sw *SQLWindowClause) expressionNode() {}
func (sw *SQLWindowClause) Line() int       { return sw.Token.Line }
func (sw *SQLWindowClause) Column() int     { return sw.Token.Column }

// SQLWindowFrame - Cadre de fenêtre
type SQLWindowFrame struct {
	Token token.Token
	Type  string // ROWS, RANGE
	Start *SQLWindowFrameBound
	End   *SQLWindowFrameBound
}

func (sw *SQLWindowFrame) String() string {
	out := sw.Type + " BETWEEN " + sw.Start.String()
	if sw.End != nil {
		out += " AND " + sw.End.String()
	} else {
		out += " AND CURRENT ROW"
	}
	return out
}

// SQLWindowFrameBound - Borne de fenêtre
type SQLWindowFrameBound struct {
	Token     token.Token
	Type      string // PRECEDING, FOLLOWING, CURRENT
	Value     Expression
	Unbounded bool
}

func (sw *SQLWindowFrameBound) String() string {
	if sw.Unbounded {
		return "UNBOUNDED " + sw.Type
	}
	if sw.Value != nil {
		return sw.Value.String() + " " + sw.Type
	}
	return "CURRENT ROW"
}

// SQLHierarchicalQuery - Requête hiérarchique
type SQLHierarchicalQuery struct {
	Token         token.Token
	StartWith     Expression
	ConnectBy     Expression
	Prior         bool
	Nocycle       bool
	OrderSiblings bool
}

func (sh *SQLHierarchicalQuery) String() string {
	out := ""
	if sh.StartWith != nil {
		out += " START WITH " + sh.StartWith.String()
	}
	if sh.ConnectBy != nil {
		out += " CONNECT BY "
		if sh.Prior {
			out += "PRIOR "
		}
		out += sh.ConnectBy.String()
	}
	if sh.Nocycle {
		out += " NOCYCLE"
	}
	if sh.OrderSiblings {
		out += " ORDER SIBLINGS"
	}
	return out
}

// Étendre SQLSelectStatement pour inclure les fonctionnalités récursives
type SQLSelectStatement struct {
	Token         token.Token
	Distinct      bool
	Select        []Expression
	From          Expression
	Joins         []*SQLJoin
	Where         Expression
	GroupBy       []Expression
	Having        Expression
	OrderBy       []*SQLOrderBy
	Limit         Expression
	Offset        Expression
	Union         *SQLSelectStatement
	UnionAll      bool
	With          *SQLWithStatement
	Hierarchical  *SQLHierarchicalQuery
	WindowClauses []*SQLWindowClause
}

func (ss *SQLSelectStatement) String() string {
	var out string

	// Clause WITH
	if ss.With != nil {
		out += ss.With.String() + " "
	}

	out += "SELECT "
	if ss.Distinct {
		out += "DISTINCT "
	}
	for i, sel := range ss.Select {
		if i > 0 {
			out += ", "
		}
		out += sel.String()
	}
	out += " FROM " + ss.From.String()

	for _, join := range ss.Joins {
		out += " " + join.String()
	}

	if ss.Where != nil {
		out += " WHERE " + ss.Where.String()
	}

	// Clause hiérarchique CONNECT BY
	if ss.Hierarchical != nil {
		out += ss.Hierarchical.String()
	}

	if len(ss.GroupBy) > 0 {
		out += " GROUP BY "
		for i, gb := range ss.GroupBy {
			if i > 0 {
				out += ", "
			}
			out += gb.String()
		}
	}

	if ss.Having != nil {
		out += " HAVING " + ss.Having.String()
	}

	// Définitions de fenêtres nommées
	if len(ss.WindowClauses) > 0 {
		out += " WINDOW "
		for i, wc := range ss.WindowClauses {
			if i > 0 {
				out += ", "
			}
			out += wc.Name.String() + " AS " + wc.String()
		}
	}

	if len(ss.OrderBy) > 0 {
		out += " ORDER BY "
		for i, ob := range ss.OrderBy {
			if i > 0 {
				out += ", "
			}
			out += ob.String()
		}
	}

	if ss.Limit != nil {
		out += " LIMIT " + ss.Limit.String()
	}

	if ss.Offset != nil {
		out += " OFFSET " + ss.Offset.String()
	}

	if ss.Union != nil {
		if ss.UnionAll {
			out += " UNION ALL " + ss.Union.String()
		} else {
			out += " UNION " + ss.Union.String()
		}
	}

	return out
}
func (ss *SQLSelectStatement) TokenLiteral() string { return ss.Token.Literal }
func (ss *SQLSelectStatement) statementNode()       {}
func (ss *SQLSelectStatement) expressionNode()      {}
func (ss *SQLSelectStatement) Line() int            { return ss.Token.Line }
func (ss *SQLSelectStatement) Column() int          { return ss.Token.Column }

// SQLRecursiveCTE - CTE récursif spécialisé
type SQLRecursiveCTE struct {
	Token     token.Token
	Name      *Identifier
	Columns   []*Identifier
	Anchor    *SQLSelectStatement // Partie anchor
	Recursive *SQLSelectStatement // Partie récursive
	UnionAll  bool
}

func (sr *SQLRecursiveCTE) String() string {
	out := sr.Name.String()
	if len(sr.Columns) > 0 {
		out += " ("
		for i, col := range sr.Columns {
			if i > 0 {
				out += ", "
			}
			out += col.String()
		}
		out += ")"
	}
	out += " AS ("
	out += sr.Anchor.String()
	if sr.UnionAll {
		out += " UNION ALL "
	} else {
		out += " UNION "
	}
	out += sr.Recursive.String()
	out += ")"
	return out
}

// ArrayType - Type de tableau
type ArrayType struct {
	Token       token.Token
	ElementType *TypeAnnotation
	Size        *IntegerLiteral // Taille fixe optionnelle
}

func (at *ArrayType) String() string {
	out := "array"
	if at.Size != nil {
		out += "[" + at.Size.String() + "]"
	}
	out += " of " + at.ElementType.String()
	return out
}

// ArrayLiteral - Littéral de tableau
type ArrayLiteral struct {
	Token    token.Token
	Elements []Expression
}

func (al *ArrayLiteral) TokenLiteral() string { return al.Token.Literal }
func (al *ArrayLiteral) String() string {
	var out string
	out += "["
	for i, el := range al.Elements {
		if i > 0 {
			out += ", "
		}
		out += el.String()
	}
	out += "]"
	return out
}
func (al *ArrayLiteral) expressionNode() {}
func (al *ArrayLiteral) Line() int       { return al.Token.Line }
func (al *ArrayLiteral) Column() int     { return al.Token.Column }

// IndexExpression - Accès par index
type IndexExpression struct {
	Token token.Token
	Left  Expression
	Index Expression
}

func (ie *IndexExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IndexExpression) String() string {
	return "(" + ie.Left.String() + "[" + ie.Index.String() + "])"
}
func (ie *IndexExpression) expressionNode() {}
func (ie *IndexExpression) Line() int       { return ie.Token.Line }
func (ie *IndexExpression) Column() int     { return ie.Token.Column }

// SliceExpression - Tranche de tableau
type SliceExpression struct {
	Token token.Token
	Left  Expression
	Start Expression
	End   Expression
}

func (se *SliceExpression) TokenLiteral() string { return se.Token.Literal }
func (se *SliceExpression) String() string {
	out := "(" + se.Left.String() + "["
	if se.Start != nil {
		out += se.Start.String()
	}
	out += ":"
	if se.End != nil {
		out += se.End.String()
	}
	out += "])"
	return out
}
func (se *SliceExpression) expressionNode() {}
func (se *SliceExpression) Line() int       { return se.Token.Line }
func (se *SliceExpression) Column() int     { return se.Token.Column }

// ArrayFunctionCall - Appel de fonction de tableau
type ArrayFunctionCall struct {
	Token     token.Token
	Function  *Identifier
	Array     Expression
	Arguments []Expression
}

func (af *ArrayFunctionCall) TokenLiteral() string { return af.Token.Literal }
func (af *ArrayFunctionCall) String() string {
	out := af.Function.String() + "(" + af.Array.String()
	for _, arg := range af.Arguments {
		out += ", " + arg.String()
	}
	out += ")"
	return out
}
func (af *ArrayFunctionCall) expressionNode() {}
func (af *ArrayFunctionCall) Line() int       { return af.Token.Line }
func (af *ArrayFunctionCall) Column() int     { return af.Token.Column }

// InExpression - Expression IN
type InExpression struct {
	Token token.Token
	Left  Expression
	Right Expression // Peut être un ArrayLiteral ou une expression
	Not   bool
}

func (ie *InExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *InExpression) String() string {
	out := "(" + ie.Left.String()
	if ie.Not {
		out += " NOT"
	}
	out += " IN " + ie.Right.String() + ")"
	return out
}
func (ie *InExpression) expressionNode() {}
func (ie *InExpression) Line() int       { return ie.Token.Line }
func (ie *InExpression) Column() int     { return ie.Token.Column }

// Mettre à jour TypeAnnotation pour supporter les tableaux
type TypeAnnotation struct {
	Token       token.Token
	Type        string
	ArrayType   *ArrayType // Pour les tableaux
	Constraints *TypeConstraints
}

func (ta *TypeAnnotation) String() string {
	if ta.ArrayType != nil {
		return ta.ArrayType.String()
	}
	out := ta.Type
	if ta.Constraints != nil {
		out += ta.Constraints.String()
	}
	return out
}

// SwitchStatement - Instruction switch
type SwitchStatement struct {
	Token       token.Token
	Expression  Expression
	Cases       []*SwitchCase
	DefaultCase *BlockStatement
}

func (ss *SwitchStatement) statementNode()       {}
func (ss *SwitchStatement) TokenLiteral() string { return ss.Token.Literal }
func (ss *SwitchStatement) String() string {
	out := "switch " + ss.Expression.String() + " {\n"
	for _, caseStmt := range ss.Cases {
		out += caseStmt.String() + "\n"
	}
	if ss.DefaultCase != nil {
		out += "default:\n"
		for _, stmt := range ss.DefaultCase.Statements {
			out += "  " + stmt.String() + "\n"
		}
	}
	out += "}"
	return out
}

// SwitchCase - Cas d'un switch
type SwitchCase struct {
	Token       token.Token
	Expressions []Expression
	Body        *BlockStatement
}

func (sc *SwitchCase) String() string {
	out := "case "
	for i, expr := range sc.Expressions {
		if i > 0 {
			out += ", "
		}
		out += expr.String()
	}
	out += ":\n"
	for _, stmt := range sc.Body.Statements {
		out += "  " + stmt.String() + "\n"
	}
	return out
}

// BreakStatement - Instruction break
type BreakStatement struct {
	Token token.Token
}

func (bs *BreakStatement) statementNode()       {}
func (bs *BreakStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BreakStatement) String() string {
	return "break;"
}

// FallthroughStatement - Instruction fallthrough
type FallthroughStatement struct {
	Token token.Token
}

func (fs *FallthroughStatement) statementNode()       {}
func (fs *FallthroughStatement) TokenLiteral() string { return fs.Token.Literal }
func (fs *FallthroughStatement) String() string {
	return "fallthrough;"
}

type StructLiteral struct {
	Token  token.Token
	Name   *Identifier
	Fields []StructFieldLit
}

func (sl *StructLiteral) TokenLiteral() string { return sl.Token.Literal }
func (sl *StructLiteral) String() string {
	var out string
	if sl.Name != nil {
		out += sl.Name.String()
	}
	out += "{\n"
	for i, el := range sl.Fields {
		if i > 0 {
			out += "\n "
		}
		out += el.String()
	}
	out += "}"
	return out
}
func (sl *StructLiteral) expressionNode() {}
func (sl *StructLiteral) Line() int       { return sl.Token.Line }
func (sl *StructLiteral) Column() int     { return sl.Token.Column }

// StructField - champ de structure
type StructFieldLit struct {
	Token token.Token
	Name  *Identifier
	Value Expression
}

func (sl *StructFieldLit) String() string {
	return sl.Name.String() + " : " + sl.Value.String()
}

// Identifier - identifiant
type FromIdentifier struct {
	Token   token.Token
	Value   Expression
	NewName Expression
}

func (fi *FromIdentifier) TokenLiteral() string { return fi.Token.Literal }
func (fi *FromIdentifier) String() string {
	return strings.TrimSpace(fmt.Sprintf("%s %s", fi.Value.String(), fi.NewName))
}
func (fi *FromIdentifier) expressionNode() {}
func (fi *FromIdentifier) Line() int       { return fi.Token.Line }
func (fi *FromIdentifier) Column() int     { return fi.Token.Column }

// Select arguments
type SelectArgs struct {
	Expr    Expression
	NewName *Identifier
}

func (sa *SelectArgs) TokenLiteral() string { return sa.Expr.String() }
func (sa *SelectArgs) String() string {
	out := sa.Expr.String()
	if sa.NewName != nil {
		out += " AS " + sa.NewName.String()
	}
	return out
}
func (sa *SelectArgs) expressionNode() {}
func (sa *SelectArgs) Line() int       { return sa.Expr.Line() }
func (sa *SelectArgs) Column() int     { return sa.Expr.Column() }

// NullLiteral - littéral null
type NullLiteral struct {
	Token token.Token
}

func (nl *NullLiteral) TokenLiteral() string { return nl.Token.Literal }
func (nl *NullLiteral) String() string       { return nl.Token.Literal }
func (sa *NullLiteral) expressionNode()      {}
func (sa *NullLiteral) Line() int            { return sa.Token.Line }
func (sa *NullLiteral) Column() int          { return sa.Token.Column }

type DurationLiteral struct {
	Token          token.Token
	Value          string
	ParsedDuration *Duration // Parsé lors de l'évaluation
}

func (dl *DurationLiteral) TokenLiteral() string { return dl.Token.Literal }
func (dl *DurationLiteral) String() string       { return dl.Token.Literal }
func (dl *DurationLiteral) expressionNode()      {}
func (dl *DurationLiteral) Line() int            { return dl.Token.Line }
func (dl *DurationLiteral) Column() int          { return dl.Token.Column }

// Structure Duration pour stocker la durée parsée
type Duration struct {
	Years   int64
	Months  int64
	Days    int64
	Hours   int64
	Minutes int64
	Seconds int64
	Nanos   int64
}
