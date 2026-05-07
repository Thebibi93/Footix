/*
Ce fichier centralise les endpoints HTTP de Footix.
Les handlers appellent la couche storage et répondent en JSON au client React.
*/
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

// RegisterRoutes configure tous les points d’entrée HTTP de l’API Footix.
func RegisterRoutes(db *sql.DB) {
	// Usage : GET /api/profile lit le profil connecté ; POST /api/profile le met à jour.
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

	// Usage : GET /api/my-predictions renvoie les pronostics de l’utilisateur connecté.
	registerJSONRoute("/api/my-predictions", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, r, http.StatusMethodNotAllowed, "Méthode non autorisée")
			return
		}
		myPredictionsHandler(w, r, db)
	})

	// Usage : GET /api/users liste les utilisateurs ; ?userId=123 lit un profil public précis.
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

	// Usage : GET /api/leagues renvoie les compétitions disponibles.
	registerJSONRoute("/api/leagues", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, r, http.StatusMethodNotAllowed, "Méthode non autorisée")
			return
		}
		getLeaguesHandler(w, r, db)
	})

	// Usage : GET /api/matches filtre les matchs par ligue ou récupère un match par ID.
	registerJSONRoute("/api/matches", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, r, http.StatusMethodNotAllowed, "Méthode non autorisée")
			return
		}
		getMatchesHandler(w, r, db)
	})

	// Usage : GET /api/stats?matchId=123 calcule les statistiques d’un match.
	registerJSONRoute("/api/stats", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, r, http.StatusMethodNotAllowed, "Méthode non autorisée")
			return
		}
		getStatsHandler(w, r, db)
	})

	// Usage : GET /api/scores lit le classement ; POST /api/scores recalcule les scores.
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

	// Usage : GET /api/feedbacks lit les messages par match, utilisateur ou séquence.
	registerJSONRoute("/api/feedbacks", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, r, http.StatusMethodNotAllowed, "Méthode non autorisée")
			return
		}
		getFeedbacksHandler(w, r, db)
	})

	// Usage : POST /api/login connecte un utilisateur avec pseudo/email et mot de passe.
	registerJSONRoute("/api/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, r, http.StatusMethodNotAllowed, "Méthode non autorisée")
			return
		}
		loginHandler(w, r, db)
	})

	// Usage : POST /api/signup crée un compte utilisateur.
	registerJSONRoute("/api/signup", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, r, http.StatusMethodNotAllowed, "Méthode non autorisée")
			return
		}
		signupHandler(w, r, db)
	})

	// Usage : POST /api/logout ferme la session courante.
	registerJSONRoute("/api/logout", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, r, http.StatusMethodNotAllowed, "Méthode non autorisée")
			return
		}
		logoutHandler(w, r, db)
	})

	// Usage : POST /api/update-profile met à jour explicitement le profil connecté.
	registerJSONRoute("/api/update-profile", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, r, http.StatusMethodNotAllowed, "Méthode non autorisée")
			return
		}
		updateProfileHandler(w, r, db)
	})

	// Usage : POST /api/predict soumet ou remplace un pronostic.
	registerJSONRoute("/api/predict", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, r, http.StatusMethodNotAllowed, "Méthode non autorisée")
			return
		}
		submitPredictionHandler(w, r, db)
	})

	// Usage : POST /api/feedback ajoute un message dans le chat d’un match.
	registerJSONRoute("/api/feedback", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, r, http.StatusMethodNotAllowed, "Méthode non autorisée")
			return
		}
		submitFeedbackHandler(w, r, db)
	})
}

// registerJSONRoute ajoute une route JSON avec gestion CORS et pré-requête OPTIONS.
func registerJSONRoute(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		if handlePreflight(w, r) {
			return
		}
		handler(w, r)
	})
}

// setCommonHeaders applique les en-têtes CORS et JSON communs à toutes les réponses.
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

// handlePreflight intercepte les requêtes OPTIONS envoyées par le navigateur.
func handlePreflight(w http.ResponseWriter, r *http.Request) bool {
	setCommonHeaders(w, r)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return true
	}
	return false
}

