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

function Home() {
  return (
    <div>
      <div className="text-center mb-10">
        <h1 className="text-4xl font-bold text-gray-800">
          Bienvenue sur Footix
        </h1>
        <p className="text-gray-600 mt-2">
          Pronostiquez les matchs de football et suivez vos performances
        </p>
      </div>
      <Leagues />
    </div>
  );
}

function App() {
  const location = useLocation();

  return (
    <AuthProvider>
      <div className="min-h-screen bg-gradient-to-br from-gray-50 to-gray-100">
        <Navbar />
        <div className="max-w-7xl mx-auto px-4 py-2">
          <Breadcrumb />
        </div>
        <main className="max-w-7xl mx-auto px-4 py-8">
          <Routes>
            <Route path="/" element={<Home />} />
            <Route
              path="/league/:leagueId"
              element={<MatchList key={location.pathname} />}
            />
            <Route path="/match/:matchId" element={<MatchDetail />} />
            <Route path="/login" element={<Login />} />
            <Route path="/signup" element={<Signup />} />
            <Route path="/profile" element={<Profile />} />
            <Route path="/logout" element={<Navigate to="/" />} />
          </Routes>
        </main>
      </div>
    </AuthProvider>
  );
}

export default App;
