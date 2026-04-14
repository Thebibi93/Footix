package storage

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// CompetitionResponse représente les détails d'une ligue (ex: PL, FL1).
type CompetitionResponse struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
	Area struct {
		Name string `json:"name"`
	} `json:"area"`
}

// MatchesResponse est l'enveloppe pour la liste des matchs.
type MatchesResponse struct {
	Matches []MatchData `json:"matches"`
}

type League struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

// SeasonInfo doit supporter deux mondes :
// - l'API football-data, où `season` est souvent un objet JSON
// - notre base SQL, où `season` est stocké comme un entier
//
// Year sert au stockage / retour API interne.
// Les autres champs servent uniquement à tolérer le payload externe.
type SeasonInfo struct {
	Year            int    `json:"-"`
	ID              int    `json:"id,omitempty"`
	StartDate       string `json:"startDate,omitempty"`
	EndDate         string `json:"endDate,omitempty"`
	CurrentMatchday int    `json:"currentMatchday,omitempty"`
	Label           string `json:"-"`
	Raw             string `json:"-"`
}

func (s SeasonInfo) hasObjectMetadata() bool {
	return s.ID != 0 || s.StartDate != "" || s.EndDate != "" || s.CurrentMatchday != 0
}

func parseSeasonYear(text string) int {
	text = strings.TrimSpace(text)
	if text == "" {
		return 0
	}
	if len(text) >= 4 {
		if y, err := strconv.Atoi(text[:4]); err == nil {
			return y
		}
	}
	if y, err := strconv.Atoi(text); err == nil {
		return y
	}
	return 0
}

func (s *SeasonInfo) UnmarshalJSON(data []byte) error {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return nil
	}

	s.Raw = string(trimmed)

	switch trimmed[0] {
	case '{':
		type alias struct {
			ID              int    `json:"id,omitempty"`
			StartDate       string `json:"startDate,omitempty"`
			EndDate         string `json:"endDate,omitempty"`
			CurrentMatchday int    `json:"currentMatchday,omitempty"`
		}
		var parsed alias
		if err := json.Unmarshal(trimmed, &parsed); err != nil {
			return err
		}
		s.ID = parsed.ID
		s.StartDate = parsed.StartDate
		s.EndDate = parsed.EndDate
		s.CurrentMatchday = parsed.CurrentMatchday
		if s.Year == 0 {
			s.Year = parseSeasonYear(parsed.StartDate)
		}
		return nil
	case '"':
		var text string
		if err := json.Unmarshal(trimmed, &text); err != nil {
			return err
		}
		s.Label = text
		if s.Year == 0 {
			s.Year = parseSeasonYear(text)
		}
		return nil
	default:
		var year int
		if err := json.Unmarshal(trimmed, &year); err == nil {
			s.Year = year
			return nil
		}
		var text string
		if err := json.Unmarshal(trimmed, &text); err == nil {
			s.Label = text
			if s.Year == 0 {
				s.Year = parseSeasonYear(text)
			}
			return nil
		}
		return nil
	}
}

// MarshalJSON privilégie un entier pour les réponses internes basées sur la base SQL.
// Si on n'a que des métadonnées objet, on renvoie l'objet toléré.
func (s SeasonInfo) MarshalJSON() ([]byte, error) {
	if s.Year != 0 {
		return json.Marshal(s.Year)
	}
	if s.hasObjectMetadata() {
		type alias struct {
			ID              int    `json:"id,omitempty"`
			StartDate       string `json:"startDate,omitempty"`
			EndDate         string `json:"endDate,omitempty"`
			CurrentMatchday int    `json:"currentMatchday,omitempty"`
		}
		return json.Marshal(alias{
			ID:              s.ID,
			StartDate:       s.StartDate,
			EndDate:         s.EndDate,
			CurrentMatchday: s.CurrentMatchday,
		})
	}
	if s.Label != "" {
		return json.Marshal(s.Label)
	}
	return json.Marshal(nil)
}

// Scan permet de lire le champ `season` depuis PostgreSQL.
func (s *SeasonInfo) Scan(src any) error {
	if src == nil {
		*s = SeasonInfo{}
		return nil
	}

	switch v := src.(type) {
	case int64:
		s.Year = int(v)
		return nil
	case int32:
		s.Year = int(v)
		return nil
	case int:
		s.Year = v
		return nil
	case float64:
		s.Year = int(v)
		return nil
	case []byte:
		text := strings.TrimSpace(string(v))
		if text == "" {
			*s = SeasonInfo{}
			return nil
		}
		year, err := strconv.Atoi(text)
		if err != nil {
			return fmt.Errorf("season scan invalide: %q", text)
		}
		s.Year = year
		return nil
	case string:
		text := strings.TrimSpace(v)
		if text == "" {
			*s = SeasonInfo{}
			return nil
		}
		year, err := strconv.Atoi(text)
		if err != nil {
			return fmt.Errorf("season scan invalide: %q", text)
		}
		s.Year = year
		return nil
	default:
		return fmt.Errorf("type season SQL non supporté: %T", src)
	}
}

