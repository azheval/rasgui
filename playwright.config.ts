import path from "node:path";

import { defineConfig } from "@playwright/test";

const testPort = 8199;
const testRunId = Date.now().toString();
const testDataDir = path.join(__dirname, "test-data", "playwright", testRunId);
const testLogDir = path.join(__dirname, "test-logs", "playwright", testRunId);
const mockRacPath = path.join(
  __dirname,
  "tests",
  "fixtures",
  process.platform === "win32" ? "rac-mock.cmd" : "rac-mock.sh",
);
const mockRasPath = path.join(
  __dirname,
  "tests",
  "fixtures",
  process.platform === "win32" ? "ras-mock.cmd" : "ras-mock.sh",
);
const webServerCommand =
  process.platform === "win32" ? "go run ./cmd/rasgui" : "chmod +x tests/fixtures/*.sh && go run ./cmd/rasgui";

export default defineConfig({
  testDir: "./tests/playwright",
  timeout: 30_000,
  use: {
    baseURL: `http://127.0.0.1:${testPort}`,
    trace: "retain-on-failure",
  },
  webServer: {
    command: webServerCommand,
    env: {
      ...process.env,
      RASGUI_HTTP_PORT: String(testPort),
      RASGUI_DATA_DIR: testDataDir,
      RASGUI_LOG_DIR: testLogDir,
      RASGUI_RAC_PATH: mockRacPath,
      RASGUI_RAS_PATH: mockRasPath,
    },
    url: `http://127.0.0.1:${testPort}/login`,
    reuseExistingServer: false,
    timeout: 120_000,
  },
});
