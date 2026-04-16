import { useEffect, useState } from "react";
import { Anchor, Breadcrumbs as MantineBreadcrumbs, Text } from "@mantine/core";
import { Link, useLocation } from "react-router-dom";
import { useSimulatedApi } from "../hooks/UseApi";

export default function Breadcrumb() {
  const location = useLocation();
  const [breadcrumbs, setBreadcrumbs] = useState([]);
  const { getLeagues, getMatchDetails } = useSimulatedApi();

  useEffect(() => {
    let active = true;

    const buildBreadcrumbs = async () => {
      const pathSegments = location.pathname.split("/").filter(Boolean);
      const crumbs = [{ name: "Accueil", path: "/" }];

      let currentPath = "";

      for (let i = 0; i < pathSegments.length; i += 1) {
        const segment = pathSegments[i];
        currentPath += `/${segment}`;

        if (segment === "league" && pathSegments[i + 1]) {
          const leagueId = pathSegments[i + 1];
          const leagues = await getLeagues();
          const league = leagues.find((item) => item.id === Number.parseInt(leagueId, 10));

          if (league) {
            crumbs.push({
              name: league.name,
              path: `${currentPath}/${leagueId}`,
            });
          }

          i += 1;
          continue;
        }

        if (segment === "match" && pathSegments[i + 1]) {
          const matchId = pathSegments[i + 1];
          const match = await getMatchDetails(matchId);

          if (match) {
            crumbs.push({
              name: `${match.homeTeam} vs ${match.awayTeam}`,
              path: `${currentPath}/${matchId}`,
            });
          }

          i += 1;
          continue;
        }

        if (segment !== "league" && segment !== "match" && Number.isNaN(Number(segment))) {
          crumbs.push({
            name: segment.charAt(0).toUpperCase() + segment.slice(1),
            path: currentPath,
          });
        }
      }

      if (active) {
        setBreadcrumbs(crumbs);
      }
    };

    buildBreadcrumbs();

    return () => {
      active = false;
    };
  }, [location.pathname, getLeagues, getMatchDetails]);

  if (breadcrumbs.length <= 1) {
    return null;
  }

  return (
    <MantineBreadcrumbs separator="›" separatorMargin="sm">
      {breadcrumbs.map((crumb, index) =>
        index === breadcrumbs.length - 1 ? (
          <Text key={crumb.path} fw={700} c="white" size="sm">
            {crumb.name}
          </Text>
        ) : (
          <Anchor key={crumb.path} component={Link} to={crumb.path} size="sm" c="cyan.3">
            {crumb.name}
          </Anchor>
        ),
      )}
    </MantineBreadcrumbs>
  );
}
