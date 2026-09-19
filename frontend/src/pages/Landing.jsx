import { Link } from "react-router-dom";
import { useAuth } from "../context/AuthContext";
import { Header } from "../components/Header";

export default function Landing() {
  const { isAuthenticated } = useAuth();

  return (
    <div>
      <Header />
      <section className="hero">
        <h1>Ask a question. Watch the room answer.</h1>
        <p className="hero-subtitle">
          Create a poll, share one link, and watch votes turn into live results as they come in —
          no refresh, no app to install.
        </p>
        <div className="hero-actions">
          <Link to={isAuthenticated ? "/dashboard" : "/signup"} className="hero-cta">
            {isAuthenticated ? "Go to your polls" : "Create a poll — it's free"}
          </Link>
          {!isAuthenticated && (
            <Link to="/login" className="hero-secondary">Log in</Link>
          )}
        </div>
      </section>

      <section className="how-it-works">
        <div className="how-step">
          <span className="how-step-number">1</span>
          <h3>Create your poll</h3>
          <p>Write a question, add up to 10 options, done in under a minute.</p>
        </div>
        <div className="how-step">
          <span className="how-step-number">2</span>
          <h3>Share the link</h3>
          <p>One link works for your whole audience — no accounts, no app downloads.</p>
        </div>
        <div className="how-step">
          <span className="how-step-number">3</span>
          <h3>Watch it live</h3>
          <p>Results update the instant a vote comes in, for everyone watching at once.</p>
        </div>
      </section>
    </div>
  );
}