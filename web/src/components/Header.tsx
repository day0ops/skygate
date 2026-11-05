import type { Meta } from "../types";
import { VersionBadge } from "./VersionBadge";
import { StatsBar } from "./StatsBar";

interface HeaderProps {
  meta: Meta | null;
  eventsPerSecond: number;
  activeFlights: number;
}

/** Renders the application header with title, version badge, and live statistics. */
export function Header({ meta, eventsPerSecond, activeFlights }: HeaderProps) {
  return (
    <header>
      <h1>
        SKYGATE <VersionBadge meta={meta} />
      </h1>
      <StatsBar eventsPerSecond={eventsPerSecond} activeFlights={activeFlights} />
    </header>
  );
}
