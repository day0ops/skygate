interface StatsBarProps {
  eventsPerSecond: number;
  activeFlights: number;
}

/** Displays the current event throughput and active flight count. */
export function StatsBar({ eventsPerSecond, activeFlights }: StatsBarProps) {
  return (
    <div className="stats">
      EVENTS/SEC: <span>{eventsPerSecond}</span> · ACTIVE FLIGHTS: <span>{activeFlights}</span>
    </div>
  );
}
