package main

import (
	"fmt"

	"github.com/akristianlopez/action/evaluator"
	"github.com/akristianlopez/action/lexer"
	"github.com/akristianlopez/action/object"
	"github.com/akristianlopez/action/parser"
	// "github.com/akristianlopez/action/evaluator"
	// "github.com/akristianlopez/action/lexer"
	// "github.com/akristianlopez/action/object"
	// "github.com/akristianlopez/action/parser"
)

func main() {
	input := `
        let x = 5 + 5;
        let y = x * 2;
        if (y > 15) {
            return true;
        } else {
            return false;
        }
    `

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		fmt.Println("Parser errors:")
		for _, msg := range p.Errors() {
			fmt.Println("\t", msg)
		}
		return
	}

	env := object.NewEnvironment()
	result := evaluator.Eval(program, env)

	if result != nil {
		fmt.Println(result.Inspect())
	}
}
