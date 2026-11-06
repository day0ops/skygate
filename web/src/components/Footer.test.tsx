import { describe, expect, it } from "vitest";
import { render, screen } from "@testing-library/react";
import { Footer } from "./Footer";

describe("Footer", () => {
  it("shows the Skygate name and build version once meta loads", () => {
    render(<Footer meta={{ version: "v0.2.1", color: "blue", buildVersion: "v0.3.0" }} />);
    expect(screen.getByText("Skygate v0.3.0")).toBeInTheDocument();
  });

  it("shows the real build version, not a canary demo label", () => {
    // routing-canary overrides `version` with a demo label like "v1"/"v2" -
    // the footer must use buildVersion instead, or every canary deployment
    // would show the wrong thing here.
    render(<Footer meta={{ version: "v1", color: "blue", buildVersion: "v0.3.0" }} />);
    expect(screen.getByText("Skygate v0.3.0")).toBeInTheDocument();
    expect(screen.queryByText("Skygate v1")).not.toBeInTheDocument();
  });

  it("renders nothing while meta is still loading", () => {
    const { container } = render(<Footer meta={null} />);
    expect(container).toBeEmptyDOMElement();
  });
});
