import react from "@vitejs/plugin-react-swc";
import dotenv from "dotenv";
import { createRequire } from "module";
import path from "path";
import { defineConfig } from "vite";
import tsconfigPaths from "vite-tsconfig-paths";
import svgr from "vite-plugin-svgr";
dotenv.config();

const require = createRequire(import.meta.url);
const reactRouterDomPath = path.dirname(require.resolve("react-router-dom/package.json"));

// https://vitejs.dev/config/
export default defineConfig(() => {
  return {
    base: "/",
    build: {
      outDir: "dist/admin",
    },
    server: {
      proxy: {
        "/frontier-api": {
          target: process.env.FRONTIER_API_URL,
          changeOrigin: true,
          rewrite: (path) => path.replace(/^\/frontier-api/, ""),
        },
        "/frontier-connect": {
          target: process.env.FRONTIER_CONNECTRPC_URL,
          changeOrigin: true,
          rewrite: (path) => path.replace(/^\/frontier-connect/, ""),
        },
      },
      fs: {
        // Allow serving files from one level up to the project root
        allow: [".."],
      },
    },
    resolve: {
      alias: {
        "react-router-dom": reactRouterDomPath,
      },
    },
    plugins: [react(), svgr(), tsconfigPaths()],
    define: {
      "process.env": process.env,
    },
  };
});
