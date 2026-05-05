import { useCallback, useState } from "react";

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? "http://localhost:8080";
let leaguesCache = null;
const matchesCache = new Map();
const matchDetailsCache = new Map();
const profileCache = new Map();

function splitRecentForm(value) {
  if (!value) {
    return [];
  }
  return String(value)
    .split(/\s+/)
    .map((item) => item.trim())
    .filter(Boolean);
}

function normalizePredictionValue(value) {
  switch (String(value || "").trim().toUpperCase()) {
    case "1":
    case "HOME":
    case "HOME_WIN":
    case "HOMEWIN":
      return "HOME_WIN";
    case "X":
    case "N":
    case "DRAW":
    case "NUL":
      return "DRAW";
    case "2":
    case "AWAY":
    case "AWAY_WIN":
    case "AWAYWIN":
      return "AWAY_WIN";
    default:
      return value;
  }
}

function formatPredictionLabel(value) {
  switch (normalizePredictionValue(value)) {
    case "HOME_WIN":
      return "Victoire domicile";
    case "DRAW":
      return "Match nul";
    case "AWAY_WIN":
      return "Victoire extérieur";
    default:
      return value || "";
  }
}

function formatDateTime(value, options = {}) {
  if (!value) {
    return "";
  }
  return new Date(value).toLocaleString("fr-FR", options);
}

