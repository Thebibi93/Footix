import { useState, useCallback } from "react";

// ----- Données mock correspondant au schéma BDD pour tester l'application -----

// 1. Ligues (Leagues)
const leagues = [
  { id: 1, name: "Ligue 1 Uber Eats", code: "L1" },
  { id: 2, name: "Premier League", code: "PL" },
  { id: 3, name: "La Liga", code: "LALIGA" },
  { id: 4, name: "Serie A", code: "SA" },
];

// 2. Équipes (Teams)
const teams = {
  1: { id: 1, name: "Paris SG", short_name: "PSG", crest_url: null },
  2: { id: 2, name: "Olympique Lyonnais", short_name: "OL", crest_url: null },
  3: {
    id: 3,
    name: "Olympique de Marseille",
    short_name: "OM",
    crest_url: null,
  },
  4: { id: 4, name: "AS Monaco", short_name: "ASM", crest_url: null },
  5: { id: 5, name: "RC Lens", short_name: "RCL", crest_url: null },
  6: { id: 6, name: "Stade Rennais", short_name: "SRFC", crest_url: null },
  7: { id: 7, name: "Manchester City", short_name: "MCI", crest_url: null },
  8: { id: 8, name: "Arsenal", short_name: "ARS", crest_url: null },
  9: { id: 9, name: "Liverpool", short_name: "LIV", crest_url: null },
  10: { id: 10, name: "Chelsea", short_name: "CHE", crest_url: null },
  11: { id: 11, name: "Real Madrid", short_name: "RMA", crest_url: null },
  12: { id: 12, name: "Barcelona", short_name: "FCB", crest_url: null },
  13: { id: 13, name: "Juventus", short_name: "JUV", crest_url: null },
  14: { id: 14, name: "Inter Milan", short_name: "INT", crest_url: null },
};

// 3. Matchs (Matches)
let matches = [
  {
    id: 1,
    league_id: 1,
    season: 2025,
    utc_date: "2025-04-20T20:45:00Z",
    home_team_id: 1,
    away_team_id: 2,
    home_score: null,
    away_score: null,
    status: "SCHEDULED",
  },
  {
    id: 2,
    league_id: 1,
    season: 2025,
    utc_date: "2025-04-21T21:00:00Z",
    home_team_id: 3,
    away_team_id: 4,
    home_score: null,
    away_score: null,
    status: "SCHEDULED",
  },
  {
    id: 3,
    league_id: 1,
    season: 2025,
    utc_date: "2025-04-22T19:00:00Z",
    home_team_id: 5,
    away_team_id: 6,
    home_score: null,
    away_score: null,
    status: "SCHEDULED",
  },
  {
    id: 101,
    league_id: 2,
    season: 2025,
    utc_date: "2025-04-20T17:30:00Z",
    home_team_id: 7,
    away_team_id: 8,
    home_score: null,
    away_score: null,
    status: "SCHEDULED",
  },
  {
    id: 102,
    league_id: 2,
    season: 2025,
    utc_date: "2025-04-21T16:00:00Z",
    home_team_id: 9,
    away_team_id: 10,
    home_score: null,
    away_score: null,
    status: "SCHEDULED",
  },
  {
    id: 201,
    league_id: 3,
    season: 2025,
    utc_date: "2025-04-20T21:00:00Z",
    home_team_id: 11,
    away_team_id: 12,
    home_score: null,
    away_score: null,
    status: "SCHEDULED",
  },
  {
    id: 301,
    league_id: 4,
    season: 2025,
    utc_date: "2025-04-21T20:45:00Z",
    home_team_id: 13,
    away_team_id: 14,
    home_score: null,
    away_score: null,
    status: "SCHEDULED",
  },
];

// Pronostics liés aux matchs
const predictions = {
  1: {
    match_id: 1,
    win_probability_home: 0.58,
    win_probability_away: 0.2,
    predicted_result: "1",
  },
  2: {
    match_id: 2,
    win_probability_home: 0.45,
    win_probability_away: 0.25,
    predicted_result: "1",
  },
  3: {
    match_id: 3,
    win_probability_home: 0.52,
    win_probability_away: 0.2,
    predicted_result: "1",
  },
  101: {
    match_id: 101,
    win_probability_home: 0.55,
    win_probability_away: 0.2,
    predicted_result: "1",
  },
  102: {
    match_id: 102,
    win_probability_home: 0.6,
    win_probability_away: 0.18,
    predicted_result: "1",
  },
  201: {
    match_id: 201,
    win_probability_home: 0.48,
    win_probability_away: 0.25,
    predicted_result: "N",
  },
  301: {
    match_id: 301,
    win_probability_home: 0.4,
    win_probability_away: 0.27,
    predicted_result: "N",
  },
};

