package optimizer

// import (
// 	"fmt"
// 	"os"
// 	"testing"

// 	"github.com/akristianlopez/action/lexer"
// 	"github.com/akristianlopez/action/parser"
// 	"github.com/akristianlopez/action/semantic"
// )

// type testCase struct {
// 	name   string
// 	src    string
// 	status int
// }

// func build_args() []testCase {
// 	res := make([]testCase, 0)
// 	res = append(res, testCase{
// 		name: "Test 1.1 : Let statement ",
// 		src: `action "Statement 1.1"
// 			  function main():integer{
// 			  	return 42;
// 			  }
// 			 start
// 			 let x = 10;
// 			 let y = 20;
// 			 let z = x + y;
// 			 let srt:string = "Hello, World!";
// 			 let pattern:string = "Hello*";
// 				if srt not like pattern {
// 					z = z + 1;
// 				}
// 				y=y+z
// 				if y between 10 and 200{
// 					y=x+main()
// 				}
// 				return y;
// 			 stop
// 			 `,
// 		status: 0,
// 	})
// 	return res
// }

// func TestAnalyze(t *testing.T) {
// 	for _, tc := range build_args() {
// 		t.Run(tc.name, func(t *testing.T) {
// 			// Étape 1: Lexical Analysis
// 			l := lexer.New(string(tc.src))

// 			// Étape 2: Parsing
// 			p := parser.New(l)
// 			action := p.ParseProgram()
// 			if p.Errors() != nil && len(p.Errors()) != 0 {
// 				fmt.Println("Erreurs de parsing:")
// 				for _, msg := range p.Errors() {
// 					fmt.Printf("\t%s\n", msg.Message())
// 				}
// 				os.Exit(1)
// 			}

// 			// Étape 3: Analyse Sémantique
// 			analyzer := semantic.NewSemanticAnalyzer()
// 			errors := analyzer.Analyze(action)

// 			if len(errors) > 0 {
// 				fmt.Printf("Erreurs sémantiques (%s):\n", tc.name)
// 				for _, msg := range errors {
// 					fmt.Printf("\t%s\n", msg)
// 				}
// 				if len(analyzer.Warnings) > 0 {
// 					fmt.Println("Avertissements:")
// 					for _, msg := range analyzer.Warnings {
// 						fmt.Printf("\t%s\n", msg)
// 					}
// 				}
// 				os.Exit(1)
// 			}
// 			fmt.Printf("✓ Action (%s) valide sémantiquement\n", tc.name)
// 			// fmt.Println(action.String())
// 			if len(analyzer.Warnings) > 0 {
// 				fmt.Println("Avertissements:")
// 				for _, msg := range analyzer.Warnings {
// 					fmt.Printf("  ⚠ %s\n", msg)
// 				}
// 			}

// 			// Étape 4: Optimisation
// 			// var optimizedProgram *ast.Program = action
// 			opt := NewOptimizer()
// 			optimizedProgram := opt.Optimize(action)
// 			// opt.Optimize(action)
// 			fmt.Printf("\n✓ Action (%s) optimisée avec succès\n Lines:%d\n", tc.name, len(optimizedProgram.Statements))
// 			if len(opt.Warnings) > 0 {
// 				fmt.Println("Avertissements d'optimisation:")
// 			}

// 			for _, msg := range opt.Warnings {
// 				fmt.Printf("  ⚠ %s\n", msg)
// 			}
// 			fmt.Printf("✓ Optimisations appliquées:\n")
// 			fmt.Printf("  - Constant folding: %d\n", opt.Stats.ConstantFolds)
// 			fmt.Printf("  - Dead code removal: %d\n", opt.Stats.DeadCodeRemovals)
// 			fmt.Printf("  - Function inlining: %d\n", opt.Stats.InlineExpansions)
// 			fmt.Printf("  - Loop optimizations: %d\n", opt.Stats.LoopOptimizations)
// 		})
// 	}
// }
