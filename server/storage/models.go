package storage

// CompetitionResponse représente les détails d'une ligue (ex: PL, FL1)
type CompetitionResponse struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
	Area struct {
		Name string `json:"name"`
	} `json:"area"`
}

// MatchesResponse est l'enveloppe pour la liste des matchs
type MatchesResponse struct {
	Matches []MatchData `json:"matches"`
}

// MatchData représente un match unique à insérer en base
type MatchData struct {
	Id       int      `json:"id"`
	UtcDate  string   `json:"utcDate"`
	Status   string   `json:"status"`
	HomeTeam TeamInfo `json:"homeTeam"`
	AwayTeam TeamInfo `json:"awayTeam"`
	Score    struct {
		FullTime struct {
			Home int `json:"home"`
			Away int `json:"away"`
		} `json:"fullTime"`
	} `json:"score"`
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

// Payloads HTTP

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