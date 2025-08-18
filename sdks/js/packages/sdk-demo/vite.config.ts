import { defineConfig, loadEnv } from 'vite';
import react from '@vitejs/plugin-react';
import path from 'path';

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '');
  const frontierUrl = env.FRONTIER_ENDPOINT || 'http://localhost:8080';
  const frontierConnectUrl =
    env.FRONTIER_CONNECT_ENDPOINT || 'http://localhost:8080';

  return {
    plugins: [react()],
    resolve: {
      alias: {
        '@': path.resolve(__dirname, './src')
      }
    },
    server: {
      port: 3000,
      proxy: {
        '/api': {
          target: frontierUrl,
          changeOrigin: true,
          rewrite: path => path.replace(/^\/api/, '')
        },
        '/frontier-connect': {
          target: frontierConnectUrl,
          changeOrigin: true,
          rewrite: path => path.replace(/^\/frontier-connect/, '')
        }
      }
    }
  };
});
