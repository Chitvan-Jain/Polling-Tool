import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { useAuth } from "../context/AuthContext";
import { api } from "../api/client";
import { PollListItem } from "../components/PollListItem";

export default function Dashboard() {
  const { logout } = useAuth();
  const [polls, setPolls] = useState(null);
  const [error, setError] = useState("");

  useEffect(() => {
    api.getMyPolls().then(setPolls).catch((err) => setError(err.message));
  }, []);

  function handleDeleted(id) {
    setPolls((prev) => prev.filter((p) => p.id !== id));
  }

  return (
    <div className="page page-wide">
      <div className="page-header">
        <h1>Your polls</h1>
        <div className="page-header-actions">
          <Link to="/dashboard/new" className="header-button">+ Create poll</Link>
          <button onClick={logout} className="header-button-secondary">Log out</button>
        </div>
      </div>

      {error && <p className="form-error">{error}</p>}
      {!polls && !error && <p>Loading your polls…</p>}

      {polls && polls.length === 0 && (
        <div className="empty-state">
          <p>You haven't created any polls yet.</p>
          <Link to="/dashboard/new" className="header-button">Create your first poll</Link>
        </div>
      )}

      {polls && polls.length > 0 && (
        <ul className="poll-list">
          {polls.map((poll) => (
            <PollListItem key={poll.id} poll={poll} onDeleted={handleDeleted} />
          ))}
        </ul>
      )}
    </div>
  );
}