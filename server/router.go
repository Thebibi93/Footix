/*
Ce fichier va centraliser la définition des points d'entrée (endpoints).
Pour ce projet de prédiction, nous avons besoin d'au moins deux routes majeures :
une pour la liste des matchs et l'autre pour les statistiques de prédiction d'un match précis.
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

const sessionCookieName = "footix_user_id"

// ==================   POINTS IMPORTANTS ==================
// 1. Chaque route est définie via http.HandleFunc, qui associe une URL à une fonction handler.
// 2. Les handlers sont des fonctions qui prennent en paramètre http.ResponseWriter et *http.Request.
// 3. La communication avec le client se fait exclusivement via JSON, en utilisant json.NewEncoder pour encoder les réponses.
// 4. Les handlers font appel à la couche de stockage pour récupérer ou manipuler les données nécessaires.
// 5. Le CORS est géré au niveau de chaque handler pour permettre les requêtes depuis le client React (en dev local).

/*
  HTTP, appelle le package storage, et renvoie du JSON.
  CORS (Cross-Origin Resource Sharing) :
  Le client React (souvent sur le port 5173) est considéré comme une
  origine différente de notre serveur Go (port 8080).
  Sans le header Access-Control-Allow-Origin, le navigateur bloquera la requête.
  Asynchronisme (AJAX) : Le client React utilisera fetch() pour appeler
  ces routes de manière asynchrone.
*/

/*  Petit descriptif

// Authentification et gestion des utilisateurs
// on a un bouton "Login" sur la page d'accueil, qui redirige vers une page de login
// on a un bouton "signup" sur la page d'accueil, qui redirige vers une page de sign up
// on a un bouton "Logout" sur la page d'accueil, qui redirige vers une page de logout
// on a un bouton "Profile" sur la page d'accueil, qui redirige vers une page de profile
// on a un bouton "My Predictions" dans le profil, qui recharge de manière asynchrone la liste des prédictions faites par l'utilisateur
// juste un petit volume lorsque l'utilisateur clique sur "My Predictions", on affiche une liste de ses prédictions passées (match, résultat prédit, résultat réel, date de la prédiction)
// qui surcharge la page de profil, et on peut revenir à la page d'accueil en cliquant sur un bouton "Back to Home" ou quelque chose du genre

// On peut aussi obtenir des infos des utilisateurs par ID
// si on est connecté, on peut aussi faire une route pour obtenir les infos de l'utilisateur connecté (ex: /api/me) et aussi une route pour mettre à jour les infos de l'utilisateur (ex: /api/updateProfile)
// et aussi afficher les infos des autres utilisateurs connectés mais pas toutes leurs infos sensibles on a un sys de chat

// Home page (on aura des boutons pour chaque ligue puis une fois qu'on clique sur une ligue,
// on affiche les matchs à venir (pas encore joué à cette heure actuelle de cette ligue)
// maintenant on est sur un match à venir on voit deux équipe
// on peut cliquer sur un match et on voit les stats de ce match
// (probabilité de victoire de chaque équipe et de nul, et aussi les derniers résultats des deux équipes)
*/

