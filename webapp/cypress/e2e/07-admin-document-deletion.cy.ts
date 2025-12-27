// SPDX-License-Identifier: AGPL-3.0-or-later
/// <reference types="cypress" />

describe('Test 7: Admin - Document Deletion', () => {
  const adminEmail = 'admin@test.com'
  const testUser = 'deletetest@example.com'
  const docId = 'doc-to-delete-' + Date.now()

  beforeEach(() => {
    cy.clearMailbox()
    cy.clearCookies()
  })

  it('should soft-delete document with signatures', () => {
    // Step 1: Login as admin and create document
    cy.loginAsAdmin()
    cy.visit('/admin')

    cy.get('[data-testid="new-doc-input"]').type(docId)
    cy.contains('button', 'Confirm').click()
    cy.url({ timeout: 10000 }).should('include', `/admin/docs/${docId}`)

    // Step 2: Add 2 expected signers
    cy.get('[data-testid="open-add-signers-btn"]').click()
    cy.wait(500)
    cy.get('[data-testid="signers-textarea"]').type(`alice@test.com\n${testUser}`, { delay: 50 })
    cy.wait(300)
    cy.get('[data-testid="add-signers-btn"]').click()

    cy.contains('alice@test.com', { timeout: 10000 }).should('be.visible')

    // Step 3: Logout and login as test user
    cy.logout()
    cy.loginViaMagicLink(testUser, `/?doc=${docId}`)

    // Step 4: User signs the document
    cy.url({ timeout: 10000 }).should('include', `/?doc=${docId}`)
    cy.contains('button', 'Confirm reading', { timeout: 10000 }).click()
    cy.contains('Reading confirmed', { timeout: 10000 }).should('be.visible')

    // Step 5: Logout and login back as admin
    cy.logout()
    cy.loginAsAdmin()

    // Step 6: Navigate to document detail
    cy.visit(`/admin/docs/${docId}`)
    cy.url({ timeout: 10000 }).should('include', `/admin/docs/${docId}`)

    // Step 7: Delete the document
    cy.contains('Danger zone', { timeout: 10000 }).should('be.visible')
    cy.contains('button', 'Delete').click()

    // Step 8: Confirm deletion in modal
    cy.contains('This action is irreversible', { timeout: 5000 }).should('be.visible')
    cy.contains(docId).should('be.visible')
    cy.contains('button', 'Delete permanently').click()

    // Step 9: Should redirect to admin dashboard
    cy.url({ timeout: 10000 }).should('eq', Cypress.config('baseUrl') + '/admin')

    // Step 10: Document should not appear in active documents list
    cy.contains(docId).should('not.exist')

    // Step 11: Logout and login as test user
    cy.logout()
    cy.loginViaMagicLink(testUser)

    // Step 12: Check /signatures page
    cy.visit('/signatures')
    cy.url({ timeout: 10000 }).should('include', '/signatures')

    // Wait for page to load completely
    cy.contains('All my confirmations', { timeout: 10000 }).should('be.visible')

    // Step 13: Signature should be visible with "deleted" indicator
    cy.contains(docId, { timeout: 10000 }).should('be.visible')

    // Look for deleted documents section (should appear when there are deleted docs)
    cy.contains('Deleted documents', { timeout: 10000 }).should('be.visible')
  })

  it('should prevent deletion confirmation without proper warning acknowledgment', () => {
    cy.loginAsAdmin()
    cy.visit('/admin')

    const safeDocId = 'safe-doc-' + Date.now()
    cy.get('[data-testid="new-doc-input"]').type(safeDocId)
    cy.contains('button', 'Confirm').click()

    cy.url({ timeout: 10000 }).should('include', `/admin/docs/${safeDocId}`)

    // Try to delete
    cy.contains('button', 'Delete').click()

    // Modal should appear with warning
    cy.contains('This action is irreversible', { timeout: 5000 }).should('be.visible')

    // Cancel deletion
    cy.contains('button', 'Cancel').click()

    // Should still be on document page
    cy.url().should('include', `/admin/docs/${safeDocId}`)

    // Document should still exist
    cy.visit('/admin')
    cy.contains(safeDocId).should('be.visible')
  })
})
