import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";

export default defineConfig({
  plugins: [react(), tailwindcss()],
  build: {
    outDir: "../internal/web/dist",
    emptyOutDir: true,
  },
  server: {
    // The Jellyfin-compatible surface uses bare paths (no shared "/api"
    // prefix), so each top-level path family is proxied individually;
    // MagicBox's own extensions stay under /magicbox.
    proxy: {
      "/System": "http://localhost:8080",
      "/Users": "http://localhost:8080",
      "/Items": "http://localhost:8080",
      "/Videos": "http://localhost:8080",
      "/Sessions": "http://localhost:8080",
      "/magicbox": "http://localhost:8080",
    },
  },
});
