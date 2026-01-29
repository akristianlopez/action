package object

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/akristianlopez/action/ast"
	"github.com/gin-gonic/gin"
)

type ObjectType string

const (
	INTEGER_OBJ      = "INTEGER"
	FLOAT_OBJ        = "FLOAT"
	BOOLEAN_OBJ      = "BOOLEAN"
	STRING_OBJ       = "STRING"
	TIME_OBJ         = "TIME"
	DATE_OBJ         = "DATE"
	NULL_OBJ         = "NULL"
	RETURN_VALUE_OBJ = "RETURN_VALUE"
	ERROR_OBJ        = "ERROR"
	FUNCTION_OBJ     = "FUNCTION"
	STRUCT_OBJ       = "STRUCT"
	DBFIELD_OBJ      = "DB_FIELD"
	DBOBJECT_OBJ     = "TABLE"
	SQL_RESULT_OBJ   = "SQL_RESULT"
	ARRAY_OBJ        = "ARRAY"
	ROWS_OBJ         = "ROWS"
	BREAK_OBJ        = "BREAK"
	FALLTHROUGH_OBJ  = "FALLTHROUGH"
	CONTINUE_OBJ     = "CONTINUE"
	DURATION_OBJ     = "DURATION"
)

type Object interface {
	Type() ObjectType
	Inspect() string
}
type Limits struct {
	_type ObjectType
	limit *map[string]Object
}

