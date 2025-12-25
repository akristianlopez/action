package optimizer

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/akristianlopez/action/ast"
	// "github.com/akristianlopez/action/token"
)

type Optimizer struct {
	Optimizations []Optimization
	Stats         OptimizationStats
	Warnings      []string
}

type OptimizationStats struct {
	ConstantFolds     int
	DeadCodeRemovals  int
	InlineExpansions  int
	LoopOptimizations int
}

type Optimization interface {
	Name() string
	Apply(*ast.Program) *ast.Program
	CanApply(*ast.Program) bool
}

type ConstantFolding struct{}
type DeadCodeElimination struct{}
type FunctionInlining struct{}
type LoopOptimization struct{}

func NewOptimizer() *Optimizer {
	return &Optimizer{
		Optimizations: []Optimization{
			&ConstantFolding{},
			&DeadCodeElimination{},
			&LoopOptimization{},
			&FunctionInlining{},
		},
		Stats:    OptimizationStats{},
		Warnings: make([]string, 0),
	}
}

var Warnings func(format string, args ...interface{})
var IncrementFolding func()
var IncrementLoopOptimization func()

func (o *Optimizer) IncrementConstantFolding() {
	o.Stats.ConstantFolds++
}

func (o *Optimizer) IncrementLoopOptimization() {
	o.Stats.LoopOptimizations++
}
func (o *Optimizer) Optimize(program *ast.Program) *ast.Program {
	optimized := program
	Warnings = o.addWarning
	IncrementFolding = o.IncrementConstantFolding
	IncrementLoopOptimization = o.IncrementLoopOptimization

	// Appliquer les optimisations en plusieurs passes
	for i := 0; i < 10; i++ { // Maximum 10 passes
		changed := false
		for _, opt := range o.Optimizations {
			oSize := len(optimized.Statements)
			if opt.CanApply(optimized) {
				optimized = opt.Apply(optimized)
				changed = true

				// Mettre à jour les statistiques
				switch opt.(type) {
				case *ConstantFolding:
					// if len(optimized.Statements) < oSize {
					// 	o.Stats.ConstantFolds++
					// }
				case *DeadCodeElimination:
					if len(optimized.Statements) < oSize {
						o.Stats.DeadCodeRemovals++
					}
				case *FunctionInlining:
					o.Stats.InlineExpansions++
				case *LoopOptimization:
					// 	o.Stats.LoopOptimizations++
				}
			}
		}

		if !changed {
			break
		}
	}

	return optimized
}

func (o *Optimizer) addWarning(format string, args ...interface{}) {
	o.Warnings = append(o.Warnings, fmt.Sprintf(format, args...))
}

// CONSTANT FOLDING
func (cf *ConstantFolding) Name() string { return "ConstantFolding" }

func (cf *ConstantFolding) CanApply(program *ast.Program) bool {
	return hasConstantExpressions(program)
}

func (cf *ConstantFolding) Apply(program *ast.Program) *ast.Program {
	optimized := &ast.Program{
		ActionName: program.ActionName,
		Statements: []ast.Statement{},
	}

	for _, stmt := range program.Statements {
		switch s := stmt.(type) {
		case *ast.LetStatements:
			for _, v := range *s {
				optimized.Statements = append(optimized.Statements, foldConstantsInStatement(&v))
			}
		default:
			optimized.Statements = append(optimized.Statements, foldConstantsInStatement(stmt))
		}
	}
	return optimized
}

func foldConstantsInStatement(stmt ast.Statement) ast.Statement {
	switch s := stmt.(type) {
	case *ast.LetStatement:
		return foldLetStatement(s)
	case *ast.ExpressionStatement:
		return foldExpressionStatement(s)
	case *ast.ReturnStatement:
		return foldReturnStatement(s)
	case *ast.BlockStatement:
		return foldBlockStatement(s)
	case *ast.ForStatement:
		return foldForStatement(s)
	case *ast.WhileStatement:
		return foldWhileStatement(s)
	case *ast.ForEachStatement:
		return foldForEachStatement(s)
	case *ast.FunctionStatement:
		return foldFunctionStatement(s)
	case *ast.IfStatement:
		return foldIfStatement(s)
	case *ast.SwitchStatement:
		return foldSwitchStatement(s)
	default:
		return s
	}
}

func foldFunctionStatement(stmt *ast.FunctionStatement) ast.Statement {
	return &ast.FunctionStatement{
		Token:      stmt.Token,
		Name:       stmt.Name,
		Parameters: stmt.Parameters,
		ReturnType: stmt.ReturnType,
		Body:       foldBlockStatement(stmt.Body),
	}
}

func foldIfStatement(stmt *ast.IfStatement) ast.Statement {
	return &ast.IfStatement{
		Token:     stmt.Token,
		Condition: foldExpression(stmt.Condition),
		Then:      foldBlockStatement(stmt.Then),
		Else:      foldBlockStatement(stmt.Else),
	}
}

func foldForEachStatement(s *ast.ForEachStatement) ast.Statement {
	return &ast.ForEachStatement{
		Token:    s.Token,
		Variable: s.Variable,
		Iterator: foldExpression(s.Iterator),
		Body:     foldBlockStatement(s.Body),
	}
}

func foldLetStatement(stmt *ast.LetStatement) *ast.LetStatement {
	if stmt.Value != nil {
		folded := foldExpression(stmt.Value)
		if folded != stmt.Value {
			// IncrementFolding()
			return &ast.LetStatement{
				Token: stmt.Token,
				Name:  stmt.Name,
				Type:  stmt.Type,
				Value: folded,
			}
		}
	}
	return stmt
}

