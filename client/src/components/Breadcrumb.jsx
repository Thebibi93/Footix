import { useEffect, useState } from "react";
import { Link, useLocation } from "react-router-dom";
import { useSimulatedApi } from "../hooks/useSimulatedApi";

// TODO : Cacher probablement les données pour un meilleur chargement
export default function Breadcrumb() {
  const location = useLocation();
  const [breadcrumbs, setBreadcrumbs] = useState([]);
  const { getLeagues, getMatchDetails } = useSimulatedApi();

  useEffect(() => {
    const buildBreadcrumbs = async () => {
      const pathSegments = location.pathname.split("/").filter(Boolean);
      const crumbs = [];

      // Accueil toujours présent
      crumbs.push({ name: "Accueil", path: "/" });

      let currentPath = "";
      for (let i = 0; i < pathSegments.length; i++) {
        const segment = pathSegments[i];
        currentPath += `/${segment}`;

        if (segment === "league" && pathSegments[i + 1]) {
          const leagueId = pathSegments[i + 1];
          const leagues = await getLeagues();
          const league = leagues.find((l) => l.id === parseInt(leagueId));
          if (league) {
            crumbs.push({
              name: league.name,
              path: currentPath + `/${leagueId}`,
            });
          }
          i++; // sauter l'id
        } else if (segment === "match" && pathSegments[i + 1]) {
          const matchId = pathSegments[i + 1];
          const match = await getMatchDetails(matchId);
          if (match) {
            // TODO : Mettre le nom de la ligue
            crumbs.push({
              name: `${match.homeTeam} vs ${match.awayTeam}`,
              path: currentPath + `/${matchId}`,
            });
          }
          i++;
        } else if (
          segment !== "league" &&
          segment !== "match" &&
          isNaN(segment)
        ) {
          // Pages spéciales comme profile, login, etc.
          const name = segment.charAt(0).toUpperCase() + segment.slice(1);
          crumbs.push({ name, path: currentPath });
        }
      }

      setBreadcrumbs(crumbs);
    };

    buildBreadcrumbs();
  }, [location, getLeagues, getMatchDetails]);

  if (breadcrumbs.length <= 1) return null; // on n'affiche que s'il y a plus que l'accueil

  return (
    <nav className="text-sm text-gray-500 mb-4" aria-label="Fil d'Ariane">
      <ol className="flex flex-wrap items-center space-x-2">
        {breadcrumbs.map((crumb, index) => (
          <li key={crumb.path} className="flex items-center">
            {index > 0 && <span className="mx-2 text-gray-400">›</span>}
            {index === breadcrumbs.length - 1 ? (
              <span className="font-medium text-gray-700">{crumb.name}</span>
            ) : (
              <Link to={crumb.path} className="hover:text-blue-600 transition">
                {crumb.name}
              </Link>
            )}
          </li>
        ))}
      </ol>
    </nav>
  );
}
