import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { useAuth } from "../context/AuthContext";
import { api } from "../api/client";

export default function Dashboard() {
	const { logout } = useAuth();
	const [polls, setPolls] = useState(null);
	const [error, setError] = useState("");
	const [copiedSlug, setCopiedSlug] = useState(null);

	useEffect(() => {
		api
			.getMyPolls()
			.then(setPolls)
			.catch((err) => setError(err.message));
	}, []);

	function copyLink(slug) {
		const url = `${window.location.origin}/vote/${slug}`;
		navigator.clipboard.writeText(url);
		setCopiedSlug(slug);
		setTimeout(() => setCopiedSlug(null), 2000);
	}

	return (
		<div className="page page-wide">
			<div className="page-header">
				<h1>Your polls</h1>
				<div className="page-header-actions">
					<Link to="/dashboard/new" className="header-button">
						+ Create poll
					</Link>
					<button onClick={logout} className="header-button-secondary">
						Log out
					</button>
				</div>
			</div>

			{error && <p className="form-error">{error}</p>}
			{!polls && !error && <p>Loading your polls…</p>}

			{polls && polls.length === 0 && (
				<div className="empty-state">
					<p>You haven't created any polls yet.</p>
					<Link to="/dashboard/new" className="header-button">
						Create your first poll
					</Link>
				</div>
			)}

			{polls && polls.length > 0 && (
				<ul className="poll-list">
					{polls.map((poll) => (
						<li key={poll.id} className="poll-list-item">
							<div className="poll-list-main">
								<h3>{poll.title}</h3>
								<div className="poll-meta">
									<span className={`status-badge status-${poll.status}`}>{poll.status}</span>
									<span>
										{poll.total_votes} vote{poll.total_votes === 1 ? "" : "s"}
									</span>
									<span>{new Date(poll.created_at).toLocaleDateString()}</span>
									{poll.expires_at && <span>Closes {new Date(poll.expires_at).toLocaleString()}</span>}
								</div>
							</div>
							<div className="poll-list-actions">
								<button onClick={() => copyLink(poll.share_slug)}>{copiedSlug === poll.share_slug ? "Copied!" : "Copy link"}</button>
								<Link to={`/vote/${poll.share_slug}`} className="header-link">
									View
								</Link>
							</div>
						</li>
					))}
				</ul>
			)}
		</div>
	);
}
