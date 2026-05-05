import { useEffect, useMemo, useState } from "react";
import {
  Alert,
  Badge,
  Button,
  Card,
  Center,
  Grid,
  Group,
  Loader,
  Paper,
  Progress,
  SimpleGrid,
  Stack,
  Text,
  ThemeIcon,
  Title,
  UnstyledButton,
} from "@mantine/core";
import { useNavigate, useParams } from "react-router-dom";
import { useApi } from "../hooks/UseApi";
import { useAuth } from "../contexts/AuthContext";
import MatchChat from "./MatchChat";

function getFormColor(result) {
  if (result === "W") {
    return "green";
  }
  if (result === "D") {
    return "yellow";
  }
  return "red";
}

function formatMatchDate(date) {
  return new Date(date).toLocaleString("fr-FR", {
    weekday: "long",
    day: "2-digit",
    month: "long",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

function getResultCopy(match) {
  if (!match.isFinished) {
    return "Rencontre ouverte";
  }
  if (match.homeScore > match.awayScore) {
    return `${match.homeTeam} s'impose`;
  }
  if (match.awayScore > match.homeScore) {
    return `${match.awayTeam} s'impose`;
  }
  return "Match nul";
}

export default function MatchDetail() {
  const { matchId } = useParams();
  const [match, setMatch] = useState(null);
  const [prediction, setPrediction] = useState("");
  const [savedPrediction, setSavedPrediction] = useState("");
  const [successMessage, setSuccessMessage] = useState("");
  const [errorMessage, setErrorMessage] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const { getMatchDetails, getUserPredictions, addPrediction, loading } = useApi();
  const { user } = useAuth();
  const navigate = useNavigate();

  useEffect(() => {
    let active = true;

    (async () => {
      try {
        const [data, predictions] = await Promise.all([
          getMatchDetails(matchId),
          user ? getUserPredictions() : Promise.resolve([]),
        ]);

        if (!active) {
          return;
        }

        setMatch(data);

        const existingPrediction = Array.isArray(predictions)
          ? predictions.find((item) => String(item.matchId) === String(matchId))
          : null;

        const nextPrediction = existingPrediction?.normalizedPredictedResult || "";
        setPrediction(nextPrediction);
        setSavedPrediction(nextPrediction);
      } catch {
        if (active) {
          setMatch(null);
          setPrediction("");
          setSavedPrediction("");
        }
      }
    })();

    return () => {
      active = false;
    };
  }, [getMatchDetails, getUserPredictions, matchId, user]);

  useEffect(() => {
    if (user) {
      return;
    }
    setPrediction("");
    setSavedPrediction("");
  }, [user]);

  useEffect(() => {
    if (!successMessage && !errorMessage) {
      return undefined;
    }

    const timeoutId = window.setTimeout(() => {
      setSuccessMessage("");
      setErrorMessage("");
    }, 3500);

    return () => window.clearTimeout(timeoutId);
  }, [successMessage, errorMessage]);

  const probabilityCards = useMemo(() => {
    if (!match) {
      return [];
    }

    const stats = match.stats || {
      homeWinProb: 33,
      drawProb: 34,
      awayWinProb: 33,
    };

    return [
      {
        label: `Victoire ${match.homeTeam}`,
        value: stats.homeWinProb,
        color: "cyan",
      },
      {
        label: "Match nul",
        value: stats.drawProb,
        color: "gray",
      },
      {
        label: `Victoire ${match.awayTeam}`,
        value: stats.awayWinProb,
        color: "green",
      },
    ];
  }, [match]);

  const predictionOptions = useMemo(() => {
    if (!match) {
      return [];
    }

    return [
      {
        value: "HOME_WIN",
        shortLabel: "1",
        title: match.homeTeam,
        subtitle: "Victoire domicile",
        color: "cyan",
      },
      {
        value: "DRAW",
        shortLabel: "N",
        title: "Match nul",
        subtitle: "Score partagé",
        color: "gray",
      },
      {
        value: "AWAY_WIN",
        shortLabel: "2",
        title: match.awayTeam,
        subtitle: "Victoire extérieur",
        color: "green",
      },
    ];
  }, [match]);

  const selectedPredictionLabel = useMemo(() => {
    const selected = predictionOptions.find((item) => item.value === prediction);
    return selected ? `${selected.shortLabel} · ${selected.title}` : "";
  }, [prediction, predictionOptions]);

  const handlePredict = async () => {
    if (!user) {
      navigate("/login");
      return;
    }
    if (!prediction) {
      return;
    }

    setSubmitting(true);
    setErrorMessage("");
    setSuccessMessage("");

    try {
      await addPrediction(matchId, prediction);
      const wasUpdate = Boolean(savedPrediction);
      setSavedPrediction(prediction);
      setSuccessMessage(
        wasUpdate ? "Pronostic mis à jour avec succès." : "Pronostic enregistré avec succès.",
      );
    } catch (err) {
      setErrorMessage(err.message || "Impossible d'enregistrer le pronostic");
    } finally {
      setSubmitting(false);
    }
  };

  if (loading && !match) {
    return (
      <Center py="xl">
        <Loader color="cyan" />
      </Center>
    );
  }

  if (!loading && !match) {
    return (
      <Alert color="red" variant="light">
        Match introuvable.
      </Alert>
    );
  }

  return (
    <Stack gap="lg">
      <Card className="glass-panel hero-panel">
        <Stack gap="lg">
          <Group justify="space-between" align="flex-start">
            <Stack gap={6}>
              <Badge variant="light" color={match.isFinished ? "gray" : "cyan"} w="fit-content">
                {match.isFinished ? "Résultat final" : match.isStarted ? "Coup d'envoi donné" : "Match à venir"}
              </Badge>
              <Title order={1}>{match.homeTeam} vs {match.awayTeam}</Title>
              <Text c="dimmed">{formatMatchDate(match.date)} · {match.leagueName}</Text>
            </Stack>
            <ThemeIcon radius="xl" size={88} variant="light" color={match.isFinished ? "green" : "cyan"}>
              <Text fw={900} size="xl">{match.isFinished ? match.scoreboard : "VS"}</Text>
            </ThemeIcon>
          </Group>

          {match.isFinished ? (
            <Paper className="kpi-card" p="lg">
              <Group justify="space-between" align="center">
                <div>
                  <Text c="dimmed" size="sm">Verdict</Text>
                  <Title order={3}>{getResultCopy(match)}</Title>
                </div>
                <Badge variant="filled" color="green">{match.scoreboard}</Badge>
              </Group>
            </Paper>
          ) : null}
        </Stack>
      </Card>

      <Grid gutter="md">
        <Grid.Col span={{ base: 12, lg: 7 }}>
          <Card className="glass-panel">
            <Stack gap="md">
              <Group justify="space-between">
                <Title order={3}>Lecture du match</Title>
                <Badge variant="outline" color="cyan">IA simple</Badge>
              </Group>

              {probabilityCards.map((item) => (
                <Paper key={item.label} className="kpi-card" p="md">
                  <Group justify="space-between" mb={8}>
                    <Text size="sm" c="dimmed">{item.label}</Text>
                    <Text fw={800}>{item.value}%</Text>
                  </Group>
                  <Progress value={item.value} color={item.color} radius="xl" size="md" />
                </Paper>
              ))}

              {match.stats?.recentForm?.home?.length ? (
                <Stack gap="sm" pt="sm">
                  <Title order={4}>Forme récente</Title>
                  <Group wrap="wrap">
                    <Text fw={700}>{match.homeTeam}</Text>
                    {match.stats.recentForm.home.map((result, index) => (
                      <Badge key={`home-${index}`} color={getFormColor(result)}>{result}</Badge>
                    ))}
                  </Group>
                  <Group wrap="wrap">
                    <Text fw={700}>{match.awayTeam}</Text>
                    {match.stats.recentForm.away.map((result, index) => (
                      <Badge key={`away-${index}`} color={getFormColor(result)}>{result}</Badge>
                    ))}
                  </Group>
                </Stack>
              ) : null}
            </Stack>
          </Card>
        </Grid.Col>

        <Grid.Col span={{ base: 12, lg: 5 }}>
          <Card className="glass-panel">
            <Stack gap="md">
              <Group justify="space-between">
                <Title order={3}>Interaction</Title>
                {user ? <Badge variant="light" color="cyan">{user.username}</Badge> : null}
              </Group>

              {match.canPredict ? (
                <>
                  {!user ? (
                    <Alert color="blue" variant="light">
                      Connectez-vous pour enregistrer votre pronostic avant le coup d'envoi.
                    </Alert>
                  ) : null}

                  <SimpleGrid cols={{ base: 1, sm: 3 }} spacing="sm">
                    {predictionOptions.map((option) => {
                      const isSelected = prediction === option.value;

                      return (
                        <UnstyledButton
                          key={option.value}
                          onClick={() => user && setPrediction(option.value)}
                          disabled={!user}
                          className="match-card"
                          style={{ width: "100%", opacity: user ? 1 : 0.62 }}
                        >
                          <Paper
                            p="md"
                            radius="lg"
                            withBorder
                            style={{
                              minHeight: 116,
                              display: "flex",
                              flexDirection: "column",
                              justifyContent: "space-between",
                              borderColor: isSelected
                                ? option.color === "green"
                                  ? "rgba(34, 197, 94, 0.6)"
                                  : option.color === "gray"
                                    ? "rgba(148, 163, 184, 0.55)"
                                    : "rgba(34, 211, 238, 0.62)"
                                : "rgba(148, 163, 184, 0.12)",
                              background: isSelected
                                ? option.color === "green"
                                  ? "linear-gradient(180deg, rgba(8, 60, 37, 0.96), rgba(5, 37, 24, 0.92))"
                                  : option.color === "gray"
                                    ? "linear-gradient(180deg, rgba(43, 52, 68, 0.96), rgba(26, 34, 48, 0.92))"
                                    : "linear-gradient(180deg, rgba(6, 50, 68, 0.96), rgba(4, 31, 43, 0.92))"
                                : "linear-gradient(180deg, rgba(10, 22, 34, 0.78), rgba(8, 18, 28, 0.9))",
                              boxShadow: isSelected ? "0 16px 32px rgba(2, 8, 23, 0.35)" : "none",
                              transition: "all 180ms ease",
                            }}
                          >
                            <Group justify="space-between" align="flex-start">
                              <Badge variant={isSelected ? "filled" : "light"} color={option.color}>
                                {option.shortLabel}
                              </Badge>
                              {isSelected ? (
                                <Badge variant="filled" color={option.color}>
                                  Choisi
                                </Badge>
                              ) : null}
                            </Group>

                            <div>
                              <Text fw={800}>{option.title}</Text>
                              <Text size="sm" c="dimmed" mt={4}>
                                {option.subtitle}
                              </Text>
                            </div>
                          </Paper>
                        </UnstyledButton>
                      );
                    })}
                  </SimpleGrid>

                  {savedPrediction ? (
                    <Alert color="cyan" variant="light">
                      Pronostic enregistré : <Text span fw={700}>{selectedPredictionLabel || savedPrediction}</Text>. Vous pouvez changer votre choix puis réenregistrer tant que le match n'a pas commencé.
                    </Alert>
                  ) : user ? (
                    <Text size="sm" c="dimmed">
                      Sélectionnez une issue ci-dessus. Le choix reste affiché sur la page et peut être modifié avant le coup d'envoi.
                    </Text>
                  ) : null}

                  <Button
                    onClick={handlePredict}
                    fullWidth
                    loading={submitting}
                    color="cyan"
                    disabled={!prediction}
                  >
                    {user
                      ? savedPrediction && savedPrediction !== prediction
                        ? "Enregistrer la modification"
                        : savedPrediction
                          ? "Réenregistrer ce pronostic"
                          : "Enregistrer mon pronostic"
                      : "Se connecter pour pronostiquer"}
                  </Button>
                </>
              ) : (
                <Alert color="gray" variant="light">
                  Ce match n'accepte plus de pronostic. L'espace conserve uniquement le résultat final.
                </Alert>
              )}

              {successMessage ? <Alert color="green" variant="light">{successMessage}</Alert> : null}
              {errorMessage ? <Alert color="red" variant="light">{errorMessage}</Alert> : null}
            </Stack>
          </Card>
        </Grid.Col>
      </Grid>

      {match.canChat ? (
        <MatchChat matchId={matchId} matchTitle={`${match.homeTeam} - ${match.awayTeam}`} />
      ) : (
        <Card className="glass-panel">
          <Stack gap="xs">
            <Title order={3}>Discussion fermée</Title>
            <Text c="dimmed">
              Le chat n'est plus affiché sur les matchs passés. Pas de complaintes sur le résultat !
            </Text>
          </Stack>
        </Card>
      )}
    </Stack>
  );
}
