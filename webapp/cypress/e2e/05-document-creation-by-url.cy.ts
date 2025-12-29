// SPDX-License-Identifier: AGPL-3.0-or-later
/// <reference types="cypress" />

describe('Test 5: Document Creation by URL', () => {
  const testUser = 'urltest@example.com'

  beforeEach(() => {
    cy.clearMailbox()
    cy.clearCookies()
  })

  it('should auto-create document from external URL', () => {
    const externalUrl = 'https://example.com/policy.pdf'

    // Step 1: Login first
    cy.loginViaMagicLink(testUser)

    // Step 2: Access document by URL
    cy.visit(`/?doc=${encodeURIComponent(externalUrl)}`)

    // Step 3: Should auto-create document
    cy.url({ timeout: 10000 }).should('include', '/?doc=')

    // Step 4: Verify document metadata contains the URL
    cy.contains('example.com/policy.pdf', { timeout: 10000 }).should('be.visible')

    // Step 5: Should be able to sign
    cy.get('[data-testid="sign-button"]').should('be.visible')
  })

  it('should accept full URL as docId and create document', () => {
    const testUrl = 'https://docs.example.com/terms-' + Date.now() + '.pdf'

    cy.loginViaMagicLink(testUser)

    // Visit with full URL
    cy.visit(`/?doc=${encodeURIComponent(testUrl)}`, { timeout: 15000 })

    // Should accept the URL as docId (no transformation)
    cy.url({ timeout: 15000 }).should('include', '/?doc=')

    // Document should be created and signable
    cy.get('[data-testid="sign-button"]', { timeout: 10000 }).should('be.visible')

    // Should display the URL in metadata
    cy.contains('docs.example.com/terms').should('be.visible')
  })

  it('should handle path-based references', () => {
    const pathRef = '/documents/handbook/employee-2025'

    cy.loginViaMagicLink(testUser)

    cy.visit(`/?doc=${encodeURIComponent(pathRef)}`)

    // Should create document
    cy.url({ timeout: 10000 }).should('include', '/?doc=')

    // Should be signable
    cy.get('[data-testid="sign-button"]').should('be.visible')
  })

  it('should handle simple docId references', () => {
    const simpleDocId = 'my-custom-doc-' + Date.now()

    cy.loginViaMagicLink(testUser)

    cy.visit(`/?doc=${simpleDocId}`)

    // Should use the same docId (no transformation)
    cy.url({ timeout: 10000 }).should('include', `/?doc=${simpleDocId}`)

    // Should create document
    cy.get('[data-testid="sign-button"]', { timeout: 10000 }).should('be.visible')
  })
})
