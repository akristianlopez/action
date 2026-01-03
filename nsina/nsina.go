package nsina

import (
	"fmt"
	"strings"
	"time"

	"github.com/akristianlopez/action/ast"
	"github.com/akristianlopez/action/object"
	// _ "github.com/go-sql-driver/mysql" // Import du driver MySQL/MariaDB
	// _ "github.com/lib/pq"              // Driver PostgreSQL
	// _ "github.com/mattn/go-sqlite3"    // Import du driver SQLite
)

var struct_id int = 0

func Eval(node ast.Node, env *object.Environment) object.Object {
	if node == nil {
		return nil
	}

	switch node := node.(type) {
	case *ast.Action:
		return evalAction(node, env)
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
	case *ast.NullLiteral:
		return &object.Null{}
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
	case *ast.AssignmentStatement:
		return evalAssignmentStatement(node, env)
	case *ast.ForEachStatement:
		return evalForEachStatement(node, env)
	case *ast.WhileStatement:
		return evalWhileStatement(node, env)
	case *ast.TypeMember:
		//TODO: A definir
		return evalTypeMember(node, env)
	case *ast.BetweenExpression:
		//TODO: A definir
		return evalBetweenExpression(node, env)
	case *ast.StructLiteral:
		//TODO: A definir
		return evalStructLiteral(node, env)
	case *ast.IfStatement:
		//TODO: A definir
		return evalIfStatement(node, env)
	case *ast.ForStatement:
		return evalForStatement(node, env)
	case *ast.ReturnStatement:
		val := Eval(node.ReturnValue, env)
		if isError(val) {
			return val
		}
		return &object.ReturnValue{Value: val}
	case *ast.SQLWithStatement:
		return evalSQLWithStatement(node, env)
	case *ast.SQLSelectStatement:
		return evalSQLSelectStatement(node, env)
	case *ast.ArrayLiteral:
		return evalArrayLiteral(node, env)
	case *ast.IndexExpression:
		return evalIndexExpression(node, env)
	case *ast.SliceExpression:
		return evalSliceExpression(node, env)
	case *ast.InExpression:
		return evalInExpression(node, env)
	case *ast.ArrayFunctionCall:
		return evalArrayFunctionCall(node, env)
	case *ast.SwitchStatement:
		return evalSwitchStatement(node, env)
	case *ast.BreakStatement:
		return evalBreakStatement(node, env)
	case *ast.DurationLiteral:
		return evalDurationLiteral(node, env)
	case *ast.FallthroughStatement:
		return evalFallthroughStatement(node, env)
	case *ast.SQLCreateObjectStatement:
		return evalSQLCreateObject(node, env)
	case *ast.SQLDropObjectStatement:
		return evalSQLDropObject(node, env)
	case *ast.SQLAlterObjectStatement:
		return evalSQLAlterObject(node, env)
	case *ast.SQLInsertStatement:
		return evalSQLInsert(node, env)
	case *ast.SQLUpdateStatement:
		return evalSQLUpdate(node, env)
	case *ast.SQLDeleteStatement:
		return evalSQLDelete(node, env)
	case *ast.SQLTruncateStatement:
		return evalSQLTruncate(node, env)
	case *ast.SQLCreateIndexStatement:
		return evalSQLCreateIndex(node, env)
	default:
		return newError("Instruction non supportée: %T", node)
	}

	// return nil
}

func evalTypeMember(node *ast.TypeMember, env *object.Environment) object.Object {
	obj, fl := env.Get(node.Left.String())
	if !fl {
		return newError("Invalid structure name '%s'", node.Left.String())
	}
	key := node.Right
	for val, ok := key.(*ast.TypeMember); ok; {
		ok = false
		if ob, o := obj.(*object.Struct); o {
			if vl, exists := ob.Fields[strings.ToLower(val.Left.String())]; exists {
				obj = vl
				key = val.Right
				val, ok = key.(*ast.TypeMember)
				continue
			}
			return newError("Invalid field name '%s'. Line:%d, column:%d", val.Left.String(),
				key.Line(), key.Column())
		}
		break
	}
	if obj.Type() == object.STRUCT_OBJ {
		ob := obj.(*object.Struct)
		if value, exists := ob.Fields[key.String()]; exists {
			return value
		}
		return newError("Invalid field name '%s'.", node.Right.String())
	}
	return newError("Invalid type of object '%s'. expected '%v', got '%v'", node.Left.String(),
		object.STRUCT_OBJ, obj.Type())
}

