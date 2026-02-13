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
	Apply(*ast.Action) *ast.Action
	CanApply(*ast.Action) bool
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
var IncrementInlineExpansion func()

func (o *Optimizer) IncrementConstantFolding() {
	o.Stats.ConstantFolds++
}

func (o *Optimizer) IncrementLoopOptimization() {
	o.Stats.LoopOptimizations++
}

func (o *Optimizer) IncrementInlineExpansion() {
	o.Stats.InlineExpansions++
}

func (o *Optimizer) Optimize(program *ast.Action) *ast.Action {
	optimized := program
	Warnings = o.addWarning
	IncrementFolding = o.IncrementConstantFolding
	IncrementLoopOptimization = o.IncrementLoopOptimization
	IncrementInlineExpansion = o.IncrementInlineExpansion

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
						o.Stats.DeadCodeRemovals += oSize - len(optimized.Statements)
					}
				case *FunctionInlining:
					// o.Stats.InlineExpansions++
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

func (cf *ConstantFolding) CanApply(program *ast.Action) bool {
	return hasConstantExpressions(program)
}

func (cf *ConstantFolding) Apply(program *ast.Action) *ast.Action {
	optimized := &ast.Action{
		ActionName: program.ActionName,
		Statements: []ast.Statement{},
	}

	for _, stmt := range program.Statements {
		// switch s := stmt.(type) {
		// case *ast.LetStatements:
		// 	for _, v := range *s {
		// 		optimized.Statements = append(optimized.Statements, foldConstantsInStatement(&v))
		// 	}
		// default:
		// 	optimized.Statements = append(optimized.Statements, foldConstantsInStatement(stmt))
		// }
		optimized.Statements = append(optimized.Statements, foldConstantsInStatement(stmt))
	}
	return optimized
}

