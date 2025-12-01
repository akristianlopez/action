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
*/
