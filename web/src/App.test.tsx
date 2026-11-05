import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { act, render, screen } from "@testing-library/react";
import { App } from "./App";

class FakeWebSocket {
  onmessage: ((event: { data: string }) => void) | null = null;
  onerror: ((event: unknown) => void) | null = null;
  onclose: (() => void) | null = null;
  close() {}
}

describe("App", () => {
  beforeEach(() => {
    vi.stubGlobal("WebSocket", FakeWebSocket);
    vi.stubGlobal(
      "fetch",
      vi.fn((url: string) => {
        if (url === "/v1/meta") {
          return Promise.resolve({ ok: true, json: async () => ({ version: "v1", color: "blue" }) });
        }
        if (url === "/v1/airports") {
          return Promise.resolve({ ok: true, json: async () => [] });
        }
        return Promise.reject(new Error(`unexpected fetch: ${url}`));
      }),
    );
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("renders the header, map, and activity log", async () => {
    await act(async () => {
      render(<App />);
    });

    expect(screen.getByText(/SKYGATE/)).toBeInTheDocument();
    expect(document.getElementById("map")).toBeInTheDocument();
    expect(screen.getByText(/LIVE ACTIVITY/)).toBeInTheDocument();
  });
});
