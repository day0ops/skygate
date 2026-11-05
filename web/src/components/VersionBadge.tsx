import type { CSSProperties } from "react";
import type { Meta } from "../types";

const COLOR_HEX: Record<string, string> = {
  green: "#4af626",
  red: "var(--accent)",
};
const DEFAULT_COLOR_HEX = "#3b82f6";

interface VersionBadgeProps {
  meta: Meta | null;
}

/** Displays the version and color from Meta, or a loading state if meta is null. */
export function VersionBadge({ meta }: VersionBadgeProps) {
  const label = meta ? `[ ${meta.version} · ${meta.color.toUpperCase()} ]` : "[ LOADING… ]";
  const color = meta ? (COLOR_HEX[meta.color] ?? DEFAULT_COLOR_HEX) : DEFAULT_COLOR_HEX;
  const style = { "--badge-color": color } as CSSProperties;

  return (
    <span className="badge" style={style}>
      {label}
    </span>
  );
}
