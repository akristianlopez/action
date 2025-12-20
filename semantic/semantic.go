package semantic

import (
	"fmt"
	"strconv"
	"strings"

	// "go/ast"

	"github.com/akristianlopez/action/ast"
	// "github.com/akristianlopez/action/object"
	// "github.com/akristianlopez/action/token"
	// "strings"
)

type SymbolType string

const (
	VariableSymbol  SymbolType = "VARIABLE"
	FunctionSymbol  SymbolType = "FUNCTION"
	StructSymbol    SymbolType = "STRUCT"
	DbObjectSymbol  SymbolType = "TABLE"
	TypeSymbol      SymbolType = "TYPE"
	ArraySymbol     SymbolType = "ARRAY"
	ParameterSymbol SymbolType = "PARAMETER"
)

type Symbol struct {
	Name     string
	Type     SymbolType
	DataType *TypeInfo
	Scope    *Scope
	Node     ast.Node
	NoOrder  int
	Index    int
}

type Scope struct {
	Parent   *Scope
	Symbols  map[string]*Symbol
	Children []*Scope
}

type TypeInfo struct {
	Name        string
	IsArray     bool
	ArraySize   int64
	ElementType *TypeInfo
	Fields      map[string]*TypeInfo // Pour les structures
}

func (ti *TypeInfo) String() string {
	if ti.IsArray {
		return fmt.Sprintf("Array of %s", ti.ElementType.String())
	}
	return ti.Name
}

type SemanticAnalyzer struct {
	CurrentScope *Scope
	GlobalScope  *Scope
	Errors       []string
	Warnings     []string
	TypeTable    map[string]*TypeInfo
	TypeSql      map[string]*TypeInfo
	inType       int
}

// var tokenList []string

func NewSemanticAnalyzer() *SemanticAnalyzer {

	globalScope := &Scope{
		Symbols: make(map[string]*Symbol),
	}

	analyzer := &SemanticAnalyzer{
		CurrentScope: globalScope,
		GlobalScope:  globalScope,
		Errors:       []string{},
		Warnings:     []string{},
		TypeTable:    make(map[string]*TypeInfo),
		TypeSql:      make(map[string]*TypeInfo),
		inType:       1,
	}

	// Enregistrement des functions standards
	analyzer.registerBuiltinFunctions()

	// Enregistrer les types de base
	analyzer.registerBuiltinTypes()

	return analyzer
}

func (sa *SemanticAnalyzer) registerBuiltinFunctions() {
	oldScope := sa.CurrentScope

	oldScope.Children = make([]*Scope, 0)
	funScope := &Scope{
		Parent:  oldScope,
		Symbols: make(map[string]*Symbol),
	}
	oldScope.Children = append(oldScope.Children, funScope)
	sa.CurrentScope = funScope
	sa.registerSymbol("val", ParameterSymbol, &TypeInfo{Name: "any"}, &ast.Identifier{Value: "val"}, -1, 0)
	sa.CurrentScope = oldScope
	sa.registerSymbol("tostring", FunctionSymbol, &TypeInfo{Name: "string"}, &ast.Identifier{Value: "tostring"}, 0)

	funScope = &Scope{
		Parent:  oldScope,
		Symbols: make(map[string]*Symbol),
	}
	oldScope.Children = append(oldScope.Children, funScope)
	sa.CurrentScope = funScope
	sa.registerSymbol("val", ParameterSymbol, &TypeInfo{Name: "any"}, &ast.Identifier{Value: "val"}, -1, 0)
	sa.CurrentScope = oldScope
	sa.registerSymbol("len", FunctionSymbol, &TypeInfo{Name: "integer"}, &ast.Identifier{Value: "val"}, 1)

	funScope = &Scope{
		Parent:  oldScope,
		Symbols: make(map[string]*Symbol),
	}
	oldScope.Children = append(oldScope.Children, funScope)
	sa.CurrentScope = funScope
	sa.registerSymbol("arr", ParameterSymbol, &TypeInfo{Name: "array", IsArray: true, ElementType: &TypeInfo{Name: "any"}}, &ast.Identifier{Value: "arr"}, -1, 0)
	sa.registerSymbol("element", ParameterSymbol, &TypeInfo{Name: "any"}, &ast.Identifier{Value: "element"}, -1, 1)
	sa.CurrentScope = oldScope
	sa.registerSymbol("append", FunctionSymbol, &TypeInfo{Name: "array", IsArray: true, ElementType: &TypeInfo{Name: "any"}}, &ast.Identifier{Value: "append"}, 2)

	funScope = &Scope{
		Parent:  oldScope,
		Symbols: make(map[string]*Symbol),
	}
	oldScope.Children = append(oldScope.Children, funScope)
	sa.CurrentScope = funScope
	sa.registerSymbol("arr", ParameterSymbol, &TypeInfo{Name: "array", IsArray: true, ElementType: &TypeInfo{Name: "any"}}, &ast.Identifier{Value: "arr"}, -1, 0)
	sa.registerSymbol("element", ParameterSymbol, &TypeInfo{Name: "any"}, &ast.Identifier{Value: "element"}, -1, 1)
	sa.CurrentScope = oldScope
	sa.registerSymbol("indexOf", FunctionSymbol, &TypeInfo{Name: "integer"}, &ast.Identifier{Value: "indexOf"}, 3)

	funScope = &Scope{
		Parent:  oldScope,
		Symbols: make(map[string]*Symbol),
	}
	oldScope.Children = append(oldScope.Children, funScope)
	sa.CurrentScope = funScope
	sa.registerSymbol("var", ParameterSymbol, &TypeInfo{Name: "any"}, &ast.Identifier{Value: "var"}, -1, 0)
	sa.CurrentScope = oldScope
	sa.registerSymbol("typeOf", FunctionSymbol, &TypeInfo{Name: "string"}, &ast.Identifier{Value: "typeOf"}, 4)
}

func (sa *SemanticAnalyzer) registerBuiltinTypes() {
	// Types primitifs
	sa.TypeTable["integer"] = &TypeInfo{Name: "integer"}
	sa.TypeTable["float"] = &TypeInfo{Name: "float"}
	sa.TypeTable["string"] = &TypeInfo{Name: "string"}
	sa.TypeTable["boolean"] = &TypeInfo{Name: "boolean"}
	sa.TypeTable["time"] = &TypeInfo{Name: "time"}
	sa.TypeTable["date"] = &TypeInfo{Name: "date"}
	sa.TypeTable["any"] = &TypeInfo{Name: "any"} // Type générique
	sa.TypeTable["duration"] = &TypeInfo{Name: "duration"}

	sa.TypeSql["integer"] = &TypeInfo{Name: "number"}
	sa.TypeSql["smallint"] = &TypeInfo{Name: "smallint"}
	sa.TypeSql["number"] = &TypeInfo{Name: "number"}
	sa.TypeSql["varchar"] = &TypeInfo{Name: "varchar"}
	sa.TypeSql["char"] = &TypeInfo{Name: "char"}
	sa.TypeSql["text"] = &TypeInfo{Name: "text"}
	sa.TypeSql["json"] = &TypeInfo{Name: "json"}

	sa.TypeSql["numeric"] = &TypeInfo{Name: "numeric"}
	sa.TypeSql["decimal"] = &TypeInfo{Name: "decimal"}
	sa.TypeSql["date"] = &TypeInfo{Name: "date"}
	sa.TypeSql["time"] = &TypeInfo{Name: "time"}
	sa.TypeSql["datetime"] = &TypeInfo{Name: "datetime"}
	sa.TypeSql["timestamp"] = &TypeInfo{Name: "timestamp"}
	sa.TypeSql["float"] = &TypeInfo{Name: "float"}
	sa.TypeSql["real"] = &TypeInfo{Name: "real"}
	sa.TypeSql["any"] = &TypeInfo{Name: "any"}
}

func (sa *SemanticAnalyzer) Analyze(program *ast.Program) []string {
	sa.visitProgram(program)
	return sa.Errors
}

func (sa *SemanticAnalyzer) visitProgram(node *ast.Program) {
	// Vérifier la structure du programme
	if node.ActionName == "" {
		sa.addError("Then action must start by 'action <nom>'")
	}

	// Visiter toutes les déclarations
	for _, stmt := range node.Statements {
		sa.visitStatement(stmt, &TypeInfo{Name: "any"})
	}
}

