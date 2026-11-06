import { useEffect, useRef, useState } from "react";
import type { Flight, FlightUpdateMessage, LogEntry } from "../types";

const RECONNECT_DELAY_MS = 2000;
const LANDED_REMOVAL_DELAY_MS = 8000;
const LOG_CAP = 30;
const RPS_INTERVAL_MS = 1000;

export interface FlightSocketState {
  flights: Map<string, Flight>;
  log: LogEntry[];
  eventsPerSecond: number;
}

/**
 * Connects to /ws/flights, reconnecting 2s after any close, and keeps
 * flight state, a capped activity log, and an events-per-second rate
 * updated as flight_update messages arrive. Landed flights are removed
 * 8s after their landed update.
 * @returns the current flights, activity log, and events-per-second rate.
 */
export function useFlightSocket(): FlightSocketState {
  const [flights, setFlights] = useState<Map<string, Flight>>(new Map());
  const [log, setLog] = useState<LogEntry[]>([]);
  const [eventsPerSecond, setEventsPerSecond] = useState(0);
  const landedTimersRef = useRef<Map<string, ReturnType<typeof setTimeout>>>(new Map());
  const logIdRef = useRef(0);
  const eventCountRef = useRef(0);

  useEffect(() => {
    let cancelled = false;
    let socket: WebSocket | undefined;
    let reconnectTimer: ReturnType<typeof setTimeout> | undefined;
    const landedTimers = landedTimersRef.current;

    function connect() {
      const proto = location.protocol === "https:" ? "wss" : "ws";
      socket = new WebSocket(`${proto}://${location.host}/ws/flights`);

      socket.onmessage = (event) => {
        const msg = JSON.parse(event.data) as FlightUpdateMessage;
        if (msg.type !== "flight_update") return;
        const incoming = msg.flight;

        setFlights((prev) => {
          const next = new Map(prev);
          next.set(incoming.id, incoming);
          return next;
        });

        if (incoming.status === "landed") {
          const existing = landedTimers.get(incoming.id);
          if (existing) clearTimeout(existing);
          const timer = setTimeout(() => {
            setFlights((prev) => {
              const next = new Map(prev);
              next.delete(incoming.id);
              return next;
            });
            landedTimers.delete(incoming.id);
          }, LANDED_REMOVAL_DELAY_MS);
          landedTimers.set(incoming.id, timer);
        }

        logIdRef.current += 1;
        const text = `${new Date().toLocaleTimeString()}  ${incoming.flightNumber} ${incoming.originCode}→${incoming.destinationCode}: ${incoming.status}`;
        setLog((prev) => [{ id: logIdRef.current, text }, ...prev].slice(0, LOG_CAP));

        eventCountRef.current += 1;
      };

      socket.onerror = (event) => {
        console.error("skygate: websocket error", event);
      };

      socket.onclose = () => {
        if (!cancelled) reconnectTimer = setTimeout(connect, RECONNECT_DELAY_MS);
      };
    }

    connect();

    const rpsInterval = setInterval(() => {
      setEventsPerSecond(eventCountRef.current);
      eventCountRef.current = 0;
    }, RPS_INTERVAL_MS);

    return () => {
      cancelled = true;
      clearTimeout(reconnectTimer);
      clearInterval(rpsInterval);
      landedTimers.forEach(clearTimeout);
      landedTimers.clear();
      socket?.close();
    };
  }, []);

  return { flights, log, eventsPerSecond };
}
