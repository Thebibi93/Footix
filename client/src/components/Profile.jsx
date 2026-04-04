import { useState } from "react";
import { useAuth } from "../contexts/AuthContext";
import PredictionHistory from "./PredictionHistory";

export default function Profile() {
  const { user } = useAuth();
  const [showPredictions, setShowPredictions] = useState(false);

  if (!user)
    return <div className="text-center py-10">Veuillez vous connecter</div>;

  return (
    <div className="max-w-2xl mx-auto">
      <div className="bg-white rounded-xl shadow-md p-6">
        <h1 className="text-2xl font-bold mb-4">Mon Profil</h1>
        <div className="mb-4">
          <p>
            <span className="font-semibold">Nom :</span> {user.name}
          </p>
          <p>
            <span className="font-semibold">Email :</span> {user.email}
          </p>
        </div>
        <button
          onClick={() => setShowPredictions(!showPredictions)}
          className="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700"
        >
          {showPredictions ? "Masquer" : "Mes pronostics"}
        </button>
        {showPredictions && <PredictionHistory />}
      </div>
    </div>
  );
}
