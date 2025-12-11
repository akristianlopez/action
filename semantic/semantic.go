package semantic

import (
	"fmt"
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
	TypeSymbol      SymbolType = "TYPE"
	ParameterSymbol SymbolType = "PARAMETER"
)

type Symbol struct {
	Name     string
	Type     SymbolType
	DataType *TypeInfo
	Scope    *Scope
	Node     ast.Node
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

type SemanticAnalyzer struct {
	CurrentScope *Scope
	GlobalScope  *Scope
	Errors       []string
	Warnings     []string
	TypeTable    map[string]*TypeInfo
}

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
	}
	returnType := &TypeInfo{}
	analyzer.registerSymbol("append", FunctionSymbol, returnType, ast.NullLiteral)
	// Enregistrer les types de base
	analyzer.registerBuiltinTypes()

	return analyzer
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
		sa.visitStatement(stmt)
	}
}

func (sa *SemanticAnalyzer) visitStatement(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.LetStatement:
		sa.visitLetStatement(s)
	case *ast.LetStatements:
		sa.visitLetStatements(s)
	case *ast.FunctionStatement:
		sa.visitFunctionStatement(s)
	case *ast.StructStatement:
		sa.visitStructStatement(s)
	case *ast.ForStatement:
		sa.visitForStatement(s)
	case *ast.SwitchStatement:
		sa.visitSwitchStatement(s)
	case *ast.ReturnStatement:
		sa.visitReturnStatement(s)
	case *ast.BlockStatement:
		sa.visitBlockStatement(s)
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
	panic("unimplemented")
}

func (sa *SemanticAnalyzer) visitSQLUpdateStatement(s *ast.SQLUpdateStatement) {
	panic("unimplemented")
}

func (sa *SemanticAnalyzer) visitSQLInsertStatement(s *ast.SQLInsertStatement) {
	panic("unimplemented")
}

func (sa *SemanticAnalyzer) visitSQLCreateObjectStatement(s *ast.SQLCreateObjectStatement) {
	panic("unimplemented")
}

func (sa *SemanticAnalyzer) visitExpressionStatement(s *ast.ExpressionStatement) {
	panic("unimplemented")
}

func (sa *SemanticAnalyzer) visitSQLSelectStatement(s *ast.SQLSelectStatement) {
	panic("unimplemented")
	// return &TypeInfo{Name: "sql_result"}
}

func (sa *SemanticAnalyzer) visitLetStatements(nodes *ast.LetStatements) {
	// Vérifier si la variable est déjà déclarée
	var varType *TypeInfo
	for _, node := range *nodes {
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

	// Enregistrer les paramètres
	oldScope := sa.CurrentScope
	sa.CurrentScope = funcScope

	for _, param := range node.Parameters {
		paramType := sa.resolveTypeAnnotation(param.Type)
		sa.registerSymbol(param.Name.Value, ParameterSymbol, paramType, param)
	}

	// Vérifier le type de retour
	var returnType *TypeInfo
	if node.ReturnType != nil {
		returnType = sa.resolveTypeAnnotation(node.ReturnType)
	} else {
		returnType = &TypeInfo{Name: "void"}
	}

	// Analyser le corps de la fonction
	sa.visitBlockStatement(node.Body)

	// Restaurer le scope
	sa.CurrentScope = oldScope

	// Enregistrer la fonction
	sa.registerSymbol(node.Name.Value, FunctionSymbol, returnType, node)
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
		structType.Fields[field.Name.Value] = fieldType
	}

	// Enregistrer le type
	sa.TypeTable[node.Name.Value] = structType
	sa.registerSymbol(node.Name.Value, StructSymbol, structType, node)
}

func (sa *SemanticAnalyzer) visitForStatement(node *ast.ForStatement) {
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
		sa.visitStatement(node.Init)
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
		sa.visitStatement(node.Update)
	}

	// Analyser le corps
	sa.visitBlockStatement(node.Body)

	// Restaurer le scope
	sa.CurrentScope = oldScope
}