// writeJSON sérialise une réponse JSON avec le statut HTTP demandé.
func writeJSON(w http.ResponseWriter, r *http.Request, status int, payload any) {
	setCommonHeaders(w, r)
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

// writeError renvoie une erreur JSON homogène au client React.
func writeError(w http.ResponseWriter, r *http.Request, status int, message string) {
	writeJSON(w, r, status, map[string]string{"error": message})
}

// decodeJSONBody lit le corps JSON d’une requête et refuse les champs inconnus.
func decodeJSONBody(r *http.Request, dst any) error {
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}

// hashPassword calcule le hash SHA-256 utilisé pour stocker ou vérifier un mot de passe.
func hashPassword(password string) string {
	sum := sha256.Sum256([]byte(password))
	return hex.EncodeToString(sum[:])
}

// passwordMatches compare un mot de passe en clair avec la valeur stockée.
func passwordMatches(storedValue, plainPassword string) bool {
	if storedValue == "" {
		return false
	}
	candidate := hashPassword(plainPassword)
	return storedValue == candidate || storedValue == plainPassword
}

// setSessionCookie crée une session persistée en base puis pose le cookie HTTP.
func setSessionCookie(w http.ResponseWriter, db *sql.DB, userID int) error {
	token, err := storage.CreateSession(db, userID)
	if err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "footix_session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   7 * 24 * 60 * 60,
	})
	return nil
}

// clearSessionCookie supprime la session serveur puis invalide le cookie côté client.
func clearSessionCookie(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if cookie, err := r.Cookie("footix_session"); err == nil {
		storage.DeleteSession(db, cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "footix_session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	})
}

// getAuthenticatedUserID récupère l’utilisateur courant depuis le cookie ou le header de test.
func getAuthenticatedUserID(r *http.Request, db *sql.DB) (int, error) {
	cookie, err := r.Cookie("footix_session")
	if err != nil {
		return 0, errors.New("cookie manquant")
	}
	return storage.ValidateSession(db, cookie.Value)
}

// normalizePredictionResult convertit les saisies possibles vers les valeurs de pronostic internes.
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

// getLeaguesHandler renvoie la liste des compétitions disponibles.
func getLeaguesHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	leagues, err := storage.GetLeagues(db)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "Erreur lors de la récupération des ligues")
		return
	}
	writeJSON(w, r, http.StatusOK, leagues)
}

// getMatchesHandler renvoie soit un match précis, soit une page de matchs filtrés par ligue.
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

// getStatsHandler calcule les probabilités et formes récentes associées à un match.
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

// profileHandler renvoie le profil complet de l’utilisateur connecté.
func profileHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	userID, err := getAuthenticatedUserID(r, db)
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

// myPredictionsHandler renvoie l’historique des pronostics de l’utilisateur connecté.
func myPredictionsHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	userID, err := getAuthenticatedUserID(r, db)
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

// getUsersHandler renvoie les utilisateurs publics nécessaires au chat et au classement.
func getUsersHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	users, err := storage.ListUsers(db)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "Erreur lors de la récupération des utilisateurs")
		return
	}
	writeJSON(w, r, http.StatusOK, users)
}

// getUserByIdHandler renvoie le profil public d’un utilisateur donné.
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

// getScoresHandler renvoie le classement des scores utilisateurs.
func getScoresHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	leaderboard, err := storage.ListScores(db)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "Erreur lors de la récupération du classement")
		return
	}
	writeJSON(w, r, http.StatusOK, leaderboard)
}

// getFeedbacksHandler renvoie les messages d’un match ou d’un utilisateur.
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

// loginHandler authentifie un utilisateur et crée sa session.
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

	setSessionCookie(w, db, user.Id)
	writeJSON(w, r, http.StatusOK, map[string]any{
		"message": "Connexion réussie",
		"user": map[string]any{
			"id":       user.Id,
			"username": user.Username,
			"email":    user.Email,
		},
	})
}

// signupHandler crée un utilisateur puis ouvre directement une session.
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

	setSessionCookie(w, db, user.ID)
	writeJSON(w, r, http.StatusCreated, map[string]any{
		"message": "Compte créé",
		"user":    user,
	})
}

// logoutHandler ferme la session courante et nettoie le cookie associé.
func logoutHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	_ = db
	clearSessionCookie(w, r, db)
	writeJSON(w, r, http.StatusOK, map[string]string{"message": "Déconnexion réussie"})
}

// updateProfileHandler met à jour les informations du profil connecté.
func updateProfileHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	userID, err := getAuthenticatedUserID(r, db)
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

// submitPredictionHandler valide et enregistre le pronostic envoyé par le client.
func submitPredictionHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	userID, err := getAuthenticatedUserID(r, db)
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

// submitFeedbackHandler ajoute un message dans le chat d’un match.
func submitFeedbackHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	userID, err := getAuthenticatedUserID(r, db)
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

// recalculateScoresHandler relance le calcul global des scores utilisateurs.
func recalculateScoresHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if err := storage.RecalculateScores(db); err != nil {
		writeError(w, r, http.StatusInternalServerError, "Erreur lors du recalcul des scores")
		return
	}
	writeJSON(w, r, http.StatusOK, map[string]string{"message": "Scores recalculés"})
}
