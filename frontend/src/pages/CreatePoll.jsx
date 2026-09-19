import { useState } from "react";
import { useAuth } from "../context/AuthContext";
import { api } from "../api/client";

export default function CreatePoll() {
  const { logout } = useAuth();
  const [title, setTitle] = useState("");
  const [options, setOptions] = useState(["", ""]);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const [createdPoll, setCreatedPoll] = useState(null);
  const [copied, setCopied] = useState(false);

  function updateOption(index, value) {
    setOptions((prev) => prev.map((opt, i) => (i === index ? value : opt)));
  }

  function addOption() {
    if (options.length >= 10) return;
    setOptions((prev) => [...prev, ""]);
  }

  function removeOption(index) {
    if (options.length <= 2) return;
    setOptions((prev) => prev.filter((_, i) => i !== index));
  }

  async function handleSubmit(e) {
    e.preventDefault();
    setError("");

    const cleanedOptions = options.map((o) => o.trim()).filter(Boolean);
    if (cleanedOptions.length < 2) {
      setError("Add at least two options.");
      return;
    }

    setLoading(true);
    try {
      const poll = await api.createPoll({ title: title.trim(), options: cleanedOptions });
      setCreatedPoll(poll);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }

  function startNewPoll() {
    setCreatedPoll(null);
    setTitle("");
    setOptions(["", ""]);
    setCopied(false);
  }

  if (createdPoll) {
    const shareUrl = `${window.location.origin}/vote/${createdPoll.share_slug}`;
    return (
      <div className="page">
        <h1>Your poll is live</h1>
        <p>{createdPoll.title}</p>
        <div className="share-link-row">
          <input readOnly value={shareUrl} onFocus={(e) => e.target.select()} />
          <button
            onClick={() => {
              navigator.clipboard.writeText(shareUrl);
              setCopied(true);
              setTimeout(() => setCopied(false), 2000);
            }}
          >
            {copied ? "Copied!" : "Copy link"}
          </button>
        </div>
        <p>Share this link with your audience so they can vote.</p>
        <button onClick={startNewPoll}>Create another poll</button>
      </div>
    );
  }

  return (
    <div className="page">
      <div className="page-header">
        <h1>Create a poll</h1>
        <button onClick={logout}>Log out</button>
      </div>
      <form onSubmit={handleSubmit} className="create-poll-form">
        {error && <p className="form-error">{error}</p>}
        <label>
          Question
          <input value={title} onChange={(e) => setTitle(e.target.value)} required minLength={3} maxLength={200} />
        </label>

        <fieldset>
          <legend>Options</legend>
          {options.map((opt, i) => (
            <div key={i} className="option-row">
              <input
                value={opt}
                onChange={(e) => updateOption(i, e.target.value)}
                placeholder={`Option ${i + 1}`}
                required
              />
              {options.length > 2 && (
                <button type="button" className="remove-option-button" onClick={() => removeOption(i)} aria-label={`Remove option ${i + 1}`}>
  ✕
</button>
              )}
            </div>
          ))}
          {options.length < 10 && (
            <button type="button" className="add-option-button" onClick={addOption}>+ Add option</button>
          )}
        </fieldset>

        <button type="submit" disabled={loading}>
          {loading ? "Creating..." : "Create poll"}
        </button>
      </form>
    </div>
  );
}