func (sa *SemanticAnalyzer) visitSwitchStatement(node *ast.SwitchStatement) {
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
		sa.visitBlockStatement(caseStmt.Body)
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
		sa.visitBlockStatement(node.DefaultCase)
		sa.CurrentScope = oldScope
	}
}

func (sa *SemanticAnalyzer) visitExpression(expr ast.Expression) *TypeInfo {
	switch e := expr.(type) {
	case *ast.Identifier:
		return sa.visitIdentifier(e)
	case *ast.IntegerLiteral:
		return &TypeInfo{Name: "integer"}
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
		return &TypeInfo{Name: "struct"}
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
	//retrouver sa definition et retourner son type
	return &TypeInfo{Name: "array"}
}

func (sa *SemanticAnalyzer) visitArrayFunctionCall(e *ast.ArrayFunctionCall) *TypeInfo {
	//retrouver la fonction dans le scope et retourner son type
	symbol := sa.lookupSymbol(e.Function.String())
	if symbol == nil {
		sa.addError("Non declared function: %s, line:%d column:%d", e.Function.Value,
			e.Function.Token.Line, e.Function.Token.Column)
		return &TypeInfo{Name: "any"}
	}
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
			return &TypeInfo{Name: "object"}
		}
	case "object":
		return &TypeInfo{Name: "table"}
	default:
		sa.addError("'%s' non supported operation on %s",
			node.Operator, rightType.Name)
	}

	/*
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
		p.registerPrefix(token.LENGTH, p.parseArrayFunctionCall)
		p.registerPrefix(token.APPEND, p.parseArrayFunctionCall)
		p.registerPrefix(token.PREPEND, p.parseArrayFunctionCall)
		p.registerPrefix(token.REMOVE, p.parseArrayFunctionCall)
		p.registerPrefix(token.NULL, p.parseNullLiteral)
		// p.registerPrefix(token.SLICE, p.parseArrayFunctionCall)
		p.registerPrefix(token.CONTAINS, p.parseArrayFunctionCall)
		p.registerPrefix(token.DURATION_LIT, p.parseDurationLiteral)

	*/
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
		if (leftType.Name == "integer" || leftType.Name == "float") &&
			(rightType.Name == "integer" || rightType.Name == "float") {
			return &TypeInfo{Name: "float"}
		}
		if leftType.Name == "string" && rightType.Name == "string" && node.Operator == "+" {
			return &TypeInfo{Name: "string"}
		}
		sa.addError("Opération '%s' non supportée entre %s et %s",
			node.Operator, leftType.Name, rightType.Name)

	case "+", "-":
		// Opérations Date/Time + Duration
		if (leftType.Name == "date" || leftType.Name == "time") && rightType.Name == "duration" {
			return leftType
		}
		if leftType.Name == "duration" && (rightType.Name == "date" || rightType.Name == "time") {
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
		if leftType.Name == "string" && rightType.Name == "string" && node.Operator == "+" {
			return &TypeInfo{Name: "string"}
		}
		sa.addError("Opération '%s' non supportée entre %s et %s",
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
		if leftType.Name == "string" && rightType.Name == "string" && node.Operator == "+" {
			return &TypeInfo{Name: "string"}
		}
		sa.addError("Opération '%s' non supportée entre %s et %s",
			node.Operator, leftType.Name, rightType.Name)

	case "==", "!=", "<", ">", "<=", ">=":
		// Comparaisons de durées
		if leftType.Name == "duration" && rightType.Name == "duration" {
			return &TypeInfo{Name: "boolean"}
		}
		// Comparaisons Date/Time + Duration
		if (leftType.Name == "date" || leftType.Name == "time") && rightType.Name == "duration" {
			sa.addWarning("Comparaison Date/Time avec Duration - conversion implicite")
			return &TypeInfo{Name: "boolean"}
		}

		// Opérations de comparaison
		if !sa.areTypesCompatible(leftType, rightType) {
			sa.addError("Comparaison impossible entre %s et %s",
				leftType.Name, rightType.Name)
		}
		return &TypeInfo{Name: "boolean"}

	case "and", "or":
		// Opérations booléennes
		if leftType.Name != "boolean" || rightType.Name != "boolean" {
			sa.addError("Opération '%s' requiert des booléens", node.Operator)
		}
		return &TypeInfo{Name: "boolean"}
	}

	return &TypeInfo{Name: "any"}
}

