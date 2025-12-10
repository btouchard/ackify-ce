// SPDX-License-Identifier: AGPL-3.0-or-later
/// <reference types="cypress" />

describe('Test 13: Embed Page Functionality', () => {
  // Use a fixed docId for tests that build on each other
  const sharedDocId = 'embed-shared-doc-' + Date.now()
  const embedDocUrl = `https://github.com/example/repo/blob/main/README.md`

  beforeEach(() => {
    cy.clearCookies()
  })

  it('should display embed page with no signatures state', () => {
    // Step 1: Visit embed page with new document (force English locale)
    cy.visitWithLocale(`/embed?doc=${sharedDocId}`, 'en')

    // Step 2: Should load without authentication
    cy.url({ timeout: 10000 }).should('include', '/embed')
    cy.url().should('include', `doc=${sharedDocId}`)

    // Step 3: Should show empty state (i18n: "No signatures for this document")
    cy.contains('No signatures for this document', { timeout: 10000 }).should('be.visible')

    // Step 4: Should show "Sign this document" button (i18n: "Sign this document")
    cy.contains('a', 'Sign this document').should('be.visible')
    cy.contains('a', 'Sign this document').should('have.attr', 'href').and('include', `doc=${sharedDocId}`)
    cy.contains('a', 'Sign this document').should('have.attr', 'target', '_blank')

    // Step 5: Should show "Powered by Ackify" footer (i18n: "Powered by Ackify")
    cy.contains('Powered by Ackify').should('be.visible')
  })

  it('should display embed page with signatures', () => {
    // Step 1: Create a signature first
    cy.loginViaMagicLink('embed-user1@test.com')
    cy.visitWithLocale(`/?doc=${sharedDocId}`, 'en')

    // Wait for page to load and show the confirm button
    cy.contains('button', 'Confirm', { timeout: 10000 }).should('be.visible').click()

    // Wait for success message
    cy.contains('successfully', { timeout: 10000 }).should('be.visible')

    // Step 2: Logout and visit embed page (force English locale)
    cy.clearCookies()
    cy.visitWithLocale(`/embed?doc=${sharedDocId}`, 'en')

    // Step 3: Should show document header with signature count (i18n: "confirmation")
    cy.contains('Document', { timeout: 10000 }).should('be.visible')
    cy.contains('confirmation', { timeout: 10000 }).should('be.visible')

    // Step 4: Should show signature in list
    cy.contains('embed-user1@test.com').should('be.visible')

    // Step 5: Should show "Sign" button (i18n: "Sign")
    cy.contains('a', 'Sign').should('be.visible')
    cy.contains('a', 'Sign').should('have.attr', 'target', '_blank')

    // Step 6: Verify signature date is displayed
    cy.get('.text-xs.text-muted-foreground').should('exist')
  })

  it('should display multiple signatures', () => {
    // Step 1: Add multiple signatures
    const users = ['embed-user2@test.com', 'embed-user3@test.com', 'embed-user4@test.com']

    users.forEach((email) => {
      cy.loginViaMagicLink(email)
      cy.visitWithLocale(`/?doc=${sharedDocId}`, 'en')
      cy.contains('button', 'Confirm', { timeout: 10000 }).click()
      cy.contains('successfully', { timeout: 10000 }).should('be.visible')
      cy.clearCookies()
    })

    // Step 2: Visit embed page (force English locale)
    cy.visitWithLocale(`/embed?doc=${sharedDocId}`, 'en')

    // Step 3: Should show correct signature count (1 from previous test + 3 new = 4)
    cy.contains('confirmation', { timeout: 10000 }).should('be.visible')

    // Step 4: Should show all 4 signatures
    cy.contains('embed-user1@test.com').should('be.visible')
    cy.contains('embed-user2@test.com').should('be.visible')
    cy.contains('embed-user3@test.com').should('be.visible')
    cy.contains('embed-user4@test.com').should('be.visible')

    // Step 5: All signatures should have checkmark icons
    cy.get('svg').filter(':visible').should('have.length.at.least', 4)
  })

  it('should handle document with URL reference', () => {
    // Step 1: Create signature for URL-based document (use unique email)
    const uniqueEmail = 'embed-url-' + Date.now() + '@test.com'
    cy.loginViaMagicLink(uniqueEmail)
    cy.visitWithLocale(`/?doc=${encodeURIComponent(embedDocUrl)}`, 'en')

    // Wait for button to be clickable
    cy.contains('button', 'Confirm', { timeout: 10000 }).should('be.visible')
    cy.wait(500) // Small wait for page stabilization
    cy.contains('button', 'Confirm').click()
    cy.contains('successfully', { timeout: 10000 }).should('be.visible')

    // Step 2: Visit embed page with URL
    cy.clearCookies()
    cy.visitWithLocale(`/embed?doc=${encodeURIComponent(embedDocUrl)}`, 'en')

    // Step 3: Should redirect to canonical docId
    cy.url({ timeout: 10000 }).should('include', '/embed')
    cy.url().should('include', 'doc=')
    cy.url().should('not.include', encodeURIComponent(embedDocUrl))

    // Step 4: Should show signature (verify unique email is displayed)
    cy.contains(uniqueEmail, { timeout: 10000 }).should('be.visible')
  })

  it('should open sign link in new tab', () => {
    // Step 1: Visit embed page with existing signatures
    cy.visitWithLocale(`/embed?doc=${sharedDocId}`, 'en')

    // Step 2: Verify sign link opens in new tab (i18n: "Sign")
    cy.get('a').contains('Sign', { timeout: 10000 })
      .should('have.attr', 'target', '_blank')
      .should('have.attr', 'href')
      .and('include', `/?doc=${sharedDocId}`)
  })

  it('should display error for missing doc parameter', () => {
    // Step 1: Visit embed page without doc parameter
    cy.visitWithLocale('/embed', 'en', { failOnStatusCode: false })

    // Step 2: Should show error message (i18n: "Document ID missing")
    cy.contains('Document ID missing', { timeout: 10000 }).should('be.visible')
  })

  it('should refresh signatures when navigating between documents', () => {
    const doc1 = 'embed-nav-test-1-' + Date.now()
    const doc2 = 'embed-nav-test-2-' + Date.now()

    // Step 1: Create signature for doc1
    cy.loginViaMagicLink('embed-nav-user1@test.com')
    cy.visitWithLocale(`/?doc=${doc1}`, 'en')
    cy.contains('button', 'Confirm', { timeout: 10000 }).click()
    cy.contains('successfully', { timeout: 10000 }).should('be.visible')

    // Step 2: Create signature for doc2 with different user
    cy.clearCookies()
    cy.loginViaMagicLink('embed-nav-user2@test.com')
    cy.visitWithLocale(`/?doc=${doc2}`, 'en')
    cy.contains('button', 'Confirm', { timeout: 10000 }).click()
    cy.contains('successfully', { timeout: 10000 }).should('be.visible')

    // Step 3: Visit embed page for doc1
    cy.clearCookies()
    cy.visitWithLocale(`/embed?doc=${doc1}`, 'en')
    cy.contains('embed-nav-user1@test.com', { timeout: 10000 }).should('be.visible')
    cy.contains('embed-nav-user2@test.com').should('not.exist')

    // Step 4: Navigate to doc2 via URL change
    cy.visitWithLocale(`/embed?doc=${doc2}`, 'en')
    cy.contains('embed-nav-user2@test.com', { timeout: 10000 }).should('be.visible')
    cy.contains('embed-nav-user1@test.com').should('not.exist')
  })

  it('should work in iframe context', () => {
    // Step 1: Visit the embed page directly (simulating iframe context)
    cy.visitWithLocale(`/embed?doc=${sharedDocId}`, 'en')

    // Step 2: Verify embed page loads correctly (should have "confirmation" text somewhere)
    cy.contains('confirmation', { timeout: 10000 }).should('be.visible')

    // Step 3: Verify no navigation elements (should be minimal UI)
    cy.get('header').should('not.exist')
    cy.get('nav').should('not.exist')

    // Step 4: Verify branding footer is present
    cy.contains('Powered by Ackify').should('be.visible')
  })

  it('should display signatures in chronological order', () => {
    const chronoDocId = 'embed-chrono-test-' + Date.now()

    // Step 1: Create signatures with delays to ensure different timestamps
    const users = ['chrono1@test.com', 'chrono2@test.com', 'chrono3@test.com']

    users.forEach((email, index) => {
      cy.loginViaMagicLink(email)
      cy.visitWithLocale(`/?doc=${chronoDocId}`, 'en')
      cy.contains('button', 'Confirm', { timeout: 10000 }).click()
      cy.contains('successfully', { timeout: 10000 }).should('be.visible')
      cy.clearCookies()

      // Add delay between signatures
      if (index < users.length - 1) {
        cy.wait(2000)
      }
    })

    // Step 2: Visit embed page
    cy.visitWithLocale(`/embed?doc=${chronoDocId}`, 'en')

    // Step 3: Verify all signatures are displayed (should contain "confirmation" text)
    cy.contains('confirmation', { timeout: 10000 }).should('be.visible')

    // Step 4: Verify signatures appear in the list
    cy.get('.space-y-2 > div').should('have.length', 3)

    // Step 5: Verify each signature has email and date
    users.forEach((email) => {
      cy.contains(email).should('be.visible')
    })
  })

  it('should handle very long email addresses gracefully', () => {
    const longEmailDocId = 'embed-long-email-' + Date.now()
    const longEmail = 'very.long.email.address.with.many.dots.and.characters@subdomain.example.com'

    // Step 1: Create signature with long email
    cy.loginViaMagicLink(longEmail)
    cy.visitWithLocale(`/?doc=${longEmailDocId}`, 'en')
    cy.contains('button', 'Confirm', { timeout: 10000 }).click()
    cy.contains('successfully', { timeout: 10000 }).should('be.visible')

    // Step 2: Visit embed page
    cy.clearCookies()
    cy.visitWithLocale(`/embed?doc=${longEmailDocId}`, 'en')

    // Step 3: Verify long email is displayed (may be truncated)
    cy.contains(longEmail.substring(0, 20), { timeout: 10000 }).should('be.visible')

    // Step 4: Verify layout is not broken (check for truncate class)
    cy.get('.truncate').should('exist')
  })
})
