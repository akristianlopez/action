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
	res = append(res, testCase{
		name: "Test 1.1 : Let statement ",
		src: `action "Statement 1.1"
			function calculer(a: integer, b: integer): integer {
				let x = 10 + 20; (* Constant folding: 30 *)
				let y = a * 2;
				return x + y;
			}
			function estPair(n: integer): boolean {
				return n % 2 == 0;
			}
			function sommeCarres(limite: integer): integer {
				let total = 0;
				(* Loop avec invariant *)
				for let i = 0; i < limite; i = i + 1 {
					let carre = i * i; (* Peut être optimisé *)
					total = total + carre;
				}
				return total;
			}
			start
				(* Expressions constantes *)
				let a = 5 * 10 + 2; (* Devrait être foldé en 52 *)
				let b = calculer(3, 4);
				(* Code mort potentiel *)
				let c = 10;
				let d = 20; (* Non utilisé *)
				(* Boucle optimisable *)
				for let i = 0; i < 1000; i = i + 1 {
					let resultat = estPair(i);
					if (resultat) {
						c=c+i
					}
				}
				return c
			stop
			 `,
		status: 0,
	})
	// res = append(res, testCase{
	// 	name: "Test 1.2 : Let statement ",
	// 	src: `action "Statement 1.1"
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
