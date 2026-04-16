import { useEffect, useMemo, useState } from "react";
import {
  Alert,
  Avatar,
  Badge,
  Button,
  Card,
  Group,
  Loader,
  Modal,
  Paper,
  ScrollArea,
  Stack,
  Text,
  TextInput,
  Title,
} from "@mantine/core";
import { useSimulatedApi } from "../hooks/UseApi";
import { useAuth } from "../contexts/AuthContext";

function initials(name) {
  return String(name || "?")
    .split(" ")
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0]?.toUpperCase())
    .join("");
}

export default function MatchChat({ matchId, matchTitle }) {
  const [messages, setMessages] = useState([]);
  const [newMessage, setNewMessage] = useState("");
  const [loadingMessages, setLoadingMessages] = useState(true);
  const [error, setError] = useState("");
  const [sending, setSending] = useState(false);
  const [selectedUserProfile, setSelectedUserProfile] = useState(null);
  const [profileOpened, setProfileOpened] = useState(false);
  const [profileLoading, setProfileLoading] = useState(false);
  const { getChatMessages, sendChatMessage, getUserProfile } = useSimulatedApi();
  const { user } = useAuth();

  const lastSeq = useMemo(
    () => messages.reduce((max, item) => Math.max(max, item.seqInChat || 0), 0),
    [messages],
  );

  useEffect(() => {
    let active = true;

    const loadInitialMessages = async () => {
      if (!matchId) {
        return;
      }

      setLoadingMessages(true);
      try {
        const data = await getChatMessages(matchId, 0, 100);
        if (active) {
          setMessages(data);
          setError("");
        }
      } catch (err) {
        if (active) {
          setMessages([]);
          setError(err.message || "Impossible de charger le chat");
        }
      } finally {
        if (active) {
          setLoadingMessages(false);
        }
      }
    };

    loadInitialMessages();

    return () => {
      active = false;
    };
  }, [getChatMessages, matchId]);

  useEffect(() => {
    if (!matchId) {
      return undefined;
    }

    const intervalId = window.setInterval(async () => {
      try {
        const freshMessages = await getChatMessages(matchId, lastSeq, 100);
        if (freshMessages.length === 0) {
          return;
        }
        setMessages((previous) => {
          const existingIds = new Set(previous.map((item) => item.id));
          const merged = [...previous];
          for (const item of freshMessages) {
            if (!existingIds.has(item.id)) {
              merged.push(item);
            }
          }
          merged.sort((a, b) => a.seqInChat - b.seqInChat);
          return merged;
        });
      } catch {
        // polling silencieux
      }
    }, 3000);

    return () => window.clearInterval(intervalId);
  }, [getChatMessages, lastSeq, matchId]);

  const openUserProfile = async (userId) => {
    setProfileOpened(true);
    setProfileLoading(true);
    try {
      const profile = await getUserProfile(userId);
      setSelectedUserProfile(profile);
    } catch {
      setSelectedUserProfile(null);
    } finally {
      setProfileLoading(false);
    }
  };

  const handleSend = async () => {
    const value = newMessage.trim();
    if (!value || !user) {
      return;
    }

    setSending(true);
    setError("");

    try {
      const created = await sendChatMessage(matchId, value);
      setMessages((previous) => [...previous, created].sort((a, b) => a.seqInChat - b.seqInChat));
      setNewMessage("");
    } catch (err) {
      setError(err.message || "Impossible d'envoyer le message");
    } finally {
      setSending(false);
    }
  };

  return (
    <>
      <Card className="glass-panel chat-card">
        <Stack gap="md">
          <Group justify="space-between" align="center">
            <div>
              <Title order={3}>Salon du match</Title>
              <Text size="sm" c="dimmed">
                {matchTitle} · Cliquez sur un pseudo pour ouvrir son profil public.
              </Text>
            </div>
            <Badge variant="light" color="cyan">
              {messages.length} messages
            </Badge>
          </Group>

          {error ? <Alert color="red" variant="light">{error}</Alert> : null}
          {!user ? (
            <Alert color="blue" variant="light">
              Connectez-vous pour participer au chat.
            </Alert>
          ) : null}

          <Paper className="kpi-card" p={0}>
            <ScrollArea h={360} offsetScrollbars>
              <Stack gap="sm" p="md">
                {loadingMessages ? <Loader color="cyan" size="sm" /> : null}

                {!loadingMessages && messages.length === 0 ? (
                  <Text c="dimmed">Aucun message pour le moment. Lancez la discussion.</Text>
                ) : null}

                {messages.map((message) => {
                  const ownMessage = user && Number(user.id) === Number(message.userId);
                  return (
                    <Group key={message.id} align="flex-start" wrap="nowrap" justify={ownMessage ? "flex-end" : "flex-start"}>
                      {!ownMessage ? <Avatar radius="xl" color="cyan">{initials(message.username)}</Avatar> : null}
                      <Paper p="sm" radius="lg" bg={ownMessage ? "cyan.9" : "dark.6"} style={{ maxWidth: "78%" }}>
                        <Group justify="space-between" gap="sm" align="flex-start">
                          <Text
                            size="sm"
                            fw={800}
                            className="user-link"
                            c={ownMessage ? "white" : "cyan.3"}
                            onClick={() => openUserProfile(message.userId)}
                          >
                            {message.username}
                          </Text>
                          <Text size="xs" c={ownMessage ? "rgba(255,255,255,0.7)" : "dimmed"}>
                            {message.timestamp}
                          </Text>
                        </Group>
                        <Text size="sm" mt={6}>{message.message}</Text>
                      </Paper>
                      {ownMessage ? <Avatar radius="xl" color="green">{initials(message.username)}</Avatar> : null}
                    </Group>
                  );
                })}
              </Stack>
            </ScrollArea>
          </Paper>

          <Group align="end" wrap="nowrap">
            <TextInput
              flex={1}
              label="Votre message"
              placeholder={user ? "Analyse, pronostic, feeling..." : "Connectez-vous pour écrire"}
              value={newMessage}
              onChange={(event) => setNewMessage(event.currentTarget.value)}
              disabled={!user || sending}
              onKeyDown={(event) => {
                if (event.key === "Enter") {
                  event.preventDefault();
                  handleSend();
                }
              }}
            />
            <Button color="cyan" onClick={handleSend} loading={sending} disabled={!user || !newMessage.trim()}>
              Envoyer
            </Button>
          </Group>
        </Stack>
      </Card>

      <Modal
        opened={profileOpened}
        onClose={() => setProfileOpened(false)}
        title="Profil du supporter"
        centered
        radius="xl"
      >
        {profileLoading ? <Loader color="cyan" size="sm" /> : null}
        {!profileLoading && selectedUserProfile ? (
          <Stack gap="md">
            <Group wrap="nowrap">
              <Avatar radius="xl" color="cyan" size={56}>{initials(selectedUserProfile.username)}</Avatar>
              <div>
                <Title order={4}>{selectedUserProfile.username}</Title>
                <Text c="dimmed" size="sm">Rang #{selectedUserProfile.rank || "-"}</Text>
              </div>
            </Group>

            <Group grow>
              <Paper className="kpi-card" p="md">
                <Text size="sm" c="dimmed">Score</Text>
                <Text fw={800} size="xl">{selectedUserProfile.score}</Text>
              </Paper>
              <Paper className="kpi-card" p="md">
                <Text size="sm" c="dimmed">Prédictions</Text>
                <Text fw={800} size="xl">{selectedUserProfile.totalPredictions}</Text>
              </Paper>
            </Group>

            <Group grow>
              <Paper className="kpi-card" p="md">
                <Text size="sm" c="dimmed">Réussite</Text>
                <Text fw={800} size="xl">{selectedUserProfile.accuracy.toFixed(0)}%</Text>
              </Paper>
              <Paper className="kpi-card" p="md">
                <Text size="sm" c="dimmed">Messages chat</Text>
                <Text fw={800} size="xl">{selectedUserProfile.chatMessages}</Text>
              </Paper>
            </Group>
          </Stack>
        ) : null}
      </Modal>
    </>
  );
}
