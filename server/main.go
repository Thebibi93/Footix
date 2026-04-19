package main

import (
	"context"
	"encoding/json"
	"fmt"
	"footix/services"
	"footix/storage" // Import de la logique DB
	"log"
	"net/http"
	"time"

	"github.com/joho/godotenv"
)

// Message définit la structure de réponse JSON pour le test
type Message struct {
	Text string `json:"text"`
}

func printCounter(c int) {
	fmt.Printf("Compteur : %d\n", c)
	time.Sleep(20 * time.Second)
}

func main() {

	fmt.Println("--- ⚽ Démarrage du serveur Footix ---")

	if err := godotenv.Load("../.env"); err != nil {
		log.Println("Aucun fichier .env trouvé, utilisation des variables d'environnement système")
	}

	// Initialisation de la base de données
	db, err := storage.InitDB()
	if err != nil {
		log.Fatalf("Erreur critique DB : %v\n", err)
	}
	defer db.Close()

	// Récupération de la configuration (pour le Token API)
	config, err := storage.LoadConfig()
	if err != nil {
		log.Printf("Impossible de charger le token : %v\n", err)
	}

	// Setup de la routine de suppression des userSessions expirés
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	for range ticker.C {
		_, _ = db.Exec(`DELETE FROM UserSessions WHERE expires_at < now()`)
	}

	// Population de la base de données
	// Liste des codes de ligues autorisés dans le Free Plan de football-data.org
	leagues := []string{"FL1", "PL", "SA", "BL1", "PD", "ELC", "CL", "EC", "DED", "PPL", "BSA", "CLI"}
	seasons := []int{2023, 2024, 2025} // saisons statique dans la DB pour les stats de prédiction

	// On lance la population dans une Goroutine pour ne pas bloquer le serveur
	fmt.Println("Début de la synchronisation avec l'API externe...")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	apiTasks := make(chan services.APITask, 512)

	// démarre d'abord le scheduler unique
	go services.StartAPIScheduler(ctx, apiTasks)

	// charge les métadonnées des ligues UNE fois
	for _, code := range leagues {
		done := make(chan error, 1)

		apiTasks <- services.APITask{
			Label: fmt.Sprintf("league metadata %s", code),
			Done:  done,
			Run: func() error {
				return services.FetchAndSaveLeague(db, config.APIToken, code)
			},
		}
	}

	// historique statique 2023-2025
	for _, code := range leagues {
		for _, season := range seasons {
			code := code
			season := season

			apiTasks <- services.APITask{
				Label: fmt.Sprintf("historical matches %s season %d", code, season),
				Run: func() error {
					return services.FetchAndSaveMatches(db, config.APIToken, code, season)
				},
			}
		}
	}

	// rafraîchissement périodique de la saison courante
	go services.FetchApi(ctx, db, config.APIToken, leagues, 2026, apiTasks)

	// Définition des routes (Router)
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Le serveur Go répond bien !\n")
	})

	// Enregistrement de toutes les autres routes de l'API
	RegisterRoutes(db)

	fmt.Println("Toutes les routes ont été bien enregistrés")

	// Lancement du serveur (Bloquant)
	fmt.Println("Serveur prêt sur http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

/*
L'Encodage (Struct Go -> Flux Client)
Pour répondre à notre client React, on doit faire l'inverse du GET.
Une requete de POST
on transforme la struct en JSON et on l'envoie dans le "tuyau" de réponse de notre serveur (http.ResponseWriter).
*/
func teamHandler(w http.ResponseWriter) {
	// Une struct remplie (venant de la DB par exemple)
	myTeam := storage.TeamInfo{
		Id:        1,
		Name:      "Paris Saint-Germain",
		ShortName: "PSG",
	}

	// On indique au navigateur qu'on envoie du JSON
	w.Header().Set("Content-Type", "application/json")

	// On crée un encodeur qui écrit directement dans 'w' (la réponse)
	json.NewEncoder(w).Encode(myTeam)
}
