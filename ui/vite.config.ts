import react from "@vitejs/plugin-react-swc";
import dotenv from "dotenv";
import { defineConfig } from "vite";
import tsconfigPaths from "vite-tsconfig-paths";
import svgr from "vite-plugin-svgr";
dotenv.config();

// https://vitejs.dev/config/
export default defineConfig(() => {
  return {
    base: "/",
    build: {
      outDir: "dist/ui",
    },
    server: {
      proxy: {
        "/frontier-api": {
          target: process.env.FRONTIER_API_URL,
          changeOrigin: true,
          rewrite: (path) => path.replace(/^\/frontier-api/, ""),
        },
      },
      fs: {
        // Allow serving files from one level up to the project root
        allow: [".."],
      },
    },
    plugins: [react(), svgr(), tsconfigPaths()],
    define: {
      "process.env": process.env,
    },
  };
});
