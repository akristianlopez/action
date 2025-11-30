package evaluator

import (
	"fmt"
	"time"

	"github.com/akristianlopez/action/ast"
	"github.com/akristianlopez/action/object"
)

func Eval(node ast.Node, env *object.Environment) object.Object {
	switch node := node.(type) {
	case *ast.Program:
		return evalProgram(node, env)
	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	case *ast.FloatLiteral:
		return &object.Float{Value: node.Value}
	case *ast.StringLiteral:
		return &object.String{Value: node.Value}
	case *ast.BooleanLiteral:
		return &object.Boolean{Value: node.Value}
	case *ast.DateTimeLiteral:
		return evalDateTimeLiteral(node)
	case *ast.Identifier:
		return evalIdentifier(node, env)
	case *ast.LetStatement:
		return evalLetStatement(node, env)
	case *ast.FunctionStatement:
		return evalFunctionStatement(node, env)
	case *ast.StructStatement:
		return evalStructStatement(node, env)
	case *ast.BlockStatement:
		return evalBlockStatement(node, env)
	case *ast.PrefixExpression:
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right)
	case *ast.InfixExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalInfixExpression(node.Operator, left, right)
	case *ast.ForStatement:
		return evalForStatement(node, env)
	case *ast.ReturnStatement:
		val := Eval(node.ReturnValue, env)
		if isError(val) {
			return val
		}
		return &object.ReturnValue{Value: val}
	case *ast.SQLSelectStatement:
		return evalSQLSelectStatement(node, env)
	}

	return nil
}

func evalProgram(program *ast.Program, env *object.Environment) object.Object {
	var result object.Object

	for _, statement := range program.Statements {
		result = Eval(statement, env)

		switch result := result.(type) {
		case *object.ReturnValue:
			return result.Value
		case *object.Error:
			return result
		}
	}

	return result
}

func evalLetStatement(let *ast.LetStatement, env *object.Environment) object.Object {
	var value object.Object

	if let.Value != nil {
		value = Eval(let.Value, env)
		if isError(value) {
			return value
		}
	} else {
		// Valeur par défaut selon le type
		if let.Type != nil {
			value = getDefaultValue(let.Type.Type)
		} else {
			value = object.NULL
		}
	}

	env.Set(let.Name.Value, value)
	return value
}

func getDefaultValue(typeName string) object.Object {
	switch typeName {
	case "integer":
		return &object.Integer{Value: 0}
	case "float":
		return &object.Float{Value: 0.0}
	case "string":
		return &object.String{Value: ""}
	case "boolean":
		return &object.Boolean{Value: false}
	case "time":
		return &object.Time{Value: time.Now()}
	case "date":
		return &object.Date{Value: time.Now()}
	default:
		return object.NULL
	}
}

func evalFunctionStatement(fn *ast.FunctionStatement, env *object.Environment) object.Object {
	function := &object.Function{
		Parameters: fn.Parameters,
		Body:       fn.Body,
		Env:        env,
	}
	env.Set(fn.Name.Value, function)
	return function
}

func evalStructStatement(st *ast.StructStatement, env *object.Environment) object.Object {
	structObj := &object.Struct{
		Name:   st.Name.Value,
		Fields: make(map[string]object.Object),
	}

	for _, field := range st.Fields {
		structObj.Fields[field.Name.Value] = getDefaultValue(field.Type.Type)
	}

	env.Set(st.Name.Value, structObj)
	return structObj
}

func evalBlockStatement(block *ast.BlockStatement, env *object.Environment) object.Object {
	var result object.Object

	for _, statement := range block.Statements {
		result = Eval(statement, env)

		if result != nil {
			rt := result.Type()
			if rt == object.RETURN_VALUE_OBJ || rt == object.ERROR_OBJ {
				return result
			}
		}
	}

	return result
}

func evalForStatement(forStmt *ast.ForStatement, env *object.Environment) object.Object {
	// Évaluer l'initialisation
	if forStmt.Init != nil {
		Eval(forStmt.Init, env)
	}

	for {
		// Évaluer la condition
		if forStmt.Condition != nil {
			condition := Eval(forStmt.Condition, env)
			if isError(condition) {
				return condition
			}

			if !isTruthy(condition) {
				break
			}
		}

		// Évaluer le corps
		result := Eval(forStmt.Body, env)
		if result != nil {
			if result.Type() == object.RETURN_VALUE_OBJ || result.Type() == object.ERROR_OBJ {
				return result
			}
		}

		// Évaluer l'update
		if forStmt.Update != nil {
			Eval(forStmt.Update, env)
		}
	}

	return object.NULL
}

