// SPDX-License-Identifier: AGPL-3.0-or-later
/// <reference types="cypress" />

describe('Test 3: Admin - Expected Signers Management', () => {
  const adminEmail = 'admin@test.com'
  const docId = 'test-admin-doc-' + Date.now()

  beforeEach(() => {
    cy.clearMailbox()
    cy.clearCookies()
  })

  it('should create document and manage expected signers', () => {
    // Step 1: Login as admin
    cy.loginAsAdmin()

    // Step 2: Navigate to admin dashboard
    cy.visit('/admin')
    cy.url({ timeout: 10000 }).should('include', '/admin')
    cy.contains('Administration', { timeout: 10000 }).should('be.visible')

    // Step 3: Create new document
    cy.get('[data-testid="doc-url-input"]').type(docId)
    cy.get('[data-testid="submit-button"]').click()

    // Step 4: Should redirect to document detail page
    cy.url({ timeout: 10000 }).should('include', `/admin/docs/${docId}`)
    cy.contains('Document').should('be.visible')

    // Step 5: Add 3 expected signers
    cy.get('[data-testid="open-add-signers-btn"]').click()

    // Modal should appear
    cy.get('[data-testid="add-signers-modal"]').should('be.visible')

    // Wait for modal to be fully rendered
    cy.wait(500)

    // Add signers (Name <email> format and plain email)
    cy.get('[data-testid="signers-textarea"]').type(
      'Alice Smith <alice@test.com>{enter}bob@test.com{enter}Charlie Brown <charlie@test.com>',
      { delay: 50 }
    )

    // Wait a bit for Vue reactivity
    cy.wait(300)

    // Submit the form
    cy.get('[data-testid="add-signers-btn"]').click()

    // Wait for modal to close
    cy.get('[data-testid="add-signers-modal"]', { timeout: 15000 }).should('not.exist')

    // Step 6: Verify signers in table
    cy.contains('alice@test.com', { timeout: 10000 }).should('be.visible')
    cy.contains('bob@test.com').should('be.visible')
    cy.contains('charlie@test.com').should('be.visible')

    // Step 7: Verify all have "Pending" status
    cy.get('table tbody tr').should('have.length.at.least', 3)
    cy.contains('Pending').should('be.visible')

    // Step 8: Verify stats
    cy.contains('Expected').should('be.visible')
    cy.contains('3').should('be.visible') // 3 expected signers
    cy.contains('Confirmed').parent().should('contain', '0') // 0 confirmed
  })

  it('should allow admin to remove expected signer', () => {
    // Login as admin and create document with signers
    cy.loginAsAdmin()
    cy.visit('/admin')

    // Create document
    const removeDocId = 'test-remove-signer-' + Date.now()
    cy.get('[data-testid="doc-url-input"]').type(removeDocId)
    cy.get('[data-testid="submit-button"]').click()

    cy.url({ timeout: 10000 }).should('include', `/admin/docs/${removeDocId}`)

    // Add 2 signers
    cy.get('[data-testid="open-add-signers-btn"]').click()

    // Wait for modal to be fully rendered
    cy.wait(500)

    cy.get('[data-testid="signers-textarea"]').type('alice@test.com{enter}bob@test.com', { delay: 50 })

    // Wait a bit for Vue reactivity
    cy.wait(300)

    // Submit the form
    cy.get('[data-testid="add-signers-btn"]').click()

    // Wait for modal to close
    cy.get('[data-testid="add-signers-modal"]', { timeout: 15000 }).should('not.exist')

    // Verify 2 signers
    cy.contains('alice@test.com', { timeout: 10000 }).should('be.visible')
    cy.contains('bob@test.com').should('be.visible')

    // Scroll to signers table to ensure visibility
    cy.contains('alice@test.com').scrollIntoView()

    // Remove alice - find the delete button in the row
    cy.contains('tr', 'alice@test.com')
      .find('button')
      .filter(':has(svg)')
      .first()
      .click()

    // Confirm removal in modal
    cy.contains('Remove expected reader', { timeout: 5000 }).should('be.visible')

    // Wait for modal animation
    cy.wait(500)

    // Find the specific confirmation text mentioning alice@test.com, then find Delete button nearby
    cy.contains('Remove alice@test.com').should('be.visible')

    // Click the Delete button in the ConfirmDialog modal only (not the Delete Document modal)
    // Find the modal overlay that contains the removal message, then click Delete within it
    cy.get('.fixed')
      .filter(':contains("Remove alice@test.com")')
      .within(() => {
        cy.get('button').contains('Delete').click()
      })

    // Wait for modal to close
    cy.contains('Remove expected reader', { timeout: 5000 }).should('not.exist')

    // Wait for API call to complete and list to refresh
    cy.wait(2000)

    // Verify alice removed (check that the row no longer exists)
    cy.get('table').within(() => {
      cy.contains('alice@test.com').should('not.exist')
    })
    cy.contains('bob@test.com').should('be.visible')
  })
})
