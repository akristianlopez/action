package object

import (
	"fmt"
	"strings"
	"time"

	"github.com/akristianlopez/action/ast"
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
	SQL_RESULT_OBJ   = "SQL_RESULT"
	ARRAY_OBJ        = "ARRAY"
	BREAK_OBJ        = "BREAK"
	FALLTHROUGH_OBJ  = "FALLTHROUGH"
	CONTINUE_OBJ     = "CONTINUE"
)

type Object interface {
	Type() ObjectType
	Inspect() string
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
func (f *Float) Inspect() string  { return fmt.Sprintf("%f", f.Value) }

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
	return fmt.Sprintf("struct %s", s.Name)
}

var (
	NULL  = &Null{}
	TRUE  = &Boolean{Value: true}
	FALSE = &Boolean{Value: false}
)

type Environment struct {
	store map[string]Object
	outer *Environment
}

func NewEnvironment() *Environment {
	s := make(map[string]Object)
	return &Environment{store: s, outer: nil}
}

func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	return env
}

func (e *Environment) Get(name string) (Object, bool) {
	obj, ok := e.store[name]
	if !ok && e.outer != nil {
		obj, ok = e.outer.Get(name)
	}
	return obj, ok
}

func (e *Environment) Set(name string, val Object) Object {
	e.store[name] = val
	return val
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
	Rows         []map[string]Object
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
