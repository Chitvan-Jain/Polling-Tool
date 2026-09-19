import { useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import { api } from "../api/client";
import { ResultsList } from "../components/ResultsList";
import { useLiveResults } from "../hooks/useLiveResults";

export default function VotePage() {
  const { slug } = useParams();
  const [poll, setPoll] = useState(null);
  const [loadError, setLoadError] = useState("");
  const [selectedOption, setSelectedOption] = useState(null);
  const [submitting, setSubmitting] = useState(false);
  const [voteError, setVoteError] = useState("");
  const [hasVoted, setHasVoted] = useState(false);
  const [alreadyVoted, setAlreadyVoted] = useState(false);

  const { data: liveResults, connectionState } = useLiveResults(poll ? slug : null);

  useEffect(() => {
    let cancelled = false;
    api
      .getPollBySlug(slug)
      .then((data) => {
        if (cancelled) return;
        setPoll(data);
        if (data.has_voted) {
          setHasVoted(true);
          setAlreadyVoted(true);
        }
      })
      .catch((err) => {
        if (!cancelled) setLoadError(err.message);
      });
    return () => {
      cancelled = true;
    };
  }, [slug]);

  async function handleVote(e) {
    e.preventDefault();
    if (!selectedOption) return;
    setSubmitting(true);
    setVoteError("");
    try {
      await api.vote(slug, selectedOption);
      setHasVoted(true);
    } catch (err) {
      if (err.status === 409) {
        setAlreadyVoted(true);
        setHasVoted(true);
      } else {
        setVoteError(err.message);
      }
    } finally {
      setSubmitting(false);
    }
  }

  if (loadError) {
    return (
      <div className="page">
        <h1>Poll not found</h1>
        <p>This link doesn't match an active poll. Double-check the URL your host shared with you.</p>
      </div>
    );
  }

  if (!poll) {
    return (
      <div className="page">
        <p>Loading poll…</p>
      </div>
    );
  }

  const showResults = hasVoted || poll.status !== "open";

  return (
    <div className="page">
      {connectionState === "reconnecting" && (
        <div className="connection-banner">Reconnecting to live results…</div>
      )}

      {showResults ? (
        <>
          <h1>{liveResults?.title ?? poll.title}</h1>
          {alreadyVoted && <p>You've already voted on this poll — here's how it's going.</p>}
          {!alreadyVoted && hasVoted && <p>Thanks for voting! Watch the results update live.</p>}
          {!hasVoted && poll.status !== "open" && <p>This poll is closed.</p>}
          {liveResults ? (
            <ResultsList results={liveResults.results} />
          ) : (
            <p>Loading live results…</p>
          )}
        </>
      ) : (
        <>
          <h1>{poll.title}</h1>
          <form onSubmit={handleVote} className="vote-form">
            {voteError && <p className="form-error">{voteError}</p>}
            <fieldset>
              {poll.options.map((opt) => (
                <label key={opt.id} className="vote-option">
                  <input
                    type="radio"
                    name="option"
                    value={opt.id}
                    checked={selectedOption === opt.id}
                    onChange={() => setSelectedOption(opt.id)}
                  />
                  {opt.text}
                </label>
              ))}
            </fieldset>
            <button type="submit" disabled={!selectedOption || submitting}>
              {submitting ? "Submitting..." : "Vote"}
            </button>
          </form>
        </>
      )}
    </div>
  );
}