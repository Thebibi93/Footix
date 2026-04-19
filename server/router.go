package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"footix/storage"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const sessionCookieName = "footix_user_id"

func RegisterRoutes(db *sql.DB) {
	registerJSONRoute("/api/profile", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			profileHandler(w, r, db)
		case http.MethodPost:
			updateProfileHandler(w, r, db)
		default:
			writeError(w, r, http.StatusMethodNotAllowed, "Méthode non autorisée")
		}
	})

	registerJSONRoute("/api/my-predictions", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, r, http.StatusMethodNotAllowed, "Méthode non autorisée")
			return
		}
		myPredictionsHandler(w, r, db)
	})

	registerJSONRoute("/api/users", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, r, http.StatusMethodNotAllowed, "Méthode non autorisée")
			return
		}
		if strings.TrimSpace(r.URL.Query().Get("userId")) != "" {
			getUserByIdHandler(w, r, db)
			return
		}
		getUsersHandler(w, r, db)
	})

	registerJSONRoute("/api/leagues", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, r, http.StatusMethodNotAllowed, "Méthode non autorisée")
			return
		}
		getLeaguesHandler(w, r, db)
	})

	registerJSONRoute("/api/matches", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, r, http.StatusMethodNotAllowed, "Méthode non autorisée")
			return
		}
		getMatchesHandler(w, r, db)
	})

	registerJSONRoute("/api/stats", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, r, http.StatusMethodNotAllowed, "Méthode non autorisée")
			return
		}
		getStatsHandler(w, r, db)
	})

	registerJSONRoute("/api/scores", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getScoresHandler(w, r, db)
		case http.MethodPost:
			recalculateScoresHandler(w, r, db)
		default:
			writeError(w, r, http.StatusMethodNotAllowed, "Méthode non autorisée")
		}
	})

	registerJSONRoute("/api/feedbacks", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, r, http.StatusMethodNotAllowed, "Méthode non autorisée")
			return
		}
		getFeedbacksHandler(w, r, db)
	})

	registerJSONRoute("/api/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, r, http.StatusMethodNotAllowed, "Méthode non autorisée")
			return
		}
		loginHandler(w, r, db)
	})

	registerJSONRoute("/api/signup", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, r, http.StatusMethodNotAllowed, "Méthode non autorisée")
			return
		}
		signupHandler(w, r, db)
	})

	registerJSONRoute("/api/logout", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, r, http.StatusMethodNotAllowed, "Méthode non autorisée")
			return
		}
		logoutHandler(w, r, db)
	})

	registerJSONRoute("/api/update-profile", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, r, http.StatusMethodNotAllowed, "Méthode non autorisée")
			return
		}
		updateProfileHandler(w, r, db)
	})

	registerJSONRoute("/api/predict", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, r, http.StatusMethodNotAllowed, "Méthode non autorisée")
			return
		}
		submitPredictionHandler(w, r, db)
	})

	registerJSONRoute("/api/feedback", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, r, http.StatusMethodNotAllowed, "Méthode non autorisée")
			return
		}
		submitFeedbackHandler(w, r, db)
	})
}

func registerJSONRoute(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		if handlePreflight(w, r) {
			return
		}
		handler(w, r)
	})
}

func setCommonHeaders(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if origin == "" {
		origin = "*"
	}
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Vary", "Origin")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-User-ID")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
}

func handlePreflight(w http.ResponseWriter, r *http.Request) bool {
	setCommonHeaders(w, r)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return true
	}
	return false
}

func writeJSON(w http.ResponseWriter, r *http.Request, status int, payload any) {
	setCommonHeaders(w, r)
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, r *http.Request, status int, message string) {
	writeJSON(w, r, status, map[string]string{"error": message})
}

func decodeJSONBody(r *http.Request, dst any) error {
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}

func hashPassword(password string) string {
	sum := sha256.Sum256([]byte(password))
	return hex.EncodeToString(sum[:])
}

func passwordMatches(storedValue, plainPassword string) bool {
	if storedValue == "" {
		return false
	}
	candidate := hashPassword(plainPassword)
	return storedValue == candidate || storedValue == plainPassword
}

func setSessionCookie(w http.ResponseWriter, userID int) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    strconv.Itoa(userID),
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   7 * 24 * 60 * 60,
	})
}

func clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	})
}

