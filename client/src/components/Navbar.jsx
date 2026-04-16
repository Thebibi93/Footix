import { Badge, Box, Button, Container, Group, Paper, Text } from "@mantine/core";
import { Link, useNavigate } from "react-router-dom";
import { useAuth } from "../contexts/AuthContext";

export default function Navbar() {
  const { user, logout } = useAuth();
  const navigate = useNavigate();

  const handleLogout = async () => {
    await logout();
    navigate("/");
  };

  return (
    <Container size="xl" h="100%" py={10}>
      <Paper className="glass-panel" h="100%" px="lg">
        <Group justify="space-between" align="center" h="100%" wrap="wrap">
          <Link to="/" style={{ display: "block", minWidth: 0 }}>
            <Box style={{ display: "flex", flexDirection: "column", justifyContent: "center", minHeight: 56 }}>
              <Text fw={900} fz={28} c="cyan.3" lh={1.05}>
                Footix
              </Text>
              <Text size="xs" c="dimmed" mt={4} lh={1.25} style={{ maxWidth: 320 }}>
                Football palpitant · chat live · pronostics
              </Text>
            </Box>
          </Link>

          <Group gap="sm" wrap="wrap">
            {user ? (
              <>
                <Badge variant="light" color="cyan">
                  {user.score ?? 0} pts
                </Badge>
                <Badge variant="outline" color="green">
                  Rang #{user.rank || "-"}
                </Badge>
                <Text size="sm" c="dimmed">
                  Bonjour, <Text span fw={700} c="white">{user.username}</Text>
                </Text>
                <Button component={Link} to="/profile" variant="light" color="cyan">
                  Mon profil
                </Button>
                <Button variant="subtle" color="red" onClick={handleLogout}>
                  Déconnexion
                </Button>
              </>
            ) : (
              <>
                <Button component={Link} to="/login" variant="subtle" color="gray">
                  Connexion
                </Button>
                <Button component={Link} to="/signup" color="cyan">
                  Inscription
                </Button>
              </>
            )}
          </Group>
        </Group>
      </Paper>
    </Container>
  );
}
