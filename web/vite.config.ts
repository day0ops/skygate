import { fileURLToPath } from "node:url";
import { dirname, resolve } from "node:path";
import { defineConfig } from "vitest/config";
import react from "@vitejs/plugin-react";

const __dirname = dirname(fileURLToPath(import.meta.url));

export default defineConfig({
  plugins: [react()],
  build: {
    outDir: resolve(__dirname, "../internal/web/static"),
    emptyOutDir: true,
  },
  test: {
    environment: "jsdom",
    setupFiles: ["./src/setupTests.ts"],
  },
});