// Utilisateurs simulés
let users = [
  {
    id: 1,
    username: "Alice",
    email: "alice@example.com",
    password_hash: "fake",
  },
  { id: 2, username: "Bob", email: "bob@example.com", password_hash: "fake" },
];

// Pronostics utilisateurs (UserPredictions) et historique (UserPredictionHistory)
let userPredictions = [
  { user_id: 1, match_id: 1, predicted_result: "1" },
  { user_id: 1, match_id: 2, predicted_result: "N" },
  { user_id: 2, match_id: 101, predicted_result: "2" },
];

let userPredictionHistory = [
  {
    user_id: 1,
    match_id: 1,
    predicted_result: "1",
    actual_result: null,
    prediction_date: "2025-03-15T10:00:00Z",
  },
  {
    user_id: 1,
    match_id: 2,
    predicted_result: "N",
    actual_result: null,
    prediction_date: "2025-03-16T14:30:00Z",
  },
  {
    user_id: 2,
    match_id: 101,
    predicted_result: "2",
    actual_result: null,
    prediction_date: "2025-03-17T09:15:00Z",
  },
];

// ChatRooms une room "main" par match
let chatRooms = {};
matches.forEach((m) => {
  chatRooms[m.id] = {
    id: m.id,
    match_id: m.id,
    room_type: "main",
    created_at: new Date().toISOString(),
  };
});

// ChatRoomCounters compteur de séquence par room
let chatRoomCounters = {};
Object.keys(chatRooms).forEach((roomId) => {
  chatRoomCounters[roomId] = { chat_room_id: parseInt(roomId), last_seq: 0 };
});

// ChatMessages stockés par room
let chatMessages = {
  1: [
    {
      id: 1,
      chat_room_id: 1,
      seq_in_room: 1,
      user_id: 1,
      message: "Qui va gagner le Classique ?",
      created_at: "2025-04-04T10:30:00Z",
    },
    {
      id: 2,
      chat_room_id: 1,
      seq_in_room: 2,
      user_id: 2,
      message: "PSG largement !",
      created_at: "2025-04-04T10:32:00Z",
    },
  ],
  101: [
    {
      id: 3,
      chat_room_id: 101,
      seq_in_room: 1,
      user_id: 1,
      message: "City va écraser Arsenal",
      created_at: "2025-04-04T11:00:00Z",
    },
  ],
};

// Initialisation des compteurs à partir des messages existants
for (const roomId in chatMessages) {
  const msgs = chatMessages[roomId];
  if (msgs.length) {
    const maxSeq = Math.max(...msgs.map((m) => m.seq_in_room));
    chatRoomCounters[roomId] = {
      chat_room_id: parseInt(roomId),
      last_seq: maxSeq,
    };
  }
}

// Obtenir le nom d'une équipe depuis son ID
function getTeamName(teamId) {
  return teams[teamId]?.name || "Équipe inconnue";
}

// Formater un match pour l'affichage (jointure avec Teams)
function enrichMatch(match) {
  const pred = predictions[match.id] || {
    win_probability_home: 33,
    win_probability_away: 33,
  };
  return {
    id: match.id,
    homeTeam: getTeamName(match.home_team_id),
    awayTeam: getTeamName(match.away_team_id),
    location: "Stade", // mock pour le moment
    date: match.utc_date,
    status: match.status,
    home_score: match.home_score,
    away_score: match.away_score,
    stats: {
      homeWinProb: Math.round(pred.win_probability_home * 100),
      drawProb:
        100 -
        Math.round(pred.win_probability_home * 100) -
        Math.round(pred.win_probability_away * 100),
      awayWinProb: Math.round(pred.win_probability_away * 100),
      recentForm: {
        home: ["W", "W", "D", "W", "L"], // mock
        away: ["L", "D", "W", "L", "D"],
      },
    },
  };
}

