package nsina

import (
	"fmt"
	"os"
	"testing"

	"github.com/akristianlopez/action/ast"
	"github.com/akristianlopez/action/lexer"
	"github.com/akristianlopez/action/object"
	"github.com/akristianlopez/action/optimizer"
	"github.com/akristianlopez/action/parser"
	"github.com/akristianlopez/action/semantic"
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
	// 	src: `action "Statement 1.1"
	// 		function calculer(a: integer, b: integer): integer {
	// 			let x = 10 + 20; (* Constant folding: 30 *)
	// 			let y = a * 2;
	// 			return x + y;
	// 		}
	// 		function estPair(n: integer): boolean {
	// 			return n % 2 == 0;
	// 		}
	// 		function sommeCarres(limite: integer): integer {
	// 			let total = 0;
	// 			(* Loop avec invariant *)
	// 			for let i = 0; i < limite; i = i + 1 {
	// 				let carre = i * i; (* Peut être optimisé *)
	// 				total = total + carre;
	// 			}
	// 			return total;
	// 		}
	// 		start
	// 			(* Expressions constantes *)
	// 			let a = 5 * 10 + 2; (* Devrait être foldé en 52 *)
	// 			let b = calculer(3, 4);
	// 			(* Code mort potentiel *)
	// 			let c = 0;
	// 			let d = 20; (* Non utilisé *)
	// 			(* Boucle optimisable *)
	// 			for let i = 0; i <1000; i = i + 1 {
	// 				let resultat = estPair(i);
	// 				if (resultat) {
	// 					c=c+i
	// 				}
	// 			}
	// 			return c
	// 		stop
	// 		 `,
	// 	status: 0,
	// })
	// res = append(res, testCase{
	// 	name: "Test 1.2 : Let statement ",
	// 	src: `action "Statement 1.2"
	// 			function main():integer{
	// 				return 42
	// 			}
	// 			function sum(a:integer, b:integer):integer{
	// 				let res:integer =a+b
	// 				return res
	// 			}
	// 		start
	// 			let c=sum(10,20) (* c = 30 *)
	// 			let result:integer
	// 			result=c+main() (* result = 72 *)
	// 			if result>=0{
	// 				let d:integer
	// 				d=sum(c,50) (* d = 80 *)
	// 				result=d (* result = 80*)
	// 			}
	// 			return result
	// 		stop
	// 		 `,
	// 	status: 0,
	// })
	// res = append(res, testCase{
	// 	name: "Test 1.3 : Let statement ",
	// 	src: `action "Statement 1.3"
	// 		start
	// 			let c=1 (* c = 30 *)
	// 			let result:integer
	// 			result=c (* result = 72 *)
	// 			if result>0{
	// 				let d:integer=50
	// 				result=d+c
	// 			}else if c==0{
	// 				let d: integer=10
	// 				result=result+d
	// 			}else{
	// 				let d: integer=30
	// 				result=result+d
	// 			}
	// 			return result
	// 		stop
	// 		 `,
	// 	status: 0,
	// })
	// res = append(res, testCase{
	// 	name: "Test 1.4 : While and Foreach ",
	// 	src: `action "Structure de controle(While, ForEach)"
	// 		Let nombres:array of integer=[0,1,2,3,4,5,6,7,8,9,10]
	// 		start
	// 			let result:integer=0
	// 			for let x of nombres{
	// 				result=result+x
	// 				if result>40 {
	// 					break;
	// 				}
	// 			}
	// 			let k=0
	// 			for k<length(nombres){
	// 				result=result+nombres[k]
	// 				if result>70{
	// 					break
	// 				}
	// 				k=k+1
	// 			}
	// 			(*nombres[0]=result*)
	// 			k=0; result=0
	// 			for let x of nombres[:3]{
	// 				result=result+x
	// 			}
	// 			let str=""
	// 			str=str+toString(result)
	// 			k=0; result=0
	// 			for let x of nombres[4:8]{
	// 				result=result+x
	// 			}
	// 			str=str+" : " +toString(result)
	// 			k=0; result=0
	// 			for let x of nombres[8:]{
	// 				result=result+x
	// 			}
	// 			str=str+" : " +toString(result) + " : nombres[0]= "+ toString(nombres[0])
	// 			return str
	// 		stop
	// 		 `,
	// 	status: 0,
	// })
	// res = append(res, testCase{
	// 	name: "Test 1.5 : switch ",
	// 	src: `action "Structure de controle(Switch)"
	// 		Let nombres:array of integer=[0,1,2,3,4,5,6,7,8,9,10]
	// 			(* Switch avec expressions *)
	// 			function evalueNote(score: integer): string {
	// 				switch (true) {
	// 					case score >= 90:
	// 						return "Excellent";
	// 					case score >= 80:
	// 						return "Très bien";
	// 					case score >= 70:
	// 						return "Bien";
	// 					case score >= 60:
	// 						return "Satisfaisant";
	// 					default:
	// 						return "Échec";
	// 				}
	// 			}
	// 			(* Switch avec multiples valeurs par case *)
	// 			function getTypeJour(numero: integer): string {
	// 				switch (numero) {
	// 					case 1, 2, 3, 4, 5:
	// 						return "Jour de semaine";
	// 					case 6, 7:
	// 						return "Weekend";
	// 					default:
	// 						return "Inconnu";
	// 				}
	// 			}
	// 			(* Switch simple avec valeurs *)
	// 			function getJourSemaine(numero: integer): string {
	// 				switch (numero) {
	// 					case 1:
	// 						return "Lundi";
	// 					case 2:
	// 						return "Mardi";
	// 					case 3:
	// 						return "Mercredi";
	// 					case 4:
	// 						return "Jeudi";
	// 					case 5:
	// 						return "Vendredi";
	// 					case 6:
	// 						return "Samedi";
	// 					case 7:
	// 						return "Dimanche";
	// 					default:
	// 						return "Numéro invalide";
	// 				}
	// 			}
	// 		start
	// 		   return evalueNote(90) + " : " +getTypeJour(5)+ " : " +getJourSemaine(5)
	// 		stop
	// 		 `,
	// 	status: 0,
	// })
	// res = append(res, testCase{
	// 	name: "Test 1.6 : type ... struct{...} ",
	// 	src: `action "Structure de controle(type ... struct{...})"
	// 		type employe struct{
	// 			matricule:string(7)
	// 			nom: string(50)
	// 			prenom:string(150)
	// 			age:integer(3)[15..150]
	// 		}
	// 		Let Employees:array of employe=[{
	// 										 Matricule:'616624-J',
	// 										 Nom:'Evu'
	// 										 Prenom:'Oscar',
	// 										 Age:14
	// 										},
	// 										{
	// 										 Matricule:'616623-M',
	// 										 Nom:'Tabi'
	// 										 Prenom:'Jean Paul'
	// 										 Age:20
	// 										},
	// 										{
	// 										 Matricule:'516624-O',
	// 										 Nom:'EKEME'
	// 										 Prenom:'Maguy'
	// 										 Age:35
	// 										},
	// 										{
	// 										 Matricule:'616624-J',
	// 										 Nom:'FRU'
	// 										 Prenom:'Paul Erick',
	// 										 Age:201
	// 										},
	// 			]
	// 		start
	// 		   Let emp:Employe=Employees[3]
	// 		   emp.age=emp.age-5
	// 		   return emp.age
	// 		stop
	// 		 `,
	// 	status: 0,
	// })
	res = append(res, testCase{
		name: "Test 1.7 : Handling type's constraints",
		src: `action "Handling type's constraints"
			type employe struct{
				matricule:string(7)
				nom: string(50)
				prenom:string(150)
				age:integer(3)[15..150]
			}
			start
			   Let emp:Employe=Employees[3]
			   emp.age=emp.age-5
			   return emp.age
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
			if p.Errors() != nil && len(p.Errors()) != 0 {
				fmt.Println("Erreurs de parsing:")
				for _, msg := range p.Errors() {
					fmt.Printf("\t%s\tline:%d, column:%d\n", msg.Message(), msg.Line(), msg.Column())
				}
				os.Exit(1)
			}

			// Étape 3: Analyse Sémantique
			analyzer := semantic.NewSemanticAnalyzer()
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

			// Étape 4: Optimisation
			var optimizedProgram *ast.Program = action
			opt := optimizer.NewOptimizer()
			optimizedProgram = opt.Optimize(action)
			// opt.Optimize(action)
			if len(opt.Warnings) > 0 {
				fmt.Println("Avertissements d'optimisation:")
			}

			for _, msg := range opt.Warnings {
				fmt.Printf("  ⚠ %s\n", msg)
			}
			fmt.Printf("✓ Optimisations appliquées:\n")
			fmt.Printf("  - Constant folding: %d\n", opt.Stats.ConstantFolds)
			fmt.Printf("  - Dead code removal: %d\n", opt.Stats.DeadCodeRemovals)
			fmt.Printf("  - Function inlining: %d\n", opt.Stats.InlineExpansions)
			fmt.Printf("  - Loop optimizations: %d\n", opt.Stats.LoopOptimizations)
			fmt.Printf("---> Remining statements size: %d\n", len(optimizedProgram.Statements))

			// Étape 5: Évaluation
			env := object.NewEnvironment()
			result := Eval(optimizedProgram, env)

			if result != nil {
				// fmt.Printf("\nErreurs d'Exécution (%s):\n", tc.name)
				if result.Type() == object.ERROR_OBJ {
					fmt.Printf("Erreur d'exécution (%s): %s\n", tc.name, result.Inspect())
					os.Exit(1)
				}
				if result.Type() != object.NULL_OBJ {
					fmt.Println(result.Inspect())
				}
			}
			fmt.Println("---")
			fmt.Println("✓ Programme exécuté avec succès")
		})
	}

}
