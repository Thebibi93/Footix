import { useState, useEffect } from "react";

// Fausses data
const leagues = [
  { id: "ligue1", name: "Ligue 1 Uber Eats", country: "France" },
  { id: "premier", name: "Premier League", country: "Angleterre" },
  { id: "laliga", name: "La Liga", country: "Espagne" },
  { id: "seriea", name: "Serie A", country: "Italie" },
];

const generateMatches = (leagueId) => {
  const commonMatches = [
    {
      id: 1,
      homeTeam: "Paris SG",
      awayTeam: "Olympique Lyonnais",
      location: "Parc des Princes",
      date: "2025-04-20T20:45:00",
      stats: {
        homeWinProb: 58,
        drawProb: 22,
        awayWinProb: 20,
        recentForm: {
          home: ["W", "W", "D", "W", "L"],
          away: ["L", "D", "W", "L", "D"],
        },
        lastMeetings: [{ date: "2024-12-15", result: "PSG 2-1 OL" }],
      },
    },
    {
      id: 2,
      homeTeam: "Olympique de Marseille",
      awayTeam: "AS Monaco",
      location: "Stade Vélodrome",
      date: "2025-04-21T21:00:00",
      stats: {
        homeWinProb: 45,
        drawProb: 30,
        awayWinProb: 25,
        recentForm: {
          home: ["D", "W", "L", "W", "W"],
          away: ["W", "L", "D", "W", "L"],
        },
      },
    },
    {
      id: 3,
      homeTeam: "RC Lens",
      awayTeam: "Stade Rennais",
      location: "Stade Bollaert-Delelis",
      date: "2025-04-22T19:00:00",
      stats: {
        homeWinProb: 52,
        drawProb: 28,
        awayWinProb: 20,
      },
    },
  ];
  if (leagueId === "premier") {
    return [
      {
        id: 101,
        homeTeam: "Manchester City",
        awayTeam: "Arsenal",
        location: "Etihad Stadium",
        date: "2025-04-20T17:30:00",
        stats: { homeWinProb: 55, drawProb: 25, awayWinProb: 20 },
      },
      {
        id: 102,
        homeTeam: "Liverpool",
        awayTeam: "Chelsea",
        location: "Anfield",
        date: "2025-04-21T16:00:00",
        stats: { homeWinProb: 60, drawProb: 22, awayWinProb: 18 },
      },
    ];
  }
  if (leagueId === "laliga") {
    return [
      {
        id: 201,
        homeTeam: "Real Madrid",
        awayTeam: "Barcelona",
        location: "Santiago Bernabéu",
        date: "2025-04-20T21:00:00",
        stats: { homeWinProb: 48, drawProb: 27, awayWinProb: 25 },
      },
    ];
  }
  if (leagueId === "seriea") {
    return [
      {
        id: 301,
        homeTeam: "Juventus",
        awayTeam: "Inter Milan",
        location: "Allianz Stadium",
        date: "2025-04-21T20:45:00",
        stats: { homeWinProb: 40, drawProb: 33, awayWinProb: 27 },
      },
    ];
  }
  return commonMatches;
};

// Fausses predis
let userPredictions = [
  {
    id: 1,
    match: "PSG vs Lyon",
    predictedResult: "1",
    actualResult: "1",
    date: "2025-03-15",
    points: 3,
  },
  {
    id: 2,
    match: "Marseille vs Monaco",
    predictedResult: "N",
    actualResult: "2",
    date: "2025-03-16",
    points: 0,
  },
];

export const useSimulatedApi = () => {
  const [loading, setLoading] = useState(false);

  const getLeagues = async () => {
    setLoading(true);
    await new Promise((r) => setTimeout(r, 500));
    setLoading(false);
    return leagues;
  };

  const getMatchesByLeague = async (leagueId) => {
    setLoading(true);
    await new Promise((r) => setTimeout(r, 500));
    setLoading(false);
    return generateMatches(leagueId);
  };

  const getMatchDetails = async (matchId) => {
    setLoading(true);
    await new Promise((r) => setTimeout(r, 300));
    // C'est pas les vraies mais bon.
    const allMatches = [
      ...generateMatches("ligue1"),
      ...generateMatches("premier"),
      ...generateMatches("laliga"),
      ...generateMatches("seriea"),
    ];
    const match = allMatches.find((m) => m.id === parseInt(matchId));
    setLoading(false);
    return match || null;
  };

  const getUserPredictions = async (userId) => {
    setLoading(true);
    await new Promise((r) => setTimeout(r, 400));
    setLoading(false);
    return [...userPredictions];
  };

  const addPrediction = async (matchId, prediction) => {
    setLoading(true);
    await new Promise((r) => setTimeout(r, 300));
    const newPred = {
      id: Date.now(),
      match: `Match ${matchId}`,
      predictedResult: prediction,
      actualResult: null,
      date: new Date().toISOString().split("T")[0],
      points: null,
    };
    userPredictions = [newPred, ...userPredictions];
    setLoading(false);
    return newPred;
  };

  let chatMessagesByMatch = {
    1: [
      { id: 1, user: "Alice", message: "Qui va gagner ?", timestamp: "10:30" },
      { id: 2, user: "Bob", message: "PSG largement !", timestamp: "10:32" },
    ],
    101: [
      {
        id: 1,
        user: "John Yakuza",
        message: "City va écraser Arsenal",
        timestamp: "11:00",
      },
    ],
  };

  const getChatMessages = async (matchId) => {
    await new Promise((r) => setTimeout(r, 300));
    return chatMessagesByMatch[matchId] || [];
  };

  const sendChatMessage = async (matchId, message, user) => {
    await new Promise((r) => setTimeout(r, 200));
    const newMsg = {
      id: Date.now(),
      user: user.name,
      message,
      timestamp: new Date().toLocaleTimeString(),
    };
    if (!chatMessagesByMatch[matchId]) {
      chatMessagesByMatch[matchId] = [];
    }
    chatMessagesByMatch[matchId] = [newMsg, ...chatMessagesByMatch[matchId]];
    return newMsg;
  };

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
