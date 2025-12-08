package parser

import (
	"fmt"
	"testing"

	"github.com/akristianlopez/action/lexer"
)

type testCase struct {
	name   string
	src    string
	status int
}

func build_args() []testCase {
	res := make([]testCase, 0)
	// res = append(res, testCase{
	// 	name: "Test 1.1 : Let statement let a, b, c",
	// 	src: `action "Check the let statement"
	// 		 start
	// 		 	let a, b, c
	// 		 stop
	// 		 `,
	// 	status: 1,
	// })
	// res = append(res, testCase{
	// 	name: "Test 1.2 : Let statement",
	// 	src: `action "Check the let statement"
	// 		 start
	// 		 	let a :integer=0,
	// 			let b=1.0
	// 			let c="my golang"
	// 		 stop
	// 		 `,
	// 	status: 1,
	// })
	// res = append(res, testCase{
	// 	name: "Test 1.3 : Let statement",
	// 	src: `action "Check the let statement"
	// 		 let a :integer=0, b=1.0;
	// 		 let c="my golang", salaire : float(6,2)[0..100000]
	// 		 start
	// 		 	let a = 5; let b = 10; let c = a + b;
	// 		 stop
	// 		 `,
	// 	status: 0,
	// })
	// res = append(res, testCase{
	// 	name: "Test 2: Check the definition of the structure",
	// 	src: `action "Check the definition of the user's type statement"
	// 		type Employe struct {
	// 			id : integer(5)[1..99999],
	// 			nom : string(50),
	// 			salaire : float(6,2)[0..100000],
	// 			actif : boolean,
	// 			date_embauche : date
	// 		}
	// 		type Student struct{
	// 			id : integer(10)[1..9999999999],
	// 			nom : string(150),
	// 			prenom:string(250),
	// 			code : string(8),
	// 			sexe: string(1)
	// 		}
	// 		type Commande struct{
	// 			id: integer,
	// 			type: string,
	// 			montant: float
	// 		}
	// 		start
	// 			let employes : Employe
	// 			let students : Student
	// 			let commandes: Commande
	// 		stop
	// 		 `,
	// 	status: 0,
	// })
	// res = append(res, testCase{
	// 	name: "Test 3.1: Check function declaration",
	// 	src: `action "Check a function definition's statement"
	// 		type Employe struct {
	// 			id : integer(5)[1..99999],
	// 			nom : string(50),
	// 			salaire : float(6,2)[0..100000],
	// 			actif : boolean,
	// 			date_embauche : date
	// 		}
	// 		(* Déclaration des fonctions *)
	// 		function calculerBonus(salaire: float, performance: integer) : float {
	// 			let bonus = salaire * (performance / 100.0);
	// 			return bonus;
	// 		}
	// 		start
	// 			let result=calculerBonus(105000.5,100)
	// 			return result
	// 		stop
	// 		 `,
	// 	status: 0,
	// })
	// res = append(res, testCase{
	// 	name: "Test 3.2: Check the definition of the structure",
	// 	src: `action "Check the definition of the user's type statement"
	// 		type Employe struct {
	// 			id : integer(5)[1..99999],
	// 			nom : string(50),
	// 			salaire : float(6,2)[0..100000],
	// 			actif : boolean,
	// 			date_embauche : date
	// 		}
	// 		type Student struct{
	// 			id : integer(10)[1..9999999999],
	// 			nom : string(150),
	// 			prenom:string(250),
	// 			code : string(8),
	// 			sexe: string(1)
	// 		}
	// 		type Commande struct{
	// 			id: integer,
	// 			nature: string,
	// 			montant: float
	// 		}
	// 		start
	// 			let employes : Employe
	// 			let students : Student
	// 			let commandes: Commande
	// 			commandes=Commande{id:0,  nature:"toto",montant:10.00}
	// 		stop
	// 		 `,
	// 	status: 0,
	// })
	// res = append(res, testCase{
	// 	name: "Test 3.3: Check the definition of the structure",
	// 	src: `action "Check the definition of the user's type statement"
	// 		start
	// 			return {}
	// 		stop
	// 		 `,
	// 	status: 0,
	// })
	// res = append(res, testCase{
	// 	name: "Test 3.4: Check the definition of the structure",
	// 	src: `action "Check the definition of the user's type statement"
	// 		start
	// 			return {id:0,  nature:"toto",montant:10.00}
	// 		stop
	// 		 `,
	// 	status: 0,
	// })
	// res = append(res, testCase{
	// 	name: "Test 3.5: Check function declaration",
	// 	src: `action "Check a function definition's statement"
	// 		(* Déclaration des fonctions *)
	// 		function calculerBonus(salaire: float, performance: integer) : float {
	// 			let bonus = salaire * (performance / 100.0);
	// 			return bonus;
	// 		}
	// 		(* Déclaration des fonctions *)
	// 		function calculerRegulier(salaire: float, performance: integer) : float {
	// 			type calculer struct{
	// 				tva : float(2,2)[0.0..100.0]
	// 				mtva:float(10,2)[0.0..9999999999.00]
	// 			}
	// 			let cal = calculer{tva:19.50,mtva:0.0}
	// 			cal.mtva = salaire * (cal.tva / 100.0);
	// 			return cal.mtva;
	// 		}
	// 		start
	// 			let result=calculerBonus(105000.5,100)
	// 			return calculer{tva:19.50,mtva:0.0}
	// 		stop
	// 		 `,
	// 	status: 0,
	// })
	// res = append(res, testCase{
	// 	name: "Test 3.6: Check function declaration",
	// 	src: `action "Check a function definition's statement"
	// 		(* Déclaration des fonctions *)
	// 		function calculerBonus(salaire: float, performance: integer) : float {
	// 			let bonus = salaire * (performance / 100.0);
	// 			return bonus;
	// 		}
	// 		(* Déclaration des fonctions *)
	// 		function calculerRegulier(salaire: float, performance: integer) : float {
	// 			type calculer struct{
	// 				tva : float(2,2)[0.0..100.0]
	// 				mtva:float(10,2)[0.0..9999999999.00]
	// 			}
	// 			let cal = calculer{tva:19.50,mtva:0.0}
	// 			cal.mtva = salaire * (cal.tva / 100.0);
	// 			return cal.mtva;
	// 		}
	// 		start
	// 			let result=calculerBonus(105000.5,100)
	// 			return calculer{tva:19.50,mtva:0.0}
	// 		stop
	// 		 `,
	// 	status: 0,
	// })
	// res = append(res, testCase{
	// 	name: "Test 3.7 : Let statement",
	// 	src: `action "Check the let statement"
	// 		 start
	// 			(* Déclaration de tableaux *)
	// 			let nombres: array[10] of integer = [1, 2, 3, 4, 5];
	// 			let noms: array of string = ["Alice", "Bob", "Charlie"];
	// 			let matrice: array of array of integer = [[1, 2], [3, 4], [5, 6]];
	// 			let vide: array of boolean = [];

	// 			(* Tableau avec contraintes *)
	// 			let scores: array[100] of integer(3)[0..100];
	// 		 stop
	// 		 `,
	// 	status: 0,
	// })
	// res = append(res, testCase{
	// 	name: "Test 3.8 : Let statement",
	// 	src: `action "Check the let statement"
	// 		 start
	// 			(* Déclaration de tableaux *)
	// 			let nombres: array[10] of integer = [1, 2, 3, 4, 5];
	// 			let noms: array of string = ["Alice", "Bob", "Charlie"];
	// 			let matrice: array of array of integer = [[1, 2], [3, 4], [5, 6]];
	// 			let vide: array of boolean = [];

	// 			(* Tableau avec contraintes *)
	// 			let scores: array[100] of integer(3)[0..100];
	// 			return []
	// 		 stop
	// 		 `,
	// 	status: 0,
	// })
	// res = append(res, testCase{
	// 	name: "Test 4.1 : Test des Structures de controle",
	// 	src: `action "Check If statement"
	// 		 start
	// 		 	let a=0, b=1, c:integer=0
	// 			if (b>a){
	// 				c=b
	// 			}
	// 			return c
	// 		 stop
	// 		 `,
	// 	status: 0,
	// })
	// res = append(res, testCase{
	// 	name: "Test 4.2 : Test des Structures de controle",
	// 	src: `action "Check If statement"
	// 		 start
	// 		 	let a=0, b=1, c:integer=0
	// 			if b>a {
	// 				c=b
	// 			}
	// 			return c
	// 		 stop
	// 		 `,
	// 	status: 0,
	// })
	// res = append(res, testCase{
	// 	name: "Test 4.3 : Test des Structures de controle",
	// 	src: `action "Check If...Else statement"
	// 		 start
	// 		 	let a=0, b=1, c:integer=0
	// 			if ((b>a) and (c==0)){
	// 				c=0 c=b
	// 			}else{
	// 				c=a
	// 			}
	// 			return c
	// 		 stop
	// 		 `,
	// 	status: 0,
	// })
	// res = append(res, testCase{
	// 	name: "Test 4.4 : Test des Structures de controle",
	// 	src: `action "Check If...Else If statement"
	// 		 start
	// 		 	let a=0, b=1, c:integer=0
	// 			if ((b>a) and (c==0)){
	// 				c=0 ; c=b
	// 			}else if(c>0){
	// 				return -1
	// 			}
	// 			return c
	// 		 stop
	// 		 `,
	// 	status: 0,
	// })
	// res = append(res, testCase{
	// 	name: "Test 4.5 : Test des Structures de controle",
	// 	src: `action "Check the For statement"
	// 		 start
	// 			(* Parcours de tableau *)
	// 			for let i = 0; i < length(nombres); i = 1 + i {
	// 				let valeur = 10;
	// 				(* Traitement... *)
	// 			}
	// 		 stop
	// 		 `,
	// 	status: 0,
	// })
	// res = append(res, testCase{
	// 	name: "Test 4.6 : Test des Structures de controle",
	// 	src: `action "Check the For statement"
	// 		 start
	// 			(* Parcours de tableau *)
	// 			for (let i = 0; i < length(nombres); i = 1 + i {
	// 				let valeur = 10;
	// 				(* Traitement... *)
	// 			}
	// 		 stop
	// 		 `,
	// 	status: 1,
	// })
	// res = append(res, testCase{
	// 	name: "Test 4.7 : Test des Structures de controle",
	// 	src: `action "Check the For statement"
	// 		 start
	// 			(* Parcours de tableau *)
	// 			for (let i = 0; i < length(nombres); i = 1 + i {
	// 				let valeur = 10;
	// 				(* Traitement... *)
	// 			}
	// 		 stop
	// 		 `,
	// 	status: 1,
	// })
	// res = append(res, testCase{
	// 	name: "Test 4.8 : Test des Structures de controle",
	// 	src: `action "Check the For statement"
	// 		 start
	// 			(* Parcours de tableau *)
	// 			for (let i = 0; i < length(nombres) and i<10); i = 1 + i) {
	// 				let valeur = 10;
	// 				(* Traitement... *)
	// 			}
	// 		 stop
	// 		 `,
	// 	status: 1,
	// })
	// res = append(res, testCase{
	// 	name: "Test 4.9 : Test des Structures de controle",
	// 	src: `action "Check the For statement"
	// 		 start
	// 			(* Parcours de tableau *)
	// 			for (let i = 0; i < length(nombres) and i<10); i = 1 + i) {
	// 				let valeur = 10;
	// 				(* Traitement... *)
	// 			}
	// 		 stop
	// 		 `,
	// 	status: 1,
	// })
	// res = append(res, testCase{
	// 	name: "Test 4.10 : Test des Structures de controle",
	// 	src: `action "Check the statement For ;...;... "
	// 		 start
	// 			(* Parcours de tableau *)
	// 			for ; i < length(nombres) and i<10; i = 1 + i {
	// 				let valeur = 10;
	// 				(* Traitement... *)
	// 			}
	// 		 stop
	// 		 `,
	// 	status: 0,
	// })
	// res = append(res, testCase{
	// 	name: "Test 4.11 : Test des Structures de controle",
	// 	src: `action "Check the statement For ;...; "
	// 		 start
	// 			(* Parcours de tableau *)
	// 			for ; i < length(nombres) and i<10; {
	// 				let valeur = 10;
	// 				(* Traitement... *)
	// 			}
	// 		 stop
	// 		 `,
	// 	status: 0,
	// })
	// res = append(res, testCase{
	// 	name: "Test 4.12 : Test des Structures de controle : For (;...;)",
	// 	src: `action "Check the statement For (;...;) "
	// 		 start
	// 			(* Parcours de tableau *)
	// 			for  (;i < length(nombres) and i<10;) {
	// 				let valeur = 10;
	// 				(* Traitement... *)
	// 			}
	// 		 stop
	// 		 `,
	// 	status: 0,
	// })
	// res = append(res, testCase{
	// 	name: "Test 4.13 : Test des Structures de controle : action 'Check the statement while'",
	// 	src: `action "Check the statement while  "
	// 		 start
	// 			(* Parcours de tableau *)
	// 			While i < length(nombres) {
	// 				let valeur = 10;
	// 				(* Traitement... *)
	// 			}
	// 		 stop
	// 		 `,
	// 	status: 0,
	// })
	// res = append(res, testCase{
	// 	name: "Test 4.14 : Test des Structures de controle : action 'Check the statement while'",
	// 	src: `action "Check the statement while  "
	// 		 start
	// 			(* Parcours de tableau *)
	// 			While (i < length(nombres)) {
	// 				let valeur = 10;
	// 				(* Traitement... *)
	// 			}
	// 		 stop
	// 		 `,
	// 	status: 0,
	// })
	// res = append(res, testCase{
	// 	name: "Test 4.14 : Test des tableaux : instruction d'affectation (access to one element)",
	// 	src: `action "Check expression with arrays  "
	// 		 start
	// 		 	let res=nombres[i]
	// 		 	return res
	// 		 stop
	// 		 `,
	// 	status: 0,
	// })
	// res = append(res, testCase{
	// 	name: "Test 4.15 : Test des tableaux : instruction d'affectation (get a slice [x:y])",
	// 	src: `action "Check expression with arrays  "
	// 		 start

	// 			let points: array of Point = [
	// 				{x: 1, y: 2},
	// 				{x: 3, y: 4},
	// 				{x: 5, y: 6}
	// 			];
	// 			(* Tranches (slices) *)
	// 			let sous_tableau = nombres[1:3];
	// 		 stop
	// 		 `,
	// 	status: 0,
	// })
	// res = append(res, testCase{
	// 	name: "Test 4.16 : Test des tableaux : instruction d'affectation (get a slice [x:])",
	// 	src: `action "Check expression with arrays  "
	// 		 start
	// 			let fin = nombres[2:];
	// 		 stop
	// 		 `,
	// 	status: 0,
	// })
	// res = append(res, testCase{
	// 	name: "Test 4.17 : Test des tableaux : instruction d'affectation (get a slice [:x])",
	// 	src: `action "Check expression with arrays  "
	// 		 start
	// 			let debut = nombres[:3];

	// 		 stop
	// 		 `,
	// 	status: 0,
	// })
	// res = append(res, testCase{
	// 	name: "Test 4.17 : Test des tableaux : instruction d'affectation (get a slice [:])",
	// 	src: `action "Check expression with arrays  "
	// 		 start
	// 			let copie = nombres[:];

	// 		 stop
	// 		 `,
	// 	status: 0,
	// })
	// res = append(res, testCase{
	// 	name: "Test 4.18 : Test des tableaux : instruction d'affectation (concat & including)",
	// 	src: `action "Check expression with arrays  "
	// 		 start
	// 			(* Concaténation
	// 			let tous = nombres || [6, 7, 8, 9, 10];
	// 			let double = nombres + nombres;

	// 			(* Vérification d'appartenance *)
	// 			let existe = 5 in nombres;
	// 			let pas_existe = 20 not in nombres;
	// 		 stop
	// 		 `,
	// 	status: 0,
	// })
	// res = append(res, testCase{
	// 	name: "Test 4.19 : Test des tableaux : instruction d'affectation (IN)",
	// 	src: `action "Check expression with arrays  "
	// 		 start
	// 			(* Vérification d'appartenance *)
	// 			let existe = 5 in nombres;
	// 		 stop
	// 		 `,
	// 	status: 0,
	// })
	// res = append(res, testCase{
	// 	name: "Test 4.20 : Test des tableaux : instruction d'affectation (NOT IN)",
	// 	src: `action "Check expression with arrays  "
	// 		 start
	// 			(* Vérification d'appartenance *)
	// 			let pas_existe = 20 not in nombres;
	// 		 stop
	// 		 `,
	// 	status: 0,
	// })
	res = append(res, testCase{
		name: "Test 4.21 : Test strutures de controle : SWITCH",
		src: `action "Check the statement switch  "
			 start

			 stop
			 `,
		status: 0,
	})

	return res
}

func TestParseProgram(t *testing.T) {
	var hasError bool
	for _, tc := range build_args() {
		fmt.Printf("\n%s is running...", tc.name)
		hasError = false
		t.Run(tc.name, func(t *testing.T) {
			l := lexer.New(tc.src)
			p := New(l)

			p.ParseProgram()
			if tc.status == 0 && len(p.Errors()) > 0 {
				fmt.Println("Erreurs de parsing:")
				hasError = true
				for _, msg := range p.Errors() {
					fmt.Printf("\n\t%s line:%d, column:%d\n", msg.Message(), msg.Line(), msg.Column())
				}
			} else if tc.status >= 1 && len(p.Errors()) == 0 {
				hasError = true
				fmt.Printf("\n\tAucune erreur n'a ete idtentifiee. Bien vouloir verifier les parametres de test")
			}
		})
		if !hasError {
			fmt.Printf("successful\n\n")
		}
	}

}
