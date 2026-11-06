// web/src/components/FlightMap.test.tsx
import { describe, expect, it } from "vitest";
import { render, screen } from "@testing-library/react";
import { FlightMap } from "./FlightMap";
import type { Airport, Flight } from "../types";

const sfo: Airport = { code: "SFO", name: "San Francisco Intl", city: "San Francisco", x: 100, y: 200, gates: [] };
const jfk: Airport = { code: "JFK", name: "John F Kennedy Intl", city: "New York", x: 900, y: 150, gates: [] };

function flight(overrides: Partial<Flight> = {}): Flight {
  return {
    id: "fl-1",
    flightNumber: "UA123",
    originCode: "SFO",
    destinationCode: "JFK",
    aircraftId: "ac-1",
    status: "airborne",
    scheduledDeparture: "2026-01-01T00:00:00Z",
    scheduledArrival: "2026-01-01T05:00:00Z",
    ...overrides,
  };
}

describe("FlightMap", () => {
  it("renders a marker for every airport and every flight", () => {
    const flights = new Map([["fl-1", flight()]]);
    render(<FlightMap airports={[sfo, jfk]} flights={flights} />);

    expect(screen.getByTestId("airport-SFO")).toBeInTheDocument();
    expect(screen.getByTestId("airport-JFK")).toBeInTheDocument();
    expect(screen.getByTestId("flight-fl-1")).toBeInTheDocument();
  });

  it("falls back to the origin airport's position when a flight has no position", () => {
    const flights = new Map([["fl-1", flight({ position: undefined })]]);
    render(<FlightMap airports={[sfo, jfk]} flights={flights} />);

    const marker = screen.getByTestId("flight-fl-1");
    expect(marker.getAttribute("transform")).toContain(`translate(${sfo.x}, ${sfo.y})`);
  });

  it("falls back to the destination airport's position when a landed flight has no position", () => {
    const flights = new Map([["fl-1", flight({ position: undefined, status: "landed" })]]);
    render(<FlightMap airports={[sfo, jfk]} flights={flights} />);

    const marker = screen.getByTestId("flight-fl-1");
    expect(marker.getAttribute("transform")).toContain(`translate(${jfk.x}, ${jfk.y})`);
  });

  it("rotates the flight marker to match its heading", () => {
    const flights = new Map([
      ["fl-1", flight({ position: { x: 500, y: 300, heading: 135, progress: 0.5 } })],
    ]);
    render(<FlightMap airports={[sfo, jfk]} flights={flights} />);

    const marker = screen.getByTestId("flight-fl-1");
    expect(marker.getAttribute("transform")).toContain("rotate(135)");
  });

  it("renders nothing when a flight has neither a position nor a resolvable origin airport", () => {
    const flights = new Map([["fl-1", flight({ position: undefined, originCode: "ZZZ" })]]);
    render(<FlightMap airports={[sfo, jfk]} flights={flights} />);

    expect(screen.queryByTestId("flight-fl-1")).not.toBeInTheDocument();
  });

  it("applies the flight's status as a CSS class", () => {
    const flights = new Map([["fl-1", flight({ status: "landed" })]]);
    render(<FlightMap airports={[sfo, jfk]} flights={flights} />);

    expect(screen.getByTestId("flight-fl-1")).toHaveClass("flight", "landed");
  });
});