func (sa *SemanticAnalyzer) visitIndexExpression(node *ast.IndexExpression) *TypeInfo {
	leftType := sa.visitExpression(node.Left)
	indexType := sa.visitExpression(node.Index)

	if !leftType.IsArray {
		sa.addError("L'indexation n'est possible que sur les tableaux")
		return &TypeInfo{Name: "any"}
	}

	if indexType.Name != "integer" {
		sa.addError("L'index doit être un entier")
	}

	return leftType.ElementType
}

func (sa *SemanticAnalyzer) visitInExpression(node *ast.InExpression) *TypeInfo {
	leftType := sa.visitExpression(node.Left)
	rightType := sa.visitExpression(node.Right)

	if !rightType.IsArray {
		sa.addError("L'opérande droit de IN doit être un tableau")
	}

	if !sa.areTypesCompatible(leftType, rightType.ElementType) {
		sa.addError("Type incompatible pour l'opérateur IN")
	}

	return &TypeInfo{Name: "boolean"}
}

// Méthodes utilitaires
func (sa *SemanticAnalyzer) resolveTypeAnnotation(ta *ast.TypeAnnotation) *TypeInfo {
	if ta.ArrayType != nil {
		elementType := sa.resolveTypeAnnotation(&ast.TypeAnnotation{
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
	if typeInfo, exists := sa.TypeTable[ta.Type]; exists {
		return typeInfo
	}

	// Type inconnu
	sa.addError("Type inconnu: %s", ta.Type)
	return &TypeInfo{Name: "any"}
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

	return t1.Name == t2.Name
}

func (sa *SemanticAnalyzer) lookupSymbol(name string) *Symbol {
	current := sa.CurrentScope
	for current != nil {
		if symbol, exists := current.Symbols[name]; exists {
			return symbol
		}
		current = current.Parent
	}
	return nil
}

func (sa *SemanticAnalyzer) registerSymbol(name string, symType SymbolType, dataType *TypeInfo, node ast.Node) {
	symbol := &Symbol{
		Name:     name,
		Type:     symType,
		DataType: dataType,
		Scope:    sa.CurrentScope,
		Node:     node,
	}
	sa.CurrentScope.Symbols[name] = symbol
}

func (sa *SemanticAnalyzer) addError(format string, args ...interface{}) {
	sa.Errors = append(sa.Errors, fmt.Sprintf(format, args...))
}

func (sa *SemanticAnalyzer) addWarning(format string, args ...interface{}) {
	sa.Warnings = append(sa.Warnings, fmt.Sprintf(format, args...))
}

// Méthodes restantes pour visiter les autres types d'expressions et instructions
func (sa *SemanticAnalyzer) visitReturnStatement(node *ast.ReturnStatement) {
	// TODO: Vérifier la compatibilité avec le type de retour de la fonction
}

func (sa *SemanticAnalyzer) visitBlockStatement(node *ast.BlockStatement) {
	// Créer un nouveau scope pour le bloc
	blockScope := &Scope{
		Parent:  sa.CurrentScope,
		Symbols: make(map[string]*Symbol),
	}
	sa.CurrentScope.Children = append(sa.CurrentScope.Children, blockScope)

	oldScope := sa.CurrentScope
	sa.CurrentScope = blockScope

	for _, stmt := range node.Statements {
		sa.visitStatement(stmt)
	}

	sa.CurrentScope = oldScope
}
