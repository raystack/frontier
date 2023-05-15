import commonjs from "@rollup/plugin-commonjs";
import resolve from "@rollup/plugin-node-resolve";
import dts from "rollup-plugin-dts";
import peerDepsExternal from "rollup-plugin-peer-deps-external";
import typescript from "rollup-plugin-typescript2";

export default [
  {
    input: "src/index.ts",
    output: [
      {
        file: "dist/cjs/index.js",
        format: "cjs",
        sourcemap: true,
        globals: {
          react: "React",
        },
      },
      {
        file: "dist/esm/index.js",
        format: "esm",
        sourcemap: true,
        globals: {
          react: "React",
        },
      },
    ],
    plugins: [
      peerDepsExternal(),
      resolve({
        browser: true,
      }),
      commonjs(),
      typescript({ clean: true, typescript: require("typescript") }),
      // terser(),
    ],
    external: ["react", "react/jsx-runtime"],
  },
  {
    input: "dist/esm/src/types/index.d.ts",
    output: [{ file: "dist/index.d.ts", format: "es" }],
    plugins: [dts()],
  },
];
