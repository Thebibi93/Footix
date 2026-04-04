import { useState } from "react";
import Match from "./components/Match";

// Données simulées
const simulatedMatches = [
  {
    id: 1,
    homeTeam: "Paris SG",
    awayTeam: "Olympique Lyonnais",
    location: "Parc des Princes",
    date: "2025-04-15T20:45:00",
    odds: { home: 1.85, draw: 3.6, away: 4.2 },
  },
  {
    id: 2,
    homeTeam: "Olympique de Marseille",
    awayTeam: "AS Monaco",
    location: "Stade Vélodrome",
    date: "2025-04-16T21:00:00",
    odds: { home: 2.1, draw: 3.4, away: 3.3 },
  },
  {
    id: 3,
    homeTeam: "RC Lens",
    awayTeam: "Stade Rennais",
    location: "Stade Bollaert-Delelis",
    date: "2025-04-17T19:00:00",
    odds: { home: 2.3, draw: 3.2, away: 2.95 },
  },
];

function App() {
  const [matches] = useState(simulatedMatches);
  const [betSlip, setBetSlip] = useState([]);
  const [stake, setStake] = useState(10);

  const addToBetSlip = (match, betType) => {
    const selection = {
      id: `${match.id}-${betType}`,
      matchId: match.id,
      matchName: `${match.homeTeam} - ${match.awayTeam}`,
      betType,
      odds: match.odds[betType],
      label: betType === "home" ? "1" : betType === "draw" ? "N" : "2",
    };
    if (!betSlip.some((s) => s.id === selection.id)) {
      setBetSlip([...betSlip, selection]);
    }
  };

  const removeFromBetSlip = (id) => {
    setBetSlip(betSlip.filter((s) => s.id !== id));
  };

  const totalOdds = betSlip.reduce((acc, sel) => acc * sel.odds, 1);
  const potentialWin = stake * totalOdds;

  const placeBet = () => {
    if (betSlip.length === 0) {
      alert("Veuillez ajouter au moins une sélection.");
      return;
    }
    const betData = {
      selections: betSlip.map((s) => ({
        matchId: s.matchId,
        betType: s.betType,
        odds: s.odds,
      })),
      stake,
      totalOdds,
      potentialWin,
    };
    console.log("Pari envoyé (simulation) :", betData);
    alert(
      `Pari placé ! Mise: ${stake}€, Gain potentiel: ${potentialWin.toFixed(2)}€`
    );
  };

  return (
    <div className="max-w-6xl mx-auto px-4 py-5">
      <header className="text-center mb-8">
        <h1 className="text-4xl font-bold text-gray-800">Footix</h1>
        <p className="text-gray-600">Paris sportifs sur le football</p>
      </header>

      <div className="flex flex-col md:flex-row gap-8">
        {/* Liste des matchs */}
        <section className="flex-[2] min-w-[300px]">
          <h2 className="text-2xl font-semibold mb-4">Matchs du jour</h2>
          <div className="flex flex-col gap-4">
            {matches.map((match) => (
              <div key={match.id} className="match-item">
                <Match
                  homeTeam={match.homeTeam}
                  awayTeam={match.awayTeam}
                  location={match.location}
                  startDate={match.date}
                />
                <div className="flex gap-3 justify-around mt-4">
                  <button
                    onClick={() => addToBetSlip(match, "home")}
                    className="bg-blue-600 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded-lg flex-1 transition-colors"
                  >
                    1 {match.odds.home}
                  </button>
                  <button
                    onClick={() => addToBetSlip(match, "draw")}
                    className="bg-blue-600 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded-lg flex-1 transition-colors"
                  >
                    N {match.odds.draw}
                  </button>
                  <button
                    onClick={() => addToBetSlip(match, "away")}
                    className="bg-blue-600 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded-lg flex-1 transition-colors"
                  >
                    2 {match.odds.away}
                  </button>
                </div>
              </div>
            ))}
          </div>
        </section>

        {/* Panier */}
        <section className="flex-1 min-w-[280px] bg-gray-800 rounded-xl p-5 shadow-md sticky top-5 self-start">
          <h2 className="text-2xl font-semibold text-white mb-4">Mon pari</h2>
          {betSlip.length === 0 ? (
            <p className="text-center text-gray-400 py-5">
              Aucune sélection pour le moment.
            </p>
          ) : (
            <>
              <ul className="list-none my-4 max-h-80 overflow-y-auto">
                {betSlip.map((sel) => (
                  <li
                    key={sel.id}
                    className="flex justify-between items-center py-2 border-b border-gray-600"
                  >
                    <div className="flex gap-2 flex-wrap">
                      <span className="font-medium text-white">
                        {sel.matchName}
                      </span>
                      <span className="bg-gray-600 px-2 py-0.5 rounded-full text-sm text-white">
                        ({sel.label})
                      </span>
                      <span className="font-bold text-yellow-400">
                        {sel.odds}
                      </span>
                    </div>
                    <button
                      className="bg-transparent border-none text-red-400 text-xl font-bold cursor-pointer hover:text-red-600"
                      onClick={() => removeFromBetSlip(sel.id)}
                    >
                      ×
                    </button>
                  </li>
                ))}
              </ul>
              <div className="mt-5 pt-3 border-t border-gray-600">
                <div className="text-white">
                  Cote totale : <strong>{totalOdds.toFixed(2)}</strong>
                </div>
                <div className="my-3 flex items-center gap-2 text-white">
                  <label>Mise (€) :</label>
                  <input
                    type="number"
                    value={stake}
                    onChange={(e) => setStake(Number(e.target.value))}
                    min="0.5"
                    step="0.5"
                    className="p-1.5 border border-gray-300 rounded-md w-24 text-gray-900"
                  />
                </div>
                <div className="text-lg my-3 text-green-400">
                  Gain potentiel : <strong>{potentialWin.toFixed(2)} €</strong>
                </div>
                <button
                  className="bg-blue-500 hover:bg-blue-600 text-white py-2 px-4 rounded-lg font-bold w-full transition-colors"
                  onClick={placeBet}
                >
                  Placer le pari
                </button>
              </div>
            </>
          )}
        </section>
      </div>
    </div>
  );
}

export default App;