function formatDate(value) {
  return formatDateTime(value, {
    day: "2-digit",
    month: "short",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

function isPastMatch(rawMatch) {
  const status = String(rawMatch?.status || "").toUpperCase();
  if (status === "FINISHED") {
    return true;
  }
  if (!rawMatch?.utcDate) {
    return false;
  }
  return new Date(rawMatch.utcDate).getTime() < Date.now() - 3 * 60 * 60 * 1000;
}

function isStartedMatch(rawMatch) {
  if (!rawMatch?.utcDate) {
    return false;
  }
  return new Date(rawMatch.utcDate).getTime() <= Date.now();
}

function mapLeague(rawLeague) {
  return {
    id: rawLeague.id,
    name: rawLeague.name,
    code: rawLeague.code,
  };
}

function mapMatch(rawMatch, fallbackLeague) {
  const finished = isPastMatch(rawMatch);
  const started = isStartedMatch(rawMatch);
  const homeScore = rawMatch.score?.fullTime?.home ?? null;
  const awayScore = rawMatch.score?.fullTime?.away ?? null;

  return {
    id: rawMatch.id,
    leagueId: rawMatch.leagueId ?? fallbackLeague?.id,
    leagueCode: rawMatch.leagueCode ?? fallbackLeague?.code,
    leagueName: rawMatch.leagueName ?? fallbackLeague?.name ?? "Compétition",
    season: rawMatch.season,
    homeTeam: rawMatch.homeTeam?.name || "Équipe domicile",
    awayTeam: rawMatch.awayTeam?.name || "Équipe extérieur",
    homeTeamId: rawMatch.homeTeam?.id,
    awayTeamId: rawMatch.awayTeam?.id,
    homeTeamShortName: rawMatch.homeTeam?.shortName || rawMatch.homeTeam?.name,
    awayTeamShortName: rawMatch.awayTeam?.shortName || rawMatch.awayTeam?.name,
    homeTeamCrest: rawMatch.homeTeam?.crest || null,
    awayTeamCrest: rawMatch.awayTeam?.crest || null,
    date: rawMatch.utcDate,
    location: rawMatch.leagueName ?? fallbackLeague?.name ?? "Compétition",
    status: rawMatch.status,
    homeScore,
    awayScore,
    isFinished: finished,
    isStarted: started,
    canPredict: !finished && !started,
    canChat: !finished,
    scoreboard: homeScore !== null && awayScore !== null ? `${homeScore} - ${awayScore}` : "-",
  };
}

function mapChatMessage(rawMessage) {
  return {
    id: rawMessage.id,
    roomId: rawMessage.roomId,
    matchId: rawMessage.matchId,
    seqInChat: rawMessage.seqInChat,
    userId: rawMessage.userId,
    username: rawMessage.username || `User #${rawMessage.userId}`,
    message: rawMessage.message,
    feedbackDate: rawMessage.feedbackDate,
    timestamp: formatDate(rawMessage.feedbackDate),
  };
}

function mapProfile(rawProfile) {
  if (!rawProfile) {
    return null;
  }

  return {
    id: rawProfile.id,
    username: rawProfile.username,
    email: rawProfile.email,
    score: rawProfile.score ?? 0,
    rank: rawProfile.rank ?? 0,
    totalPredictions: rawProfile.totalPredictions ?? 0,
    correctPredictions: rawProfile.correctPredictions ?? 0,
    accuracy: Number(rawProfile.accuracy ?? 0),
    chatMessages: rawProfile.chatMessages ?? 0,
  };
}

async function apiFetch(path, options = {}) {
  const headers = new Headers(options.headers || {});
  const isJsonBody =
    options.body !== undefined && options.body !== null && !(options.body instanceof FormData);

  if (isJsonBody && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }

  const response = await fetch(`${API_BASE_URL}${path}`, {
    credentials: "include",
    ...options,
    headers,
    body: isJsonBody && typeof options.body !== "string"
      ? JSON.stringify(options.body)
      : options.body,
  });

  const text = await response.text();
  let data = null;
  try {
    data = text ? JSON.parse(text) : null;
  } catch {
    data = null;
  }

  if (!response.ok) {
    throw new Error(data?.error || data?.message || "Erreur serveur");
  }

  return data;
}

async function ensureLeagues() {
  if (leaguesCache) {
    return leaguesCache;
  }
  const data = await apiFetch("/api/leagues", { method: "GET" });
  leaguesCache = Array.isArray(data) ? data.map(mapLeague) : [];
  return leaguesCache;
}

function matchesCacheKey(leagueCode, bucket, page, pageSize) {
  return `${leagueCode}:${bucket}:${page}:${pageSize}`;
}

export const useApi = () => {
  const [loading, setLoading] = useState(false);

  const withLoading = useCallback(async (fn) => {
    setLoading(true);
    try {
      return await fn();
    } finally {
      setLoading(false);
    }
  }, []);

  const getLeagues = useCallback(async () => withLoading(ensureLeagues), [withLoading]);

  const getMatchesByLeague = useCallback(async (leagueId, options = {}) => {
    return withLoading(async () => {
      const bucket = options.bucket || "upcoming";
      const page = options.page || 1;
      const pageSize = options.pageSize || 12;
      const leagues = await ensureLeagues();
      const league = leagues.find((item) => String(item.id) === String(leagueId));
      if (!league) {
        return { items: [], page: 1, pageSize, total: 0, totalPages: 0, bucket };
      }

      const key = matchesCacheKey(league.code, bucket, page, pageSize);
      if (matchesCache.has(key)) {
        return matchesCache.get(key);
      }

      const data = await apiFetch(
        `/api/matches?league=${encodeURIComponent(league.code)}&bucket=${encodeURIComponent(bucket)}&page=${encodeURIComponent(page)}&pageSize=${encodeURIComponent(pageSize)}`,
        { method: "GET" },
      );

      const response = {
        items: Array.isArray(data?.items) ? data.items.map((item) => mapMatch(item, league)) : [],
        page: data?.page ?? page,
        pageSize: data?.pageSize ?? pageSize,
        total: data?.total ?? 0,
        totalPages: data?.totalPages ?? 0,
        bucket: data?.bucket ?? bucket,
        league,
      };

      matchesCache.set(key, response);
      for (const item of response.items) {
        matchDetailsCache.set(String(item.id), item);
      }
      return response;
    });
  }, [withLoading]);

  const getMatchDetails = useCallback(async (matchId) => {
    return withLoading(async () => {
      const rawMatch = await apiFetch(`/api/matches?matchId=${encodeURIComponent(matchId)}`, {
        method: "GET",
      });
      if (!rawMatch) {
        return null;
      }

      const leagues = await ensureLeagues();
      const fallbackLeague = leagues.find((league) => league.code === rawMatch.leagueCode || league.id === rawMatch.leagueId);
      const baseMatch = mapMatch(rawMatch, fallbackLeague);

      let stats = null;
      try {
        stats = await apiFetch(`/api/stats?matchId=${encodeURIComponent(matchId)}`, {
          method: "GET",
        });
      } catch {
        stats = null;
      }

      const detailedMatch = {
        ...baseMatch,
        stats: {
          homeWinProb: Math.round(stats?.homeWinProb ?? 33),
          drawProb: Math.round(stats?.drawProb ?? 34),
          awayWinProb: Math.round(stats?.awayWinProb ?? 33),
          recentForm: {
            home: splitRecentForm(stats?.homeLastResults),
            away: splitRecentForm(stats?.awayLastResults),
          },
        },
      };
      matchDetailsCache.set(String(matchId), detailedMatch);
      return detailedMatch;
    });
  }, [withLoading]);

  const getProfileSummary = useCallback(async () => {
    return withLoading(async () => {
      const data = await apiFetch("/api/profile", { method: "GET" });
      const mapped = mapProfile(data);
      if (mapped) {
        profileCache.set(String(mapped.id), mapped);
      }
      return mapped;
    });
  }, [withLoading]);

  const getUserProfile = useCallback(async (userId) => {
    return withLoading(async () => {
      const cached = profileCache.get(String(userId));
      if (cached) {
        return cached;
      }
      const data = await apiFetch(`/api/users?userId=${encodeURIComponent(userId)}`, { method: "GET" });
      const mapped = mapProfile(data);
      if (mapped) {
        profileCache.set(String(userId), mapped);
      }
      return mapped;
    });
  }, [withLoading]);

  const getLeaderboard = useCallback(async () => {
    return withLoading(async () => {
      const data = await apiFetch("/api/scores", { method: "GET" });
      return Array.isArray(data) ? data : [];
    });
  }, [withLoading]);

  const getUserPredictions = useCallback(async () => {
    return withLoading(async () => {
      const data = await apiFetch("/api/my-predictions", { method: "GET" });
      return Array.isArray(data)
        ? data.map((item, index) => {
            const normalizedPredictedResult = normalizePredictionValue(item.predictedResult);
            const normalizedActualResult = normalizePredictionValue(item.actualResult);

            return {
              id: `${item.matchId}-${item.predictionDate}-${index}`,
              matchId: item.matchId,
              match: `${item.homeTeam} vs ${item.awayTeam}`,
              predictedResult: formatPredictionLabel(item.predictedResult),
              rawPredictedResult: item.predictedResult,
              normalizedPredictedResult,
              actualResult: formatPredictionLabel(item.actualResult),
              rawActualResult: item.actualResult,
              normalizedActualResult,
              date: formatDate(item.predictionDate),
              matchDate: formatDate(item.matchDate),
              points:
                item.actualResult &&
                normalizedPredictedResult === normalizedActualResult
                  ? 1
                  : 0,
            };
          })
        : [];
    });
  }, [withLoading]);

  const addPrediction = useCallback(async (matchId, predictedResult) => {
    return withLoading(async () => {
      const normalized = normalizePredictionValue(predictedResult);
      return apiFetch("/api/predict", {
        method: "POST",
        body: {
          matchId: Number(matchId),
          predictedResult: normalized,
        },
      });
    });
  }, [withLoading]);

  const getChatMessages = useCallback(async (matchId, afterSeq = 0, limit = 100) => {
    return withLoading(async () => {
      const data = await apiFetch(
        `/api/feedbacks?matchId=${encodeURIComponent(matchId)}&afterSeq=${encodeURIComponent(afterSeq)}&limit=${encodeURIComponent(limit)}`,
        { method: "GET" },
      );

      const messages = Array.isArray(data) ? data.map(mapChatMessage) : [];
      messages.sort((a, b) => a.seqInChat - b.seqInChat);
      return messages;
    });
  }, [withLoading]);

  const sendChatMessage = useCallback(async (matchId, message) => {
    return withLoading(async () => {
      const created = await apiFetch("/api/feedback", {
        method: "POST",
        body: {
          matchId: Number(matchId),
          message,
        },
      });
      return mapChatMessage(created);
    });
  }, [withLoading]);

  return {
    loading,
    getLeagues,
    getMatchesByLeague,
    getMatchDetails,
    getProfileSummary,
    getUserProfile,
    getLeaderboard,
    getUserPredictions,
    addPrediction,
    getChatMessages,
    sendChatMessage,
  };
};