func (l *Limits) Valid(value Object) (bool, string) {
	switch l._type {
	case INTEGER_OBJ:
		return l.isValidInt(value.(*Integer))
	case FLOAT_OBJ:
		return l.isValidFloat(value.(*Float))
	case STRING_OBJ:
		return l.isValidString(value.(*String))
	case TIME_OBJ:
		return l.isValidTime(value.(*Time))
	case DATE_OBJ:
		return l.isValidDate(value.(*Date))
	case DURATION_OBJ:
		return l.isValidDuration(value.(*Duration))
	}
	return false, "Type non supported"
}
func (l *Limits) SetType(t ObjectType) {
	l._type = t
}
func (l *Limits) Type() ObjectType {
	return l._type
}
func (l *Limits) Set(name string, value Object) bool {
	switch l._type {
	case INTEGER_OBJ:
		return l.setIntLimit(name, value)
	case FLOAT_OBJ:
		return l.setFloatLimit(name, value)
	case STRING_OBJ:
		return l.setStringLimit(name, value)
	case TIME_OBJ:
		return l.setTimeLimit(name, value)
	case DATE_OBJ:
		return l.setDateLimit(name, value)
	case DURATION_OBJ:
		return l.setDurationLimit(name, value)
	}
	return false
}
func (i *Limits) setIntLimit(name string, value Object) bool {
	if i.limit == nil {
		o := make(map[string]Object)
		i.limit = &o
	}
	if value.Type() != INTEGER_OBJ {
		return false
	}
	(*i.limit)[strings.ToLower(name)] = value
	return true
}
func (l *Limits) isValidInt(value *Integer) (bool, string) {
	result := true
	if l.limit == nil {
		return result, ""
	}
	m, o := (*l.limit)[strings.ToLower("min")]
	if o {
		val := m.(*Integer).Value
		result = result && val <= value.Value
		if !result {
			return false, "Value '" + value.Inspect() + "' is lower than '" + m.(*Integer).Inspect() + "'"
		}
	}
	m, o = (*l.limit)[strings.ToLower("max")]
	if o {
		val := m.(*Integer).Value
		result = result && val >= value.Value
		if !result {
			return false, "Value '" + value.Inspect() + "' is greater than '" + m.(*Integer).Inspect() + "'"
		}
	}
	m, o = (*l.limit)[strings.ToLower("MaxDigits")]
	if o {
		val := m.(*Integer).Value
		result = result && int64(len(m.(*Integer).Inspect())) <= val
		if !result {
			return false, "Value '" + value.Inspect() + "' too much digits than expected"
		}
	}
	return result, ""
}
func (l *Limits) setFloatLimit(name string, value Object) bool {
	if l.limit == nil {
		o := make(map[string]Object)
		l.limit = &o
	}
	if value.Type() != INTEGER_OBJ && value.Type() != FLOAT_OBJ {
		return false
	}
	if value.Type() == INTEGER_OBJ {
		(*l.limit)[strings.ToLower(name)] = &Float{Value: float64(value.(*Integer).Value)}
		return true
	}
	(*l.limit)[strings.ToLower(name)] = value
	return true
}
func (l *Limits) isValidFloat(value *Float) (bool, string) {
	result := true
	if l.limit == nil {
		return result, ""
	}
	m, o := (*l.limit)[strings.ToLower("min")]
	if o {
		val := m.(*Float).Value
		result = result && val <= value.Value
		if !result {
			return false, "Value '" + value.Inspect() + "' is lower than '" + m.(*Float).Inspect() + "'"
		}
	}
	m, o = (*l.limit)[strings.ToLower("max")]
	if o {
		val := m.(*Float).Value
		result = result && val >= value.Value
		if !result {
			return false, "Value '" + value.Inspect() + "' is greater than '" + m.(*Float).Inspect() + "'"
		}
	}
	m, o = (*l.limit)[strings.ToLower("MaxDigits")]
	if o {
		val := m.(*Integer).Value
		result = result && int64(countDigitsBeforeDecimal(value.Value)) <= val
		if !result {
			return false, "Value '" + value.Inspect() + "' too much digits than expected"
		}
	}
	return result, ""
}
func (l *Limits) setStringLimit(name string, value Object) bool {
	if l.limit == nil {
		o := make(map[string]Object)
		l.limit = &o
	}
	(*l.limit)[strings.ToLower(name)] = value
	return true
}
func (l *Limits) isValidString(value *String) (bool, string) {
	result := true
	if l.limit == nil {
		return result, ""
	}
	m, o := (*l.limit)[strings.ToLower("MaxLength")]
	if o {
		val := m.(*Integer).Value
		result = result && int64(len(value.Inspect())) <= val
		if !result {
			return false, "String '" + value.Inspect() + "' too long than expected"
		}
	}
	return result, ""
}
func (l *Limits) setTimeLimit(name string, value Object) bool {
	if value.Type() != TIME_OBJ {
		return false
	}
	if l.limit == nil {
		o := make(map[string]Object)
		l.limit = &o
	}
	(*l.limit)[strings.ToLower(name)] = value
	return true
}
func (l *Limits) isValidTime(value *Time) (bool, string) {
	result := true
	if l.limit == nil {
		return result, ""
	}

	m, o := (*l.limit)[strings.ToLower("min")]
	if o {
		val := m.(*Time).Value
		result = result && val.Compare(value.Value) <= 0
		if !result {
			return false, "Value '" + value.Inspect() + "' is lower than '" + m.(*Time).Inspect() + "'"
		}
	}
	m, o = (*l.limit)[strings.ToLower("max")]
	if o {
		val := m.(*Time).Value
		result = result && val.Compare(value.Value) >= 0
		if !result {
			return false, "Value '" + value.Inspect() + "' is greater than '" + m.(*Time).Inspect() + "'"
		}
	}
	return result, ""
}
func (l *Limits) setDateLimit(name string, value Object) bool {
	if value.Type() != DATE_OBJ {
		return false
	}
	if l.limit == nil {
		o := make(map[string]Object)
		l.limit = &o
	}
	(*l.limit)[strings.ToLower(name)] = value
	return true
}
func (l *Limits) isValidDate(value *Date) (bool, string) {
	result := true
	if l.limit == nil {
		return result, ""
	}

	m, o := (*l.limit)[strings.ToLower("min")]
	if o {
		val := m.(*Time).Value
		result = result && val.Compare(value.Value) <= 0
		if !result {
			return false, "Value '" + value.Inspect() + "' is lower than '" + m.(*Time).Inspect() + "'"
		}
	}
	m, o = (*l.limit)[strings.ToLower("max")]
	if o {
		val := m.(*Time).Value
		result = result && val.Compare(value.Value) >= 0
		if !result {
			return false, "Value '" + value.Inspect() + "' is greater than '" + m.(*Time).Inspect() + "'"
		}
	}
	return result, ""
}
func (l *Limits) setDurationLimit(name string, value Object) bool {
	if value.Type() != DURATION_OBJ {
		return false
	}
	if l.limit == nil {
		o := make(map[string]Object)
		l.limit = &o
	}
	(*l.limit)[strings.ToLower(name)] = value
	return true
}
func (l *Limits) isValidDuration(d *Duration) (bool, string) {
	result := true
	if l.limit == nil {
		return result, ""
	}

	m, o := (*l.limit)[strings.ToLower("min")]
	if o {
		val := m.(*Duration).Nanoseconds
		result = result && val <= d.Nanoseconds
		if !result {
			return false, "Value '" + d.Inspect() + "' is lower than '" + m.(*Duration).Inspect() + "'"
		}
	}
	m, o = (*l.limit)[strings.ToLower("max")]
	if o {
		val := m.(*Duration).Nanoseconds
		result = result && val >= d.Nanoseconds
		if !result {
			return false, "Value '" + d.Inspect() + "' is greater than '" + m.(*Duration).Inspect() + "'"
		}
	}
	return result, ""
}

