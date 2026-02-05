// SPDX-License-Identifier: AGPL-3.0-or-later
/// <reference types="cypress" />

describe('Test 2: Signature Uniqueness Verification', () => {
  const testEmail = 'unique-user@example.com'

  beforeEach(() => {
    cy.clearMailbox()
    cy.clearCookies()
  })

  it('should enforce one signature per user per document', () => {
    const docId = 'test-unique-doc-' + Date.now()

    // Step 1: Login via MagicLink
    cy.loginViaMagicLink(testEmail, `/?doc=${docId}`)

    // Step 2: Sign the document
    cy.url({ timeout: 10000 }).should('include', `/?doc=${docId}`)
    cy.confirmReading()

    // Step 3: Verify success and button changes
    cy.contains('Reading confirmed', { timeout: 10000 }).should('be.visible')
    cy.contains('Confirmed').should('be.visible')

    // Step 4: Verify signature in list (user can see their own signature)
    cy.contains('Existing confirmations').should('be.visible')
    cy.contains(testEmail).should('be.visible') // User sees their own signature

    // Step 5: Reload the page
    cy.reload()

    // Step 6: Verify "already signed" status persists
    cy.contains('Confirmed', { timeout: 10000 }).should('be.visible')
    cy.contains('button', 'Confirm reading').should('not.exist')

    // Step 7: Verify signature still in list
    cy.contains(testEmail).should('be.visible')
  })

  it('should prevent duplicate signature via API', () => {
    const docId = 'test-unique-api-' + Date.now()

    // Login and sign
    cy.loginViaMagicLink(testEmail, `/?doc=${docId}`)
    cy.url({ timeout: 10000 }).should('include', `/?doc=${docId}`)

    // Get CSRF token
    cy.request('/api/v1/csrf').then((csrfResponse) => {
      const csrfToken = csrfResponse.body.data.token

      // First signature - should succeed
      cy.request({
        method: 'POST',
        url: '/api/v1/signatures',
        headers: {
          'X-CSRF-Token': csrfToken
        },
        body: {
          docId: docId,
          referer: window.location.href
        }
      }).then((response) => {
        expect(response.status).to.eq(201)
      })

      // Second signature attempt - should fail with 409 Conflict
      cy.request({
        method: 'POST',
        url: '/api/v1/signatures',
        headers: {
          'X-CSRF-Token': csrfToken
        },
        body: {
          docId: docId,
          referer: window.location.href
        },
        failOnStatusCode: false
      }).then((response) => {
        expect(response.status).to.eq(409)
        expect(response.body.error.message).to.include('already signed')
      })
    })
  })
})
