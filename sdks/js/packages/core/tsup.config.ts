import { defineConfig } from 'tsup';

export default defineConfig(() => [
  // Core API
  {
    entry: ['src/index.ts'],
    format: ['cjs', 'esm'],
    dts: true
  },
  // React APIs
  {
    entry: ['react/index.ts'],
    outDir: 'react/dist',
    banner: {
      js: "'use client'"
    },
    format: ['cjs', 'esm'],
    external: ['react', 'svelte', 'vue', 'solid-js'],
    dts: true
  }
]);
