export function ResultsList({ results }) {
  const total = results.reduce((sum, r) => sum + r.count, 0);

  return (
    <ul className="results-list">
      {results.map((r) => {
        const pct = total > 0 ? Math.round((r.count / total) * 100) : 0;
        return (
          <li key={r.id} className="results-row">
            <div className="results-label">
              <span>{r.text}</span>
              <span>
                {r.count} vote{r.count === 1 ? "" : "s"} ({pct}%)
              </span>
            </div>
            <div className="results-bar-track">
              <div className="results-bar-fill" style={{ width: `${pct}%` }} />
            </div>
          </li>
        );
      })}
    </ul>
  );
}