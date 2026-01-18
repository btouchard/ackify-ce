// SPDX-License-Identifier: AGPL-3.0-or-later
/// <reference types="cypress" />

/**
 * Demo video script for Ackify
 * This test is designed to be recorded as a product demonstration video.
 * It showcases the main features with appropriate pauses for readability.
 *
 * Features demonstrated:
 * - PDF upload with integrated viewer
 * - Require full read before signature
 * - MagicLink authentication
 * - Signature workflow
 * - Admin management (signers, reminders, status)
 * - Signature history
 */

const PAUSE_SHORT = 1200   // Brief pause between actions
const PAUSE_MEDIUM = 1800  // Pause to let user see result
const PAUSE_LONG = 2500    // Pause for important moments
const PAUSE_XLONG = 3500   // Extra long for key features

describe('Ackify Demo Video', () => {
  const adminEmail = 'admin@test.com'
  const alice = 'alice@demo.com'
  const bob = 'bob@demo.com'
  let docId: string

  before(() => {
    // Generate unique docId for this demo run
    docId = 'policy-' + Date.now()
  })

  beforeEach(() => {
    cy.clearMailbox()
    cy.clearCookies()
  })

  it('Part 1: Admin uploads PDF document with full-read requirement', () => {
    cy.log('**SCENE 1: Admin Login**')
    cy.loginAsAdmin()
    cy.wait(PAUSE_MEDIUM)

    cy.log('**SCENE 2: Navigate to Documents page**')
    cy.visit('/documents')
    cy.wait(PAUSE_MEDIUM)

    cy.log('**SCENE 3: Upload PDF Document**')
    // Click upload button
    cy.get('[data-testid="upload-button"]', { timeout: 10000 }).click()
    cy.wait(PAUSE_SHORT)

    // Select the PDF file
    cy.get('input[type="file"]').selectFile('cypress/fixtures/pdf-exemple.pdf', { force: true })
    cy.wait(PAUSE_MEDIUM)

    // File name should be visible
    cy.get('[data-testid="selected-file-name"]')
      .should('be.visible')
      .and('contain', 'pdf-exemple.pdf')
    cy.wait(PAUSE_SHORT)

    cy.log('**SCENE 4: Configure document options**')
    // Open options panel
    cy.get('[data-testid="options-toggle"]').click()
    cy.wait(PAUSE_SHORT)

    // Set document title
    cy.get('#doc-title').clear().type('Company Security Policy 2025', { delay: 40 })
    cy.wait(PAUSE_SHORT)

    // Enable "Require full read" option
    cy.contains('label', /Require full reading|Exiger la lecture/)
      .find('input[type="checkbox"]')
      .check({ force: true })
    cy.wait(PAUSE_MEDIUM)

    cy.log('**SCENE 5: Submit document**')
    // Intercept upload to get doc_id
    cy.intercept('POST', '/api/v1/documents/upload').as('uploadDoc')

    cy.get('[data-testid="submit-button"]').click()

    // Wait for upload and get docId
    cy.wait('@uploadDoc').then((interception) => {
      expect(interception.response?.statusCode).to.eq(201)
      docId = interception.response?.body?.data?.doc_id
      expect(docId).to.exist
      cy.log(`Document created with ID: ${docId}`)
    })

    cy.wait(PAUSE_LONG)
  })

  it('Part 2: Admin adds expected signers and sends reminders', () => {
    cy.log('**SCENE 6: Navigate to Admin Document Detail**')
    cy.loginAsAdmin()
    cy.visit(`/admin/docs/${docId}`)
    cy.wait(PAUSE_MEDIUM)

    // Document title should be visible
    cy.get('[data-testid="document-title-input"]')
      .should('have.value', 'Company Security Policy 2025')
    cy.wait(PAUSE_SHORT)

    cy.log('**SCENE 7: Add Expected Signers**')
    cy.get('[data-testid="open-add-signers-btn"]').click()
    cy.wait(PAUSE_SHORT)

    cy.get('[data-testid="signers-textarea"]')
      .type(`${alice}\n${bob}`, { delay: 40 })
    cy.wait(PAUSE_SHORT)

    cy.get('[data-testid="add-signers-btn"]').click()
    cy.wait(PAUSE_MEDIUM)

    // Verify signers are added
    cy.contains(alice, { timeout: 10000 }).should('be.visible')
    cy.contains(bob).should('be.visible')
    cy.wait(PAUSE_MEDIUM)

    cy.log('**SCENE 8: Document Status - 0% Complete**')
    cy.contains('Confirmed').parent().should('contain', '0')
    cy.contains('Expected').parent().should('contain', '2')
    cy.wait(PAUSE_LONG)

    cy.log('**SCENE 9: Send Email Reminders**')
    cy.clearMailbox()
    cy.contains('button', 'Send reminders').click()
    cy.wait(PAUSE_SHORT)

    cy.get('[data-testid="confirm-button"]').click()
    cy.contains(/Reminder.*sent|sent successfully/, { timeout: 15000 }).should('be.visible')
    cy.wait(PAUSE_LONG)

    cy.logout()
  })

  it('Part 3: Alice reads PDF (with 100% progress) and signs', () => {
    cy.log('**SCENE 10: Alice receives email and logs in**')
    cy.loginViaMagicLink(alice, `/?doc=${docId}`)
    cy.wait(PAUSE_MEDIUM)

    cy.log('**SCENE 11: PDF Viewer with progress tracking**')
    cy.url({ timeout: 10000 }).should('include', `/?doc=${docId}`)

    // Wait for PDF viewer to load
    cy.contains('Company Security Policy 2025', { timeout: 15000 }).should('be.visible')
    cy.wait(PAUSE_MEDIUM)

    // Show progress bar (should start at 0% or low)
    cy.contains(/\d+%/).should('be.visible')
    cy.wait(PAUSE_SHORT)

    cy.log('**SCENE 12: Scroll through PDF to read it**')
    // Scroll through the PDF content
    // The PDF viewer container - we need to scroll inside it
    cy.get('.pdf-container, .document-content, [class*="viewer"]')
      .first()
      .then($el => {
        if ($el.length) {
          // Scroll in increments to show progress
          const scrollSteps = 5
          for (let i = 1; i <= scrollSteps; i++) {
            cy.wrap($el).scrollTo('bottom', { duration: 800, ensureScrollable: false })
            cy.wait(400)
          }
        }
      })

    cy.wait(PAUSE_MEDIUM)

    // Progress should now be 100%
    cy.contains('100%', { timeout: 10000 }).should('be.visible')
    cy.wait(PAUSE_LONG)

    cy.log('**SCENE 13: Alice confirms reading**')
    // Check certify checkbox
    cy.contains('label', /I certify|Je certifie/, { timeout: 10000 })
      .find('input[type="checkbox"]')
      .check({ force: true })
    cy.wait(PAUSE_SHORT)

    // Confirm button should now be enabled
    cy.contains('button', /Confirm reading|Confirmer la lecture/, { timeout: 10000 })
      .should('be.visible')
      .and('not.be.disabled')
      .click()

    cy.log('**SCENE 14: Signature confirmed!**')
    cy.contains('Reading confirmed', { timeout: 15000 }).should('be.visible')
    cy.wait(PAUSE_XLONG)

    cy.logout()
  })

  it('Part 4: Admin checks progress - 50% complete', () => {
    cy.log('**SCENE 15: Admin reviews progress**')
    cy.loginAsAdmin()
    cy.visit(`/admin/docs/${docId}`)
    cy.wait(PAUSE_MEDIUM)

    cy.log('**SCENE 16: Document Status - 50% Complete**')
    cy.contains('Confirmed', { timeout: 10000 }).parent().should('contain', '1')
    cy.contains('Pending').parent().should('contain', '1')
    cy.wait(PAUSE_MEDIUM)

    // Alice shows as confirmed
    cy.contains('tr', alice).should('contain', 'Confirmed')
    cy.wait(PAUSE_SHORT)

    // Bob still pending
    cy.contains('tr', bob).should('contain', 'Pending')
    cy.wait(PAUSE_LONG)

    cy.logout()
  })

  it('Part 5: Bob signs the document', () => {
    cy.log('**SCENE 17: Bob logs in and views document**')
    cy.loginViaMagicLink(bob, `/?doc=${docId}`)
    cy.wait(PAUSE_MEDIUM)

    cy.url({ timeout: 10000 }).should('include', `/?doc=${docId}`)

    // Wait for PDF viewer
    cy.contains('Company Security Policy 2025', { timeout: 15000 }).should('be.visible')
    cy.wait(PAUSE_SHORT)

    cy.log('**SCENE 18: Bob scrolls to read PDF**')
    // Scroll through PDF
    cy.get('.pdf-container, .document-content, [class*="viewer"]')
      .first()
      .then($el => {
        if ($el.length) {
          cy.wrap($el).scrollTo('bottom', { duration: 1500, ensureScrollable: false })
        }
      })

    cy.wait(PAUSE_MEDIUM)

    cy.log('**SCENE 19: Bob confirms reading**')
    cy.contains('label', /I certify|Je certifie/, { timeout: 10000 })
      .find('input[type="checkbox"]')
      .check({ force: true })
    cy.wait(PAUSE_SHORT)

    cy.contains('button', /Confirm reading|Confirmer la lecture/, { timeout: 10000 })
      .should('be.visible')
      .and('not.be.disabled')
      .click()

    cy.contains('Reading confirmed', { timeout: 15000 }).should('be.visible')
    cy.wait(PAUSE_LONG)

    cy.logout()
  })

  it('Part 6: Admin sees 100% completion', () => {
    cy.log('**SCENE 20: Final Status - 100% Complete**')
    cy.loginAsAdmin()
    cy.visit(`/admin/docs/${docId}`)
    cy.wait(PAUSE_MEDIUM)

    // All signed!
    cy.contains('Confirmed', { timeout: 10000 }).parent().should('contain', '2')
    cy.contains('Expected').parent().should('contain', '2')
    cy.contains('Pending').parent().should('contain', '0')
    cy.wait(PAUSE_MEDIUM)

    // Both signers confirmed
    cy.contains('tr', alice).should('contain', 'Confirmed')
    cy.contains('tr', bob).should('contain', 'Confirmed')
    cy.wait(PAUSE_LONG)

    cy.log('**SCENE 21: Admin Documents Overview**')
    cy.visit('/admin')
    cy.wait(PAUSE_MEDIUM)

    // Show our document in the list
    cy.contains('Company Security Policy 2025', { timeout: 10000 }).should('be.visible')
    cy.wait(PAUSE_LONG)
  })

  it('Part 7: User views their signature history', () => {
    cy.log('**SCENE 22: My Signatures Page**')
    cy.loginViaMagicLink(alice)
    cy.wait(PAUSE_MEDIUM)

    cy.visit('/signatures')
    cy.wait(PAUSE_MEDIUM)

    cy.log('**SCENE 23: Signature history with document**')
    // Show alice's signatures - should see our document
    cy.contains('Company Security Policy 2025', { timeout: 10000 }).should('be.visible')
    cy.wait(PAUSE_SHORT)

    // Show stats
    cy.contains('Total confirmations', { timeout: 5000 }).should('be.visible')
    cy.wait(PAUSE_XLONG)
  })
})
