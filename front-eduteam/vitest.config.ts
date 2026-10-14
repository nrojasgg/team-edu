import { defineConfig } from 'vitest/config'
import vue from '@vitejs/plugin-vue'
import { fileURLToPath } from 'url'
import { resolve, dirname } from 'path'

const __dirname = dirname(fileURLToPath(import.meta.url))

export default defineConfig({
  plugins: [vue()],
  // Make import.meta.client === true for all test files so composables
  // that branch on it (useAuth, useSemester …) take the browser-side path.
  define: {
    'import.meta.client': JSON.stringify(true),
    'import.meta.server': JSON.stringify(false),
  },
  resolve: {
    alias: {
      // Allow imports like '~/composables/...' inside source files if needed
      '~': resolve(__dirname, '.'),
    },
  },
  test: {
    environment: 'happy-dom',
    globals: true,
    // Run setup file before every test suite
    setupFiles: ['./tests/setup.js'],
    include: [
      'tests/**/*.{spec,test}.{js,ts}',
      '**/*.{spec,test}.{js,ts}',
    ],
    exclude: ['node_modules', '.nuxt', '.output', 'dist'],
    coverage: {
      provider: 'istanbul',
      // Output directory for coverage reports
      reportsDirectory: 'coverage',
      reporter: ['lcov', 'text', 'html'],
      include: [
        'components/**/*.vue',
        'composables/**/*.{js,ts}',
        'pages/**/*.vue',
        'plugins/**/*.{js,ts}',
        'middleware/**/*.{js,ts}',
      ],
      exclude: [
        'node_modules/**',
        '.nuxt/**',
        '.output/**',
        'dist/**',
        '**/*.spec.{js,ts}',
        '**/*.test.{js,ts}',
      ],
    },
  },
})