func foldConstantsInStatement(stmt ast.Statement) ast.Statement {
	switch s := stmt.(type) {
	case *ast.LetStatement:
		return foldLetStatement(s)
	case *ast.ExpressionStatement:
		return foldExpressionStatement(s)
	case *ast.AssignmentStatement:
		return foldAssignmentStatement(s)
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
	case *ast.CatchStatement:
		return foldCatchStatement(s)
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

func foldCatchStatement(stmt *ast.CatchStatement) ast.Statement {
	return &ast.CatchStatement{
		Token:      stmt.Token,
		Statements: foldBlockStatement(stmt.Statements),
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
	case *ast.BetweenExpression:
		return foldBetweenExpression(e)
	case *ast.LikeExpression:
		return foldLikeExpression(e)
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

func (dce *DeadCodeElimination) CanApply(program *ast.Action) bool {
	return true // Toujours applicable
}

func (dce *DeadCodeElimination) Apply(program *ast.Action) *ast.Action {
	//Eliminating of the unused let statements
	optimized := &ast.Action{
		ActionName: program.ActionName,
		Statements: []ast.Statement{},
	}
	for _, stmt := range program.Statements {
		switch s := stmt.(type) {
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
		case *ast.CatchStatement:
			if !isDeadCode(s, program) {
				optimized.Statements = append(optimized.Statements, s)
				continue
			}
			Warnings("Dead code eliminated: Catch statement at Line:%d, column:%d has empty body.",
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

func isUsed(name string, actions *ast.Action) bool {
	// Vérifier si la variable est utilisée dans le programme
	for _, stmt := range actions.Statements {
		if isFunctionUsedInStatement(stmt, name) {
			return true
		}
		// switch s := stmt.(type) {
		// case *ast.LetStatement:
		// 	// if s.Name.Value == name {
		// 	// 	return true // La variable est définie ici
		// 	// }
		// 	if !strings.EqualFold(name, s.Name.Value) && s.Value != nil && isVariableUsedInExpression(s.Value, name) {
		// 		return true // La variable est utilisée dans une expression
		// 	}
		// case *ast.LetStatements:
		// 	for _, v := range *s {
		// 		// if v.Name.Value == name {
		// 		// 	return true // La variable est définie ici
		// 		// }
		// 		if !strings.EqualFold(name, v.Name.Value) && v.Value != nil && isVariableUsedInExpression(v.Value, name) {
		// 			return true // La variable est définie ici
		// 		}
		// 	}
		// case *ast.ExpressionStatement:
		// 	if isVariableUsedInExpression(s.Expression, name) {
		// 		return true // La variable est utilisée dans une expression
		// 	}
		// case *ast.ReturnStatement:
		// 	if isVariableUsedInExpression(s.ReturnValue, name) {
		// 		return true // La variable est utilisée dans une valeur de retour
		// 	}
		// default:
		// 	if isFunctionUsedInStatement(stmt, name) {
		// 		return true
		// 	}
		// }
	}
	return false
}

func isVariableUsedInExpression(expr ast.Expression, name string) bool {
	switch e := expr.(type) {
	case *ast.Identifier:
		return strings.EqualFold(e.Value, name)
	case *ast.TypeMember:
		if t, o := e.Left.(*ast.Identifier); o && t != nil {
			return strings.EqualFold(t.Value, name)
		}
		return false
	case *ast.TypeExternalCall:
		return isVariableUsedInExpression(e.Name, name) || isVariableUsedInExpression(e.Action, name)
	case *ast.InfixExpression:
		return isVariableUsedInExpression(e.Left, name) || isVariableUsedInExpression(e.Right, name)
	case *ast.IndexExpression:
		return isVariableUsedInExpression(e.Left, name) || isVariableUsedInExpression(e.Index, name)
	case *ast.PrefixExpression:
		return isVariableUsedInExpression(e.Right, name)
	case *ast.BetweenExpression:
		return isVariableUsedInExpression(e.Base, name) || isVariableUsedInExpression(e.Left, name) ||
			isVariableUsedInExpression(e.Right, name)
	case *ast.LikeExpression:
		return isVariableUsedInExpression(e.Left, name) || isVariableUsedInExpression(e.Right, name)
	case *ast.AssignmentStatement:
		return isVariableUsedInExpression(e.Variable, name) || isVariableUsedInExpression(e.Value, name)
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

func isDeadCode(stmt ast.Statement, actions *ast.Action) bool {
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
		return false //isPureExpression(s.Expression)
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
	case *ast.CatchStatement:
		// Si les deux branches sont mortes, le if est mort
		return s.Statements == nil || isDeadCode(s.Statements, actions)
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
		allDead := true
		if len(s.Statements) == 0 {
			return allDead
		}
		for _, st := range s.Statements {
			allDead = allDead && isDeadCode(st, actions)
		}
		return allDead
	case *ast.ReturnStatement:
		// Un return est mort s'il n'y a pas de valeur de retour ou si la valeur est pure
		if s.ReturnValue == nil {
			return true
		}
		return false //isPureExpression(s.ReturnValue)
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

func isDefinedStructDead(s *ast.StructStatement, actions *ast.Action) bool {
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

func isStructNameUsedAsType(node ast.Statement, name string) bool {
	if node == nil {
		return false
	}
	switch st := node.(type) {
	case *ast.LetStatement:
		return scanForTypeLikeField(st, name)
	case *ast.BlockStatement:
		if st == nil {
			return false
		}
		if len(st.Statements) == 0 {
			return false
		}
		for _, v := range st.Statements {
			if isStructNameUsedAsType(v, name) {
				return true
			}
		}
		return false
	case *ast.IfStatement:
		return isStructNameUsedAsType(st.Then, name) || isStructNameUsedAsType(st.Else, name)
	case *ast.CatchStatement:
		return isStructNameUsedAsType(st.Statements, name)
	case *ast.ForEachStatement:
		return isStructNameUsedAsType(st.Body, name)
	case *ast.ForStatement:
		return isStructNameUsedAsType(st.Init, name) || isStructNameUsedAsType(st.Body, name) ||
			isStructNameUsedAsType(st.Update, name)
	case *ast.WhileStatement:
		return isStructNameUsedAsType(st.Body, name)
	case *ast.FunctionStatement:
		if st == nil {
			return false
		}
		for _, arg := range st.Parameters {
			if scanForTypeLikeField(arg, name) {
				return true
			}
		}
		return isStructNameUsedAsType(st.Body, name)
	case *ast.SwitchStatement:
		if st == nil {
			return false
		}
		for _, stm := range st.Cases {
			if isStructNameUsedAsType(stm.Body, name) {
				return true
			}
		}
		return isStructNameUsedAsType(st.DefaultCase, name)
	case *ast.StructStatement:
		if st == nil {
			return false
		}
		for _, stm := range st.Fields {
			if isTypeLike(stm.Type, name) {
				return true
			}
		}
	}
	return false
}

func isTypeLike(let *ast.TypeAnnotation, name string) bool {
	if let == nil {
		return false
	}
	if let.ArrayType == nil {
		return strings.EqualFold(let.Type, name)
	}
	if let.ArrayType.ElementType != nil {
		return strings.EqualFold(let.ArrayType.ElementType.Type, name)
	}
	return false
}

func scanForTypeLikeField(node ast.Node, name string) bool {
	if node == nil {
		return false
	}
	switch let := node.(type) {
	case *ast.LetStatement:
		return isTypeLike(let.Type, name)
	case *ast.FunctionParameter:
		return isTypeLike(let.Type, name)
	}
	return false
}

func isFunctionDead(fn *ast.FunctionStatement, program *ast.Action) bool {
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
	case *ast.AssignmentStatement:
		return isVariableUsedInExpression(s.Variable, name) ||
			isVariableUsedInExpression(s.Value, name)
	case *ast.TypeMember:
		if t, o := s.Left.(*ast.Identifier); o && t != nil {
			return strings.EqualFold(t.Value, name)
		}
		return false
	case *ast.ReturnStatement:
		if s.ReturnValue != nil {
			return isVariableUsedInExpression(s.ReturnValue, name)
		}
	case *ast.LetStatement:
		if s.Value != nil {
			return isVariableUsedInExpression(s.Value, name)
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
	case *ast.CatchStatement:
		if s.Statements != nil && isFunctionUsedInStatement(s.Statements, name) {
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

func (lo *LoopOptimization) CanApply(program *ast.Action) bool {
	return hasLoops(program)
}

func (lo *LoopOptimization) Apply(program *ast.Action) *ast.Action {
	optimized := &ast.Action{
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
		if v, ok := st.(*ast.LetStatement); ok {
			if v.Name != nil {
				declared[strings.ToLower(v.Name.Value)] = struct{}{}
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
		if s, ok := st.(*ast.LetStatement); s != nil && ok {
			if s.Name != nil {
				declared[strings.ToLower(s.Name.Value)] = struct{}{}
			}
		}
		// switch s := st.(type) {
		// case *ast.LetStatement:
		// 	if s.Name != nil {
		// 		declared[strings.ToLower(s.Name.Value)] = struct{}{}
		// 	}
		// case *ast.LetStatements:
		// 	for _, v := range *s {
		// 		if v.Name != nil {
		// 			declared[strings.ToLower(v.Name.Value)] = struct{}{}
		// 		}
		// 	}
		// }
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
		if s, ok := st.(*ast.LetStatement); s != nil && ok {
			if s.Name != nil {
				declared[strings.ToLower(s.Name.Value)] = struct{}{}
			}
		}
		// switch s := st.(type) {
		// case *ast.LetStatement:
		// 	if s.Name != nil {
		// 		declared[strings.ToLower(s.Name.Value)] = struct{}{}
		// 	}
		// case *ast.LetStatements:
		// 	for _, v := range *s {
		// 		if v.Name != nil {
		// 			declared[strings.ToLower(v.Name.Value)] = struct{}{}
		// 		}
		// 	}
		// }
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

func (fi *FunctionInlining) CanApply(program *ast.Action) bool {
	return hasSmallFunctions(program)
}

func (fi *FunctionInlining) Apply(program *ast.Action) *ast.Action {
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
	optimized := &ast.Action{
		ActionName: program.ActionName,
		Statements: []ast.Statement{},
	}

	for _, stmt := range otherStatements {
		optimized.Statements = append(optimized.Statements, inlineFunctionsInStatement(stmt, functions))
	}

	// Garder seulement les fonctions non inlineables
	arr := make([]ast.Statement, 0)
	for _, fn := range functions {
		if !shouldInline(fn) || len(fn.Parameters) > 0 {
			arr = append(arr, fn)
		}
	}
	if len(arr) > 0 {
		optimized.Statements = append(arr, optimized.Statements...)
	}
	return optimized
}

func shouldInline(fn *ast.FunctionStatement) bool {
	// Inline les petites fonctions
	bodySize := estimateFunctionSize(fn)
	return bodySize == 1 // Seuil arbitraire
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
		return 3
	case *ast.LetStatement:
		return 2
	case *ast.ReturnStatement:
		return 1
	case *ast.IfStatement, *ast.CatchStatement:
		return 4
	case *ast.ForStatement, *ast.WhileStatement, *ast.ForEachStatement:
		return 5
	case *ast.SQLAlterObjectStatement, *ast.SQLCreateIndexStatement,
		*ast.SQLInsertStatement, *ast.SQLDropObjectStatement, *ast.SQLDeleteStatement,
		*ast.SQLTruncateStatement, *ast.SQLUpdateStatement, *ast.SQLSelectStatement,
		*ast.SQLWithStatement:
		return 10
	default:
		return 3
	}
}

// Fonctions utilitaires pour détecter les opportunités d'optimisation
func hasConstantExpressions(program *ast.Action) bool {
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

func hasLoops(program *ast.Action) bool {
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

func hasSmallFunctions(program *ast.Action) bool {
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

// Implémentations des autres méthodes de folding
func foldAssignmentStatement(stmt *ast.AssignmentStatement) *ast.AssignmentStatement {
	folded := foldExpression(stmt.Value)
	if folded != stmt.Value {
		return &ast.AssignmentStatement{
			Token:    stmt.Token,
			Variable: stmt.Variable,
			Value:    folded,
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

func floatVal(e ast.Expression) float64 {
	if e == nil {
		return 0.0
	}
	switch s := e.(type) {
	case *ast.IntegerLiteral:
		return float64(s.Value)
	case *ast.FloatLiteral:
		return float64(s.Value)
	}
	return 0.0
}

// NOTE: minimal stub implementations to satisfy references from foldExpression.
// These are conservative no-ops for now and can be expanded to fold inner
// expressions when the exact AST field names/structure are known.

func foldBetweenExpression(expr *ast.BetweenExpression) ast.Expression {
	if expr == nil {
		return nil
	}
	if expr.Base != nil && expr.Left != nil && expr.Right != nil &&
		isConstant(expr.Base) && isConstant(expr.Left) && isConstant(expr.Right) {
		// Currently return as-is; implement inner folding later if needed.

		switch expr.Base.(type) {
		case *ast.IntegerLiteral, *ast.FloatLiteral:
			val := floatVal(expr.Base)
			min := floatVal(expr.Left)
			max := floatVal(expr.Right)
			value := (val >= min && val <= max)
			if expr.Not {
				value = !value
			}
			IncrementFolding()
			return &ast.BooleanLiteral{
				Token: expr.Token,
				Value: value,
			}
		default:
			// return expr
		}
	}
	return expr
}

func foldLikeExpression(expr *ast.LikeExpression) ast.Expression {
	if expr == nil {
		return nil
	}
	if expr.Left != nil && expr.Right != nil &&
		isConstant(expr.Left) && isConstant(expr.Right) {
		// Currently return as-is; implement inner folding later if needed.

		switch expr.Left.(type) {
		case *ast.StringLiteral:
			val := expr.Left.(*ast.StringLiteral)
			pattern := expr.Right.(*ast.StringLiteral)
			id := strings.Index(pattern.Value, "*")
			res := false
			if len(val.Value) > id && id > 0 {
				res = strings.EqualFold(val.Value[0:id], pattern.Value[0:id])
			}
			if expr.Not {
				res = !res
			}
			IncrementFolding()
			return &ast.BooleanLiteral{
				Token: expr.Token,
				Value: res,
			}
		default:
			// return expr
		}
	}
	// Currently return as-is; implement inner folding later if needed.
	return expr
}

func inlineFunctionsInStatement(stmt ast.Statement, functions map[string]*ast.FunctionStatement) ast.Statement {
	if stmt == nil {
		return stmt
	}
	switch s := stmt.(type) {
	case *ast.ExpressionStatement:
		newExpr := inlineFunctionsInExpression(s.Expression, functions)
		if newExpr != s.Expression {
			return &ast.ExpressionStatement{Token: s.Token, Expression: newExpr}
		}
		return s
	case *ast.LetStatement:
		if s.Value != nil {
			newVal := inlineFunctionsInExpression(s.Value, functions)
			if newVal != s.Value {
				return &ast.LetStatement{
					Token: s.Token,
					Name:  s.Name,
					Type:  s.Type,
					Value: newVal,
				}
			}
		}
		return s
	// case *ast.LetStatements:
	// 	changed := false
	// 	out := make([]ast.Statement, 0, len(*s))
	// 	for _, v := range *s {
	// 		ls := v
	// 		if ls.Value != nil {
	// 			newVal := inlineFunctionsInExpression(ls.Value, functions)
	// 			if newVal != ls.Value {
	// 				ls = ast.LetStatement{
	// 					Token: ls.Token,
	// 					Name:  ls.Name,
	// 					Type:  ls.Type,
	// 					Value: newVal,
	// 				}
	// 				changed = true
	// 			}
	// 		}
	// 		out = append(out, &ls)
	// 	}
	// 	if changed {
	// 		// convert back to []*ast.LetStatement if that's the original shape;
	// 		// we use []ast.Statement which is acceptable to the optimizer pipeline
	// 		return &ast.BlockStatement{
	// 			Token:      (*s)[0].Token,
	// 			Statements: out,
	// 		}
	// 	}
	// 	return s
	case *ast.AssignmentStatement:
		target := inlineFunctionsInExpression(s.Variable, functions)
		value := inlineFunctionsInExpression(s.Value, functions)
		if s.Variable != target || value != s.Value {
			return &ast.AssignmentStatement{Token: s.Token, Variable: target, Value: value}
		}
		return s
	case *ast.ReturnStatement:
		if s.ReturnValue != nil {
			newVal := inlineFunctionsInExpression(s.ReturnValue, functions)
			if newVal != s.ReturnValue {
				return &ast.ReturnStatement{Token: s.Token, ReturnValue: newVal}
			}
		}
		return s
	case *ast.BlockStatement:
		newBlock := &ast.BlockStatement{Token: s.Token, Statements: []ast.Statement{}}
		changed := false
		for _, st := range s.Statements {
			ns := inlineFunctionsInStatement(st, functions)
			newBlock.Statements = append(newBlock.Statements, ns)
			if ns != st {
				changed = true
			}
		}
		if changed {
			return newBlock
		}
		return s
	case *ast.IfStatement:
		cond := inlineFunctionsInExpression(s.Condition, functions)
		then := foldBlockStatement(s.Then)
		elseBlk := foldBlockStatement(s.Else)
		// inline inside then/else
		then = inlineBlockStatements(then, functions)
		elseBlk = inlineBlockStatements(elseBlk, functions)
		if cond != s.Condition || then != s.Then || elseBlk != s.Else {
			return &ast.IfStatement{
				Token:     s.Token,
				Condition: cond,
				Then:      then,
				Else:      elseBlk,
			}
		}
		return s
	case *ast.CatchStatement:
		then := foldBlockStatement(s.Statements)
		// inline inside then/else
		then = inlineBlockStatements(then, functions)
		if then != s.Statements {
			return &ast.CatchStatement{
				Token:      s.Token,
				Statements: then,
			}
		}
		return s
	case *ast.ForStatement:
		init := inlineFunctionsInStatement(s.Init, functions)
		cond := inlineFunctionsInExpression(s.Condition, functions)
		update := inlineFunctionsInStatement(s.Update, functions)
		body := inlineBlockStatements(s.Body, functions)
		if init != s.Init || cond != s.Condition || update != s.Update || body != s.Body {
			return &ast.ForStatement{
				Token:     s.Token,
				Init:      init,
				Condition: cond,
				Update:    update,
				Body:      body,
			}
		}
		return s
	case *ast.WhileStatement:
		cond := inlineFunctionsInExpression(s.Condition, functions)
		body := inlineBlockStatements(s.Body, functions)
		if cond != s.Condition || body != s.Body {
			return &ast.WhileStatement{
				Token:     s.Token,
				Condition: cond,
				Body:      body,
			}
		}
		return s
	case *ast.ForEachStatement:
		iter := inlineFunctionsInExpression(s.Iterator, functions)
		body := inlineBlockStatements(s.Body, functions)
		if iter != s.Iterator || body != s.Body {
			return &ast.ForEachStatement{
				Token:    s.Token,
				Variable: s.Variable,
				Iterator: iter,
				Body:     body,
			}
		}
		return s
	case *ast.SwitchStatement:
		expr := inlineFunctionsInExpression(s.Expression, functions)
		changed := expr != s.Expression
		newCases := []*ast.SwitchCase{}
		for _, c := range s.Cases {
			newBody := inlineBlockStatements(c.Body, functions)
			newExprs := []ast.Expression{}
			for _, e := range c.Expressions {
				ne := inlineFunctionsInExpression(e, functions)
				newExprs = append(newExprs, ne)
				if ne != e {
					changed = true
				}
			}
			newCases = append(newCases, &ast.SwitchCase{
				Token:       c.Token,
				Expressions: newExprs,
				Body:        newBody,
			})
			if newBody != c.Body {
				changed = true
			}
		}
		def := s.DefaultCase
		if def != nil {
			newDef := inlineBlockStatements(def, functions)
			if newDef != def {
				changed = true
			}
			def = newDef
		}
		if changed {
			return &ast.SwitchStatement{
				Token:       s.Token,
				Expression:  expr,
				Cases:       newCases,
				DefaultCase: def,
			}
		}
		return s
	default:
		return s
	}
}

func inlineBlockStatements(b *ast.BlockStatement, functions map[string]*ast.FunctionStatement) *ast.BlockStatement {
	if b == nil {
		return nil
	}
	changed := false
	newBlk := &ast.BlockStatement{Token: b.Token, Statements: []ast.Statement{}}
	for _, st := range b.Statements {
		ns := inlineFunctionsInStatement(st, functions)
		newBlk.Statements = append(newBlk.Statements, ns)
		if ns != st {
			changed = true
		}
	}
	if changed {
		return newBlk
	}
	return b
}

func inlineFunctionsInExpression(expr ast.Expression, functions map[string]*ast.FunctionStatement) ast.Expression {
	if expr == nil {
		return nil
	}
	switch e := expr.(type) {
	case *ast.Identifier:
		return e
	case *ast.InfixExpression:
		l := inlineFunctionsInExpression(e.Left, functions)
		r := inlineFunctionsInExpression(e.Right, functions)
		if l != e.Left || r != e.Right {
			return &ast.InfixExpression{Token: e.Token, Left: l, Operator: e.Operator, Right: r}
		}
		return e
	case *ast.PrefixExpression:
		r := inlineFunctionsInExpression(e.Right, functions)
		if r != e.Right {
			return &ast.PrefixExpression{Token: e.Token, Operator: e.Operator, Right: r}
		}
		return e
	case *ast.BetweenExpression:
		base := inlineFunctionsInExpression(e.Base, functions)
		l := inlineFunctionsInExpression(e.Left, functions)
		r := inlineFunctionsInExpression(e.Right, functions)
		if base != e.Base || l != e.Left || r != e.Right {
			return &ast.BetweenExpression{
				Token: e.Token, Base: base, Left: l, Right: r, Not: e.Not,
			}
		}
		return e
	case *ast.LikeExpression:
		l := inlineFunctionsInExpression(e.Left, functions)
		r := inlineFunctionsInExpression(e.Right, functions)
		if l != e.Left || r != e.Right {
			return &ast.LikeExpression{Token: e.Token, Left: l, Right: r, Not: e.Not}
		}
		return e
	case *ast.AssignmentStatement:
		target := inlineFunctionsInExpression(e.Variable, functions)
		value := inlineFunctionsInExpression(e.Value, functions)
		if e.Variable != target || value != e.Value {
			return &ast.AssignmentStatement{Token: e.Token, Variable: target, Value: value}
		}
		return e
	case *ast.ArrayFunctionCall:
		// recurse into function expression first

		fnExpr := inlineFunctionsInExpression(e.Function, functions)
		argsChanged := false
		newArgs := make([]ast.Expression, 0, len(e.Arguments))
		for _, a := range e.Arguments {
			na := inlineFunctionsInExpression(a, functions)
			newArgs = append(newArgs, na)
			if na != a {
				argsChanged = true
			}
		}
		// attempt inlining only when function is a simple identifier and the target function is present
		if id, ok := fnExpr.(*ast.Identifier); ok {
			if fn, found := functions[id.Value]; found && shouldInline(fn) {
				// conservative: inline only functions without parameters and with a single return statement
				rv := reflect.ValueOf(fn.Parameters)
				paramCount := 0
				if rv.IsValid() && rv.Kind() == reflect.Slice {
					paramCount = rv.Len()
				}
				if paramCount == 0 && fn.Body != nil && len(fn.Body.Statements) == 1 {
					if ret, ok := fn.Body.Statements[0].(*ast.ReturnStatement); ok && ret.ReturnValue != nil {
						// inline by returning a copy of the return expression (and recursively inline inside it)
						IncrementInlineExpansion()
						inlined := inlineFunctionsInExpression(ret.ReturnValue, functions)
						return inlined
					}
				}
			}
		}
		// otherwise rebuild node if any child changed
		if fnExpr != e.Function || argsChanged || (e.Array != nil && inlineFunctionsInExpression(e.Array, functions) != e.Array) {
			IncrementInlineExpansion()
			return &ast.ArrayFunctionCall{
				Token:     e.Token,
				Function:  fnExpr.(*ast.Identifier),
				Array:     inlineFunctionsInExpression(e.Array, functions),
				Arguments: newArgs,
			}
		}
		return e
	default:
		// unknown expression types: return as-is
		return e
	}
}

func foldWhileStatement(stmt *ast.WhileStatement) *ast.WhileStatement {
	return &ast.WhileStatement{
		Token:     stmt.Token,
		Condition: foldExpression(stmt.Condition),
		Body:      foldBlockStatement(stmt.Body),
	}
}
