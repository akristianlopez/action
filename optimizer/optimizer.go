package optimizer

import (
	"github.com/akristianlopez/action/ast"
	// "github.com/akristianlopez/action/token"
)

type Optimizer struct {
	Optimizations []Optimization
	Stats         OptimizationStats
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
		Stats: OptimizationStats{},
	}
}

func (o *Optimizer) Optimize(program *ast.Program) *ast.Program {
	optimized := program

	// Appliquer les optimisations en plusieurs passes
	for i := 0; i < 10; i++ { // Maximum 10 passes
		changed := false
		for _, opt := range o.Optimizations {
			if opt.CanApply(optimized) {
				optimized = opt.Apply(optimized)
				changed = true

				// Mettre à jour les statistiques
				switch opt.(type) {
				case *ConstantFolding:
					o.Stats.ConstantFolds++
				case *DeadCodeElimination:
					o.Stats.DeadCodeRemovals++
				case *FunctionInlining:
					o.Stats.InlineExpansions++
				case *LoopOptimization:
					o.Stats.LoopOptimizations++
				}
			}
		}

		if !changed {
			break
		}
	}

	return optimized
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
	case *ast.ReturnStatement:
		return foldReturnStatement(s)
	case *ast.BlockStatement:
		return foldBlockStatement(s)
	case *ast.ForStatement:
		return foldForStatement(s)
	case *ast.SwitchStatement:
		return foldSwitchStatement(s)
	default:
		return s
	}
}

func foldLetStatement(stmt *ast.LetStatement) *ast.LetStatement {
	if stmt.Value != nil {
		folded := foldExpression(stmt.Value)
		if folded != stmt.Value {
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
					return &ast.IntegerLiteral{
						Token: expr.Token,
						Value: l.Value + r.Value,
					}
				}
			}
			if l, ok := left.(*ast.FloatLiteral); ok {
				if r, ok := right.(*ast.FloatLiteral); ok {
					return &ast.FloatLiteral{
						Token: expr.Token,
						Value: l.Value + r.Value,
					}
				}
			}

		case "-":
			if l, ok := left.(*ast.IntegerLiteral); ok {
				if r, ok := right.(*ast.IntegerLiteral); ok {
					return &ast.IntegerLiteral{
						Token: expr.Token,
						Value: l.Value - r.Value,
					}
				}
			}

		case "*":
			if l, ok := left.(*ast.IntegerLiteral); ok {
				if r, ok := right.(*ast.IntegerLiteral); ok {
					return &ast.IntegerLiteral{
						Token: expr.Token,
						Value: l.Value * r.Value,
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
	optimized := &ast.Program{
		ActionName: program.ActionName,
		Statements: []ast.Statement{},
	}

	for _, stmt := range program.Statements {
		if !isDeadCode(stmt) {
			optimized.Statements = append(optimized.Statements, stmt)
		}
	}

	return optimized
}

func isDeadCode(stmt ast.Statement) bool {
	// Identifier le code mort (variables non utilisées, etc.)
	switch s := stmt.(type) {
	case *ast.LetStatement:
		// TODO: Vérifier si la variable est utilisée
		return false
	case *ast.ExpressionStatement:
		// Les expressions sans effet de bord peuvent être mortes
		return isPureExpression(s.Expression)
	default:
		return false
	}
}

func isPureExpression(expr ast.Expression) bool {
	// Vérifier si l'expression n'a pas d'effet de bord
	switch expr.(type) {
	case *ast.IntegerLiteral, *ast.FloatLiteral, *ast.StringLiteral, *ast.BooleanLiteral:
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
	default:
		return s
	}
}

func optimizeForLoop(stmt *ast.ForStatement) *ast.ForStatement {
	// Optimization: déplacer les expressions invariantes hors de la boucle
	optimized := &ast.ForStatement{
		Token:     stmt.Token,
		Init:      stmt.Init,
		Condition: stmt.Condition,
		Update:    stmt.Update,
		Body:      &ast.BlockStatement{Token: stmt.Body.Token},
	}

	// TODO: Implémenter loop-invariant code motion
	optimized.Body.Statements = stmt.Body.Statements

	return optimized
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
				return &ast.IntegerLiteral{
					Token: expr.Token,
					Value: -intLit.Value,
				}
			}
			if floatLit, ok := operand.(*ast.FloatLiteral); ok {
				return &ast.FloatLiteral{
					Token: expr.Token,
					Value: -floatLit.Value,
				}
			}
		case "!":
			if boolLit, ok := operand.(*ast.BooleanLiteral); ok {
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
