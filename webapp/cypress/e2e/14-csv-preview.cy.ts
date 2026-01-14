// SPDX-License-Identifier: AGPL-3.0-or-later
/// <reference types="cypress" />

describe('Test 14: CSV Import Preview', () => {
  const testDocId = 'csv-preview-test-' + Date.now()

  beforeEach(() => {
    cy.clearCookies()
    cy.clearMailbox()
  })

  it('should show CSV preview before import', () => {
    // Step 1: Login as admin and create document
    cy.loginAsAdmin()
    cy.visit('/admin')
    cy.get('[data-testid="doc-url-input"]').type(testDocId)
    cy.get('[data-testid="submit-button"]').click()
    cy.url({ timeout: 10000 }).should('include', `/admin/docs/${testDocId}`)

    // Step 2: Click Import CSV button
    cy.contains('button', 'Import CSV').click()

    // Step 3: Upload CSV file with valid emails
    const validCsv = `email,name,notes
alice@test.com,Alice Smith,Marketing team
bob@test.com,Bob Johnson,Development team
charlie@test.com,Charlie Brown,Design team`

    cy.get('input[type="file"]').selectFile({
      contents: Cypress.Buffer.from(validCsv),
      fileName: 'signers.csv',
      mimeType: 'text/csv'
    }, { force: true })

    // Step 4: Click Analyze button
    cy.contains('button', 'Analyze').click()

    // Step 5: Should show preview with summary (check for Valid label and count)
    cy.contains('Valid', { timeout: 10000 }).should('be.visible')
    cy.get('.text-emerald-600').contains('3').should('be.visible')

    // Step 6: Should show preview table with emails
    cy.contains('alice@test.com').should('be.visible')
    cy.contains('Alice Smith').should('be.visible')
    cy.contains('bob@test.com').should('be.visible')
    cy.contains('charlie@test.com').should('be.visible')

    // Step 7: Confirm import (button shows count)
    cy.contains('button', 'Import 3 reader').click()

    // Step 8: Should close modal and show imported signers
    cy.get('input[type="file"]', { timeout: 5000 }).should('not.exist')
    cy.contains('alice@test.com', { timeout: 10000 }).should('be.visible')
    cy.contains('bob@test.com').should('be.visible')
    cy.contains('charlie@test.com').should('be.visible')
  })

  it('should detect existing emails in CSV preview', () => {
    // Step 1: Login as admin and navigate to test document
    cy.loginAsAdmin()
    cy.visit(`/admin/docs/${testDocId}`)
    cy.url({ timeout: 10000 }).should('include', `/admin/docs/${testDocId}`)

    // Step 2: Click Import CSV button
    cy.contains('button', 'Import CSV').click()

    // Step 3: Upload CSV with existing + new emails
    const csvWithExisting = `email,name
alice@test.com,Alice Updated
david@test.com,David New`

    cy.get('input[type="file"]').selectFile({
      contents: Cypress.Buffer.from(csvWithExisting),
      fileName: 'signers-existing.csv',
      mimeType: 'text/csv'
    }, { force: true })

    // Step 4: Click Analyze button
    cy.contains('button', 'Analyze').click()

    // Step 5: Should show preview with existing email detected
    cy.contains('Valid', { timeout: 10000 }).should('be.visible')
    cy.get('.text-emerald-600').contains('1').should('be.visible')
    cy.contains('Already exist').should('be.visible')
    cy.get('.text-amber-600').contains('1').should('be.visible')

    // Step 6: Should show existing email in preview with "Existing" badge
    cy.contains('alice@test.com').should('be.visible')
    cy.contains('Existing').should('be.visible')

    // Step 7: Should show new email
    cy.contains('david@test.com').should('be.visible')

    // Step 8: Confirm import (should only import new email)
    cy.contains('button', 'Import 1 reader').click()

    // Step 9: Verify new email was added
    cy.get('input[type="file"]', { timeout: 5000 }).should('not.exist')
    cy.contains('david@test.com', { timeout: 10000 }).should('be.visible')

    // Step 10: Verify total count (original 3 + 1 new = 4)
    cy.get('table tbody tr').should('have.length', 4)
  })

  it('should detect invalid emails in CSV preview', () => {
    const invalidCsvDocId = 'csv-invalid-test-' + Date.now()

    // Step 1: Login and create new document
    cy.loginAsAdmin()
    cy.visit('/admin')
    cy.get('[data-testid="doc-url-input"]').type(invalidCsvDocId)
    cy.get('[data-testid="submit-button"]').click()
    cy.url({ timeout: 10000 }).should('include', `/admin/docs/${invalidCsvDocId}`)

    // Step 2: Click Import CSV button
    cy.contains('button', 'Import CSV').click()

    // Step 3: Upload CSV with invalid emails
    const invalidCsv = `email,name
valid@test.com,Valid User
invalid-email,Invalid User
@missing-local.com,Missing Local
missing-domain@,Missing Domain`

    cy.get('input[type="file"]').selectFile({
      contents: Cypress.Buffer.from(invalidCsv),
      fileName: 'signers-invalid.csv',
      mimeType: 'text/csv'
    }, { force: true })

    // Step 4: Click Analyze button
    cy.contains('button', 'Analyze').click()

    // Step 5: Should show preview with invalid emails detected
    cy.contains('Valid', { timeout: 10000 }).should('be.visible')
    cy.get('.text-emerald-600').contains('1').should('be.visible')
    cy.contains('Invalid').should('be.visible')
    cy.get('.text-red-600').contains('3').should('be.visible')

    // Step 6: Invalid count is shown (invalid emails not displayed in table)

    // Step 7: Confirm import (should only import valid email)
    cy.contains('button', 'Import 1 reader').click()

    // Step 8: Verify only valid email was added
    cy.get('input[type="file"]', { timeout: 5000 }).should('not.exist')
    cy.contains('valid@test.com', { timeout: 10000 }).should('be.visible')
    cy.get('table tbody tr').should('have.length', 1)
  })

  it('should handle CSV with only headers (no data)', () => {
    const emptyDocId = 'csv-empty-test-' + Date.now()

    // Step 1: Login and create new document
    cy.loginAsAdmin()
    cy.visit('/admin')
    cy.get('[data-testid="doc-url-input"]').type(emptyDocId)
    cy.get('[data-testid="submit-button"]').click()
    cy.url({ timeout: 10000 }).should('include', `/admin/docs/${emptyDocId}`)

    // Step 2: Click Import CSV button
    cy.contains('button', 'Import CSV').click()

    // Step 3: Upload CSV with only headers
    const emptyCsv = `email,name,notes`

    cy.get('input[type="file"]').selectFile({
      contents: Cypress.Buffer.from(emptyCsv),
      fileName: 'signers-empty.csv',
      mimeType: 'text/csv'
    }, { force: true })

    // Step 4: Click Analyze button
    cy.contains('button', 'Analyze').click()

    // Step 5: Should show 0 valid entries or show disabled import button
    cy.contains('Valid', { timeout: 10000 }).should('be.visible')
    // Import button should be disabled when no valid entries
    cy.contains('button', /Import \d+ reader/).should('be.disabled')
  })

  it('should handle CSV with missing email column', () => {
    const missingColDocId = 'csv-missing-col-' + Date.now()

    // Step 1: Login and create new document
    cy.loginAsAdmin()
    cy.visit('/admin')
    cy.get('[data-testid="doc-url-input"]').type(missingColDocId)
    cy.get('[data-testid="submit-button"]').click()
    cy.url({ timeout: 10000 }).should('include', `/admin/docs/${missingColDocId}`)

    // Step 2: Click Import CSV button
    cy.contains('button', 'Import CSV').click()

    // Step 3: Upload CSV without email column
    const missingColCsv = `name,notes
Alice Smith,Marketing
Bob Johnson,Development`

    cy.get('input[type="file"]').selectFile({
      contents: Cypress.Buffer.from(missingColCsv),
      fileName: 'signers-no-email.csv',
      mimeType: 'text/csv'
    }, { force: true })

    // Step 4: Click Analyze button
    cy.contains('button', 'Analyze').click()

    // Step 5: Should show error about missing email column or all entries invalid
    // The API should return an error or show all as invalid
    cy.contains(/email|Invalid|error/i, { timeout: 10000 }).should('be.visible')
  })

  it('should handle large CSV file preview', () => {
    const largeCsvDocId = 'csv-large-test-' + Date.now()

    // Step 1: Login and create new document
    cy.loginAsAdmin()
    cy.visit('/admin')
    cy.get('[data-testid="doc-url-input"]').type(largeCsvDocId)
    cy.get('[data-testid="submit-button"]').click()
    cy.url({ timeout: 10000 }).should('include', `/admin/docs/${largeCsvDocId}`)

    // Step 2: Generate large CSV with 50 emails
    let largeCsv = 'email,name,notes\n'
    for (let i = 1; i <= 50; i++) {
      largeCsv += `user${i}@test.com,User ${i},Test user ${i}\n`
    }

    // Step 3: Click Import CSV button
    cy.contains('button', 'Import CSV').click()

    // Step 4: Upload large CSV
    cy.get('input[type="file"]').selectFile({
      contents: Cypress.Buffer.from(largeCsv),
      fileName: 'signers-large.csv',
      mimeType: 'text/csv'
    }, { force: true })

    // Step 5: Click Analyze button
    cy.contains('button', 'Analyze').click()

    // Step 6: Should show preview with all 50 valid emails
    cy.contains('Valid', { timeout: 10000 }).should('be.visible')
    cy.get('.text-emerald-600').contains('50').should('be.visible')

    // Step 7: Preview table should show some emails
    cy.contains('user1@test.com').should('be.visible')

    // Step 8: Confirm import
    cy.contains('button', 'Import 50 reader').click()

    // Step 9: Should show success (may take a moment)
    cy.get('input[type="file"]', { timeout: 10000 }).should('not.exist')

    // Step 10: Verify some imported emails appear in table
    cy.contains('user1@test.com', { timeout: 15000 }).should('be.visible')
  })

  it('should allow canceling CSV import from preview', () => {
    // Step 1: Login and navigate to test document
    cy.loginAsAdmin()
    cy.visit(`/admin/docs/${testDocId}`)
    cy.url({ timeout: 10000 }).should('include', `/admin/docs/${testDocId}`)

    // Step 2: Wait for table to load and get initial signer count
    cy.get('table tbody tr', { timeout: 10000 }).then(($rows) => {
      const initialCount = $rows.length

      // Step 3: Click Import CSV button
      cy.contains('button', 'Import CSV').click()

      // Step 4: Upload CSV
      const cancelCsv = `email,name
cancel1@test.com,Cancel User 1
cancel2@test.com,Cancel User 2`

      cy.get('input[type="file"]').selectFile({
        contents: Cypress.Buffer.from(cancelCsv),
        fileName: 'signers-cancel.csv',
        mimeType: 'text/csv'
      }, { force: true })

      // Step 5: Click Analyze button
      cy.contains('button', 'Analyze').click()

      // Step 6: Wait for preview
      cy.contains('Valid', { timeout: 10000 }).should('be.visible')
      cy.get('.text-emerald-600').contains('2').should('be.visible')

      // Step 7: Click Cancel button
      cy.contains('button', 'Cancel').click()

      // Step 8: Should close modal
      cy.get('input[type="file"]', { timeout: 5000 }).should('not.exist')

      // Step 9: Verify no emails were imported
      cy.get('table tbody tr').should('have.length', initialCount)
      cy.contains('cancel1@test.com').should('not.exist')
      cy.contains('cancel2@test.com').should('not.exist')
    })
  })

  it('should show preview with mixed valid, existing, and invalid emails', () => {
    // Step 1: Login and navigate to test document
    cy.loginAsAdmin()
    cy.visit(`/admin/docs/${testDocId}`)
    cy.url({ timeout: 10000 }).should('include', `/admin/docs/${testDocId}`)

    // Step 2: Click Import CSV button
    cy.contains('button', 'Import CSV').click()

    // Step 3: Upload CSV with mixed emails
    const mixedCsv = `email,name,notes
new-user@test.com,New User,Should be imported
alice@test.com,Alice Duplicate,Already exists
invalid-email,Invalid User,Should be rejected
another-new@test.com,Another New,Should be imported`

    cy.get('input[type="file"]').selectFile({
      contents: Cypress.Buffer.from(mixedCsv),
      fileName: 'signers-mixed.csv',
      mimeType: 'text/csv'
    }, { force: true })

    // Step 4: Click Analyze button
    cy.contains('button', 'Analyze').click()

    // Step 5: Should show preview with accurate counts
    cy.contains('Valid', { timeout: 10000 }).should('be.visible')
    cy.get('.text-emerald-600').contains('2').should('be.visible')
    cy.contains('Already exist').should('be.visible')
    cy.get('.text-amber-600').contains('1').should('be.visible')
    cy.contains('Invalid').should('be.visible')
    cy.get('.text-red-600').contains('1').should('be.visible')

    // Step 6: Verify valid and existing entries are displayed (invalid not shown in table)
    cy.contains('new-user@test.com').should('be.visible')
    cy.contains('another-new@test.com').should('be.visible')
    cy.contains('alice@test.com').should('be.visible')
    cy.contains('Existing').should('be.visible')

    // Step 7: Confirm import
    cy.contains('button', 'Import 2 reader').click()

    // Step 8: Verify only new valid emails were added
    cy.get('input[type="file"]', { timeout: 5000 }).should('not.exist')
    cy.contains('new-user@test.com', { timeout: 10000 }).should('be.visible')
    cy.contains('another-new@test.com').should('be.visible')
  })
})
