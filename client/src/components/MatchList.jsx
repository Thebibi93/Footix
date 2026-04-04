import { useEffect, useState } from "react";
import { useParams, Link } from "react-router-dom";
import { useSimulatedApi } from "../hooks/useSimulatedApi";

export default function MatchList() {
  const { leagueId } = useParams();
  const [matches, setMatches] = useState([]);
  const { getMatchesByLeague, loading } = useSimulatedApi();

  useEffect(() => {
    let isMounted = true;
    if (leagueId) {
      getMatchesByLeague(leagueId).then((data) => {
        if (isMounted) setMatches(data);
      });
    }
    return () => {
      isMounted = false;
    };
  }, [leagueId, getMatchesByLeague]);

  if (loading)
    return <div className="text-center py-10">Chargement des matchs...</div>;

  return (
    <div className="space-y-4">
      <h2 className="text-2xl font-bold mb-4">Matchs à venir</h2>
      {matches.length === 0 && <p>Aucun match trouvé.</p>}
      {matches.map((match) => (
        <Link
          to={`/match/${match.id}`}
          key={match.id}
          className="block bg-white rounded-xl p-5 shadow hover:shadow-md transition"
        >
          <div className="flex justify-between items-center">
            <div className="flex-1 text-center">
              <span className="font-bold text-lg">{match.homeTeam}</span>
            </div>
            <div className="px-4 text-gray-500">vs</div>
            <div className="flex-1 text-center">
              <span className="font-bold text-lg">{match.awayTeam}</span>
            </div>
          </div>
          <div className="text-sm text-gray-500 text-center mt-2">
            📍 {match.location} | 📅{" "}
            {new Date(match.date).toLocaleString("fr-FR")}
          </div>
        </Link>
      ))}
    </div>
  );
}
