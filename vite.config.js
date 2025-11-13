import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import svgr from 'vite-plugin-svgr';
import { rmSync } from 'node:fs';
import { resolve } from 'node:path';

const cleanStaticDir = () => ({
  name: 'clean-static-dir',
  apply: 'build',
  buildStart() {
    const target = resolve(process.cwd(), 'dist/static');
    rmSync(target, { recursive: true, force: true });
  },
});

export default defineConfig({
  plugins: [
    cleanStaticDir(),
    react(),
    svgr(),
  ],
  resolve: {
    conditions: ['mui-modern', 'module', 'import', 'browser', 'default'],
  },
  server: {
    port: 3000,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        ws: true,
      },
    },
  },
  build: {
    target: 'esnext',
    outDir: 'dist',
    assetsDir: 'static',
    emptyOutDir: false,
    chunkSizeWarningLimit: 1024,
  },
});
