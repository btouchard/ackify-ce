// SPDX-License-Identifier: AGPL-3.0-or-later
/// <reference types="cypress" />

describe('Test 9: Complete End-to-End Workflow', () => {
  const adminEmail = 'admin@test.com'
  const alice = 'alice@test.com'
  const bob = 'bob@test.com'
  const charlie = 'charlie@test.com'
  const docId = 'policy-2025-' + Date.now()

  beforeEach(() => {
    cy.clearMailbox()
    cy.clearCookies()
  })

  it('should complete full document lifecycle: create → signers → reminders → signatures → completion', () => {
    // ===== STEP 1: Admin creates document =====
    cy.log('STEP 1: Admin creates document')
    cy.loginAsAdmin()
    cy.visit('/admin')

    cy.get('[data-testid="doc-url-input"]').type(docId)
    cy.get('[data-testid="submit-button"]').click()
    cy.url({ timeout: 10000 }).should('include', `/admin/docs/${docId}`)

    // ===== STEP 2: Admin adds 3 expected signers =====
    cy.log('STEP 2: Admin adds 3 expected signers')
    cy.get('[data-testid="open-add-signers-btn"]').click()
    cy.wait(500)
    cy.get('[data-testid="signers-textarea"]').type(`${alice}\n${bob}\n${charlie}`, { delay: 50 })
    cy.wait(300)
    cy.get('[data-testid="add-signers-btn"]').click()

    cy.contains(alice, { timeout: 10000 }).should('be.visible')
    cy.contains(bob).should('be.visible')
    cy.contains(charlie).should('be.visible')

    // Verify stats: 0/3 signed (0%)
    cy.contains('Confirmed').parent().should('contain', '0')
    cy.contains('Expected').parent().should('contain', '3')

    // ===== STEP 3: Admin sends reminders → 3 emails sent =====
    cy.log('STEP 3: Admin sends reminders to all signers')
    cy.clearMailbox()

    cy.contains('button', 'Send reminders').click()
    cy.contains('Send reminders', { timeout: 5000 }).should('be.visible')
    cy.get('[data-testid="confirm-button"]').click()

    cy.contains(/Reminder.*sent|sent successfully/, { timeout: 10000 }).should('be.visible')

    // Verify 3 emails sent (alice, bob, charlie)
    cy.waitForEmail(alice, 'Reminder', 15000).should('exist')
    cy.waitForEmail(bob, 'Reminder', 15000).should('exist')
    cy.waitForEmail(charlie, 'Reminder', 15000).should('exist')

    // ===== STEP 4: Alice logs in and signs =====
    cy.log('STEP 4: Alice signs the document')
    cy.logout()
    cy.loginViaMagicLink(alice, `/?doc=${docId}`)

    cy.url({ timeout: 10000 }).should('include', `/?doc=${docId}`)
    cy.confirmReading()
    cy.contains('Reading confirmed', { timeout: 10000 }).should('be.visible')

    // ===== STEP 5: Verify stats: 1/3 signed (33%) =====
    cy.log('STEP 5: Verify completion stats after Alice signs')
    cy.logout()
    cy.loginAsAdmin()
    cy.visit(`/admin/docs/${docId}`)

    cy.contains('Confirmed', { timeout: 10000 }).parent().should('contain', '1')
    cy.contains('Pending').parent().should('contain', '2')

    // ===== STEP 6: Bob logs in and signs =====
    cy.log('STEP 6: Bob signs the document')
    cy.logout()
    cy.loginViaMagicLink(bob, `/?doc=${docId}`)

    cy.url({ timeout: 10000 }).should('include', `/?doc=${docId}`)
    cy.confirmReading()
    cy.contains('Reading confirmed', { timeout: 10000 }).should('be.visible')

    // ===== STEP 7: Verify stats: 2/3 signed (66%) =====
    cy.log('STEP 7: Verify completion stats after Bob signs')
    cy.logout()
    cy.loginAsAdmin()
    cy.visit(`/admin/docs/${docId}`)

    cy.contains('Confirmed', { timeout: 10000 }).parent().should('contain', '2')
    cy.contains('Pending').parent().should('contain', '1')

    // ===== STEP 8: Admin sends new reminder → 1 email (charlie only) =====
    cy.log('STEP 8: Admin sends reminder to remaining signer')
    cy.clearMailbox()

    cy.contains('button', 'Send reminders').click()
    cy.contains('Send reminders', { timeout: 5000 }).should('be.visible')
    cy.get('[data-testid="confirm-button"]').click()

    cy.contains(/Reminder.*sent|sent successfully/, { timeout: 10000 }).should('be.visible')

    // Only Charlie should receive email
    cy.waitForEmail(charlie, 'Reminder', 15000).should('exist')

    // Alice and Bob should NOT receive new emails
    cy.request(`${Cypress.env('mailhogUrl')}/api/v2/messages?limit=50`).then((response) => {
      const messages = response.body.items || []
      const aliceReminder = messages.filter((msg: any) => {
        const recipients = msg.To || []
        return recipients.some((to: any) => `${to.Mailbox}@${to.Domain}` === alice) &&
               msg.Content?.Headers?.Subject?.[0]?.includes('Reminder')
      })
      const bobReminder = messages.filter((msg: any) => {
        const recipients = msg.To || []
        return recipients.some((to: any) => `${to.Mailbox}@${to.Domain}` === bob) &&
               msg.Content?.Headers?.Subject?.[0]?.includes('Reminder')
      })

      // Should have 0 new reminders for alice and bob
      expect(aliceReminder).to.have.length(0)
      expect(bobReminder).to.have.length(0)
    })

    // ===== STEP 9: Charlie signs =====
    cy.log('STEP 9: Charlie signs the document')
    cy.logout()
    cy.loginViaMagicLink(charlie, `/?doc=${docId}`)

    cy.url({ timeout: 10000 }).should('include', `/?doc=${docId}`)
    cy.confirmReading()
    cy.contains('Reading confirmed', { timeout: 10000 }).should('be.visible')

    // ===== STEP 10: Verify stats: 3/3 signed (100% completion) =====
    cy.log('STEP 10: Verify 100% completion')
    cy.logout()
    cy.loginAsAdmin()
    cy.visit(`/admin/docs/${docId}`)

    cy.contains('Confirmed', { timeout: 10000 }).parent().should('contain', '3')
    cy.contains('Expected').parent().should('contain', '3')

    // All signers should show "Confirmed" status
    cy.contains('tr', alice).should('contain', 'Confirmed')
    cy.contains('tr', bob).should('contain', 'Confirmed')
    cy.contains('tr', charlie).should('contain', 'Confirmed')

    // No pending signers
    cy.contains('Pending').parent().should('contain', '0')
  })
})