type Integer struct {
	Value int64
}

func (i *Integer) Type() ObjectType { return INTEGER_OBJ }
func (i *Integer) Inspect() string  { return fmt.Sprintf("%d", i.Value) }

type Float struct {
	Value float64
}

func (f *Float) Type() ObjectType { return FLOAT_OBJ }
func (f *Float) Inspect() string {
	return fmt.Sprintf("%f", f.Value)
}

type Boolean struct {
	Value bool
}

func (b *Boolean) Type() ObjectType { return BOOLEAN_OBJ }
func (b *Boolean) Inspect() string  { return fmt.Sprintf("%t", b.Value) }

type String struct {
	Value string
}

func (s *String) Type() ObjectType { return STRING_OBJ }
func (s *String) Inspect() string  { return s.Value }

type DBField struct {
	Value string
}

func (s *DBField) Type() ObjectType { return DBFIELD_OBJ }
func (s *DBField) Inspect() string  { return s.Value }

type Time struct {
	Value time.Time
}

func (t *Time) Type() ObjectType { return TIME_OBJ }
func (t *Time) Inspect() string  { return t.Value.Format("15:04:05") }

type Date struct {
	Value time.Time
}

func (d *Date) Type() ObjectType { return DATE_OBJ }
func (d *Date) Inspect() string  { return d.Value.Format("2006-01-02") }

type Null struct{}

func (n *Null) Type() ObjectType { return NULL_OBJ }
func (n *Null) Inspect() string  { return "null" }

type ReturnValue struct {
	Value Object
}

func (rv *ReturnValue) Type() ObjectType { return RETURN_VALUE_OBJ }
func (rv *ReturnValue) Inspect() string  { return rv.Value.Inspect() }

type Error struct {
	Message string
}

func (e *Error) Type() ObjectType { return ERROR_OBJ }
func (e *Error) Inspect() string  { return "ERREUR: " + e.Message }

type Function struct {
	Parameters []*ast.FunctionParameter
	Body       *ast.BlockStatement
	Env        *Environment
}

func (f *Function) Type() ObjectType { return FUNCTION_OBJ }
func (f *Function) Inspect() string {
	return "function"
}

type Struct struct {
	Name   string
	Fields map[string]Object
}

func (s *Struct) Type() ObjectType { return STRUCT_OBJ }
func (s *Struct) Inspect() string {
	out := ""
	if s.Fields != nil {
		for k, v := range s.Fields {
			if out == "" {
				out = fmt.Sprintf("%s: %s", k, v.Inspect())
				continue
			}
			out = fmt.Sprintf("%s,\n %s: %s", out, k, v.Inspect())
		}
	}
	return fmt.Sprintf("%s{%s}", s.Name, out)
}

