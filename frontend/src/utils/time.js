import { useEffect, useState } from "react";

export function formatDuration(ms) {
  if (ms <= 0) return null;
  if (ms < 60000) return "less than a minute";

  const totalMinutes = Math.floor(ms / 60000);
  const days = Math.floor(totalMinutes / (60 * 24));
  const hours = Math.floor((totalMinutes % (60 * 24)) / 60);
  const minutes = totalMinutes % 60;

  const parts = [];
  if (days > 0) parts.push(`${days} day${days === 1 ? "" : "s"}`);
  if (hours > 0) parts.push(`${hours} hour${hours === 1 ? "" : "s"}`);
  if (minutes > 0) parts.push(`${minutes} minute${minutes === 1 ? "" : "s"}`);

  return parts.join(", ");
}

export function useCountdown(targetIso) {
  const [now, setNow] = useState(Date.now());

  useEffect(() => {
    if (!targetIso) return;
    const interval = setInterval(() => setNow(Date.now()), 30000);
    return () => clearInterval(interval);
  }, [targetIso]);

  if (!targetIso) return null;

  const diff = new Date(targetIso).getTime() - now;
  if (diff <= 0) return "Closed";

  return `Closes in ${formatDuration(diff)}`;
}