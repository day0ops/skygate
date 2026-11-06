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
      {Array.from(flights.values()).map((flight) => {
        // Landed flights have arrived at their destination, not their origin -
        // falling back to origin here previously made every landed flight
        // visibly snap backward across the map while it waited to be removed.
        const fallbackCode = flight.status === "landed" ? flight.destinationCode : flight.originCode;
        return (
          <FlightMarker key={flight.id} flight={flight} fallbackPosition={positionByCode.get(fallbackCode)} />
        );
      })}
    </svg>
  );
}
