import type { Airport, Flight } from "../types";
import { AirportMarker } from "./AirportMarker";
import { FlightMarker } from "./FlightMarker";

interface FlightMapProps {
  airports: Airport[];
  flights: Map<string, Flight>;
}

/** Renders the SVG canvas of airport and flight markers. */
export function FlightMap({ airports, flights }: FlightMapProps) {
  const positionByCode = new Map(airports.map((a) => [a.code, { x: a.x, y: a.y }]));

  return (
    <svg id="map" viewBox="0 0 1600 650" xmlns="http://www.w3.org/2000/svg">
      {airports.map((airport) => (
        <AirportMarker key={airport.code} airport={airport} />
      ))}
      {Array.from(flights.values()).map((flight) => (
        <FlightMarker key={flight.id} flight={flight} fallbackPosition={positionByCode.get(flight.originCode)} />
      ))}
    </svg>
  );
}
