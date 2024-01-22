import react from "@vitejs/plugin-react-swc";
import dotenv from "dotenv";
import { defineConfig } from "vite";
import tsconfigPaths from "vite-tsconfig-paths";
dotenv.config();

const FRONTIER_API_URL =
  process.env.FRONTIER_API_URL || "http://localhost:8000";

// https://vitejs.dev/config/
export default defineConfig(({ command, mode }) => {
  return {
    base: "/",
    build: {
      outDir: "dist/ui",
    },
    server: {
      proxy: {
        "/frontier-api": {
          target: FRONTIER_API_URL,
          changeOrigin: true,
          rewrite: (path) => path.replace(/^\/frontier-api/, ""),
        },
      },
      fs: {
        // Allow serving files from one level up to the project root
        allow: [".."],
      },
    },
    plugins: [react(), tsconfigPaths()],
    define: {
      "process.env": process.env,
    },
  };
});
