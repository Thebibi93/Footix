import { useEffect, useState } from "react";
import { useSimulatedApi } from "../hooks/useSimulatedApi";
import { useAuth } from "../contexts/AuthContext";

export default function PredictionHistory() {
  const [predictions, setPredictions] = useState([]);
  const { getUserPredictions, loading } = useSimulatedApi();
  const { user } = useAuth();

  useEffect(() => {
    if (user) {
      getUserPredictions(user.id).then(setPredictions);
    }
  }, [user]);

  if (loading) return <div className="mt-4 text-center">Chargement...</div>;

  return (
    <div className="mt-6">
      <h3 className="text-xl font-semibold mb-3">Historique des pronostics</h3>
      {predictions.length === 0 ? (
        <p className="text-gray-500">Aucun pronostic pour le moment.</p>
      ) : (
        <div className="space-y-2">
          {predictions.map((p) => (
            <div
              key={p.id}
              className="border p-3 rounded flex justify-between items-center"
            >
              <div>
                <div className="font-medium">{p.match}</div>
                <div className="text-sm text-gray-500">
                  Pronostic: {p.predictedResult} | Date: {p.date}
                </div>
              </div>
              <div
                className={`font-bold ${p.actualResult ? (p.predictedResult === p.actualResult ? "text-green-600" : "text-red-600") : "text-gray-400"}`}
              >
                {p.actualResult
                  ? p.predictedResult === p.actualResult
                    ? `+${p.points} pts`
                    : "0 pt"
                  : "À venir"}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
