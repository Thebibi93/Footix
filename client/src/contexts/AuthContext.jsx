import { createContext, useContext, useState, useEffect } from "react";

const AuthContext = createContext();

export const useAuth = () => useContext(AuthContext);

export const AuthProvider = ({ children }) => {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // Check localStorage pr le moment...
    const storedUser = localStorage.getItem("footix_user");
    if (storedUser) {
      setUser(JSON.parse(storedUser));
    }
    setLoading(false);
  }, []);

  const login = (email, password) => {
    // On mettra après le vrai login pour le moment c'est du mock
    if (!email.includes("@")) {
      throw new Error("Invalid credentials");
    }
    const newUser = { id: 1, email, name: email.split("@")[0] };
    setUser(newUser);
    localStorage.setItem("footix_user", JSON.stringify(newUser));
    return newUser;
  };

  const signup = (email, password, name) => {
    const newUser = {
      id: Date.now(),
      email,
      name: name || email.split("@")[0],
    };
    setUser(newUser);
    localStorage.setItem("footix_user", JSON.stringify(newUser));
    return newUser;
  };

  const logout = () => {
    setUser(null);
    localStorage.removeItem("footix_user");
  };

  return (
    <AuthContext.Provider value={{ user, loading, login, signup, logout }}>
      {children}
    </AuthContext.Provider>
  );
};
