package ast

import (
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
}

// // DeclareSection - section des déclarations
// type DeclareSection struct {
// 	Token        token.Token // token DECLARE
// 	Declarations []Statement
// 	EndToken     token.Token // token du délimiteur de fin
// }

// func (ds *DeclareSection) statementNode()       {}
// func (ds *DeclareSection) TokenLiteral() string { return ds.Token.Literal }
//
//	func (ds *DeclareSection) String() string {
//		var out string
//		out += "declare {\n"
//		for _, decl := range ds.Declarations {
//			out += "  " + decl.String() + "\n"
//		}
//		out += "}"
//		return out
//	}
//
// DeclarationBlock - Bloc de déclarations avec mot-clé
type DeclarationBlock struct {
	Token     token.Token // token DECLARATION
	Structs   []*StructStatement
	Functions []*FunctionLiteral
	Variables []*LetStatement // Variables globales
}

func (db *DeclarationBlock) statementNode()       {}
func (db *DeclarationBlock) TokenLiteral() string { return db.Token.Literal }
func (db *DeclarationBlock) String() string {
	var out string
	out += "declaration\n"
	for _, st := range db.Structs {
		out += "  " + st.String()
	}
	for _, fn := range db.Functions {
		out += "  " + fn.String()
	}
	for _, v := range db.Variables {
		out += "  " + v.String()
	}
	out += "end\n"
	return out
}

// ActionStatement - modifié pour inclure la section declare
// Programme - nœud racine
type ActionStatement struct {
	Token        token.Token       // token ACTION
	Name         *StringLiteral    //*StringLiteral
	Declarations *DeclarationBlock // Section des déclarations optionnelle
	Body         *BlockStatement
	Stop         token.Token // token STOP
	Start        token.Token // token START
}

func (as *ActionStatement) statementNode()       {}
func (as *ActionStatement) TokenLiteral() string { return as.Token.Literal }
func (as *ActionStatement) String() string {
	var out string
	out += "action " + as.Name.String() + " "
	if as.Declarations != nil {
		out += as.Declarations.String() + " "
	}
	out += as.Body.String()
	out += " stop"
	return out
}

// StructStatement - déclaration de structure
type StructStatement struct {
	Token  token.Token // token STRUCT
	Name   *Identifier
	Fields []*StructField
}

func (ss *StructStatement) statementNode()       {}
func (ss *StructStatement) TokenLiteral() string { return ss.Token.Literal }
func (ss *StructStatement) String() string {
	var out string
	out += "struct " + ss.Name.String() + " {"
	for i, field := range ss.Fields {
		if i > 0 {
			out += ", "
		}
		out += field.String()
	}
	out += "}"
	return out
}

// StructField - champ d'une structure
type StructField struct {
	Token token.Token
	Name  *Identifier
	Type  *Identifier // Pour l'instant, types simples
}

func (sf *StructField) expressionNode()      {}
func (sf *StructField) TokenLiteral() string { return sf.Token.Literal }
func (sf *StructField) String() string {
	return sf.Name.String() + ":" + sf.Type.String()
}

// StructLiteral - instance de structure
type StructLiteral struct {
	Token  token.Token // token STRUCT ou NEW
	Type   *Identifier
	Fields map[*Identifier]Expression
}

func (sl *StructLiteral) expressionNode()      {}
func (sl *StructLiteral) TokenLiteral() string { return sl.Token.Literal }
func (sl *StructLiteral) String() string {
	var out string
	pairs := []string{}
	for key, value := range sl.Fields {
		pairs = append(pairs, key.String()+":"+value.String())
	}
	out += sl.Type.String() + " {"
	out += strings.Join(pairs, ", ")
	out += "}"
	return out
}

// MemberExpression - accès aux membres d'une structure (obj.champ)
type MemberExpression struct {
	Token  token.Token // token DOT
	Object Expression
	Member *Identifier
}

func (me *MemberExpression) expressionNode()      {}
func (me *MemberExpression) TokenLiteral() string { return me.Token.Literal }
func (me *MemberExpression) String() string {
	return me.Object.String() + "." + me.Member.String()
}

// NewExpression - création d'instance avec "new"
type NewExpression struct {
	Token     token.Token // token NEW
	Type      *Identifier
	Arguments []Expression
}

func (ne *NewExpression) expressionNode()      {}
func (ne *NewExpression) TokenLiteral() string { return ne.Token.Literal }
func (ne *NewExpression) String() string {
	var out string
	args := []string{}
	for _, a := range ne.Arguments {
		args = append(args, a.String())
	}
	out += "new " + ne.Type.String() + "("
	out += strings.Join(args, ", ")
	out += ")"
	return out
}

// LetStatement - instruction d'assignation (let x = 5)
type LetStatement struct {
	Token token.Token // token LET
	Name  *Identifier
	Value Expression
}

