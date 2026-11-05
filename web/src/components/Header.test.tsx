import { describe, expect, it } from "vitest";
import { render, screen } from "@testing-library/react";
import { Header } from "./Header";

describe("Header", () => {
  it("renders the title, badge, and stats together", () => {
    render(<Header meta={{ version: "v1", color: "blue" }} eventsPerSecond={3} activeFlights={5} />);
    expect(screen.getByText(/SKYGATE/)).toBeInTheDocument();
    expect(screen.getByText(/v1/)).toBeInTheDocument();
    expect(screen.getByText("3")).toBeInTheDocument();
    expect(screen.getByText("5")).toBeInTheDocument();
  });
});
