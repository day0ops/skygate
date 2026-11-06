import { describe, expect, it } from "vitest";
import { render, screen } from "@testing-library/react";
import { VersionBadge } from "./VersionBadge";

describe("VersionBadge", () => {
  it("shows a loading state when meta is null", () => {
    render(<VersionBadge meta={null} />);
    expect(screen.getByText(/loading/i)).toBeInTheDocument();
  });

  it("shows the version and color once meta loads", () => {
    render(<VersionBadge meta={{ version: "v1.2.3", color: "green", buildVersion: "v0.3.0" }} />);
    expect(screen.getByText(/v1\.2\.3/)).toBeInTheDocument();
    expect(screen.getByText(/GREEN/)).toBeInTheDocument();
  });

  it("resolves the red color to the accent token", () => {
    render(<VersionBadge meta={{ version: "v1.2.3", color: "red", buildVersion: "v0.3.0" }} />);
    expect(screen.getByText(/RED/)).toHaveStyle({ "--badge-color": "var(--accent)" });
  });

  it("falls back to the default blue for an unrecognized color", () => {
    render(<VersionBadge meta={{ version: "v1.2.3", color: "blue", buildVersion: "v0.3.0" }} />);
    expect(screen.getByText(/BLUE/)).toHaveStyle({ "--badge-color": "#3b82f6" });
  });
});
