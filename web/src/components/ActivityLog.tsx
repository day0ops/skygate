import type { LogEntry } from "../types";

interface ActivityLogProps {
  entries: LogEntry[];
}

/** Renders a list of log entries in the given order. */
export function ActivityLog({ entries }: ActivityLogProps) {
  return (
    <aside id="log">
      <h2>[ LIVE ACTIVITY ]</h2>
      <ul id="log-list">
        {entries.map((entry) => (
          <li key={entry.id}>{entry.text}</li>
        ))}
      </ul>
    </aside>
  );
}