func (sa *SemanticAnalyzer) visitStatement(stmt ast.Statement, t *TypeInfo) {
	switch s := stmt.(type) {
	case *ast.LetStatement:
		sa.visitLetStatement(s)
	case *ast.LetStatements:
		sa.visitLetStatements(s)
	case *ast.FunctionStatement:
		sa.visitFunctionStatement(s)
	case *ast.StructStatement:
		sa.visitStructStatement(s)
	case *ast.IfStatement:
		sa.visitIfStatement(s, t)
	case *ast.WhileStatement:
		sa.visitWhileStatement(s, t)
	case *ast.ForEachStatement:
		sa.visitForEachStatement(s, t)
	case *ast.ForStatement:
		sa.visitForStatement(s, t)
	case *ast.SwitchStatement:
		sa.visitSwitchStatement(s, t)
	case *ast.ReturnStatement:
		sa.visitReturnStatement(s, t)
	case *ast.BlockStatement:
		sa.visitBlockStatement(s, t)
	case *ast.ExpressionStatement:
		sa.visitExpressionStatement(s)
	case *ast.SQLCreateObjectStatement:
		sa.visitSQLCreateObjectStatement(s)
	case *ast.SQLInsertStatement:
		sa.visitSQLInsertStatement(s)
	case *ast.SQLUpdateStatement:
		sa.visitSQLUpdateStatement(s)
	case *ast.SQLDeleteStatement:
		sa.visitSQLDeleteStatement(s)
	case *ast.SQLSelectStatement:
		sa.visitSQLSelectStatement(s)
	}
}

func (sa *SemanticAnalyzer) visitSQLDeleteStatement(s *ast.SQLDeleteStatement) {
	if s.From == nil {
		sa.addError("Define the right object where datas should be deleted eventually. line:%d, column:%d",
			s.Line(), s.Column())
		return
	}
	if s.Where == nil {
		sa.addError("The condition in the clause <where> is needed. line:%d, column:%d",
			s.Line(), s.Column())
		return
	}
	tokenList := make([]string, 0)
	tokenList = append(tokenList, lower(s.From.Value))
	oldScope := sa.CurrentScope
	scope := &Scope{
		Parent:  sa.CurrentScope,
		Symbols: make(map[string]*Symbol),
	}
	sa.CurrentScope = scope
	sa.registerTempoSymbols(tokenList)

	condType := sa.visitExpression(s.Where)
	if condType.Name != "boolean" {
		sa.addError("The condition of a for loop must be boolean. line:%d column:%d",
			s.Line(), s.Column())
		sa.CurrentScope = oldScope
		return
	}
	// tokenList := make([]string, 0)
	// tokenList = append(tokenList, strings.ToLower(s.From.Value))
	//Verify that each time, we have a.b, a exists in the list
	sa.visitSQLExpressionWithDotToken(tokenList, s.Where)
	//Check left operand, right operand and operator
	sa.CurrentScope = oldScope
}

func (sa *SemanticAnalyzer) visitSQLUpdateStatement(s *ast.SQLUpdateStatement) {
	if s.ObjectName == nil {
		sa.addError("Define the right name of the object. line:%d, column:%d",
			s.Line(), s.Column())
		return
	}
	if s.Set == nil {
		sa.addError("The condition in the clause <where> is needed. line:%d, column:%d",
			s.Line(), s.Column())
		return
	}
	tokenList := make([]string, 0)
	tokenList = append(tokenList, lower(s.ObjectName.Value))
	oldScope := sa.CurrentScope
	scope := &Scope{
		Parent:  sa.CurrentScope,
		Symbols: make(map[string]*Symbol),
	}
	sa.CurrentScope = scope
	sa.registerTempoSymbols(tokenList)
	for _, v := range s.Set {
		if v.Column == nil {
			sa.addError("Define the right name of the column. line:%d, column:%d",
				s.Line(), s.Column())
		}
		info := sa.visitExpression(v.Value)
		if _, exists := sa.TypeSql[lower(info.Name)]; !exists {
			sa.addError("This column '%s[%s]' is not defined. Maybe, it's a field of %s. line:%d, column:%d",
				v.Column, info.Name, s.ObjectName.Value, s.Line(), s.Column())
		}
	}
	if s.Where != nil {
		condType := sa.visitExpression(s.Where)
		if condType.Name != "boolean" {
			sa.addError("The condition of a for loop must be boolean. line:%d column:%d",
				s.Line(), s.Column())
			sa.CurrentScope = oldScope
			return
		}
		// tokenList := make([]string, 0)
		// tokenList = append(tokenList, strings.ToLower(s.ObjectName.Value))
		//Verify that each time, we have a.b, a exists in the list
		sa.visitSQLExpressionWithDotToken(tokenList, s.Where)
		//Check left operand, right operand and operator
	}
	sa.CurrentScope = oldScope
}

func (sa *SemanticAnalyzer) visitSQLInsertStatement(s *ast.SQLInsertStatement) {
	if s.ObjectName == nil {
		sa.addError("The name of the object is missing. line:%d, column:%d", s.Line(), s.Column())
		return
	}
	if s.Select == nil {
		for _, v := range s.Values {
			if len(s.Columns) > len(v.Values) {
				sa.addError("Too few values. line:%d, column:%d", v.Token.Line, v.Token.Column)
				return
			}
			if len(s.Columns) < len(v.Values) {
				sa.addError("Too much values. line:%d, column:%d", v.Token.Line, v.Token.Column)
				return
			}
			for _, e := range v.Values {
				t := sa.visitExpression(e)
				if _, exists := sa.TypeSql[lower(t.Name)]; !exists {
					if lower(t.Name) == "string" {
						continue
					}
					sa.addError("'%s' invalid expression. line:%d, column:%d", e.String(),
						e.Line(), e.Column())
				}
			}
		}
		return
	}
	if len(s.Values) > 0 {
		sa.addError("Bad insert statement. line:%d, column:%d", s.Line(), s.Column())
		return
	}
	if len(s.Columns) > 0 {
		if len(s.Select.Select) > len(s.Columns) {
			sa.addError("Too much values. line:%d, column:%d", s.Line(), s.Column())
			return
		}
		if len(s.Select.Select) < len(s.Columns) {
			sa.addError("Too few values. line:%d, column:%d", s.Line(), s.Column())
			return
		}
	}
	sa.visitSQLSelectStatement(s.Select)
}
func (sa *SemanticAnalyzer) visitSQLTypeConstraint(v *ast.SQLDataType) {
	//Check the type and it's constraint
	// if v == nil {
	// 	sa.addError("The column '%s' must have datatype. line:%d, column:%d",
	// 		v.Name.Value, v.Token.Line, v.Token.Column)
	// 	return
	// }
	if _, exists := sa.TypeSql[strings.ToLower(v.Name)]; !exists {
		sa.addError("'%s' invalid expression. line:%d, column:%d", v.Name,
			v.Token.Line, v.Token.Column)
		return
	}
	if v.Length != nil && v.Length.Value > 0 {
		switch strings.ToLower(v.Name) {
		case "varchar", "char":
			break
		default:
			sa.addError("'%d' invalid type '%s' constraint. line:%d, column:%d",
				v.Length.Value, v.Name, v.Token.Line, v.Token.Column)
		}
	}
	if v.Precision != nil && v.Precision.Value > 0 {
		switch strings.ToLower(v.Name) {
		case "number", "numeric", "decimal":
			break
		default:
			sa.addError("'%d' invalid type '%s' constraint. line:%d, column:%d",
				v.Length.Value, v.Name, v.Token.Line, v.Token.Column)
		}
	}
	if v.Scale != nil && v.Scale.Value > 0 {
		switch strings.ToLower(v.Name) {
		case "number", "numeric", "decimal":
			break
		default:
			sa.addError("'%d' invalid type '%s' constraint. line:%d, column:%d",
				v.Length.Value, v.Name, v.Token.Line, v.Token.Column)
		}
	}
}
func (sa *SemanticAnalyzer) visitSQLColumnConstraints(names []string, v *ast.SQLConstraint) {
	// Token      token.Token
	// Name       *Identifier
	// Type       string // PRIMARY KEY, FOREIGN KEY, etc.
	// Columns    []*Identifier
	// References *SQLReference
	// Check      Expression
	if v == nil {
		return
	}
	if v.Name == nil {
		sa.addError("Define the name of the column. line:%d, column:%d", v.Token.Line, v.Token.Column)
		return
	}
	for _, n := range v.Columns {
		if !contains(names, strings.ToLower(n.Value)) {
			sa.addError("This '%s' does not exist. line:%d, column:%d", n.Value, n.Line(), n.Column())
			continue
		}
	}
	if v.References != nil && len(v.References.Columns) > 0 {
		if len(v.References.Columns) > len(v.Columns) {
			sa.addError("Too much columns. line:%d, column:%d", v.References.Token.Line, v.References.Token.Column)
		}
		if len(v.References.Columns) < len(v.Columns) {
			sa.addError("Too few columns. line:%d, column:%d", v.References.Token.Line, v.References.Token.Column)
		}
	}
	t := sa.visitExpression(v.Check)
	if _, exists := sa.TypeSql[strings.ToLower(t.Name)]; !exists {
		sa.addError("'%s' invalid expression. line:%d, column:%d",
			v.Check.String(), v.Check.Line(), v.Check.Column())
	}
}

