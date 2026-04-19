import { useState } from "react";
import {
  Alert,
  Anchor,
  Button,
  Center,
  Paper,
  PasswordInput,
  Stack,
  Text,
  TextInput,
  Title,
} from "@mantine/core";
import { Link, useNavigate } from "react-router-dom";
import { useAuth } from "../../contexts/AuthContext";

export default function Login() {
  const [identifier, setIdentifier] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const { login } = useAuth();
  const navigate = useNavigate();

  const handleSubmit = async (event) => {
    event.preventDefault();
    setError("");
    setSubmitting(true);

    try {
      await login(identifier, password);
      navigate("/");
    } catch (err) {
      setError(err.message || "Identifiants invalides");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Center mih="70vh">
      <Paper className="glass-panel auth-card" w="100%" maw={500} p="xl">
        <Stack gap="lg">
          <div>
            <Title order={2} ta="center">Connexion</Title>
            <Text c="dimmed" ta="center" size="sm" mt={6}>
              Rejoignez l'arène des pronostics et retrouvez vos points en temps réel.
            </Text>
          </div>

          {error ? <Alert color="red" variant="light">{error}</Alert> : null}

          <form onSubmit={handleSubmit}>
            <Stack gap="md">
              <TextInput
                label="Nom d'utilisateur ou email"
                placeholder="idris ou vous@example.com"
                value={identifier}
                onChange={(event) => setIdentifier(event.currentTarget.value)}
                required
              />
              <PasswordInput
                label="Mot de passe"
                placeholder="Votre mot de passe"
                value={password}
                onChange={(event) => setPassword(event.currentTarget.value)}
                required
              />
              <Button type="submit" fullWidth loading={submitting} color="cyan">
                Se connecter
              </Button>
            </Stack>
          </form>

          <Text ta="center" size="sm" c="dimmed">
            Pas encore de compte ? <Anchor component={Link} to="/signup">Créer mon profil</Anchor>
          </Text>
        </Stack>
      </Paper>
    </Center>
  );
}
