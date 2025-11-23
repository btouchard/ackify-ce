// SPDX-License-Identifier: AGPL-3.0-or-later
/// <reference types="cypress" />

describe('Test 1: User Signature Workflow (MagicLink)', () => {
  const testEmail = 'user1@example.com'
  const docId = 'test-document-1'

  beforeEach(() => {
    cy.clearMailbox()
    cy.clearCookies()
  })

  it('should complete full signature workflow from document access to signature', () => {
    // Step 1: Login first via MagicLink
    cy.loginViaMagicLink(testEmail, `/?doc=${docId}`)

    // Step 2: Should be on document page after login
    cy.url({ timeout: 10000 }).should('include', `/?doc=${docId}`)

    // Step 3: Verify user is authenticated
    cy.request('/api/v1/users/me').then((response) => {
      expect(response.status).to.eq(200)
      expect(response.body.data.email).to.equal(testEmail)
    })

    // Step 4: Sign the document
    cy.contains('button', 'Confirm reading', { timeout: 10000 }).should('be.visible').click()

    // Step 5: Verify success message
    cy.contains('Reading confirmed', { timeout: 10000 }).should('be.visible')

    // Step 6: Verify signature appears in the list
    cy.contains('confirmation', { timeout: 5000 }).should('be.visible')
    cy.contains(testEmail).should('be.visible')

    // Step 7: Verify button is no longer present (already confirmed)
    cy.contains('button', 'Confirm reading').should('not.exist')
  })
})