func (sa *SemanticAnalyzer) visitSQLCreateObjectStatement(s *ast.SQLCreateObjectStatement) {
	if s.ObjectName == nil {
		sa.addError("The name of the object is missing. line:%d, column:%d", s.Token.Line, s.Token.Column)
		return
	}
	if len(s.Columns) == 0 {
		sa.addError("Define at least one column. line:%d, column:%d", s.Token.Line, s.Token.Column)
		return
	}
	//Browsing columns
	names := make([]string, 0)
	for _, v := range s.Columns {
		if contains(names, strings.ToLower(v.Name.Value)) {
			sa.addError("This column '%s' is already existed. line:%d, column:%d",
				v.Name.Value, v.Token.Line, v.Token.Column)
			continue
		}
		names = append(names, strings.ToLower(v.Name.Value))
		sa.visitSQLTypeConstraint(v.DataType)
	}
	//Browsing Constraints
	constNames := make([]string, 0)
	constType := make([]string, 0)
	for _, e := range s.Constraints {
		sa.visitSQLColumnConstraints(names, e)
		if contains(constNames, strings.ToLower(e.Name.Value)) {
			sa.addError("This constraint '%s' already exist. line:%d, column:%d", e.Name.Value, e.Name.Line(), e.Name.Column())
			continue
		}
		constNames = append(constNames, strings.ToLower(e.Name.Value))
		flag := contains(constType, strings.ToLower(e.Type))
		if flag && strings.ToLower(e.Type) == "primary key" {
			sa.addError("Primary key already exist. line:%d, column:%d", e.Name.Line(), e.Name.Column())
			continue
		}
		if !flag {
			constType = append(constType, strings.ToLower(e.Type))
		}
	}
}

func (sa *SemanticAnalyzer) canReceivedValue(s ast.Expression) *TypeInfo {
	switch exp := s.(type) {
	case *ast.Identifier:
		return sa.visitIdentifier(exp)
	case *ast.IndexExpression:
		return sa.visitIndexExpression(exp)
	case *ast.TypeMember:
		return sa.visitTypeMember(exp)
	}
	return &TypeInfo{Name: "void"}
}
func (sa *SemanticAnalyzer) visitExpressionStatement(s *ast.ExpressionStatement) {
	if s == nil {
		return
	}
	switch expr := s.Expression.(type) {
	case *ast.InfixExpression:
		switch expr.Operator {
		case "=": //assignment
			l := sa.canReceivedValue(expr.Left)
			ti := sa.visitExpression(expr.Right)
			if !sa.areSameType(l, ti) {
				sa.addError("Type of '%s' does not match the type of '%s'. Line:%d, column:%d",
					expr.Left.String(), expr.Right.String(), expr.Line(), expr.Column())
			}
			/*
				case "[": //Array's element
					sa.addError("Invalid expression. Line:%d, column:%d", expr.Line(), expr.Column())
				case ".": //Object's member
					sa.addError("Invalid expression. Line:%d, column:%d", expr.Line(), expr.Column())
			*/
		default: //Unkown
			sa.addError("Invalid expression. Line:%d, column:%d", expr.Line(), expr.Column())
		}
	case *ast.ArrayFunctionCall:
		sa.visitArrayFunctionCall(expr)
	default:
		sa.addError("Invalid expression '%s'. Line:%d, column:%d", expr.String(), expr.Line(), expr.Column())
	}
}

func lIsInFrom(name string, sj []*ast.SQLJoin) bool {
	res := false
	if len(sj) > 0 {
		for _, s := range sj {
			n := s.Table.(*ast.FromIdentifier)
			if n.NewName != nil {
				nn := n.NewName.(*ast.Identifier)
				res = strings.EqualFold(nn.Value, name)
				if res {
					break
				}
			}
			switch n.Value.(type) {
			case *ast.Identifier:
				nn := n.Value.(*ast.Identifier)
				res = strings.EqualFold(nn.Value, name)
				if res {
					break
				}
			}
		}
	}
	return res
}

func contains(slice []string, element string) bool {
	if len(slice) == 0 {
		return false
	}
	for _, v := range slice {
		if v == element {
			return true
		}
	}
	return false
}

func (sa *SemanticAnalyzer) visitObjectInFromClause(se ast.Expression) *string {
	if se == nil {
		sa.addError("From clause must have at least one object. line:%d, column:%d", 0, 0)
		return nil
	}
	res := ""
	switch s := se.(type) {
	case *ast.FromIdentifier:
		switch v := s.Value.(type) {
		case *ast.Identifier:
			if s.NewName == nil {
				res = strings.ToLower(v.Value)
				return &res
			}
			switch s.NewName.(type) {
			case *ast.Identifier:
				res = strings.ToLower(v.Value)
				return &res
			default:
				sa.addError("'%s' invalid statement here. line:%d, column:%d", s.NewName.String(), s.NewName.Line(), s.NewName.Column())
				return nil
			}
		case *ast.SQLSelectStatement:
			if s.NewName == nil {
				sa.addError("New name expected. line:%d, column:%d", s.Value.Line(), s.Value.Column())
				break
			}
			switch nn := s.NewName.(type) {
			case *ast.Identifier:
				res = strings.ToLower(nn.Value)
			default:
				sa.addError("'%s' invalid expression. New name expected. line:%d, column:%d",
					nn.String(), nn.Line(), nn.Column())
				return nil
			}
		default:
			sa.addError("'%s' invalid statement here. line:%d, column:%d", s.Value.String(), s.Value.Line(), s.Value.Column())
			return nil
		}
	default:
		sa.addError("Unknown expression '%s' here. line:%d, column:%d", s.String(), s.Line(), s.Column())
		return nil
	}
	return &res
}

func (sa *SemanticAnalyzer) visitSQLExpressionWithDotToken(tab []string, expr ast.Expression) {
	switch expr.(type) {
	case *ast.TypeMember:
		tm := expr.(*ast.TypeMember)
		switch left := tm.Left.(type) {
		case *ast.Identifier:
			if !contains(tab, strings.ToLower(left.Value)) {
				sa.addError("'%s' is not an object. line:%d, column:%d",
					left.Value, left.Token.Line, left.Token.Column)

			}
		default:
			sa.addError("'%s' is not an object. line:%d, column:%d",
				left.String(), left.Line(), left.Column())
		}
	case *ast.InfixExpression:
		ie := expr.(*ast.InfixExpression)
		switch strings.ToLower(ie.Operator) {
		case "and", "or", "+", "-", ">", ">=",
			"<", "<=", "*", "/", "!=":
			sa.visitSQLExpressionWithDotToken(tab, ie.Left)
			sa.visitSQLExpressionWithDotToken(tab, ie.Right)
		}
	default:
		return
	}
}

