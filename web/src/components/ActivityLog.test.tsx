import { describe, expect, it } from "vitest";
import { render, screen } from "@testing-library/react";
import { ActivityLog } from "./ActivityLog";

describe("ActivityLog", () => {
  it("renders every entry in the given order", () => {
    render(
      <ActivityLog
        entries={[
          { id: 2, text: "UA2 SFO→JFK: airborne" },
          { id: 1, text: "UA1 SFO→JFK: departed" },
        ]}
      />,
    );

    const items = screen.getAllByRole("listitem");
    expect(items).toHaveLength(2);
    expect(items[0]).toHaveTextContent("UA2 SFO→JFK: airborne");
    expect(items[1]).toHaveTextContent("UA1 SFO→JFK: departed");
  });

  it("shows the panel header", () => {
    render(<ActivityLog entries={[]} />);
    expect(screen.getByText(/LIVE ACTIVITY/)).toBeInTheDocument();
  });
});
