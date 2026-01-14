// SPDX-License-Identifier: AGPL-3.0-or-later
/// <reference types="cypress" />

describe('Test 8: Admin Route Protection', () => {
  const regularUser = 'regular@example.com'
  const adminEmail = 'admin@test.com'
  const testDocId = 'protected-doc-' + Date.now()

  beforeEach(() => {
    cy.clearMailbox()
    cy.clearCookies()
  })

  it('should redirect non-admin users from /admin', () => {
    // Step 1: Login as regular user
    cy.loginViaMagicLink(regularUser)

    // Step 2: Try to access /admin
    cy.visit('/admin')

    // Step 3: Should redirect to home page
    cy.url({ timeout: 10000 }).should('eq', Cypress.config('baseUrl') + '/')

    // Step 4: Should not see admin content
    cy.contains('Administration').should('not.exist')
  })

  it('should redirect non-admin users from /admin/docs/:docId', () => {
    // Login as regular user
    cy.loginViaMagicLink(regularUser)

    // Try to access admin document detail page
    cy.visit(`/admin/docs/${testDocId}`)

    // Should redirect to home
    cy.url({ timeout: 10000 }).should('eq', Cypress.config('baseUrl') + '/')
  })

  it('should allow admin users to access /admin', () => {
    // Step 1: Login as admin
    cy.loginAsAdmin()

    // Step 2: Access /admin
    cy.visit('/admin')

    // Step 3: Should stay on admin page
    cy.url({ timeout: 10000 }).should('include', '/admin')

    // Step 4: Should see admin dashboard
    cy.contains('Administration', { timeout: 10000 }).should('be.visible')
    cy.get('[data-testid="doc-url-input"]').should('be.visible')
  })

  it('should redirect unauthenticated users to auth page', () => {
    // Ensure we're completely logged out
    cy.clearCookies()
    cy.clearLocalStorage()

    // Don't login - just try to access admin
    cy.visit('/admin')

    // Should redirect to auth (this is the key verification)
    cy.url({ timeout: 10000 }).should('include', '/auth')

    // Verify we're not redirected elsewhere (e.g., not at /admin)
    cy.url().should('not.include', '/admin/docs')
  })

  it('should preserve redirect after login', () => {
    const targetDoc = 'redirect-test-' + Date.now()

    // Create document first as admin
    cy.loginAsAdmin()
    cy.visit('/admin')
    cy.get('[data-testid="doc-url-input"]').type(targetDoc)
    cy.get('[data-testid="submit-button"]').click()
    cy.url({ timeout: 10000 }).should('include', `/admin/docs/${targetDoc}`)

    // Logout
    cy.logout()

    // Now login as admin again with redirect parameter
    cy.loginViaMagicLink(adminEmail, `/admin/docs/${targetDoc}`)

    // Should be redirected to the target document
    cy.url({ timeout: 10000 }).should('include', `/admin/docs/${targetDoc}`)
    cy.contains('Document', { timeout: 10000 }).should('be.visible')
  })

  it('should block API access to admin endpoints for non-admin users', () => {
    // Login as regular user
    cy.loginViaMagicLink(regularUser)

    // Try to access admin API endpoint
    cy.request({
      url: '/api/v1/admin/documents',
      failOnStatusCode: false
    }).then((response) => {
      // Should return 403 Forbidden
      expect(response.status).to.eq(403)
    })
  })

  it('should allow API access to admin endpoints for admin users', () => {
    // Login as admin
    cy.loginAsAdmin()

    // Access admin API endpoint
    cy.request('/api/v1/admin/documents').then((response) => {
      // Should succeed
      expect(response.status).to.eq(200)
      expect(response.body.data).to.be.an('array')
    })
  })
})
