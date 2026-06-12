import { fileURLToPath, URL } from 'node:url'
import vue from '@vitejs/plugin-vue'
import { defineConfig, loadEnv } from 'vite'
import { validateTeacherApiRuntimeEnv } from './src/services/runtimeEnv'

const projectRoot = fileURLToPath(new URL('.', import.meta.url))

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, projectRoot, '')
  const isProductionBuild = mode === 'production'

  if (isProductionBuild) {
    validateTeacherApiRuntimeEnv(
      {
        DEV: false,
        PROD: true,
        VITE_SERVER_WEB_API_MODE: env.VITE_SERVER_WEB_API_MODE,
        VITE_SERVER_WEB_API_BASE_URL: env.VITE_SERVER_WEB_API_BASE_URL,
      },
      true,
    )
  }

  return {
    plugins: [vue()],
    resolve: {
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url)),
      },
    },
    test: {
      environment: 'jsdom',
      globals: true,
      setupFiles: ['./src/test/setup.ts'],
    },
  }
})
