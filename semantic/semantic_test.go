package semantic

import (
	"fmt"
	"os"
	"testing"

	"github.com/akristianlopez/action/lexer"
	"github.com/akristianlopez/action/parser"
)

type testCase struct {
	name   string
	src    string
	status int
}

func build_args() []testCase {
	res := make([]testCase, 0)
	// res = append(res, testCase{
	// 	name: "Test 1.1 : Let statement ",
	// 	src: `action "Check the let statement (Global position)"
	//          Let numero:Integer(5)[10..200],ratio:float(2,1)=2.0
	// 		 Let flag:Boolean=true, message:String="Hello World"
	// 		 Let step:Integer(2)=1
	// 		 start
	// 		   return message + " "+ toString(ratio+numero+step) ;
	// 		 stop
	// 		 `,
	// 	status: 0,
	// })
	// res = append(res, testCase{
	// 	name: "Test 1.2 : Let statement ",
	// 	src: `action "Check the let statement (Global position bis)"
	// 		 type CustomType struct {
	// 		     field1: Integer
	// 		     field2: String
	// 		 }
	//          Let var1:CustomType,message:String="Hello World";
	// 		 Let step:Integer
	// 		 start
	// 			(* let var2 = CustomType{field1: 10, field2: "Test"} *)
	// 			var1.field2="Updated"
	// 			(* var1.field1=10
	// 			step=var1.field1*100/2
	// 		    return var1.field2 + message; *)
	// 		 stop
	// 		 `,
	// 	status: 0,
	// })
	// res = append(res, testCase{
	// 	name: "Test 1.3 : Let statement (with Literal structure of the specific type)",
	// 	src: `action "Check the let statement (with Literal structure)"
	// 		 type CustomType struct {
	// 		     field1: Integer
	// 		     field2: String
	// 		 }
	// 		 start
	// 			let var2 = CustomType{field1: 10, field2: "Test"}
	// 			let var3 = {field1: 10, field2: "Test"}
	// 			let var1 :CustomType
	// 			var1.field1=20
	// 			var1.field2="Hello";
	// 			let message:String="Hello World";
	// 			var2.field1=var2.field1*210+var3.field1+var1.field1;
	// 		    return var2.field2 + message+ toString(var2.field1)+var1.field2;
	// 		 stop
	// 		 `,
	// 	status: 0,
	// })
	// res = append(res, testCase{
	// 	name: "Test 1.4 : Let statement (with Literal structure without type)",
	// 	src: `action "Check the let statement (with Literal structure)"
	// 		 start
	// 			let var2 = {field1: 10, field2: "Test"}
	// 			let var1={field1:20, field2:"Hello"};
	// 			var2=var1;
	// 		    return var2.field1 + var1.field1;
	// 		 stop
	// 		 `,
	// 	status: 0,
	// })
	// res = append(res, testCase{
	// 	name: "Test 1.5 : function with a definition and call",
	// 	src: `action "Check the function statement (with definition and call)"
	// 		 function addition(a:Integer,b:Integer): Integer{
	// 		    return a + b;
	// 		 }
	// 		 function addfloat(a:Float(5,2),b:Float(5,2)): Float(5,2){
	// 		 	let e:Float(5,2)=10.0;
	// 		    return e + a + b;
	// 		 }
	// 		 start
	// 		 	let a:Integer, b:Float(5,2);
	// 			let d:float(5,2);
	// 			a=10;b=20.5;
	// 			d= addfloat(15.5,4.5);
	// 			addition(a,10);
	// 		    return addition(a,10)+addfloat(d,4.5);
	// 		 stop
	// 		 `,
	// 	status: 0,
	// })
	// res = append(res, testCase{
	// 	name: "Test 1.6 : function with with no result and those with structure as result",
	// 	src: `action "Check the function statement (with no result and those with structure as result)"
	// 		 type Employé struct{
	// 		 	first_name: string(30)
	// 			Last_name: string(50)
	// 			age:integer(3)
	// 			matricule:string(8)
	// 		 }
	// 		 function addition(a:Integer,b:Integer){
	// 		    let c=a + b;
	// 		 }
	// 		 function InitEmployé(n:string, f:string,age:integer): Employé{
	// 		    return {first_name:f, Last_name:f,age:age,matricule:'000000'};
	// 		 }
	// 		 function noResultFunction(){
	// 		    let x:Integer=10;
	// 			x=x+20;
	// 		 }
	// 		 start
	// 			addition(10,20);
	// 			noResultFunction()
	// 		 	return InitEmployé('google','Golang',3)
	// 		 stop
	// 		 `,
	// 	status: 0,
	// })
	// res = append(res, testCase{
	// 	name: "Test 1.7 : array type",
	// 	src: `action "Check the function statement (with no result and those with structure as result)"
	// 		type Employé struct{
	// 		 	first_name: string(30)
	// 			Last_name: string(50)
	// 			age:integer(3)
	// 			matricule:string(8)
	// 		 }
	// 		 function InitEmployé(n:string, f:string,age:integer): Employé{
	// 		    return Employé{first_name:f, Last_name:f,age:age,matricule:'000000'};
	// 		 }
	// 		 Let arr: array[10] of Employé
	// 		 start
	// 		 	let arrInteger: array of Integer=[1,2,3,4,5]
	// 			arrInteger[2]=10
	// 			arr=append(arr, InitEmployé('google','Golang',3))
	// 		 	return arr;
	// 		 stop
	// 		 `,
	// 	status: 0,
	// })
	// res = append(res, testCase{
	// 	name: "Test 1.8 : array literal, slice, len, and append",
	// 	src: `action "Check the array literal statement (with array literal, slice, len, and append)"
	// 		let nombres: array[10] of integer = [1, 2, 3, 4, 5];
	// 		let noms: array of string = ["Alice", "Bob", "Charlie"];
	// 		let vide: array of boolean = [];
	// 		let matrice: array of array of array of integer = [[[1, 2]], [[3, 4]], [[5, 6]]]
	// 		start
	// 			(* Tranches (slices) *)
	// 			let sous_tableau = nombres[1:3];
	// 			let fin = nombres[2:];
	// 			let debut = nombres[:3];
	// 			let copie = nombres[:];
	// 			(* Concaténation *)
	// 			let tous = nombres || [6, 7, 8, 9, 10];
	// 			let double = nombres + nombres;
	// 		stop
	// 		 `,
	// 	status: 0,
	// })
	// res = append(res, testCase{
	// 	name: "Test 1.9 : array literal, slice, len, and append",
	// 	src: `action "Check the array literal statement (with array literal, slice, len, and append)"
	// 		let nombres: array[10] of integer = [1, 2, 3, 4, 5];
	// 		let noms: array of string = ["Alice", "Bob", "Charlie"];
	// 		let vide: array of boolean = [];
	// 		let matrice: array of array of array of integer = [[[1, 2]], [[3, 4]], [[5, 6]]]
	// 		start
	// 			(* Tranches (slices) *)
	// 			let sous_tableau = nombres[1:3];
	// 			let fin = nombres[2:];
	// 			let debut = nombres[:3];
	// 			let copie = nombres[:];
	// 			(* Concaténation *)
	// 			let tous = nombres || [6, 7, 8, 9, 10];
	// 			let double = nombres + nombres;
	// 		stop
	// 		 `,
	// 	status: 0,
	// })
	// res = append(res, testCase{
	// 	name: "Test 1.20 : Belonging in an array statement",
	// 	src: `action "Check belonging in array statement"
	// 		let nombres: array[10] of integer = [1, 2, 3, 4, 5];
	// 		type Employé struct{
	// 		 	first_name: string(30)
	// 			Last_name: string(50)
	// 			age:integer(3)
	// 			matricule:string(8)
	// 		 }
	// 		start
	// 			(* Vérification d'appartenance *)
	// 			let existe = 5 in nombres;
	// 			let pas_existe = 20 not in nombres;
	// 			let position = indexOf(nombres, 3);
	// 		stop
	// 		 `,
	// 	status: 0,
	// })

	res = append(res, testCase{
		name: "Test 1.21 : DateTime and Duration types",
		src: `action "Check DateTime and Duration types operations"
			let nombres: array[10] of integer = [1, 2, 3, 4, 5];
			start
				(* Vérification d'appartenance *)
				let existe = 5 in nombres;
				let pas_existe = 20 not in nombres;
				let position = indexOf(nombres, 3); 
			stop
			 `,
		status: 0,
	})
	return res
}

