package semantic

// import (
// 	"context"
// 	"fmt"
// 	"os"
// 	"testing"

// 	"github.com/akristianlopez/action/lexer"
// 	"github.com/akristianlopez/action/parser"
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
// 		src: `action "Check the let statement (Global position)"
// 	         Let numero:Integer(5)[10..200],ratio:float(2,1)=2.0
// 			 Let flag:Boolean=true, message:String="Hello World"
// 			 Let step:Integer(2)=1
// 			 start
// 			   return message + " "+ toString(ratio+numero+step) ;
// 			 stop
// 			 `,
// 		status: 0,
// 	})
// 	res = append(res, testCase{
// 		name: "Test 1.2 : Let statement ",
// 		src: `action "Check the let statement (Global position bis)"
// 			 type CustomType struct {
// 			     field1: Integer
// 			     field2: String
// 			 }
// 	         Let var1:CustomType,message:String="Hello World";
// 			 Let step:Integer
// 			 start
// 				(* let var2 = CustomType{field1: 10, field2: "Test"} *)
// 				var1.field2="Updated"
// 				(* var1.field1=10
// 				step=var1.field1*100/2
// 			    return var1.field2 + message; *)
// 			 stop
// 			 `,
// 		status: 0,
// 	})
// 	res = append(res, testCase{
// 		name: "Test 1.3 : Let statement (with Literal structure of the specific type)",
// 		src: `action "Check the let statement (with Literal structure)"
// 			 type CustomType struct {
// 			     field1: Integer
// 			     field2: String
// 			 }
// 			 start
// 				let var2 = CustomType{field1: 10, field2: "Test"}
// 				let var3 = {field1: 10, field2: "Test"}
// 				let var1 :CustomType
// 				var1.field1=20
// 				var1.field2="Hello";
// 				let message:String="Hello World";
// 				var2.field1=var2.field1*210+var3.field1+var1.field1;
// 			    return var2.field2 + message+ toString(var2.field1)+var1.field2;
// 			 stop
// 			 `,
// 		status: 0,
// 	})
// 	res = append(res, testCase{
// 		name: "Test 1.4 : Let statement (with Literal structure without type)",
// 		src: `action "Check the let statement (with Literal structure)"
// 			 start
// 				let var2 = {field1: 10, field2: "Test"}
// 				let var1={field1:20, field2:"Hello"};
// 				var2=var1;
// 			    return var2.field1 + var1.field1;
// 			 stop
// 			 `,
// 		status: 0,
// 	})
// 	res = append(res, testCase{
// 		name: "Test 1.5 : function with a definition and call",
// 		src: `action "Check the function statement (with definition and call)"
// 			 function addition(a:Integer,b:Integer): Integer{
// 			    return a + b;
// 			 }
// 			 function addfloat(a:Float(5,2),b:Float(5,2)): Float(5,2){
// 			 	let e:Float(5,2)=10.0;
// 			    return e + a + b;
// 			 }
// 			 start
// 			 	let a:Integer, b:Float(5,2);
// 				let d:float(5,2);
// 				a=10;b=20.5;
// 				d= addfloat(15.5,4.5);
// 				addition(a,10);
// 			    return addition(a,10)+addfloat(d,4.5);
// 			 stop
// 			 `,
// 		status: 0,
// 	})
// 	res = append(res, testCase{
// 		name: "Test 1.6 : function with with no result and those with structure as result",
// 		src: `action "Check the function statement (with no result and those with structure as result)"
// 			 type Employé struct{
// 			 	first_name: string(30)
// 				Last_name: string(50)
// 				age:integer(3)
// 				matricule:string(8)
// 			 }
// 			 function addition(a:Integer,b:Integer){
// 			    let c=a + b;
// 			 }
// 			 function InitEmployé(n:string, f:string,age:integer): Employé{
// 			    return {first_name:f, Last_name:f,age:age,matricule:'000000'};
// 			 }
// 			 function noResultFunction(){
// 			    let x:Integer=10;
// 				x=x+20;
// 			 }
// 			 start
// 				addition(10,20);
// 				noResultFunction()
// 			 	return InitEmployé('google','Golang',3)
// 			 stop
// 			 `,
// 		status: 0,
// 	})
// 	res = append(res, testCase{
// 		name: "Test 1.7 : array type",
// 		src: `action "Check the function statement (with no result and those with structure as result)"
// 			type Employé struct{
// 			 	first_name: string(30)
// 				Last_name: string(50)
// 				age:integer(3)
// 				matricule:string(8)
// 			 }
// 			 function InitEmployé(n:string, f:string,age:integer): Employé{
// 			    return Employé{first_name:f, Last_name:f,age:age,matricule:'000000'};
// 			 }
// 			 Let arr: array[10] of Employé
// 			 start
// 			 	let arrInteger: array of Integer=[1,2,3,4,5]
// 				arrInteger[2]=10
// 				arr=append(arr, InitEmployé('google','Golang',3))
// 			 	return arr;
// 			 stop
// 			 `,
// 		status: 0,
// 	})
// 	res = append(res, testCase{
// 		name: "Test 1.8 : array literal, slice, len, and append",
// 		src: `action "Check the array literal statement (with array literal, slice, len, and append)"
// 			let nombres: array[10] of integer = [1, 2, 3, 4, 5];
// 			let noms: array of string = ["Alice", "Bob", "Charlie"];
// 			let vide: array of boolean = [];
// 			let matrice: array of array of array of integer = [[[1, 2]], [[3, 4]], [[5, 6]]]
// 			start
// 				(* Tranches (slices) *)
// 				let sous_tableau = nombres[1:3];
// 				let fin = nombres[2:];
// 				let debut = nombres[:3];
// 				let copie = nombres[:];
// 				(* Concaténation *)
// 				let tous = nombres + [6, 7, 8, 9, 10];
// 				let double = nombres + nombres;
// 			stop
// 			 `,
// 		status: 0,
// 	})
// 	res = append(res, testCase{
// 		name: "Test 1.9 : array literal, slice, len, and append",
// 		src: `action "Check the array literal statement (with array literal, slice, len, and append)"
// 			let nombres: array[10] of integer = [1, 2, 3, 4, 5];
// 			let noms: array of string = ["Alice", "Bob", "Charlie"];
// 			let vide: array of boolean = [];
// 			let matrice: array of array of array of integer = [[[1, 2]], [[3, 4]], [[5, 6]]]
// 			start
// 				(* Tranches (slices) *)
// 				let sous_tableau = nombres[1:3];
// 				let fin = nombres[2:];
// 				let debut = nombres[:3];
// 				let copie = nombres[:];
// 				(* Concaténation *)
// 				let tous = nombres + [6, 7, 8, 9, 10];
// 				let double = nombres + nombres;
// 			stop
// 			 `,
// 		status: 0,
// 	})
// 	res = append(res, testCase{
// 		name: "Test 1.10 : Belonging in an array statement",
// 		src: `action "Check belonging in array statement"
// 			let nombres: array[10] of integer = [1, 2, 3, 4, 5];
// 			let matrice: array of array of integer = [[1, 2], [3, 4], [5, 6]];
// 			type Employé struct{
// 			 	first_name: string(30)
// 				Last_name: string(50)
// 				age:integer(3)
// 				matricule:string(8)
// 			 }
// 			start
// 				(* Tableaux multidimensionnels *)
// 				let element = matrice[1][0];
// 				matrice[0] = [10, 20];
// 				(* Vérification d'appartenance *)
// 				let existe = 5 in nombres;
// 				let pas_existe = 20 not in nombres;
// 				let position = indexOf(nombres, 3);
// 			stop
// 			 `,
// 		status: 0,
// 	})
// 	res = append(res, testCase{
// 		name: "Test 1.11 : DateTime and Duration types",
// 		src: `action "Check DateTime and Duration types operations"
// 			(* Tableaux de dates/times *)
// 			let dates_importantes: array of date = [#2024-01-01#, #2024-07-14#, #2024-12-25#];
// 			let horaires: array of time = [#09:00:00#, #12:00:00#, #18:00:00#];
// 			start
// 				(* Tableaux multidimensionnels *)
// 				let element = #2years 3months 5days# + #4hours 30minutes#;
// 				let duree: duration
// 				let now: dateTime = #2024-06-15 10:00:00#;
// 				duree = now-dates_importantes[0];
// 				duree=now - #2023-12-31 08:00:00#;
// 				now=now + #1day 2hours#;
// 			stop
// 			 `,
// 		status: 0,
// 	})
// 	res = append(res, testCase{
// 		name: "Test 1.12 : Conditon expression",
// 		src: `action "Check the If statement (if ... else ...)"
// 			start
// 				let a=10, b=20.5
// 				if (a>b) and ((b==0)) {
// 					b=a*b+10
// 				}else{
// 					b=b*a+20
// 				}
// 				if (a>b) and ((b==0)) {
// 					b=a*b+10
// 				}else if a>20 {
// 					b=b*a+20
// 				}else{
// 					a=0
// 				}
// 			stop
// 			 `,
// 		status: 0,
// 	})
// 	res = append(res, testCase{
// 		name: "Test 1.13 : For statement (for let x of y)",
// 		src: `action "Check the For statement (for let x of y)"
// 			Let elements: array of integer=[1,2,3,4,5,6,7,8,9,0]
// 			start
// 				let a=10, b=20.5
// 				for(let x of elements){
// 					if x%2==0 {
// 						a=a+x
// 					}else {
// 						b=b+x
// 					}
// 				}
// 				for let x of elements{
// 					if x%3==0 {
// 						a=a+x
// 					}else {
// 						b=b+x
// 					}
// 			}
// 			return a>b
// 			stop
// 			 `,
// 		status: 0,
// 	})
// 	res = append(res, testCase{
// 		name: "Test 1.14 : For statement (for condition...)",
// 		src: `action "Check the For statement (for condition...)"
// 			Let elements: array of integer=[1,2,3,4,5,6,7,8,9,0]
// 			start
// 				let a=10, b=20.5, k=1
// 				let x=10
// 				for(b>a and k<len(elements)){
// 					if x%2==0 {
// 						a=a+x
// 					}else {
// 						b=b+x
// 					}
// 					k=k+1
// 				}
// 				for b>a and k<len(elements){
// 					if x%2==0 {
// 						a=a+x
// 					}else {
// 						b=b+x
// 					}
// 					k=k+1
// 				}
// 				return a>b
// 			stop
// 			 `,
// 		status: 0,
// 	})
// 	res = append(res, testCase{
// 		name: "Test 1.15 : For statement (for ...;...;...)",
// 		src: `action "Check the For statement (for ...;...;...)"
// 			Let elements: array of integer=[1,2,3,4,5,6,7,8,9,0]
// 			start
// 				let a=10, b=20.5
// 				for(let k=1;k<=len(elements);k=k+1){
// 					if elements[k]%2==0 {
// 						a=a+elements[k]
// 					}else {
// 						b=b+elements[k]
// 					}
// 				}
// 				for let k=1;k<=len(elements);k=k+1{
// 					if elements[k]%2==0 {
// 						a=a+elements[k]
// 					}else {
// 						b=b+elements[k]
// 					}
// 				}
// 				let k=1
// 				for ;k<=len(elements);k=k+1{
// 					if elements[k]%2==0 {
// 						a=a+elements[k]
// 					}else {
// 						b=b+elements[k]
// 					}
// 				}
// 				k=0
// 				for ;k<=len(elements);{
// 					if elements[k]%2==0 {
// 						a=a+elements[k]
// 					}else {
// 						b=b+elements[k]
// 					}
// 					k=k+1
// 				}
// 				return a>b
// 			stop
// 			 `,
// 		status: 0,
// 	})
// 	res = append(res, testCase{
// 		name: "Test 1.16 : Switch statement (Switch)",
// 		src: `action "Check the For statement (Switch)"
// 				(* Switch simple avec valeurs *)
// 				function getJourSemaine(numero: integer): string {
// 					switch (numero) {
// 						case 1:
// 							return "Lundi";
// 						case 2:
// 							return "Mardi";
// 						case 3:
// 							return "Mercredi";
// 						case 4:
// 							return "Jeudi";
// 						case 5:
// 							return "Vendredi";
// 						case 6:
// 							return "Samedi";
// 						case 7:
// 							return "Dimanche";
// 						default:
// 							return "Numéro invalide";
// 					}
// 				}
// 				(* Switch avec multiples valeurs par case *)
// 				function getTypeJour(numero: integer): string {
// 					switch (numero) {
// 						case 1, 2, 3, 4, 5:
// 							return "Jour de semaine";
// 						case 6, 7:
// 							return "Weekend";
// 						default:
// 							return "Inconnu";
// 					}
// 				}
// 				(* Switch avec expressions *)
// 				function evalueNote(score: integer): string {
// 					switch (true) {
// 						case score >= 90:
// 							return "Excellent";
// 						case score >= 80:
// 							return "Très bien";
// 						case score >= 70:
// 							return "Bien";
// 						case score >= 60:
// 							return "Satisfaisant";
// 						default:
// 							return "Échec";
// 					}
// 				}
// 			start
// 			stop
// 			 `,
// 		status: 0,
// 	})
// 	res = append(res, testCase{
// 		name: "Test 1.17 : Switch statement (Switch bis)",
// 		src: `action "Check the For statement (Switch)"
// 				(* Switch pour validation de formulaire *)
// 				function validerFormulaire(nom: string, age: integer, email: string): array of string {
// 					let erreurs: array of string = [];
// 					switch (true) {
// 						case len(nom) == 0:
// 							erreurs = append(erreurs, "Le nom est requis");
// 							break;
// 						case len(nom) < 2:
// 							erreurs = append(erreurs, "Le nom est trop court");
// 							break;
// 					}
// 					switch (true) {
// 						case age < 0:
// 							erreurs = append(erreurs, "L'âge doit être positif");
// 							break;
// 						case age < 18:
// 							erreurs = append(erreurs, "Vous devez avoir au moins 18 ans");
// 							break;
// 						case age > 120:
// 							erreurs = append(erreurs, "Âge invalide");
// 							break;
// 					}
// 					switch (true) {
// 						case "@" not in email:
// 							erreurs = append(erreurs, "Email invalide");
// 							break;
// 						case "." not in email:
// 							erreurs = append(erreurs, "Email invalide");
// 							break;
// 					}
// 					return erreurs;
// 				}
// 				(* Switch avec différents types *)
// 				function describeValue(valeur: any): string {
// 					switch (typeOf(valeur)) {
// 						case "integer":
// 							return "Nombre entier: " + toString(valeur);
// 						case "float":
// 							return "Nombre décimal: " + toString(valeur);
// 						case "string":
// 							return "Chaîne: '" + toString(valeur) + "'";
// 						case "boolean":
// 							if (valeur) {
// 								return "Vrai";
// 							} else {
// 								return "Faux";
// 							}
// 						case "array":
// 							return "Tableau de " + tostring(len(valeur)) + " éléments";
// 						default:
// 							return "Type inconnu";
// 					}
// 				}
// 			start
// 				(* Gestion des commandes *)
// 				let statut_commande = "expédiée";
// 				let message:string
// 				switch (statut_commande) {
// 					case "nouvelle":
// 						message="La commande est nouvelle";
// 						break;
// 					case "traitement":
// 						message="La commande est en cours de traitement";
// 						break;
// 					case "expédiée":
// 						message="La commande a été expédiée";
// 						fallthrough;
// 					case "livraison":
// 						message="En cours de livraison";
// 						break;
// 					case "livrée":
// 						message="Commande livrée avec succès";
// 						break;
// 					case "annulée":
// 						message="Commande annulée";
// 						break;
// 					default:
// 						message="Statut inconnu";
// 				}
// 				(* Catégorisation d'âge *)
// 				let age = 25;
// 				let categorie = "";
// 				switch (true) {
// 					case age < 0:
// 						categorie = "Âge invalide";
// 						break;
// 					case age < 13:
// 						categorie = "Enfant";
// 						break;
// 					case age < 18:
// 						categorie = "Adolescent";
// 						break;
// 					case age < 65:
// 						categorie = "Adulte";
// 						break;
// 					default:
// 						categorie = "Senior";
// 				}
// 				message="Catégorie: " + categorie;
// 				(* Gestion des erreurs HTTP *)
// 				let code_http = 404;
// 				switch (code_http) {
// 					case 200, 201, 204:
// 						message = "Succès";
// 						break;
// 					case 400:
// 						message = "Mauvaise requête";
// 						break;
// 					case 401:
// 						message = "Non autorisé";
// 						break;
// 					case 403:
// 						message = "Interdit";
// 						break;
// 					case 404:
// 						message = "Non trouvé";
// 						break;
// 					case 500:
// 						message = "Erreur serveur";
// 						break;
// 					default:
// 						if (code_http >= 100 and code_http < 200) {
// 							message = "Information";
// 						} else if (code_http >= 300 and code_http < 400) {
// 							message = "Redirection";
// 						} else {
// 							message = "Code inconnu";
// 						}
// 				}
// 				(* Switch avec énumérations *)
// 				let couleur = "rouge";
// 				let code_couleur = "";
// 				switch (couleur) {
// 					case "rouge":
// 						code_couleur = "#FF0000";
// 						break;
// 					case "vert":
// 						code_couleur = "#00FF00";
// 						break;
// 					case "bleu":
// 						code_couleur = "#0000FF";
// 						break;
// 					case "jaune":
// 						code_couleur = "#FFFF00";
// 						break;
// 					case "violet":
// 						code_couleur = "#800080";
// 						break;
// 					default:
// 						code_couleur = "#000000"; (* noir par défaut *)
// 				}
// 				(* Switch dans une boucle *)
// 				let nombres = [1, 2, 3, 4, 5, 10, 15, 20];
// 				for let i = 0; i < len(nombres); i = i + 1 {
// 					switch (nombres[i]) {
// 						case 1, 2, 3:
// 							message="Petit nombre: " + toString(nombres[i]);
// 							break;
// 						case 4, 5:
// 							message="Nombre moyen: " + toString(nombres[i]);
// 							break;
// 						case 10:
// 							message="Dix";
// 							break;
// 						case 15:
// 							message="Quinze";
// 							break;
// 						case 20:
// 							message="Vingt";
// 							break;
// 					}
// 				}
// 				(* Switch complexe avec conditions *)
// 				let temperature = 22;
// 				let humidite = 65;
// 				let conditions = "";
// 				switch (true) {
// 					case temperature > 30 and humidite > 70:
// 						conditions = "Très chaud et humide";
// 						break;
// 					case temperature > 25 and humidite > 60:
// 						conditions = "Chaud et humide";
// 						break;
// 					case temperature < 0:
// 						conditions = "Gel";
// 						break;
// 					case temperature < 10 and humidite > 80:
// 						conditions = "Froid et humide";
// 						break;
// 					default:
// 						conditions = "Conditions normales";
// 				}
// 				(* Utilisation *)
// 				let validation = validerFormulaire("Alice", 25, "alice@example.com");
// 			stop
// 			 `,
// 		status: 0,
// 	})
// 	res = append(res, testCase{
// 		name: "Test 1.18 : SQL Insert statement (INSERT INTO ... VALUES ...)",
// 		src: `action "Check the SQL Insert statement (INSERT INTO ... VALUES ...)"
// 			start
// 				let result: Integer
// 				(* Insertion de données *)
// 				INSERT INTO Départements (id, nom, budget)
// 				VALUES (1, 'IT', 1000000.00),
// 					(2, 'RH', 500000.00),
// 					(3, 'Finance', 750000.00)
// 				switch (result) {
// 					case 3:
// 						(* Succès *)
// 						let result2: Integer
// 						(* Insertion de données avec des types différents *)
// 						INSERT INTO Employés (id, nom, salaire, département, date_embauche)
// 						VALUES (1, 'Alice Dupont', 55000.00, 'IT', #2023-01-15#),
// 							(2, 'Bob Martin', 48000.00, 'RH', #2023-03-20#),
// 							(3, 'Charlie Durand', 62000.00, 'IT', #2022-11-10#)
// 						break
// 					default:
// 						(*// Échec *)
// 						break
// 				}
// 				(* Insertion de données à partir d'une sous-requête*)
// 				INSERT INTO Employés (id, nom, salaire, département, date_embauche)
// 				SELECT 4, 'Diana Lopez', 70000.00, 'Finance', #2024-02-05#
// 				FROM info
// 				WHERE info.nom == 'Diana Lopez' and info.date_embauche==#2024-02-05#;
// 			stop
// 			 `,
// 		status: 1,
// 	})
// 	res = append(res, testCase{
// 		name: "Test 1.18 : SQL Update statement (UPDATE ... SET ... WHERE ...)",
// 		src: `action "Check the SQL Update statement (UPDATE ... SET ... WHERE ...)"
// 			start
// 				(* Mise à jour *)
// 				UPDATE Employés
// 				SET salaire = Employés.salaire * 1.05
// 				WHERE Employés.département == 'IT';
// 			stop
// 			 `,
// 		status: 0,
// 	})
// 	res = append(res, testCase{
// 		name: "Test 1.19 : SQL Update statement (DELETE FROM ... WHERE ...)",
// 		src: `action "Check the SQL Update statement (DELETE FROM ... WHERE ...)"
// 			start
// 				(* Suppression *)
// 				DELETE FROM Employés
// 				WHERE Employés.actif == false;
// 			stop
// 			 `,
// 		status: 0,
// 	})
// 	res = append(res, testCase{
// 		name: "Test 1.20 : SQL Select statement (SELECT ... FROM ... WHERE ...)",
// 		src: `action "Check the SQL Update statement (SELECT ... FROM ... WHERE ...)"
// 			start
// 				(* Requêtes SELECT avancées
// 				let employes_actifs = SELECT e.nom, e.salaire, d.nom as département
// 									FROM Employés e
// 									INNER JOIN Départements d ON e.département == d.nom
// 									WHERE e.actif == true
// 									ORDER BY e.salaire DESC;*)
// 				let stats_departements = SELECT département, AVG(salaire) as salaire_moyen, COUNT( * ) as nb_employes
// 										FROM Employés
// 										GROUP BY département
// 										HAVING AVG(salaire) > 50000;
// 			stop
// 			 `,
// 		status: 0,
// 	})
// 	res = append(res, testCase{
// 		name: "Test 1.21 : SQL CREATE OBJECT statement (CREATE OBJECT, CREATE INDEX, CREATE UNIQUE)",
// 		src: `action "Check the SQL Update statement (CREATE OBJECT, CREATE INDEX, CREATE UNIQUE)"
// 			start
// 				(* Création des objets *)
// 				CREATE OBJECT IF NOT EXISTS Employés (
// 					id INTEGER PRIMARY KEY,
// 					nom VARCHAR(50) NOT NULL,
// 					salaire NUMERIC(10,2),
// 					département VARCHAR(30),
// 					date_embauche DATE,
// 					actif BOOLEAN DEFAULT true
// 				);
// 				CREATE OBJECT Départements (
// 					id INTEGER PRIMARY KEY,
// 					nom VARCHAR(50) UNIQUE NOT NULL,
// 					budget NUMERIC(12,2)
// 				);
// 				(* Création d'index *)
// 				CREATE INDEX idx_employes_departement ON Employés(département);
// 				CREATE UNIQUE INDEX idx_employes_nom ON Employés(nom);
// 				(* ALTER TABLE *)
// 				ALTER OBJECT Employés
// 				ADD COLUMN email VARCHAR(100),
// 				ADD CONSTRAINT fk_departement FOREIGN KEY (département) REFERENCES Départements(nom);
// 			stop
// 			 `,
// 		status: 0,
// 	})
// 	res = append(res, testCase{
// 		name: "Test 1.22 : SQL CREATE OBJECT statement (ALTER OBJECT ... ADD COLUMN ..., ADD CONSTRAINT ... FOREIGN KEY ...)",
// 		src: `action "Check the SQL Update statement (ALTER OBJECT ... )"
// 			start
// 				(* ALTER TABLE *)
// 				ALTER OBJECT Employés
// 				ADD COLUMN email VARCHAR(100),
// 				ADD CONSTRAINT fk_departement FOREIGN KEY (département) REFERENCES Départements(nom);
// 				return true;
// 			stop
// 			 `,
// 		status: 0,
// 	})
// 	res = append(res, testCase{
// 		name: "Test 1.23 : SQL Select statement (WITH RECURSIVE ..)",
// 		src: `action "Check the SQL Select statement (WITH RECURSIVE ... )"
// 			start
// 				(* Agrégation récursive : budget par branche *)
// 				WITH RECURSIVE BudgetsBranche AS (
// 					SELECT
// 						id,
// 						nom,
// 						parent_id,
// 						budget as budget_direct,
// 						budget as budget_total
// 					FROM Organisation
// 					WHERE parent_id IS NULL
// 					UNION ALL
// 					SELECT
// 						o.id,
// 						o.nom,
// 						o.parent_id,
// 						o.budget as budget_direct,
// 						o.budget + bb.budget_total as budget_total
// 					FROM Organisation o
// 					INNER JOIN BudgetsBranche bb ON o.parent_id == bb.id
// 				)
// 				SELECT
// 					nom,
// 					budget_direct,
// 					budget_total
// 				FROM BudgetsBranche
// 				ORDER BY budget_total DESC;
// 			stop
// 			 `,
// 		status: 0,
// 	})

// 	return res
// }

// func TestAnalyze(t *testing.T) {
// 	ctx := context.Background()
// 	for _, tc := range build_args() {
// 		t.Run(tc.name, func(t *testing.T) {
// 			// Étape 1: Lexical Analysis
// 			l := lexer.New(string(tc.src))

// 			// Étape 2: Parsing
// 			p := parser.New(l)
// 			action := p.ParseAction()
// 			if p.Errors() != nil && len(p.Errors()) != 0 {
// 				fmt.Println("Erreurs de parsing:")
// 				for _, msg := range p.Errors() {
// 					fmt.Printf("\t%s\n", msg.Message())
// 				}
// 				os.Exit(1)
// 			}

// 			// Étape 3: Analyse Sémantique
// 			analyzer := NewSemanticAnalyzer(ctx, nil)
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

// 		})
// 	}
// }
