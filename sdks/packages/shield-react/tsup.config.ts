import externalGlobal from "esbuild-plugin-external-global";
import type { Options } from "tsup";

export const tsup: Options = {
  entry: ["src/index.ts"],
  dts: true,
  clean: true,
  sourcemap: true,
  external: ["react", "react-dom"],
  format: ["cjs", "esm", "iife"],
  skipNodeModulesBundle: true,
  globalName: "Shield",
  target: "es6",
  banner: {
    js: "'use client'",
  },
  /**
   * Couple build options for UMD/iife build.
   *  - externalGlobalPlugin to use window.React instead of trying to bundle it.
   *  - for iife build, set platform to "browser" and define process.env.NODE_ENV
   */
  esbuildPlugins: [
    externalGlobal.externalGlobalPlugin({
      react: "window.React",
      "react-dom": "window.ReactDOM",
    }),
  ],

  esbuildOptions: (options, context) => {
    if (context.format === "iife") {
      options.minify = true;
      options.platform = "browser";
      options.define!["process.env.NODE_ENV"] = '"production"';
    }
  },
};
