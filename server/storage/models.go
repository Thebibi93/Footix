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
    Id        int      `json:"id"`
    UtcDate   string   `json:"utcDate"`
    Status    string   `json:"status"`
    HomeTeam  TeamInfo `json:"homeTeam"`
    AwayTeam  TeamInfo `json:"awayTeam"`
    Score     struct {
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
    HomeWinProb  float64 `json:"homeWinProb"`
    AwayWinProb  float64 `json:"awayWinProb"`
    DrawProb     float64 `json:"drawProb"`
    HomeLastResults string `json:"homeLastResults"`
    AwayLastResults string `json:"awayLastResults"`
}