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
	res = append(res, testCase{
		name: "Test 1.1 : Let statement ",
		src: `action "Check the let statement (Global position)"
             Let numero:Integer(5)
			 start
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
					fmt.Printf("\t%s\n", msg)
				}
				os.Exit(1)
			}

			// Étape 3: Analyse Sémantique
			analyzer := semantic.NewSemanticAnalyzer()
			errors := analyzer.Analyze(action)

			if len(errors) > 0 {
				fmt.Println("Erreurs sémantiques:")
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
			fmt.Println("✓ Programme valide sémantiquement")
			if len(analyzer.Warnings) > 0 {
				fmt.Println("Avertissements:")
				for _, msg := range analyzer.Warnings {
					fmt.Printf("  ⚠ %s\n", msg)
				}
			}

		})
	}
}
