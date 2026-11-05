import type { Flight, Position } from "../types";

const PLANE_PATH =
  "M21 16v-2l-8-5V3.5c0-.83-.67-1.5-1.5-1.5S10 2.67 10 3.5V9l-8 5v2l8-2.5V19l-2.5 1.5V22l4-1 4 1v-1.5L13 19v-5.5l8 2.5z";
const PLANE_ICON_SIZE = 24;
const PLANE_SCALE = 0.6;

interface FlightMarkerProps {
  flight: Flight;
  fallbackPosition?: Pick<Position, "x" | "y">;
}

/**
 * Renders a flight as a plane icon rotated to its heading, using
 * fallbackPosition when the flight has no live position. Renders nothing
 * if neither a position nor a fallback is available.
 */
export function FlightMarker({ flight, fallbackPosition }: FlightMarkerProps) {
  const pos = flight.position ?? fallbackPosition;
  if (!pos) return null;

  const heading = flight.position?.heading ?? 0;
  const halfIconUnscaled = PLANE_ICON_SIZE / 2;

  return (
    <g
      className={`flight ${flight.status}`}
      transform={`translate(${pos.x}, ${pos.y}) rotate(${heading}) scale(${PLANE_SCALE}) translate(${-halfIconUnscaled}, ${-halfIconUnscaled})`}
      data-testid={`flight-${flight.id}`}
    >
      <path d={PLANE_PATH} />
    </g>
  );
}
