import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'
import { dirname, resolve } from 'path'
import { fileURLToPath } from 'url'

const __dirname = dirname(fileURLToPath(import.meta.url))

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, __dirname, '')
  const apiTarget = env.VITE_API_PROXY_TARGET || 'http://127.0.0.1:8090'
  const secure = apiTarget.startsWith('https')

  return {
    root: __dirname,
    plugins: [vue()],
    build: {
      outDir: resolve(__dirname, 'dist'),
      emptyOutDir: true,
    },
    server: {
      port: 5173,
      strictPort: true,
      proxy: {
        '/api': { target: apiTarget, changeOrigin: true, secure },
        '/health': { target: apiTarget, changeOrigin: true, secure },
      },
    },
  }
})
