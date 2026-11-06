import { describe, expect, it } from "vitest";
import { render, screen } from "@testing-library/react";
import { ColorLegend } from "./ColorLegend";

describe("ColorLegend", () => {
  it("shows a label for every flight status color group", () => {
    render(<ColorLegend />);
    expect(screen.getByText(/ground \/ holding/i)).toBeInTheDocument();
    expect(screen.getByText(/in flight/i)).toBeInTheDocument();
    expect(screen.getByText(/landed/i)).toBeInTheDocument();
    expect(screen.getByText(/cancelled/i)).toBeInTheDocument();
  });
});
