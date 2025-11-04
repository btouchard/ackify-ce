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

export {}
