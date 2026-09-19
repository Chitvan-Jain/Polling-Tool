import { Link } from "react-router-dom";
import { useAuth } from "../context/AuthContext";

export function Header() {
  const { isAuthenticated, logout } = useAuth();

  return (
    <header className="site-header">
      <Link to="/" className="brand">Pulse</Link>
      <nav className="header-nav">
        {isAuthenticated ? (
          <>
            <Link to="/dashboard" className="header-link">Dashboard</Link>
            <button onClick={logout} className="header-button-secondary">Log out</button>
          </>
        ) : (
          <>
            <Link to="/login" className="header-link">Log in</Link>
            <Link to="/signup" className="header-button">Sign up</Link>
          </>
        )}
      </nav>
    </header>
  );
}