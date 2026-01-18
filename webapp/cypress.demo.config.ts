// SPDX-License-Identifier: AGPL-3.0-or-later
import { defineConfig } from 'cypress'

export default defineConfig({
  e2e: {
    baseUrl: 'http://localhost:8080',
    specPattern: 'cypress/e2e/demo.cy.ts',
    supportFile: 'cypress/support/e2e.ts',
    fixturesFolder: 'cypress/fixtures',

    // Video recording enabled for demo
    video: true,
    videoCompression: 15, // Lower = better quality
    videosFolder: 'cypress/videos/demo',

    // Screenshots
    screenshotOnRunFailure: true,
    screenshotsFolder: 'cypress/screenshots/demo',

    // Longer timeouts for demo with pauses
    defaultCommandTimeout: 20000,
    requestTimeout: 20000,
    pageLoadTimeout: 60000,

    // Viewport for the app (will be larger with Cypress UI)
    viewportWidth: 1400,
    viewportHeight: 900,

    env: {
      mailhogUrl: 'http://localhost:8025',
    },

    setupNodeEvents(on, config) {
      // Force larger window size for Chrome headless
      on('before:browser:launch', (browser, launchOptions) => {
        if (browser.name === 'chrome' && browser.isHeadless) {
          // Set window size for headless Chrome
          launchOptions.args.push('--window-size=1920,1080')
          launchOptions.args.push('--force-device-scale-factor=1')
        }
        return launchOptions
      })

      on('after:spec', (spec, results) => {
        if (results.video) {
          console.log(`Video recorded: ${results.video}`)
        }
      })
      return config
    },
  },
})