type DBStruct struct {
	Name   string
	Fields map[string]Object
}

func (s *DBStruct) Type() ObjectType { return DBOBJECT_OBJ }
func (s *DBStruct) Inspect() string {
	return fmt.Sprintf("Table %s", s.Name)
}

var (
	NULL  = &Null{}
	TRUE  = &Boolean{Value: true}
	FALSE = &Boolean{Value: false}
)

type Environment struct {
	store         map[string]Object
	outer         *Environment
	limits        *map[string]Limits
	db            *sql.DB
	ctx           *gin.Context
	hasFilter     func(ctx *gin.Context, table string) bool
	getFilter     func(ctx *gin.Context, table, newName string) (ast.Expression, bool)
	dbname        string
	params        *map[string]Object
	disableUpdate bool
	disabledDDL   bool
	external      func(ctx *gin.Context, srv, name string, args map[string]Object) (Object, bool)
	signature     func(ctx *gin.Context, serviceName, methodName string) ([]*ast.StructField, *ast.TypeAnnotation, error)
}

func (env *Environment) Context() context.Context {
	return env.ctx
}
func (env *Environment) DBName() string {
	return env.dbname
}
func NewEnvironment(ctx *gin.Context, db *sql.DB, hf func(ctx *gin.Context, table string) bool,
	gf func(ctx *gin.Context, table, newName string) (ast.Expression, bool), dbname string, params map[string]Object,
	disableUpdate, disabledDDL bool, sign func(ctx *gin.Context, serviceName, methodName string) ([]*ast.StructField, *ast.TypeAnnotation, error),
	external func(ctx *gin.Context, srv, name string, args map[string]Object) (Object, bool)) *Environment {
	s := make(map[string]Object)

	return &Environment{store: s, outer: nil, limits: nil, db: db, ctx: ctx,
		hasFilter: hf, getFilter: gf, dbname: dbname, params: &params,
		disableUpdate: disableUpdate, disabledDDL: disabledDDL, external: external, signature: sign}
}
func (env *Environment) IsParams(name string) bool {
	if env.params == nil {
		return false
	}
	_, ok := (*env.params)[strings.ToLower(name)]
	return ok
}
func (env *Environment) Params(name string) Object {
	if env.params == nil {
		return NULL
	}
	val, ok := (*env.params)[strings.ToLower(name)]
	if !ok {
		return NULL
	}
	return val
}
func (env *Environment) Filter(table, newName string) (ast.Expression, bool) {
	if env.getFilter == nil {
		return nil, false
	}
	return env.getFilter(env.ctx, table, newName)
}
func (env *Environment) IsFiltered(table string) bool {
	if env.hasFilter == nil {
		return false
	}
	return env.hasFilter(env.ctx, table)
}
func (env *Environment) Exec(strSQL string, args ...any) (sql.Result, error) {
	if env.db != nil && strSQL != "" {
		if env.ctx == nil {
			return nil, errors.New("Nsina: Context is not defined")
		}
		if len(args) == 0 {
			return env.db.ExecContext(env.ctx, strSQL)
		}
		return env.db.ExecContext(env.ctx, strSQL, args...)
	}
	if env.db == nil {
		return nil, errors.New("Nsina: no defined database")
	}
	return nil, errors.New("Nsina: no query to be executed")
}
func (env *Environment) External(srv, name string, args map[string]Object) (Object, bool) {
	if env.external == nil {
		return nil, false
	}
	return env.external(env.ctx, srv, name, args)
}
func (env *Environment) Signature(srv, name string) ([]*ast.StructField, *ast.TypeAnnotation, error) {
	return env.signature(env.ctx, srv, name)
}
func (env *Environment) Query(strSQL string, args ...any) (*sql.Rows, error) {
	if env.db != nil && strSQL != "" {
		if env.ctx == nil {
			return nil, errors.New("Nsina: Context is not defined")
		}
		if len(args) == 0 {
			return env.db.QueryContext(env.ctx, strSQL)
		}
		return env.db.QueryContext(env.ctx, strSQL, args)
	}
	if env.db == nil {
		return nil, errors.New("Nsina: no defined database")
	}
	return nil, errors.New("Nsina: no query to be executed")
}