func (sa *SemanticAnalyzer) visitSQLSelectStatement(ss *ast.SQLSelectStatement) {
	//check for select argumens
	if ss.Select == nil {
		sa.addError("select must have at least one field. line:%d, column:%d", ss.Line(), ss.Column())
		return
	}
	if ss.From == nil {
		sa.addError("select must have at least one object in the clause from. line:%d, column:%d", ss.Line(), ss.Column())
		return
	}
	argList := make([]string, 0)
	for _, f := range ss.Select {
		field := f.(*ast.SelectArgs)
		switch s := field.Expr.(type) {
		case *ast.Identifier:
			if !contains(argList, lower(s.Value)) {
				argList = append(argList, lower(s.Value))
			}
			continue
		case *ast.StringLiteral, *ast.DurationLiteral, *ast.BooleanLiteral,
			*ast.FloatLiteral, *ast.IntegerLiteral, *ast.DateTimeLiteral:
			str := field.Expr.String()
			if field.NewName != nil {
				str = field.NewName.Value
				// sa.addError("Then '%s' must have a new name. line:%d, column:%d",
				// 	s.String(), s.Line(), s.Column())
				// break
			}
			if !contains(argList, lower(str)) {
				argList = append(argList, lower(str))
			}
		case *ast.InfixExpression:
			t := field.Expr.(*ast.InfixExpression)
			if t.Operator != "." { //We should add RARR when we would need to take into account the sub-object
				sa.addError("'%s' invalid operation in the select clause. line:%d, column:%d",
					t.Operator, t.Right.Line(), t.Right.Column())
				continue
			}
			switch t.Left.(type) {
			case *ast.Identifier:
				n := t.Left.(*ast.Identifier)
				if !lIsInFrom(n.Value, ss.Joins) {
					sa.addError("'%s' is not an object. line:%d, column:%d",
						n.Value, n.Token.Line, n.Token.Column)

				}
				switch t.Right.(type) {
				case *ast.Identifier, *ast.StringLiteral:
					continue
				default:
					sa.addError("'%s' can not be a new name. line:%d, column:%d",
						n.Value, n.Token.Line, n.Token.Column)
				}
			case *ast.ArrayFunctionCall:
				n := t.Left.(*ast.ArrayFunctionCall)
				if sa.lookupSymbol(n.Function.Value) != nil {
					sa.addError("'%s' can not be used in the select clause. line:%d, column:%d",
						t.Right.String(), t.Right.Line(), t.Right.Column())
				}
				//Check the function argument format but this will be done after
				if n.Array == nil && len(n.Arguments) == 0 {
					sa.addError("Function '%s' must have at least one argument. line:%d, column:%d",
						n.Function.String(), s.Line(), s.Column())
				}
				switch e := t.Right.(type) {
				case *ast.Identifier, *ast.StringLiteral:
					if field.NewName != nil {
						if !contains(argList, strings.ToLower(field.NewName.Value)) {
							argList = append(argList, strings.ToLower(field.NewName.Value))
						}
						continue
					}
					if !contains(argList, strings.ToLower(e.String())) {
						argList = append(argList, strings.ToLower(e.String()))
					}
				default:
					sa.addError("'%s' can not be a new name. line:%d, column:%d",
						t.Right.String(), t.Right.Line(), t.Right.Column())
				}
			}
		case *ast.ArrayFunctionCall:
			if sa.lookupSymbol(s.Function.Value) != nil {
				sa.addError("'%s' can not be used in the select clause. line:%d, column:%d",
					s.Function.String(), s.Line(), s.Column())
			}
			//Check the function argument format but this will be done after
			if s.Array == nil && len(s.Arguments) == 0 {
				sa.addError("Function '%s' must have at least one argument. line:%d, column:%d",
					s.Function.String(), s.Line(), s.Column())
			}
		default:
			sa.addError("'%s' invalid . line:%d, column:%d",
				s.String(), s.Line(), s.Column())
		}

	}

	//Check the clause From expression
	tokenList := make([]string, 0)
	cf := sa.visitObjectInFromClause(ss.From)
	if cf != nil {
		tokenList = append(tokenList, *cf)
	}

	//Look on the join clauses
	if ss.Joins != nil {
		for _, fm := range ss.Joins {
			cf = sa.visitObjectInFromClause(fm.Table)
			if !contains(tokenList, *cf) {
				tokenList = append(tokenList, *cf)
				continue
			}
			sa.addError("'%s' already exists. Line:%d, column:%d", *cf, fm.Table.Line(), fm.Table.Column())
			//check the clause ON globally
			condType := sa.visitExpression(fm.On)
			if condType.Name != "boolean" {
				sa.addError("The condition of a for loop must be boolean. line:%d column:%d",
					ss.Line(), ss.Column())
			}
			//Verify that each time, we have a.b, a exists in the list
			sa.visitSQLExpressionWithDotToken(tokenList, fm.On)
		}
	}
	//Check the clause where
	if ss.Where != nil {
		oldscope := sa.CurrentScope
		scope := Scope{
			Parent:  sa.CurrentScope,
			Symbols: make(map[string]*Symbol),
		}
		sa.CurrentScope = &scope
		sa.registerTempoSymbols(tokenList)
		condType := sa.visitExpression(ss.Where)
		sa.CurrentScope = oldscope
		if condType.Name != "boolean" {
			sa.addError("The condition of a for loop must be boolean. line:%d column:%d",
				ss.Line(), ss.Column())
		}
		//Verify that each time, we have a.b, a exists in the list
		sa.visitSQLExpressionWithDotToken(tokenList, ss.Where)
		//Check left operand, right operand and operator
	}
	//Check the clause Having
	if ss.Having != nil {
		condType := sa.visitExpression(ss.Having)
		if condType.Name != "boolean" {
			sa.addError("The condition of a for loop must be boolean. line:%d column:%d",
				ss.Line(), ss.Column())
		}
	}
	//Verify that each time, we have a.b, a exists in the list
	if ss.Where != nil {
		sa.visitSQLExpressionWithDotToken(tokenList, ss.Where)
	}
	//Check the clause Group by
	if ss.GroupBy != nil {
		for _, v := range ss.GroupBy {
			switch t := v.(type) {
			case *ast.InfixExpression:
				if t.Operator == "." {
					sa.visitSQLExpressionWithDotToken(tokenList, t)
					continue
				}
				sa.addError("Invalid operation '%s'. line:%d, column:%d", t.Operator, t.Line(), t.Column())
			case *ast.IntegerLiteral:
				//verify if the value of the literal is between 0 and length of the select arguments list
				if t.Value <= 0 || t.Value >= int64(len(argList)) {
					sa.addError("Index '%d' out of box. line:%d column:%d", t.Value,
						t.Line(), t.Column())
				}
			case *ast.StringLiteral:
				//Verify that this literal exists into the select rguments list
				if !contains(argList, strings.ToLower(t.Value)) {
					sa.addError("Invalid express '%s'. line:%d column:%d", t.String(),
						t.Line(), t.Column())
				}
			case *ast.ArrayFunctionCall:
				//very that this function was call in the select clause
				if !contains(argList, strings.ToLower(t.String())) {
					sa.addError("Invalid express '%s'. line:%d column:%d", t.String(),
						t.Line(), t.Column())
				}
			default:
				sa.addError("Invalid expression '%s'. line:%d, column:%d", t.String(), t.Line(), t.Column())
			}
		}
	}
	//Check the claude Order by
	if ss.OrderBy != nil {
		for _, v := range ss.OrderBy {
			switch t := v.Expression.(type) {
			case *ast.InfixExpression:
				if t.Operator == "." {
					sa.visitSQLExpressionWithDotToken(tokenList, t)
					if !contains(argList, strings.ToLower(t.String())) {
						sa.addError("Field '%s'does not exist. line:%d, column:%d", t.String(), t.Line(), t.Column())
					}
				}
				sa.addError("Invalid operation '%s'. line:%d, column:%d", t.Operator, t.Line(), t.Column())
			case *ast.Identifier, *ast.StringLiteral:
				if !contains(argList, strings.ToLower(t.String())) {
					sa.addError("Field '%s'does not exist. line:%d, column:%d", t.String(), t.Line(), t.Column())
				}
				sa.addError("Invalid operation '%s'. line:%d, column:%d", t.String(), t.Line(), t.Column())
			default:
				sa.addError("Invalid expression '%s'. line:%d, column:%d", t.String(), t.Line(), t.Column())
			}
		}
	}
	if ss.Union != nil {
		sa.visitSQLSelectStatement(ss.Union)
	}
}

func (sa *SemanticAnalyzer) visitLetStatements(nodes *ast.LetStatements) {
	// Vérifier si la variable est déjà déclarée
	var varType *TypeInfo
	for _, node := range *nodes {
		varType = nil
		if sa.lookupSymbol(node.Name.Value) != nil {
			sa.addError("Variable '%s' already declared. line:%d column:%d",
				node.Name.Value, node.Name.Token.Line, node.Name.Token.Column)
			return
		}
		if node.Type != nil {
			varType = sa.resolveTypeAnnotation(node.Type)
		}
		// Si une valeur est fournie, vérifier la compatibilité des types
		if node.Value != nil {
			valueType := sa.visitExpression(node.Value)
			if varType != nil && !sa.areTypesCompatible(varType, valueType) {
				sa.addError("Type mismatch for the variable '%s': expected %s, got %s. line:%d column:%d",
					node.Name.Value, varType.String(), valueType.String(), node.Token.Line, node.Token.Column)
			}
			// Si le type n'est pas spécifié, l'inférer
			if varType == nil {
				varType = valueType
			}
		}
		// Enregistrer la variable
		sa.registerSymbol(node.Name.Value, VariableSymbol, varType, &node)
	}
}

func (sa *SemanticAnalyzer) visitLetStatement(node *ast.LetStatement) {
	// Vérifier si la variable est déjà déclarée
	var varType *TypeInfo
	if sa.lookupSymbol(node.Name.Value) != nil {
		sa.addError("Variable '%s' already declared. line:%d column:%d",
			node.Name.Value, node.Name.Token.Line, node.Name.Token.Column)
		return
	}
	if node.Type != nil {
		varType = sa.resolveTypeAnnotation(node.Type)
	}
	// Si une valeur est fournie, vérifier la compatibilité des types
	if node.Value != nil {
		valueType := sa.visitExpression(node.Value)

		if varType != nil && !sa.areTypesCompatible(varType, valueType) {
			sa.addError("Type mismatch for the variable '%s': expected %s, got %s. line:%d column:%d",
				node.Name.Value, varType.Name, valueType.Name, node.Token.Line, node.Token.Column)
		}
		// Si le type n'est pas spécifié, l'inférer
		if varType == nil {
			varType = valueType
		}
	}
	// Enregistrer la variable
	sa.registerSymbol(node.Name.Value, VariableSymbol, varType, node)
}

