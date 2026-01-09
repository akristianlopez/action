package action

import (
	"context"
	"database/sql"

	"github.com/akristianlopez/action/ast"
	"github.com/akristianlopez/action/lexer"
	"github.com/akristianlopez/action/nsina"
	"github.com/akristianlopez/action/object"
	"github.com/akristianlopez/action/optimizer"
	"github.com/akristianlopez/action/parser"
	"github.com/akristianlopez/action/semantic"
)

type Action struct {
	ctx      context.Context
	db       *sql.DB
	dbname   string
	error    []string
	Warnings []string
}

func NewAction(ctx context.Context, db *sql.DB, dbname string) *Action {
	return &Action{ctx: ctx, db: db, dbname: dbname, error: make([]string, 0)}
}
func (action *Action) Interprete(src string, canHandle func(table, field, operation string) (bool, string),
	hasFilter func(table string) bool, getFilter func(table, newName string) (ast.Expression, bool),
	params map[string]object.Object) (object.Object, []string) {
	lex := lexer.New(src)
	p := parser.New(lex)
	act := p.ParseAction()
	if p.Errors() != nil && len(p.Errors()) != 0 {
		for _, msg := range p.Errors() {
			action.error = append(action.error, msg.String())
		}
		return object.NULL, action.error
	}
	analyzer := semantic.NewSemanticAnalyzer(action.ctx, action.db, canHandle)
	errors := analyzer.Analyze(act)
	if len(analyzer.Warnings) > 0 {
		action.Warnings = append(action.Warnings, analyzer.Warnings...)
	}
	if len(errors) > 0 {
		action.error = append(action.error, errors...)
		return object.NULL, action.error
	}
	opt := optimizer.NewOptimizer()
	optimizedProgram := opt.Optimize(act)
	// opt.Optimize(action)
	if len(opt.Warnings) > 0 {
		action.Warnings = append(action.Warnings, opt.Warnings...)
	}
	env := object.NewEnvironment(action.ctx, action.db, hasFilter, getFilter, action.dbname, params)
	result := nsina.Eval(optimizedProgram, env)
	return result, action.error
}
func (action *Action) Expression(src, table, newName string, canHandle func(table, field, operation string) (bool, string)) (ast.Expression, []string) {
	lex := lexer.New(src)
	p := parser.New(lex)
	act := p.ParseExpression()
	if p.Errors() != nil && len(p.Errors()) != 0 {
		for _, msg := range p.Errors() {
			action.error = append(action.error, msg.String())
		}
		return nil, action.error
	}
	analyzer := semantic.NewSemanticAnalyzer(action.ctx, action.db, canHandle)
	analyzer.AnalyzeExpression(table, newName, act)
	if len(analyzer.Warnings) > 0 {
		action.Warnings = append(action.Warnings, analyzer.Warnings...)
	}
	if len(analyzer.Errors) > 0 {
		action.error = append(action.error, analyzer.Errors...)
		return nil, action.error
	}
	return act, action.error
}