func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment(outer.ctx, outer.db, outer.hasFilter, outer.getFilter, outer.dbname, nil,
		outer.disableUpdate, outer.disabledDDL, outer.signature, outer.external)
	env.outer = outer
	env.limits = nil
	return env
}
func (e *Environment) IsUpdateAllowed() bool {
	return !e.disableUpdate
}
func (e *Environment) IsDDLAllowed() bool {
	return !e.disabledDDL
}

func (e *Environment) Get(name string) (Object, bool) {
	obj, ok := e.store[strings.ToLower(name)]
	if !ok && e.outer != nil {
		obj, ok = e.outer.Get(name)
	}
	return obj, ok
}

func (e *Environment) isExist(name string) (*Environment, bool) {
	_, ok := e.store[strings.ToLower(name)]
	if !ok && e.outer != nil {
		return e.outer.isExist(name)
	}
	if ok {
		return e, ok
	}
	return nil, false
}

func (e *Environment) Set(name string, val Object) Object {
	n, ok := e.isExist(name)
	if ok {
		n.store[strings.ToLower(name)] = val
		return val
	}
	e.store[strings.ToLower(name)] = val
	return val
}

func (e *Environment) Clear() {
	e.store = make(map[string]Object)
}

func (e *Environment) Limit(name string, val *Limits) {
	if val == nil {
		return
	}
	if e.limits == nil {
		o := make(map[string]Limits)
		e.limits = &o
	}
	(*e.limits)[strings.ToLower(name)] = *val
}

func (e *Environment) HasLimits(name string) bool {
	if e.limits == nil {
		return false
	}
	_, ok := (*e.limits)[strings.ToLower(name)]
	return ok
}
func (e *Environment) GetLimitEnv(name string) *Environment {
	env := e
	if env.limits == nil {
		env = env.outer
	}
	if env == nil {
		return nil
	}
	for !env.HasLimits(name) {
		env = env.outer
		if env == nil {
			break
		}
	}
	return env
}
func (e *Environment) Valid(name string, value Object) (bool, string) {
	if e.limits == nil {
		return true, ""
	}
	if lim, ok := (*e.limits)[strings.ToLower(name)]; ok {
		return lim.Valid(value)
	}
	return true, ""
}

func (e *Environment) IsStructExist(node *Struct, env *Environment) string {
	// keys := make([]string, 0)
	Scope := env
	returnType := ""
	for {
		ok := false
		for _, val := range env.store {
			if val.Type() == STRUCT_OBJ {
				st := val.(*Struct)
				ok = true
				for name, field := range node.Fields {
					currentType := field.Type()
					expectedType, exists := st.Fields[strings.ToLower(name)]
					if !exists || ((expectedType.Type() != currentType) &&
						(expectedType.Type() != NULL_OBJ || currentType != STRUCT_OBJ)) {
						ok = false
						break
					}
				}
				if ok {
					returnType = st.Name
					break
				}
			}
		}
		if ok {
			break
		}
		Scope = Scope.outer
		if Scope == nil {
			break
		}
	}
	return returnType
}

// SQLTable - Représente une table/objet SQL
type SQLTable struct {
	Name    string
	Columns map[string]*SQLColumn
	Data    []map[string]Object
	Indexes map[string]*SQLIndex
}

func (st *SQLTable) Type() ObjectType { return "SQL_TABLE" }
func (st *SQLTable) Inspect() string {
	return fmt.Sprintf("TABLE %s (%d colonnes, %d lignes)", st.Name, len(st.Columns), len(st.Data))
}

// SQLColumn - Colonne d'une table
type SQLColumn struct {
	Name         string
	Type         string
	PrimaryKey   bool
	NotNull      bool
	Unique       bool
	DefaultValue Object
}

// SQLIndex - Index sur une table
type SQLIndex struct {
	Name    string
	Columns []string
	Unique  bool
}