func (sa *SemanticAnalyzer) visitFunctionStatement(node *ast.FunctionStatement) {
	// Vérifier si la fonction est déjà déclarée
	if sa.lookupSymbol(node.Name.Value) != nil {
		sa.addError("Function '%s' already declared", node.Name.Value)
		return
	}

	// Créer un nouveau scope pour la fonction
	funcScope := &Scope{
		Parent:  sa.CurrentScope,
		Symbols: make(map[string]*Symbol),
	}
	sa.CurrentScope.Children = append(sa.CurrentScope.Children, funcScope)
	pos := len(sa.CurrentScope.Children) - 1
	// Enregistrer les paramètres
	oldScope := sa.CurrentScope
	sa.CurrentScope = funcScope

	for k, param := range node.Parameters {
		paramType := sa.resolveTypeAnnotation(param.Type)
		sa.registerSymbol(param.Name.Value, ParameterSymbol, paramType, param, -1, k)
	}

	// Vérifier le type de retour
	var returnType *TypeInfo
	if node.ReturnType != nil {
		returnType = sa.resolveTypeAnnotation(node.ReturnType)
	} else {
		returnType = &TypeInfo{Name: "void"}
	}

	// Analyser le corps de la fonction
	sa.visitBlockStatement(node.Body, returnType)

	// Restaurer le scope
	sa.CurrentScope = oldScope

	// Enregistrer la fonction
	sa.registerSymbol(node.Name.Value, FunctionSymbol, returnType, node, pos)
}

func (sa *SemanticAnalyzer) visitStructStatement(node *ast.StructStatement) {
	// Vérifier si la structure est déjà déclarée
	if sa.lookupSymbol(node.Name.Value) != nil {
		sa.addError("Type '%s' already declared. line:%d column:%d", node.Name.Value,
			node.Token.Line, node.Token.Column)
		return
	}

	// Créer le type de structure
	structType := &TypeInfo{
		Name:   node.Name.Value,
		Fields: make(map[string]*TypeInfo),
	}

	// Analyser les champs
	for _, field := range node.Fields {
		fieldType := sa.resolveTypeAnnotation(field.Type)
		structType.Fields[lower(field.Name.Value)] = fieldType
	}

	// Enregistrer le type
	sa.TypeTable[lower(node.Name.Value)] = structType
	sa.registerSymbol(node.Name.Value, StructSymbol, structType, node)
}

func (sa *SemanticAnalyzer) visitForStatement(node *ast.ForStatement, t *TypeInfo) {
	// Créer un nouveau scope pour la boucle
	loopScope := &Scope{
		Parent:  sa.CurrentScope,
		Symbols: make(map[string]*Symbol),
	}
	sa.CurrentScope.Children = append(sa.CurrentScope.Children, loopScope)

	oldScope := sa.CurrentScope
	sa.CurrentScope = loopScope

	// Analyser l'initialisation
	if node.Init != nil {
		sa.visitStatement(node.Init, t)
	}

	// Analyser la condition
	if node.Condition != nil {
		condType := sa.visitExpression(node.Condition)
		if condType.Name != "boolean" && condType.Name != "any" {
			sa.addError("The condition of a for loop must be boolean. line:%d column:%d",
				node.Token.Line, node.Token.Column)
		}
	}

	// Analyser l'update
	if node.Update != nil {
		sa.visitStatement(node.Update, t)
	}

	// Analyser le corps
	sa.visitBlockStatement(node.Body, t)

	// Restaurer le scope
	sa.CurrentScope = oldScope
}

func (sa *SemanticAnalyzer) visitIfStatement(node *ast.IfStatement, t *TypeInfo) {
	// Créer un nouveau scope pour la boucle
	loopScope := &Scope{
		Parent:  sa.CurrentScope,
		Symbols: make(map[string]*Symbol),
	}
	sa.CurrentScope.Children = append(sa.CurrentScope.Children, loopScope)

	oldScope := sa.CurrentScope
	sa.CurrentScope = loopScope

	// Analyser la condition
	if node.Condition != nil {
		condType := sa.visitExpression(node.Condition)
		if condType.Name != "boolean" && condType.Name != "any" {
			sa.addError("The condition of a If statement must be boolean. line:%d column:%d",
				node.Token.Line, node.Token.Column)
			return
		}
	}

	// Analyser l'update
	if node.Then != nil {
		sa.visitStatement(node.Then, t)
	}

	if node.Else != nil {
		sa.visitStatement(node.Else, t)
	}
	// Restaurer le scope
	sa.CurrentScope = oldScope
}

func (sa *SemanticAnalyzer) visitWhileStatement(node *ast.WhileStatement, t *TypeInfo) {
	// Créer un nouveau scope pour la boucle
	loopScope := &Scope{
		Parent:  sa.CurrentScope,
		Symbols: make(map[string]*Symbol),
	}
	sa.CurrentScope.Children = append(sa.CurrentScope.Children, loopScope)

	oldScope := sa.CurrentScope
	sa.CurrentScope = loopScope

	// Analyser la condition
	if node.Condition != nil {
		condType := sa.visitExpression(node.Condition)
		if condType.Name != "boolean" && condType.Name != "any" {
			sa.addError("The condition of a If statement must be boolean. line:%d column:%d",
				node.Token.Line, node.Token.Column)
			return
		}
	}

	// Analyser l'update
	if node.Body != nil {
		sa.visitStatement(node.Body, t)
	}

	// Restaurer le scope
	sa.CurrentScope = oldScope
}

func (sa *SemanticAnalyzer) visitForEachStatement(node *ast.ForEachStatement, t *TypeInfo) {
	// Créer un nouveau scope pour la boucle
	loopScope := &Scope{
		Parent:  sa.CurrentScope,
		Symbols: make(map[string]*Symbol),
	}
	if node.Variable == nil {
		sa.addError("Variable must be defined. Line:%d, column:%d", node.Token.Line, node.Token.Column)
		return
	}
	symbol := sa.lookupSymbol(node.Variable.Value)
	if symbol != nil {
		sa.addError("This variable '%s' already exists. Line:%d, column:%d", node.Variable.Value,
			node.Variable.Line(), node.Variable.Column())
		return
	}
	sa.CurrentScope.Children = append(sa.CurrentScope.Children, loopScope)
	varType := sa.visitExpression(node.Iterator)
	if varType == nil || node.Iterator == nil {
		sa.addError("Iterator '%s' must have a type. Line:%d, column:%d", node.Iterator.String(),
			node.Iterator.Line(), node.Iterator.Column())
		return
	}
	if !varType.IsArray {
		sa.addError("'%s' must be an iterator. Line:%d, column:%d", node.Iterator.String(),
			node.Iterator.Line(), node.Iterator.Column())
		return
	}
	oldScope := sa.CurrentScope
	sa.CurrentScope = loopScope
	sa.registerSymbol(node.Variable.Value, VariableSymbol, varType.ElementType, node)

	// Analyser l'update
	if node.Body != nil {
		sa.visitStatement(node.Body, t)
	}

	// Restaurer le scope
	sa.CurrentScope = oldScope
}

func (sa *SemanticAnalyzer) visitSwitchStatement(node *ast.SwitchStatement, t *TypeInfo) {
	// Analyser l'expression du switch
	switchType := sa.visitExpression(node.Expression)

	// Analyser les cases
	for _, caseStmt := range node.Cases {
		for _, expr := range caseStmt.Expressions {
			caseType := sa.visitExpression(expr)
			if !sa.areTypesCompatible(switchType, caseType) {
				sa.addError("Type incompatible dans case: attendu %s, got %s",
					switchType.Name, caseType.Name)
			}
		}

		// Créer un scope pour le case
		caseScope := &Scope{
			Parent:  sa.CurrentScope,
			Symbols: make(map[string]*Symbol),
		}
		sa.CurrentScope.Children = append(sa.CurrentScope.Children, caseScope)

		oldScope := sa.CurrentScope
		sa.CurrentScope = caseScope
		sa.visitBlockStatement(caseStmt.Body, t)
		sa.CurrentScope = oldScope
	}

	// Analyser le default
	if node.DefaultCase != nil {
		defaultScope := &Scope{
			Parent:  sa.CurrentScope,
			Symbols: make(map[string]*Symbol),
		}
		sa.CurrentScope.Children = append(sa.CurrentScope.Children, defaultScope)

		oldScope := sa.CurrentScope
		sa.CurrentScope = defaultScope
		sa.visitBlockStatement(node.DefaultCase, t)
		sa.CurrentScope = oldScope
	}
}