func foldExpression(expr ast.Expression) ast.Expression {
	switch e := expr.(type) {
	case *ast.InfixExpression:
		return foldInfixExpression(e)
	case *ast.PrefixExpression:
		return foldPrefixExpression(e)
	default:
		return e
	}
}

func foldInfixExpression(expr *ast.InfixExpression) ast.Expression {
	left := foldExpression(expr.Left)
	right := foldExpression(expr.Right)

	// Vérifier si les deux côtés sont des littéraux
	if isConstant(left) && isConstant(right) {
		// Évaluer l'expression à la compilation
		switch expr.Operator {
		case "+":
			if l, ok := left.(*ast.IntegerLiteral); ok {
				if r, ok := right.(*ast.IntegerLiteral); ok {
					IncrementFolding()
					return &ast.IntegerLiteral{
						Token: expr.Token,
						Value: l.Value + r.Value,
					}
				}
			}
			if l, ok := left.(*ast.FloatLiteral); ok {
				if r, ok := right.(*ast.FloatLiteral); ok {
					IncrementFolding()
					return &ast.FloatLiteral{
						Token: expr.Token,
						Value: l.Value + r.Value,
					}
				}
			}

		case "-":
			if l, ok := left.(*ast.IntegerLiteral); ok {
				if r, ok := right.(*ast.IntegerLiteral); ok {
					IncrementFolding()
					return &ast.IntegerLiteral{
						Token: expr.Token,
						Value: l.Value - r.Value,
					}
				}
			}
			if l, ok := left.(*ast.FloatLiteral); ok {
				if r, ok := right.(*ast.FloatLiteral); ok {
					IncrementFolding()
					return &ast.FloatLiteral{
						Token: expr.Token,
						Value: l.Value - r.Value,
					}
				}
			}

		case "*":
			if l, ok := left.(*ast.IntegerLiteral); ok {
				if r, ok := right.(*ast.IntegerLiteral); ok {
					IncrementFolding()
					return &ast.IntegerLiteral{
						Token: expr.Token,
						Value: l.Value * r.Value,
					}
				}
			}
			if l, ok := left.(*ast.FloatLiteral); ok {
				if r, ok := right.(*ast.FloatLiteral); ok {
					IncrementFolding()
					return &ast.FloatLiteral{
						Token: expr.Token,
						Value: l.Value * r.Value,
					}
				}
			}
		case "/":
			if l, ok := left.(*ast.IntegerLiteral); ok {
				if r, ok := right.(*ast.IntegerLiteral); ok {
					if r.Value != 0 {
						IncrementFolding()
						return &ast.IntegerLiteral{
							Token: expr.Token,
							Value: l.Value / r.Value,
						}
					}
				}
			}
			if l, ok := left.(*ast.FloatLiteral); ok {
				if r, ok := right.(*ast.FloatLiteral); ok {
					if r.Value != 0 {
						IncrementFolding()
						return &ast.FloatLiteral{
							Token: expr.Token,
							Value: l.Value / r.Value,
						}
					}
				}
			}
		case "==":
			if l, ok := left.(*ast.IntegerLiteral); ok {
				if r, ok := right.(*ast.IntegerLiteral); ok {
					IncrementFolding()
					return &ast.BooleanLiteral{
						Token: expr.Token,
						Value: l.Value == r.Value,
					}
				}
			}
		case "!=":
			if l, ok := left.(*ast.IntegerLiteral); ok {
				if r, ok := right.(*ast.IntegerLiteral); ok {
					IncrementFolding()
					return &ast.BooleanLiteral{
						Token: expr.Token,
						Value: l.Value != r.Value,
					}
				}
			}
		case "<":
			if l, ok := left.(*ast.IntegerLiteral); ok {
				if r, ok := right.(*ast.IntegerLiteral); ok {
					IncrementFolding()
					return &ast.BooleanLiteral{
						Token: expr.Token,
						Value: l.Value < r.Value,
					}
				}
			}
		case ">":
			if l, ok := left.(*ast.IntegerLiteral); ok {
				if r, ok := right.(*ast.IntegerLiteral); ok {
					IncrementFolding()
					return &ast.BooleanLiteral{
						Token: expr.Token,
						Value: l.Value > r.Value,
					}
				}
			}
		case "<=":
			if l, ok := left.(*ast.IntegerLiteral); ok {
				if r, ok := right.(*ast.IntegerLiteral); ok {
					IncrementFolding()
					return &ast.BooleanLiteral{
						Token: expr.Token,
						Value: l.Value <= r.Value,
					}
				}
			}
		case ">=":
			if l, ok := left.(*ast.IntegerLiteral); ok {
				if r, ok := right.(*ast.IntegerLiteral); ok {
					IncrementFolding()
					return &ast.BooleanLiteral{
						Token: expr.Token,
						Value: l.Value >= r.Value,
					}
				}
			}
		}
	}

	return &ast.InfixExpression{
		Token:    expr.Token,
		Left:     left,
		Operator: expr.Operator,
		Right:    right,
	}
}

func isConstant(expr ast.Expression) bool {
	//Add datetime literals and other constants
	switch expr.(type) {
	case *ast.IntegerLiteral, *ast.FloatLiteral, *ast.StringLiteral,
		*ast.BooleanLiteral, *ast.DurationLiteral:
		return true
	default:
		return false
	}
}