export const useSimulatedApi = () => {
  const [loading, setLoading] = useState(false);

  // Récupérer toutes les ligues
  const getLeagues = useCallback(async () => {
    setLoading(true);
    await new Promise((r) => setTimeout(r, 500));
    setLoading(false);
    return leagues;
  }, []);

  // Récupérer les matchs d'une ligue donnée
  const getMatchesByLeague = useCallback(async (leagueId) => {
    setLoading(true);
    await new Promise((r) => setTimeout(r, 500));
    const filtered = matches.filter((m) => m.league_id === parseInt(leagueId));
    const enriched = filtered.map(enrichMatch);
    setLoading(false);
    return enriched;
  }, []);

  // Récupérer les détails d'un match
  const getMatchDetails = useCallback(async (matchId) => {
    setLoading(true);
    await new Promise((r) => setTimeout(r, 300));
    const match = matches.find((m) => m.id === parseInt(matchId));
    if (!match) {
      setLoading(false);
      return null;
    }
    const enriched = enrichMatch(match);
    setLoading(false);
    return enriched;
  }, []);

  // Récupérer l'historique des pronostics d'un utilisateur
  const getUserPredictions = useCallback(async (userId) => {
    setLoading(true);
    await new Promise((r) => setTimeout(r, 400));
    const userPreds = userPredictionHistory.filter(
      (ph) => ph.user_id === parseInt(userId),
    );
    const enriched = userPreds.map((ph) => {
      const match = matches.find((m) => m.id === ph.match_id);
      const matchName = match
        ? `${getTeamName(match.home_team_id)} - ${getTeamName(match.away_team_id)}`
        : `Match ${ph.match_id}`;
      let points = null;
      if (ph.actual_result) {
        points = ph.predicted_result === ph.actual_result ? 3 : 0;
      }
      return {
        id: `${ph.user_id}_${ph.match_id}`,
        match: matchName,
        predictedResult: ph.predicted_result,
        actualResult: ph.actual_result,
        date: new Date(ph.prediction_date).toISOString().split("T")[0],
        points: points,
      };
    });
    setLoading(false);
    return enriched;
  }, []);

  // Ajouter un pronostic utilisateur (dans UserPredictions et UserPredictionHistory)
  const addPrediction = useCallback(async (matchId, prediction) => {
    setLoading(true);
    await new Promise((r) => setTimeout(r, 300));
    const userId = 1; // À remplacer par l'ID de l'utilisateur connecté
    // Vérifier si le pronostic existe déjà
    const existingIndex = userPredictions.findIndex(
      (up) => up.user_id === userId && up.match_id === parseInt(matchId),
    );
    if (existingIndex !== -1) {
      userPredictions[existingIndex].predicted_result = prediction;
    } else {
      userPredictions.push({
        user_id: userId,
        match_id: parseInt(matchId),
        predicted_result: prediction,
      });
    }
    // Ajouter dans l'historique
    const historyEntry = {
      user_id: userId,
      match_id: parseInt(matchId),
      predicted_result: prediction,
      actual_result: null,
      prediction_date: new Date().toISOString(),
    };
    const historyIndex = userPredictionHistory.findIndex(
      (h) => h.user_id === userId && h.match_id === parseInt(matchId),
    );
    if (historyIndex !== -1) {
      userPredictionHistory[historyIndex] = historyEntry;
    } else {
      userPredictionHistory.push(historyEntry);
    }
    setLoading(false);
    return { success: true };
  }, []);

  // Chat par match (basé sur ChatRooms, ChatRoomCounters, ChatMessages)
  const getChatMessages = useCallback(async (matchId) => {
    await new Promise((r) => setTimeout(r, 300));
    const roomId = matchId; // on utilise matchId comme chat_room_id (id de la room = id du match)
    const msgs = chatMessages[roomId] || [];
    // Enrichir avec le nom d'utilisateur
    const enriched = msgs.map((msg) => ({
      id: msg.id,
      user: users.find((u) => u.id === msg.user_id)?.username || "Inconnu",
      message: msg.message,
      timestamp: new Date(msg.created_at).toLocaleTimeString(),
    }));
    // Trier par séquence croissante (ancien en premier) ou décroissant selon affichage
    return enriched.reverse();
  }, []);

  const sendChatMessage = useCallback(async (matchId, message, user) => {
    await new Promise((r) => setTimeout(r, 200));
    const roomId = matchId;
    // Récupérer l'utilisateur réel (pour obtenir son id)
    const dbUser = users.find((u) => u.username === user.name) || {
      id: 1,
      username: user.name,
    };
    // Incrémenter le compteur de la room
    if (!chatRoomCounters[roomId]) {
      chatRoomCounters[roomId] = {
        chat_room_id: parseInt(roomId),
        last_seq: 0,
      };
    }
    chatRoomCounters[roomId].last_seq += 1;
    const newSeq = chatRoomCounters[roomId].last_seq;
    // Créer le message
    const newMsg = {
      id: Date.now(),
      chat_room_id: parseInt(roomId),
      seq_in_room: newSeq,
      user_id: dbUser.id,
      message: message,
      created_at: new Date().toISOString(),
    };
    if (!chatMessages[roomId]) chatMessages[roomId] = [];
    chatMessages[roomId].push(newMsg);
    // Retourner le message formaté pour le front
    return {
      id: newMsg.id,
      user: dbUser.username,
      message: message,
      timestamp: new Date().toLocaleTimeString(),
    };
  }, []);

  return {
    loading,
    getLeagues,
    getMatchesByLeague,
    getMatchDetails,
    getUserPredictions,
    addPrediction,
    getChatMessages,
    sendChatMessage,
  };
};