// SQLResult - Résultat d'une opération SQL
type SQLResult struct {
	Message      string
	RowsAffected int64
	Columns      []string
	Rows         *sql.Rows //[]map[string]Object
}

func (sr *SQLResult) Type() ObjectType { return SQL_RESULT_OBJ }
func (sr *SQLResult) Inspect() string {
	if sr.Message != "" {
		return sr.Message
	}
	return fmt.Sprintf("SQLResult(%d ligne(s) affectée(s))", sr.RowsAffected)
}

// HierarchicalTree - Arbre hiérarchique
type HierarchicalTree struct {
	Roots []*HierarchicalNode
	Nodes map[string]*HierarchicalNode
}

func (ht *HierarchicalTree) Type() ObjectType { return "HIERARCHICAL_TREE" }
func (ht *HierarchicalTree) Inspect() string {
	return fmt.Sprintf("HierarchicalTree(%d racines, %d nœuds)", len(ht.Roots), len(ht.Nodes))
}

// HierarchicalNode - Nœud dans un arbre hiérarchique
type HierarchicalNode struct {
	ID       string
	Data     map[string]Object
	Level    int
	Parent   *HierarchicalNode
	Children []*HierarchicalNode
}

// WindowState - État pour le calcul des fonctions de fenêtrage
type WindowState struct {
	Partition  []map[string]Object
	CurrentRow int
	OrderBy    []string
}

// Array - Type tableau
type Array struct {
	Elements    []Object
	ElementType string // Type des éléments (optionnel)
	FixedSize   bool   // Taille fixe
	Size        int64  // Taille maximale
}

func (a *Array) Type() ObjectType { return ARRAY_OBJ }
func (a *Array) Inspect() string {
	var out strings.Builder
	out.WriteString("[")

	for i, element := range a.Elements {
		if i > 0 {
			out.WriteString(", ")
		}
		out.WriteString(element.Inspect())
	}

	out.WriteString("]")
	return out.String()
}

type Break struct{}

func (b *Break) Type() ObjectType { return BREAK_OBJ }
func (b *Break) Inspect() string  { return "break" }

// Fallthrough - Type fallthrough
type Fallthrough struct {
	Value bool
}

func (f *Fallthrough) Type() ObjectType { return FALLTHROUGH_OBJ }
func (f *Fallthrough) Inspect() string  { return "fallthrough" }

// Continue - Type continue (pour les boucles)
type Continue struct{}

func (c *Continue) Type() ObjectType { return CONTINUE_OBJ }
func (c *Continue) Inspect() string  { return "continue" }

type Duration struct {
	Nanoseconds int64  // Stockage interne en nanosecondes
	Original    string // Représentation originale
}

func (d *Duration) Type() ObjectType { return DURATION_OBJ }
func (d *Duration) Inspect() string {
	if d.Original != "" {
		return d.Original
	}
	return formatDuration(d.Nanoseconds)
}

