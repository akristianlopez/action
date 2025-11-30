package main

import (
	"fmt"
	"os"

	"github.com/akristianlopez/action/lexer"
	"github.com/akristianlopez/action/nsina"
	"github.com/akristianlopez/action/object"
	"github.com/akristianlopez/action/parser"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: lang <filename>")
		os.Exit(1)
	}

	// filename := os.Args[1]
	// input, err := ioutil.ReadFile(filename)
	// if err != nil {
	//     fmt.Printf("Erreur de lecture du fichier: %s\n", err)
	//     os.Exit(1)
	// }
	input := exampleProgram
	l := lexer.New(string(input))
	p := parser.New(l)

	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		fmt.Println("Erreurs de parsing:")
		for _, msg := range p.Errors() {
			fmt.Printf("\t%s\n", msg)
		}
		os.Exit(1)
	}

	env := object.NewEnvironment()
	result := nsina.Eval(program, env)

	if result != nil {
		if result.Type() == object.ERROR_OBJ {
			fmt.Printf("Erreur d'exécution: %s\n", result.Inspect())
			os.Exit(1)
		}
		fmt.Println(result.Inspect())
	}
}

// Exemple de programme test
const exampleProgram = `
(* 
 * Exemple de programme avec toutes les fonctionnalités
 *)

action "Gestion des Employés"

(* Déclaration des structures *)
struct Employé {
    id : integer(5)[1..99999],
    nom : string(50),
    salaire : float(6,2)[0..100000],
    actif : boolean,
    date_embauche : date
}

(* Déclaration des fonctions *)
function calculerBonus(salaire: float, performance: integer) : float {
    let bonus = salaire * (performance / 100.0);
    return bonus;
}

(* Déclaration des variables *)
let employés = object Employé;
let compteur : integer = 0;

start

(* Boucle for style Go *)
for let i = 0; i < 10; i = i + 1 {
    let message = "Itération " + i;
}

(* Requête SQL avec OBJECT *)
let résultats = SELECT e.nom, e.salaire 
                FROM employés AS e 
                WHERE e.salaire > 50000 
                AND e.actif = true;

(* Traitement avec date et time *)
let aujourdhui = #2024-01-15#;
let maintenant = #14:30:00#;

(* Instructions multiples sur une ligne *)
let a = 5; let b = 10; let c = a + b;

stop
`
