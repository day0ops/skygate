import { useEffect, useState } from "react";
import type { Meta } from "../types";

const RETRY_DELAY_MS = 2000;

/**
 * Fetches /v1/meta once and retries every 2s on failure.
 * @returns the loaded Meta, or null until the first successful load completes.
 */
export function useMeta(): Meta | null {
  const [meta, setMeta] = useState<Meta | null>(null);

  useEffect(() => {
    let cancelled = false;
    let timer: ReturnType<typeof setTimeout> | undefined;

    async function load() {
      try {
        const res = await fetch("/v1/meta");
        if (!res.ok) throw new Error(`status ${res.status}`);
        const data = (await res.json()) as Meta;
        if (!cancelled) setMeta(data);
      } catch (err) {
        console.error("skygate: failed to load /v1/meta, retrying in 2s", err);
        if (!cancelled) timer = setTimeout(load, RETRY_DELAY_MS);
      }
    }

    load();

    return () => {
      cancelled = true;
      clearTimeout(timer);
    };
  }, []);

  return meta;
}
