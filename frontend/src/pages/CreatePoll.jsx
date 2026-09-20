import { useState } from "react";
import { api } from "../api/client";
import { Link } from "react-router-dom";
import { useCountdown } from "../utils/time";
export default function CreatePoll() {
	const [title, setTitle] = useState("");
	const [options, setOptions] = useState(["", ""]);
	const [error, setError] = useState("");
	const [loading, setLoading] = useState(false);
	const [createdPoll, setCreatedPoll] = useState(null);
	const [copied, setCopied] = useState(false);
	const [durationChoice, setDurationChoice] = useState("");
	const [customDateTime, setCustomDateTime] = useState("");
	const closesText = useCountdown(createdPoll?.expires_at);
	const minDateTimeLocal = toDatetimeLocalString(new Date(Date.now() + 5 * 60000));
	const DURATION_OPTIONS = [
		{ label: "No expiry", value: "" },
		{ label: "1 hour", value: "60" },
		{ label: "3 hours", value: "180" },
		{ label: "12 hours", value: "720" },
		{ label: "1 day", value: "1440" },
		{ label: "7 days", value: "10080" },
		{ label: "Custom", value: "custom" },
	];
	function updateOption(index, value) {
		setOptions((prev) => prev.map((opt, i) => (i === index ? value : opt)));
	}
	function toDatetimeLocalString(date) {
		const pad = (n) => String(n).padStart(2, "0");
		return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}`;
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
		let durationMinutes = 0;
		if (durationChoice === "custom") {
			if (!customDateTime) {
				setError("Pick a close date and time.");
				setLoading(false);
				return;
			}
			const target = new Date(customDateTime).getTime();
			durationMinutes = Math.round((target - Date.now()) / 60000);
			if (durationMinutes < 5) {
				setError("Custom close time must be at least 5 minutes from now.");
				setLoading(false);
				return;
			}
		} else if (durationChoice) {
			durationMinutes = parseInt(durationChoice, 10);
		}
		try {
			const poll = await api.createPoll({
				title: title.trim(),
				options: cleanedOptions,
				duration_minutes: durationMinutes,
			});
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
				{closesText && <p>{closesText}.</p>}
				<div className="button-row">
					<button onClick={startNewPoll}>Create another poll</button>
					<Link to="/dashboard" className="header-button-secondary">
						View all my polls
					</Link>
				</div>
			</div>
		);
	}

	return (
		<div className="page">
			<div className="page-header">
				<h1>Create a poll</h1>
				<Link to="/dashboard" className="header-link">
					← Back to dashboard
				</Link>
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
							<input value={opt} onChange={(e) => updateOption(i, e.target.value)} placeholder={`Option ${i + 1}`} required />
							{options.length > 2 && (
								<button type="button" className="remove-option-button" onClick={() => removeOption(i)} aria-label={`Remove option ${i + 1}`}>
									✕
								</button>
							)}
						</div>
					))}
					{options.length < 10 && (
						<button type="button" className="add-option-button" onClick={addOption}>
							+ Add option
						</button>
					)}
				</fieldset>
				<label>
					Voting window
					<select value={durationChoice} onChange={(e) => setDurationChoice(e.target.value)}>
						{DURATION_OPTIONS.map((opt) => (
							<option key={opt.value} value={opt.value}>
								{opt.label}
							</option>
						))}
					</select>
				</label>

				{durationChoice === "custom" && (
					<label>
						Custom close date & time
						<input type="datetime-local" value={customDateTime} min={minDateTimeLocal} onChange={(e) => setCustomDateTime(e.target.value)} />
					</label>
				)}
				<button type="submit" disabled={loading}>
					{loading ? "Creating..." : "Create poll"}
				</button>
			</form>
		</div>
	);
}