func getAuthenticatedUserID(r *http.Request) (int, error) {
	if cookie, err := r.Cookie(sessionCookieName); err == nil {
		id, convErr := strconv.Atoi(cookie.Value)
		if convErr == nil && id > 0 {
			return id, nil
		}
	}

	if header := strings.TrimSpace(r.Header.Get("X-User-ID")); header != "" {
		id, err := strconv.Atoi(header)
		if err == nil && id > 0 {
			return id, nil
		}
	}

	return 0, errors.New("utilisateur non authentifié")
}

func normalizePredictionResult(value string) (string, error) {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case "1", "HOME", "HOME_WIN", "HOMEWIN", "DOMICILE":
		return "HOME_WIN", nil
	case "X", "N", "DRAW", "NUL":
		return "DRAW", nil
	case "2", "AWAY", "AWAY_WIN", "AWAYWIN", "EXTERIEUR", "EXTÉRIEUR":
		return "AWAY_WIN", nil
	default:
		return "", errors.New("predictedResult doit valoir HOME_WIN, DRAW ou AWAY_WIN")
	}
}

func getLeaguesHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	leagues, err := storage.GetLeagues(db)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "Erreur lors de la récupération des ligues")
		return
	}
	writeJSON(w, r, http.StatusOK, leagues)
}

func getMatchesHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if matchIDStr := strings.TrimSpace(r.URL.Query().Get("matchId")); matchIDStr != "" {
		matchID, err := strconv.Atoi(matchIDStr)
		if err != nil || matchID <= 0 {
			writeError(w, r, http.StatusBadRequest, "Paramètre matchId invalide")
			return
		}
		match, err := storage.GetMatchByID(db, matchID)
		if err == sql.ErrNoRows {
			writeError(w, r, http.StatusNotFound, "Match introuvable")
			return
		}
		if err != nil {
			writeError(w, r, http.StatusInternalServerError, "Erreur lors de la récupération du match")
			return
		}
		writeJSON(w, r, http.StatusOK, match)
		return
	}

	leagueCode := strings.TrimSpace(r.URL.Query().Get("league"))
	if leagueCode == "" {
		writeError(w, r, http.StatusBadRequest, "Le paramètre 'league' est requis")
		return
	}

	bucket := strings.TrimSpace(r.URL.Query().Get("bucket"))
	if bucket != "upcoming" && bucket != "past" {
		writeError(w, r, http.StatusBadRequest, "Le paramètre 'bucket' doit être à upcoming ou past ")
		return
	}

	page := 1
	if value := strings.TrimSpace(r.URL.Query().Get("page")); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil && parsed > 0 {
			page = parsed
		}
	}

	pageSize := 12
	if value := strings.TrimSpace(r.URL.Query().Get("pageSize")); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil && parsed > 0 && parsed <= 50 {
			pageSize = parsed
		}
	}

	response, err := storage.GetMatchesPageByLeague(db, leagueCode, bucket, page, pageSize)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "Erreur lors de la récupération des matchs")
		return
	}

	writeJSON(w, r, http.StatusOK, response)
}

func getStatsHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	matchID := strings.TrimSpace(r.URL.Query().Get("matchId"))
	if matchID == "" {
		writeError(w, r, http.StatusBadRequest, "Paramètre matchId manquant")
		return
	}

	stats, err := storage.GetEnrichedMatchStats(db, matchID)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "Erreur lors du calcul des statistiques")
		return
	}

	writeJSON(w, r, http.StatusOK, stats)
}

func profileHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	userID, err := getAuthenticatedUserID(r)
	if err != nil {
		writeError(w, r, http.StatusUnauthorized, "Utilisateur non connecté")
		return
	}

	user, err := storage.GetProfileByID(db, userID)
	if err == sql.ErrNoRows {
		writeError(w, r, http.StatusNotFound, "Utilisateur introuvable")
		return
	}
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "Erreur lors de la récupération du profil")
		return
	}

	writeJSON(w, r, http.StatusOK, user)
}

func myPredictionsHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	userID, err := getAuthenticatedUserID(r)
	if err != nil {
		writeError(w, r, http.StatusUnauthorized, "Utilisateur non connecté")
		return
	}

	predictions, err := storage.ListUserPredictionHistory(db, userID)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "Erreur lors de la récupération des prédictions")
		return
	}

	writeJSON(w, r, http.StatusOK, predictions)
}

func getUsersHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	users, err := storage.ListUsers(db)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "Erreur lors de la récupération des utilisateurs")
		return
	}
	writeJSON(w, r, http.StatusOK, users)
}

func getUserByIdHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	userID, err := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("userId")))
	if err != nil || userID <= 0 {
		writeError(w, r, http.StatusBadRequest, "Paramètre userId invalide")
		return
	}

	user, err := storage.GetUserByID(db, userID)
	if err == sql.ErrNoRows {
		writeError(w, r, http.StatusNotFound, "Utilisateur introuvable")
		return
	}
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "Erreur lors de la récupération de l'utilisateur")
		return
	}

	writeJSON(w, r, http.StatusOK, user)
}

func getScoresHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	leaderboard, err := storage.ListScores(db)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "Erreur lors de la récupération du classement")
		return
	}
	writeJSON(w, r, http.StatusOK, leaderboard)
}

func getFeedbacksHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	matchIDStr := strings.TrimSpace(r.URL.Query().Get("matchId"))
	userIDStr := strings.TrimSpace(r.URL.Query().Get("userId"))
	afterSeqStr := strings.TrimSpace(r.URL.Query().Get("afterSeq"))
	limitStr := strings.TrimSpace(r.URL.Query().Get("limit"))

	limit := 100
	if limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 && parsed <= 500 {
			limit = parsed
		}
	}

	if matchIDStr == "" && userIDStr == "" {
		writeError(w, r, http.StatusBadRequest, "Il faut fournir matchId ou userId")
		return
	}

	switch {
	case matchIDStr != "":
		matchID, convErr := strconv.ParseInt(matchIDStr, 10, 64)
		if convErr != nil || matchID <= 0 {
			writeError(w, r, http.StatusBadRequest, "Paramètre matchId invalide")
			return
		}

		afterSeq := int64(0)
		if afterSeqStr != "" {
			afterSeq, convErr = strconv.ParseInt(afterSeqStr, 10, 64)
			if convErr != nil || afterSeq < 0 {
				writeError(w, r, http.StatusBadRequest, "Paramètre afterSeq invalide")
				return
			}
		}

		messages, err := storage.ListChatMessagesByMatchID(db, matchID, afterSeq, limit)
		if err != nil {
			writeError(w, r, http.StatusInternalServerError, "Erreur lors de la récupération des messages")
			return
		}
		writeJSON(w, r, http.StatusOK, messages)

	case userIDStr != "":
		userID, convErr := strconv.ParseInt(userIDStr, 10, 64)
		if convErr != nil || userID <= 0 {
			writeError(w, r, http.StatusBadRequest, "Paramètre userId invalide")
			return
		}

		messages, err := storage.ListChatMessagesByUserID(db, userID, limit)
		if err != nil {
			writeError(w, r, http.StatusInternalServerError, "Erreur lors de la récupération des messages")
			return
		}
		writeJSON(w, r, http.StatusOK, messages)
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var payload storage.AuthPayload
	if err := decodeJSONBody(r, &payload); err != nil {
		writeError(w, r, http.StatusBadRequest, "Corps JSON invalide")
		return
	}

	identifier := strings.TrimSpace(payload.Username)
	password := payload.Password
	if identifier == "" && strings.TrimSpace(payload.Email) != "" {
		identifier = strings.TrimSpace(payload.Email)
	}
	if identifier == "" || strings.TrimSpace(password) == "" {
		writeError(w, r, http.StatusBadRequest, "username/email et password sont requis")
		return
	}

	user, err := storage.GetUserByIdentifier(db, identifier)
	if err == sql.ErrNoRows {
		writeError(w, r, http.StatusUnauthorized, "Identifiants invalides")
		return
	}
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "Erreur lors du login")
		return
	}

	if !passwordMatches(user.PasswordHash, password) {
		writeError(w, r, http.StatusUnauthorized, "Identifiants invalides")
		return
	}

	setSessionCookie(w, user.Id)
	writeJSON(w, r, http.StatusOK, map[string]any{
		"message": "Connexion réussie",
		"user": map[string]any{
			"id":       user.Id,
			"username": user.Username,
			"email":    user.Email,
		},
	})
}

func signupHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var payload storage.AuthPayload
	if err := decodeJSONBody(r, &payload); err != nil {
		writeError(w, r, http.StatusBadRequest, "Corps JSON invalide")
		return
	}

	payload.Username = strings.TrimSpace(payload.Username)
	payload.Email = strings.TrimSpace(payload.Email)
	payload.Password = strings.TrimSpace(payload.Password)

	if payload.Username == "" || payload.Email == "" || payload.Password == "" {
		writeError(w, r, http.StatusBadRequest, "username, email et password sont requis")
		return
	}

	user, err := storage.CreateUser(db, payload.Username, payload.Email, hashPassword(payload.Password))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "Impossible de créer le compte (username ou email déjà utilisé)")
		return
	}

	setSessionCookie(w, user.ID)
	writeJSON(w, r, http.StatusCreated, map[string]any{
		"message": "Compte créé",
		"user":    user,
	})
}

func logoutHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	_ = db
	clearSessionCookie(w)
	writeJSON(w, r, http.StatusOK, map[string]string{"message": "Déconnexion réussie"})
}

func updateProfileHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	userID, err := getAuthenticatedUserID(r)
	if err != nil {
		writeError(w, r, http.StatusUnauthorized, "Utilisateur non connecté")
		return
	}

	var payload storage.UpdateProfilePayload
	if err := decodeJSONBody(r, &payload); err != nil {
		writeError(w, r, http.StatusBadRequest, "Corps JSON invalide")
		return
	}

	payload.Username = strings.TrimSpace(payload.Username)
	payload.Email = strings.TrimSpace(payload.Email)
	payload.Password = strings.TrimSpace(payload.Password)

	user, err := storage.UpdateUserProfile(db, userID, payload, hashPassword)
	if err == sql.ErrNoRows {
		writeError(w, r, http.StatusNotFound, "Utilisateur introuvable")
		return
	}
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "Impossible de mettre à jour le profil")
		return
	}

	writeJSON(w, r, http.StatusOK, map[string]any{
		"message": "Profil mis à jour",
		"user":    user,
	})
}

func submitPredictionHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	userID, err := getAuthenticatedUserID(r)
	if err != nil {
		writeError(w, r, http.StatusUnauthorized, "Utilisateur non connecté")
		return
	}

	var payload storage.PredictionPayload
	if err := decodeJSONBody(r, &payload); err != nil {
		writeError(w, r, http.StatusBadRequest, "Corps JSON invalide")
		return
	}

	if payload.MatchID <= 0 {
		writeError(w, r, http.StatusBadRequest, "matchId invalide")
		return
	}

	normalizedResult, err := normalizePredictionResult(payload.PredictedResult)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	if err := storage.SaveUserPrediction(db, userID, payload.MatchID, normalizedResult); err != nil {
		if strings.Contains(err.Error(), "déjà") || strings.Contains(err.Error(), "commencé") {
			writeError(w, r, http.StatusBadRequest, err.Error())
			return
		}
		if err == sql.ErrNoRows {
			writeError(w, r, http.StatusNotFound, "Match introuvable")
			return
		}
		writeError(w, r, http.StatusInternalServerError, "Erreur lors de l'enregistrement de la prédiction")
		return
	}

	writeJSON(w, r, http.StatusOK, map[string]any{
		"message":         "Prédiction enregistrée",
		"userId":          userID,
		"matchId":         payload.MatchID,
		"predictedResult": normalizedResult,
	})
}

func submitFeedbackHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	userID, err := getAuthenticatedUserID(r)
	if err != nil {
		writeError(w, r, http.StatusUnauthorized, "Utilisateur non connecté")
		return
	}

	var payload storage.FeedbackPayload
	if err := decodeJSONBody(r, &payload); err != nil {
		writeError(w, r, http.StatusBadRequest, "Corps JSON invalide")
		return
	}

	payload.Message = strings.TrimSpace(payload.Message)
	if payload.MatchID <= 0 {
		writeError(w, r, http.StatusBadRequest, "matchId invalide")
		return
	}
	if payload.Message == "" {
		writeError(w, r, http.StatusBadRequest, "message requis")
		return
	}

	message, err := storage.CreateChatMessageByMatchID(db, int64(payload.MatchID), userID, payload.Message)
	if err != nil {
		if strings.Contains(err.Error(), "terminé") {
			writeError(w, r, http.StatusBadRequest, err.Error())
			return
		}
		if err == sql.ErrNoRows {
			writeError(w, r, http.StatusNotFound, "Match ou utilisateur introuvable")
			return
		}
		writeError(w, r, http.StatusConflict, "Impossible d'enregistrer le message pour le moment, réessaie")
		return
	}

	writeJSON(w, r, http.StatusCreated, message)
}

func recalculateScoresHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if err := storage.RecalculateScores(db); err != nil {
		writeError(w, r, http.StatusInternalServerError, "Erreur lors du recalcul des scores")
		return
	}
	writeJSON(w, r, http.StatusOK, map[string]string{"message": "Scores recalculés"})
}