// DEAD CODE ELIMINATION
func (dce *DeadCodeElimination) Name() string { return "DeadCodeElimination" }

func (dce *DeadCodeElimination) CanApply(program *ast.Program) bool {
	return true // Toujours applicable
}

func (dce *DeadCodeElimination) Apply(program *ast.Program) *ast.Program {
	//Eliminating of the unused let statements
	optimized := &ast.Program{
		ActionName: program.ActionName,
		Statements: []ast.Statement{},
	}
	for _, stmt := range program.Statements {
		switch s := stmt.(type) {
		case *ast.LetStatements:
			var usedLets []ast.Statement
			for _, v := range *s {
				if !isDeadCode(&v, program) {
					usedLets = append(usedLets, &v)
					continue
				}
				Warnings("Dead code eliminated: variable '%s' is not used. Line:%d, column:%d", v.Name.Value,
					v.Token.Line, v.Token.Column)
			}
			if len(usedLets) > 0 {
				optimized.Statements = append(optimized.Statements, usedLets...)
			}
			continue
		case *ast.LetStatement:
			if !isDeadCode(s, program) {
				optimized.Statements = append(optimized.Statements, s)
				continue
			}
			Warnings("Dead code eliminated: variable '%s' is not used. Line:%d, column:%d", s.Name.Value,
				s.Token.Line, s.Token.Column)
			continue
		case *ast.FunctionStatement:
			if !isDeadCode(s, program) {
				optimized.Statements = append(optimized.Statements, s)
				continue
			}
			Warnings("Dead code eliminated: function '%s' is not used. Line:%d, column:%d", s.Name.Value,
				s.Token.Line, s.Token.Column)
			continue
		case *ast.StructStatement:
			if !isDeadCode(s, program) {
				optimized.Statements = append(optimized.Statements, s)
				continue
			}
			Warnings("Dead code eliminated: struct '%s' is not used. Line:%d, column:%d", s.Name.Value,
				s.Token.Line, s.Token.Column)
			continue
		case *ast.IfStatement:
			if !isDeadCode(s, program) {
				optimized.Statements = append(optimized.Statements, s)
				continue
			}
			if isUnrichabled(s.Condition) && s.Else != nil {
				Warnings("Dead code eliminated: Then statement at Line:%d, column:%d is not reachable.",
					s.Token.Line, s.Token.Column)
				optimized.Statements = append(optimized.Statements, s.Else)
				continue
			}
			Warnings("Dead code eliminated: if statement at Line:%d, column:%d has empty body.",
				s.Token.Line, s.Token.Column)
		case *ast.WhileStatement:
			if !isDeadCode(s, program) {
				optimized.Statements = append(optimized.Statements, s)
				continue
			}
			if isUnrichabled(s.Condition) {
				Warnings("Dead code eliminated: while statement at Line:%d, column:%d is not reachable.",
					s.Token.Line, s.Token.Column)
				continue
			}
			Warnings("Dead code eliminated: while statement at Line:%d, column:%d has empty body.",
				s.Token.Line, s.Token.Column)
		case *ast.ForStatement:
			if !isDeadCode(s, program) {
				optimized.Statements = append(optimized.Statements, s)
				continue
			}
			if isUnrichabled(s.Condition) {
				Warnings("Dead code eliminated: for statement at Line:%d, column:%d is not reachable.",
					s.Token.Line, s.Token.Column)
				continue
			}
			Warnings("Dead code eliminated: for statement at Line:%d, column:%d has empty body.",
				s.Token.Line, s.Token.Column)
		case *ast.ForEachStatement:
			if !isDeadCode(s, program) {
				optimized.Statements = append(optimized.Statements, s)
				continue
			}
			Warnings("Dead code eliminated: foreach statement at Line:%d, column:%d has empty body.",
				s.Token.Line, s.Token.Column)
		case *ast.SwitchStatement:
			if !isDeadCode(s, program) {
				optimized.Statements = append(optimized.Statements, s)
				continue
			}
			Warnings("Dead code eliminated: switch statement at Line:%d, column:%d has all cases dead.",
				s.Token.Line, s.Token.Column)
		case *ast.ReturnStatement:
			if !isDeadCode(s, program) {
				optimized.Statements = append(optimized.Statements, s)
				continue
			}
			Warnings("Dead code eliminated: return statement at Line:%d, column:%d has no effect.",
				s.Token.Line, s.Token.Column)
		default:
			optimized.Statements = append(optimized.Statements, stmt)
		}
	}
	return optimized
}

func isUsed(name string, actions *ast.Program) bool {
	// Vérifier si la variable est utilisée dans le programme
	for _, stmt := range actions.Statements {
		switch s := stmt.(type) {
		case *ast.LetStatement:
			// if s.Name.Value == name {
			// 	return true // La variable est définie ici
			// }
			if !strings.EqualFold(name, s.Name.Value) && s.Value != nil && isVariableUsedInExpression(s.Value, name) {
				return true // La variable est utilisée dans une expression
			}
		case *ast.LetStatements:
			for _, v := range *s {
				// if v.Name.Value == name {
				// 	return true // La variable est définie ici
				// }
				if !strings.EqualFold(name, v.Name.Value) && v.Value != nil && isVariableUsedInExpression(v.Value, name) {
					return true // La variable est définie ici
				}
			}
		case *ast.ExpressionStatement:
			if isVariableUsedInExpression(s.Expression, name) {
				return true // La variable est utilisée dans une expression
			}
		case *ast.ReturnStatement:
			if isVariableUsedInExpression(s.ReturnValue, name) {
				return true // La variable est utilisée dans une valeur de retour
			}
		default:
			if isFunctionUsedInStatement(stmt, name) {
				return true
			}
		}
	}
	return false
}

