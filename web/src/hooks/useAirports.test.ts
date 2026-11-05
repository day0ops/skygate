import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { act, renderHook } from "@testing-library/react";
import { useAirports } from "./useAirports";
import type { Airport } from "../types";

const sfo: Airport = { code: "SFO", name: "San Francisco Intl", city: "San Francisco", x: 100, y: 200, gates: [] };

describe("useAirports", () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    vi.useRealTimers();
  });

  it("returns the airport list on a successful fetch", async () => {
    const fetchMock = vi.fn().mockResolvedValueOnce({
      ok: true,
      json: async () => [sfo],
    });
    vi.stubGlobal("fetch", fetchMock);

    const { result } = renderHook(() => useAirports());

    await act(async () => {});

    expect(result.current).toEqual([sfo]);
  });

  it("retries every 2s on failure until it succeeds", async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce({ ok: false, status: 500 })
      .mockResolvedValueOnce({ ok: true, json: async () => [sfo] });
    vi.stubGlobal("fetch", fetchMock);

    const { result } = renderHook(() => useAirports());

    await act(async () => {});
    expect(fetchMock).toHaveBeenCalledTimes(1);
    expect(result.current).toEqual([]);

    await act(async () => {
      await vi.advanceTimersByTimeAsync(2000);
    });

    expect(fetchMock).toHaveBeenCalledTimes(2);
    expect(result.current).toEqual([sfo]);
  });
});