func TestAnalyze(t *testing.T) {
	for _, tc := range build_args() {
		t.Run(tc.name, func(t *testing.T) {
			// Étape 1: Lexical Analysis
			l := lexer.New(string(tc.src))

			// Étape 2: Parsing
			p := parser.New(l)
			action := p.ParseProgram()
			if len(p.Errors()) != 0 {
				fmt.Println("Erreurs de parsing:")
				for _, msg := range p.Errors() {
					fmt.Printf("\t%s\n", msg.Message())
				}
				os.Exit(1)
			}

			// Étape 3: Analyse Sémantique
			analyzer := NewSemanticAnalyzer()
			errors := analyzer.Analyze(action)

			if len(errors) > 0 {
				fmt.Printf("Erreurs sémantiques (%s):\n", tc.name)
				for _, msg := range errors {
					fmt.Printf("\t%s\n", msg)
				}
				if len(analyzer.Warnings) > 0 {
					fmt.Println("Avertissements:")
					for _, msg := range analyzer.Warnings {
						fmt.Printf("\t%s\n", msg)
					}
				}
				os.Exit(1)
			}
			fmt.Printf("✓ Action (%s) valide sémantiquement\n", tc.name)
			// fmt.Println(action.String())
			if len(analyzer.Warnings) > 0 {
				fmt.Println("Avertissements:")
				for _, msg := range analyzer.Warnings {
					fmt.Printf("  ⚠ %s\n", msg)
				}
			}

		})
	}
}
