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
                AND e.actif == true;

(* Traitement avec date et time *)
let aujourdhui =#2024-01-15#;
let maintenant = #14:30:00#;

(* Instructions multiples sur une ligne *)
let a = 5; let b = 10; let c = a + b;

stop
`

	// if len(os.Args) < 2 {
	// 	fmt.Println("Usage: lang <filename>")
	// 	os.Exit(1)
	// }

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
			fmt.Printf("\n\t%s line:%d, column:%d", msg.Message(), msg.Line(), msg.Column())
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

/*
const exampleProgram = `
action "Gestion Base de Données"

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

(* Création d'index *)
CREATE INDEX idx_employes_departement ON Employés(département);
CREATE UNIQUE INDEX idx_employes_nom ON Employés(nom);

(* Insertion de données *)
INSERT INTO Départements (id, nom, budget)
VALUES (1, 'IT', 1000000.00),
       (2, 'RH', 500000.00),
       (3, 'Finance', 750000.00);

INSERT INTO Employés (id, nom, salaire, département, date_embauche)
VALUES (1, 'Alice Dupont', 55000.00, 'IT', #2023-01-15#),
       (2, 'Bob Martin', 48000.00, 'RH', #2023-03-20#),
       (3, 'Charlie Durand', 62000.00, 'IT', #2022-11-10#);

start

(* Requêtes SELECT avancées *)
let employes_actifs = SELECT e.nom, e.salaire, d.nom as département
                      FROM Employés e
                      INNER JOIN Départements d ON e.département = d.nom
                      WHERE e.actif = true
                      ORDER BY e.salaire DESC;

(* Mise à jour *)
UPDATE Employés
SET salaire = salaire * 1.05
WHERE département = 'IT';

(* Suppression *)
DELETE FROM Employés
WHERE actif = false;

(* Requête avec GROUP BY et HAVING *)
let stats_departements = SELECT département, AVG(salaire) as salaire_moyen, COUNT(*) as nb_employes
                         FROM Employés
                         GROUP BY département
                         HAVING AVG(salaire) > 50000;

(* ALTER TABLE *)
ALTER OBJECT Employés
ADD COLUMN email VARCHAR(100),
ADD CONSTRAINT fk_departement FOREIGN KEY (département) REFERENCES Départements(nom);

stop
`

const exampleArrayProgram = `
action "Gestion des Tableaux"

(* Déclaration de tableaux *)
let nombres: array[10] of integer = [1, 2, 3, 4, 5];
let noms: array of string = ["Alice", "Bob", "Charlie"];
let matrice: array of array of integer = [[1, 2], [3, 4], [5, 6]];
let vide: array of boolean = [];

(* Tableau avec contraintes *)
let scores: array[100] of integer(3)[0..100];

start

(* Accès aux éléments *)
let premier = nombres[0];
let dernier = nombres[length(nombres) - 1];

(* Modification d'éléments *)
nombres[0] = 100;
nombres[2] = nombres[1] + nombres[2];

(* Tranches (slices) *)
let sous_tableau = nombres[1:3];
let fin = nombres[2:];
let debut = nombres[:3];
let copie = nombres[:];

(* Concaténation *)
let tous = nombres || [6, 7, 8, 9, 10];
let double = nombres + nombres;

(* Vérification d'appartenance *)
let existe = 5 in nombres;
let pas_existe = 20 not in nombres;

let contient = contains(nombres, 3);
let position = index_of(nombres, 3); (* À implémenter *)

(* Fonctions de tableau *)
let taille = length(nombres);
let avec_ajout = append(nombres, 11);
let avec_insertion = prepend(nombres, 0);
let sans_element = remove(nombres, 2);
let partie = slice(nombres, 1, 4);

(* Tableaux multidimensionnels *)
let element = matrice[1][0];
matrice[0] = [10, 20];

(* Parcours de tableau *)
for let i = 0; i < length(nombres); i = i + 1 {
    let valeur = nombres[i];
    (* Traitement... *)
}

(* Tableaux avec SQL *)
let employes_ids = SELECT id FROM Employés WHERE actif = true;
let salaires: array of float = SELECT salaire FROM Employés;

(* Utilisation avec les structures *)
struct Point {
    x: integer,
    y: integer
}

let points: array of Point = [
    {x: 1, y: 2},
    {x: 3, y: 4},
    {x: 5, y: 6}
];

(* Tableaux de dates/times *)
let dates_importantes: array of date = [#2024-01-01#, #2024-07-14#, #2024-12-25#];
let horaires: array of time = [#09:00:00#, #12:00:00#, #18:00:00#];

(* Fonctions qui retournent des tableaux *)
function getNombresPairs(limite: integer): array of integer {
    let resultat: array of integer = [];
    for let i = 0; i <= limite; i = i + 1 {
        if i % 2 == 0 {
            resultat = append(resultat, i);
        }
    }
    return resultat;
}

let pairs = getNombresPairs(10);

(* Algorithmes sur les tableaux *)
function somme(tableau: array of integer): integer {
    let total = 0;
    for let i = 0; i < length(tableau); i = i + 1 {
        total = total + tableau[i];
    }
    return total;
}

function maximum(tableau: array of integer): integer {
    if length(tableau) == 0 {
        return 0;
    }
    let max = tableau[0];
    for let i = 1; i < length(tableau); i = i + 1 {
        if tableau[i] > max {
            max = tableau[i];
        }
    }
    return max;
}

function filtrer(tableau: array of integer, condition: function): array of integer {
    let resultat: array of integer = [];
    for let i = 0; i < length(tableau); i = i + 1 {
        if condition(tableau[i]) {
            resultat = append(resultat, tableau[i]);
        }
    }
    return resultat;
}

(* Utilisation *)
let mes_nombres = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10];
let total = somme(mes_nombres);
let plus_grand = maximum(mes_nombres);
let pairs_seulement = filtrer(mes_nombres, function(x: integer): boolean {
    return x % 2 == 0;
});

stop

*/
