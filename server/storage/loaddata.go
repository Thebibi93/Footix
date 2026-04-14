package storage

import (
	"database/sql"
	"fmt"
	"strings"
)

// SaveTeam insère une équipe ou la met à jour
func SaveTeam(db *sql.DB, id int, name string, shortName string, crestURL string) error {
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
            status = EXCLUDED.status,
            utc_date = EXCLUDED.utc_date,
            season = EXCLUDED.season,
            league_id = EXCLUDED.league_id`

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

	_, err = db.Exec(`
		INSERT INTO ChatRooms (match_id, room_type)
		VALUES ($1, 'main')
		ON CONFLICT (match_id, room_type) DO NOTHING`, m.Id)
	return err
}

// SaveLeague insère une compétition ou la met à jour si elle existe déjà
func SaveLeague(db *sql.DB, id int, name string, code string) error {
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

func GetLeagues(db *sql.DB) ([]League, error) {
	rows, err := db.Query(`
		SELECT id, name, code
		FROM Leagues
		ORDER BY name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	leagues := make([]League, 0)
	for rows.Next() {
		var league League
		if err := rows.Scan(&league.Id, &league.Name, &league.Code); err != nil {
			return nil, err
		}
		leagues = append(leagues, league)
	}

	return leagues, rows.Err()
}

/*
	PPM =  (Victoires x 3) + Nuls​  /  Nombre_de_Matchs_Joués
	Ex: Si une équipe a 10 matchs joués, 5 victoires, 3 nuls et 2 défaites :
	PPM = (5 x 3) + 3 / 10 = 1.8
*/

// GetMatchStats calcule les probabilités de victoire pour un match donné.
func GetMatchStats(db *sql.DB, matchID string) (MatchStats, error) {
	var stats MatchStats
	var homeID, awayID int

	err := db.QueryRow("SELECT home_team_id, away_team_id FROM Matches WHERE id = $1", matchID).Scan(&homeID, &awayID)
	if err != nil {
		return stats, err
	}

	homePoints := getTeamPoints(db, homeID)
	awayPoints := getTeamPoints(db, awayID)

	total := homePoints + awayPoints
	if total > 0 {
		stats.HomeWinProb = (homePoints / total) * 100
		stats.AwayWinProb = (awayPoints / total) * 100
		stats.DrawProb = 100 - (stats.HomeWinProb + stats.AwayWinProb)
	} else {
		stats.HomeWinProb, stats.AwayWinProb, stats.DrawProb = 33.3, 33.3, 33.4
	}

	return stats, nil
}

func getTeamPoints(db *sql.DB, teamID int) float64 {
	var points float64
	query := `
        SELECT COALESCE(SUM(
            CASE
                WHEN (home_team_id = $1 AND home_score > away_score) OR (away_team_id = $1 AND away_score > home_score) THEN 3
                WHEN home_score = away_score THEN 1
                ELSE 0
            END
        ), 0)
        FROM Matches
        WHERE (home_team_id = $1 OR away_team_id = $1) AND status = 'FINISHED'`

	_ = db.QueryRow(query, teamID).Scan(&points)
	return points
}

func matchBucketClause(bucket string) string {
	switch strings.ToLower(strings.TrimSpace(bucket)) {
	case "past", "finished":
		return "(m.status = 'FINISHED' OR m.utc_date < NOW() - INTERVAL '3 hours')"
	case "all":
		return "TRUE"
	default:
		return "(m.status <> 'FINISHED' AND m.utc_date >= NOW() - INTERVAL '3 hours')"
	}
}