// RegisterRoutes configure tous les points d'entrée de l'API
func RegisterRoutes(db *sql.DB) {

	// ========== ============================ =========================
	//  ===== ======== ==== Côté GET ===== ======= ======== ====
	// ========== ============================ =========================

	// Usage :
	// - GET  /api/profile                -> obtenir les infos du profil de l'utilisateur connecté
	// - POST /api/profile                -> mettre à jour le profil (ex: {"email":"test@mail.com","password":"newpass"})
	// - POST /api/update-profile         -> alias explicite pour la mise à jour du profil
	http.HandleFunc("/api/profile", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			profileHandler(w, r, db)
		case http.MethodPost:
			updateProfileHandler(w, r, db)
		default:
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		}
	})

	// Usage :
	// - GET /api/my-predictions          -> obtenir la liste des prédictions de l'utilisateur connecté
	http.HandleFunc("/api/my-predictions", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
			return
		}
		myPredictionsHandler(w, r, db)
	})

	// Usage :
	// - GET /api/users                   -> obtenir la liste de tous les utilisateurs (pour le sys de chat)
	// - GET /api/users?userId=123        -> obtenir les infos d'un utilisateur spécifique (pour le sys de chat)
	http.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
			return
		}

		if r.URL.Query().Get("userId") != "" {
			getUserByIdHandler(w, r, db)
			return
		}

		getUsersHandler(w, r, db)
	})


	// Route pour récupérer les matchs d'une ligue
	// Usage :
	// - GET /api/matches?league=FL1      -> récupérer les matchs d'une ligue
	http.HandleFunc("/api/matches", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
			return
		}
		getMatchesHandler(w, r, db)
	})

	// Route pour obtenir des statistiques d'un match spécifique
	// Usage :
	// - GET /api/stats?matchId=123       -> obtenir les stats d'un match
	http.HandleFunc("/api/stats", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
			return
		}
		getStatsHandler(w, r, db)
	})

	// Usage :
	// - GET  /api/scores                 -> obtenir le classement des utilisateurs
	// - POST /api/scores                 -> recalculer les scores des utilisateurs
	http.HandleFunc("/api/scores", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getScoresHandler(w, r, db)
		case http.MethodPost:
			recalculateScoresHandler(w, r, db)
		default:
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		}
	})

	// Usage :
	// - GET /api/feedbacks?matchId=123               -> obtenir les messages d'un chat de match donné
	// - GET /api/feedbacks?userId=456                -> obtenir les messages envoyés par un utilisateur donné
	// - GET /api/feedbacks?matchId=123&afterSeq=99   -> obtenir uniquement les nouveaux messages après la séquence 99
	http.HandleFunc("/api/feedbacks", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
			return
		}
		getFeedbacksHandler(w, r, db)
	})

	// ========== ============================ =========================
	//  ===== ======== ==== Côté POST ===== ======= ======== ====
	// ========== ============================ =========================

	// Usage :
	// - POST /api/login                  -> connexion avec body JSON {"username":"idris","password":"secret"}
	http.HandleFunc("/api/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
			return
		}
		loginHandler(w, r, db)
	})

	// Usage :
	// - POST /api/signup                 -> inscription avec body JSON {"username":"idris","email":"idris@mail.com","password":"secret"}
	http.HandleFunc("/api/signup", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
			return
		}
		signupHandler(w, r, db)
	})

	// Usage :
	// - POST /api/logout                 -> déconnexion (pas besoin de body, on se base sur la session/cookie)
	http.HandleFunc("/api/logout", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
			return
		}
		logoutHandler(w, r, db)
	})

	// Usage :
	// - POST /api/update-profile         -> mettre à jour le profil avec body JSON {"email":"new@mail.com","password":"newpass"}
	http.HandleFunc("/api/update-profile", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
			return
		}
		updateProfileHandler(w, r, db)
	})

	// Usage :
	// - POST /api/predict                -> soumettre une prédiction avec body JSON {"matchId":123,"predictedResult":"HOME_WIN"}
	http.HandleFunc("/api/predict", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
			return
		}
		submitPredictionHandler(w, r, db)
	})

	// Usage :
	// - POST /api/feedback               -> envoyer un message dans le chat avec body JSON {"matchId":123,"message":"Très beau match"}
	http.HandleFunc("/api/feedback", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
			return
		}
		submitFeedbackHandler(w, r, db)
	})
}

// ====================== ============Implémentation des CallBack ==========================================================
//======================= ====================== ===========================================================================
//======================== ====================== ==========================================================================
//======================== ====================== ==========================================================================
//======================== ====================== ==========================================================================
//======================== ====================== ==========================================================================

// =========================
// Helpers HTTP / JSON / CORS
// =========================

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

