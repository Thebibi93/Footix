import { useEffect, useMemo, useState } from "react";
import {
  Badge,
  Card,
  Center,
  Divider,
  Group,
  Loader,
  Pagination,
  SegmentedControl,
  SimpleGrid,
  Stack,
  Text,
  ThemeIcon,
  Title,
} from "@mantine/core";
import { useParams, Link } from "react-router-dom";
import { useApi } from "../hooks/UseApi";

function formatMatchDate(date) {
  return new Date(date).toLocaleString("fr-FR", {
    weekday: "short",
    day: "2-digit",
    month: "short",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

function statusLabel(match) {
  if (match.isFinished) {
    return "Terminé";
  }
  if (match.isStarted) {
    return "En cours";
  }
  return "À venir";
}

export default function MatchList() {
  const { leagueId } = useParams();
  const [league, setLeague] = useState(null);
  const [bucket, setBucket] = useState("upcoming");
  const [page, setPage] = useState(1);
  const [response, setResponse] = useState({ items: [], page: 1, totalPages: 0, total: 0 });
  const [error, setError] = useState("");
  const { getLeagues, getMatchesByLeague, loading } = useApi();

  useEffect(() => {
    setPage(1);
  }, [bucket, leagueId]);

  useEffect(() => {
    let active = true;

    (async () => {
      try {
        const leagues = await getLeagues();
        const currentLeague = leagues.find((item) => String(item.id) === String(leagueId)) || null;
        const data = await getMatchesByLeague(leagueId, {
          bucket,
          page,
          pageSize: 12,
        });

        if (active) {
          setLeague(currentLeague);
          setResponse(data);
          setError("");
        }
      } catch (err) {
        if (active) {
          setError(err.message || "Impossible de charger les matchs");
        }
      }
    })();

    return () => {
      active = false;
    };
  }, [leagueId, bucket, page, getLeagues, getMatchesByLeague]);

  const title = useMemo(
    () => (bucket === "past" ? "Matchs passés" : "Matchs à venir"),
    [bucket],
  );

  return (
    <Stack gap="lg">
      <Card className="glass-panel hero-panel">
        <Group justify="space-between" align="flex-start" gap="md">
          <div>
            <Badge variant="light" color="cyan" className="section-label">
              {league?.code || "Ligue"}
            </Badge>
            <Title order={1} mt={8}>{league?.name || "Compétition"}</Title>
            <Text c="dimmed" mt={6} maw={760}>
              Alternez entre les rencontres ouvertes aux pronostics et les résultats déjà scellés.
            </Text>
          </div>
          <Group gap="xs">
            <Badge variant="outline" color="green">{response.total} matchs</Badge>
            <Badge variant="light" color="cyan">Page {response.page}</Badge>
          </Group>
        </Group>

        <Divider my="lg" />

        <SegmentedControl
          value={bucket}
          onChange={setBucket}
          data={[
            { label: "À venir", value: "upcoming" },
            { label: "Passés", value: "past" },
          ]}
          fullWidth
        />
      </Card>

      <Group justify="space-between" align="center">
        <div>
          <Title order={2}>{title}</Title>
          <Text c="dimmed" size="sm">
            {bucket === "past"
              ? "Consultez les scores finaux et les résultats des affiches déjà jouées."
              : "Pronostiquez et échangez avant le coup d'envoi sur les rencontres encore ouvertes."}
          </Text>
        </div>
      </Group>

      {loading && response.items.length === 0 ? (
        <Center py="xl">
          <Loader color="cyan" />
        </Center>
      ) : null}

      {error ? <Text c="red.3">{error}</Text> : null}

      {!loading && !error && response.items.length === 0 ? (
        <Center py="xl">
          <Text c="dimmed">Aucun match trouvé pour cet onglet.</Text>
        </Center>
      ) : null}

      <SimpleGrid cols={{ base: 1, lg: 2 }} spacing="md">
        {response.items.map((match) => (
          <Card key={match.id} component={Link} to={`/match/${match.id}`} className="glass-panel match-card">
            <Stack gap="md">
              <Group justify="space-between">
                <Badge color={match.isFinished ? "gray" : "cyan"} variant={match.isFinished ? "light" : "filled"}>
                  {statusLabel(match)}
                </Badge>
                <Text size="sm" c="dimmed">
                  {formatMatchDate(match.date)}
                </Text>
              </Group>

              <Group justify="space-between" wrap="nowrap" align="center">
                <Stack gap={4} style={{ flex: 1 }}>
                  <Text fw={800} size="lg">{match.homeTeam}</Text>
                  <Text c="dimmed" size="sm">Domicile</Text>
                </Stack>

                {match.isFinished ? (
                  <ThemeIcon radius="xl" size={58} variant="light" color="green">
                    <Text fw={900} size="lg">{match.scoreboard}</Text>
                  </ThemeIcon>
                ) : (
                  <Badge variant="outline" size="lg">VS</Badge>
                )}

                <Stack gap={4} style={{ flex: 1, textAlign: "right" }}>
                  <Text fw={800} size="lg">{match.awayTeam}</Text>
                  <Text c="dimmed" size="sm">Extérieur</Text>
                </Stack>
              </Group>

              <Group justify="space-between" align="center">
                <Text size="sm" c="dimmed">{match.location}</Text>
                <Text size="sm" fw={700} c="cyan.3">Voir le détail →</Text>
              </Group>
            </Stack>
          </Card>
        ))}
      </SimpleGrid>

      {response.totalPages > 1 ? (
        <Center pt="md">
          <Pagination value={page} onChange={setPage} total={response.totalPages} color="cyan" />
        </Center>
      ) : null}
    </Stack>
  );
}