func isVariableUsedInExpression(expr ast.Expression, name string) bool {
	switch e := expr.(type) {
	case *ast.Identifier:
		return strings.EqualFold(e.Value, name)
	case *ast.InfixExpression:
		return isVariableUsedInExpression(e.Left, name) || isVariableUsedInExpression(e.Right, name)
	case *ast.PrefixExpression:
		return isVariableUsedInExpression(e.Right, name)
	case *ast.ArrayFunctionCall:
		if e.Function != nil && isVariableUsedInExpression(e.Function, name) {
			return true
		}
		if isVariableUsedInExpression(e.Array, name) {
			return true
		}
		for _, arg := range e.Arguments {
			if isVariableUsedInExpression(arg, name) {
				return true
			}
		}
	}
	return false
}

func isDeadCode(stmt ast.Statement, actions *ast.Program) bool {
	// Identifier le code mort (variables non utilisées, etc.)
	switch s := stmt.(type) {
	case *ast.LetStatement:
		return !isUsed(s.Name.Value, actions)
	case *ast.FunctionStatement:
		// isFunctionDead vérifie si une fonction est utilisée ailleurs dans le programme.
		// Une référence depuis sa propre définition (p.ex. appel récursif) n'est pas considérée comme utilisation externe.
		return isFunctionDead(s, actions)
	case *ast.StructStatement:
		// Les structures ne sont pas considérées comme du code mort ici
		return isDefinedStructDead(s, actions)
	case *ast.ExpressionStatement:
		// Les expressions sans effet de bord peuvent être mortes
		return isPureExpression(s.Expression)
	case *ast.IfStatement:
		// Si les deux branches sont mortes, le if est mort
		thenDead := s.Then == nil || isDeadCode(s.Then, actions)
		elseDead := s.Else == nil || isDeadCode(s.Else, actions)
		if thenDead && elseDead {
			return true
		}
		if isUnrichabled(s.Condition) {
			return true
		}
		return false
	case *ast.WhileStatement:
		// Si le corps de la boucle est vide ou mort, la boucle est morte
		if s.Body == nil || len(s.Body.Statements) == 0 {
			return true
		}
		if isUnrichabled(s.Condition) {
			return true
		}
		return isDeadCode(s.Body, actions)
	case *ast.ForStatement:
		// Si le corps de la boucle est vide ou mort, la boucle est morte
		if s.Body == nil || len(s.Body.Statements) == 0 {
			return true
		}
		if isUnrichabled(s.Condition) {
			return true
		}
		return isDeadCode(s.Body, actions)
	case *ast.ForEachStatement:
		// Si le corps de la boucle est vide ou mort, la boucle est morte
		if s.Body == nil || len(s.Body.Statements) == 0 {
			return true
		}
		return isDeadCode(s.Body, actions)
	case *ast.SwitchStatement:
		// Si tous les cas sont morts, le switch est mort
		allDead := true
		for _, c := range s.Cases {
			if c.Body != nil && !isDeadCode(c.Body, actions) {
				allDead = false
				break
			}
		}
		if s.DefaultCase != nil && !isDeadCode(s.DefaultCase, actions) {
			allDead = false
		}
		return allDead
	case *ast.BlockStatement:
		// Si tous les statements dans le bloc sont morts, le bloc est mort
		for _, st := range s.Statements {
			if !isDeadCode(st, actions) {
				return false
			}
		}
		return true
	case *ast.ReturnStatement:
		// Un return est mort s'il n'y a pas de valeur de retour ou si la valeur est pure
		if s.ReturnValue == nil {
			return true
		}
		return isPureExpression(s.ReturnValue)
	default:
		return false
	}
}

func isUnrichabled(expr ast.Expression) bool {
	// Détermine si une expression est toujours fausse (p.ex. while(false))
	if boolLit, ok := expr.(*ast.BooleanLiteral); ok {
		return !boolLit.Value
	}
	return false
}

func isDefinedStructDead(s *ast.StructStatement, actions *ast.Program) bool {
	// Retourne true si la structure n'est utilisée nulle part comme type.
	// Vérifie les déclarations du programme (variables, fonctions, champs, paramètres, retours, etc.)
	if s == nil || s.Name == nil {
		return true
	}
	name := s.Name.Value

	for _, stmt := range actions.Statements {
		// ignorer la définition elle-même
		if st, ok := stmt.(*ast.StructStatement); ok && st == s {
			continue
		}
		if isStructNameUsedAsType(stmt, name) {
			return false
		}
	}

	return true
}

//
// Helpers: recherche récursive via reflection dans les champs dont le nom
// suggère qu'ils représentent des types (Type, ReturnType, Parameters, Fields, etc.)
// NOTE: this implementation requires importing "reflect".
// importedReflectPlaceholder struct{} // placeholder to hint that reflect is used; remove if you add the import

func isStructNameUsedAsType(node interface{}, name string) bool {
	if node == nil {
		return false
	}
	return scanForTypeLikeField(reflect.ValueOf(node), name)
}

