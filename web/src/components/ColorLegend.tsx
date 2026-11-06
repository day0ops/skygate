const GROUPS = [
  { label: "GROUND / HOLDING", color: "var(--status-amber)" },
  { label: "IN FLIGHT", color: "var(--status-blue)" },
  { label: "LANDED", color: "var(--status-landed)" },
  { label: "CANCELLED", color: "var(--status-cancelled)" },
];

/** Explains what each flight marker color on the map means. */
export function ColorLegend() {
  return (
    <ul className="legend">
      {GROUPS.map((group) => (
        <li key={group.label}>
          <span className="legend-swatch" style={{ backgroundColor: group.color }} />
          {group.label}
        </li>
      ))}
    </ul>
  );
}
