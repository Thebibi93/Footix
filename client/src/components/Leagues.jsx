import { useEffect, useState } from "react";
import { useSimulatedApi } from "../hooks/useSimulatedApi";
import { Link } from "react-router-dom";

export default function Leagues() {
  const [leagues, setLeagues] = useState([]);
  const { getLeagues, loading } = useSimulatedApi();

  useEffect(() => {
    getLeagues().then(setLeagues);
  }, []);

  if (loading)
    return <div className="text-center py-10">Chargement des ligues...</div>;

  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
      {leagues.map((league) => (
        <Link
          key={league.id}
          to={`/league/${league.id}`}
          className="bg-white rounded-xl shadow-md p-6 hover:shadow-lg transition transform hover:-translate-y-1"
        >
          <h3 className="text-xl font-bold text-gray-800">{league.name}</h3>
          <p className="text-gray-500">{league.country}</p>
        </Link>
      ))}
    </div>
  );
}
