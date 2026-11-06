import type { Meta } from "../types";
import { VersionBadge } from "./VersionBadge";
import { StatsBar } from "./StatsBar";
import { ColorLegend } from "./ColorLegend";

interface HeaderProps {
  meta: Meta | null;
  eventsPerSecond: number;
  activeFlights: number;
}

/** Renders the application header with title, version badge, color legend, and live statistics. */
export function Header({ meta, eventsPerSecond, activeFlights }: HeaderProps) {
  return (
    <header>
      <div>
        <h1>
          SKYGATE <VersionBadge meta={meta} />
        </h1>
        <ColorLegend />
      </div>
      <StatsBar eventsPerSecond={eventsPerSecond} activeFlights={activeFlights} />
    </header>
  );
}