func evalAction(program *ast.Action, env *object.Environment) object.Object {
	var result object.Object

	for _, statement := range program.Statements {
		select {
		case <-env.Context().Done():
			return newError("%s: Canceled by the user", "Nsina")
		default:
			result = Eval(statement, env)

			switch result := result.(type) {
			case *object.ReturnValue:
				return result.Value
			case *object.Error:
				return result
			}
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
	if let.Type != nil {
		env.Limit(let.Name.Value, defConstraints(let.Type, env))
		if st, ok := value.(*object.Struct); ok && st.Name == "" && let.Type.Type != "" {
			objtype := env.IsStructExist(st, env)
			if objtype == "" {
				struct_id++
				objtype = fmt.Sprintf("%s_%d", "struct_id", struct_id)
				defType(objtype, value, env)
			} else { //check if all values are ok
				for k, v := range st.Fields {
					en := env.GetLimitEnv(fmt.Sprintf("%s.%s", objtype, k))
					if en != nil {
						ok, msg := en.Valid(fmt.Sprintf("%s.%s", objtype, k), v)
						if !ok {
							return newError(msg+". Line:%d, column:%d", let.Value.Line(), let.Value.Column())
						}
					}
				}
			}
			st.Name = objtype
			value = st
		}
	}
	if ok, msg := env.Valid(let.Name.Value, value); !ok {
		return &object.Error{Message: msg}
	}
	return env.Set(let.Name.Value, value)
}

func defType(objtype string, value object.Object, env *object.Environment) {
	structObj := &object.Struct{
		Name:   objtype,
		Fields: make(map[string]object.Object),
	}
	st := value.(*object.Struct)
	for name, field := range st.Fields {
		structObj.Fields[strings.ToLower(name)] = field
	}
	env.Set(objtype, structObj)
}

func evalAssignmentStatement(node *ast.AssignmentStatement, env *object.Environment) object.Object {
	// Évaluer la valeur droite
	value := Eval(node.Value, env)
	if isError(value) {
		return value
	}

	// Gérer différentes cibles d'affectation
	switch target := node.Variable.(type) {
	case *ast.Identifier:
		// Assignation simple à une variable env.HasLimits(target.Value)
		if en := env.GetLimitEnv(target.Value); en != nil {
			ok, msg := en.Valid(target.Value, value)
			if !ok {
				return newError(msg+".Line:%d, column:%d", target.Value, target.Line(), target.Column())
			}
		}
		res := env.Set(target.Value, value)
		if res == object.NULL {
			return newError("Invalid name '%s'. Line:%d, column:%d", target.Value, target.Line(), target.Column())
		}
		return value
	case *ast.TypeMember:
		obj, fl := env.Get(target.Left.String())
		if !fl {
			return newError("Invalid structure name '%s'", target.Left.String())
		}
		key := target.Right
		for val, ok := key.(*ast.TypeMember); ok; {
			if ob, o := obj.(*object.Struct); o {
				if vl, exists := ob.Fields[val.Left.String()]; exists {
					obj = vl
					key = val.Right
					continue
				}
				return newError("Invalid field name '%s'. Line:%d, column:%d", key.String(),
					key.Line(), key.Column())
			}
			break
		}
		if obj.Type() == object.STRUCT_OBJ {
			ob := obj.(*object.Struct)
			if _, exists := ob.Fields[key.String()]; exists {

				if en := env.GetLimitEnv(ob.Name + "." + key.String()); en != nil {
					ok, msg := en.Valid(ob.Name+"."+key.String(), value)
					if !ok {
						return newError(msg+".Line:%d, column:%d", key.Line(), key.Column())
					}
				}
				ob.Fields[key.String()] = value
				return value
			}
			return newError("Invalid field name '%s'.", key.String())
		}
		return newError("Invalid type of object '%s'. expected '%v', got '%v'", target.Left.String(),
			object.STRUCT_OBJ, obj.Type())
	case *ast.IndexExpression:
		// Assignation à un élément d'un tableau (ex: arr[0] = x)
		// Évaluer la partie gauche (doit être un tableau ou une structure mutable)
		leftObj := Eval(target.Left, env)
		if isError(leftObj) {
			return leftObj
		}

		indexObj := Eval(target.Index, env)
		if isError(indexObj) {
			return indexObj
		}

		// Supporter les tableaux
		if arr, ok := leftObj.(*object.Array); ok {
			if indexObj.Type() != object.INTEGER_OBJ {
				return newError("L'index doit être un entier, got %s", indexObj.Type())
			}
			idx := indexObj.(*object.Integer).Value
			if idx < 0 || idx >= int64(len(arr.Elements)) {
				return newError("Index hors de portée: %d", idx)
			}
			if en := env.GetLimitEnv(target.Left.String()); en != nil {
				ok, msg := en.Valid(target.Left.String(), value)
				if !ok {
					return newError(msg+".Line:%d, column:%d", target.Left.Line(), target.Left.Column())
				}
			}
			arr.Elements[idx] = value
			return value
		}

		// Si la partie gauche est un identificateur référencant un tableau dans l'environnement,
		// Eval(target.Left, env) retourne la valeur actuelle (déjà traitée ci-dessus).
		// Pour d'autres types, on ne supporte pas l'assignation par index ici.
		return newError("Impossible d'assigner par index sur %s", leftObj.Type())
	default:
		return newError("Cible d'assignation invalide: %T", node.Variable)
	}

}

func defConstraints(ta *ast.TypeAnnotation, env *object.Environment) *object.Limits {
	if ta == nil || ta.Constraints == nil || env == nil {
		return nil
	}
	tc := ta.Constraints
	tp := ta
	if ta.ArrayType != nil && ta.Constraints != nil {
		tp = ta.ArrayType.ElementType
		tc = ta.ArrayType.ElementType.Constraints
	}
	// var result object.Limits
	result := object.Limits{}
	switch strings.ToLower(tp.Type) {
	case "integer":
		result.SetType(object.INTEGER_OBJ)
		if tc.MaxDigits != nil {
			result.Set("MaxDigits", Eval(tc.MaxDigits, env))
		}
		if tc.IntegerRange != nil && tc.IntegerRange.Min != nil {
			result.Set("Min", Eval(tc.IntegerRange.Min, env))
		}
		if tc.IntegerRange != nil && tc.IntegerRange.Max != nil {
			result.Set("Max", Eval(tc.IntegerRange.Max, env))
		}
		return &result
	case "float":
		result.SetType(object.FLOAT_OBJ)
		if tc.MaxDigits != nil {
			result.Set("MaxDigits", Eval(tc.MaxDigits, env))
		}
		if tc.DecimalPlaces != nil {
			result.Set("DecimalPlaces", Eval(tc.MaxDigits, env))
		}
		if tc.IntegerRange != nil && tc.IntegerRange.Min != nil {
			result.Set("Min", Eval(tc.IntegerRange.Min, env))
		}
		if tc.IntegerRange != nil && tc.IntegerRange.Max != nil {
			result.Set("Max", Eval(tc.IntegerRange.Max, env))
		}
		return &result
	case "string":
		result.SetType(object.STRING_OBJ)
		if tc.MaxLength != nil {
			result.Set("MaxLength", Eval(tc.MaxLength, env))
		}
		if tc.IntegerRange != nil && tc.IntegerRange.Min != nil {
			result.Set("Min", Eval(tc.IntegerRange.Min, env))
		}
		if tc.IntegerRange != nil && tc.IntegerRange.Max != nil {
			result.Set("Max", Eval(tc.IntegerRange.Max, env))
		}
		return &result
	case "time":
		result.SetType(object.TIME_OBJ)
		if tc.IntegerRange != nil && tc.IntegerRange.Min != nil {
			result.Set("Min", Eval(tc.IntegerRange.Min, env))
		}
		if tc.IntegerRange != nil && tc.IntegerRange.Max != nil {
			result.Set("Max", Eval(tc.IntegerRange.Max, env))
		}
		return &result
	case "date":
		result.SetType(object.DATE_OBJ)
		if tc.IntegerRange != nil && tc.IntegerRange.Min != nil {
			result.Set("Min", Eval(tc.IntegerRange.Min, env))
		}
		if tc.IntegerRange != nil && tc.IntegerRange.Max != nil {
			result.Set("Max", Eval(tc.IntegerRange.Max, env))
		}
		return &result
	case "duration":
		result.SetType(object.DURATION_OBJ)
		if tc.IntegerRange != nil && tc.IntegerRange.Min != nil {
			result.Set("Min", Eval(tc.IntegerRange.Min, env))
		}
		if tc.IntegerRange != nil && tc.IntegerRange.Max != nil {
			result.Set("Max", Eval(tc.IntegerRange.Max, env))
		}
		return &result
	default:
		return nil
	}
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
	case "datetime":
		return &object.Date{Value: time.Now()}
	case "array":
		return &object.Array{Elements: []object.Object{}}
	case "duration":
		return &object.Duration{Nanoseconds: 0}
	default:
		// Vérifier si c'est un type tableau
		if strings.HasPrefix(typeName, "array") {
			return &object.Array{
				Elements:    []object.Object{},
				ElementType: extractElementType(typeName),
			}
		}
		return object.NULL
	}
}

func evalFunctionStatement(fn *ast.FunctionStatement, env *object.Environment) object.Object {
	function := &object.Function{
		Parameters: fn.Parameters,
		Body:       fn.Body,
		Env:        object.NewEnclosedEnvironment(env),
	}
	if fn.ReturnType != nil {
		env.Limit(fn.Name.Value, defConstraints(fn.ReturnType, env))
	}

	env.Set(fn.Name.Value, function)
	return function
}

func evalStructStatement(st *ast.StructStatement, env *object.Environment) object.Object {
	structObj := &object.Struct{
		Name:   st.Name.Value,
		Fields: make(map[string]object.Object),
	}
	objName := st.Name.Value
	for _, field := range st.Fields {
		structObj.Fields[strings.ToLower(field.Name.Value)] = getDefaultValue(field.Type.Type)
		if field.Type != nil {
			env.Limit(objName+"."+field.Name.Value, defConstraints(field.Type, env))
		}
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

			// Gérer break, continue, fallthrough dans les blocs
			if rt == object.BREAK_OBJ || rt == object.FALLTHROUGH_OBJ ||
				rt == object.CONTINUE_OBJ || rt == object.RETURN_VALUE_OBJ ||
				rt == object.ERROR_OBJ {
				return result
			}
		}
	}

	return result
}

func evalForStatement(forStmt *ast.ForStatement, env *object.Environment) object.Object {
	scope := object.NewEnclosedEnvironment(env)
	bodyEnv := object.NewEnclosedEnvironment(scope)

	if forStmt.Init != nil {
		initResult := Eval(forStmt.Init, scope) //env
		if isError(initResult) {
			return initResult
		}
	}

	for {
		// Évaluer la condition
		if forStmt.Condition != nil {
			condition := Eval(forStmt.Condition, scope) //env
			if isError(condition) {
				return condition
			}

			if !isTruthy(condition) {
				break
			}
		}

		// Évaluer le corps
		bodyEnv.Clear()
		result := evalForBody(forStmt.Body, bodyEnv) //env

		if result != nil {
			rt := result.Type()

			if rt == object.BREAK_OBJ {
				break
			}

			if rt == object.CONTINUE_OBJ {
				// Continue passe à l'itération suivante
				goto update
			}

			if rt == object.RETURN_VALUE_OBJ || rt == object.ERROR_OBJ {
				return result
			}
		}

	update:
		// Évaluer l'update
		if forStmt.Update != nil {
			updateResult := Eval(forStmt.Update, scope) //env
			if isError(updateResult) {
				return updateResult
			}
		}
	}

	return object.NULL
}

