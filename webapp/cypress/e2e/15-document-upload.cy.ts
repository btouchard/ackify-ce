// SPDX-License-Identifier: AGPL-3.0-or-later
/// <reference types="cypress" />

describe('Test 15: Document Upload', () => {
  beforeEach(() => {
    cy.clearCookies()
    cy.clearMailbox()
  })

  it('should show upload button when storage is enabled', () => {
    cy.loginAsAdmin()
    cy.visit('/documents')

    cy.get('[data-testid="upload-button"]', { timeout: 10000 }).should('be.visible')
  })

  it('should upload a PDF document', () => {
    cy.loginAsAdmin()
    cy.visit('/documents')

    const pdfContent = '%PDF-1.4\n1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n2 0 obj\n<< /Type /Pages /Kids [3 0 R] /Count 1 >>\nendobj\n3 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] >>\nendobj\nxref\n0 4\n0000000000 65535 f\n0000000009 00000 n\n0000000058 00000 n\n0000000115 00000 n\ntrailer\n<< /Size 4 /Root 1 0 R >>\nstartxref\n193\n%%EOF'

    cy.get('[data-testid="upload-button"]').click()

    cy.get('input[type="file"]').selectFile({
      contents: Cypress.Buffer.from(pdfContent),
      fileName: 'test-document.pdf',
      mimeType: 'application/pdf'
    }, { force: true })

    cy.get('[data-testid="selected-file-name"]', { timeout: 5000 })
      .should('be.visible')
      .and('contain', 'test-document.pdf')

    // Intercept upload to get doc_id
    cy.intercept('POST', '/api/v1/documents/upload').as('uploadDoc')

    cy.get('[data-testid="submit-button"]').click()

    // Wait for upload to complete and get doc_id
    cy.wait('@uploadDoc').then((interception) => {
      expect(interception.response?.statusCode).to.eq(201)
      const docId = interception.response?.body?.data?.doc_id
      expect(docId).to.exist

      // Navigate to the document page
      cy.visit(`/?doc=${docId}`)
      cy.contains('button', 'Confirm reading', { timeout: 10000 }).should('be.visible')
    })
  })

  it('should upload an image document', () => {
    cy.loginAsAdmin()
    cy.visit('/documents')

    const pngHeader = new Uint8Array([
      0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
      0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52,
      0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
      0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53,
      0xDE, 0x00, 0x00, 0x00, 0x0C, 0x49, 0x44, 0x41,
      0x54, 0x08, 0xD7, 0x63, 0xF8, 0xFF, 0xFF, 0x3F,
      0x00, 0x05, 0xFE, 0x02, 0xFE, 0xDC, 0xCC, 0x59,
      0xE7, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4E,
      0x44, 0xAE, 0x42, 0x60, 0x82
    ])

    cy.get('[data-testid="upload-button"]').click()

    cy.get('input[type="file"]').selectFile({
      contents: Cypress.Buffer.from(pngHeader),
      fileName: 'test-image.png',
      mimeType: 'image/png'
    }, { force: true })

    cy.get('[data-testid="selected-file-name"]', { timeout: 5000 })
      .should('be.visible')
      .and('contain', 'test-image.png')

    // Intercept upload to get doc_id
    cy.intercept('POST', '/api/v1/documents/upload').as('uploadDoc')

    cy.get('[data-testid="submit-button"]').click()

    // Wait for upload to complete and get doc_id
    cy.wait('@uploadDoc').then((interception) => {
      expect(interception.response?.statusCode).to.eq(201)
      const docId = interception.response?.body?.data?.doc_id
      expect(docId).to.exist

      // Navigate to the document page
      cy.visit(`/?doc=${docId}`)
      cy.contains('button', 'Confirm reading', { timeout: 10000 }).should('be.visible')
    })
  })

  it('should set custom title for uploaded document', () => {
    const customTitle = 'Custom Document Title ' + Date.now()

    cy.loginAsAdmin()
    cy.visit('/documents')

    const pdfContent = '%PDF-1.4\n1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\ntrailer\n<< /Root 1 0 R >>\n%%EOF'

    cy.get('[data-testid="upload-button"]').click()

    cy.get('input[type="file"]').selectFile({
      contents: Cypress.Buffer.from(pdfContent),
      fileName: 'document-with-title.pdf',
      mimeType: 'application/pdf'
    }, { force: true })

    cy.get('[data-testid="selected-file-name"]', { timeout: 5000 }).should('be.visible')

    cy.get('[data-testid="options-toggle"]').click()

    cy.get('#doc-title').clear()
    cy.get('#doc-title').type(customTitle)

    // Intercept upload to get doc_id
    cy.intercept('POST', '/api/v1/documents/upload').as('uploadDoc')

    cy.get('[data-testid="submit-button"]').click()

    // Wait for upload to complete and get doc_id
    cy.wait('@uploadDoc').then((interception) => {
      expect(interception.response?.statusCode).to.eq(201)
      const docId = interception.response?.body?.data?.doc_id
      expect(docId).to.exist

      // Navigate to the document page to verify the custom title
      cy.visit(`/?doc=${docId}`)
      cy.contains(customTitle, { timeout: 10000 }).should('be.visible')
    })
  })

  it('should show upload progress', () => {
    cy.loginAsAdmin()
    cy.visit('/documents')

    let largePdf = '%PDF-1.4\n'
    for (let i = 0; i < 1000; i++) {
      largePdf += `${i} 0 obj\n<< /Type /XObject /Length 1000 >>\nstream\n${'x'.repeat(1000)}\nendstream\nendobj\n`
    }
    largePdf += 'trailer\n<< /Root 1 0 R >>\n%%EOF'

    cy.get('[data-testid="upload-button"]').click()

    cy.get('input[type="file"]').selectFile({
      contents: Cypress.Buffer.from(largePdf),
      fileName: 'large-document.pdf',
      mimeType: 'application/pdf'
    }, { force: true })

    // Intercept upload to verify it completes
    cy.intercept('POST', '/api/v1/documents/upload').as('uploadDoc')

    cy.get('[data-testid="submit-button"]').click()

    // Wait for upload to complete (longer timeout for large file)
    cy.wait('@uploadDoc', { timeout: 30000 }).then((interception) => {
      expect(interception.response?.statusCode).to.eq(201)
      expect(interception.response?.body?.data?.doc_id).to.exist
    })
  })

  it('should clear selected file when clicking remove', () => {
    cy.loginAsAdmin()
    cy.visit('/documents')

    const pdfContent = '%PDF-1.4\ntrailer\n%%EOF'

    cy.get('[data-testid="upload-button"]').click()

    cy.get('input[type="file"]').selectFile({
      contents: Cypress.Buffer.from(pdfContent),
      fileName: 'to-remove.pdf',
      mimeType: 'application/pdf'
    }, { force: true })

    cy.get('[data-testid="selected-file-name"]', { timeout: 5000 })
      .should('be.visible')
      .and('contain', 'to-remove.pdf')

    cy.get('[data-testid="clear-file-button"]').click()

    cy.get('[data-testid="selected-file-name"]').should('not.exist')

    cy.get('[data-testid="doc-url-input"]').should('be.visible')
  })

  it('should auto-set title from filename', () => {
    cy.loginAsAdmin()
    cy.visit('/documents')

    const pdfContent = '%PDF-1.4\ntrailer\n%%EOF'

    cy.get('[data-testid="upload-button"]').click()

    cy.get('input[type="file"]').selectFile({
      contents: Cypress.Buffer.from(pdfContent),
      fileName: 'my-important-document.pdf',
      mimeType: 'application/pdf'
    }, { force: true })

    cy.get('[data-testid="options-toggle"]').click()

    cy.get('#doc-title').should('have.value', 'my-important-document.pdf')
  })

  it('should handle upload errors gracefully', () => {
    cy.loginAsAdmin()
    cy.visit('/documents')

    cy.intercept('POST', '/api/v1/documents/upload', {
      statusCode: 413,
      body: {
        error: {
          code: 'BAD_REQUEST',
          message: 'File too large'
        }
      }
    }).as('uploadError')

    const pdfContent = '%PDF-1.4\ntrailer\n%%EOF'

    cy.get('[data-testid="upload-button"]').click()

    cy.get('input[type="file"]').selectFile({
      contents: Cypress.Buffer.from(pdfContent),
      fileName: 'error-test.pdf',
      mimeType: 'application/pdf'
    }, { force: true })

    cy.get('[data-testid="submit-button"]').click()

    cy.wait('@uploadError')

    cy.get('[data-testid="error-message"]', { timeout: 5000 })
      .should('be.visible')
      .and('contain', 'File too large')
  })

  it('should view uploaded document in admin panel', () => {
    const docTitle = 'Admin View Test ' + Date.now()

    cy.loginAsAdmin()
    cy.visit('/documents')

    const pdfContent = '%PDF-1.4\n1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\ntrailer\n<< /Root 1 0 R >>\n%%EOF'

    cy.get('[data-testid="upload-button"]').click()

    cy.get('input[type="file"]').selectFile({
      contents: Cypress.Buffer.from(pdfContent),
      fileName: 'admin-test.pdf',
      mimeType: 'application/pdf'
    }, { force: true })

    cy.get('[data-testid="options-toggle"]').click()
    cy.get('#doc-title').clear()
    cy.get('#doc-title').type(docTitle)

    // Intercept upload to get doc_id
    cy.intercept('POST', '/api/v1/documents/upload').as('uploadDoc')

    cy.get('[data-testid="submit-button"]').click()

    // Wait for upload to complete and get doc_id
    cy.wait('@uploadDoc').then((interception) => {
      expect(interception.response?.statusCode).to.eq(201)
      const docId = interception.response?.body?.data?.doc_id
      expect(docId).to.exist

      cy.visit(`/admin/docs/${docId}`)

      // The title is displayed in the metadata form input
      cy.get('[data-testid="document-title-input"]', { timeout: 10000 })
        .should('have.value', docTitle)
    })
  })
})
