import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

// During local dev, proxy /api to the Go backend on :8080.
// In the container, nginx handles the /api proxy (see nginx.conf).
export default defineConfig({
  plugins: [react()],
  server: {
    port: 3000,
    proxy: {
      "/api": {
        target: "http://localhost:8080",
        changeOrigin: true,
      },
    },
  },
});
