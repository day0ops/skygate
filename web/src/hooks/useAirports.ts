import { useEffect, useState } from "react";
import type { Airport } from "../types";

const RETRY_DELAY_MS = 2000;

/**
 * Fetches /v1/airports once and retries every 2s on failure.
 * @returns the loaded airports, or an empty array until the first successful load completes.
 */
export function useAirports(): Airport[] {
  const [airports, setAirports] = useState<Airport[]>([]);

  useEffect(() => {
    let cancelled = false;
    let timer: ReturnType<typeof setTimeout> | undefined;

    async function load() {
      try {
        const res = await fetch("/v1/airports");
        if (!res.ok) throw new Error(`status ${res.status}`);
        const data = (await res.json()) as Airport[];
        if (!cancelled) setAirports(data);
      } catch (err) {
        console.error("skygate: failed to load /v1/airports, retrying in 2s", err);
        if (!cancelled) timer = setTimeout(load, RETRY_DELAY_MS);
      }
    }

    load();

    return () => {
      cancelled = true;
      clearTimeout(timer);
    };
  }, []);

  return airports;
}