func (sa *SemanticAnalyzer) visitExpression(expr ast.Expression) *TypeInfo {
	if expr == nil {
		return nil
	}
	switch e := expr.(type) {
	case *ast.Identifier:
		return sa.visitIdentifier(e)
	case *ast.IntegerLiteral:
		return &TypeInfo{Name: "integer"}
	case *ast.TypeMember:
		return sa.visitTypeMember(e)
	case *ast.FloatLiteral:
		return &TypeInfo{Name: "float"}
	case *ast.StringLiteral:
		return &TypeInfo{Name: "string"}
	case *ast.BooleanLiteral:
		return &TypeInfo{Name: "boolean"}
	case *ast.DateTimeLiteral:
		if e.IsTime {
			return &TypeInfo{Name: "time"}
		}
		return &TypeInfo{Name: "date"}
	case *ast.DurationLiteral:
		return &TypeInfo{Name: "duration"}
	case *ast.ArrayLiteral:
		return sa.visitArrayLiteral(e)
	case *ast.StructLiteral:
		return sa.visitStructLiteral(e)
	case *ast.NullLiteral:
		return &TypeInfo{Name: "null"}
	case *ast.InfixExpression:
		return sa.visitInfixExpression(e)
	case *ast.PrefixExpression:
		return sa.visitPrefixExpression(e)
	case *ast.IndexExpression:
		return sa.visitIndexExpression(e)
	case *ast.SliceExpression:
		return sa.visitSliceExpression(e)
	case *ast.InExpression:
		return sa.visitInExpression(e)
	case *ast.ArrayFunctionCall:
		return sa.visitArrayFunctionCall(e)
	case *ast.SQLSelectStatement, *ast.SQLWithStatement:
		return &TypeInfo{Name: "sql_result"}
	case *ast.SQLInsertStatement, *ast.SQLUpdateStatement,
		*ast.SQLDeleteStatement:
		return &TypeInfo{Name: "integer"}
	default:
		return &TypeInfo{Name: "any"}
	}
}

func (sa *SemanticAnalyzer) visitSliceExpression(e *ast.SliceExpression) *TypeInfo {
	ts := sa.visitExpression(e.Start)
	te := sa.visitExpression(e.End)
	if ts != nil && ts.Name != "integer" {
		sa.addError("This expression '%s' must be integer. line:%d column:%d",
			e.String(), e.Start.Line(), e.Start.Column())
	}
	if te != nil && te.Name != "integer" {
		sa.addError("This expression '%s' must be integer. line:%d column:%d",
			e.String(), e.End.Line(), e.End.Column())
	}
	symbol := sa.lookupSymbol(e.Left.String())
	if symbol == nil {
		sa.addError("Non declared identifier: %s line:%d column:%d", e.Left.String(),
			e.Left.Line(), e.Left.Column())
		return &TypeInfo{Name: "void"}
	}
	return symbol.DataType
}

func (sa *SemanticAnalyzer) visitArrayFunctionCall(e *ast.ArrayFunctionCall) *TypeInfo {
	//retrouver la fonction dans le scope et retourner son type
	oldScope := sa.CurrentScope
	symbol := sa.lookupSymbol(e.Function.String())
	if symbol == nil {
		sa.addError("Non declared function: %s. line:%d column:%d", e.Function.Value,
			e.Function.Token.Line, e.Function.Token.Column)
		sa.CurrentScope = oldScope
		return &TypeInfo{Name: "void"}
	}
	Scope := symbol.Scope.Children[symbol.Index]
	if len(Scope.Symbols) == 0 && e.Array == nil {
		sa.CurrentScope = oldScope
		return symbol.DataType
	}
	if Scope == nil && e.Array != nil {
		sa.addError("The function '%s' does not have argument(s). line:%d column:%d", e.Function.Value,
			e.Function.Token.Line, e.Function.Token.Column)
		sa.CurrentScope = oldScope
		return &TypeInfo{Name: "void"}
	}
	if e.Array == nil && len(Scope.Symbols) > 0 {
		sa.addError("The function '%s' must have argument(s). line:%d column:%d", e.Function.Value,
			e.Function.Token.Line, e.Function.Token.Column)
		sa.CurrentScope = oldScope
		return &TypeInfo{Name: "void"}
	}

	if len(e.Arguments) > 0 && len(Scope.Symbols)-1 == 0 {
		sa.addError("The function '%s' does not have argument(s). line:%d column:%d", e.Function.Value,
			e.Function.Token.Line, e.Function.Token.Column)
		sa.CurrentScope = oldScope
		return &TypeInfo{Name: "void"}
	}
	if len(Scope.Symbols) != (len(e.Arguments) + 1) {
		sa.addError("The function '%s' expects %d argument(s), but got %d. line:%d column:%d", e.Function.Value,
			len(Scope.Symbols), len(e.Arguments), e.Function.Token.Line, e.Function.Token.Column)
		sa.CurrentScope = oldScope
		return &TypeInfo{Name: "void"}
	}
	args := make(map[int]string)

	for k, v := range Scope.Symbols {
		args[v.NoOrder] = k
	}
	currentType := sa.visitExpression(e.Array)
	expectedType := Scope.Symbols[args[0]]
	if !sa.areSameType(expectedType.DataType, currentType) {
		sa.addError("Type mismatch for argument '%s' in function '%s': expected %s, got %s. line:%d column:%d",
			e.Array.String(), e.Function.Value, expectedType.DataType.Name, currentType.Name,
			e.Function.Token.Line, e.Function.Token.Column)
		return &TypeInfo{Name: "void"}
	}
	var exists bool

	for k, arg := range e.Arguments {
		currentType = sa.visitExpression(arg)
		expectedType, exists = Scope.Symbols[args[k+1]]
		if !exists {
			sa.addError("The function '%s' does not have argument '%s'. line:%d column:%d", e.Function.Value,
				arg.String(), e.Function.Token.Line, e.Function.Token.Column)
			continue
		}
		if !sa.areSameType(expectedType.DataType, currentType) {
			sa.addError("Type mismatch for argument '%s' in function '%s': expected %s, got %s. line:%d column:%d",
				arg.String(), e.Function.Value, expectedType.DataType.Name, currentType.Name,
				e.Function.Token.Line, e.Function.Token.Column)
		}
	}
	sa.CurrentScope = oldScope
	return symbol.DataType
}

func (sa *SemanticAnalyzer) visitIdentifier(node *ast.Identifier) *TypeInfo {
	symbol := sa.lookupSymbol(node.Value)
	if symbol == nil {
		sa.addError("Non declared identifier: %s line:%d column:%d", node.Value, node.Token.Line,
			node.Token.Column)
		return &TypeInfo{Name: "any"}
	}
	return symbol.DataType
}

func (sa *SemanticAnalyzer) visitTypeMember(node *ast.TypeMember) *TypeInfo {
	switch t := node.Left.(type) {
	case *ast.Identifier:
		l := sa.lookupSymbol(t.Value)
		if l == nil {
			sa.addError("Non declared variable '%s'. line:%d, column:%d", node.String(),
				node.Line(), node.Column())
			return &TypeInfo{Name: "void"}
		}
		if l.Type == DbObjectSymbol {
			return &TypeInfo{Name: "any"}
		}
		if l.Type == VariableSymbol && !l.DataType.IsArray &&
			len(l.DataType.Fields) > 0 {
			switch node.Right.(type) {
			case *ast.Identifier:
				ta, exists := l.DataType.Fields[strings.ToLower(node.Right.(*ast.Identifier).Value)]
				if !exists {
					sa.addError("Field '%s' does not exist. line:%d, column:%d", node.Right.String(),
						node.Right.Line(), node.Right.Column())
					return &TypeInfo{Name: "void"}
				}
				return ta
			case *ast.ArrayFunctionCall:
				sa.addError("Invalid expression '%s'. line:%d, column:%d", node.Right.String(),
					node.Right.Line(), node.Right.Column())
				return &TypeInfo{Name: "void"}
			default:
				sa.addError("Invalid expression '%s'. line:%d, column:%d", node.Right.String(),
					node.Right.Line(), node.Right.Column())
				return &TypeInfo{Name: "void"}
			}
		}
	case *ast.TypeMember:
		return sa.visitTypeMember(t)
	case *ast.ArrayFunctionCall:
		sa.addError("Invalid expression '%s'. line:%d, column:%d", node.Left.String(),
			node.Left.Line(), node.Left.Column())
	default:
		sa.addError("Invalid expression '%s'. line:%d, column:%d", node.Left.String(),
			node.Left.Line(), node.Left.Column())
	}
	return &TypeInfo{Name: "void"}
}

func (sa *SemanticAnalyzer) visitArrayLiteral(node *ast.ArrayLiteral) *TypeInfo {
	if len(node.Elements) == 0 {
		return &TypeInfo{
			Name:        "array",
			IsArray:     true,
			ElementType: &TypeInfo{Name: "any"},
		}
	}

	// Vérifier que tous les éléments ont le même type
	firstType := sa.visitExpression(node.Elements[0])
	for i, elem := range node.Elements {
		elemType := sa.visitExpression(elem)
		if !sa.areTypesCompatible(firstType, elemType) {
			sa.addError("Type incompatible dans le tableau à la position %d", i)
		}
	}

	return &TypeInfo{
		Name:        "array",
		IsArray:     true,
		ElementType: firstType,
	}
}

