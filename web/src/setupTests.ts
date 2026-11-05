import "@testing-library/jest-dom/vitest";
import { afterEach } from "vitest";
import { cleanup } from "@testing-library/react";

// Unmount rendered components after each test; vitest.config.ts doesn't set
// `globals: true`, so Testing Library's auto-cleanup never registers on its own.
afterEach(cleanup);