func scanMatchRow(scanner interface{ Scan(dest ...any) error }) (MatchData, error) {
	var m MatchData
	var homeScore sql.NullInt64
	var awayScore sql.NullInt64

	err := scanner.Scan(
		&m.Id,
		&m.LeagueID,
		&m.LeagueCode,
		&m.LeagueName,
		&m.Season,
		&m.UtcDate,
		&m.Status,
		&homeScore,
		&awayScore,
		&m.HomeTeam.Id,
		&m.HomeTeam.Name,
		&m.HomeTeam.ShortName,
		&m.HomeTeam.Crest,
		&m.AwayTeam.Id,
		&m.AwayTeam.Name,
		&m.AwayTeam.ShortName,
		&m.AwayTeam.Crest,
	)
	if err != nil {
		return MatchData{}, err
	}

	if homeScore.Valid {
		m.Score.FullTime.Home = int(homeScore.Int64)
	}
	if awayScore.Valid {
		m.Score.FullTime.Away = int(awayScore.Int64)
	}

	return m, nil
}

func GetMatchByID(db *sql.DB, matchID int) (MatchData, error) {
	query := `
        SELECT
            m.id,
            l.id,
            l.code,
            l.name,
            m.season,
            TO_CHAR(m.utc_date AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"'),
            m.status,
            m.home_score,
            m.away_score,
            t1.id,
            t1.name,
            t1.short_name,
            t1.crest_url,
            t2.id,
            t2.name,
            t2.short_name,
            t2.crest_url
        FROM Matches m
        JOIN Leagues l ON m.league_id = l.id
        JOIN Teams t1 ON m.home_team_id = t1.id
        JOIN Teams t2 ON m.away_team_id = t2.id
        WHERE m.id = $1`

	row := db.QueryRow(query, matchID)
	return scanMatchRow(row)
}

func GetMatchesPageByLeague(db *sql.DB, leagueCode, bucket string, page, pageSize int) (MatchListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 12
	}
	if pageSize > 50 {
		pageSize = 50
	}

	clause := matchBucketClause(bucket)
	response := MatchListResponse{
		Items:    make([]MatchData, 0),
		Page:     page,
		PageSize: pageSize,
		Bucket:   strings.ToLower(strings.TrimSpace(bucket)),
	}
	if response.Bucket == "" {
		response.Bucket = "upcoming"
	}

	countQuery := fmt.Sprintf(`
        SELECT COUNT(*)
        FROM Matches m
        JOIN Leagues l ON m.league_id = l.id
        WHERE l.code = $1 AND %s`, clause)

	if err := db.QueryRow(countQuery, leagueCode).Scan(&response.Total); err != nil {
		return response, err
	}

	if response.Total == 0 {
		return response, nil
	}

	response.TotalPages = (response.Total + pageSize - 1) / pageSize
	if page > response.TotalPages {
		page = response.TotalPages
		response.Page = page
	}
	offset := (page - 1) * pageSize

	dataQuery := fmt.Sprintf(`
        SELECT
            m.id,
            l.id,
            l.code,
            l.name,
            m.season,
            TO_CHAR(m.utc_date AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"'),
            m.status,
            m.home_score,
            m.away_score,
            t1.id,
            t1.name,
            t1.short_name,
            t1.crest_url,
            t2.id,
            t2.name,
            t2.short_name,
            t2.crest_url
        FROM Matches m
        JOIN Leagues l ON m.league_id = l.id
        JOIN Teams t1 ON m.home_team_id = t1.id
        JOIN Teams t2 ON m.away_team_id = t2.id
        WHERE l.code = $1 AND %s
        ORDER BY m.utc_date DESC
        LIMIT $2 OFFSET $3`, clause)

	rows, err := db.Query(dataQuery, leagueCode, pageSize, offset)
	if err != nil {
		return response, err
	}
	defer rows.Close()

	for rows.Next() {
		m, err := scanMatchRow(rows)
		if err != nil {
			return response, err
		}
		response.Items = append(response.Items, m)
	}

	return response, rows.Err()
}

// GetMatchesByLeague garde la compatibilité avec l'ancien appel sans pagination.
func GetMatchesByLeague(db *sql.DB, leagueCode string) ([]MatchData, error) {
	response, err := GetMatchesPageByLeague(db, leagueCode, "all", 1, 50)
	if err != nil {
		return nil, err
	}
	return response.Items, nil
}