func evalDateTimeLiteral(dt *ast.DateTimeLiteral) object.Object {
	// Enlever les # et parser la date/time
	value := dt.Literal[1 : len(dt.Literal)-1]

	if dt.IsTime {
		t, err := time.Parse("15:04:05", value)
		if err != nil {
			t, err = time.Parse("15:04", value)
			if err != nil {
				return newError("Format de temps invalide: %s", value)
			}
		}
		return &object.Time{Value: t}
	} else {
		// Essayer différents formats de date
		formats := []string{"2006-01-02", "02/01/2006", "01/02/2006"}
		for _, format := range formats {
			t, err := time.Parse(format, value)
			if err == nil {
				return &object.Date{Value: t}
			}
		}
		return newError("Format de date invalide: %s", value)
	}
}

func evalSQLSelectStatement(selectStmt *ast.SQLSelectStatement, env *object.Environment) object.Object {
	// Implémentation simplifiée pour la démonstration
	// Dans une vraie implémentation, cela interagirait avec une base de données

	result := &object.SQLResult{
		Columns: []string{},
		Rows:    []map[string]object.Object{},
	}

	// Traiter la clause SELECT
	for _, expr := range selectStmt.Select {
		if ident, ok := expr.(*ast.Identifier); ok {
			if ident.Value == "*" {
				// Sélectionner toutes les colonnes
				// Implémentation simplifiée
			} else {
				result.Columns = append(result.Columns, ident.Value)
			}
		}
	}

	// Traiter la clause FROM
	from := evalFromClause(selectStmt.From, env)
	if isError(from) {
		return from
	}

	// Traiter la clause WHERE
	if selectStmt.Where != nil {
		whereResult := Eval(selectStmt.Where, env)
		if isError(whereResult) {
			return whereResult
		}
	}

	return result
}

func evalFromClause(from ast.Expression, env *object.Environment) object.Object {
	switch from := from.(type) {
	case *ast.Identifier:
		// Rechercher l'objet dans l'environnement
		obj, ok := env.Get(from.Value)
		if !ok {
			return newError("Objet non trouvé: %s", from.Value)
		}
		return obj
	default:
		return Eval(from, env)
	}
}

// Les fonctions evalPrefixExpression, evalInfixExpression, evalIdentifier, etc.
// suivent le même pattern que dans un évaluateur standard...

func evalPrefixExpression(operator string, right object.Object) object.Object {
	switch operator {
	case "!":
		return evalBangOperatorExpression(right)
	case "-":
		return evalMinusPrefixOperatorExpression(right)
	default:
		return newError("Opérateur inconnu: %s%s", operator, right.Type())
	}
}

func evalInfixExpression(operator string, left, right object.Object) object.Object {
	switch {
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalIntegerInfixExpression(operator, left, right)
	case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
		return evalStringInfixExpression(operator, left, right)
	case operator == "==":
		return &object.Boolean{Value: left == right}
	case operator == "!=":
		return &object.Boolean{Value: left != right}
	default:
		return newError("Type mismatch: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalIntegerInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value

	switch operator {
	case "+":
		return &object.Integer{Value: leftVal + rightVal}
	case "-":
		return &object.Integer{Value: leftVal - rightVal}
	case "*":
		return &object.Integer{Value: leftVal * rightVal}
	case "/":
		return &object.Integer{Value: leftVal / rightVal}
	case "<":
		return &object.Boolean{Value: leftVal < rightVal}
	case ">":
		return &object.Boolean{Value: leftVal > rightVal}
	case "==":
		return &object.Boolean{Value: leftVal == rightVal}
	case "!=":
		return &object.Boolean{Value: leftVal != rightVal}
	default:
		return newError("Opérateur inconnu: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalStringInfixExpression(operator string, left, right object.Object) object.Object {
	if operator != "+" {
		return newError("Opérateur inconnu: %s %s %s", left.Type(), operator, right.Type())
	}

	leftVal := left.(*object.String).Value
	rightVal := right.(*object.String).Value
	return &object.String{Value: leftVal + rightVal}
}

func evalBangOperatorExpression(right object.Object) object.Object {
	switch right {
	case object.TRUE:
		return object.FALSE
	case object.FALSE:
		return object.TRUE
	case object.NULL:
		return object.TRUE
	default:
		return object.FALSE
	}
}

func evalMinusPrefixOperatorExpression(right object.Object) object.Object {
	if right.Type() != object.INTEGER_OBJ && right.Type() != object.FLOAT_OBJ {
		return newError("Opérateur inconnu: -%s", right.Type())
	}

	switch right := right.(type) {
	case *object.Integer:
		return &object.Integer{Value: -right.Value}
	case *object.Float:
		return &object.Float{Value: -right.Value}
	default:
		return newError("Opérateur inconnu: -%s", right.Type())
	}
}

func evalIdentifier(node *ast.Identifier, env *object.Environment) object.Object {
	if val, ok := env.Get(node.Value); ok {
		return val
	}

	return newError("Identifiant non trouvé: " + node.Value)
}

func isTruthy(obj object.Object) bool {
	switch obj {
	case object.NULL:
		return false
	case object.TRUE:
		return true
	case object.FALSE:
		return false
	default:
		return true
	}
}

func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ
	}
	return false
}
