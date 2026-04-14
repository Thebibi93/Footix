package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// GetLeagueIDByCode récupère l'ID numérique d'une ligue à partir de son code mnémonique (ex: 'FL1').
func GetLeagueIDByCode(db *sql.DB, leagueCode string) (int, error) {
	var id int
	query := `SELECT id FROM Leagues WHERE code = $1`
	err := db.QueryRow(query, leagueCode).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("ligue introuvable pour le code %s : %v", leagueCode, err)
	}
	return id, nil
}

func getUserRank(db *sql.DB, userID int) (int, error) {
	var rank int
	err := db.QueryRow(`
        SELECT ranked.rank
        FROM (
            SELECT u.id,
                   RANK() OVER (ORDER BY COALESCE(s.score, 0) DESC, u.username ASC) AS rank
            FROM Users u
            LEFT JOIN UserScores s ON s.user_id = u.id
        ) ranked
        WHERE ranked.id = $1`, userID).Scan(&rank)
	if err != nil {
		return 0, err
	}
	return rank, nil
}

func getUserProfileSummary(db *sql.DB, userID int, includeEmail bool) (UserProfileSummary, error) {
	var profile UserProfileSummary
	var email sql.NullString

	err := db.QueryRow(`
        SELECT
            u.id,
            u.username,
            u.email,
            COALESCE(s.score, 0) AS score,
            COALESCE(pred.total_predictions, 0) AS total_predictions,
            COALESCE(pred.correct_predictions, 0) AS correct_predictions,
            COALESCE(chat.chat_messages, 0) AS chat_messages
        FROM Users u
        LEFT JOIN UserScores s ON s.user_id = u.id
        LEFT JOIN (
            SELECT
                h.user_id,
                COUNT(*)::INT AS total_predictions,
                COUNT(*) FILTER (
                    WHERE h.actual_result IS NOT NULL
                      AND h.actual_result <> ''
                      AND h.predicted_result = h.actual_result
                )::INT AS correct_predictions
            FROM UserPredictionHistory h
            GROUP BY h.user_id
        ) pred ON pred.user_id = u.id
        LEFT JOIN (
            SELECT user_id, COUNT(*)::INT AS chat_messages
            FROM ChatMessages
            GROUP BY user_id
        ) chat ON chat.user_id = u.id
        WHERE u.id = $1`, userID).Scan(
		&profile.ID,
		&profile.Username,
		&email,
		&profile.Score,
		&profile.TotalPredictions,
		&profile.CorrectPredictions,
		&profile.ChatMessages,
	)
	if err != nil {
		return UserProfileSummary{}, err
	}

	if includeEmail && email.Valid {
		profile.Email = email.String
	}
	if profile.TotalPredictions > 0 {
		profile.Accuracy = float64(profile.CorrectPredictions) * 100 / float64(profile.TotalPredictions)
	}

	profile.Rank, err = getUserRank(db, userID)
	if err != nil {
		return UserProfileSummary{}, err
	}

	return profile, nil
}

func GetProfileByID(db *sql.DB, userID int) (UserProfileSummary, error) {
	return getUserProfileSummary(db, userID, true)
}

func GetUserByID(db *sql.DB, userID int) (UserProfileSummary, error) {
	return getUserProfileSummary(db, userID, false)
}