func scanForTypeLikeField(v reflect.Value, name string) bool {
	if !v.IsValid() {
		return false
	}
	// Dereference pointers
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return false
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Struct:
		t := v.Type()
		for i := 0; i < t.NumField(); i++ {
			fieldVal := v.Field(i)
			fieldName := t.Field(i).Name
			lname := strings.ToLower(fieldName)

			// Si le nom du champ suggère qu'il s'agit d'un emplacement de type, rechercher l'identifiant dedans.
			if strings.Contains(lname, "type") || strings.Contains(lname, "return") ||
				strings.Contains(lname, "param") || strings.Contains(lname, "field") ||
				strings.Contains(lname, "fields") || strings.Contains(lname, "typ") {
				if containsIdentifierWithName(fieldVal, name) {
					return true
				}
			}

			// Toujours descendre récursivement, au cas où la structure de types serait imbriquée.
			if scanForTypeLikeField(fieldVal, name) {
				return true
			}
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			if scanForTypeLikeField(v.Index(i), name) {
				return true
			}
		}
	}
	return false
}

func containsIdentifierWithName(v reflect.Value, name string) bool {
	if !v.IsValid() {
		return false
	}
	// Dereference pointers
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return false
		}
		v = v.Elem()
	}

	identType := reflect.TypeOf((*ast.Identifier)(nil)).Elem()
	if v.Type() == identType {
		f := v.FieldByName("Value")
		if f.IsValid() && f.Kind() == reflect.String {
			return strings.EqualFold(f.String(), name)
		}
		return false
	}

	switch v.Kind() {
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			if containsIdentifierWithName(v.Field(i), name) {
				return true
			}
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			if containsIdentifierWithName(v.Index(i), name) {
				return true
			}
		}
	}
	return false
}

func isFunctionDead(fn *ast.FunctionStatement, program *ast.Program) bool {
	name := fn.Name.Value
	for _, stmt := range program.Statements {
		// ignorer la définition même
		if f, ok := stmt.(*ast.FunctionStatement); ok && f == fn {
			continue
		}
		if isFunctionUsedInStatement(stmt, name) {
			return false
		}
	}
	return true
}

func isFunctionUsedInStatement(stmt ast.Statement, name string) bool {
	switch s := stmt.(type) {
	case *ast.ExpressionStatement:
		return isVariableUsedInExpression(s.Expression, name)
	case *ast.ReturnStatement:
		if s.ReturnValue != nil {
			return isVariableUsedInExpression(s.ReturnValue, name)
		}
	case *ast.LetStatement:
		if s.Value != nil {
			return isVariableUsedInExpression(s.Value, name)
		}
	case *ast.LetStatements:
		for _, v := range *s {
			if v.Value != nil && isVariableUsedInExpression(v.Value, name) {
				return true
			}
		}
	case *ast.BlockStatement:
		for _, st := range s.Statements {
			if isFunctionUsedInStatement(st, name) {
				return true
			}
		}
	case *ast.IfStatement:
		if s.Condition != nil && isVariableUsedInExpression(s.Condition, name) {
			return true
		}
		if s.Then != nil && isFunctionUsedInStatement(s.Then, name) {
			return true
		}
		if s.Else != nil && isFunctionUsedInStatement(s.Else, name) {
			return true
		}
	case *ast.ForStatement:
		if s.Init != nil && isFunctionUsedInStatement(s.Init, name) {
			return true
		}
		if s.Condition != nil && isVariableUsedInExpression(s.Condition, name) {
			return true
		}
		if s.Update != nil && isFunctionUsedInStatement(s.Update, name) {
			return true
		}
		if s.Body != nil && isFunctionUsedInStatement(s.Body, name) {
			return true
		}
	case *ast.WhileStatement:
		if s.Condition != nil && isVariableUsedInExpression(s.Condition, name) {
			return true
		}
		if s.Body != nil && isFunctionUsedInStatement(s.Body, name) {
			return true
		}
	case *ast.ForEachStatement:
		if s.Iterator != nil && isVariableUsedInExpression(s.Iterator, name) {
			return true
		}
		if s.Body != nil && isFunctionUsedInStatement(s.Body, name) {
			return true
		}
	case *ast.SwitchStatement:
		if s.Expression != nil && isVariableUsedInExpression(s.Expression, name) {
			return true
		}
		for _, c := range s.Cases {
			for _, e := range c.Expressions {
				if isVariableUsedInExpression(e, name) {
					return true
				}
			}
			if c.Body != nil && isFunctionUsedInStatement(c.Body, name) {
				return true
			}
		}
		if s.DefaultCase != nil && isFunctionUsedInStatement(s.DefaultCase, name) {
			return true
		}
	case *ast.FunctionStatement:
		// Si c'est une autre fonction, vérifier son corps (mais pas si même nom)
		if s.Name != nil && strings.EqualFold(s.Name.Value, name) {
			return false
		}
		if s.Body != nil && isFunctionUsedInStatement(s.Body, name) {
			return true
		}
	}
	return false
}

func isPureExpression(expr ast.Expression) bool {
	// Vérifier si l'expression n'a pas d'effet de bord
	switch expr.(type) {
	case *ast.IntegerLiteral, *ast.FloatLiteral, *ast.StringLiteral,
		*ast.BooleanLiteral, *ast.DurationLiteral, *ast.DateTimeLiteral:
		return true
	case *ast.InfixExpression:
		// Les opérations arithmétiques sont pures
		return true
	default:
		return false
	}
}

// LOOP OPTIMIZATION
func (lo *LoopOptimization) Name() string { return "LoopOptimization" }

