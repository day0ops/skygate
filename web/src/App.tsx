import { useAirports } from "./hooks/useAirports";
import { useMeta } from "./hooks/useMeta";
import { useFlightSocket } from "./hooks/useFlightSocket";
import { Header } from "./components/Header";
import { FlightMap } from "./components/FlightMap";
import { ActivityLog } from "./components/ActivityLog";

/** Composes the header, flight map, and activity log into the skygate dashboard. */
export function App() {
  const meta = useMeta();
  const airports = useAirports();
  const { flights, log, eventsPerSecond } = useFlightSocket();

  return (
    <>
      <Header meta={meta} eventsPerSecond={eventsPerSecond} activeFlights={flights.size} />
      <main>
        <FlightMap airports={airports} flights={flights} />
        <ActivityLog entries={log} />
      </main>
    </>
  );
}