func (sa *SemanticAnalyzer) ifExists(node *ast.StructLiteral) *TypeInfo {
	// keys := make([]string, 0)
	oldScope := sa.CurrentScope
	Scope := sa.CurrentScope
	var returnType *TypeInfo
	for {
		ok := false
		for _, sym := range Scope.Symbols {
			if sym.Type == StructSymbol {
				ok = true
				for _, field := range node.Fields {
					currentType := sa.visitExpression(field.Value)
					expectedType, exists := sym.DataType.Fields[lower(field.Name.Value)]
					if !exists || expectedType.Name != currentType.Name {
						ok = false
						break
					}
				}
				if ok {
					returnType = sym.DataType
					break
				}
			}
		}
		if ok {
			break
		}
		Scope = Scope.Parent
		if Scope == nil {
			break
		}
	}

	sa.CurrentScope = oldScope
	return returnType
}

func (sa *SemanticAnalyzer) visitStructLiteral(node *ast.StructLiteral) *TypeInfo {
	// Name   *Identifier
	// Fields []StructFieldLit
	if node == nil {
		return &TypeInfo{Name: "void"}
	}
	var resultType *TypeInfo
	if node.Name != nil {
		structType := sa.lookupSymbol(lower(node.Name.Value))
		if structType == nil {
			sa.addError("Type '%s' not declared. line:%d column:%d", node.Name.Value,
				node.Token.Line, node.Token.Column)
			return &TypeInfo{Name: "void"}
		}
		resultType = structType.DataType
	}
	if node.Name == nil {
		newInlineType := sa.ifExists(node)
		if newInlineType != nil {
			return newInlineType
		}
		newInlineType = &TypeInfo{
			Name:   "internal_struct_" + strconv.Itoa(sa.inType),
			Fields: make(map[string]*TypeInfo),
		}
		sa.inType++
		for _, field := range node.Fields {
			fieldType := sa.visitExpression(field.Value)
			newInlineType.Fields[lower(field.Name.Value)] = fieldType
		}
		sa.TypeTable[lower(newInlineType.Name)] = newInlineType
		sa.registerSymbol(newInlineType.Name, StructSymbol, newInlineType, node)
		return newInlineType
	}
	for _, elem := range node.Fields {
		elemType := sa.visitExpression(elem.Value)
		expectedType, exists := resultType.Fields[lower(elem.Name.Value)]
		if !exists {
			sa.addError("Field '%s' does not exist in type '%s'. line:%d column:%d",
				elem.Name.Value, resultType.Name, elem.Name.Token.Line, elem.Name.Token.Column)
			continue
		}
		if !sa.areTypesCompatible(expectedType, elemType) {
			sa.addError("Type '%s' mismatch. line:%d, column:%d", elem.Name.Value,
				elem.Name.Token.Line, elem.Name.Token.Column)
		}
	}
	return resultType
}

func (sa *SemanticAnalyzer) visitPrefixExpression(node *ast.PrefixExpression) *TypeInfo {
	rightType := sa.visitExpression(node.Right)
	switch node.Operator {
	case "-", "+":
		if rightType.Name == "integer" {
			return &TypeInfo{Name: "integer"}
		}
		if rightType.Name == "float" ||
			rightType.Name == "numeric" || rightType.Name == "decimal" {
			return &TypeInfo{Name: "float"}
		}
		sa.addError("'%s' non supported operation on %s",
			node.Operator, rightType.Name)
	case "not":
		if rightType.Name == "boolean" {
			return &TypeInfo{Name: rightType.Name}
		}
	case "is":
		if rightType.Name == "null" {
			return &TypeInfo{Name: "boolean"}
		}
	case "object":
		return &TypeInfo{Name: "table"}
	default:
		sa.addError("'%s' non supported operation on %s",
			node.Operator, rightType.Name)
	}

	return &TypeInfo{Name: "any"}
}
func (sa *SemanticAnalyzer) visitInfixExpression(node *ast.InfixExpression) *TypeInfo {
	leftType := sa.visitExpression(node.Left)
	rightType := sa.visitExpression(node.Right)

	switch node.Operator {
	case "%":
		// Opérations arithmétiques
		if leftType.Name == "integer" && rightType.Name == "integer" {
			return &TypeInfo{Name: "integer"}
		}
		sa.addError("Non supported operation '%s' between %s and %s",
			node.Operator, leftType.Name, rightType.Name)

	case "+", "-":
		// Opérations Date/Time + Duration
		if (leftType.Name == "date" || leftType.Name == "datetime" || leftType.Name == "time") && rightType.Name == "duration" {
			return leftType
		}
		if leftType.Name == "duration" && (rightType.Name == "datetime" || rightType.Name == "date" || rightType.Name == "time") {
			return rightType
		}
		// Duration + Duration
		if leftType.Name == "duration" && rightType.Name == "duration" {
			return &TypeInfo{Name: "duration"}
		}
		// Duration + Number = Duration
		if leftType.Name == "duration" && (rightType.Name == "integer" || rightType.Name == "float") {
			return &TypeInfo{Name: "duration"}
		}
		if (leftType.Name == "integer" || leftType.Name == "float") && rightType.Name == "duration" {
			return &TypeInfo{Name: "duration"}
		}
		if leftType.Name == "integer" && rightType.Name == "integer" {
			return &TypeInfo{Name: "integer"}
		}
		if (leftType.Name == "integer" || leftType.Name == "float") &&
			(rightType.Name == "integer" || rightType.Name == "float") {
			return &TypeInfo{Name: "float"}
		}
		if leftType.Name == "date" && (rightType.Name == "date" || rightType.Name == "datetime") && node.Operator == "-" {
			return &TypeInfo{Name: "duration"}
		}
		if (leftType.Name == "date" || leftType.Name == "datetime") && rightType.Name == "date" && node.Operator == "-" {
			return &TypeInfo{Name: "duration"}
		}
		if leftType.Name == "datetime" && (rightType.Name == "date" || rightType.Name == "datetime") && node.Operator == "-" {
			return &TypeInfo{Name: "duration"}
		}
		if (leftType.Name == "date" || leftType.Name == "datetime") && rightType.Name == "datetime" && node.Operator == "-" {
			return &TypeInfo{Name: "duration"}
		}
		if leftType.Name == "string" && rightType.Name == "string" && node.Operator == "+" {
			return &TypeInfo{Name: "string"}
		}
		if leftType.IsArray && rightType.IsArray &&
			sa.areTypesCompatible(leftType, rightType) &&
			node.Operator == "+" {
			return leftType
		}
		sa.addError("Unsupported operation '%s' between %s and %s",
			node.Operator, leftType.Name, rightType.Name)

	case "*", "/":
		// Duration * Number = Duration
		if leftType.Name == "duration" && (rightType.Name == "integer" || rightType.Name == "float") {
			return &TypeInfo{Name: "duration"}
		}
		if (leftType.Name == "integer" || leftType.Name == "float") && rightType.Name == "duration" {
			return &TypeInfo{Name: "duration"}
		}
		// Duration / Duration = Number
		if leftType.Name == "duration" && rightType.Name == "duration" {
			return &TypeInfo{Name: "float"}
		}
		// Duration / Number = Duration
		if leftType.Name == "duration" && (rightType.Name == "integer" || rightType.Name == "float") {
			return &TypeInfo{Name: "duration"}
		}
		if leftType.Name == "integer" && rightType.Name == "integer" {
			return &TypeInfo{Name: "integer"}
		}
		if (leftType.Name == "integer" || leftType.Name == "float") &&
			(rightType.Name == "integer" || rightType.Name == "float") {
			return &TypeInfo{Name: "float"}
		}
		if (leftType.Name == "integer" || leftType.Name == "float") &&
			(rightType.Name == "any") {
			return &TypeInfo{Name: leftType.Name}
		}
		if (rightType.Name == "integer" || rightType.Name == "float") &&
			(leftType.Name == "any") {
			return &TypeInfo{Name: leftType.Name}
		}
		sa.addError("Non supported '%s' operation between %s and %s",
			node.Operator, leftType.Name, rightType.Name)

	case "==", "!=", "<", ">", "<=", ">=":
		// Comparaisons de durées
		if leftType.Name == "duration" && rightType.Name == "duration" {
			return &TypeInfo{Name: "boolean"}
		}
		// Comparaisons Date/Time + Duration
		if (leftType.Name == "date" || leftType.Name == "time") && rightType.Name == "duration" {
			sa.addWarning("Comparison Date/Time with Duration - implicite conversion")
			return &TypeInfo{Name: "boolean"}
		}

		// Opérations de comparaison
		if !sa.areTypesCompatible(leftType, rightType) {
			sa.addError("Non authorize comparision between %s and %s",
				leftType.Name, rightType.Name)
			return &TypeInfo{Name: "void"}
		}
		return &TypeInfo{Name: "boolean"}

	case "and", "or":
		// Opérations booléennes
		if leftType.Name != "boolean" || rightType.Name != "boolean" {
			sa.addError("Operation '%s' requires booleans", node.Operator)
		}
		return &TypeInfo{Name: "boolean"}

	case "||":
		if leftType.IsArray && rightType.IsArray {
			if !sa.areTypesCompatible(leftType.ElementType, rightType.ElementType) {
				sa.addError("Impossible to concat arrays because of type mismatch: %s et %s",
					leftType.ElementType.Name, rightType.ElementType.Name)
				return &TypeInfo{Name: "void"}
			}
			return leftType.ElementType
		}
		// Opérations de concaténation de chaînes
		if leftType.Name != "string" || rightType.Name != "string" {
			sa.addError("invalid operation. Both operands of '||' must have the same type (string, array). got %s and %s",
				leftType.Name, rightType.Name)
			return &TypeInfo{Name: "void"}
		}
		return &TypeInfo{Name: "string"}
	default:
		sa.addError("Opérateur inconnu: %s. Line:%d, column:%d", node.Operator,
			node.Token.Line, node.Token.Column)
	}
	return &TypeInfo{Name: "any"}
}

