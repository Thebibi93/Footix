import { useEffect, useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { useSimulatedApi } from "../hooks/useSimulatedApi";
import { useAuth } from "../contexts/AuthContext";
import MatchChat from "./MatchChat";

export default function MatchDetail() {
  const { matchId } = useParams();
  const [match, setMatch] = useState(null);
  const [prediction, setPrediction] = useState("");
  const [message, setMessage] = useState("");
  const { getMatchDetails, addPrediction, loading } = useSimulatedApi();
  const { user } = useAuth();
  const navigate = useNavigate();

  useEffect(() => {
    getMatchDetails(matchId).then(setMatch);
  }, [matchId]);

  const handlePredict = async () => {
    if (!user) {
      navigate("/login");
      return;
    }
    if (!prediction) return;
    await addPrediction(matchId, prediction);
    setMessage("Pronostic enregistré !");
    setTimeout(() => setMessage(""), 3000);
  };

  if (loading || !match)
    return <div className="text-center py-10">Chargement...</div>;

  const stats = match.stats || {
    homeWinProb: 33,
    drawProb: 34,
    awayWinProb: 33,
    recentForm: { home: [], away: [] },
  };

  return (
    <div className="max-w-3xl mx-auto">
      <div className="bg-white rounded-xl shadow-lg p-6">
        <h1 className="text-3xl font-bold text-center mb-2">
          {match.homeTeam} vs {match.awayTeam}
        </h1>
        <p className="text-center text-gray-500 mb-6">
          {new Date(match.date).toLocaleString("fr-FR")} - {match.location}
        </p>

        <div className="border-t pt-4 mb-6">
          <h2 className="text-xl font-semibold mb-3">Probabilités</h2>
          <div className="flex gap-4 text-center">
            <div className="flex-1 bg-blue-100 p-3 rounded">
              <span className="block font-bold text-2xl">
                {stats.homeWinProb}%
              </span>{" "}
              Victoire {match.homeTeam}
            </div>
            <div className="flex-1 bg-gray-100 p-3 rounded">
              <span className="block font-bold text-2xl">
                {stats.drawProb}%
              </span>{" "}
              Nul
            </div>
            <div className="flex-1 bg-blue-100 p-3 rounded">
              <span className="block font-bold text-2xl">
                {stats.awayWinProb}%
              </span>{" "}
              Victoire {match.awayTeam}
            </div>
          </div>
        </div>

        {stats.recentForm.home.length > 0 && (
          <div className="mb-6">
            <h3 className="font-semibold">Forme récente</h3>
            <div className="flex gap-2 mt-1">
              <span className="font-medium">{match.homeTeam}:</span>{" "}
              {stats.recentForm.home.map((res, i) => (
                <span
                  key={i}
                  className={`px-2 py-1 rounded ${res === "W" ? "bg-green-500" : res === "D" ? "bg-yellow-500" : "bg-red-500"} text-white text-sm`}
                >
                  {res}
                </span>
              ))}
            </div>
            <div className="flex gap-2 mt-1">
              <span className="font-medium">{match.awayTeam}:</span>{" "}
              {stats.recentForm.away.map((res, i) => (
                <span
                  key={i}
                  className={`px-2 py-1 rounded ${res === "W" ? "bg-green-500" : res === "D" ? "bg-yellow-500" : "bg-red-500"} text-white text-sm`}
                >
                  {res}
                </span>
              ))}
            </div>
          </div>
        )}

        {user && (
          <div className="mt-6 p-4 bg-gray-50 rounded-lg">
            <h3 className="font-bold mb-2">Faire un pronostic</h3>
            <div className="flex gap-3">
              <button
                onClick={() => setPrediction("1")}
                className={`flex-1 py-2 rounded ${prediction === "1" ? "bg-blue-600 text-white" : "bg-gray-200"}`}
              >
                1 (domicile)
              </button>
              <button
                onClick={() => setPrediction("N")}
                className={`flex-1 py-2 rounded ${prediction === "N" ? "bg-blue-600 text-white" : "bg-gray-200"}`}
              >
                N (nul)
              </button>
              <button
                onClick={() => setPrediction("2")}
                className={`flex-1 py-2 rounded ${prediction === "2" ? "bg-blue-600 text-white" : "bg-gray-200"}`}
              >
                2 (extérieur)
              </button>
            </div>
            <button
              onClick={handlePredict}
              className="mt-3 w-full bg-green-600 text-white py-2 rounded hover:bg-green-700"
            >
              Enregistrer mon pronostic
            </button>
            {message && (
              <p className="text-green-600 text-center mt-2">{message}</p>
            )}

            <MatchChat
              matchId={matchId}
              matchTitle={`${match.homeTeam} - ${match.awayTeam}`}
            />
          </div>
        )}
      </div>
    </div>
  );
}
