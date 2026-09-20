import { useState } from "react";
import { Link } from "react-router-dom";
import { api } from "../api/client";
import { useCountdown } from "../utils/time";

export function PollListItem({ poll, onDeleted }) {
  const [copied, setCopied] = useState(false);
  const [deleting, setDeleting] = useState(false);
  const closesText = useCountdown(poll.expires_at);

  function copyLink() {
    const url = `${window.location.origin}/vote/${poll.share_slug}`;
    navigator.clipboard.writeText(url);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  }

  async function handleDelete() {
    const confirmed = window.confirm(`Delete "${poll.title}"? This can't be undone.`);
    if (!confirmed) return;
    setDeleting(true);
    try {
      await api.deletePoll(poll.id);
      onDeleted(poll.id);
    } catch (err) {
      alert(err.message);
      setDeleting(false);
    }
  }

  return (
    <li className="poll-list-item">
      <div className="poll-list-main">
        <h3>{poll.title}</h3>
        <div className="poll-meta">
          <span className={`status-badge status-${poll.status}`}>{poll.status}</span>
          <span>{poll.total_votes} vote{poll.total_votes === 1 ? "" : "s"}</span>
          <span>{new Date(poll.created_at).toLocaleDateString()}</span>
          {closesText && <span>{closesText}</span>}
        </div>
      </div>
      <div className="poll-list-actions">
        <button onClick={copyLink}>{copied ? "Copied!" : "Copy link"}</button>
        <Link to={`/vote/${poll.share_slug}`} className="header-link">View</Link>
        <button onClick={handleDelete} disabled={deleting} className="delete-button">
          {deleting ? "Deleting..." : "Delete"}
        </button>
      </div>
    </li>
  );
}