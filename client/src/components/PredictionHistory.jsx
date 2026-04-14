import { useEffect, useMemo, useState } from "react";
import { Alert, Badge, Card, Center, Group, Loader, Pagination, Stack, Text, Title } from "@mantine/core";
import { useSimulatedApi } from "../hooks/useSimulatedApi";
import { useAuth } from "../contexts/AuthContext";

const PAGE_SIZE = 6;

function getHistoryStatus(prediction) {
  if (!prediction.actualResult) {
    return {
      label: "À venir",
      color: "gray",
      variant: "light",
    };
  }

  if (prediction.points > 0) {
    return {
      label: `+${prediction.points} pt`,
      color: "green",
      variant: "filled",
    };
  }

  return {
    label: "0 pt",
    color: "red",
    variant: "light",
  };
}

export default function PredictionHistory() {
  const [predictions, setPredictions] = useState([]);
  const [error, setError] = useState("");
  const [page, setPage] = useState(1);
  const { getUserPredictions, loading } = useSimulatedApi();
  const { user } = useAuth();

  useEffect(() => {
    let active = true;

    if (!user) {
      return undefined;
    }

    (async () => {
      try {
        const data = await getUserPredictions();
        if (active) {
          setPredictions(data);
          setError("");
          setPage(1);
        }
      } catch (err) {
        if (active) {
          setPredictions([]);
          setError(err.message || "Impossible de charger l'historique");
        }
      }
    })();

    return () => {
      active = false;
    };
  }, [getUserPredictions, user]);

  const totalPages = Math.max(1, Math.ceil(predictions.length / PAGE_SIZE));
  const pageItems = useMemo(
    () => predictions.slice((page - 1) * PAGE_SIZE, page * PAGE_SIZE),
    [predictions, page],
  );

  if (loading) {
    return (
      <Center py="md">
        <Loader size="sm" color="cyan" />
      </Center>
    );
  }

  return (
    <Stack gap="md">
      <Group justify="space-between">
        <Title order={3}>Historique des pronostics</Title>
        <Badge variant="light" color="cyan">{predictions.length} entrées</Badge>
      </Group>

      {error ? (
        <Alert color="red" variant="light">
          {error}
        </Alert>
      ) : null}

      {!error && predictions.length === 0 ? (
        <Alert color="gray" variant="light">
          Aucun pronostic pour le moment.
        </Alert>
      ) : null}

      {!error && pageItems.map((prediction) => {
        const status = getHistoryStatus(prediction);

        return (
          <Card key={prediction.id} className="glass-panel profile-card">
            <Group justify="space-between" align="flex-start">
              <div>
                <Text fw={700}>{prediction.match}</Text>
                <Text size="sm" c="dimmed">Pronostic : {prediction.predictedResult}</Text>
                <Text size="sm" c="dimmed">Créé le : {prediction.date}</Text>
                <Text size="sm" c="dimmed">Match : {prediction.matchDate}</Text>
                {prediction.actualResult ? (
                  <Text size="sm" c="dimmed">Résultat réel : {prediction.actualResult}</Text>
                ) : null}
              </div>
              <Badge color={status.color} variant={status.variant}>{status.label}</Badge>
            </Group>
          </Card>
        );
      })}

      {predictions.length > PAGE_SIZE ? (
        <Center pt="sm">
          <Pagination total={totalPages} value={page} onChange={setPage} color="cyan" />
        </Center>
      ) : null}
    </Stack>
  );
}