func (sa *SemanticAnalyzer) visitIndexExpression(node *ast.IndexExpression) *TypeInfo {
	leftType := sa.visitExpression(node.Left)
	indexType := sa.visitExpression(node.Index)
	if indexType.Name != "integer" {
		sa.addError("Index must be an integer. line:%d, column:%d", node.Index.Line(), node.Index.Column())
		return &TypeInfo{Name: "void"}
	}
	if leftType == nil {
		sa.addError("Non declared variable '%s'. line:%d, column:%d", node.Left.String(),
			node.Left.Line(), node.Left.Column())
		return &TypeInfo{Name: "void"}
	}
	if !leftType.IsArray {
		sa.addError("The variable '%s' is not an array. line:%d, column:%d", node.Left.String(),
			node.Left.Line(), node.Left.Column())
		return &TypeInfo{Name: "void"}
	}
	if leftType.ElementType == nil {
		sa.addError("The variable '%s' has no element type. line:%d, column:%d", node.Left.String(),
			node.Left.Line(), node.Left.Column())
		return &TypeInfo{Name: "void"}
	}
	return leftType.ElementType
}

func (sa *SemanticAnalyzer) visitInExpression(node *ast.InExpression) *TypeInfo {
	leftType := sa.visitExpression(node.Left)
	rightType := sa.visitExpression(node.Right)

	if rightType.Name == "string" && sa.areTypesCompatible(leftType, rightType) {
		return &TypeInfo{Name: "boolean"}
	}

	if !rightType.IsArray {
		sa.addError("'%s' must be an array type in IN operation. line:%d, column:%d",
			node.Right.String(), node.Right.Line(), node.Right.Column())
		return &TypeInfo{Name: "void"}
	}

	if !sa.areTypesCompatible(leftType, rightType.ElementType) {
		sa.addError("Type '%s' mismatch for IN. line:%d, column:%d",
			leftType.Name, node.Left.Line(), node.Left.Column())
		return &TypeInfo{Name: "void"}
	}

	return &TypeInfo{Name: "boolean"}
}

// Méthodes utilitaires
func (sa *SemanticAnalyzer) resolveTypeAnnotation(ta *ast.TypeAnnotation) *TypeInfo {
	if ta.ArrayType != nil {
		var elementType *TypeInfo
		// Vérifier les tableaux multidimensionnels
		if ta.ArrayType.ElementType != nil &&
			ta.ArrayType.ElementType.ArrayType != nil {
			elementType = sa.resolveTypeAnnotation(&ast.TypeAnnotation{
				Token:     ta.ArrayType.ElementType.Token,
				Type:      ta.ArrayType.ElementType.Type,
				ArrayType: ta.ArrayType.ElementType.ArrayType,
			})
			return &TypeInfo{
				Name:        "array",
				IsArray:     true,
				ArraySize:   sa.getArraySize(ta.ArrayType.Size),
				ElementType: elementType,
			}
		}
		elementType = sa.resolveTypeAnnotation(&ast.TypeAnnotation{
			Token: ta.ArrayType.ElementType.Token,
			Type:  ta.ArrayType.ElementType.Type,
		})
		return &TypeInfo{
			Name:        "array",
			IsArray:     true,
			ArraySize:   sa.getArraySize(ta.ArrayType.Size),
			ElementType: elementType,
		}
	}

	// Vérifier si c'est un type défini
	if typeInfo, exists := sa.TypeTable[lower(ta.Type)]; exists {
		return typeInfo
	}

	// Type inconnu
	sa.addError("Type inconnu: %s", ta.Type)
	return &TypeInfo{Name: "any"}
}
func lower(s string) string {
	return strings.ToLower(s)
}
func (sa *SemanticAnalyzer) getArraySize(size *ast.IntegerLiteral) int64 {
	if size != nil {
		return size.Value
	}
	return -1 // Taille dynamique
}

func (sa *SemanticAnalyzer) areTypesCompatible(t1, t2 *TypeInfo) bool {
	if t1.Name == "any" || t2.Name == "any" {
		return true
	}

	if t1.IsArray && t2.IsArray {
		return sa.areTypesCompatible(t1.ElementType, t2.ElementType)
	}

	// Conversion implicite integer -> float
	if t1.Name == "integer" && t2.Name == "float" {
		return true
	}
	if t1.Name == "float" && t2.Name == "integer" {
		return true
	}
	if t1.IsArray && t2.IsArray {
		return t1.ElementType.Name == t2.ElementType.Name
	}
	return t1.Name == t2.Name
}

func (sa *SemanticAnalyzer) areSameType(t1, t2 *TypeInfo) bool {
	if t1.Name == "any" || t2.Name == "any" {
		return true
	}

	if t1.IsArray && t2.IsArray {
		return sa.areSameType(t1.ElementType, t2.ElementType)
	}

	if t1.Name == "float" && t2.Name == "integer" {
		return true
	}

	return t1.Name == t2.Name
}

func (sa *SemanticAnalyzer) lookupSymbol(name string) *Symbol {
	current := sa.CurrentScope
	for current != nil {
		if symbol, exists := current.Symbols[strings.ToLower(name)]; exists {
			return symbol
		}
		current = current.Parent
	}
	return nil
}

func (sa *SemanticAnalyzer) registerSymbol(name string, symType SymbolType, dataType *TypeInfo,
	node ast.Node, pos ...int) {
	index := -1
	noOrder := -1
	if len(pos) > 0 {
		index = pos[0]
	}
	if len(pos) > 1 {
		noOrder = pos[1]
	}

	symbol := &Symbol{
		Name:     name,
		Type:     symType,
		DataType: dataType,
		Scope:    sa.CurrentScope,
		Node:     node,
		Index:    index,
		NoOrder:  noOrder,
	}
	sa.CurrentScope.Symbols[lower(name)] = symbol
}

func (sa *SemanticAnalyzer) registerTempoSymbols(names []string) {
	for _, name := range names {
		symbol := &Symbol{
			Name:     name,
			Type:     DbObjectSymbol,
			DataType: &TypeInfo{Name: "any"},
			Scope:    sa.CurrentScope,
			Node:     nil,
			Index:    -1,
			NoOrder:  -1,
		}
		sa.CurrentScope.Symbols[lower(name)] = symbol
	}
}

func (sa *SemanticAnalyzer) addError(format string, args ...interface{}) {
	sa.Errors = append(sa.Errors, fmt.Sprintf(format, args...))
}

func (sa *SemanticAnalyzer) addWarning(format string, args ...interface{}) {
	sa.Warnings = append(sa.Warnings, fmt.Sprintf(format, args...))
}

// Méthodes restantes pour visiter les autres types d'expressions et instructions
func (sa *SemanticAnalyzer) visitReturnStatement(node *ast.ReturnStatement, t *TypeInfo) {
	if (t == nil || t.Name == "void") && node.ReturnValue != nil {
		sa.addError("Fonction does not return a value. line:%d column:%d",
			node.Token.Line, node.Token.Column)
		return
	}
	ti := sa.visitExpression(node.ReturnValue)
	if !sa.areTypesCompatible(t, ti) {
		sa.addError("Type of the Return value mismatch: expected %s, got %s. line:%d column:%d",
			t.Name, ti.Name, node.Token.Line, node.Token.Column)
	}
}

func (sa *SemanticAnalyzer) visitBlockStatement(node *ast.BlockStatement, t *TypeInfo) {
	// Créer un nouveau scope pour le bloc
	blockScope := &Scope{
		Parent:  sa.CurrentScope,
		Symbols: make(map[string]*Symbol),
	}
	sa.CurrentScope.Children = append(sa.CurrentScope.Children, blockScope)

	oldScope := sa.CurrentScope
	sa.CurrentScope = blockScope

	for _, stmt := range node.Statements {
		sa.visitStatement(stmt, t)
	}

	sa.CurrentScope = oldScope
}
