import { Badge, Box, Button, Container, Group, Paper, Text } from "@mantine/core";
import { useMediaQuery } from "@mantine/hooks";
import { Link, useNavigate } from "react-router-dom";
import { useAuth } from "../contexts/AuthContext";

export default function Navbar() {
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  const isMobile = useMediaQuery("(max-width: 36em)");

  const handleLogout = async () => {
    await logout();
    navigate("/");
  };

  return (
    <Container size="xl" h="100%" py={10}>
      <Paper className="glass-panel" h="100%" px={isMobile ? "sm" : "lg"}>
        <Group justify="space-between" align="center" h="100%" wrap="wrap" gap={isMobile ? "xs" : "sm"}>
          <Link to="/" style={{ display: "block", minWidth: 0 }}>
            <Box style={{ display: "flex", flexDirection: "column", justifyContent: "center", minHeight: 56 }}>
              <Text fw={900} fz={isMobile ? 24 : 28} c="cyan.3" lh={1.05}>
                Footix
              </Text>
              <Text size="xs" c="dimmed" mt={4} lh={1.25} style={{ maxWidth: 320 }} visibleFrom="sm">
                Football palpitant · chat live · pronostics
              </Text>
            </Box>
          </Link>

          <Group gap={isMobile ? 4 : "sm"} wrap="wrap">
            {user ? (
              <>
                <Badge variant="light" color="cyan" size={isMobile ? "sm" : "md"}>
                  {user.score ?? 0} pts
                </Badge>
                <Badge variant="outline" color="green" size={isMobile ? "sm" : "md"}>
                  Rang #{user.rank || "-"}
                </Badge>
                <Text size="sm" c="dimmed" visibleFrom="sm">
                  Bonjour, <Text span fw={700} c="white">{user.username}</Text>
                </Text>
                <Button component={Link} to="/profile" variant="light" color="cyan" size={isMobile ? "xs" : "sm"}>
                  Profil
                </Button>
                <Button variant="subtle" color="red" onClick={handleLogout} size={isMobile ? "xs" : "sm"}>
                  Déco
                </Button>
              </>
            ) : (
              <>
                <Button component={Link} to="/login" variant="subtle" color="gray" size={isMobile ? "xs" : "sm"}>
                  Connexion
                </Button>
                <Button component={Link} to="/signup" color="cyan" size={isMobile ? "xs" : "sm"}>
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