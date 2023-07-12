import type { Options } from "tsup";

const env = process.env.NODE_ENV;
const isProduction = env !== "development";

export const tsup: Options = {
  dts: true,
  clean: true,
  splitting: false,
  minify: isProduction,
  watch: !isProduction,
  target: "es2020",
  format: ["cjs", "esm"],
  skipNodeModulesBundle: true,
  entryPoints: ["src/index.ts"],
  publicDir: true,
};
