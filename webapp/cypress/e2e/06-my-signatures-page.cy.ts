// SPDX-License-Identifier: AGPL-3.0-or-later
/// <reference types="cypress" />

describe('Test 6: My Signatures Page', () => {
  const testUser = 'mysigsuser@example.com'
  const doc1 = 'doc-alpha-' + Date.now()
  const doc2 = 'doc-beta-' + Date.now()
  const doc3 = 'doc-gamma-' + Date.now()

  beforeEach(() => {
    cy.clearMailbox()
    cy.clearCookies()
  })

  it('should display all user signatures with stats', () => {
    // Step 1: Login
    cy.loginViaMagicLink(testUser)

    // Step 2: Sign 3 different documents
    // Sign doc1
    cy.visit(`/?doc=${doc1}`)
    cy.confirmReading()
    cy.contains('Reading confirmed', { timeout: 10000 }).should('be.visible')

    // Sign doc2
    cy.visit(`/?doc=${doc2}`)
    cy.confirmReading()
    cy.contains('Reading confirmed', { timeout: 10000 }).should('be.visible')

    // Sign doc3
    cy.visit(`/?doc=${doc3}`)
    cy.confirmReading()
    cy.contains('Reading confirmed', { timeout: 10000 }).should('be.visible')

    // Step 3: Navigate to /signatures
    cy.visit('/signatures')
    cy.url({ timeout: 10000 }).should('include', '/signatures')

    // Step 4: Verify page title and stats cards
    cy.contains('My reading confirmations', { timeout: 10000 }).should('be.visible')
    cy.contains('Total confirmations').should('be.visible')
    cy.contains('Unique documents').should('be.visible')
    cy.contains('Last confirmation').should('be.visible')

    // Step 5: Verify all 3 signatures appear in list
    cy.contains(doc1).should('be.visible')
    cy.contains(doc2).should('be.visible')
    cy.contains(doc3).should('be.visible')
  })

  it('should allow searching signatures by docId', () => {
    // Login and sign documents
    cy.loginViaMagicLink(testUser)

    const searchDoc1 = 'search-foo-' + Date.now()
    const searchDoc2 = 'search-bar-' + Date.now()

    // Sign 2 documents
    cy.visit(`/?doc=${searchDoc1}`)
    cy.confirmReading()

    cy.visit(`/?doc=${searchDoc2}`)
    cy.confirmReading()

    // Navigate to signatures page
    cy.visit('/signatures')

    // Both should be visible initially
    cy.contains(searchDoc1, { timeout: 10000 }).should('be.visible')
    cy.contains(searchDoc2).should('be.visible')

    // Search for "foo"
    cy.get('input[placeholder*="Search"]').type('foo')

    // Only searchDoc1 should be visible
    cy.contains(searchDoc1).should('be.visible')
    cy.contains(searchDoc2).should('not.exist')

    // Clear search
    cy.get('input[placeholder*="Search"]').clear()

    // Both should be visible again
    cy.contains(searchDoc1).should('be.visible')
    cy.contains(searchDoc2).should('be.visible')
  })

  it('should show empty state when no signatures', () => {
    const newUser = 'nosigs-' + Date.now() + '@example.com'

    cy.loginViaMagicLink(newUser)

    cy.visit('/signatures')

    // Should show 0 confirmations in stats
    cy.contains('Total confirmations', { timeout: 10000 }).should('be.visible')
    cy.contains('0 results').should('be.visible')
  })

  it('should require authentication', () => {
    // Try to access without login
    cy.visit('/signatures')

    // Should redirect to auth
    cy.url({ timeout: 10000 }).should('include', '/auth')
  })
})