func formatDuration(nanos int64) string {
	if nanos == 0 {
		return "#0s#"
	}

	var parts []string

	// Années (approximatif: 365.25 jours)
	if nanos >= int64(365.25*24*60*60*1e9) {
		years := nanos / int64(365.25*24*60*60*1e9)
		parts = append(parts, fmt.Sprintf("%dy", years))
		nanos %= int64(365.25 * 24 * 60 * 60 * 1e9)
	}

	// Mois (approximatif: 30.44 jours)
	if nanos >= int64(30.44*24*60*60*1e9) {
		months := nanos / int64(30.44*24*60*60*1e9)
		parts = append(parts, fmt.Sprintf("%dmo", months))
		nanos %= int64(30.44 * 24 * 60 * 60 * 1e9)
	}

	// Jours
	if nanos >= 24*60*60*1e9 {
		days := nanos / (24 * 60 * 60 * 1e9)
		parts = append(parts, fmt.Sprintf("%dd", days))
		nanos %= 24 * 60 * 60 * 1e9
	}

	// Heures
	if nanos >= 60*60*1e9 {
		hours := nanos / (60 * 60 * 1e9)
		parts = append(parts, fmt.Sprintf("%dh", hours))
		nanos %= 60 * 60 * 1e9
	}

	// Minutes
	if nanos >= 60*1e9 {
		minutes := nanos / (60 * 1e9)
		parts = append(parts, fmt.Sprintf("%dm", minutes))
		nanos %= 60 * 1e9
	}

	// Secondes
	if nanos >= 1e9 {
		seconds := nanos / 1e9
		parts = append(parts, fmt.Sprintf("%ds", seconds))
		nanos %= 1e9
	}

	// Millisecondes
	if nanos >= 1e6 {
		ms := nanos / 1e6
		parts = append(parts, fmt.Sprintf("%dms", ms))
		nanos %= 1e6
	}

	// Microsecondes
	if nanos >= 1e3 {
		us := nanos / 1e3
		parts = append(parts, fmt.Sprintf("%dus", us))
		nanos %= 1e3
	}

	// Nanosecondes
	if nanos > 0 {
		parts = append(parts, fmt.Sprintf("%dns", nanos))
	}

	return "#" + strings.Join(parts, " ") + "#"
}

// Fonctions utilitaires pour Duration
func NewDuration(nanos int64) *Duration {
	return &Duration{Nanoseconds: nanos}
}
func ParseDuration(literal string) (*Duration, error) {
	if len(literal) < 3 || literal[0] != '#' || literal[len(literal)-1] != '#' {
		return nil, fmt.Errorf("format de durée invalide")
	}

	// Enlever les #
	str := literal[1 : len(literal)-1]
	str = strings.TrimSpace(str)

	if str == "" {
		return NewDuration(0), nil
	}

	var totalNanos int64

	// Parser les différentes unités
	parts := strings.Fields(str)
	for _, part := range parts {
		// Trouver la dernière lettre qui est l'unité
		i := len(part) - 1
		for i >= 0 && !unicode.IsDigit(rune(part[i])) {
			i--
		}

		if i < 0 {
			continue
		}

		valueStr := part[:i+1]
		unit := part[i+1:]

		value, err := strconv.ParseInt(valueStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("valeur invalide dans la durée: %s", valueStr)
		}

		switch unit {
		case "y", "year", "years":
			totalNanos += value * int64(365.25*24*60*60*1e9)
		case "mo", "month", "months":
			totalNanos += value * int64(30.44*24*60*60*1e9)
		case "w", "week", "weeks":
			totalNanos += value * 7 * 24 * 60 * 60 * 1e9
		case "d", "day", "days":
			totalNanos += value * 24 * 60 * 60 * 1e9
		case "h", "hour", "hours":
			totalNanos += value * 60 * 60 * 1e9
		case "m", "min", "minute", "minutes":
			totalNanos += value * 60 * 1e9
		case "s", "sec", "second", "seconds":
			totalNanos += value * 1e9
		case "ms", "millisecond", "milliseconds":
			totalNanos += value * 1e6
		case "us", "µs", "microsecond", "microseconds":
			totalNanos += value * 1e3
		case "ns", "nanosecond", "nanoseconds":
			totalNanos += value
		default:
			return nil, fmt.Errorf("unité de durée inconnue: %s", unit)
		}
	}

	return &Duration{
		Nanoseconds: totalNanos,
		Original:    literal,
	}, nil
}

// countDigitsBeforeDecimal returns the number of digits in the integer part
func countDigitsBeforeDecimal(f float64) int {
	f = math.Abs(f)
	if f < 1 {
		return 1 // e.g., 0.x has 1 digit before decimal
	}
	return int(math.Floor(math.Log10(f))) + 1
}

// countDigitsAfterDecimal returns the number of digits after the decimal point
func countDigitsAfterDecimal(f float64) int {
	// Convert to string without scientific notation
	s := strconv.FormatFloat(f, 'f', -1, 64)
	if strings.Contains(s, ".") {
		return len(strings.Split(s, ".")[1])
	}
	return 0
}