// Package sha256 implements the SHA224 and SHA256 hash algorithms as defined in FIPS 180-4.
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

	// petit fallback pratique pour tester rapidement côté dev / Postman
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
	case "X", "DRAW", "NUL":
		return "DRAW", nil
	case "2", "AWAY", "AWAY_WIN", "AWAYWIN", "EXTERIEUR", "EXTÉRIEUR":
		return "AWAY_WIN", nil
	default:
		return "", errors.New("predictedResult doit valoir HOME_WIN, DRAW ou AWAY_WIN")
	}
}

// =========================
// GET routes
// =========================

func getMatchesHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if handlePreflight(w, r) {
		return
	}
	if r.Method != http.MethodGet {
		writeError(w, r, http.StatusMethodNotAllowed, "Méthode non autorisée")
		return
	}

	leagueCode := strings.TrimSpace(r.URL.Query().Get("league"))
	if leagueCode == "" {
		writeError(w, r, http.StatusBadRequest, "Le paramètre 'league' est requis")
		return
	}

	matches, err := storage.GetMatchesByLeague(db, leagueCode)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "Erreur lors de la récupération des matchs")
		return
	}

	writeJSON(w, r, http.StatusOK, matches)
}

func getStatsHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if handlePreflight(w, r) {
		return
	}
	if r.Method != http.MethodGet {
		writeError(w, r, http.StatusMethodNotAllowed, "Méthode non autorisée")
		return
	}

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
	if handlePreflight(w, r) {
		return
	}

	switch r.Method {
	case http.MethodGet:
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

	case http.MethodPost:
		updateProfileHandler(w, r, db)

	default:
		writeError(w, r, http.StatusMethodNotAllowed, "Méthode non autorisée")
	}
}

func myPredictionsHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if handlePreflight(w, r) {
		return
	}
	if r.Method != http.MethodGet {
		writeError(w, r, http.StatusMethodNotAllowed, "Méthode non autorisée")
		return
	}

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

// =========================
// POST routes
// =========================

func loginHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if handlePreflight(w, r) {
		return
	}
	if r.Method != http.MethodPost {
		writeError(w, r, http.StatusMethodNotAllowed, "Méthode non autorisée")
		return
	}

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
	if handlePreflight(w, r) {
		return
	}
	if r.Method != http.MethodPost {
		writeError(w, r, http.StatusMethodNotAllowed, "Méthode non autorisée")
		return
	}

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
	if handlePreflight(w, r) {
		return
	}
	if r.Method != http.MethodPost {
		writeError(w, r, http.StatusMethodNotAllowed, "Méthode non autorisée")
		return
	}

	clearSessionCookie(w)
	writeJSON(w, r, http.StatusOK, map[string]string{"message": "Déconnexion réussie"})
}

func updateProfileHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if handlePreflight(w, r) {
		return
	}
	if r.Method != http.MethodPost {
		writeError(w, r, http.StatusMethodNotAllowed, "Méthode non autorisée")
		return
	}

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
	if handlePreflight(w, r) {
		return
	}
	if r.Method != http.MethodPost {
		writeError(w, r, http.StatusMethodNotAllowed, "Méthode non autorisée")
		return
	}

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
		if strings.Contains(err.Error(), "déjà terminé") {
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
	if handlePreflight(w, r) {
		return
	}
	if r.Method != http.MethodPost {
		writeError(w, r, http.StatusMethodNotAllowed, "Méthode non autorisée")
		return
	}

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
		writeError(w, r, http.StatusConflict, "Impossible d'enregistrer le message pour le moment, réessaie")
		return
	}

	writeJSON(w, r, http.StatusCreated, message)
}

func recalculateScoresHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if handlePreflight(w, r) {
		return
	}
	if r.Method != http.MethodPost {
		writeError(w, r, http.StatusMethodNotAllowed, "Méthode non autorisée")
		return
	}

	if err := storage.RecalculateScores(db); err != nil {
		writeError(w, r, http.StatusInternalServerError, "Erreur lors du recalcul des scores")
		return
	}

	writeJSON(w, r, http.StatusOK, map[string]string{"message": "Scores recalculés"})
}
