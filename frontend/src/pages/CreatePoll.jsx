import { useState } from "react";
import { Link } from "react-router-dom";
import { api } from "../api/client";
import { Select } from "../components/Select";
import { DateTimePicker } from "../components/DateTimePicker";
import { useCountdown } from "../utils/time";

const DURATION_OPTIONS = [
	{ label: "No expiry", value: "" },
	{ label: "1 hour", value: "60" },
	{ label: "3 hours", value: "180" },
	{ label: "12 hours", value: "720" },
	{ label: "1 day", value: "1440" },
	{ label: "7 days", value: "10080" },
	{ label: "Custom", value: "custom" },
];

export default function CreatePoll() {
	const [title, setTitle] = useState("");
	const [options, setOptions] = useState(["", ""]);
	const [durationChoice, setDurationChoice] = useState("");
	const [customDate, setCustomDate] = useState(null);
	const [error, setError] = useState("");
	const [loading, setLoading] = useState(false);
	const [createdPoll, setCreatedPoll] = useState(null);
	const [copied, setCopied] = useState(false);
	const [resultVisibility, setResultVisibility] = useState("always");
	const closesText = useCountdown(createdPoll?.expires_at);
	const minDate = new Date(Date.now() + 5 * 60000);

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

		let durationMinutes = 0;
		if (durationChoice === "custom") {
			if (!customDate) {
				setError("Pick a close date and time.");
				return;
			}
			durationMinutes = Math.round((customDate.getTime() - Date.now()) / 60000);
			if (durationMinutes < 5) {
				setError("Custom close time must be at least 5 minutes from now.");
				return;
			}
		} else if (durationChoice) {
			durationMinutes = parseInt(durationChoice, 10);
		}

		setLoading(true);
		try {
			const poll = await api.createPoll({
				title: title.trim(),
				options: cleanedOptions,
				result_visibility: resultVisibility,
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
		setDurationChoice("");
		setCustomDate(null);
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

				<Select label="Voting window" value={durationChoice} onChange={setDurationChoice} options={DURATION_OPTIONS} />

				{durationChoice === "custom" && (
					<label>
						Custom close date &amp; time
						<DateTimePicker value={customDate} onChange={setCustomDate} minDate={minDate} />
					</label>
				)}
				<Select
					label="Results visibility"
					value={resultVisibility}
					onChange={setResultVisibility}
					options={[
						{ value: "always", label: "Always public" },
						{ value: "after_end", label: "Public after end date" },
						{ value: "after_voting", label: "Public after voting" },
						{ value: "private", label: "Not public" },
					]}
				/>
				<button type="submit" disabled={loading}>
					{loading ? "Creating..." : "Create poll"}
				</button>
			</form>
		</div>
	);
}
