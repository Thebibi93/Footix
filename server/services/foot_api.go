// Package services gère la logique métier liée aux interactions avec les API externes.
// Conformément aux contraintes du projet, il utilise exclusivement le package net/http[cite: 13].
package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"footix/storage"
	"net/http"
	"time"
)

// ==================   PRINCIPES DE CONCEPTION ==================
// 1. Séparation claire entre la logique métier (services) et la persistance (storage).
// 2. Utilisation de structures Go pour modéliser les données JSON reçues de l'API.
// 3. Respect strict du "Rate Limit" imposé par le plan gratuit de football-data.org (10 requêtes/minute).
// 4. Gestion robuste des erreurs avec des messages clairs pour faciliter le debugging.
// 5. Encapsulation de la logique d'importation dans une fonction dédiée (InitialPopulate) pour une meilleure maintenabilité.

/*
	la communication entre le client et le serveur repose entièrement sur cet échange JSON.
	En Go, tout repose sur les Struct Tags (les annotations entre backticks `json:"..."`)
	qui servent de dictionnaire de traduction entre le monde du JSON (souvent en camelCase)
	et le monde de Go (en PascalCase).
	1. Le Décodage (Flux API-> Struct Go)

	C'est ce que on fait avec json.NewDecoder. On utilise un Decoder lorsqu'on lit
	un flux de données (un io.Reader), comme le corps d'une réponse HTTP.

	Pourquoi utiliser NewDecoder
	C'est plus efficace en mémoire car il traite les données au fur et à mesure
	qu'elles arrivent sur le réseau, sans avoir à charger tout le texte en une seule grosse chaîne de caractères.
*/

// ======================= ======================== ===============

// FetchAndSaveLeague contacte l'API externe pour récupérer les métadonnées d'une compétition.
// Cette fonction répond à la contrainte de peupler la base de données via une API dynamique[cite: 19, 23].
func FetchAndSaveLeague(db *sql.DB, token string, leagueCode string) error {
	// Construction de l'URL pour une compétition spécifique (ex: PL, FL1)
	url := fmt.Sprintf("https://api.football-data.org/v4/competitions/%s", leagueCode)

	// Initialisation de la requête GET
	req, _ := http.NewRequest("GET", url, nil)

	// Ajout du jeton d'authentification requis par les conditions d'utilisation de l'API
	req.Header.Set("X-Auth-Token", token)

	// Exécution de l'appel HTTP via le client standard de Go
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		return fmt.Errorf("erreur API League (Statut %d): %v", resp.StatusCode, err)
	}
	// Fermeture du corps de la réponse pour libérer les ressources (important pour éviter les fuites de mémoire)
	defer resp.Body.Close()

	// Décodage du flux JSON reçu directement dans la structure Go définie dans le stockage
	var data storage.CompetitionResponse
	// On crée le décodeur sur le flux (resp.Body)
	// Decode(&team) "verse" les données dans la struct via son pointeur
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return err
	}

	// Appel de la couche de persistance pour sauvegarder ou mettre à jour la ligue
	return storage.SaveLeague(db, data.Id, data.Name, data.Code)
}

// FetchAndSaveMatches récupère l'historique des matchs pour une saison précise.
// Les données récupérées serviront de base pour l'algorithme de prédiction statistique.
func FetchAndSaveMatches(db *sql.DB, token string, leagueCode string, season int) error {
	// RÉCUPÉRATION DYNAMIQUE DE L'ID
	// On cherche l'ID en base associé au code (ex: "PL") pour éviter les erreurs d'intégrité.
	leagueID, err := storage.GetLeagueIDByCode(db, leagueCode)
	if err != nil {
		return fmt.Errorf("impossible de procéder aux matchs : %v", err)
	}

	// Utilisation du paramètre "season" pour obtenir les données historiques dynamiques
	url := fmt.Sprintf("https://api.football-data.org/v4/competitions/%s/matches?season=%d", leagueCode, season)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("X-Auth-Token", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		return fmt.Errorf("erreur API Matches pour la saison %d: %v", season, err)
	}
	defer resp.Body.Close()

	var data storage.MatchesResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return err
	}

	// Itération sur chaque match reçu dans la réponse JSON
	for _, m := range data.Matches {
		// IMPORTANT : Sauvegarder les équipes AVANT les matchs à cause des clés étrangères (Foreign Keys).
		// Chaque match contient les informations minimales sur les équipes (Home et Away).
		storage.SaveTeam(db, m.HomeTeam.Id, m.HomeTeam.Name, m.HomeTeam.ShortName, m.HomeTeam.Crest)
		storage.SaveTeam(db, m.AwayTeam.Id, m.AwayTeam.Name, m.AwayTeam.ShortName, m.AwayTeam.Crest)

		// Sauvegarde du match avec son score et son statut (FINISHED, SCHEDULED, etc.)
		err := storage.SaveMatch(db, leagueID, season, m)
		if err != nil {
			fmt.Printf("Erreur lors de l'insertion du match %d: %v\n", m.Id, err)
		}
	}

	fmt.Printf("Import réussi : Ligue %s | Saison %d | %d matchs\n", leagueCode, season, len(data.Matches))
	return nil
}


/*
	Là le serveur va tourner en permanence et on gros chaque 10 minutes on
	essaie de voir si y'a de nouveau match à importer pour la saison en cours
	on les fetch à la DB
*/
// FetchApi lance une synchronisation périodique pour une ligue/saison.
// Elle s'exécute immédiatement une première fois, puis toutes les 10 minutes.
// L'arrêt se fait proprement via le context.
func FetchApi(ctx context.Context, db *sql.DB, token string, leagues []string, season int, apiTasks chan APITask) {
	const refreshInterval = 10 * time.Minute

	enqueueAll := func() {
		fmt.Printf("Planification du rafraîchissement de la saison %d pour %d ligues\n", season, len(leagues))

		for _, leagueCode := range leagues {
			leagueCode := leagueCode

			apiTasks <- APITask{
				Label: fmt.Sprintf("refresh matches %s season %d", leagueCode, season),
				Run: func() error {
					return FetchAndSaveMatches(db, token, leagueCode, season)
				}, 
			}
		}
	}

	// Premier envoi immédiat
	enqueueAll()

	ticker := time.NewTicker(refreshInterval)
	defer ticker.Stop()

	for {
		select {
			case <-ctx.Done():
				fmt.Println("Arrêt de FetchApi")
				return

			case <-ticker.C:
				enqueueAll()
		}
	}
}
