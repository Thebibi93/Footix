import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";

const AuthContext = createContext();
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? "http://localhost:8080";

// useAuth expose le contexte d’authentification aux composants React.
export const useAuth = () => useContext(AuthContext);

// authFetch centralise les appels authentifiés vers l’API Go.
async function authFetch(path, options = {}) {
  const headers = new Headers(options.headers || {});
  const isJsonBody =
    options.body !== undefined && options.body !== null && !(options.body instanceof FormData);

  if (isJsonBody && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }

  const response = await fetch(`${API_BASE_URL}${path}`, {
    credentials: "include",
    ...options,
    headers,
    body: isJsonBody && typeof options.body !== "string"
      ? JSON.stringify(options.body)
      : options.body,
  });

  const text = await response.text();
  let data = null;
  try {
    data = text ? JSON.parse(text) : null;
  } catch {
    data = null;
  }

  if (!response.ok) {
    throw new Error(data?.error || data?.message || "Erreur serveur");
  }

  return data;
}

// normalizeUser uniformise le profil reçu du serveur avant stockage côté client.
function normalizeUser(rawUser) {
  if (!rawUser) {
    return null;
  }

  return {
    id: rawUser.id,
    username: rawUser.username,
    email: rawUser.email,
    score: rawUser.score ?? 0,
    rank: rawUser.rank ?? 0,
    totalPredictions: rawUser.totalPredictions ?? 0,
    correctPredictions: rawUser.correctPredictions ?? 0,
    accuracy: rawUser.accuracy ?? 0,
    chatMessages: rawUser.chatMessages ?? 0,
  };
}

// AuthProvider maintient la session utilisateur et fournit les actions d’authentification.
export const AuthProvider = ({ children }) => {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);

  // refreshProfile recharge le profil courant depuis l’API.
  const refreshProfile = useCallback(async () => {
    try {
      const profile = await authFetch("/api/profile", { method: "GET" });
      const normalized = normalizeUser(profile);
      setUser(normalized);
      return normalized;
    } catch {
      setUser(null);
      return null;
    }
  }, []);

  useEffect(() => {
    let active = true;

    (async () => {
      try {
        const profile = await authFetch("/api/profile", { method: "GET" });
        if (active) {
          setUser(normalizeUser(profile));
        }
      } catch {
        if (active) {
          setUser(null);
        }
      } finally {
        if (active) {
          setLoading(false);
        }
      }
    })();

    return () => {
      active = false;
    };
  }, []);

  // login authentifie l’utilisateur puis synchronise son profil local.
  const login = useCallback(async (identifier, password) => {
    await authFetch("/api/login", {
      method: "POST",
      body: {
        username: identifier,
        password,
      },
    });

    const profile = await authFetch("/api/profile", { method: "GET" });
    const normalized = normalizeUser(profile);
    setUser(normalized);
    return normalized;
  }, []);

  // signup crée un compte puis synchronise le profil local.
  const signup = useCallback(async (username, email, password) => {
    await authFetch("/api/signup", {
      method: "POST",
      body: {
        username,
        email,
        password,
      },
    });

    const profile = await authFetch("/api/profile", { method: "GET" });
    const normalized = normalizeUser(profile);
    setUser(normalized);
    return normalized;
  }, []);

  // logout ferme la session serveur puis nettoie l’état utilisateur local.
  const logout = useCallback(async () => {
    try {
      await authFetch("/api/logout", { method: "POST" });
    } finally {
      setUser(null);
    }
  }, []);

  const value = useMemo(
    () => ({ user, loading, login, signup, logout, refreshProfile }),
    [user, loading, login, signup, logout, refreshProfile],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};
