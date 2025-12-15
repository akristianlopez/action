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
	res = append(res, testCase{
		name: "Test 1.2 : Let statement ",
		src: `action "Check the let statement (Global position)"
			 type CustomType struct {
			     field1: Integer
			     field2: String
			 }
             Let var1:CustomType,	
			  message:String="Hello World"
			 Let step:Integer(2)=1
			 start
				let var2 = CustomType{field1: 10, field2: "Test"}
				var1.field2="Updated"
				var1.field1=100
			    return var1.field2 + toString(step) ; 
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
