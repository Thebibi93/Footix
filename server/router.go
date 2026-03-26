/*
	Ce fichier va centraliser la définition des points d'entrée (endpoints).
	Pour ce projet de prédiction, nous avons besoin d'au moins deux routes majeures :
	une pour la liste des matchs et l'autre pour les statistiques de prédiction d'un match précis.
*/
package main

import (
	"database/sql"
	"encoding/json"
	"footix/storage"
	"net/http"
)


// ==================   POINTS IMPORTANTS ==================
// 1. Chaque route est définie via http.HandleFunc, qui associe une URL à une fonction handler.
// 2. Les handlers sont des fonctions qui prennent en paramètre http.ResponseWriter et *http.Request.
// 3. La communication avec le client se fait exclusivement via JSON, en utilisant json.NewEncoder pour encoder les réponses.
// 4. Les handlers font appel à la couche de stockage pour récupérer ou manipuler les données nécessaires.
// 5. Le CORS est géré au niveau de chaque handler pour permettre les requêtes depuis le client React (en dev local).

/* 
  HTTP, appelle le package storage, et renvoie du JSON.
  CORS (Cross-Origin Resource Sharing) : C'est le piège numéro 1. 
  Le client React (souvent sur le port 5173) est considéré comme une 
  origine différente de notre serveur Go (port 8080). 
  Sans le header Access-Control-Allow-Origin, le navigateur bloquera la requête.
  Asynchronisme (AJAX) : Le client React utilisera fetch() pour appeler
  ces routes de manière asynchrone, ce qui respecte une autre contrainte forte du sujet.
*/





// RegisterRoutes configure tous les points d'entrée de l'API
func RegisterRoutes(db *sql.DB) {
	// Route pour tester la connexion (déjà faite)
	http.HandleFunc("/api/hello", helloHandler)

	// Route pour récupérer les matchs d'une ligue
	// Usage: /api/matches?league=FL1
	http.HandleFunc("/api/matches", func(w http.ResponseWriter, r *http.Request) {
		getMatchesHandler(w, r, db)
	})

	// Route pour obtenir la prédiction d'un match spécifique
	// Usage: /api/predict?matchId=123
	http.HandleFunc("/api/predict", func(w http.ResponseWriter, r *http.Request) {
		getPredictionHandler(w, r, db)
	})
}

// getMatchesHandler récupère la liste des matchs depuis la DB pour une ligue précise.
func getMatchesHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
    // Autoriser les requêtes depuis le client React (CORS)
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Content-Type", "application/json")

    // Récupérer le code de la ligue dans l'URL (ex: ?league=FL1)
    leagueCode := r.URL.Query().Get("league")
    if leagueCode == "" {
        http.Error(w, "Le paramètre 'league' est requis", http.StatusBadRequest)
        return
    }

    // Appeler la couche storage pour récupérer les données
    matches, err := storage.GetMatchesByLeague(db, leagueCode)
    if err != nil {
        http.Error(w, "Erreur lors de la récupération des matchs", http.StatusInternalServerError)
        return
    }

    // Encoder la liste en JSON pour le client 
    json.NewEncoder(w).Encode(matches)
}

func getPredictionHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Gestion du CORS (Indispensable pour React)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	// Récupération des paramètres
	matchID := r.URL.Query().Get("matchId")
	if matchID == "" {
		http.Error(w, "Paramètre matchId manquant", http.StatusBadRequest)
		return
	}

	// Appel à la DB (Logique à créer dans storage)
	stats, err := storage.GetMatchStats(db, matchID)
	if err != nil {
		http.Error(w, "Erreur lors du calcul", http.StatusInternalServerError)
		return
	}

	// Envoi de la réponse au client
	json.NewEncoder(w).Encode(stats)
}