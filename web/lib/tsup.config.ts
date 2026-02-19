import cssModulesPlugin from 'esbuild-css-modules-plugin';
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
    dts: true,
    loader: {
      '.svg': 'dataurl',
      '.png': 'dataurl'
    },
    esbuildPlugins: [cssModulesPlugin({ localsConvention: 'camelCase' })]
  },
  // Hooks APIs
  {
    entry: ['hooks/index.ts'],
    outDir: 'hooks/dist',
    banner: {
      js: "'use client'"
    },
    format: ['cjs', 'esm'],
    external: ['react'],
    dts: true
  },
  // Admin APIs
  {
    entry: ['admin/index.ts'],
    outDir: 'admin/dist',
    banner: {
      js: "'use client'"
    },
    format: ['cjs', 'esm'],
    target: 'es2020',
    external: [
      'react',
      'react-dom',
      'react-router-dom',
      'zod',
      '@radix-ui/react-form',
      'sonner',
      'react-loading-skeleton',
      '@stitches/react',
      'usehooks-ts'
    ],
    dts: true,
    loader: {
      '.jpg': 'dataurl',
      '.png': 'dataurl'
    },
    esbuildPlugins: [cssModulesPlugin({ localsConvention: 'camelCase' })]
  }
]);
