// SPDX-License-Identifier: AGPL-3.0-or-later
/// <reference types="cypress" />
// ***********************************************
// This example commands.ts shows you how to
// create various custom commands and overwrite
// existing commands.
//
// For more comprehensive examples of custom
// commands please read more here:
// https://on.cypress.io/custom-commands
// ***********************************************

declare global {
  namespace Cypress {
    interface Chainable {
      /**
       * Visit a page with locale set to English
       * @param url - URL to visit
       * @param options - Visit options
       */
      visitWithLocale(url: string, locale?: string, options?: Partial<Cypress.VisitOptions>): Chainable<Cypress.AUTWindow>

      /**
       * Login via MagicLink authentication
       * @param email - Email address to login with
       * @param redirectTo - Optional redirect URL after login
       */
      loginViaMagicLink(email: string, redirectTo?: string): Chainable<void>

      /**
       * Login as admin via MagicLink
       */
      loginAsAdmin(): Chainable<void>

      /**
       * Logout current user
       */
      logout(): Chainable<void>
    }
  }
}

Cypress.Commands.add('visitWithLocale', (url: string, locale: string = 'en', options?: Partial<Cypress.VisitOptions>) => {
  return cy.visit(url, {
    ...options,
    onBeforeLoad: (win) => {
      win.localStorage.setItem('locale', locale)
      if (options?.onBeforeLoad) {
        options.onBeforeLoad(win)
      }
    }
  })
})

Cypress.Commands.add('loginViaMagicLink', (email: string, redirectTo?: string) => {
  const baseUrl = Cypress.config('baseUrl')

  // Clear mailbox first
  cy.clearMailbox()

  // Request magic link
  const authUrl = redirectTo ? `/auth?redirect=${encodeURIComponent(redirectTo)}` : '/auth'
  cy.visitWithLocale(authUrl)

  cy.get('input[type="email"]', { timeout: 10000 }).should('be.visible').clear().type(email)
  cy.contains('button', 'Send Magic Link').should('be.visible').click()

  // Wait for success message
  cy.contains('Check your email', { timeout: 10000 }).should('be.visible')

  // Get magic link from email (subject from backend i18n: email.magic_link.subject)
  const emailSubject = 'Your login link' // en.json: email.magic_link.subject
  cy.waitForEmail(email, emailSubject, 30000).then((message) => {
    cy.extractMagicLink(message).then((magicLink) => {
      // Visit magic link
      cy.visit(magicLink)

      // Wait for redirect to complete
      cy.url({ timeout: 10000 }).should('not.include', '/auth/magic-link/verify')

      // Verify authentication
      cy.request('/api/v1/users/me').its('status').should('eq', 200)
    })
  })
})

Cypress.Commands.add('loginAsAdmin', () => {
  cy.loginViaMagicLink('admin@test.com')
})

Cypress.Commands.add('logout', () => {
  cy.request('/api/v1/auth/logout').then(() => {
    cy.clearCookies()
  })
})

export {}
