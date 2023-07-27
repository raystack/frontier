import { defineConfig } from "tsup";
import pkg from "./package.json";
const external = [...Object.keys(pkg.dependencies || {})];

export default defineConfig(() => [
  {
    entryPoints: ["src/index.ts"],
    dts: true,
    clean: true,
    sourcemap: true,
    format: ["cjs", "esm"],
    target: "node16",
    external,
  },
  {
    entry: { bundle: "src/index.ts" },
    format: ["iife"],
    globalName: "frontierdev",
    clean: false,
    minify: true,
    platform: "browser",
    dts: false,
    name: "frontier",
    // esbuild `globalName` option generates `var frontierdev = (() => {})()`
    // and var is not guaranteed to assign to the global `window` object so we make sure to assign it
    footer: {
      js: "window.__SHIELD__ = frontierdev",
    },
    outExtension({ format, options }) {
      return {
        js: ".min.js",
      };
    },
    esbuildOptions(options, ctx) {
      options.entryNames = "frontier";
    },
  },
]);
