import type { Airport } from "../types";

const PLANE_PATH =
  "M21 16v-2l-8-5V3.5c0-.83-.67-1.5-1.5-1.5S10 2.67 10 3.5V9l-8 5v2l8-2.5V19l-2.5 1.5V22l4-1 4 1v-1.5L13 19v-5.5l8 2.5z";
const ICON_SIZE = 24;
const ICON_SCALE = 0.5;

interface AirportMarkerProps {
  airport: Airport;
}

/** Renders an airport as a static outline plane icon with its code labeled alongside. */
export function AirportMarker({ airport }: AirportMarkerProps) {
  const { x, y, code } = airport;
  const halfIconUnscaled = ICON_SIZE / 2;

  return (
    <g className="airport" data-testid={`airport-${code}`}>
      <path
        d={PLANE_PATH}
        className="airport-icon"
        transform={`translate(${x}, ${y}) scale(${ICON_SCALE}) translate(${-halfIconUnscaled}, ${-halfIconUnscaled})`}
      />
      <text x={x + 8} y={y - 8} className="airport-label">
        {code}
      </text>
    </g>
  );
}
