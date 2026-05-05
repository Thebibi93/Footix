import { useEffect, useMemo, useState } from "react";
import {
  Alert,
  Badge,
  Card,
  Center,
  Grid,
  Group,
  Loader,
  Paper,
  Stack,
  Tabs,
  Text,
  Title,
} from "@mantine/core";
import { useAuth } from "../contexts/AuthContext";
import { useApi } from "../hooks/UseApi";
import PredictionHistory from "./PredictionHistory";

export default function Profile() {
  const { user, loading: authLoading } = useAuth();
  const { getProfileSummary, getLeaderboard, loading } = useApi();
  const [profile, setProfile] = useState(null);
  const [leaderboard, setLeaderboard] = useState([]);
  const [error, setError] = useState("");

  useEffect(() => {
    let active = true;

    if (!user?.id) {
      setProfile(null);
      setLeaderboard([]);
      setError("");
      return undefined;
    }

    (async () => {
      try {
        const [profileData, leaderboardData] = await Promise.all([
          getProfileSummary(),
          getLeaderboard(),
        ]);

        if (!active) {
          return;
        }

        setProfile(profileData);
        setLeaderboard(Array.isArray(leaderboardData) ? leaderboardData : []);
        setError("");
      } catch (err) {
        if (!active) {
          return;
        }
        setError(err.message || "Impossible de charger le profil");
      }
    })();

    return () => {
      active = false;
    };
  }, [user?.id, getProfileSummary, getLeaderboard]);

  const surroundingUsers = useMemo(() => {
    if (!profile || leaderboard.length === 0) {
      return [];
    }

    const index = leaderboard.findIndex(
      (entry) => Number(entry.userId) === Number(profile.id),
    );

    if (index === -1) {
      return leaderboard.slice(0, 5);
    }

    return leaderboard.slice(
      Math.max(0, index - 2),
      Math.min(leaderboard.length, index + 3),
    );
  }, [leaderboard, profile]);

  if (authLoading || loading) {
    return (
      <Center py="xl">
        <Loader color="cyan" />
      </Center>
    );
  }

  if (!user) {
    return (
      <Center py="xl">
        <Alert color="blue" variant="light">
          Veuillez vous connecter pour accéder à votre profil.
        </Alert>
      </Center>
    );
  }

  return (
    <Stack gap="lg">
      <Card className="glass-panel hero-panel">
        <Stack gap="md">
          <Group justify="space-between" align="flex-start">
            <div>
              <Badge variant="light" color="cyan" className="section-label">
                Espace personnel
              </Badge>
              <Title order={1} mt={8}>
                {profile?.username || user.username}
              </Title>
              <Text c="dimmed" mt={6}>
                Visualisez vos points, votre rang, votre précision et l'historique de vos paris en un seul endroit.
              </Text>
            </div>
            <Badge variant="filled" color="green">
              {profile?.score ?? user.score ?? 0} pts
            </Badge>
          </Group>

          {error ? (
            <Alert color="red" variant="light">
              {error}
            </Alert>
          ) : null}

          <Grid gutter="md">
            <Grid.Col span={{ base: 12, sm: 6, lg: 3 }}>
              <Paper className="kpi-card" p="lg">
                <Text size="sm" c="dimmed">Rang</Text>
                <Title order={2}>#{profile?.rank || user.rank || "-"}</Title>
              </Paper>
            </Grid.Col>
            <Grid.Col span={{ base: 12, sm: 6, lg: 3 }}>
              <Paper className="kpi-card" p="lg">
                <Text size="sm" c="dimmed">Prédictions</Text>
                <Title order={2}>{profile?.totalPredictions ?? user.totalPredictions ?? 0}</Title>
              </Paper>
            </Grid.Col>
            <Grid.Col span={{ base: 12, sm: 6, lg: 3 }}>
              <Paper className="kpi-card" p="lg">
                <Text size="sm" c="dimmed">Réussite</Text>
                <Title order={2}>
                  {(profile?.accuracy ?? user.accuracy ?? 0).toFixed(0)}%
                </Title>
              </Paper>
            </Grid.Col>
            <Grid.Col span={{ base: 12, sm: 6, lg: 3 }}>
              <Paper className="kpi-card" p="lg">
                <Text size="sm" c="dimmed">Messages envoyés</Text>
                <Title order={2}>{profile?.chatMessages ?? user.chatMessages ?? 0}</Title>
              </Paper>
            </Grid.Col>
          </Grid>
        </Stack>
      </Card>

      <Tabs defaultValue="overview" color="cyan">
        <Tabs.List>
          <Tabs.Tab value="overview">Vue d'ensemble</Tabs.Tab>
          <Tabs.Tab value="history">Historique</Tabs.Tab>
        </Tabs.List>

        <Tabs.Panel value="overview" pt="md">
          <Grid gutter="md">
            <Grid.Col span={{ base: 12, lg: 5 }}>
              <Card className="glass-panel profile-card">
                <Stack gap="sm">
                  <Title order={3}>Carte identité</Title>
                  <Text c="dimmed">Pseudo</Text>
                  <Text fw={700}>{profile?.username || user.username}</Text>
                  <Text c="dimmed" mt="sm">Email</Text>
                  <Text fw={700}>{profile?.email || user.email}</Text>
                  <Text c="dimmed" mt="sm">Prédictions correctes</Text>
                  <Text fw={700}>{profile?.correctPredictions ?? user.correctPredictions ?? 0}</Text>
                </Stack>
              </Card>
            </Grid.Col>

            <Grid.Col span={{ base: 12, lg: 7 }}>
              <Card className="glass-panel profile-card">
                <Stack gap="sm">
                  <Group justify="space-between">
                    <Title order={3}>Autour de vous au classement</Title>
                    <Badge variant="light" color="cyan">Top dynamique</Badge>
                  </Group>

                  {surroundingUsers.map((entry, idx) => {
                    const isMe = Number(entry.userId) === Number(profile?.id || user?.id);
                    const rank = leaderboard.findIndex(
                      (item) => Number(item.userId) === Number(entry.userId),
                    ) + 1;

                    return (
                      <Group key={`${entry.userId}-${entry.score}-${idx}`} justify="space-between">
                        <Group gap="sm">
                          <Badge
                            variant={isMe ? "filled" : "light"}
                            color={isMe ? "cyan" : "gray"}
                          >
                            #{rank}
                          </Badge>
                          <Text fw={isMe ? 700 : 500}>{entry.username}</Text>
                        </Group>
                        <Text fw={700}>{entry.score} pts</Text>
                      </Group>
                    );
                  })}
                </Stack>
              </Card>
            </Grid.Col>
          </Grid>
        </Tabs.Panel>

        <Tabs.Panel value="history" pt="md">
          <PredictionHistory />
        </Tabs.Panel>
      </Tabs>
    </Stack>
  );
}