// Value permet d'écrire la saison en SQL si nécessaire.
func (s SeasonInfo) Value() (driver.Value, error) {
	if s.Year == 0 {
		return nil, nil
	}
	return int64(s.Year), nil
}

// MatchData représente un match unique à insérer en base et à renvoyer au front.
type MatchData struct {
	Id         int        `json:"id"`
	LeagueID   int        `json:"leagueId,omitempty"`
	LeagueCode string     `json:"leagueCode,omitempty"`
	LeagueName string     `json:"leagueName,omitempty"`
	Season     SeasonInfo `json:"season"`
	UtcDate    string     `json:"utcDate"`
	Status     string     `json:"status"`
	HomeTeam   TeamInfo   `json:"homeTeam"`
	AwayTeam   TeamInfo   `json:"awayTeam"`
	Score      struct {
		FullTime struct {
			Home int `json:"home"`
			Away int `json:"away"`
		} `json:"fullTime"`
	} `json:"score"`
}

type MatchListResponse struct {
	Items      []MatchData `json:"items"`
	Page       int         `json:"page"`
	PageSize   int         `json:"pageSize"`
	Total      int         `json:"total"`
	TotalPages int         `json:"totalPages"`
	Bucket     string      `json:"bucket"`
}

type TeamInfo struct {
	Id        int    `json:"id"`
	Name      string `json:"name"`
	ShortName string `json:"shortName"`
	Crest     string `json:"crest"`
}

type MatchStats struct {
	HomeWinProb     float64 `json:"homeWinProb"`
	AwayWinProb     float64 `json:"awayWinProb"`
	DrawProb        float64 `json:"drawProb"`
	HomeLastResults string  `json:"homeLastResults"`
	AwayLastResults string  `json:"awayLastResults"`
}

type User struct {
	Id           int    `json:"id"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	PasswordHash string `json:"passwordHash"`
}

type PublicUser struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email,omitempty"`
}

type UserProfileSummary struct {
	ID                 int     `json:"id"`
	Username           string  `json:"username"`
	Email              string  `json:"email,omitempty"`
	Score              int     `json:"score"`
	Rank               int     `json:"rank"`
	TotalPredictions   int     `json:"totalPredictions"`
	CorrectPredictions int     `json:"correctPredictions"`
	Accuracy           float64 `json:"accuracy"`
	ChatMessages       int     `json:"chatMessages"`
}

type UserPrediction struct {
	UserId          int    `json:"userId"`
	MatchId         int    `json:"matchId"`
	PredictedResult string `json:"predictedResult"`
}

type UserScore struct {
	UserId int `json:"userId"`
	Score  int `json:"score"`
}

type ScoreEntry struct {
	UserID   int    `json:"userId"`
	Username string `json:"username"`
	Score    int    `json:"score"`
}

type UserPredictionHistory struct {
	UserId          int    `json:"userId"`
	MatchId         int    `json:"matchId"`
	PredictedResult string `json:"predictedResult"`
	ActualResult    string `json:"actualResult,omitempty"`
	PredictionDate  string `json:"predictionDate"`
}

type PredictionHistoryItem struct {
	UserID          int    `json:"userId"`
	MatchID         int    `json:"matchId"`
	PredictedResult string `json:"predictedResult"`
	ActualResult    string `json:"actualResult,omitempty"`
	PredictionDate  string `json:"predictionDate"`
	HomeTeam        string `json:"homeTeam"`
	AwayTeam        string `json:"awayTeam"`
	MatchDate       string `json:"matchDate"`
}

type ChatMessage struct {
	ID           int64  `json:"id"`
	RoomID       int64  `json:"roomId,omitempty"`
	MatchID      int64  `json:"matchId"`
	SeqInChat    int64  `json:"seqInChat"`
	UserID       int64  `json:"userId"`
	Username     string `json:"username,omitempty"`
	Message      string `json:"message"`
	FeedbackDate string `json:"feedbackDate"`
}

// Payloads HTTP.
type AuthPayload struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UpdateProfilePayload struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type PredictionPayload struct {
	MatchID         int    `json:"matchId"`
	PredictedResult string `json:"predictedResult"`
}

type FeedbackPayload struct {
	MatchID int    `json:"matchId"`
	Message string `json:"message"`
}
