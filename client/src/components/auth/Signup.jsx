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

export default function Signup() {
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const { signup } = useAuth();
  const navigate = useNavigate();

  const handleSubmit = async (event) => {
    event.preventDefault();
    setError("");
    setSubmitting(true);

    try {
      await signup(name, email, password);
      navigate("/");
    } catch (err) {
      setError(err.message || "Impossible de créer le compte");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Center mih="70vh">
      <Paper className="glass-panel auth-card" w="100%" maw={520} p="xl">
        <Stack gap="lg">
          <div>
            <Title order={2} ta="center">Créer un compte</Title>
            <Text c="dimmed" ta="center" size="sm" mt={6}>
              Entrez dans Footix, cumulez des points et construisez votre réputation de pronostiqueur.
            </Text>
          </div>

          {error ? <Alert color="red" variant="light">{error}</Alert> : null}

          <form onSubmit={handleSubmit}>
            <Stack gap="md">
              <TextInput
                label="Nom d'utilisateur"
                placeholder="Votre pseudo"
                value={name}
                onChange={(event) => setName(event.currentTarget.value)}
                required
              />
              <TextInput
                label="Email"
                type="email"
                placeholder="vous@example.com"
                value={email}
                onChange={(event) => setEmail(event.currentTarget.value)}
                required
              />
              <PasswordInput
                label="Mot de passe"
                placeholder="Choisissez un mot de passe"
                value={password}
                onChange={(event) => setPassword(event.currentTarget.value)}
                required
              />
              <Button type="submit" fullWidth loading={submitting} color="cyan">
                Créer mon compte
              </Button>
            </Stack>
          </form>

          <Text ta="center" size="sm" c="dimmed">
            Déjà inscrit ? <Anchor component={Link} to="/login">Connexion</Anchor>
          </Text>
        </Stack>
      </Paper>
    </Center>
  );
}
