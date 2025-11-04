// SPDX-License-Identifier: AGPL-3.0-or-later
import { defineConfig } from 'cypress'

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
    },
    setupNodeEvents(on, config) {
      // implement node event listeners here
      return config
    },
  },
})
