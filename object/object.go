package object

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
	"unicode"

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
	DURATION_OBJ     = "DURATION"
)

type Object interface {
	Type() ObjectType
	Inspect() string
}

type Constraints struct {
	MaxDigits     int
	DecimalPlaces int
	MaxLength     int
}
type Integer struct {
	Value int64
	limit *map[string]Object
}

func (i *Integer) Set(name string, value Object) bool {
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
func (i *Integer) IsValid() (bool, string) {
	result := true
	if i.limit == nil {
		return result, ""
	}
	m, o := (*i.limit)[strings.ToLower("min")]
	if o {
		val := m.(*Integer).Value
		result = result && val <= i.Value
		if !result {
			return false, "Value '" + i.Inspect() + "' is lower than '" + strconv.FormatInt(val, 10) + "'"
		}
	}
	m, o = (*i.limit)[strings.ToLower("max")]
	if o {
		val := m.(*Integer).Value
		result = result && val >= i.Value
		if !result {
			return false, "Value '" + i.Inspect() + "' is greater than '" + strconv.FormatInt(val, 10) + "'"
		}
	}
	m, o = (*i.limit)[strings.ToLower("MaxLength")]
	if o {
		val := m.(*Integer).Value
		result = result && int64(len(i.Inspect())) <= val
		if !result {
			return false, "Value '" + i.Inspect() + "' too much digits than expected"
		}
	}
	return result, ""
}
func (i *Integer) Type() ObjectType { return INTEGER_OBJ }
func (i *Integer) Inspect() string  { return fmt.Sprintf("%d", i.Value) }

type Float struct {
	Value float64
	limit *map[string]Object
}

func (f *Float) Set(name string, value Object) bool {
	if f.limit == nil {
		o := make(map[string]Object)
		f.limit = &o
	}
	if value.Type() != INTEGER_OBJ && value.Type() != FLOAT_OBJ {
		return false
	}
	if value.Type() == INTEGER_OBJ {
		if value.(*Integer).limit != nil {
			if t, e := (*value.(*Integer).limit)["min"]; e {
				(*value.(*Integer).limit)["min"] = &Float{Value: float64(t.(*Integer).Value), limit: nil}
			}
			if t, e := (*value.(*Integer).limit)["max"]; e {
				(*value.(*Integer).limit)["max"] = &Float{Value: float64(t.(*Integer).Value), limit: nil}
			}
		}
		(*f.limit)[strings.ToLower(name)] = &Float{Value: float64(value.(*Integer).Value), limit: value.(*Integer).limit}
		return true
	}
	(*f.limit)[strings.ToLower(name)] = value
	return true
}
func (f *Float) IsValid() (bool, string) {
	result := true
	if f.limit == nil {
		return result, ""
	}
	m, o := (*f.limit)[strings.ToLower("min")]
	if o {
		val := m.(*Float).Value
		result = result && val <= f.Value
		if !result {
			return false, "Value '" + f.Inspect() + "' is lower than '" + m.(*Float).Inspect() + "'"
		}
	}
	m, o = (*f.limit)[strings.ToLower("max")]
	if o {
		val := m.(*Float).Value
		result = result && val >= f.Value
		if !result {
			return false, "Value '" + f.Inspect() + "' is greater than '" + m.(*Float).Inspect() + "'"
		}
	}
	m, o = (*f.limit)[strings.ToLower("MaxDigits")]
	if o {
		val := m.(*Integer).Value
		result = result && int64(countDigitsBeforeDecimal(f.Value)) <= val
		if !result {
			return false, "Value '" + f.Inspect() + "' too much digits than expected"
		}
	}
	return result, ""
}
func (f *Float) Type() ObjectType { return FLOAT_OBJ }
func (f *Float) Inspect() string {
	m, o := (*f.limit)[strings.ToLower("DecimalPlaces")]
	if o {
		str := "%." + m.(*Integer).Inspect() + "f"
		return fmt.Sprintf(str, f.Value)
	}
	return fmt.Sprintf("%f", f.Value)
}

type Boolean struct {
	Value bool
}

func (b *Boolean) Type() ObjectType { return BOOLEAN_OBJ }
func (b *Boolean) Inspect() string  { return fmt.Sprintf("%t", b.Value) }

type String struct {
	Value string
	limit *map[string]Object
}

func (s *String) Set(name string, value Object) bool {
	if s.limit == nil {
		o := make(map[string]Object)
		s.limit = &o
	}
	(*s.limit)[strings.ToLower(name)] = value
	return true
}

func (s *String) IsValid() (bool, string) {
	result := true
	if s.limit == nil {
		return result, ""
	}
	m, o := (*s.limit)[strings.ToLower("MaxLength")]
	if o {
		val := m.(*Integer).Value
		result = result && int64(len(s.Inspect())) <= val
		if !result {
			return false, "String '" + s.Inspect() + "' too long than expected"
		}
	}
	return result, ""
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
