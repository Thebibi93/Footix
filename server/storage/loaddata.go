package storage

import (
	"database/sql"
	"fmt"
)

// SaveTeam insère une équipe ou la met à jour
func SaveTeam(db *sql.DB, id int, name string, shortName string, crestURL string) error {
	// Utilisation de placeholders ($1, $2...) pour éviter les injections SQL
	query := `
		INSERT INTO Teams (id, name, short_name, crest_url)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE 
		SET name = EXCLUDED.name, short_name = EXCLUDED.short_name, crest_url = EXCLUDED.crest_url`

	_, err := db.Exec(query, id, name, shortName, crestURL)
	return err
}

func SaveMatch(db *sql.DB, leagueID int, season int, m MatchData) error {
	query := `
        INSERT INTO Matches (id, league_id, season, utc_date, home_team_id, away_team_id, home_score, away_score, status)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
        ON CONFLICT (id) DO UPDATE SET 
            home_score = EXCLUDED.home_score, 
            away_score = EXCLUDED.away_score, 
            status = EXCLUDED.status`

	_, err := db.Exec(query,
		m.Id,
		leagueID,
		season,
		m.UtcDate,
		m.HomeTeam.Id,
		m.AwayTeam.Id,
		m.Score.FullTime.Home,
		m.Score.FullTime.Away,
		m.Status,
	)
	if err != nil {
		return err
	}

	// Chaque match possède sa room principale de chat.
	_, err = db.Exec(`
		INSERT INTO ChatRooms (match_id, room_type)
		VALUES ($1, 'main')
		ON CONFLICT (match_id, room_type) DO NOTHING`, m.Id)
	return err
}

// SaveLeague insère une compétition ou la met à jour si elle existe déjà
func SaveLeague(db *sql.DB, id int, name string, code string) error {
	// Note : On utilise l'ID de l'API comme Primary Key pour faciliter les jointures
	query := `
        INSERT INTO Leagues (id, name, code)
        VALUES ($1, $2, $3)
        ON CONFLICT (id) DO UPDATE 
        SET name = EXCLUDED.name, code = EXCLUDED.code`

	_, err := db.Exec(query, id, name, code)
	if err != nil {
		return fmt.Errorf("erreur lors de la sauvegarde de la ligue %s: %v", code, err)
	}
	return nil
}

// ====================== LOGIQUE DE PRÉDICTION ======================
// ====================== ===================== ======================
// ====================== ===================== ======================

/*
	PPM =  (Victoires x 3) + Nuls​  /  Nombre_de_Matchs_Joués
	Ex: Si une équipe a 10 matchs joués, 5 victoires, 3 nuls et 2 défaites :
	PPM = (5 x 3) + 3 / 10 = 1.8
*/

// GetMatchStats calcule les probabilités de victoire pour un match donné.
func GetMatchStats(db *sql.DB, matchID string) (MatchStats, error) {
	var stats MatchStats
	var homeID, awayID int

	// Trouver les IDs des deux équipes pour ce match
	err := db.QueryRow("SELECT home_team_id, away_team_id FROM Matches WHERE id = $1", matchID).Scan(&homeID, &awayID)
	if err != nil {
		return stats, err
	}

	// Calculer le ratio de points pour chaque équipe sur la saison actuelle
	// Cette sous-requête compte les points (3 pour V, 1 pour N)
	homePoints := getTeamPoints(db, homeID)
	awayPoints := getTeamPoints(db, awayID)

	// Calcul de probabilité simple (Force Home / Force Totale)
	total := homePoints + awayPoints
	if total > 0 {
		stats.HomeWinProb = (homePoints / total) * 100
		stats.AwayWinProb = (awayPoints / total) * 100
		stats.DrawProb = 100 - (stats.HomeWinProb + stats.AwayWinProb)
	} else {
		// Si pas assez de données, on met 33% partout
		stats.HomeWinProb, stats.AwayWinProb, stats.DrawProb = 33.3, 33.3, 33.4
	}

	return stats, nil
}

// Fonction utilitaire pour calculer les points d'une équipe
func getTeamPoints(db *sql.DB, teamID int) float64 {
	var points float64
	query := `
        SELECT SUM(
            CASE 
                WHEN (home_team_id = $1 AND home_score > away_score) OR (away_team_id = $1 AND away_score > home_score) THEN 3
                WHEN home_score = away_score THEN 1
                ELSE 0 
            END
        ) FROM Matches WHERE (home_team_id = $1 OR away_team_id = $1) AND status = 'FINISHED'`

	db.QueryRow(query, teamID).Scan(&points)
	return points
}

// GetMatchesByLeague récupère tous les matchs d'une ligue avec les détails des équipes.
func GetMatchesByLeague(db *sql.DB, leagueCode string) ([]MatchData, error) {
	// Requête SQL avec jointures pour récupérer les noms et logos des équipes
	query := `
        SELECT 
            m.id, m.utc_date, m.status, m.home_score, m.away_score,
            t1.id, t1.name, t1.short_name, t1.crest_url,
            t2.id, t2.name, t2.short_name, t2.crest_url
        FROM Matches m
        JOIN Leagues l ON m.league_id = l.id
        JOIN Teams t1 ON m.home_team_id = t1.id
        JOIN Teams t2 ON m.away_team_id = t2.id
        WHERE l.code = $1
        ORDER BY m.utc_date DESC
        LIMIT 50`

	rows, err := db.Query(query, leagueCode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var matches []MatchData

	for rows.Next() {
		var m MatchData
		// On scanne les colonnes dans l'ordre de la requête SELECT
		err := rows.Scan(
			&m.Id, &m.UtcDate, &m.Status, &m.Score.FullTime.Home, &m.Score.FullTime.Away,
			&m.HomeTeam.Id, &m.HomeTeam.Name, &m.HomeTeam.ShortName, &m.HomeTeam.Crest,
			&m.AwayTeam.Id, &m.AwayTeam.Name, &m.AwayTeam.ShortName, &m.AwayTeam.Crest,
		)
		if err != nil {
			return nil, err
		}
		matches = append(matches, m)
	}

	return matches, nil
}