func (lo *LoopOptimization) CanApply(program *ast.Program) bool {
	return hasLoops(program)
}

func (lo *LoopOptimization) Apply(program *ast.Program) *ast.Program {
	optimized := &ast.Program{
		ActionName: program.ActionName,
		Statements: []ast.Statement{},
	}

	for _, stmt := range program.Statements {
		optimized.Statements = append(optimized.Statements, optimizeLoopInStatement(stmt))
	}

	return optimized
}

func optimizeLoopInStatement(stmt ast.Statement) ast.Statement {
	switch s := stmt.(type) {
	case *ast.ForStatement:
		return optimizeForLoop(s)
	case *ast.WhileStatement:
		return optimizeWhileLoop(s)
	case *ast.ForEachStatement:
		return optimizeForEachLoop(s)
	default:
		return s
	}
}

func optimizeForLoop(stmt *ast.ForStatement) ast.Statement {
	// Loop-invariant code motion (conservative heuristic):
	// - Collect variables declared inside the loop body (let ...).
	// - Move out only pure statements (pure expressions or let with pure value)
	//   that do not reference any variable declared inside the loop.
	// - We only handle single LetStatement and ExpressionStatement moves (we
	//   avoid splitting grouped LetStatements for simplicity).

	if stmt == nil || stmt.Body == nil || len(stmt.Body.Statements) == 0 {
		return stmt
	}

	// Collect declared variable names inside the loop body (lower-cased).
	declared := map[string]struct{}{}
	for _, st := range stmt.Body.Statements {
		switch s := st.(type) {
		case *ast.LetStatement:
			if s.Name != nil {
				declared[strings.ToLower(s.Name.Value)] = struct{}{}
			}
		case *ast.LetStatements:
			for _, v := range *s {
				if v.Name != nil {
					declared[strings.ToLower(v.Name.Value)] = struct{}{}
				}
			}
		}
	}

	var moved []ast.Statement
	var remaining []ast.Statement

	for _, st := range stmt.Body.Statements {
		movedThis := false
		switch s := st.(type) {
		case *ast.ExpressionStatement:
			// Move only pure expressions that don't reference declared vars.
			if s.Expression != nil && isPureExpression(s.Expression) && !exprUsesAny(s.Expression, declared) {
				moved = append(moved, s)
				movedThis = true
			}
		case *ast.LetStatement:
			// Move only let statements with a pure value and no dependency on declared vars.
			if s.Value != nil && isPureExpression(s.Value) && !exprUsesAny(s.Value, declared) {
				moved = append(moved, s)
				movedThis = true
			}
		}
		if !movedThis {
			remaining = append(remaining, st)
		}
	}

	// If nothing moved, return original loop unchanged.
	if len(moved) == 0 {
		return stmt
	}
	IncrementLoopOptimization()
	// Construct the optimized loop with remaining body statements.
	optimizedLoop := &ast.ForStatement{
		Token:     stmt.Token,
		Init:      stmt.Init,
		Condition: stmt.Condition,
		Update:    stmt.Update,
		Body: &ast.BlockStatement{
			Token:      stmt.Body.Token,
			Statements: remaining,
		},
	}

	// Return a block that first runs the moved statements then the loop.
	return &ast.BlockStatement{
		Token:      stmt.Token,
		Statements: append(append([]ast.Statement{}, moved...), optimizedLoop),
	}
}

func optimizeWhileLoop(stmt *ast.WhileStatement) ast.Statement {
	// Loop-invariant code motion (conservative heuristic):
	// - Collect variables declared inside the loop body (let ...).
	// - Move out only pure statements (pure expressions or let with pure value)
	//   that do not reference any variable declared inside the loop.
	// - We only handle single LetStatement and ExpressionStatement moves (we
	//   avoid splitting grouped LetStatements for simplicity).

	if stmt == nil || stmt.Body == nil || len(stmt.Body.Statements) == 0 {
		return stmt
	}

	// Collect declared variable names inside the loop body (lower-cased).
	declared := map[string]struct{}{}
	for _, st := range stmt.Body.Statements {
		switch s := st.(type) {
		case *ast.LetStatement:
			if s.Name != nil {
				declared[strings.ToLower(s.Name.Value)] = struct{}{}
			}
		case *ast.LetStatements:
			for _, v := range *s {
				if v.Name != nil {
					declared[strings.ToLower(v.Name.Value)] = struct{}{}
				}
			}
		}
	}

	var moved []ast.Statement
	var remaining []ast.Statement

	for _, st := range stmt.Body.Statements {
		movedThis := false
		switch s := st.(type) {
		case *ast.ExpressionStatement:
			// Move only pure expressions that don't reference declared vars.
			if s.Expression != nil && isPureExpression(s.Expression) && !exprUsesAny(s.Expression, declared) {
				moved = append(moved, s)
				movedThis = true
			}
		case *ast.LetStatement:
			// Move only let statements with a pure value and no dependency on declared vars.
			if s.Value != nil && isPureExpression(s.Value) && !exprUsesAny(s.Value, declared) {
				moved = append(moved, s)
				movedThis = true
			}
		}
		if !movedThis {
			remaining = append(remaining, st)
		}
	}

	// If nothing moved, return original loop unchanged.
	if len(moved) == 0 {
		return stmt
	}
	IncrementLoopOptimization()
	// Construct the optimized loop with remaining body statements.
	optimizedLoop := &ast.WhileStatement{
		Token:     stmt.Token,
		Condition: stmt.Condition,
		Body: &ast.BlockStatement{
			Token:      stmt.Body.Token,
			Statements: remaining,
		},
	}

	// Return a block that first runs the moved statements then the loop.
	return &ast.BlockStatement{
		Token:      stmt.Token,
		Statements: append(append([]ast.Statement{}, moved...), optimizedLoop),
	}
}