func ListUsers(db *sql.DB) ([]PublicUser, error) {
	rows, err := db.Query(`SELECT id, username FROM Users ORDER BY username ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]PublicUser, 0)
	for rows.Next() {
		var u PublicUser
		if err := rows.Scan(&u.ID, &u.Username); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func GetUserByIdentifier(db *sql.DB, identifier string) (User, error) {
	var user User
	err := db.QueryRow(`
        SELECT id, username, email, password_hash
        FROM Users
        WHERE username = $1 OR email = $1`, identifier).
		Scan(&user.Id, &user.Username, &user.Email, &user.PasswordHash)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func CreateUser(db *sql.DB, username, email, passwordHash string) (PublicUser, error) {
	var user PublicUser
	err := db.QueryRow(`
        INSERT INTO Users (username, email, password_hash)
        VALUES ($1, $2, $3)
        RETURNING id, username, email`, username, email, passwordHash).
		Scan(&user.ID, &user.Username, &user.Email)
	if err != nil {
		return PublicUser{}, err
	}

	_, _ = db.Exec(`INSERT INTO UserScores (user_id, score) VALUES ($1, 0) ON CONFLICT (user_id) DO NOTHING`, user.ID)
	return user, nil
}

func UpdateUserProfile(db *sql.DB, userID int, payload UpdateProfilePayload, passwordHashFunc func(string) string) (PublicUser, error) {
	var currentUsername, currentEmail, currentPasswordHash string
	err := db.QueryRow(`SELECT username, email, password_hash FROM Users WHERE id = $1`, userID).
		Scan(&currentUsername, &currentEmail, &currentPasswordHash)
	if err != nil {
		return PublicUser{}, err
	}

	newUsername := currentUsername
	newEmail := currentEmail
	newPasswordHash := currentPasswordHash

	if payload.Username != "" {
		newUsername = payload.Username
	}
	if payload.Email != "" {
		newEmail = payload.Email
	}
	if payload.Password != "" {
		newPasswordHash = passwordHashFunc(payload.Password)
	}

	_, err = db.Exec(`
        UPDATE Users
        SET username = $1, email = $2, password_hash = $3
        WHERE id = $4`, newUsername, newEmail, newPasswordHash, userID)
	if err != nil {
		return PublicUser{}, err
	}

	return PublicUser{ID: userID, Username: newUsername, Email: newEmail}, nil
}

func ListUserPredictionHistory(db *sql.DB, userID int) ([]PredictionHistoryItem, error) {
	query := `
        SELECT
            h.user_id,
            h.match_id,
            h.predicted_result,
            COALESCE(h.actual_result, ''),
            TO_CHAR(h.prediction_date AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"'),
            hm.name,
            am.name,
            m.utc_date
        FROM UserPredictionHistory h
        JOIN Matches m ON m.id = h.match_id
        JOIN Teams hm ON hm.id = m.home_team_id
        JOIN Teams am ON am.id = m.away_team_id
        WHERE h.user_id = $1
        ORDER BY h.prediction_date DESC`

	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	predictions := make([]PredictionHistoryItem, 0)
	for rows.Next() {
		var item PredictionHistoryItem
		var matchDate time.Time
		if err := rows.Scan(
			&item.UserID,
			&item.MatchID,
			&item.PredictedResult,
			&item.ActualResult,
			&item.PredictionDate,
			&item.HomeTeam,
			&item.AwayTeam,
			&matchDate,
		); err != nil {
			return nil, err
		}
		item.MatchDate = matchDate.UTC().Format(time.RFC3339)
		predictions = append(predictions, item)
	}

	return predictions, rows.Err()
}

func GetEnrichedMatchStats(db *sql.DB, matchID string) (MatchStats, error) {
	stats, err := GetMatchStats(db, matchID)
	if err != nil {
		return MatchStats{}, err
	}

	var homeID, awayID int
	err = db.QueryRow(`SELECT home_team_id, away_team_id FROM Matches WHERE id = $1`, matchID).
		Scan(&homeID, &awayID)
	if err == nil {
		stats.HomeLastResults = getLastResultsString(db, homeID, 5)
		stats.AwayLastResults = getLastResultsString(db, awayID, 5)
	}

	return stats, nil
}

func getLastResultsString(db *sql.DB, teamID int, limit int) string {
	query := `
        SELECT home_team_id, away_team_id, home_score, away_score
        FROM Matches
        WHERE status = 'FINISHED'
          AND (home_team_id = $1 OR away_team_id = $1)
        ORDER BY utc_date DESC
        LIMIT $2`

	rows, err := db.Query(query, teamID, limit)
	if err != nil {
		return ""
	}
	defer rows.Close()

	results := make([]string, 0, limit)
	for rows.Next() {
		var homeID, awayID int
		var homeScore, awayScore sql.NullInt64
		if err := rows.Scan(&homeID, &awayID, &homeScore, &awayScore); err != nil {
			return fmt.Sprint(results)
		}
		if !homeScore.Valid || !awayScore.Valid {
			continue
		}
		if teamID == homeID {
			switch {
			case homeScore.Int64 > awayScore.Int64:
				results = append(results, "W")
			case homeScore.Int64 == awayScore.Int64:
				results = append(results, "D")
			default:
				results = append(results, "L")
			}
		} else {
			switch {
			case awayScore.Int64 > homeScore.Int64:
				results = append(results, "W")
			case awayScore.Int64 == homeScore.Int64:
				results = append(results, "D")
			default:
				results = append(results, "L")
			}
		}
	}

	result := ""
	for i, s := range results {
		if i > 0 {
			result += " "
		}
		result += s
	}
	return result
}

func ListScores(db *sql.DB) ([]ScoreEntry, error) {
	query := `
        SELECT u.id, u.username, COALESCE(s.score, 0) AS score
        FROM Users u
        LEFT JOIN UserScores s ON s.user_id = u.id
        ORDER BY score DESC, u.username ASC`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	leaderboard := make([]ScoreEntry, 0)
	for rows.Next() {
		var item ScoreEntry
		if err := rows.Scan(&item.UserID, &item.Username, &item.Score); err != nil {
			return nil, err
		}
		leaderboard = append(leaderboard, item)
	}
	return leaderboard, rows.Err()
}

func SaveUserPrediction(db *sql.DB, userID int, matchID int, predictedResult string) error {
	var status string
	var kickoff time.Time
	var homeScore, awayScore sql.NullInt64
	err := db.QueryRow(`SELECT status, utc_date, home_score, away_score FROM Matches WHERE id = $1`, matchID).
		Scan(&status, &kickoff, &homeScore, &awayScore)
	if err != nil {
		return err
	}

	statusUpper := strings.ToUpper(strings.TrimSpace(status))
	if statusUpper == "FINISHED" || kickoff.UTC().Before(time.Now().UTC()) {
		return errors.New("Impossible de prédire un match déjà commencé ou terminé")
	}

	_, err = db.Exec(`
        INSERT INTO UserPredictions (user_id, match_id, predicted_result)
        VALUES ($1, $2, $3)
        ON CONFLICT (user_id, match_id) DO UPDATE
        SET predicted_result = EXCLUDED.predicted_result`, userID, matchID, predictedResult)
	if err != nil {
		return err
	}

	actualResult := actualResultFromScores(homeScore, awayScore)
	_, err = db.Exec(`
        INSERT INTO UserPredictionHistory (user_id, match_id, predicted_result, actual_result, prediction_date)
        VALUES ($1, $2, $3, $4, NOW())
        ON CONFLICT (user_id, match_id) DO UPDATE
        SET predicted_result = EXCLUDED.predicted_result,
            actual_result = EXCLUDED.actual_result,
            prediction_date = NOW()`, userID, matchID, predictedResult, actualResult)
	return err
}

func actualResultFromScores(homeScore, awayScore sql.NullInt64) string {
	if !homeScore.Valid || !awayScore.Valid {
		return ""
	}
	if homeScore.Int64 > awayScore.Int64 {
		return "HOME_WIN"
	}
	if awayScore.Int64 > homeScore.Int64 {
		return "AWAY_WIN"
	}
	return "DRAW"
}

func RecalculateScores(db *sql.DB) error {
	_, err := db.Exec(`
        UPDATE UserPredictionHistory h
        SET actual_result = CASE
            WHEN m.home_score > m.away_score THEN 'HOME_WIN'
            WHEN m.away_score > m.home_score THEN 'AWAY_WIN'
            ELSE 'DRAW'
        END
        FROM Matches m
        WHERE h.match_id = m.id AND m.status = 'FINISHED'`)
	if err != nil {
		return err
	}

	_, _ = db.Exec(`
        INSERT INTO UserScores (user_id, score)
        SELECT id, 0 FROM Users
        ON CONFLICT (user_id) DO NOTHING`)

	_, err = db.Exec(`
        UPDATE UserScores s
        SET score = computed.score
        FROM (
            SELECT h.user_id, COUNT(*)::INT AS score
            FROM UserPredictionHistory h
            WHERE h.actual_result IS NOT NULL
              AND h.actual_result <> ''
              AND h.predicted_result = h.actual_result
            GROUP BY h.user_id
        ) AS computed
        WHERE s.user_id = computed.user_id`)
	if err != nil {
		return err
	}

	_, _ = db.Exec(`
        UPDATE UserScores
        SET score = 0
        WHERE user_id NOT IN (
            SELECT DISTINCT user_id
            FROM UserPredictionHistory
            WHERE actual_result IS NOT NULL
              AND actual_result <> ''
              AND predicted_result = actual_result
        )`)

	return nil
}

func GetOrCreateMainChatRoomIDByMatchID(tx *sql.Tx, matchID int64) (int64, error) {
	var roomID int64
	err := tx.QueryRow(`SELECT id FROM ChatRooms WHERE match_id = $1 AND room_type = 'main'`, matchID).Scan(&roomID)
	if err == nil {
		return roomID, nil
	}
	if err != sql.ErrNoRows {
		return 0, err
	}

	err = tx.QueryRow(`
        INSERT INTO ChatRooms (match_id, room_type)
        VALUES ($1, 'main')
        RETURNING id`, matchID).Scan(&roomID)
	if err != nil {
		return 0, err
	}
	return roomID, nil
}

func ListChatMessagesByMatchID(db *sql.DB, matchID int64, afterSeq int64, limit int) ([]ChatMessage, error) {
	rows, err := db.Query(`
        SELECT m.id, r.id, r.match_id, m.seq_in_room, m.user_id, u.username, m.message,
               TO_CHAR(m.created_at AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"')
        FROM ChatMessages m
        JOIN ChatRooms r ON r.id = m.chat_room_id
        JOIN Users u ON u.id = m.user_id
        WHERE r.match_id = $1
          AND r.room_type = 'main'
          AND m.seq_in_room > $2
        ORDER BY m.seq_in_room ASC
        LIMIT $3`, matchID, afterSeq, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := make([]ChatMessage, 0)
	for rows.Next() {
		var item ChatMessage
		if err := rows.Scan(&item.ID, &item.RoomID, &item.MatchID, &item.SeqInChat, &item.UserID, &item.Username, &item.Message, &item.FeedbackDate); err != nil {
			return nil, err
		}
		messages = append(messages, item)
	}
	return messages, rows.Err()
}

func ListChatMessagesByUserID(db *sql.DB, userID int64, limit int) ([]ChatMessage, error) {
	rows, err := db.Query(`
        SELECT m.id, r.id, r.match_id, m.seq_in_room, m.user_id, u.username, m.message,
               TO_CHAR(m.created_at AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"')
        FROM ChatMessages m
        JOIN ChatRooms r ON r.id = m.chat_room_id
        JOIN Users u ON u.id = m.user_id
        WHERE m.user_id = $1
        ORDER BY m.created_at DESC
        LIMIT $2`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := make([]ChatMessage, 0)
	for rows.Next() {
		var item ChatMessage
		if err := rows.Scan(&item.ID, &item.RoomID, &item.MatchID, &item.SeqInChat, &item.UserID, &item.Username, &item.Message, &item.FeedbackDate); err != nil {
			return nil, err
		}
		messages = append(messages, item)
	}
	return messages, rows.Err()
}

func CreateChatMessageByMatchID(db *sql.DB, matchID int64, userID int, message string) (ChatMessage, error) {
	ctx := context.Background()

	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return ChatMessage{}, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var matchStatus string
	if err = tx.QueryRow(`SELECT status FROM Matches WHERE id = $1`, matchID).Scan(&matchStatus); err != nil {
		return ChatMessage{}, err
	}
	if strings.EqualFold(strings.TrimSpace(matchStatus), "FINISHED") {
		return ChatMessage{}, errors.New("chat fermé pour un match terminé")
	}

	var existingUserID int
	if err = tx.QueryRow(`SELECT id FROM Users WHERE id = $1`, userID).Scan(&existingUserID); err != nil {
		return ChatMessage{}, err
	}

	roomID, err := GetOrCreateMainChatRoomIDByMatchID(tx, matchID)
	if err != nil {
		return ChatMessage{}, err
	}

	_, err = tx.Exec(`
        INSERT INTO ChatRoomCounters (chat_room_id, last_seq)
        VALUES ($1, 0)
        ON CONFLICT (chat_room_id) DO NOTHING`, roomID)
	if err != nil {
		return ChatMessage{}, err
	}

	var nextSeq int64
	err = tx.QueryRow(`
        UPDATE ChatRoomCounters
        SET last_seq = last_seq + 1
        WHERE chat_room_id = $1
        RETURNING last_seq`, roomID).Scan(&nextSeq)
	if err != nil {
		return ChatMessage{}, err
	}

	var created ChatMessage
	created.MatchID = matchID
	created.RoomID = roomID
	created.Message = message

	err = tx.QueryRow(`
        INSERT INTO ChatMessages (chat_room_id, seq_in_room, user_id, message)
        VALUES ($1, $2, $3, $4)
        RETURNING id, seq_in_room, user_id,
                  TO_CHAR(created_at AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"')`,
		roomID, nextSeq, userID, message).
		Scan(&created.ID, &created.SeqInChat, &created.UserID, &created.FeedbackDate)
	if err != nil {
		return ChatMessage{}, err
	}

	err = tx.QueryRow(`SELECT username FROM Users WHERE id = $1`, userID).Scan(&created.Username)
	if err != nil {
		return ChatMessage{}, err
	}

	if err = tx.Commit(); err != nil {
		return ChatMessage{}, err
	}

	return created, nil
}
