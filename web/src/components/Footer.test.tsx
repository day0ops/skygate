import { describe, expect, it } from "vitest";
import { render, screen } from "@testing-library/react";
import { Footer } from "./Footer";

describe("Footer", () => {
  it("shows the Skygate name and build version once meta loads", () => {
    render(<Footer meta={{ version: "v0.2.1", color: "blue" }} />);
    expect(screen.getByText("Skygate v0.2.1")).toBeInTheDocument();
  });

  it("renders nothing while meta is still loading", () => {
    const { container } = render(<Footer meta={null} />);
    expect(container).toBeEmptyDOMElement();
  });
});
