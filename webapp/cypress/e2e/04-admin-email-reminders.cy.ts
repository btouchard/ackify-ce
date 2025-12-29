// SPDX-License-Identifier: AGPL-3.0-or-later
/// <reference types="cypress" />

describe('Test 4: Admin - Email Reminders', () => {
  const adminEmail = 'admin@test.com'
  const docId = 'test-reminders-' + Date.now()
  const alice = 'alice@test.com'
  const bob = 'bob@test.com'

  beforeEach(() => {
    cy.clearMailbox()
    cy.clearCookies()
  })

  it('should send reminders only to unsigned users', () => {
    // Step 1: Login as admin and create document
    cy.loginAsAdmin()
    cy.visit('/admin')

    cy.get('[data-testid="admin-new-doc-input"]').type(docId)
    cy.get('[data-testid="admin-create-doc-btn"]').click()
    cy.url({ timeout: 10000 }).should('include', `/admin/docs/${docId}`)

    // Step 2: Add 2 expected signers
    cy.get('[data-testid="add-signers-btn"]').click()

    // Wait for modal
    cy.wait(500)

    cy.get('[data-testid="add-signers-textarea"]').type(`${alice}{enter}${bob}`, { delay: 50 })

    // Wait for Vue reactivity
    cy.wait(300)

    // Submit form
    cy.get('[data-testid="add-signers-submit"]').click()

    // Verify signers added
    cy.contains(alice, { timeout: 10000 }).should('be.visible')
    cy.contains(bob).should('be.visible')

    // Step 3: Logout admin and login as Alice
    cy.logout()
    cy.loginViaMagicLink(alice, `/?doc=${docId}`)

    // Step 4: Alice signs the document
    cy.url({ timeout: 10000 }).should('include', `/?doc=${docId}`)
    cy.get('[data-testid="sign-button"]', { timeout: 10000 }).should('be.visible').click()
    cy.get('[data-testid="sign-success"]', { timeout: 10000 }).should('be.visible')

    // Step 5: Logout Alice and login back as admin
    cy.logout()
    cy.loginAsAdmin()

    // Step 6: Navigate to document detail page
    cy.visit(`/admin/docs/${docId}`)
    cy.url({ timeout: 10000 }).should('include', `/admin/docs/${docId}`)

    // Step 7: Verify stats (1 signed, 1 pending)
    cy.contains('Signed').parent().should('contain', '1')
    cy.contains('Pending').parent().should('contain', '1')

    // Step 8: Send reminders to all pending
    cy.clearMailbox() // Clear previous emails

    // Click send reminders button
    cy.get('[data-testid="send-reminders-btn"]').click()

    // Confirm in modal
    cy.contains('Send reminders', { timeout: 5000 }).should('be.visible')
    cy.contains('button', 'Confirm').click({ force: true })

    // Step 9: Wait for API call to complete
    cy.wait(2000)

    // Step 10: Verify email sent to Bob only (not Alice)
    cy.waitForEmail(bob, 'Reminder', 15000).then((message) => {
      expect(message.To[0].Mailbox + '@' + message.To[0].Domain).to.equal(bob)

      // Verify email content contains document link
      const body = message.Content?.Body || ''
      expect(body).to.include(docId)
    })

    // Step 11: Verify no email sent to Alice (already signed)
    cy.request(`${Cypress.env('mailhogUrl')}/api/v2/messages?limit=50`).then((response) => {
      const messages = response.body.items || []
      const aliceEmail = messages.find((msg: any) => {
        const recipients = msg.To || []
        return recipients.some((to: any) => `${to.Mailbox}@${to.Domain}` === alice) &&
               msg.Content?.Headers?.Subject?.[0]?.includes('Reminder')
      })
      expect(aliceEmail).to.be.undefined
    })

    // Test passed - reminders sent only to pending users
  })
})
