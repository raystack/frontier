import react from "@vitejs/plugin-react-swc";
import dotenv from "dotenv";
import { defineConfig, type Plugin } from "vite";
import tsconfigPaths from "vite-tsconfig-paths";
import svgr from "vite-plugin-svgr";
import fs from "node:fs";
import path from "node:path";
dotenv.config();

/**
 * Vite plugin that serves a local JSON file at `/configs` during development.
 * Edit `configs.dev.json` in the project root to change the config
 * (including terminology overrides like organization → workspace).
 */
function devConfigsPlugin(): Plugin {
  return {
    name: "dev-configs",
    configureServer(server) {
      server.middlewares.use("/configs", (_req, res) => {
        const configPath = path.resolve(__dirname, "configs.dev.json");
        try {
          const content = fs.readFileSync(configPath, "utf-8");
          // Re-read on every request so changes are picked up without restart
          JSON.parse(content); // validate JSON
          res.setHeader("Content-Type", "application/json");
          res.end(content);
        } catch {
          res.statusCode = 500;
          res.end(JSON.stringify({ error: "Failed to read configs.dev.json" }));
        }
      });
    },
  };
}

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
    plugins: [devConfigsPlugin(), react(), svgr(), tsconfigPaths()],
    resolve: {
      // Force a single React runtime across app + linked workspace SDK packages
      dedupe: ["react", "react-dom"],
    },
    define: {
      "process.env": process.env,
    },
  };
});