func optimizeForEachLoop(stmt *ast.ForEachStatement) ast.Statement {
	// Loop-invariant code motion (conservative heuristic):
	// - Collect variables declared inside the loop body (let ...).
	// - Move out only pure statements (pure expressions or let with pure value)
	//   that do not reference any variable declared inside the loop.
	// - We only handle single LetStatement and ExpressionStatement moves (we
	//   avoid splitting grouped LetStatements for simplicity).

	if stmt == nil || stmt.Body == nil || len(stmt.Body.Statements) == 0 {
		return stmt
	}

	// Collect declared variable names inside the loop body (lower-cased).
	declared := map[string]struct{}{}
	for _, st := range stmt.Body.Statements {
		switch s := st.(type) {
		case *ast.LetStatement:
			if s.Name != nil {
				declared[strings.ToLower(s.Name.Value)] = struct{}{}
			}
		case *ast.LetStatements:
			for _, v := range *s {
				if v.Name != nil {
					declared[strings.ToLower(v.Name.Value)] = struct{}{}
				}
			}
		}
	}

	var moved []ast.Statement
	var remaining []ast.Statement

	for _, st := range stmt.Body.Statements {
		movedThis := false
		switch s := st.(type) {
		case *ast.ExpressionStatement:
			// Move only pure expressions that don't reference declared vars.
			if s.Expression != nil && isPureExpression(s.Expression) && !exprUsesAny(s.Expression, declared) {
				moved = append(moved, s)
				movedThis = true
			}
		case *ast.LetStatement:
			// Move only let statements with a pure value and no dependency on declared vars.
			if s.Value != nil && isPureExpression(s.Value) && !exprUsesAny(s.Value, declared) {
				moved = append(moved, s)
				movedThis = true
			}
		}
		if !movedThis {
			remaining = append(remaining, st)
		}
	}

	// If nothing moved, return original loop unchanged.
	if len(moved) == 0 {
		return stmt
	}
	IncrementLoopOptimization()
	// Construct the optimized loop with remaining body statements.
	optimizedLoop := &ast.ForEachStatement{
		Token:    stmt.Token,
		Variable: stmt.Variable,
		Iterator: stmt.Iterator,
		Body: &ast.BlockStatement{
			Token:      stmt.Body.Token,
			Statements: remaining,
		},
	}

	// Return a block that first runs the moved statements then the loop.
	return &ast.BlockStatement{
		Token:      stmt.Token,
		Statements: append(append([]ast.Statement{}, moved...), optimizedLoop),
	}
}

// exprUsesAny returns true if expr references any identifier name present in the set.
func exprUsesAny(expr ast.Expression, set map[string]struct{}) bool {
	if expr == nil {
		return false
	}
	switch e := expr.(type) {
	case *ast.Identifier:
		_, ok := set[strings.ToLower(e.Value)]
		return ok
	case *ast.InfixExpression:
		return exprUsesAny(e.Left, set) || exprUsesAny(e.Right, set)
	case *ast.PrefixExpression:
		return exprUsesAny(e.Right, set)
	case *ast.ArrayFunctionCall:
		if e.Function != nil && exprUsesAny(e.Function, set) {
			return true
		}
		if exprUsesAny(e.Array, set) {
			return true
		}
		for _, a := range e.Arguments {
			if exprUsesAny(a, set) {
				return true
			}
		}
		return false
	// Add other expression kinds that may contain identifiers as needed.
	default:
		return false
	}
}

// FUNCTION INLINING
func (fi *FunctionInlining) Name() string { return "FunctionInlining" }

func (fi *FunctionInlining) CanApply(program *ast.Program) bool {
	return hasSmallFunctions(program)
}

func (fi *FunctionInlining) Apply(program *ast.Program) *ast.Program {
	// Collecter les fonctions
	functions := make(map[string]*ast.FunctionStatement)
	var otherStatements []ast.Statement

	for _, stmt := range program.Statements {
		if fn, ok := stmt.(*ast.FunctionStatement); ok {
			functions[fn.Name.Value] = fn
		} else {
			otherStatements = append(otherStatements, stmt)
		}
	}

	// Appliquer l'inline
	optimized := &ast.Program{
		ActionName: program.ActionName,
		Statements: []ast.Statement{},
	}

	for _, stmt := range otherStatements {
		optimized.Statements = append(optimized.Statements, inlineFunctionsInStatement(stmt, functions))
	}

	// Garder seulement les fonctions non inlineables
	for _, fn := range functions {
		if !shouldInline(fn) {
			optimized.Statements = append(optimized.Statements, fn)
		}
	}

	return optimized
}

func shouldInline(fn *ast.FunctionStatement) bool {
	// Inline les petites fonctions
	bodySize := estimateFunctionSize(fn)
	return bodySize <= 5 // Seuil arbitraire
}

func estimateFunctionSize(fn *ast.FunctionStatement) int {
	size := 0
	for _, stmt := range fn.Body.Statements {
		size += estimateStatementSize(stmt)
	}
	return size
}

