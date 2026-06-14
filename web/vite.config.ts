import path from "path"
import tailwindcss from "@tailwindcss/vite"
import react from "@vitejs/plugin-react"
import { defineConfig } from "vite"

// https://vite.dev/config/
export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
  server: {
    // Mirror the container's nginx: /api goes to the agent-node container
    // (docker compose publishes nginx on 5500), which strips the prefix.
    proxy: {
      "/api": { target: "http://localhost:5500", changeOrigin: true },
    },
  },
})