func (ls *LetStatement) statementNode()       {}
func (ls *LetStatement) TokenLiteral() string { return ls.Token.Literal }
func (ls *LetStatement) String() string {
	var out string
	out += ls.TokenLiteral() + " "
	out += ls.Name.String()
	out += " = "
	if ls.Value != nil {
		out += ls.Value.String()
	}
	out += ";"
	return out
}

// ReturnStatement - instruction return
type ReturnStatement struct {
	Token       token.Token // token RETURN
	ReturnValue Expression
}

func (rs *ReturnStatement) statementNode()       {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }
func (rs *ReturnStatement) String() string {
	var out string
	out += rs.TokenLiteral() + " "
	if rs.ReturnValue != nil {
		out += rs.ReturnValue.String()
	}
	out += ";"
	return out
}

// ExpressionStatement - expression comme instruction
type ExpressionStatement struct {
	Token      token.Token // premier token de l'expression
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

// Identifier - identificateur (variable)
type Identifier struct {
	Token token.Token // token IDENT
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string       { return i.Value }

// IntegerLiteral - entier
type IntegerLiteral struct {
	Token token.Token
	Value int64
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) String() string       { return il.Token.Literal }

// IntegerLiteral - entier
type StringLiteral struct {
	Token token.Token
	Value string
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }
func (sl *StringLiteral) String() string       { return sl.Value }

// Boolean - booléen
type Boolean struct {
	Token token.Token
	Value bool
}

func (b *Boolean) expressionNode()      {}
func (b *Boolean) TokenLiteral() string { return b.Token.Literal }
func (b *Boolean) String() string       { return b.Token.Literal }

// PrefixExpression - expression préfixe (!true, -5)
type PrefixExpression struct {
	Token    token.Token // token préfixe (!, -)
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }
func (pe *PrefixExpression) String() string {
	return "(" + pe.Operator + pe.Right.String() + ")"
}

// InfixExpression - expression infixe (5 + 5, a * b)
type InfixExpression struct {
	Token    token.Token // token opérateur
	Left     Expression
	Operator string
	Right    Expression
}

func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *InfixExpression) String() string {
	return "(" + ie.Left.String() + " " + ie.Operator + " " + ie.Right.String() + ")"
}

// IfExpression - expression conditionnelle
type IfExpression struct {
	Token       token.Token // token IF
	Condition   Expression
	Consequence *BlockStatement
	Alternative *BlockStatement
}

func (ie *IfExpression) expressionNode()      {}
func (ie *IfExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IfExpression) String() string {
	var out string
	out += "if" + ie.Condition.String() + " " + ie.Consequence.String()
	if ie.Alternative != nil {
		out += "else " + ie.Alternative.String()
	}
	return out
}

// BlockStatement - bloc d'instructions
type BlockStatement struct {
	Token      token.Token // token {
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

// FunctionLiteral - Literal de fonction
type FunctionLiteral struct {
	Token      token.Token // token FUNCTION
	Name       *Identifier
	Parameters []*FunctionParameter
	Body       *BlockStatement
}

func (fl *FunctionLiteral) expressionNode()      {}
func (fl *FunctionLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FunctionLiteral) String() string {
	var out string
	out += "fn " + fl.Name.String() + "("
	for i, p := range fl.Parameters {
		out += p.String()
		if i < len(fl.Parameters)-1 {
			out += ", "
		}
	}
	out += ") {\n"
	out += fl.Body.String()
	out += "}\n"
	return out
}

// FunctionParameter - Paramètre de fonction
type FunctionParameter struct {
	Token token.Token // token IDENT
	Name  *Identifier
	Type  *Identifier
}

func (fp *FunctionParameter) expressionNode()      {}
func (fp *FunctionParameter) TokenLiteral() string { return fp.Token.Literal }
func (fp *FunctionParameter) String() string {
	return fp.Name.String() + " : " + fp.Type.String()
}

// CallExpression - appel de fonction
type CallExpression struct {
	Token     token.Token // token '('
	Function  Expression
	Arguments []Expression
}

func (ce *CallExpression) expressionNode()      {}
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Literal }
func (ce *CallExpression) String() string {
	var out string
	args := []string{}
	for _, a := range ce.Arguments {
		args = append(args, a.String())
	}
	out += ce.Function.String()
	out += "("
	out += strings.Join(args, ", ")
	out += ")"
	return out
}

// FunctionStatement - déclaration de fonction
type FunctionStatement struct {
	Token    token.Token // token FUNCTION
	Name     *Identifier
	Function *FunctionLiteral
}

func (fs *FunctionStatement) statementNode()       {}
func (fs *FunctionStatement) TokenLiteral() string { return fs.Token.Literal }
func (fs *FunctionStatement) String() string {
	var out string
	out += "fn " + fs.Name.String()
	out += fs.Function.String()
	return out
}
