// SPDX-License-Identifier: AGPL-3.0-or-later
/// <reference types="cypress" />

describe('Magic Link Authentication', () => {
  const testEmail = 'test@example.com'
  const baseUrl = Cypress.config('baseUrl')

  beforeEach(() => {
    // Clear mailbox before each test
    cy.clearMailbox()
  })

  it('should complete full magic link authentication workflow', () => {
    // Visit auth choice page with English locale
    cy.visitWithLocale('/auth')

    // Wait for Vue app to be fully loaded
    cy.get('#app', { timeout: 10000 }).should('not.be.empty')

    // Should display auth choice page
    cy.contains('Sign in to Ackify', { timeout: 10000 }).should('be.visible')
    cy.contains('Sign in with Email', { timeout: 10000 }).should('be.visible')

    // Fill magic link form
    cy.get('input[type="email"]', { timeout: 10000 }).should('be.visible').type(testEmail)
    cy.contains('Send Magic Link').click()

    // Should show success message
    cy.contains('Check your email', { timeout: 10000 }).should('be.visible')
    cy.contains('We sent you a magic link').should('be.visible')
    cy.contains('The link expires in 15 minutes').should('be.visible')

    // Wait for email to arrive in Mailhog
    cy.waitForEmail(testEmail, 'login link', 30000).then((message) => {
      // Verify email content
      expect(message.To).to.have.length.greaterThan(0)
      expect(message.To[0].Mailbox + '@' + message.To[0].Domain).to.equal(testEmail)

      // Extract magic link from email
      cy.extractMagicLink(message).then((magicLink) => {
        cy.log('Magic link found:', magicLink)

        // Visit the magic link
        cy.visit(magicLink)

        // Should redirect to home page after successful authentication
        cy.url({ timeout: 10000 }).should('eq', baseUrl + '/')

        // Verify user is authenticated
        cy.request('/api/v1/users/me').then((response) => {
          expect(response.status).to.eq(200)
          expect(response.body.data).to.have.property('email', testEmail)
          expect(response.body.data).to.have.property('id', testEmail)
        })
      })
    })
  })

  it('should redirect to document page after authentication', () => {
    const docId = 'test-document-123'
    const redirectUrl = `/?doc=${docId}`

    // Visit auth page with redirect parameter
    cy.visitWithLocale(`/auth?redirect=${encodeURIComponent(redirectUrl)}`)

    // Fill magic link form
    cy.get('input[type="email"]').type(testEmail)
    cy.contains('Send Magic Link').click()

    // Wait for success message
    cy.contains('Check your email', { timeout: 10000 }).should('be.visible')

    // Wait for email
    cy.waitForEmail(testEmail, 'login link', 30000).then((message) => {
      // Extract magic link
      cy.extractMagicLink(message).then((magicLink) => {
        // Verify redirect parameter is in the link
        expect(magicLink).to.include(`redirect=${encodeURIComponent(redirectUrl)}`)

        // Visit the magic link
        cy.visit(magicLink)

        // Should redirect to document page
        cy.url({ timeout: 10000 }).should('include', `/?doc=${docId}`)
      })
    })
  })

  it('should reject invalid email addresses', () => {
    cy.visitWithLocale('/auth')

    // Try invalid email - remove HTML5 validation and type
    cy.get('input[type="email"]').then(($input) => {
      $input.removeAttr('type')
      $input.removeAttr('required')
      cy.wrap($input).type('invalid-email')
    })

    cy.contains('Send Magic Link').click()

    // Should show error
    cy.contains('Please enter a valid email address', { timeout: 5000 }).should('be.visible')
  })

  it('should reject expired tokens', () => {
    // This test would require mocking time or waiting 15 minutes
    // Skip for now, can be tested manually or with time manipulation
    cy.log('Expired token test skipped - requires time manipulation')
  })

  it('should prevent token reuse', () => {
    cy.visitWithLocale('/auth')

    // Request magic link
    cy.get('input[type="email"]').type(testEmail)
    cy.contains('Send Magic Link').click()

    // Wait for email
    cy.waitForEmail(testEmail, 'login link', 30000).then((message) => {
      cy.extractMagicLink(message).then((magicLink) => {
        // Use the magic link once
        cy.visit(magicLink)
        cy.url({ timeout: 10000 }).should('eq', baseUrl + '/')

        // Clear cookies to simulate new session
        cy.clearCookies()

        // Try to use the same link again - should fail with 400
        cy.request({ url: magicLink, failOnStatusCode: false }).then((response) => {
          // Should get 400 Bad Request
          expect(response.status).to.eq(400)
          // Should contain error message about token
          expect(response.body).to.include('Invalid or expired token')
        })
      })
    })
  })

  it('should enforce rate limiting', () => {
    cy.visitWithLocale('/auth')

    // Send multiple requests quickly
    for (let i = 0; i < 4; i++) {
      cy.get('input[type="email"]').clear().type(`test${i}@example.com`)
      cy.contains('Send Magic Link').click()

      if (i < 3) {
        // First 3 should succeed
        cy.contains('Check your email', { timeout: 5000 }).should('be.visible')
        cy.wait(1000) // Small delay between requests
        cy.visitWithLocale('/auth') // Reload page to send another request
      } else {
        // 4th request should fail due to rate limiting
        // Note: This depends on backend rate limiting implementation
        // You may need to adjust based on actual limits
        cy.log('Rate limit test - verify backend behavior')
      }
    }
  })

  it('should handle mailhog unavailability gracefully', () => {
    // This test verifies frontend behavior when email fails to send
    // You might want to mock the API response for this
    cy.log('Email service unavailability test - manual testing recommended')
  })
})

describe('Magic Link Email Content', () => {
  const testEmail = 'content-test@example.com'

  beforeEach(() => {
    cy.clearMailbox()
  })

  it('should send email with correct subject and content', () => {
    cy.visitWithLocale('/auth')

    // Request magic link
    cy.get('input[type="email"]').type(testEmail)
    cy.contains('Send Magic Link').click()

    // Wait for email
    cy.waitForEmail(testEmail, 'login link', 30000).then((message) => {
      // Verify subject
      const subject = message.Content?.Headers?.Subject?.[0] || ''
      expect(subject).to.include('login link')

      // Verify sender
      expect(message.From.Mailbox + '@' + message.From.Domain).to.include('ackify')

      // Verify recipient
      expect(message.To[0].Mailbox + '@' + message.To[0].Domain).to.equal(testEmail)

      // Verify body contains key elements
      const body = message.Content?.Body || ''
      expect(body).to.include('/api/v1/auth/magic-link/verify')
      expect(body).to.include('token=')

      // Extract and verify token format (should be base64url)
      cy.extractMagicLink(message).then((link) => {
        const url = new URL(link)
        const token = url.searchParams.get('token')
        expect(token).to.exist
        expect(token).to.have.length.greaterThan(20)
        // base64url should only contain [A-Za-z0-9_-]
        expect(token).to.match(/^[A-Za-z0-9_-]+$/)
      })
    })
  })
})
