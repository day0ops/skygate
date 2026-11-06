import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { act, renderHook } from "@testing-library/react";
import { useFlightSocket } from "./useFlightSocket";
import type { Flight, FlightUpdateMessage } from "../types";

class FakeWebSocket {
  static instances: FakeWebSocket[] = [];
  url: string;
  onmessage: ((event: { data: string }) => void) | null = null;
  onerror: ((event: unknown) => void) | null = null;
  onclose: (() => void) | null = null;
  closed = false;

  constructor(url: string) {
    this.url = url;
    FakeWebSocket.instances.push(this);
  }

  close() {
    this.closed = true;
  }

  emitMessage(msg: FlightUpdateMessage) {
    this.onmessage?.({ data: JSON.stringify(msg) });
  }

  emitClose() {
    this.onclose?.();
  }
}

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
    position: { x: 10, y: 20, heading: 90, progress: 0.5 },
    ...overrides,
  };
}

describe("useFlightSocket", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    FakeWebSocket.instances = [];
    vi.stubGlobal("WebSocket", FakeWebSocket);
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    vi.useRealTimers();
  });

  it("upserts a flight from a flight_update message", async () => {
    const { result } = renderHook(() => useFlightSocket());
    const socket = FakeWebSocket.instances[0];
    const f = flight();

    await act(async () => {
      socket.emitMessage({ type: "flight_update", flight: f });
    });

    expect(result.current.flights.get(f.id)).toEqual(f);
  });

  it("reconnects 2s after the socket closes", async () => {
    renderHook(() => useFlightSocket());
    expect(FakeWebSocket.instances).toHaveLength(1);

    await act(async () => {
      FakeWebSocket.instances[0].emitClose();
    });
    expect(FakeWebSocket.instances).toHaveLength(1);

    await act(async () => {
      await vi.advanceTimersByTimeAsync(2000);
    });
    expect(FakeWebSocket.instances).toHaveLength(2);
  });

  it("removes a landed flight 8s after the status update", async () => {
    const { result } = renderHook(() => useFlightSocket());
    const socket = FakeWebSocket.instances[0];
    const landed = flight({ status: "landed" });

    await act(async () => {
      socket.emitMessage({ type: "flight_update", flight: landed });
    });
    expect(result.current.flights.has(landed.id)).toBe(true);

    await act(async () => {
      await vi.advanceTimersByTimeAsync(7999);
    });
    expect(result.current.flights.has(landed.id)).toBe(true);

    await act(async () => {
      await vi.advanceTimersByTimeAsync(1);
    });
    expect(result.current.flights.has(landed.id)).toBe(false);
  });

  it("caps the activity log at 30 entries, newest first", async () => {
    const { result } = renderHook(() => useFlightSocket());
    const socket = FakeWebSocket.instances[0];

    await act(async () => {
      for (let i = 0; i < 31; i++) {
        socket.emitMessage({
          type: "flight_update",
          flight: flight({ id: `fl-${i}`, flightNumber: `UA${i}` }),
        });
      }
    });

    expect(result.current.log).toHaveLength(30);
    expect(result.current.log[0].text).toContain("UA30 SFO→JFK: airborne");
    expect(result.current.log[29].text).toContain("UA1 SFO→JFK: airborne");
  });

  it("counts events in the last second and resets every 1s", async () => {
    const { result } = renderHook(() => useFlightSocket());
    const socket = FakeWebSocket.instances[0];

    await act(async () => {
      socket.emitMessage({ type: "flight_update", flight: flight({ id: "a" }) });
      socket.emitMessage({ type: "flight_update", flight: flight({ id: "b" }) });
    });
    expect(result.current.eventsPerSecond).toBe(0); // not yet reported until the interval ticks

    await act(async () => {
      await vi.advanceTimersByTimeAsync(1000);
    });
    expect(result.current.eventsPerSecond).toBe(2);

    await act(async () => {
      await vi.advanceTimersByTimeAsync(1000);
    });
    expect(result.current.eventsPerSecond).toBe(0);
  });
});
