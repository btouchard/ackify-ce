import { defineConfig } from 'vitest/config'
import vue from '@vitejs/plugin-vue'
import { fileURLToPath, URL } from 'node:url'
import istanbul from 'vite-plugin-istanbul'

const isE2ECoverage = process.env.CYPRESS_COVERAGE === 'true'

export default defineConfig({
  plugins: [
    vue(),
    // Instrument code for E2E coverage (only when CYPRESS_COVERAGE=true)
    ...(isE2ECoverage ? [istanbul({
      include: 'src/**/*',
      exclude: ['node_modules/**', 'tests/**', 'cypress/**', 'dist/**'],
      extension: ['.js', '.ts', '.vue'],
      requireEnv: false,
      forceBuildInstrument: true,
      cypress: true
    })] : [])
  ],
  build: {
    // Disable minification when instrumenting for coverage
    ...(isE2ECoverage ? {
      minify: false,
      sourcemap: true
    } : {}),
    rollupOptions: {
      onwarn(warning, warn) {
        // Suppress "currentInstance" not exported warning from vue-i18n
        // This is a known issue with vue-i18n accessing internal Vue APIs
        if (
          warning.code === 'MISSING_EXPORT' &&
          warning.exporter?.includes('vue.runtime.esm-bundler.js') &&
          warning.message.includes('currentInstance')
        ) {
          return
        }
        warn(warning)
      }
    }
  },
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    }
  },
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        secure: false
      },
      '/oauth2': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        secure: false
      }
    }
  },
  // Vitest configuration for unit tests
  test: {
    globals: true,
    environment: 'happy-dom',
    setupFiles: './tests/setup.ts',
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json', 'html', 'lcov'],
      reportsDirectory: './coverage',
      exclude: [
        'node_modules/',
        'tests/',
        'cypress/',
        '**/*.spec.ts',
        '**/*.test.ts',
        'dist/',
        '.eslintrc.cjs',
        'vite.config.ts',
        'cypress.config.ts',
        'src/main.ts'
      ],
      include: ['src/**/*.{js,ts,vue}']
      // Note: No thresholds set - coverage is for information only
      // Add thresholds when baseline coverage is established
    }
  }
})
