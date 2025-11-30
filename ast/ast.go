package ast

import "github.com/akristianlopez/action/token"

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

// TypeAnnotation - annotation de type avec contraintes
type TypeAnnotation struct {
	Token       token.Token
	Type        string
	Constraints *TypeConstraints
}

func (ta *TypeAnnotation) String() string {
	out := ta.Type
	if ta.Constraints != nil {
		out += ta.Constraints.String()
	}
	return out
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
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string       { return i.Value }

// IntegerLiteral - littéral entier
type IntegerLiteral struct {
	Token token.Token
	Value int64
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) String() string       { return il.Token.Literal }

// FloatLiteral - littéral flottant
type FloatLiteral struct {
	Token token.Token
	Value float64
}

func (fl *FloatLiteral) expressionNode()      {}
func (fl *FloatLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FloatLiteral) String() string       { return fl.Token.Literal }

// StringLiteral - littéral chaîne
type StringLiteral struct {
	Token token.Token
	Value string
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }
func (sl *StringLiteral) String() string       { return "\"" + sl.Value + "\"" }

// BooleanLiteral - littéral booléen
type BooleanLiteral struct {
	Token token.Token
	Value bool
}

func (b *BooleanLiteral) expressionNode()      {}
func (b *BooleanLiteral) TokenLiteral() string { return b.Token.Literal }
func (b *BooleanLiteral) String() string       { return b.Token.Literal }

// DateTimeLiteral - littéral date/time
type DateTimeLiteral struct {
	Token  token.Token
	Value  string
	IsTime bool
}

func (dt *DateTimeLiteral) expressionNode()      {}
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

// SQLSelectStatement - requête SQL
type SQLSelectStatement struct {
	Token  token.Token
	Select []Expression
	From   Expression
	Joins  []*SQLJoin
	Where  Expression
}

func (ss *SQLSelectStatement) expressionNode()      {}
func (ss *SQLSelectStatement) TokenLiteral() string { return ss.Token.Literal }
func (ss *SQLSelectStatement) String() string {
	var out string
	out += "SELECT "
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
