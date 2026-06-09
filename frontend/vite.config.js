import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

const configuredHosts = [
  "logarift.local",
  ...(process.env.VITE_ALLOWED_HOSTS || "")
    .split(",")
    .map((host) => host.trim())
    .filter(Boolean),
  ...(process.env.__VITE_ADDITIONAL_SERVER_ALLOWED_HOSTS || "")
    .split(",")
    .map((host) => host.trim())
    .filter(Boolean),
];

const allowedHosts = Array.from(new Set(configuredHosts));

export default defineConfig({
  plugins: [react()],
  server: {
    host: "0.0.0.0",
    port: 5173,
    strictPort: true,
    allowedHosts,
  },
});
