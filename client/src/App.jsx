import {
  AppShell,
  Badge,
  Container,
  Group,
  Paper,
  SimpleGrid,
  Stack,
  Text,
  ThemeIcon,
  Title,
} from "@mantine/core";
import { Routes, Route, Navigate, useLocation } from "react-router-dom";
import { AuthProvider } from "./contexts/AuthContext";
import Navbar from "./components/Navbar";
import Leagues from "./components/Leagues";
import MatchList from "./components/MatchList";
import MatchDetail from "./components/MatchDetail";
import Login from "./components/auth/Login";
import Signup from "./components/auth/Signup";
import Profile from "./components/Profile";
import Breadcrumb from "./components/Breadcrumb";

const featureCards = [
  {
    title: "Deux vues de calendrier",
    description: "Observez les matchs à venir et les résultats terminés.",
    emoji: "📅",
  },
  {
    title: "Des pronostics palpitants",
    description: "Ressentez l'adréaline d'un pari 100% Hallal.",
    emoji: "🎯",
  },
  {
    title: "Communauté vivante",
    description: "Intéragissez avec d'autres fans de foots.",
    emoji: "💬",
  },
];

function Home() {
  return (
    <Stack gap="xl">
      <Paper className="glass-panel hero-panel" p={{ base: "xl", md: "2rem" }}>
        <Stack gap="lg">

          <Stack gap={6} maw={760}>
            <Title order={1} fz={{ base: 34, md: 54 }} lh={1.02}>
              Footix est un hub de pronostics, de scores et de conversations live.
            </Title>
            <Text size="lg" c="dimmed" maw={680}>
              Explorez les compétitions, basculez entre les matchs à venir et les affiches passées,
              suivez vos points et discutez avec la communauté dans une interface élégante et responsive.
            </Text>
          </Stack>

          <SimpleGrid cols={{ base: 1, md: 3 }} spacing="md">
            {featureCards.map((feature) => (
              <Paper key={feature.title} className="glass-panel profile-card" p="lg">
                <Group align="flex-start" wrap="nowrap">
                  <ThemeIcon radius="xl" size={46} variant="light" color="cyan">
                    <span style={{ fontSize: 20 }}>{feature.emoji}</span>
                  </ThemeIcon>
                  <Stack gap={4}>
                    <Text fw={700}>{feature.title}</Text>
                    <Text size="sm" c="dimmed">
                      {feature.description}
                    </Text>
                  </Stack>
                </Group>
              </Paper>
            ))}
          </SimpleGrid>
        </Stack>
      </Paper>

      <Leagues />
    </Stack>
  );
}

function AppContent() {
  const location = useLocation();

  return (
    <div className="footix-page">
      <div className="stadium-lines" aria-hidden="true" />

      <AppShell header={{ height: 96 }} padding="md" className="page-content">
        <AppShell.Header>
          <Navbar />
        </AppShell.Header>

        <AppShell.Main>
          <Container size="xl" py="lg">
            <Stack gap="md">
              <Breadcrumb />
              <Routes>
                <Route path="/" element={<Home />} />
                <Route path="/league/:leagueId" element={<MatchList key={location.pathname} />} />
                <Route path="/match/:matchId" element={<MatchDetail />} />
                <Route path="/login" element={<Login />} />
                <Route path="/signup" element={<Signup />} />
                <Route path="/profile" element={<Profile />} />
                <Route path="/logout" element={<Navigate to="/" />} />
              </Routes>
            </Stack>
          </Container>
        </AppShell.Main>
      </AppShell>
    </div>
  );
}

function App() {
  return (
    <AuthProvider>
      <AppContent />
    </AuthProvider>
  );
}

export default App;
