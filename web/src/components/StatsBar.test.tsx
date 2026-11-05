import { describe, expect, it } from "vitest";
import { render, screen } from "@testing-library/react";
import { StatsBar } from "./StatsBar";

describe("StatsBar", () => {
  it("renders events/sec and active flight counts", () => {
    render(<StatsBar eventsPerSecond={7} activeFlights={12} />);
    expect(screen.getByText("7")).toBeInTheDocument();
    expect(screen.getByText("12")).toBeInTheDocument();
  });
});
