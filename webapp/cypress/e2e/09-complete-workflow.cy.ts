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

    cy.get('input#newDocId, input#newDocIdMobile').first().type(docId)
    cy.contains('button', 'Confirm').click()
    cy.url({ timeout: 10000 }).should('include', `/admin/docs/${docId}`)

    // ===== STEP 2: Admin adds 3 expected signers =====
    cy.log('STEP 2: Admin adds 3 expected signers')
    cy.contains('button', 'Add').click()
    cy.wait(500)
    cy.get('textarea[placeholder*="Jane"]').type(`${alice}\n${bob}\n${charlie}`, { delay: 50 })
    cy.wait(300)
    cy.get('button[type="submit"]').contains('Add').click()

    cy.contains(alice, { timeout: 10000 }).should('be.visible')
    cy.contains(bob).should('be.visible')
    cy.contains(charlie).should('be.visible')

    // Verify stats: 0/3 signed (0%)
    cy.contains('Signed').parent().should('contain', '0')
    cy.contains('Expected').parent().should('contain', '3')

    // ===== STEP 3: Admin sends reminders → 3 emails sent =====
    cy.log('STEP 3: Admin sends reminders to all signers')
    cy.clearMailbox()

    cy.contains('button', 'Send reminders').click()
    cy.contains('This action is irreversible', { timeout: 5000 }).should('be.visible')
    cy.contains('button', 'Confirm').click()

    cy.contains(/Reminder.*sent|sent successfully/, { timeout: 10000 }).should('be.visible')

    // Verify 3 emails sent (alice, bob, charlie)
    cy.waitForEmail(alice, 'Reminder', 15000).should('exist')
    cy.waitForEmail(bob, 'Reminder', 15000).should('exist')
    cy.waitForEmail(charlie, 'Reminder', 15000).should('exist')

    // ===== STEP 4: Alice logs in from email and signs =====
    cy.log('STEP 4: Alice signs the document')
    cy.logout()

    // Get Alice's reminder email and extract auth link
    cy.waitForEmail(alice, 'Reminder', 15000).then((message) => {
      let body = message.Content?.Body || ''

      // Decode quoted-printable encoding first (=\r\n are soft line breaks, =XX are hex)
      body = body
        .replace(/=\r?\n/g, '') // Remove soft line breaks
        .replace(/=([0-9A-F]{2})/g, (_, hex) => String.fromCharCode(parseInt(hex, 16)))

      // Extract reminder auth link from email (reminder-link/verify)
      const linkMatch = body.match(/(https?:\/\/[^\s"<]+\/api\/v1\/auth\/reminder-link\/verify\?token=[^\s"<]+)/)
      expect(linkMatch).to.not.be.null

      const authLink = linkMatch![1]

      cy.log('Auth link from email:', authLink)

      // Visit auth link directly (it will authenticate and redirect to document)
      cy.visit(authLink)

      // Should redirect to document
      cy.url({ timeout: 10000 }).should('include', `/?doc=${docId}`)

      // Alice signs
      cy.contains('button', 'Confirm reading', { timeout: 10000 }).click()
      cy.contains('Reading confirmed', { timeout: 10000 }).should('be.visible')
    })

    // ===== STEP 5: Verify stats: 1/3 signed (33%) =====
    cy.log('STEP 5: Verify completion stats after Alice signs')
    cy.logout()
    cy.loginAsAdmin()
    cy.visit(`/admin/docs/${docId}`)

    cy.contains('Signed', { timeout: 10000 }).parent().should('contain', '1')
    cy.contains('Pending').parent().should('contain', '2')
    cy.contains('33%').should('be.visible') // 33% completion

    // ===== STEP 6: Bob logs in and signs =====
    cy.log('STEP 6: Bob signs the document')
    cy.logout()
    cy.loginViaMagicLink(bob, `/?doc=${docId}`)

    cy.url({ timeout: 10000 }).should('include', `/?doc=${docId}`)
    cy.contains('button', 'Confirm reading', { timeout: 10000 }).click()
    cy.contains('Reading confirmed', { timeout: 10000 }).should('be.visible')

    // ===== STEP 7: Verify stats: 2/3 signed (66%) =====
    cy.log('STEP 7: Verify completion stats after Bob signs')
    cy.logout()
    cy.loginAsAdmin()
    cy.visit(`/admin/docs/${docId}`)

    cy.contains('Signed', { timeout: 10000 }).parent().should('contain', '2')
    cy.contains('Pending').parent().should('contain', '1')
    cy.contains('67%').should('be.visible') // 67% completion (2/3 rounded)

    // ===== STEP 8: Admin sends new reminder → 1 email (charlie only) =====
    cy.log('STEP 8: Admin sends reminder to remaining signer')
    cy.clearMailbox()

    cy.contains('button', 'Send reminders').click()
    cy.contains('This action is irreversible', { timeout: 5000 }).should('be.visible')
    cy.contains('button', 'Confirm').click()

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
    cy.contains('button', 'Confirm reading', { timeout: 10000 }).click()
    cy.contains('Reading confirmed', { timeout: 10000 }).should('be.visible')

    // ===== STEP 10: Verify stats: 3/3 signed (100% completion) =====
    cy.log('STEP 10: Verify 100% completion')
    cy.logout()
    cy.loginAsAdmin()
    cy.visit(`/admin/docs/${docId}`)

    cy.contains('Signed', { timeout: 10000 }).parent().should('contain', '3')
    cy.contains('Expected').parent().should('contain', '3')
    cy.contains('100%').should('be.visible') // 100% completion

    // All signers should show "Confirmed" status
    cy.contains('tr', alice).should('contain', 'Confirmed')
    cy.contains('tr', bob).should('contain', 'Confirmed')
    cy.contains('tr', charlie).should('contain', 'Confirmed')

    // No pending signers
    cy.contains('Pending').parent().should('contain', '0')
  })
})
