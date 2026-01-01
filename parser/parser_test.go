package parser

import (
	"fmt"
	"os"
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
	res = append(res, testCase{
		name: "Test 1.1 : Let statement let a, b, c",
		src: `action "Check the let statement"
			 start
			 	let a, b, c
			 stop
			 `,
		status: 1,
	})
	res = append(res, testCase{
		name: "Test 1.2 : Let statement",
		src: `action "Check the let statement"
			 start
			 	let a :integer=0,
				let b=1.0
				let c="my golang"
			 stop
			 `,
		status: 1,
	})
	res = append(res, testCase{
		name: "Test 1.3 : Let statement",
		src: `action "Check the let statement"
			 let a :integer=0, b=1.0;
			 let c="my golang", salaire : float(6,2)[0..100000]
			 start
			 	let a = 5; let b = 10; let c = a + b;
			 stop
			 `,
		status: 0,
	})
	res = append(res, testCase{
		name: "Test 2: Check the definition of the structure",
		src: `action "Check the definition of the user's type statement"
			type Employe struct {
				id : integer(5)[1..99999],
				nom : string(50),
				salaire : float(6,2)[0..100000],
				actif : boolean,
				date_embauche : date
			}
			type Student struct{
				id : integer(10)[1..9999999999],
				nom : string(150),
				prenom:string(250),
				code : string(8),
				sexe: string(1)
			}
			type Commande struct{
				id: integer,
				type: string,
				montant: float
			}
			start
				let employes : Employe
				let students : Student
				let commandes: Commande
			stop
			 `,
		status: 0,
	})
	res = append(res, testCase{
		name: "Test 3.1: Check function declaration",
		src: `action "Check a function definition's statement"
			type Employe struct {
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
			start
				let result=calculerBonus(105000.5,100)
				return result
			stop
			 `,
		status: 0,
	})
	res = append(res, testCase{
		name: "Test 3.2: Check the definition of the structure",
		src: `action "Check the definition of the user's type statement"
			type Employe struct {
				id : integer(5)[1..99999],
				nom : string(50),
				salaire : float(6,2)[0..100000],
				actif : boolean,
				date_embauche : date
			}
			type Student struct{
				id : integer(10)[1..9999999999],
				nom : string(150),
				prenom:string(250),
				code : string(8),
				sexe: string(1)
			}
			type Commande struct{
				id: integer,
				nature: string,
				montant: float
			}
			start
				let employes : Employe
				let students : Student
				let commandes: Commande
				commandes=Commande{id:0,  nature:"toto",montant:10.00}
			stop
			 `,
		status: 0,
	})
	res = append(res, testCase{
		name: "Test 3.3: Check the definition of the structure",
		src: `action "Check the definition of the user's type statement"
			start
				return {}
			stop
			 `,
		status: 0,
	})
	res = append(res, testCase{
		name: "Test 3.4: Check the definition of the structure",
		src: `action "Check the definition of the user's type statement"
			start
				return {id:0,  nature:"toto",montant:10.00}
			stop
			 `,
		status: 0,
	})
	res = append(res, testCase{
		name: "Test 3.5: Check function declaration",
		src: `action "Check a function definition's statement"
			(* Déclaration des fonctions *)
			function calculerBonus(salaire: float, performance: integer) : float {
				let bonus = salaire * (performance / 100.0);
				return bonus;
			}
			(* Déclaration des fonctions *)
			function calculerRegulier(salaire: float, performance: integer) : float {
				type calculer struct{
					tva : float(2,2)[0.0..100.0]
					mtva:float(10,2)[0.0..9999999999.00]
				}
				let cal = calculer{tva:19.50,mtva:0.0}
				cal.mtva = salaire * (cal.tva / 100.0);
				return cal.mtva;
			}
			start
				let result=calculerBonus(105000.5,100)
				return calculer{tva:19.50,mtva:0.0}
			stop
			 `,
		status: 0,
	})
	res = append(res, testCase{
		name: "Test 3.6: Check function declaration",
		src: `action "Check a function definition's statement"
			(* Déclaration des fonctions *)
			function calculerBonus(salaire: float, performance: integer) : float {
				let bonus = salaire * (performance / 100.0);
				return bonus;
			}
			(* Déclaration des fonctions *)
			function calculerRegulier(salaire: float, performance: integer) : float {
				type calculer struct{
					tva : float(2,2)[0.0..100.0]
					mtva:float(10,2)[0.0..9999999999.00]
				}
				let cal = calculer{tva:19.50,mtva:0.0}
				cal.mtva = salaire * (cal.tva / 100.0);
				return cal.mtva;
			}
			start
				let result=calculerBonus(105000.5,100)
				return calculer{tva:19.50,mtva:0.0}
			stop
			 `,
		status: 0,
	})
	res = append(res, testCase{
		name: "Test 3.7 : Let statement",
		src: `action "Check the let statement"
			 start
				(* Déclaration de tableaux *)
				let nombres: array[10] of integer = [1, 2, 3, 4, 5];
				let noms: array of string = ["Alice", "Bob", "Charlie"];
				let matrice: array of array of integer = [[1, 2], [3, 4], [5, 6]];
				let vide: array of boolean = [];
				(* Tableau avec contraintes *)
				let scores: array[100] of integer(3)[0..100];
			 stop
			 `,
		status: 0,
	})
	res = append(res, testCase{
		name: "Test 3.8 : Let statement",
		src: `action "Check the let statement"
			 start
				(* Déclaration de tableaux *)
				let nombres: array[10] of integer = [1, 2, 3, 4, 5];
				let noms: array of string = ["Alice", "Bob", "Charlie"];
				let matrice: array of array of integer = [[1, 2], [3, 4], [5, 6]];
				let vide: array of boolean = [];
				(* Tableau avec contraintes *)
				let scores: array[100] of integer(3)[0..100];
				return []
			 stop
			 `,
		status: 0,
	})
	res = append(res, testCase{
		name: "Test 4.1 : Test des Structures de controle",
		src: `action "Check If statement"
			 start
			 	let a=0, b=1, c:integer=0
				if (b>a){
					c=b
				}
				return c
			 stop
			 `,
		status: 0,
	})
	res = append(res, testCase{
		name: "Test 4.2 : Test des Structures de controle",
		src: `action "Check If statement"
			 start
			 	let a=0, b=1, c:integer=0
				if b>a {
					c=b
				}
				return c
			 stop
			 `,
		status: 0,
	})
	res = append(res, testCase{
		name: "Test 4.3 : Test des Structures de controle",
		src: `action "Check If...Else statement"
			 start
			 	let a=0, b=1, c:integer=0
				if ((b>a) and (c==0)){
					c=0 c=b
				}else{
					c=a
				}
				return c
			 stop
			 `,
		status: 0,
	})
	res = append(res, testCase{
		name: "Test 4.4 : Test des Structures de controle",
		src: `action "Check If...Else If statement"
			 start
			 	let a=0, b=1, c:integer=0
				if ((b>a) and (c==0)){
					c=0 ; c=b
				}else if(c>0){
					return -1
				}
				return c
			 stop
			 `,
		status: 0,
	})
	res = append(res, testCase{
		name: "Test 4.5 : Test des Structures de controle",
		src: `action "Check the For statement"
			 start
				(* Parcours de tableau *)
				for let i = 0; i < length(nombres); i = 1 + i {
					let valeur = 10;
					(* Traitement... *)
				}
			 stop
			 `,
		status: 0,
	})
	res = append(res, testCase{
		name: "Test 4.6 : Test des Structures de controle",
		src: `action "Check the For statement"
			 start
				(* Parcours de tableau *)
				for (let i = 0; i < length(nombres); i = 1 + i {
					let valeur = 10;
					(* Traitement... *)
				}
			 stop
			 `,
		status: 1,
	})
	res = append(res, testCase{
		name: "Test 4.7 : Test des Structures de controle",
		src: `action "Check the For statement"
			 start
				(* Parcours de tableau *)
				for (let i = 0; i < length(nombres); i = 1 + i {
					let valeur = 10;
					(* Traitement... *)
				}
			 stop
			 `,
		status: 1,
	})
	res = append(res, testCase{
		name: "Test 4.8 : Test des Structures de controle",
		src: `action "Check the For statement"
			 start
				(* Parcours de tableau *)
				for (let i = 0; i < length(nombres) and i<10); i = 1 + i) {
					let valeur = 10;
					(* Traitement... *)
				}
			 stop
			 `,
		status: 1,
	})
	res = append(res, testCase{
		name: "Test 4.9 : Test des Structures de controle",
		src: `action "Check the For statement"
			 start
				(* Parcours de tableau *)
				for (let i = 0; i < length(nombres) and i<10); i = 1 + i) {
					let valeur = 10;
					(* Traitement... *)
				}
			 stop
			 `,
		status: 1,
	})
	res = append(res, testCase{
		name: "Test 4.10 : Test des Structures de controle",
		src: `action "Check the statement For ;...;... "
			 start
				(* Parcours de tableau *)
				for ; i < length(nombres) and i<10; i = 1 + i {
					let valeur = 10;
					(* Traitement... *)
				}
			 stop
			 `,
		status: 0,
	})
	res = append(res, testCase{
		name: "Test 4.11 : Test des Structures de controle",
		src: `action "Check the statement For ;...; "
			 start
				(* Parcours de tableau *)
				for ; i < length(nombres) and i<10; {
					let valeur = 10;
					(* Traitement... *)
				}
			 stop
			 `,
		status: 0,
	})
	res = append(res, testCase{
		name: "Test 4.12 : Test des Structures de controle : For (;...;)",
		src: `action "Check the statement For (;...;) "
			 start
				(* Parcours de tableau *)
				for  (;i < length(nombres) and i<10;) {
					let valeur = 10;
					(* Traitement... *)
				}
			 stop
			 `,
		status: 0,
	})
	res = append(res, testCase{
		name: "Test 4.13 : Test des Structures de controle : action 'Check the statement while'",
		src: `action "Check the statement while  "
			 start
				(* Parcours de tableau *)
				for (i < length(nombres)) {
					let valeur = 10;
					(* Traitement... *)
				}
				for i < length(nombres) {
					let valeur = 10;
					(* Traitement... *)
				}
			 stop
			 `,
		status: 0,
	})
	res = append(res, testCase{
		name: "Test 4.14 : Test des Structures de controle : action 'Check the statement while'",
		src: `action "Check the statement while  "
			 start
				(* Parcours de tableau *)
				for (i < length(nombres)) {
					let valeur = 10;
					(* Traitement... *)
				}
			 stop
			 `,
		status: 0,
	})
	res = append(res, testCase{
		name: "Test 4.14 : Test des tableaux : instruction d'affectation (access to one element)",
		src: `action "Check expression with arrays  "
			 start
			 	let res=nombres[i]
			 	return res
			 stop
			 `,
		status: 0,
	})
	res = append(res, testCase{
		name: "Test 4.15 : Test des tableaux : instruction d'affectation (get a slice [x:y])",
		src: `action "Check expression with arrays  "
			 start
				let points: array of Point = [
					{x: 1, y: 2},
					{x: 3, y: 4},
					{x: 5, y: 6}
				];
				(* Tranches (slices) *)
				let sous_tableau = nombres[1:3];
			 stop
			 `,
		status: 0,
	})
	res = append(res, testCase{
		name: "Test 4.16 : Test des tableaux : instruction d'affectation (get a slice [x:])",
		src: `action "Check expression with arrays  "
			 start
				let fin = nombres[2:];
			 stop
			 `,
		status: 0,
	})
	res = append(res, testCase{
		name: "Test 4.17 : Test des tableaux : instruction d'affectation (get a slice [:x])",
		src: `action "Check expression with arrays  "
			 start
				let debut = nombres[:3];
			 stop
			 `,
		status: 0,
	})
	res = append(res, testCase{
		name: "Test 4.17 : Test des tableaux : instruction d'affectation (get a slice [:])",
		src: `action "Check expression with arrays  "
			 start
				let copie = nombres[:];
			 stop
			 `,
		status: 0,
	})
	res = append(res, testCase{
		name: "Test 4.18 : Test des tableaux : instruction d'affectation (concat & including)",
		src: `action "Check expression with arrays  "
			 start
				(* Concaténation
				let tous = nombres || [6, 7, 8, 9, 10];
				let double = nombres + nombres;
				(* Vérification d'appartenance *)
				let existe = 5 in nombres;
				let pas_existe = 20 not in nombres;
			 stop
			 `,
		status: 0,
	})
	res = append(res, testCase{
		name: "Test 4.19 : Test des tableaux : instruction d'affectation (IN)",
		src: `action "Check expression with arrays  "
			 start
				(* Vérification d'appartenance *)
				let existe = 5 in nombres;
			 stop
			 `,
		status: 0,
	})
	res = append(res, testCase{
		name: "Test 4.20 : Test des tableaux : instruction d'affectation (NOT IN)",
		src: `action "Check expression with arrays  "
			 start
				(* Vérification d'appartenance *)
				let pas_existe = 20 not in nombres;
			 stop
			 `,
		status: 0,
	})
	res = append(res, testCase{
		name: "Test 4.21 : Test strutures de controle : SWITCH (x)",
		src: `action "Check the statement switch  "
			 start
				(* Switch avec constantes *)
				switch (b) {
					case 52:
						print("Valeur attendue");
						break;
					default:
						print("Autre valeur");
						break
				}
			 stop
			 `,
		status: 0,
	})
	res = append(res, testCase{
		name: "Test 4.22 : Test strutures de controle : SWITCH (x) dans une function",
		src: `action "Check the statement switch(x)  "
				(* Switch simple avec valeurs *)
				function getJourSemaine(numero: integer): string {
					switch (numero) {
						case 1:
							return "Lundi";
						case 2:
							return "Mardi";
						case 3:
							return "Mercredi";
						case 4:
							return "Jeudi";
						case 5:
							return "Vendredi";
						case 6:
							return "Samedi";
						case 7:
							return "Dimanche";
						default:
							return "Numéro invalide";
					}
				}
			start
			  return getJourSemaine(1)
			stop
			 `,
		status: 0,
	})
	res = append(res, testCase{
		name: "Test 4.23 : Test strutures de controle : SWITCH (x) with case with multiple value",
		src: `action "Check the statement switch(x)  "
			(* Switch avec multiples valeurs par case *)
			function getTypeJour(numero: integer): string {
				switch (numero) {
					case 1, 2, 3, 4, 5:
						return "Jour de semaine";
					case 6, 7:
						return "Weekend";
					default:
						return "Inconnu";
				}
			}
			start
			  return getTypeJour(1)
			stop
			 `,
		status: 0,
	})
	res = append(res, testCase{
		name: "Test 4.24 : Test strutures de controle : SWITCH (true) with bool expression",
		src: `action "Check the statement switch(x)  "
			(* Switch avec expressions *)
			function evalueNote(score: integer): string {
				switch (true) {
					case score >= 90:
						return "Excellent";
					case score >= 80:
						return "Très bien";
					case score >= 70:
						return "Bien";
					case score >= 60:
						return "Satisfaisant";
					default:
						return "Échec";
				}
			}
			start
			  return evalueNote(1)
			stop
			 `,
		status: 0,
	})
	res = append(res, testCase{
		name: "Test 4.25 : Test strutures de controle : SWITCH (fn(b)) with string value",
		src: `action "Check the statement switch(x)  "
			(* Switch avec différents types *)
			function describeValue(valeur: any): string {
				switch (typeOf(valeur)) {
					case "integer":
						return "Nombre entier: " + valeur;
					case "float":
						return "Nombre décimal: " + valeur;
					case "string":
						return 'Chaîne:' + valeur + "'";
					case "boolean":
						if (valeur) {
							return "Vrai";
						} else {
							return "Faux";
						}
					case "array":
						return "Tableau de " + length(valeur) + " éléments";
					default:
						return "Type inconnu";
				}
			}
			start
			  return evalueNote(1)
			stop
			 `,
		status: 0,
	})
	res = append(res, testCase{
		name: "Test 4.26 : Test strutures de controle : SWITCH multiple switch",
		src: `action "Check the statement switch(x)  "
			start
				(* Gestion des commandes *)
				let statut_commande = "expédiée";
				switch (statut_commande) {
					case "nouvelle":
						print("La commande est nouvelle");
						break;
					case "traitement":
						print("La commande est en cours de traitement");
						break;
					case "expédiée":
						print("La commande a été expédiée");
						fallthrough;
					case "livraison":
						print("En cours de livraison");
						break;
					case "livrée":
						print("Commande livrée avec succès");
						break;
					case "annulée":
						print("Commande annulée");
						break;
					default:
						print("Statut inconnu");
				}
				(* Catégorisation d'âge *)
				let age = 25;
				let categorie = "";
				switch (true) {
					case age < 0:
						categorie = "Âge invalide";
						break;
					case age < 13:
						categorie = "Enfant";
						break;
					case age < 18:
						categorie = "Adolescent";
						break;
					case age < 65:
						categorie = "Adulte";
						break;
					default:
						categorie = "Senior";
				}
				print("Catégorie: " + categorie);
				(* Gestion des erreurs HTTP *)
				let code_http = 404;
				let message = "";
				switch (code_http) {
					case 200, 201, 204:
						message = "Succès";
						break;
					case 400:
						message = "Mauvaise requête";
						break;
					case 401:
						message = "Non autorisé";
						break;
					case 403:
						message = "Interdit";
						break;
					case 404:
						message = "Non trouvé";
						break;
					case 500:
						message = "Erreur serveur";
						break;
					default:
						if (code_http >= 100 and code_http < 200) {
							message = "Information";
						} else if (code_http >= 300 and code_http < 400) {
							message = "Redirection";
						} else {
							message = "Code inconnu";
						}
				}
				print("Message HTTP: " + message);
				(* Switch avec énumérations *)
				let couleur = "rouge";
				let code_couleur = "";
				switch (couleur) {
					case "rouge":
						code_couleur = "#FF0000";
						break;
					case "vert":
						code_couleur = "#00FF00";
						break;
					case "bleu":
						code_couleur = "#0000FF";
						break;
					case "jaune":
						code_couleur = "#FFFF00";
						break;
					case "violet":
						code_couleur = "#800080";
						break;
					default:
						code_couleur = "#000000"; (* noir par défaut *)
				}
				(* Switch dans une boucle *)
				let nombres = [1, 2, 3, 4, 5, 10, 15, 20];
				for let i = 0; i < length(nombres); i = i + 1 {
					switch (nombres[i]) {
						case 1, 2, 3:
							print("Petit nombre: " + nombres[i]);
							break;
						case 4, 5:
							print("Nombre moyen: " + nombres[i]);
							break;
						case 10:
							print("Dix");
							break;
						case 15:
							print("Quinze");
							break;
						case 20:
							print("Vingt");
							break;
					}
				}
				(* Switch avec dates *)
				let jour_semaine = #2024-01-15#.dayOfWeek() ; (* Lundi *)
				let type_journee = "";
				switch (jour_semaine) {
					case 1, 2, 3, 4, 5:
						type_journee = "Jour de travail";
						break;
					case 6:
						type_journee = "Samedi - repos";
						break;
					case 7:
						type_journee = "Dimanche - weekend";
						break;
				}
				(* Switch complexe avec conditions *)
				let temperature = 22;
				let humidite = 65;
				let conditions = "";
				switch (true) {
					case temperature > 30 and humidite > 70:
						conditions = "Très chaud et humide";
						break;
					case temperature > 25 and humidite > 60:
						conditions = "Chaud et humide";
						break;
					case temperature < 0:
						conditions = "Gel";
						break;
					case temperature < 10 and humidite > 80:
						conditions = "Froid et humide";
						break;
					default:
						conditions = "Conditions normales";
				}
			stop
			 `,
		status: 0,
	})
	res = append(res, testCase{
		name: "Test 4.27 : Test strutures de controle : DateTime litteral with a function",
		src: `action "Check the DateTime litteral"
			(* Switch avec différents types *)
			start
				return #2024-01-15#.dayOfWeek();  (* Lundi *)
			stop
			 `,
		status: 0,
	})
	res = append(res, testCase{
		name: "Test 4.28 : Test strutures de controle : For (let .. of ...)",
		src: `action "Check the DateTime litteral"
			(* Switch avec différents types *)
			start
				For (let a of [1,2,3, 4]) {
					a=50*10+2
				}
			stop
			 `,
		status: 0,
	})
	res = append(res, testCase{
		name: "Test 4.29 : Test strutures de controle : For let ... of ..",
		src: `action "Check the DateTime litteral"
			(* Switch avec différents types *)
			start
				For let a of [1,2,3, 4] {
					a=50*10+2
				}
			stop
			 `,
		status: 0,
	})
	res = append(res, testCase{
		name: "Test 5.1 : Test of the SQL Statements : SELECT simple sans where",
		src: `action "Check the DateTime litteral"
			(* Switch avec différents types *)
			start
				SELECT salaire FROM Employés
			stop
			 `,
		status: 0,
	})
	res = append(res, testCase{
		name: "Test 5.2 : Test of the SQL Statements : SELECT simple avec where",
		src: `action "Check the DateTime litteral"
			start
				SELECT id FROM Employés WHERE actif == true;
			stop
			 `,
		status: 0,
	})
	res = append(res, testCase{
		name: "Test 5.3 : Test of the SQL Statements : SELECT simple avec where",
		src: `action "Check the DateTime litteral"
			start
				(* Requêtes SELECT avancées *)
				SELECT e.nom, e.salaire, d.nom as département
				FROM Employés e
					INNER JOIN Départements d ON e.département == d.nom
				WHERE e.actif == true
				ORDER BY e.salaire DESC;
			stop
			 `,
		status: 0,
	})
	res = append(res, testCase{
		name: "Test 5.4 : Test of the SQL Statements : SELECT simple avec where",
		src: `action "Check the DateTime litteral"
			start
				(* Requêtes SELECT avancées *)
				SELECT e.nom, e.salaire, d.nom as département
				FROM Employés e
					INNER JOIN Départements d ON e.département == d.nom
				WHERE e.actif == true
				ORDER BY e.salaire DESC;
				SELECT e.nom, e.salaire
				FROM employés e
				WHERE e.salaire > 50000
					AND e.actif == true;
			stop
			 `,
		status: 0,
	})
	res = append(res, testCase{
		name: "Test 5.5 : Test of the SQL Statements : Advanced SELECT with a function in the clause select",
		src: `action "Check the DateTime litteral"
			start
				(* Requêtes SELECT avancées *)
				SELECT département, AVG(salaire) as salaire_moyen, COUNT(*) as nb_employes
	            FROM Employés
	            GROUP BY département
	            HAVING AVG(salaire) > 50000;
			stop
			 `,
		status: 1, //due to the comment mark '(*' but if you want the compilation
		// process to go smoothly then surround asterisk with space.
	})
	res = append(res, testCase{
		name: "Test 5.6 : Test of the SQL Statements : Advanced SELECT with a function in the clause select",
		src: `action "Requêtes SELECT avancées"
			start
				(* Requêtes SELECT avancées *)
				SELECT
					o.id,
					o.nom,
					o.parent_id,
					o.niveau,
					o.budget,
					ao.niveau_hiérarchique + 1,
					ao.chemin + ' -> ' + o.nom
				FROM Organisation o
				INNER JOIN ArbreOrganisation ao ON o.parent_id == ao.id;
			stop
			 `,
		status: 0,
	})
	res = append(res, testCase{
		name: "Test 5.7 : Test of the SQL Statements : assign advanced select to the variable",
		src: `action "Requêtes SELECT avancées"
			start
				(* Requêtes SELECT avancées *)
				let employes_actifs = SELECT e.nom, e.salaire, d.nom as département
									FROM Employés e
									INNER JOIN Départements d ON e.département == d.nom
									WHERE e.actif == true
									ORDER BY e.salaire DESC;
			stop
			 `,
		status: 0,
	})
	res = append(res, testCase{
		name: "Test 5.8 : Test of the SQL Statements : Advanced SELECT Recursive select",
		src: `action "Requête récursive pour l'arbre complet de l'organisation"
			start
				(* Requête récursive pour l'arbre complet de l'organisation *)
				WITH RECURSIVE ArbreOrganisation AS (
					(* -- Anchor : les racines (sans parent) *)
					SELECT
						id,
						nom,
						parent_id,
						niveau,
						budget,
						0 as niveau_hiérarchique,
						'' as chemin
					FROM Organisation
					WHERE parent_id IS NULL
					UNION ALL
					(* -- Partie récursive : les enfants *)
					SELECT
						o.id,
						o.nom,
						o.parent_id,
						o.niveau,
						o.budget,
						ao.niveau_hiérarchique + 1,
						ao.chemin + ' -> ' + o.nom
					FROM Organisation o
					INNER JOIN ArbreOrganisation ao ON o.parent_id == ao.id
				)
				SELECT
					niveau_hiérarchique,
					nom,
					niveau,
					budget,
					chemin
				FROM ArbreOrganisation
				ORDER BY niveau_hiérarchique, nom;
			stop
			 `,
		status: 0,
	})
	res = append(res, testCase{
		name: "Test 5.9 : Test of the SQL Statements : assign advanced select to the variable",
		src: `action "Requêtes SELECT avancées"
			start
				(* Détection des cycles avec requête récursive *)
				WITH RECURSIVE DetectionCycle AS (
					SELECT
						id,
						nom,
						parent_id,
						(* ARRAY[id] as chemin, *)
						false as cycle
					FROM Organisation
					UNION ALL
					SELECT
						o.id,
						o.nom,
						o.parent_id,
						dc.chemin + o.id (*,
						o.id = ANY(dc.chemin) as cycle *)
					FROM Organisation o
					INNER JOIN DetectionCycle dc ON o.parent_id == dc.id
					WHERE NOT dc.cycle
				)
				SELECT DISTINCT
					nom,
					chemin
				FROM DetectionCycle
				WHERE cycle == true;
			stop
			 `,
		status: 0,
	})
	res = append(res, testCase{
		name: "Test 5.10 : Test of the SQL Statements : INSERT ",
		src: `action "Requêtes INSERT INTO"
			start
				(* Insertion de données *)
				INSERT INTO Départements (id, nom, budget)
				VALUES (1, 'IT', 1000000.00),
					(2, 'RH', 500000.00),
					(3, 'Finance', 750000.00);
				INSERT INTO Employés (id, nom, salaire, département, date_embauche)
				VALUES (1, 'Alice Dupont', 55000.00, 'IT', #2023-01-15#),
					(2, 'Bob Martin', 48000.00, 'RH', #2023-03-20#),
					(3, 'Charlie Durand', 62000.00, 'IT', #2022-11-10#);
				INSERT INTO Organisation (id, nom, parent_id, niveau, budget) VALUES
				(1, 'Entreprise', NULL, 'Direction', 10000000.00),
				(2, 'IT', 1, 'Département', 2000000.00),
				(3, 'RH', 1, 'Département', 800000.00),
				(4, 'Développement', 2, 'Service', 1200000.00),
				(5, 'Infrastructure', 2, 'Service', 800000.00),
				(6, 'Recrutement', 3, 'Service', 400000.00),
				(7, 'Formation', 3, 'Service', 300000.00),
				(8, 'Backend', 4, 'Équipe', 600000.00),
				(9, 'Frontend', 4, 'Équipe', 400000.00),
				(10, 'Base de données', 5, 'Équipe', 300000.00);
	 		stop
	 		 `,
		status: 0,
	})
	res = append(res, testCase{
		name: "Test 5.10 : Test of the SQL Statements : UPDATE ",
		src: `action "Requêtes UPDATE"
			start
				(* Mise à jour *)
				UPDATE Employés
				SET salaire = salaire * 1.05
				WHERE département == 'IT';
			stop
			 `,
		status: 0,
	})
	res = append(res, testCase{
		name: "Test 5.11 : Test of the SQL Statements : DELETE ",
		src: `action "Requêtes DELETE"
			start
				(* Suppression *)
				DELETE FROM Employés
				WHERE actif == false;
			stop
			 `,
		status: 0,
	})
	res = append(res, testCase{
		name: "Test 5.12 : Test of the SQL Statements : CREATE OBJECT ",
		src: `action "Requêtes CREATE OBJECT"
			start
				(* Création des objets *)
				CREATE OBJECT IF NOT EXISTS Employés (
					id INTEGER PRIMARY KEY,
					nom VARCHAR(50) NOT NULL,
					salaire NUMERIC(10,2),
					département VARCHAR(30),
					date_embauche DATE,
					actif BOOLEAN DEFAULT true
				);
				CREATE OBJECT Départements (
					id INTEGER PRIMARY KEY,
					nom VARCHAR(50) UNIQUE NOT NULL,
					budget NUMERIC(12,2)
				);
			stop
			 `,
		status: 0,
	})
	res = append(res, testCase{
		name: "Test 5.13 : Test of the SQL Statements : CREATE AN INDEX ",
		src: `action "Requêtes CREATE AN INDEX"
			start
				(* Création d'index *)
				 CREATE INDEX idx_employes_departement ON Employés(département);
				 CREATE UNIQUE INDEX idx_employes_nom ON Employés(nom);
			stop
			 `,
		status: 0,
	})
	res = append(res, testCase{
		name: "Test 5.14 : Test of the SQL Statements : ALTER ",
		src: `action "Requêtes SQl ALTER"
			start
				(* ALTER TABLE *)
				ALTER OBJECT Employés
				ADD COLUMN email VARCHAR(100),
				ADD CONSTRAINT fk_departement FOREIGN KEY (département) REFERENCES Départements(nom);
				stop
			 `,
		status: 0,
	})
	res = append(res, testCase{
		name: "Test 5.15 : Test of the duration literal ",
		src: `action "Gestion des Durées"
				(* Fonctions avec durées *)
				function ajouterJours(date1: date, jours: integer): date {
					return date1 + #1d# * jours;
				}
				function dureeTotale(taches: array of duration): duration {
					let total: duration = #0s#;
					for let i = 0; i < length(taches); i = i + 1 {
						total = total + taches[i];
					}
					return total;
				}
				function formatDureeHumain(d: duration): string {
					if (d < #1m#) {
						return "Moins d'une minute";
					} else if (d < #1h#) {
						return "Quelques minutes";
					} else if (d < #1d#) {
						return "Quelques heures";
					} else if (d < #7d#) {
						return "Quelques jours";
					} else if (d < #30d#) {
						return "Quelques semaines";
					} else {
						return "Plusieurs mois";
					}
				}
				(* Structures avec durées *)
				type Tache struct{
					nom: string,
					duree_estimee: duration,
					duree_reelle: duration,
					date_echeance: date
				}
				type  Projet struct{
					nom: string,
					taches: array of Tache,
					date_debut: date,
					date_fin: date
				}
				function dureeTotaleProjet(p: Projet): duration {
					let total: duration = #0s#;
					for let i = 0; i < length(p.taches); i = i + 1 {
						total = total + p.taches[i].duree_estimee;
					}
					return total;
				}
				function tempsRestant(p: Projet): duration {
					let maintenant = #now#;  (* Fonction hypothétique pour l'instant courant *)
					if (maintenant > p.date_fin) {
						return #0s#;
					}
					return p.date_fin - maintenant;
				}
				start
					(* Déclaration de variables de type duration *)
					let duree1: duration = #1h 30m#;
					let duree2: duration = #45m#;
					let duree_complexe: duration = #2d 3h 15m 30s#;
					let duree_precise: duration = #1.5s 500ms#;
					(* Opérations sur les durées *)
					let total = duree1 + duree2;                     (* #2h 15m# *)
					let difference = duree1 - duree2;                (* #45m# *)
					let double = duree1 * 2;                         (* #3h# *)
					let moitie = duree1 / 2;                         (* #45m# *)
					let ratio = duree1 / duree2;                     (* 2.0 *)
					(* Comparaisons *)
					let est_plus_long = duree1 > duree2;             (* true *)
					let egal = #1h# == #60m#;                        (* true *)
					(* Opérations avec dates et temps *)
					let aujourdhui = #2024-01-15#;
					let demain = aujourdhui + #1d#;
					let dans_une_semaine = aujourdhui + #7d#;
					let dans_deux_heures = #14:30:00# + #2h#;
					(* Calcul d'intervalle *)
					let date_debut = #2024-01-01#;
					let date_fin = #2024-01-15#;
					let intervalle = date_fin - date_debut;          (* #14d# *)
					(* Durées avec différentes unités *)
					let une_annee = #1y#;
					let un_mois = #1mo#;
					let une_semaine = #1w#;
					let un_jour = #1d#;
					let une_heure = #1h#;
					let une_minute = #1m#;
					let une_seconde = #1s#;
					let une_milliseconde = #1ms#;
					let une_microseconde = #1us#;
					let une_nanoseconde = #1ns#;
					(* Tableaux de durées *)
					let durees_projet: array of duration = [
						#1d#,
						#2d#,
						#3d#,
						#1w#
					];
					let total_projet: duration = #0s#;
					for let i = 0; i < length(durees_projet); i = i + 1 {
						total_projet = total_projet + durees_projet[i];
					}
					(* Calcul de salaire horaire *)
					let heures_travaillees = #160h#;
					let salaire_mensuel = 3000.00;
					let taux_horaire = salaire_mensuel / (heures_travaillees / #1h#);
					(* Planification de projet *)
					let duree_phase1 = #2w#;
					let duree_phase2 = #3w#;
					let duree_phase3 = #1w#;
					let duree_totale = duree_phase1 + duree_phase2 + duree_phase3;
					let date_debut_projet = #2024-01-15#;
					let date_fin_phase1 = date_debut_projet + duree_phase1;
					let date_fin_phase2 = date_fin_phase1 + duree_phase2;
					let date_fin_projet = date_fin_phase2 + duree_phase3;
					(* Suivi du temps *)
					let temps_session1 = #25m#;
					let temps_session2 = #30m#;
					let temps_session3 = #20m#;
					let temps_pause = #5m#;
					let temps_total_session = temps_session1 + temps_session2 + temps_session3;
					let temps_total_avec_pauses = temps_total_session + (temps_pause * 2);
					(* Conversion entre unités *)
					let une_journee = #24h#;
					let en_minutes = une_journee / #1m#;  (* 1440 *)
	(*
					(* Validation de durée
					function validerDuree(d: duration, min: duration, max: duration): boolean {
						return d >= min and d <= max;
					}
					let duree_valide = validerDuree(#8h#, #1h#, #12h#);  (* true
	*)
				stop
				 `,
		status: 0,
	})
	res = append(res, testCase{
		name: "Test 5.16 : Test of the duration literal ",
		src: `action "Gestion des Durées"
			start
				(* Durées avec contraintes *)
				let duree_max: duration(#100d#) = #50d#;  (* Durée maximale de 100 jours *)
				let duree_min: duration[#1h#..#24h#] = #8h#;  (* Entre 1h et 24h *)
				(* Validation de durée
				function validerDuree(d: duration, min: duration, max: duration): boolean {
					return d >= min and d <= max;
				}
				let duree_valide = validerDuree(#8h#, #1h#, #12h#);  (* true
			stop
			 `,
		status: 0,
	})
	res = append(res, testCase{
		name: "Test 5.17 : Test of the advanced select : SELECT with Select in the clause from ",
		src: `action "Requêtes SQl ALTER"
			start
				(* SELECT with SELECT in the clause From *)
				Select t.a, t.b, oo.g, oo.kal
				from table1 t Inner join (select g, kal, id from object2) oo ON (oo.id==t.id)
			stop
			 `,
		status: 0,
	})
	return res
}

func TestParseProgram(t *testing.T) {
	has := false

	for _, tc := range build_args() {
		fmt.Printf("\n%s is running...", tc.name)
		hasError := false
		t.Run(tc.name, func(t *testing.T) {
			l := lexer.New(tc.src)
			p := New(l)

			p.ParseAction()
			if tc.status == 0 && len(p.Errors()) > 0 {
				fmt.Println("Erreurs de parsing:")
				hasError = true
				has = true
				for _, msg := range p.Errors() {
					fmt.Printf("\n\t%s line:%d, column:%d\n", msg.Message(), msg.Line(), msg.Column())
				}
			} else if tc.status >= 1 && len(p.Errors()) == 0 {
				hasError = true
				has = true
				fmt.Printf("\n\tAucune erreur n'a ete idtentifiee. Bien vouloir verifier les parametres de test")
			}
		})
		if !hasError {
			fmt.Printf("successful\n\n")
		}
	}
	if has {
		os.Exit(1)
	}
}
