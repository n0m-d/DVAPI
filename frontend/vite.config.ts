import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";


import tailwindcss from "@tailwindcss/vite";

// https://vite.dev/config/
export default defineConfig(async () => ({
  plugins: [vue(), tailwindcss()],
  resolve: {
    alias: {
      "@": "/src",
    },
  },
  server: {
    port: 5173,
    strictPort: true,
    host: "0.0.0.0",
  },
}));
