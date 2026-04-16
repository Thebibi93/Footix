import { useEffect, useState } from "react";
import {
  Badge,
  Card,
  Center,
  Group,
  Loader,
  SimpleGrid,
  Stack,
  Text,
  ThemeIcon,
  Title,
} from "@mantine/core";
import { Link } from "react-router-dom";
import { motion } from "framer-motion";
import { useSimulatedApi } from "../hooks/UseApi";

const MotionCard = motion.create(Card);

export default function Leagues() {
  const [leagues, setLeagues] = useState([]);
  const [error, setError] = useState("");
  const { getLeagues, loading } = useSimulatedApi();

  useEffect(() => {
    let active = true;

    (async () => {
      try {
        const data = await getLeagues();
        if (active) {
          setLeagues(data);
          setError("");
        }
      } catch (err) {
        if (active) {
          setError(err.message || "Impossible de charger les ligues");
        }
      }
    })();

    return () => {
      active = false;
    };
  }, [getLeagues]);

  if (loading && leagues.length === 0) {
    return (
      <Center py="xl">
        <Loader color="cyan" />
      </Center>
    );
  }

  return (
    <Stack gap="lg">
      <Group justify="space-between" align="end">
        <div>
          <Badge variant="light" color="green" className="section-label">
            Compétitions
          </Badge>
          <Title order={2} mt={8}>
            Choisissez votre ligue
          </Title>
          <Text c="dimmed" mt={4}>
            Parcourez les affiches à venir, les résultats passés et les performances des équipes.
          </Text>
        </div>
        <Text c="dimmed">{leagues.length} ligues disponibles</Text>
      </Group>

      {error ? <Text c="red.3">{error}</Text> : null}

      <SimpleGrid cols={{ base: 1, sm: 2, lg: 4 }} spacing="md">
        {leagues.map((league, index) => (
          <MotionCard
            key={league.id}
            component={Link}
            to={`/league/${league.id}`}
            className="glass-panel match-card"
            initial={{ opacity: 0, y: 18 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.35, delay: index * 0.05 }}
          >
            <Stack gap="md">
              <Group justify="space-between" align="flex-start">
                <Badge variant="light" color="cyan">
                  {league.code}
                </Badge>
                <ThemeIcon variant="light" radius="xl" color="green" size={42}>
                  ⚽
                </ThemeIcon>
              </Group>

              <div>
                <Title order={3}>{league.name}</Title>
                <Text size="sm" c="dimmed" mt={6}>
                  Accès rapide aux matchs à venir, aux résultats et aux discussions liées à la compétition.
                </Text>
              </div>

              <Text size="sm" fw={700} c="cyan.3">
                Ouvrir la ligue →
              </Text>
            </Stack>
          </MotionCard>
        ))}
      </SimpleGrid>
    </Stack>
  );
}
