import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { act, renderHook } from "@testing-library/react";
import { useMeta } from "./useMeta";

describe("useMeta", () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    vi.useRealTimers();
  });

  it("returns meta on a successful fetch", async () => {
    const fetchMock = vi.fn().mockResolvedValueOnce({
      ok: true,
      json: async () => ({ version: "v1.2.3", color: "green" }),
    });
    vi.stubGlobal("fetch", fetchMock);

    const { result } = renderHook(() => useMeta());

    await act(async () => {});

    expect(result.current).toEqual({ version: "v1.2.3", color: "green" });
  });

  it("retries every 2s on failure until it succeeds", async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce({ ok: false, status: 503 })
      .mockResolvedValueOnce({
        ok: true,
        json: async () => ({ version: "v1", color: "blue" }),
      });
    vi.stubGlobal("fetch", fetchMock);

    const { result } = renderHook(() => useMeta());

    await act(async () => {});
    expect(fetchMock).toHaveBeenCalledTimes(1);
    expect(result.current).toBeNull();

    await act(async () => {
      await vi.advanceTimersByTimeAsync(2000);
    });

    expect(fetchMock).toHaveBeenCalledTimes(2);
    expect(result.current).toEqual({ version: "v1", color: "blue" });
  });
});
