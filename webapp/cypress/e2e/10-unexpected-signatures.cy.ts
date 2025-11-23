// SPDX-License-Identifier: AGPL-3.0-or-later
/// <reference types="cypress" />

describe('Test 10: Unexpected Signatures Tracking', () => {
  const adminEmail = 'admin@test.com'
  const alice = 'alice@test.com'
  const bob = 'bob@test.com'
  const charlie = 'charlie-unexpected@test.com' // Not in expected list
  const docId = 'unexpected-test-' + Date.now()

  beforeEach(() => {
    cy.clearMailbox()
    cy.clearCookies()
  })

  it('should track unexpected signatures separately from expected ones', () => {
    // ===== STEP 1: Admin creates document with 2 expected signers =====
    cy.log('STEP 1: Create document with expected signers')
    cy.loginAsAdmin()
    cy.visit('/admin')

    cy.get('input#newDocId, input#newDocIdMobile').first().type(docId)
    cy.contains('button', 'Confirm').click()
    cy.url({ timeout: 10000 }).should('include', `/admin/docs/${docId}`)

    // Add alice and bob as expected signers
    cy.contains('button', 'Add').click()
    cy.wait(500)
    cy.get('textarea[placeholder*="Jane"]').type(`${alice}\n${bob}`, { delay: 50 })
    cy.wait(300)
    cy.get('button[type="submit"]').contains('Add').click()

    cy.contains(alice, { timeout: 10000 }).should('be.visible')
    cy.contains(bob).should('be.visible')

    // Verify initial stats
    cy.contains('Expected').parent().should('contain', '2')
    cy.contains('Signed').parent().should('contain', '0')

    // ===== STEP 2: Charlie (not expected) accesses and signs the document =====
    cy.log('STEP 2: Unexpected user (Charlie) signs the document')
    cy.logout()
    cy.loginViaMagicLink(charlie, `/?doc=${docId}`)

    cy.url({ timeout: 10000 }).should('include', `/?doc=${docId}`)

    // Charlie can still sign (document is not restricted)
    cy.contains('button', 'Confirm reading', { timeout: 10000 }).should('be.visible').click()
    cy.contains('Reading confirmed', { timeout: 10000 }).should('be.visible')

    // ===== STEP 3: Verify in admin that Charlie appears in "Unexpected Signatures" section =====
    cy.log('STEP 3: Verify unexpected signature appears in admin panel')
    cy.logout()
    cy.loginAsAdmin()
    cy.visit(`/admin/docs/${docId}`)

    // Expected signers section
    cy.contains(/Expected.*[Rr]eaders/, { timeout: 10000 }).should('be.visible')
    cy.contains(alice).should('be.visible')
    cy.contains(bob).should('be.visible')

    // Stats should show: 0/2 expected signed, but there's an unexpected signature
    cy.contains('Expected').parent().should('contain', '2')
    cy.contains('Signed').parent().should('contain', '0') // 0 expected signed

    // Unexpected signatures section should exist
    cy.contains(/Unexpected|Additional.*confirmations/, { timeout: 10000 }).should('be.visible')
    cy.contains(/Additional.*confirmations.*users.*not.*expected|not on the expected readers list/)
      .should('be.visible')

    // Charlie should appear in unexpected section
    cy.contains(charlie).should('be.visible')

    // Badge showing count of unexpected signatures
    cy.contains('1').should('be.visible')

    // ===== STEP 4: Alice (expected) signs =====
    cy.log('STEP 4: Expected user (Alice) signs')
    cy.logout()
    cy.loginViaMagicLink(alice, `/?doc=${docId}`)

    cy.url({ timeout: 10000 }).should('include', `/?doc=${docId}`)
    cy.contains('button', 'Confirm reading', { timeout: 10000 }).click()
    cy.contains('Reading confirmed', { timeout: 10000 }).should('be.visible')

    // ===== STEP 5: Verify stats update correctly =====
    cy.log('STEP 5: Verify stats with mix of expected and unexpected')
    cy.logout()
    cy.loginAsAdmin()
    cy.visit(`/admin/docs/${docId}`)

    // Expected stats: 1/2 signed (50%)
    cy.contains('Signed', { timeout: 10000 }).parent().should('contain', '1')
    cy.contains('Expected').parent().should('contain', '2')
    cy.contains('50%').should('be.visible')

    // Alice should show "Confirmed" in expected section
    cy.contains('tr', alice).should('contain', 'Confirmed')

    // Bob should show "Pending"
    cy.contains('tr', bob).should('contain', 'Pending')

    // Charlie should still be in unexpected section
    cy.contains(/Unexpected|Additional.*confirmations/).should('be.visible')
    cy.contains(charlie).should('be.visible')
  })

  it('should handle multiple unexpected signatures', () => {
    const multiDocId = 'multi-unexpected-' + Date.now()
    const expected1 = 'expected1@test.com'
    const unexpected1 = 'unexpected1@test.com'
    const unexpected2 = 'unexpected2@test.com'

    // Create document with 1 expected signer
    cy.loginAsAdmin()
    cy.visit('/admin')

    cy.get('input#newDocId, input#newDocIdMobile').first().type(multiDocId)
    cy.contains('button', 'Confirm').click()

    cy.url({ timeout: 10000 }).should('include', `/admin/docs/${multiDocId}`)

    cy.contains('button', 'Add').click()
    cy.wait(500)
    cy.get('textarea[placeholder*="Jane"]').type(expected1, { delay: 50 })
    cy.wait(300)
    cy.get('button[type="submit"]').contains('Add').click()

    cy.contains(expected1, { timeout: 10000 }).should('be.visible')

    // Two unexpected users sign
    cy.logout()
    cy.loginViaMagicLink(unexpected1, `/?doc=${multiDocId}`)
    cy.contains('button', 'Confirm reading', { timeout: 10000 }).click()

    cy.logout()
    cy.loginViaMagicLink(unexpected2, `/?doc=${multiDocId}`)
    cy.contains('button', 'Confirm reading', { timeout: 10000 }).click()

    // Verify both appear in unexpected section
    cy.logout()
    cy.loginAsAdmin()
    cy.visit(`/admin/docs/${multiDocId}`)

    cy.contains(/Unexpected|Additional.*confirmations/, { timeout: 10000 }).should('be.visible')
    cy.contains(unexpected1).should('be.visible')
    cy.contains(unexpected2).should('be.visible')

    // Badge should show 2
    cy.contains(/Unexpected|Additional.*confirmations/).parent().should('contain', '2')
  })
})
