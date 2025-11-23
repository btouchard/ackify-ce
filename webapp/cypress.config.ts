// SPDX-License-Identifier: AGPL-3.0-or-later
import { defineConfig } from 'cypress'
// @ts-ignore - no types available for @cypress/code-coverage
import codeCoverageTask from '@cypress/code-coverage/task'

export default defineConfig({
  e2e: {
    baseUrl: 'http://localhost:8080',
    specPattern: 'cypress/e2e/**/*.cy.{js,jsx,ts,tsx}',
    supportFile: 'cypress/support/e2e.ts',
    fixturesFolder: 'cypress/fixtures',
    video: false,
    screenshotOnRunFailure: true,
    defaultCommandTimeout: 10000,
    requestTimeout: 10000,
    env: {
      mailhogUrl: 'http://localhost:8025',
      // Enable coverage collection
      codeCoverage: {
        exclude: ['cypress/**/*.*', 'tests/**/*.*']
      }
    },
    setupNodeEvents(on, config) {
      // Register code coverage plugin
      codeCoverageTask(on, config)
      return config
    },
  },
})