func estimateStatementSize(stmt ast.Statement) int {
	switch stmt.(type) {
	case *ast.ExpressionStatement:
		return 1
	case *ast.LetStatement:
		return 2
	case *ast.ReturnStatement:
		return 1
	case *ast.IfStatement:
		return 3
	default:
		return 1
	}
}

// Fonctions utilitaires pour détecter les opportunités d'optimisation
func hasConstantExpressions(program *ast.Program) bool {
	for _, stmt := range program.Statements {
		if containsConstantExpression(stmt) {
			return true
		}
	}
	return false
}

func containsConstantExpression(stmt ast.Statement) bool {
	switch s := stmt.(type) {
	case *ast.LetStatement:
		return s.Value != nil && isConstant(s.Value)
	case *ast.ExpressionStatement:
		return isConstant(s.Expression)
	case *ast.ReturnStatement:
		return s.ReturnValue != nil && isConstant(s.ReturnValue)
	default:
		return false
	}
}

func hasLoops(program *ast.Program) bool {
	for _, stmt := range program.Statements {
		if containsLoop(stmt) {
			return true
		}
	}
	return false
}

func containsLoop(stmt ast.Statement) bool {
	switch s := stmt.(type) {
	case *ast.ForStatement:
		return true
	case *ast.BlockStatement:
		for _, subStmt := range s.Statements {
			if containsLoop(subStmt) {
				return true
			}
		}
	}
	return false
}

func hasSmallFunctions(program *ast.Program) bool {
	for _, stmt := range program.Statements {
		if fn, ok := stmt.(*ast.FunctionStatement); ok {
			if shouldInline(fn) {
				return true
			}
		}
	}
	return false
}

// Implémentations des autres méthodes de folding
func foldExpressionStatement(stmt *ast.ExpressionStatement) *ast.ExpressionStatement {
	folded := foldExpression(stmt.Expression)
	if folded != stmt.Expression {
		return &ast.ExpressionStatement{
			Token:      stmt.Token,
			Expression: folded,
		}
	}
	return stmt
}

func foldReturnStatement(stmt *ast.ReturnStatement) *ast.ReturnStatement {
	if stmt.ReturnValue != nil {
		folded := foldExpression(stmt.ReturnValue)
		if folded != stmt.ReturnValue {
			return &ast.ReturnStatement{
				Token:       stmt.Token,
				ReturnValue: folded,
			}
		}
	}
	return stmt
}

func foldBlockStatement(stmt *ast.BlockStatement) *ast.BlockStatement {
	if stmt == nil {
		return nil
	}
	folded := &ast.BlockStatement{
		Token:      stmt.Token,
		Statements: []ast.Statement{},
	}

	for _, s := range stmt.Statements {
		folded.Statements = append(folded.Statements, foldConstantsInStatement(s))
	}

	return folded
}

func foldForStatement(stmt *ast.ForStatement) *ast.ForStatement {
	return &ast.ForStatement{
		Token:     stmt.Token,
		Init:      foldConstantsInStatement(stmt.Init),
		Condition: foldExpression(stmt.Condition),
		Update:    foldConstantsInStatement(stmt.Update),
		Body:      foldBlockStatement(stmt.Body),
	}
}

func foldSwitchStatement(stmt *ast.SwitchStatement) *ast.SwitchStatement {
	folded := &ast.SwitchStatement{
		Token:       stmt.Token,
		Expression:  foldExpression(stmt.Expression),
		Cases:       []*ast.SwitchCase{},
		DefaultCase: nil,
	}

	for _, c := range stmt.Cases {
		foldedCase := &ast.SwitchCase{
			Token:       c.Token,
			Expressions: []ast.Expression{},
			Body:        foldBlockStatement(c.Body),
		}

		for _, expr := range c.Expressions {
			foldedCase.Expressions = append(foldedCase.Expressions, foldExpression(expr))
		}

		folded.Cases = append(folded.Cases, foldedCase)
	}

	if stmt.DefaultCase != nil {
		folded.DefaultCase = foldBlockStatement(stmt.DefaultCase)
	}

	return folded
}

func foldPrefixExpression(expr *ast.PrefixExpression) ast.Expression {
	operand := foldExpression(expr.Right)

	if isConstant(operand) {
		switch expr.Operator {
		case "-":
			if intLit, ok := operand.(*ast.IntegerLiteral); ok {
				IncrementFolding()
				return &ast.IntegerLiteral{
					Token: expr.Token,
					Value: -intLit.Value,
				}
			}
			if floatLit, ok := operand.(*ast.FloatLiteral); ok {
				IncrementFolding()
				return &ast.FloatLiteral{
					Token: expr.Token,
					Value: -floatLit.Value,
				}
			}
		case "!":
			if boolLit, ok := operand.(*ast.BooleanLiteral); ok {
				IncrementFolding()
				return &ast.BooleanLiteral{
					Token: expr.Token,
					Value: !boolLit.Value,
				}
			}
		}
	}

	return &ast.PrefixExpression{
		Token:    expr.Token,
		Operator: expr.Operator,
		Right:    operand,
	}
}

func inlineFunctionsInStatement(stmt ast.Statement, functions map[string]*ast.FunctionStatement) ast.Statement {
	// TODO: Implémenter l'inlining des fonctions
	return stmt
}

func foldWhileStatement(stmt *ast.WhileStatement) *ast.WhileStatement {
	return &ast.WhileStatement{
		Token:     stmt.Token,
		Condition: foldExpression(stmt.Condition),
		Body:      foldBlockStatement(stmt.Body),
	}
}
