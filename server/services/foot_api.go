// Package services gère la logique métier liée aux interactions avec les API externes.
// Conformément aux contraintes du projet, il utilise exclusivement le package net/http[cite: 13].
package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"footix/storage"
	"io"
	"net/http"
	"strings"
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
// Package services gère la logique métier liée aux interactions avec les API externes.
// Conformément aux contraintes du projet, il utilise exclusivement le package net/http.

func executeAPIRequest(token string, url string, target any) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("création requête API impossible: %w", err)
	}
	req.Header.Set("X-Auth-Token", token)

	client := &http.Client{Timeout: 25 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("appel API impossible: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		msg := strings.TrimSpace(string(body))
		if msg == "" {
			msg = http.StatusText(resp.StatusCode)
		}
		return fmt.Errorf("API HTTP %d: %s", resp.StatusCode, msg)
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("décodage JSON impossible: %w", err)
	}

	return nil
}

// FetchAndSaveLeague contacte l'API externe pour récupérer les métadonnées d'une compétition.
func FetchAndSaveLeague(db *sql.DB, token string, leagueCode string) error {
	url := fmt.Sprintf("https://api.football-data.org/v4/competitions/%s", leagueCode)

	var data storage.CompetitionResponse
	if err := executeAPIRequest(token, url, &data); err != nil {
		return fmt.Errorf("erreur API League %s: %w", leagueCode, err)
	}

	return storage.SaveLeague(db, data.Id, data.Name, data.Code)
}

// FetchAndSaveMatches récupère l'historique des matchs pour une saison précise.
func FetchAndSaveMatches(db *sql.DB, token string, leagueCode string, season int) error {
	leagueID, err := storage.GetLeagueIDByCode(db, leagueCode)
	if err != nil {
		return fmt.Errorf("impossible de procéder aux matchs: %w", err)
	}

	url := fmt.Sprintf("https://api.football-data.org/v4/competitions/%s/matches?season=%d", leagueCode, season)

	var data storage.MatchesResponse
	if err := executeAPIRequest(token, url, &data); err != nil {
		return fmt.Errorf("erreur API Matches pour la saison %d: %w", season, err)
	}

	for _, m := range data.Matches {
		if err := storage.SaveTeam(db, m.HomeTeam.Id, m.HomeTeam.Name, m.HomeTeam.ShortName, m.HomeTeam.Crest); err != nil {
			fmt.Printf("Erreur lors de l'insertion de l'équipe domicile %d: %v\n", m.HomeTeam.Id, err)
		}
		if err := storage.SaveTeam(db, m.AwayTeam.Id, m.AwayTeam.Name, m.AwayTeam.ShortName, m.AwayTeam.Crest); err != nil {
			fmt.Printf("Erreur lors de l'insertion de l'équipe extérieure %d: %v\n", m.AwayTeam.Id, err)
		}

		err := storage.SaveMatch(db, leagueID, season, m)
		if err != nil {
			fmt.Printf("Erreur lors de l'insertion du match %d: %v\n", m.Id, err)
		}
	}

	fmt.Printf("Import réussi : Ligue %s | Saison %d | %d matchs\n", leagueCode, season, len(data.Matches))
	return nil
}

// FetchApi lance une synchronisation périodique pour une ligue/saison.
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
