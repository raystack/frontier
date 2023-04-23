import react from "@vitejs/plugin-react-swc";
import dotenv from "dotenv";
import { defineConfig } from "vite";
import tsconfigPaths from "vite-tsconfig-paths";
dotenv.config();

// https://vitejs.dev/config/
export default defineConfig({
  build: {
    outDir: "dist/ui",
  },
  server: {
    proxy: {
      "/admin": {
        target: process.env.SHILD_API_URL,
        changeOrigin: true,
      },
    },
  },
  plugins: [react(), tsconfigPaths()],
});
