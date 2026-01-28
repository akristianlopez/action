package action

import (
	"context"
	"database/sql"
	"strings"

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
	warnings []string
}

func NewAction(ctx context.Context, db *sql.DB, dbname string) *Action {
	return &Action{ctx: ctx, db: db, dbname: dbname, error: make([]string, 0)}
}
func (action *Action) Interpret(src string, canHandle func(table, field, operation string) (bool, string),
	hasFilter func(table string) bool, getFilter func(table, newName string) (ast.Expression, bool),
	params map[string]object.Object, disableUpdate, disabledDDL bool,
	serviceExists func(serviceName string) bool,
	signature func(serviceName, methodName string) ([]*ast.StructField, *ast.TypeAnnotation, error),
	external func(ctx context.Context, srv, name string, args map[string]object.Object) (object.Object, bool)) (object.Object, []string) {
	lex := lexer.New(src)
	p := parser.New(lex)
	act := p.ParseAction()
	if p.Errors() != nil && len(p.Errors()) != 0 {
		for _, msg := range p.Errors() {
			action.error = append(action.error, msg.String())
		}
		return object.NULL, action.error
	}
	analyzer := semantic.NewSemanticAnalyzer(action.ctx, action.db, canHandle, serviceExists, signature)
	errors := analyzer.Analyze(act)
	if len(analyzer.Warnings) > 0 {
		action.setWarnings(append(action.Warnings(), analyzer.Warnings...))
	}
	if len(errors) > 0 {
		action.error = append(action.error, errors...)
		return object.NULL, action.error
	}
	opt := optimizer.NewOptimizer()
	optimizedProgram := opt.Optimize(act)
	// opt.Optimize(action)
	if len(opt.Warnings) > 0 {
		action.setWarnings(append(action.Warnings(), opt.Warnings...))
	}
	env := object.NewEnvironment(action.ctx, action.db, hasFilter, getFilter, action.dbname, params,
		disableUpdate, disabledDDL, signature, external)
	result := nsina.Eval(optimizedProgram, env)
	return result, action.AllMessages()
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
	analyzer := semantic.NewSemanticAnalyzer(action.ctx, action.db, canHandle, nil, nil)
	analyzer.AnalyzeExpression(table, newName, act)
	if len(analyzer.Warnings) > 0 {
		action.setWarnings(append(action.Warnings(), analyzer.Warnings...))
	}
	if len(analyzer.Errors) > 0 {
		action.error = append(action.error, analyzer.Errors...)
		return nil, action.error
	}
	return act, action.error
}
func (action *Action) Check(src, id, table, newName string, canHandle func(table, field, operation string) (bool, string), serviceExists func(serviceName string) bool,
	signature func(serviceName, methodName string) ([]*ast.StructField, *ast.TypeAnnotation, error)) (bool, []string) {
	lex := lexer.New(src)
	p := parser.New(lex)
	switch strings.ToLower(id) {
	case "action":
		act := p.ParseAction()
		if p.Errors() != nil && len(p.Errors()) != 0 {
			for _, msg := range p.Errors() {
				action.error = append(action.error, msg.String())
			}
			return false, action.error
		}
		analyzer := semantic.NewSemanticAnalyzer(action.ctx, action.db, canHandle, serviceExists, signature)
		analyzer.Analyze(act)
		if len(analyzer.Warnings) > 0 {
			action.setWarnings(append(action.Warnings(), analyzer.Warnings...))
		}
		if len(analyzer.Errors) > 0 {
			action.error = append(action.error, analyzer.Errors...)
			return false, action.error
		}
		return true, nil
	case "expression":
		act := p.ParseExpression()
		if p.Errors() != nil && len(p.Errors()) != 0 {
			for _, msg := range p.Errors() {
				action.error = append(action.error, msg.String())
			}
			return false, action.error
		}
		analyzer := semantic.NewSemanticAnalyzer(action.ctx, action.db, canHandle, nil, nil)
		analyzer.AnalyzeExpression(table, newName, act)
		if len(analyzer.Warnings) > 0 {
			action.setWarnings(append(action.Warnings(), analyzer.Warnings...))
		}
		if len(analyzer.Errors) > 0 {
			action.error = append(action.error, analyzer.Errors...)
			return false, action.error
		}
		return true, nil
	default:
		res := make([]string, 0)
		res = append(res, "Invalid id")
		return false, nil
	}
}
func (action *Action) Errors() []string {
	return action.error
}
func (action *Action) Warnings() []string {
	return action.warnings
}
func (action *Action) AllMessages() []string {
	messages := make([]string, 0)
	messages = append(messages, action.error...)
	messages = append(messages, action.warnings...)
	return messages
}
func (action *Action) HasErrors() bool {
	return len(action.error) > 0
}
func (action *Action) HasWarnings() bool {
	return len(action.warnings) > 0
}
func (action *Action) setWarnings(warnings []string) {
	action.warnings = warnings
}

func (action *Action) ClearMessages() {
	action.error = make([]string, 0)
	action.warnings = make([]string, 0)
}
func (action *Action) ClearErrors() {
	action.error = make([]string, 0)
}
func (action *Action) ClearWarnings() {
	action.warnings = make([]string, 0)
}
func (action *Action) Signature(src string) ([]*ast.StructField, *ast.TypeAnnotation, []string) {
	lex := lexer.New(src)
	p := parser.New(lex)
	args, retType := p.ParseSignature()
	if p.Errors() != nil && len(p.Errors()) != 0 {
		for _, msg := range p.Errors() {
			action.error = append(action.error, msg.String())
		}
		return nil, nil, action.error
	}
	return args, retType, nil
}