func evalDateTimeLiteral(dt *ast.DateTimeLiteral) object.Object {
	// Enlever les # et parser la date/time
	// value := dt.Literal[1 : len(dt.Literal)-1]
	value := dt.Value[1 : len(dt.Value)-1]
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

func toString(selectStmt *ast.SQLSelectStatement, env *object.Environment) object.Object {
	strSelect := ""
	for _, ex := range selectStmt.Select {
		if strSelect == "" {
			strSelect = ex.String()
		} else {
			strSelect = fmt.Sprintf("%s, %s", strSelect, ex.String())
		}
	}
	from := evalFromClause(selectStmt.From, env)
	strFrom := selectStmt.From.String()
	if isError(from) {
		return from
	}
	// Traiter la champ Join avant de passer a la clause where
	// puis executer la requete SQL et charger les resultats
	for _, step := range selectStmt.Joins {
		from = evalFromClause(step.Table, env)
		strFrom = fmt.Sprintf("%s %s JOIN %s ON %s", strFrom, step.Type, step.Table.String(), step.On.String())
	}

	strSQL := fmt.Sprintf("SELECT %s\nFROM %s", strSelect, strFrom)
	// Traiter la clause WHERE
	if selectStmt.Where != nil {
		whereResult := Eval(selectStmt.Where, env)
		strSQL = fmt.Sprintf("%s\nWHERE (%s)", strSQL, whereResult.Inspect())
		if isError(whereResult) {
			return whereResult
		}
	}
	if selectStmt.GroupBy != nil {
		strGroup := ""
		for k, v := range selectStmt.GroupBy {
			if k == 0 {
				strGroup = v.String()
				continue
			}
			strGroup = fmt.Sprintf("%s, %s", strGroup, v.String())
		}
		strSQL = fmt.Sprintf("%s\nGROUP BY \n%s", strSQL, strGroup)
	}
	if selectStmt.Having != nil {
		strHaving := Eval(selectStmt.Having, env)
		strSQL = fmt.Sprintf("%s\nHAVING(%s)", strSQL, strHaving.Inspect())
	}
	if selectStmt.Union != nil {
		strSQL = fmt.Sprintf("%s\nUNION\n (%s)", strSQL, toString(selectStmt.Union, env))
	}
	return &object.String{Value: strSQL}
}

func evalSQLSelectStatement(selectStmt *ast.SQLSelectStatement, env *object.Environment) object.Object {
	// Implémentation simplifiée pour la démonstration
	// Dans une vraie implémentation, cela interagirait avec une base de données

	result := &object.SQLResult{
		Columns: []string{},
		Rows:    []map[string]object.Object{},
	}

	// Traiter la clause SELECT
	for _, ex := range selectStmt.Select {

		if e, True := ex.(*ast.SelectArgs); True {
			if _, ok := e.Expr.(*ast.TypeMember); ok {
				result.Columns = append(result.Columns, e.String())
				continue
			}
			if ident, ok := e.Expr.(*ast.Identifier); ok { //expr.(*ast.Identifier)
				if ident.Value == "*" {
					// Sélectionner toutes les colonnes
					// Implémentation simplifiée
					continue
				} else {
					result.Columns = append(result.Columns, ident.Value)
				}
			}
		}

	}

	// Traiter la clause FROM
	// Build SQL String to run in the database
	// strSelect := toString(selectStmt)
	toString(selectStmt, env)
	//Executer la requete SQL puis retourner le resultat.
	return result
}

func evalFromClause(from ast.Expression, env *object.Environment) object.Object {
	switch from := from.(type) {
	case *ast.Identifier:
		// Rechercher l'objet dans l'environnement
		obj, ok := env.Get(strings.ToLower(from.Value))
		if !ok {
			return newError("Objet non trouvé: %s", from.Value)
		}
		return obj
	case *ast.FromIdentifier:
		// n := Eval(from.Value, env)
		obj, ok := env.Get(strings.ToLower(from.Value.String()))
		if !ok {
			return newError("Objet non trouvé: %s", from.Value)
		}
		if from.NewName != nil {
			env.Set(from.NewName.String(), obj)
		}
		return obj
	default:
		return Eval(from, env)
	}
}

// Les fonctions evalPrefixExpression, evalInfixExpression, evalIdentifier, etc.
// suivent le même pattern que dans un évaluateur standard...

func evalStructLiteral(node *ast.StructLiteral, env *object.Environment) object.Object {
	// Créer un objet struct littéral avec ses champs évalués
	structObj := &object.Struct{
		Name:   "",
		Fields: make(map[string]object.Object),
	}

	for _, f := range node.Fields {
		val := Eval(f.Value, env)
		if isError(val) {
			return val
		}
		structObj.Fields[strings.ToLower(f.Name.Value)] = val
	}
	objtype := env.IsStructExist(structObj, env)
	if objtype == "" {
		struct_id++
		objtype = fmt.Sprintf("%s_%d", "struct_id", struct_id)
		defType(objtype, structObj, env)
	} else { //check if all values are ok
		for k, v := range structObj.Fields {
			en := env.GetLimitEnv(k)
			if en != nil {
				ok, msg := en.Valid(k, v)
				if !ok {
					return newError(msg+". Line:%d, column:%d", node.Line(), node.Column())
				}
			}
		}
	}
	structObj.Name = objtype
	return structObj
}

func evalIfStatement(node *ast.IfStatement, env *object.Environment) object.Object {
	// Évaluer la condition
	condition := Eval(node.Condition, env)
	if isError(condition) {
		return condition
	}
	scope := object.NewEnclosedEnvironment(env)
	// Si la condition est vraie, évaluer le bloc conséquence
	if isTruthy(condition) {
		return evalBlockStatement(node.Then, scope)
	}

	// Sinon, si une alternative existe, l'évaluer
	if node.Else != nil {
		return evalBlockStatement(node.Else, scope)
	}

	// Par défaut, retourner NULL
	return object.NULL
}

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
	case (left.Type() == object.FLOAT_OBJ || left.Type() == object.INTEGER_OBJ) && (right.Type() == object.INTEGER_OBJ ||
		right.Type() == object.FLOAT_OBJ):
		return evalFloatInfixExpression(operator, left, right)
	case left.Type() == object.STRING_OBJ || right.Type() == object.STRING_OBJ:
		return evalStringInfixExpression(operator, left, right)
	case operator == "==":
		return &object.Boolean{Value: left == right}
	case operator == "!=":
		return &object.Boolean{Value: left != right}
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalIntegerInfixExpression(operator, left, right)
	case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
		return evalStringInfixExpression(operator, left, right)
	case left.Type() == object.ARRAY_OBJ && right.Type() == object.ARRAY_OBJ:
		return evalArrayInfixExpression(operator, left, right)
	case operator == "==":
		return &object.Boolean{Value: objectsEqual(left, right)}
	case operator == "!=":
		return &object.Boolean{Value: !objectsEqual(left, right)}
	case left.Type() == object.DURATION_OBJ && right.Type() == object.DURATION_OBJ:
		return evalDurationInfixExpression(operator, left, right)
	case left.Type() == object.DURATION_OBJ && (right.Type() == object.INTEGER_OBJ || right.Type() == object.FLOAT_OBJ):
		return evalDurationInfixExpression(operator, left, right)
	case (left.Type() == object.INTEGER_OBJ || left.Type() == object.FLOAT_OBJ) && right.Type() == object.DURATION_OBJ:
		return evalDurationInfixExpression(operator, left, right)
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
	case "%":
		return &object.Integer{Value: leftVal % rightVal}
	case "<":
		return &object.Boolean{Value: leftVal < rightVal}
	case "<=":
		return &object.Boolean{Value: leftVal <= rightVal}
	case ">":
		return &object.Boolean{Value: leftVal > rightVal}
	case ">=":
		return &object.Boolean{Value: leftVal >= rightVal}
	case "==":
		return &object.Boolean{Value: leftVal == rightVal}
	case "!=":
		return &object.Boolean{Value: leftVal != rightVal}
	default:
		return newError("Opérateur inconnu: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalFloatInfixExpression(operator string, left, right object.Object) object.Object {
	var rightVal, leftVal float64
	_, ok := left.(*object.Float)
	if ok {
		leftVal = left.(*object.Float).Value
	} else {
		leftVal = float64(left.(*object.Integer).Value)
	}
	_, ok = right.(*object.Float)
	if ok {
		rightVal = right.(*object.Float).Value
	} else {
		rightVal = float64(right.(*object.Integer).Value)
	}
	switch operator {
	case "+":
		return &object.Float{Value: leftVal + rightVal}
	case "-":
		return &object.Float{Value: leftVal - rightVal}
	case "*":
		return &object.Float{Value: leftVal * rightVal}
	case "/":
		return &object.Float{Value: leftVal / rightVal}
	case "<":
		return &object.Boolean{Value: leftVal < rightVal}
	case "<=":
		return &object.Boolean{Value: leftVal <= rightVal}
	case ">":
		return &object.Boolean{Value: leftVal > rightVal}
	case ">=":
		return &object.Boolean{Value: leftVal >= rightVal}
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
	// leftVal := left.Inspect()
	// rightVal := right.Inspect()
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

	return newError("%s", "Identifiant non trouvé: "+node.Value)
}

func isTruthy(obj object.Object) bool {
	if obj == nil || obj == object.NULL {
		return false
	}
	switch v := obj.(type) {
	case *object.Boolean:
		return v.Value
	default:
		return false
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

func evalSQLCreateObject(stmt *ast.SQLCreateObjectStatement, env *object.Environment) object.Object {
	// Créer une nouvelle table/objet dans l'environnement
	table := &object.SQLTable{
		Name:    stmt.ObjectName.Value,
		Columns: make(map[string]*object.SQLColumn),
		Data:    []map[string]object.Object{},
		Indexes: make(map[string]*object.SQLIndex),
	}

	// Traiter les définitions de colonnes
	for _, colDef := range stmt.Columns {
		column := &object.SQLColumn{
			Name: colDef.Name.Value,
			Type: colDef.DataType.Name,
		}

		// Appliquer les contraintes
		for _, constraint := range colDef.Constraints {
			switch constraint.Type {
			case "PRIMARY KEY":
				column.PrimaryKey = true
			case "NOT NULL":
				column.NotNull = true
			case "UNIQUE":
				column.Unique = true
			case "DEFAULT":
				column.DefaultValue = Eval(constraint.Expression, env)
			}
		}

		table.Columns[colDef.Name.Value] = column
	}

	// Stocker la table dans l'environnement
	env.Set(stmt.ObjectName.Value, table)
	strSQL := stmt.String()
	res, err := env.Exec(strSQL)
	if err == nil {
		r, _ := res.RowsAffected()
		return &object.SQLResult{
			Message:      fmt.Sprintf("OBJECT %s créé avec succès", stmt.ObjectName.Value),
			RowsAffected: r,
		}
	}
	return newError("Nsina: %s", err.Error())
}

func evalSQLDropObject(stmt *ast.SQLDropObjectStatement, env *object.Environment) object.Object {
	// Vérifier si l'objet existe
	if _, ok := env.Get(stmt.ObjectName.Value); !ok {
		if stmt.IfExists {
			return &object.SQLResult{
				Message:      "Aucun objet à supprimer",
				RowsAffected: 0,
			}
		}
		return newError("L'objet %s n'existe pas", stmt.ObjectName.Value)
	}

	// Supprimer l'objet de l'environnement
	env.Set(stmt.ObjectName.Value, object.NULL)

	return &object.SQLResult{
		Message:      fmt.Sprintf("OBJECT %s supprimé avec succès", stmt.ObjectName.Value),
		RowsAffected: 0,
	}
}

func evalSQLAlterObject(stmt *ast.SQLAlterObjectStatement, env *object.Environment) object.Object {
	// Récupérer l'objet existant
	obj, ok := env.Get(stmt.ObjectName.Value)
	if !ok {
		return newError("L'objet %s n'existe pas", stmt.ObjectName.Value)
	}

	table, ok := obj.(*object.SQLTable)
	if !ok {
		return newError("%s n'est pas un OBJECT", stmt.ObjectName.Value)
	}

	// Appliquer les actions ALTER
	for _, action := range stmt.Actions {
		switch action.Type {
		case "ADD":
			if action.Column != nil {
				// Ajouter une nouvelle colonne
				column := &object.SQLColumn{
					Name: action.Column.Name.Value,
					Type: action.Column.DataType.Name,
				}
				table.Columns[action.Column.Name.Value] = column

				// Mettre à jour les données existantes avec la valeur par défaut
				defaultValue := getDefaultSQLValue(action.Column.DataType.Name)
				for i := range table.Data {
					table.Data[i][action.Column.Name.Value] = defaultValue
				}
			}
		case "MODIFY":
			// Modifier une colonne existante
			// Implémentation simplifiée
		case "DROP":
			if action.ColumnName != nil {
				// Supprimer une colonne
				delete(table.Columns, action.ColumnName.Value)
				for i := range table.Data {
					delete(table.Data[i], action.ColumnName.Value)
				}
			}
		}
	}

	return &object.SQLResult{
		Message:      fmt.Sprintf("OBJECT %s modifié avec succès", stmt.ObjectName.Value),
		RowsAffected: 0,
	}
}

func evalSQLInsert(stmt *ast.SQLInsertStatement, env *object.Environment) object.Object {
	// Récupérer l'objet
	obj, ok := env.Get(stmt.ObjectName.Value)
	if !ok {
		return newError("L'objet %s n'existe pas", stmt.ObjectName.Value)
	}

	table, ok := obj.(*object.SQLTable)
	if !ok {
		return newError("%s n'est pas un OBJECT", stmt.ObjectName.Value)
	}

	rowsAffected := 0

	if stmt.Select != nil {
		// INSERT ... SELECT
		result := Eval(stmt.Select, env).(*object.SQLResult)
		// Implémentation de l'insertion des résultats
		rowsAffected = len(result.Rows)
	} else {
		// INSERT ... VALUES
		for _, values := range stmt.Values {
			row := make(map[string]object.Object)

			for i, value := range values.Values {
				var colName string
				if i < len(stmt.Columns) {
					colName = stmt.Columns[i].Value
				} else {
					// Utiliser l'ordre des colonnes de la table
					colIndex := 0
					for name := range table.Columns {
						if colIndex == i {
							colName = name
							break
						}
						colIndex++
					}
				}

				evaluatedValue := Eval(value, env)
				row[colName] = evaluatedValue
			}

			table.Data = append(table.Data, row)
			rowsAffected++
		}
	}

	return &object.SQLResult{
		Message:      fmt.Sprintf("%d ligne(s) insérée(s)", rowsAffected),
		RowsAffected: int64(rowsAffected),
	}
}

func evalSQLUpdate(stmt *ast.SQLUpdateStatement, env *object.Environment) object.Object {
	obj, ok := env.Get(stmt.ObjectName.Value)
	if !ok {
		return newError("L'objet %s n'existe pas", stmt.ObjectName.Value)
	}

	table, ok := obj.(*object.SQLTable)
	if !ok {
		return newError("%s n'est pas un OBJECT", stmt.ObjectName.Value)
	}

	rowsAffected := 0

	for _, row := range table.Data {
		// Créer un environnement local pour la ligne
		rowEnv := object.NewEnclosedEnvironment(env)
		for colName, value := range row {
			rowEnv.Set(colName, value)
		}

		// Évaluer la condition WHERE
		shouldUpdate := true
		if stmt.Where != nil {
			condition := Eval(stmt.Where, rowEnv)
			if isError(condition) {
				return condition
			}
			shouldUpdate = isTruthy(condition)
		}

		if shouldUpdate {
			// Appliquer les modifications SET
			for _, setClause := range stmt.Set {
				newValue := Eval(setClause.Value, rowEnv)
				if isError(newValue) {
					return newValue
				}
				row[setClause.Column.Value] = newValue
			}
			rowsAffected++
		}
	}

	return &object.SQLResult{
		Message:      fmt.Sprintf("%d ligne(s) modifiée(s)", rowsAffected),
		RowsAffected: int64(rowsAffected),
	}
}

func evalSQLDelete(stmt *ast.SQLDeleteStatement, env *object.Environment) object.Object {
	obj, ok := env.Get(stmt.From.Value)
	if !ok {
		return newError("L'objet %s n'existe pas", stmt.From.Value)
	}

	table, ok := obj.(*object.SQLTable)
	if !ok {
		return newError("%s n'est pas un OBJECT", stmt.From.Value)
	}

	rowsAffected := 0
	var newData []map[string]object.Object

	for _, row := range table.Data {
		// Créer un environnement local pour la ligne
		rowEnv := object.NewEnclosedEnvironment(env)
		for colName, value := range row {
			rowEnv.Set(colName, value)
		}

		// Évaluer la condition WHERE
		shouldDelete := true
		if stmt.Where != nil {
			condition := Eval(stmt.Where, rowEnv)
			if isError(condition) {
				return condition
			}
			shouldDelete = isTruthy(condition)
		}

		if shouldDelete {
			rowsAffected++
		} else {
			newData = append(newData, row)
		}
	}

	table.Data = newData

	return &object.SQLResult{
		Message:      fmt.Sprintf("%d ligne(s) supprimée(s)", rowsAffected),
		RowsAffected: int64(rowsAffected),
	}
}

func evalSQLTruncate(stmt *ast.SQLTruncateStatement, env *object.Environment) object.Object {
	obj, ok := env.Get(stmt.ObjectName.Value)
	if !ok {
		return newError("L'objet %s n'existe pas", stmt.ObjectName.Value)
	}

	table, ok := obj.(*object.SQLTable)
	if !ok {
		return newError("%s n'est pas un OBJECT", stmt.ObjectName.Value)
	}

	rowsAffected := len(table.Data)
	table.Data = []map[string]object.Object{}

	return &object.SQLResult{
		Message:      fmt.Sprintf("OBJECT %s vidé (%d ligne(s) supprimée(s))", stmt.ObjectName.Value, rowsAffected),
		RowsAffected: int64(rowsAffected),
	}
}

func evalSQLCreateIndex(stmt *ast.SQLCreateIndexStatement, env *object.Environment) object.Object {
	obj, ok := env.Get(stmt.ObjectName.Value)
	if !ok {
		return newError("L'objet %s n'existe pas", stmt.ObjectName.Value)
	}

	table, ok := obj.(*object.SQLTable)
	if !ok {
		return newError("%s n'est pas un OBJECT", stmt.ObjectName.Value)
	}

	// Créer l'index
	index := &object.SQLIndex{
		Name:    stmt.IndexName.Value,
		Columns: make([]string, len(stmt.Columns)),
		Unique:  stmt.Unique,
	}

	for i, col := range stmt.Columns {
		index.Columns[i] = col.Value
	}

	table.Indexes[stmt.IndexName.Value] = index

	return &object.SQLResult{
		Message:      fmt.Sprintf("INDEX %s créé avec succès", stmt.IndexName.Value),
		RowsAffected: 0,
	}
}

func getDefaultSQLValue(dataType string) object.Object {
	switch dataType {
	case "integer", "int":
		return &object.Integer{Value: 0}
	case "float", "numeric", "decimal":
		return &object.Float{Value: 0.0}
	case "varchar", "char", "text":
		return &object.String{Value: ""}
	case "boolean", "bool":
		return object.FALSE
	case "date":
		return &object.Date{Value: time.Now()}
	case "time", "timestamp", "datetime":
		return &object.Time{Value: time.Now()}
	default:
		return object.NULL
	}
}

func evalSQLWithStatement(stmt *ast.SQLWithStatement, env *object.Environment) object.Object {
	// Créer un nouvel environnement pour les CTE
	cteEnv := object.NewEnclosedEnvironment(env)

	// Évaluer les CTEs
	for _, cte := range stmt.CTEs {
		result := evalCommonTableExpression(cte, cteEnv)
		if isError(result) {
			return result
		}
		cteEnv.Set(cte.Name.Value, result)
	}

	// Évaluer la requête principale dans l'environnement avec CTE
	return Eval(stmt.Select, cteEnv)
}

func evalCommonTableExpression(cte *ast.SQLCommonTableExpression, env *object.Environment) object.Object {
	// Évaluer la requête CTE
	result := Eval(cte.Query, env)
	if isError(result) {
		return result
	}

	// Stocker le résultat comme table temporaire
	if sqlResult, ok := result.(*object.SQLResult); ok {
		table := &object.SQLTable{
			Name:    cte.Name.Value,
			Columns: make(map[string]*object.SQLColumn),
			Data:    sqlResult.Rows,
		}

		// Définir les colonnes
		for i, colName := range sqlResult.Columns {
			if i < len(cte.Columns) {
				colName = cte.Columns[i].Value
			}
			table.Columns[colName] = &object.SQLColumn{
				Name: colName,
				Type: "dynamic", // Type déterminé dynamiquement
			}
		}

		return table
	}

	return newError("Le CTE doit retourner un résultat SQL")
}

func evalRecursiveCTE(cte *ast.SQLRecursiveCTE, env *object.Environment) object.Object {
	// Évaluer la partie anchor
	anchorResult := Eval(cte.Anchor, env)
	if isError(anchorResult) {
		return anchorResult
	}

	anchorTable, ok := anchorResult.(*object.SQLTable)
	if !ok {
		return newError("L'anchor doit retourner une table")
	}

	// Créer la table de résultat
	resultTable := &object.SQLTable{
		Name:    cte.Name.Value,
		Columns: anchorTable.Columns,
		Data:    make([]map[string]object.Object, 0),
	}

	// Ajouter les données anchor
	resultTable.Data = append(resultTable.Data, anchorTable.Data...)

	// Itération récursive
	maxIterations := 1000 // Limite de sécurité
	iteration := 0
	previousCount := 0

	for iteration < maxIterations {
		// Créer un environnement avec les données actuelles
		recursiveEnv := object.NewEnclosedEnvironment(env)
		recursiveEnv.Set(cte.Name.Value, resultTable)

		// Évaluer la partie récursive
		recursiveResult := Eval(cte.Recursive, recursiveEnv)
		if isError(recursiveResult) {
			return recursiveResult
		}

		recursiveTable, ok := recursiveResult.(*object.SQLTable)
		if !ok {
			return newError("La partie récursive doit retourner une table")
		}

		// Vérifier si on a de nouvelles données
		if len(recursiveTable.Data) == 0 {
			break // Point fixe atteint
		}

		// Ajouter les nouvelles données (éviter les doublons)
		for _, newRow := range recursiveTable.Data {
			if !containsRow(resultTable.Data, newRow) {
				resultTable.Data = append(resultTable.Data, newRow)
			}
		}

		// Vérifier la convergence
		if len(resultTable.Data) == previousCount {
			break // Aucun nouveau row ajouté
		}

		previousCount = len(resultTable.Data)
		iteration++
	}

	if iteration >= maxIterations {
		return newError("Limite d'itérations récursives atteinte")
	}

	return &object.SQLResult{
		Columns:      getColumnNames(resultTable),
		Rows:         resultTable.Data,
		RowsAffected: int64(len(resultTable.Data)),
	}
}

func evalHierarchicalQuery(selectStmt *ast.SQLSelectStatement, env *object.Environment) object.Object {
	// Récupérer la table source
	fromResult := Eval(selectStmt.From, env)
	if isError(fromResult) {
		return fromResult
	}

	sourceTable, ok := fromResult.(*object.SQLTable)
	if !ok {
		return newError("La source doit être une table")
	}

	// Construire l'arbre hiérarchique
	tree := buildHierarchicalTree(sourceTable, selectStmt.Hierarchical, env)
	if isError(tree) {
		return tree
	}

	// Parcourir l'arbre et construire le résultat
	resultRows := traverseHierarchicalTree(tree, selectStmt, env)

	return &object.SQLResult{
		Columns:      getColumnNames(sourceTable),
		Rows:         resultRows,
		RowsAffected: int64(len(resultRows)),
	}
}

func buildHierarchicalTree(table *object.SQLTable, hierarchical *ast.SQLHierarchicalQuery, env *object.Environment) object.Object {
	// Implémentation simplifiée de la construction d'arbre
	// Dans une implémentation réelle, cela utiliserait les clauses
	// START WITH et CONNECT BY pour construire la hiérarchie

	tree := &object.HierarchicalTree{
		Nodes: make(map[string]*object.HierarchicalNode),
	}

	// Identifier la colonne clé et parent
	keyColumn := "id"
	parentColumn := "parent_id"

	// Construire les nœuds
	for i, row := range table.Data {
		node := &object.HierarchicalNode{
			Data:     row,
			Level:    0,
			Children: []*object.HierarchicalNode{},
		}

		if id, ok := row[keyColumn]; ok {
			node.ID = id.Inspect()
		} else {
			node.ID = fmt.Sprintf("node_%d", i)
		}

		tree.Nodes[node.ID] = node
	}

	// Construire les relations parent-enfant
	for _, node := range tree.Nodes {
		if parentID, ok := node.Data[parentColumn]; ok {
			if parentNode, exists := tree.Nodes[parentID.Inspect()]; exists {
				parentNode.Children = append(parentNode.Children, node)
				node.Parent = parentNode
			}
		} else {
			// Nœud racine
			tree.Roots = append(tree.Roots, node)
		}
	}

	// Calculer les niveaux
	for _, root := range tree.Roots {
		calculateLevels(root, 0)
	}

	return tree
}

func calculateLevels(node *object.HierarchicalNode, level int) {
	node.Level = level
	for _, child := range node.Children {
		calculateLevels(child, level+1)
	}
}

func traverseHierarchicalTree(treeObj object.Object, selectStmt *ast.SQLSelectStatement, env *object.Environment) []map[string]object.Object {
	tree, ok := treeObj.(*object.HierarchicalTree)
	if !ok {
		return []map[string]object.Object{}
	}

	var result []map[string]object.Object

	// Parcours en profondeur d'abord
	for _, root := range tree.Roots {
		traverseNode(root, &result, selectStmt, env)
	}

	return result
}

func traverseNode(node *object.HierarchicalNode, result *[]map[string]object.Object, selectStmt *ast.SQLSelectStatement, env *object.Environment) {
	// Ajouter le nœud courant
	row := make(map[string]object.Object)
	for k, v := range node.Data {
		row[k] = v
	}

	// Ajouter les colonnes hiérarchiques
	row["level"] = &object.Integer{Value: int64(node.Level)}
	if node.Parent != nil {
		if parentID, ok := node.Parent.Data["id"]; ok {
			row["parent_id"] = parentID
		}
	}

	*result = append(*result, row)

	// Parcourir les enfants
	for _, child := range node.Children {
		traverseNode(child, result, selectStmt, env)
	}
}

func evalWindowFunction(function *ast.SQLWindowFunction, env *object.Environment) object.Object {
	// Implémentation simplifiée des fonctions de fenêtrage
	switch function.Name {
	case "ROW_NUMBER":
		return evalRowNumber(function, env)
	case "RANK":
		return evalRank(function, env)
	case "DENSE_RANK":
		return evalDenseRank(function, env)
	case "LAG":
		return evalLag(function, env)
	case "LEAD":
		return evalLead(function, env)
	default:
		return newError("Fonction de fenêtrage non supportée: %s", function.Name)
	}
}

func evalRowNumber(function *ast.SQLWindowFunction, env *object.Environment) object.Object {
	// Dans une implémentation réelle, cela calculerait le numéro de ligne
	// dans la partition et l'ordre définis
	return &object.Integer{Value: 1}
}

func evalRank(function *ast.SQLWindowFunction, env *object.Environment) object.Object {
	// Dans une implémentation réelle, cela calculerait le rank
	// dans la partition et l'ordre définis
	return &object.Integer{Value: 1}
}

func evalDenseRank(function *ast.SQLWindowFunction, env *object.Environment) object.Object {
	// Dans une implémentation réelle, cela calculerait le rank
	// dans la partition et l'ordre définis
	return &object.Integer{Value: 1}
}

func evalLag(function *ast.SQLWindowFunction, env *object.Environment) object.Object {
	// Dans une implémentation réelle, cela calculerait le rank
	// dans la partition et l'ordre définis
	return &object.Integer{Value: 1}
}

func evalLead(function *ast.SQLWindowFunction, env *object.Environment) object.Object {
	// Dans une implémentation réelle, cela calculerait le rank
	// dans la partition et l'ordre définis
	return &object.Integer{Value: 1}
}

// Fonctions utilitaires
func containsRow(rows []map[string]object.Object, row map[string]object.Object) bool {
	for _, existingRow := range rows {
		if rowsEqual(existingRow, row) {
			return true
		}
	}
	return false
}

func rowsEqual(row1, row2 map[string]object.Object) bool {
	if len(row1) != len(row2) {
		return false
	}

	for k, v1 := range row1 {
		if v2, exists := row2[k]; !exists || v1.Inspect() != v2.Inspect() {
			return false
		}
	}

	return true
}

func getColumnNames(table *object.SQLTable) []string {
	var columns []string
	for colName := range table.Columns {
		columns = append(columns, colName)
	}
	return columns
}

func evalArrayLiteral(node *ast.ArrayLiteral, env *object.Environment) object.Object {
	elements := make([]object.Object, len(node.Elements))

	for i, element := range node.Elements {
		evaluated := Eval(element, env)
		if isError(evaluated) {
			return evaluated
		}
		elements[i] = evaluated
	}

	return &object.Array{Elements: elements}
}

func evalIndexExpression(node *ast.IndexExpression, env *object.Environment) object.Object {
	left := Eval(node.Left, env)
	if isError(left) {
		return left
	}

	index := Eval(node.Index, env)
	if isError(index) {
		return index
	}

	switch {
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		return evalArrayIndexExpression(left, index)
	case left.Type() == object.STRING_OBJ && index.Type() == object.INTEGER_OBJ:
		return evalStringIndexExpression(left, index)
	default:
		return newError("Opération d'indexation non supportée: %s[%s]",
			left.Type(), index.Type())
	}
}

func evalArrayIndexExpression(array, index object.Object) object.Object {
	arrayObject := array.(*object.Array)
	idx := index.(*object.Integer).Value
	max := int64(len(arrayObject.Elements) - 1)

	if idx < 0 || idx > max {
		return object.NULL
	}

	return arrayObject.Elements[idx]
}

func evalStringIndexExpression(str, index object.Object) object.Object {
	strObject := str.(*object.String)
	idx := index.(*object.Integer).Value

	if idx < 0 || idx >= int64(len(strObject.Value)) {
		return object.NULL
	}

	return &object.String{Value: string(strObject.Value[idx])}
}

func evalSliceExpression(node *ast.SliceExpression, env *object.Environment) object.Object {
	left := Eval(node.Left, env)
	if isError(left) {
		return left
	}

	var start, end int64

	if node.Start != nil {
		startObj := Eval(node.Start, env)
		if isError(startObj) {
			return startObj
		}
		if startObj.Type() != object.INTEGER_OBJ {
			return newError("L'index de début doit être un entier, got %s", startObj.Type())
		}
		start = startObj.(*object.Integer).Value
	} else {
		start = 0
	}

	if node.End != nil {
		endObj := Eval(node.End, env)
		if isError(endObj) {
			return endObj
		}
		if endObj.Type() != object.INTEGER_OBJ {
			return newError("L'index de fin doit être un entier, got %s", endObj.Type())
		}
		end = endObj.(*object.Integer).Value
	}

	switch left := left.(type) {
	case *object.Array:
		return evalArraySlice(left, start, end)
	case *object.String:
		return evalStringSlice(left, start, end)
	default:
		return newError("Opération de slice non supportée sur %s", left.Type())
	}
}

func evalArraySlice(array *object.Array, start, end int64) object.Object {
	length := int64(len(array.Elements))

	// Gestion des index négatifs
	if start < 0 {
		start = length + start
		if start < 0 {
			start = 0
		}
	}

	if end <= 0 {
		end = length + end
	}

	if end > length {
		end = length
	}

	if start > end {
		start = end
	}

	if start < 0 || start >= length || end < 0 {
		return &object.Array{Elements: []object.Object{}}
	}

	result := make([]object.Object, end-start)
	copy(result, array.Elements[start:end])

	return &object.Array{Elements: result}
}

func evalStringSlice(str *object.String, start, end int64) object.Object {
	length := int64(len(str.Value))

	// Gestion des index négatifs
	if start < 0 {
		start = length + start
		if start < 0 {
			start = 0
		}
	}

	if end <= 0 {
		end = length + end
	}

	if end > length {
		end = length
	}

	if start > end {
		start = end
	}

	if start < 0 || start >= length || end < 0 {
		return &object.String{Value: ""}
	}

	return &object.String{Value: str.Value[start:end]}
}

func evalInExpression(node *ast.InExpression, env *object.Environment) object.Object {
	left := Eval(node.Left, env)
	if isError(left) {
		return left
	}

	right := Eval(node.Right, env)
	if isError(right) {
		return right
	}

	var contains bool

	switch right := right.(type) {
	case *object.Array:
		contains = arrayContains(right, left)
	case *object.String:
		if left.Type() != object.STRING_OBJ {
			return newError("L'opérateur IN sur les chaînes nécessite une chaîne à gauche")
		}
		leftStr := left.(*object.String).Value
		rightStr := right.Value
		contains = strings.Contains(rightStr, leftStr)
	default:
		return newError("L'opérande droit de IN doit être un tableau ou une chaîne, got %s", right.Type())
	}

	if node.Not {
		return &object.Boolean{Value: !contains}
	}
	return &object.Boolean{Value: contains}
}

func arrayContains(array *object.Array, element object.Object) bool {
	for _, el := range array.Elements {
		if objectsEqual(el, element) {
			return true
		}
	}
	return false
}

func objectsEqual(a, b object.Object) bool {
	if a.Type() != b.Type() {
		if (a.Type() == object.INTEGER_OBJ && b.Type() == object.FLOAT_OBJ) ||
			(b.Type() == object.INTEGER_OBJ && a.Type() == object.FLOAT_OBJ) {
			var v, w float64
			if s, ok := a.(*object.Integer); ok {
				v = float64(s.Value)
			}
			v = a.(*object.Float).Value
			if s, ok := b.(*object.Integer); ok {
				w = float64(s.Value)
			}
			w = b.(*object.Float).Value
			return v == w
		}
		return false
	}

	switch a := a.(type) {
	case *object.Integer:
		b := b.(*object.Integer)
		return a.Value == b.Value
	case *object.Float:
		b := b.(*object.Float)
		return a.Value == b.Value
	case *object.String:
		b := b.(*object.String)
		return a.Value == b.Value
	case *object.Boolean:
		b := b.(*object.Boolean)
		return a.Value == b.Value
	default:
		return a == b
	}
}

func evalArrayFunctionCall(node *ast.ArrayFunctionCall, env *object.Environment) object.Object {
	//How to save the context before running the function
	f, ok := env.Get(node.Function.Value)
	if ok && f.Type() == object.FUNCTION_OBJ {
		//run the function
		ob := f.(*object.Function)
		for k, field := range ob.Parameters {
			if k == 0 {
				val := Eval(node.Array, env)
				if isError(val) {
					return val
				}
				ob.Env.Set(field.Name.Value, val)
				continue
			}
			val := Eval(node.Arguments[k-1], env)
			if isError(val) {
				return val
			}
			ob.Env.Set(field.Name.Value, val)
		}
		val := Eval(ob.Body, ob.Env)
		if val == nil {
			return nil
		}
		return val.(*object.ReturnValue).Value
	}
	switch strings.ToLower(node.Function.Value) {
	case "tostring":
		//TODO: A definir
		val := Eval(node.Array, env)
		if isError(val) {
			return val
		}
		return &object.String{Value: val.Inspect()}
	}
	array := Eval(node.Array, env)
	if isError(array) {
		return array
	}

	if array.Type() != object.ARRAY_OBJ {
		switch strings.ToLower(node.Function.Value) {
		case "len", "length":
			return &object.Integer{Value: int64(len(array.Inspect()))}
		case "tostring":
			return &object.String{Value: array.Inspect()}
		}
		return newError("La fonction %s attend un tableau en argument", node.Function.Value)
	}

	arr := array.(*object.Array)

	switch strings.ToLower(node.Function.Value) {
	case "length":
		return &object.Integer{Value: int64(len(arr.Elements))}
	case "len":
		return &object.Integer{Value: int64(len(arr.Elements))}
	case "append":
		if len(node.Arguments) != 1 {
			return newError("La fonction append attend exactement 1 argument")
		}
		element := Eval(node.Arguments[0], env)
		if isError(element) {
			return element
		}
		return arrayAppend(arr, element)

	case "prepend":
		if len(node.Arguments) != 1 {
			return newError("La fonction prepend attend exactement 1 argument")
		}
		element := Eval(node.Arguments[0], env)
		if isError(element) {
			return element
		}
		return arrayPrepend(arr, element)

	case "remove":
		if len(node.Arguments) != 1 {
			return newError("La fonction remove attend exactement 1 argument")
		}
		index := Eval(node.Arguments[0], env)
		if isError(index) {
			return index
		}
		if index.Type() != object.INTEGER_OBJ {
			return newError("L'index pour remove doit être un entier")
		}
		return arrayRemove(arr, index.(*object.Integer).Value)

	case "slice":
		if len(node.Arguments) > 2 {
			return newError("La fonction slice attend 1 ou 2 arguments")
		}

		var start, end int64
		if len(node.Arguments) >= 1 {
			startObj := Eval(node.Arguments[0], env)
			if isError(startObj) {
				return startObj
			}
			if startObj.Type() != object.INTEGER_OBJ {
				return newError("Le début du slice doit être un entier")
			}
			start = startObj.(*object.Integer).Value
		}

		if len(node.Arguments) == 2 {
			endObj := Eval(node.Arguments[1], env)
			if isError(endObj) {
				return endObj
			}
			if endObj.Type() != object.INTEGER_OBJ {
				return newError("La fin du slice doit être un entier")
			}
			end = endObj.(*object.Integer).Value
		} else {
			end = int64(len(arr.Elements))
		}

		return evalArraySlice(arr, start, end)

	case "contains":
		if len(node.Arguments) != 1 {
			return newError("La fonction contains attend exactement 1 argument")
		}
		element := Eval(node.Arguments[0], env)
		if isError(element) {
			return element
		}
		return &object.Boolean{Value: arrayContains(arr, element)}

	default:
		return newError("Fonction de tableau inconnue: %s", node.Function.Value)
	}
}

func arrayAppend(arr *object.Array, element object.Object) object.Object {
	newElements := make([]object.Object, len(arr.Elements)+1)
	copy(newElements, arr.Elements)
	newElements[len(arr.Elements)] = element
	return &object.Array{Elements: newElements}
}

func arrayPrepend(arr *object.Array, element object.Object) object.Object {
	newElements := make([]object.Object, len(arr.Elements)+1)
	newElements[0] = element
	copy(newElements[1:], arr.Elements)
	return &object.Array{Elements: newElements}
}

func arrayRemove(arr *object.Array, index int64) object.Object {
	if index < 0 || index >= int64(len(arr.Elements)) {
		return arr
	}

	newElements := make([]object.Object, len(arr.Elements)-1)
	copy(newElements, arr.Elements[:index])
	copy(newElements[index:], arr.Elements[index+1:])
	return &object.Array{Elements: newElements}
}

func evalArrayInfixExpression(operator string, left, right object.Object) object.Object {
	leftArray := left.(*object.Array)
	rightArray := right.(*object.Array)

	switch operator {
	case "+", "||": // Concaténation
		newElements := make([]object.Object, len(leftArray.Elements)+len(rightArray.Elements))
		copy(newElements, leftArray.Elements)
		copy(newElements[len(leftArray.Elements):], rightArray.Elements)
		return &object.Array{Elements: newElements}
	case "==":
		if len(leftArray.Elements) != len(rightArray.Elements) {
			return &object.Boolean{Value: false}
		}
		for i := range leftArray.Elements {
			if !objectsEqual(leftArray.Elements[i], rightArray.Elements[i]) {
				return &object.Boolean{Value: false}
			}
		}
		return &object.Boolean{Value: true}
	case "!=":
		return &object.Boolean{Value: !evalArrayInfixExpression("==", left, right).(*object.Boolean).Value}
	default:
		return newError("Opérateur inconnu: %s %s %s", left.Type(), operator, right.Type())
	}
}

func extractElementType(arrayType string) string {
	// Extrait le type d'élément d'une chaîne comme "array of integer"
	parts := strings.Split(arrayType, " of ")
	if len(parts) > 1 {
		return parts[1]
	}
	return "any"
}

// Ajouter une fonction pour créer des tableaux avec taille fixe
func NewFixedSizeArray(size int64, elementType string) *object.Array {
	elements := make([]object.Object, size)
	defaultElement := getDefaultValue(elementType)

	for i := range elements {
		elements[i] = defaultElement
	}

	return &object.Array{
		Elements:    elements,
		ElementType: elementType,
		FixedSize:   true,
		Size:        size,
	}
}
func evalSwitchStatement(node *ast.SwitchStatement, env *object.Environment) object.Object {
	// Évaluer l'expression du switch
	scope := object.NewEnclosedEnvironment(env)
	switchValue := Eval(node.Expression, scope)
	if isError(switchValue) {
		return switchValue
	}

	var result object.Object
	matched := false

	// Vérifier chaque case
	for _, caseStmt := range node.Cases {
		if !matched {
			// Vérifier si l'une des expressions du case correspond
			for _, caseExpr := range caseStmt.Expressions {
				caseValue := Eval(caseExpr, scope)
				if isError(caseValue) {
					return caseValue
				}

				// Comparer les valeurs
				if objectsEqual(switchValue, caseValue) {
					matched = true
					result = evalSwitchCaseBody(caseStmt.Body, scope)
					if isError(result) || isBreakOrReturn(result) {
						return cleanReturn(result)
					}
					break
				}
			}
		}
	}

	// Si aucun case ne correspond, exécuter le default
	if !matched && node.DefaultCase != nil {
		result = evalBlockStatement(node.DefaultCase, scope)
		if isError(result) {
			return result
		}
	}

	if result == nil {
		return object.NULL
	}

	return result
}

func evalSwitchCaseBody(body *ast.BlockStatement, env *object.Environment) object.Object {
	var result object.Object

	for _, stmt := range body.Statements {
		result = Eval(stmt, env)

		if result != nil {
			rt := result.Type()

			// Si on rencontre un break, on sort du switch
			if rt == object.BREAK_OBJ {
				return object.NULL
			}

			// Si on rencontre un return ou une erreur, on propage
			if rt == object.RETURN_VALUE_OBJ || rt == object.ERROR_OBJ {
				return result
			}

			// Si on rencontre un fallthrough, on continue au case suivant
			if rt == object.FALLTHROUGH_OBJ {
				// fallthrough se comporte comme si on n'avait pas de break
				// mais continue à vérifier les cases suivants
				return &object.Fallthrough{Value: true}
			}
		}
	}

	return result
}

func evalBreakStatement(node *ast.BreakStatement, env *object.Environment) object.Object {
	return &object.Break{}
}

func evalFallthroughStatement(node *ast.FallthroughStatement, env *object.Environment) object.Object {
	return &object.Fallthrough{Value: true}
}

// Fonctions utilitaires
func isBreakOrReturn(obj object.Object) bool {
	if obj == nil {
		return false
	}
	rt := obj.Type()
	return rt == object.BREAK_OBJ || rt == object.RETURN_VALUE_OBJ
}

func cleanReturn(obj object.Object) object.Object {
	if obj != nil && obj.Type() == object.BREAK_OBJ {
		return object.NULL
	}
	return obj
}

func evalForBody(body *ast.BlockStatement, env *object.Environment) object.Object {
	var result object.Object

	for _, statement := range body.Statements {
		result = Eval(statement, env)

		if result != nil {
			rt := result.Type()

			// Gérer les instructions de contrôle de flux
			if rt == object.BREAK_OBJ || rt == object.CONTINUE_OBJ ||
				rt == object.RETURN_VALUE_OBJ || rt == object.ERROR_OBJ {
				return result
			}
		}
	}

	return result
}

func evalWhileStatement(whileStmt *ast.WhileStatement, env *object.Environment) object.Object {
	scope := object.NewEnclosedEnvironment(env)
	for {
		// Évaluer la condition
		condition := Eval(whileStmt.Condition, env)
		if isError(condition) {
			return condition
		}

		if !isTruthy(condition) {
			break
		}

		// Évaluer le corps
		result := evalForBody(whileStmt.Body, scope)

		if result != nil {
			rt := result.Type()

			if rt == object.BREAK_OBJ {
				break
			}

			if rt == object.CONTINUE_OBJ {
				continue
			}

			if rt == object.RETURN_VALUE_OBJ || rt == object.ERROR_OBJ {
				return result
			}
		}
	}

	return object.NULL
}

// Ajouter l'évaluation des littéraux de durée
func evalDurationLiteral(node *ast.DurationLiteral, env *object.Environment) object.Object {
	duration, err := object.ParseDuration(node.Value)
	if err != nil {
		return newError("Invalid duration : %s", err)
	}
	return duration
}

// Ajouter les opérations sur les durées
func evalDurationInfixExpression(operator string, left, right object.Object) object.Object {
	leftDuration := left.(*object.Duration)
	rightDuration := right.(*object.Duration)

	switch operator {
	case "+":
		return &object.Duration{
			Nanoseconds: leftDuration.Nanoseconds + rightDuration.Nanoseconds,
		}
	case "-":
		return &object.Duration{
			Nanoseconds: leftDuration.Nanoseconds - rightDuration.Nanoseconds,
		}
	case "*":
		// Multiplication par un entier
		if right.Type() == object.INTEGER_OBJ {
			return &object.Duration{
				Nanoseconds: leftDuration.Nanoseconds * right.(*object.Integer).Value,
			}
		}
		if right.Type() == object.FLOAT_OBJ {
			return &object.Duration{
				Nanoseconds: int64(float64(leftDuration.Nanoseconds) * right.(*object.Float).Value),
			}
		}
		return newError("Impossible to multiply a duration with %s", right.Type())
	case "/":
		// Division par un nombre
		if right.Type() == object.INTEGER_OBJ {
			if right.(*object.Integer).Value == 0 {
				return newError("Division by zero")
			}
			return &object.Duration{
				Nanoseconds: leftDuration.Nanoseconds / right.(*object.Integer).Value,
			}
		}
		if right.Type() == object.FLOAT_OBJ {
			if right.(*object.Float).Value == 0 {
				return newError("Division by zero")
			}
			return &object.Duration{
				Nanoseconds: int64(float64(leftDuration.Nanoseconds) / right.(*object.Float).Value),
			}
		}
		// Division de deux durées donne un ratio
		if right.Type() == object.DURATION_OBJ {
			if rightDuration.Nanoseconds == 0 {
				return newError("Division by zero")
			}
			return &object.Float{
				Value: float64(leftDuration.Nanoseconds) / float64(rightDuration.Nanoseconds),
			}
		}
		return newError("Impossible to divide a duration by %s", right.Type())
	case "==":
		return &object.Boolean{Value: leftDuration.Nanoseconds == rightDuration.Nanoseconds}
	case "!=":
		return &object.Boolean{Value: leftDuration.Nanoseconds != rightDuration.Nanoseconds}
	case "<":
		return &object.Boolean{Value: leftDuration.Nanoseconds < rightDuration.Nanoseconds}
	case ">":
		return &object.Boolean{Value: leftDuration.Nanoseconds > rightDuration.Nanoseconds}
	case "<=":
		return &object.Boolean{Value: leftDuration.Nanoseconds <= rightDuration.Nanoseconds}
	case ">=":
		return &object.Boolean{Value: leftDuration.Nanoseconds >= rightDuration.Nanoseconds}
	default:
		return newError("Unknown operation: %s %s %s", left.Type(), operator, right.Type())
	}
}

// Ajouter l'évaluation des opérations avec Date/Time et Duration
func evalDateTimeDurationOperations(left, right object.Object, operator string) object.Object {
	switch left := left.(type) {
	case *object.Duration:
		if right.Type() == object.DATE_OBJ {
			duration := left
			date := right.(*object.Date)
			// Ajouter la durée à la date
			newTime := date.Value.Add(time.Duration(duration.Nanoseconds))
			return &object.Date{Value: newTime}
		}
		if right.Type() == object.TIME_OBJ {
			duration := left
			timeObj := right.(*object.Time)
			// Ajouter la durée au temps
			newTime := timeObj.Value.Add(time.Duration(duration.Nanoseconds))
			return &object.Time{Value: newTime}
		}
	case *object.Date:
		if right.Type() == object.DURATION_OBJ {
			duration := right.(*object.Duration)
			// Ajouter la durée à la date
			newTime := left.Value.Add(time.Duration(duration.Nanoseconds))
			return &object.Date{Value: newTime}
		}
	case *object.Time:
		if right.Type() == object.DURATION_OBJ {
			duration := right.(*object.Duration)
			// Ajouter la durée au temps
			newTime := left.Value.Add(time.Duration(duration.Nanoseconds))
			return &object.Time{Value: newTime}
		}
	}
	return newError("Non supported operation entre %s et %s", left.Type(), right.Type())
}

func evalBetweenExpression(node *ast.BetweenExpression, env *object.Environment) object.Object {
	value := Eval(node.Base, env)
	if isError(value) {
		return value
	}

	low := Eval(node.Left, env)
	if isError(low) {
		return low
	}

	high := Eval(node.Right, env)
	if isError(high) {
		return high
	}

	// Compare values
	lowComp := evalInfixExpression("<", low, value)
	if isError(lowComp) {
		return lowComp
	}

	highComp := evalInfixExpression("<", value, high)
	if isError(highComp) {
		return highComp
	}

	result := &object.Boolean{
		Value: lowComp.(*object.Boolean).Value && highComp.(*object.Boolean).Value,
	}

	if node.Not {
		return &object.Boolean{Value: !result.Value}
	}

	return result
}

func evalForEachStatement(n ast.Node, env *object.Environment) object.Object {
	// Accept ast.Node to avoid compile issues if called from Eval with a typed node;
	// then assert to the expected *ast.ForEachStatement.
	node, ok := n.(*ast.ForEachStatement)
	if !ok {
		return newError("Invalid node pour foreach")
	}

	collection := Eval(node.Iterator, env)
	if isError(collection) {
		return collection
	}

	switch coll := collection.(type) {
	case *object.SQLResult:
		for _, el := range coll.Rows {
			loopEnv := object.NewEnclosedEnvironment(env)

			// Si une variable clé est présente, la définir (index)
			// if node.Key != nil {
			// 	loopEnv.Set(node.Key.Value, &object.Integer{Value: int64(i)})
			// }

			// // Si une variable valeur est présente, la définir
			// if node.Value != nil {
			// 	loopEnv.Set(node.Value.Value, el)
			// }
			loopEnv.Set(node.Variable.Value, &object.Struct{Name: "", Fields: el})
			// Évaluer le corps avec l'environnement local
			result := evalForBody(node.Body, loopEnv)
			if result != nil {
				rt := result.Type()

				if rt == object.BREAK_OBJ {
					// sortir de la boucle
					return object.NULL
				}

				if rt == object.CONTINUE_OBJ {
					// passer à l'itération suivante
					continue
				}

				if rt == object.RETURN_VALUE_OBJ || rt == object.ERROR_OBJ {
					return result
				}
			}
		}
		return object.NULL
	case *object.Array:
		for _, el := range coll.Elements {
			loopEnv := object.NewEnclosedEnvironment(env)

			// Si une variable clé est présente, la définir (index)
			// if node.Key != nil {
			// 	loopEnv.Set(node.Key.Value, &object.Integer{Value: int64(i)})
			// }

			// // Si une variable valeur est présente, la définir
			// if node.Value != nil {
			// 	loopEnv.Set(node.Value.Value, el)
			// }
			loopEnv.Set(node.Variable.Value, el)
			// Évaluer le corps avec l'environnement local
			result := evalForBody(node.Body, loopEnv)
			if result != nil {
				rt := result.Type()

				if rt == object.BREAK_OBJ {
					// sortir de la boucle
					return object.NULL
				}

				if rt == object.CONTINUE_OBJ {
					// passer à l'itération suivante
					continue
				}

				if rt == object.RETURN_VALUE_OBJ || rt == object.ERROR_OBJ {
					return result
				}
			}
		}
		return object.NULL
	case *object.Struct:
		for _, field := range coll.Fields {
			loopEnv := object.NewEnclosedEnvironment(env)

			// Définir la variable valeur avec le champ actuel
			loopEnv.Set(node.Variable.Value, field)

			// Évaluer le corps avec l'environnement local
			result := evalForBody(node.Body, loopEnv)
			if result != nil {
				rt := result.Type()

				if rt == object.BREAK_OBJ {
					// sortir de la boucle
					return object.NULL
				}

				if rt == object.CONTINUE_OBJ {
					// passer à l'itération suivante
					continue
				}

				if rt == object.RETURN_VALUE_OBJ || rt == object.ERROR_OBJ {
					return result
				}
			}
		}
		return object.NULL
	default:
		return newError("L'opérande de foreach n'est pas itérable: %s", collection.Type())
	}
}

/* Fonctions utilitaires supplémentaires
function toHours(d: duration): float {
	return d / #1h#;
}

function toDays(d: duration): float {
	return d / #1d#;
}

function toMinutes(d: duration): float {
	return d / #1m#;
}

(* Formatage personnalisé *)
function formatDuree(d: duration): string {
	let jours = d / #1d#;
	let heures = (d % #1d#) / #1h#;
	let minutes = (d % #1h#) / #1m#;
	let secondes = (d % #1m#) / #1s#;

	return jours + "d " + heures + "h " + minutes + "m " + secondes + "s";
}
*/
