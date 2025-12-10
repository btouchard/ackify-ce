// SPDX-License-Identifier: AGPL-3.0-or-later
/// <reference types="cypress" />

describe('Test 11: Webhooks CRUD Operations', () => {
  const webhookTitle = 'Test Webhook ' + Date.now()
  const webhookUrl = 'https://webhook.site/' + Date.now()

  beforeEach(() => {
    cy.clearCookies()
  })

  it('should allow admin to access webhooks page', () => {
    // Step 1: Login as admin
    cy.loginAsAdmin()

    // Step 2: Navigate to webhooks page
    cy.visit('/admin/webhooks')
    cy.url({ timeout: 10000 }).should('include', '/admin/webhooks')

    // Step 3: Verify page elements
    cy.contains('Webhooks', { timeout: 10000 }).should('be.visible')
    cy.contains('button', 'New webhook').should('be.visible')
  })

  it('should create a new webhook', () => {
    // Step 1: Login and navigate to webhooks
    cy.loginAsAdmin()
    cy.visit('/admin/webhooks')

    // Step 2: Click create webhook button
    cy.contains('button', 'New webhook').click()

    // Step 3: Should navigate to new webhook form
    cy.url({ timeout: 10000 }).should('include', '/admin/webhooks/new')
    cy.contains('New webhook', { timeout: 10000 }).should('be.visible')

    // Step 4: Fill in webhook form (using Input components with v-model)
    cy.get('input[type="text"]').first().type(webhookTitle)
    cy.get('input[type="url"]').type(webhookUrl)
    cy.get('input[type="text"]').eq(1).type('test_webhook_secret_123')

    // Step 5: Select event types (checkboxes)
    cy.get('input[type="checkbox"]').then(($checkboxes) => {
      // Select first 3 event types
      cy.wrap($checkboxes[0]).check()
      cy.wrap($checkboxes[1]).check()
      cy.wrap($checkboxes[2]).check()
    })

    // Step 6: Add description (optional)
    cy.get('textarea').type('E2E test webhook for document and signature events')

    // Step 7: Submit form
    cy.get('button[type="submit"]').click()

    // Step 8: Should redirect back to webhooks list
    cy.url({ timeout: 10000 }).should('eq', Cypress.config('baseUrl') + '/admin/webhooks')

    // Step 9: Verify webhook appears in list
    cy.contains(webhookTitle, { timeout: 10000 }).should('be.visible')
    cy.contains(webhookUrl).should('be.visible')

    // Step 10: Verify webhook is active (check for "Active" status)
    cy.contains('table tbody tr', webhookTitle).within(() => {
      cy.contains('Active').should('be.visible')
    })
  })

  it('should edit webhook', () => {
    // Step 1: Login and navigate to webhooks
    cy.loginAsAdmin()
    cy.visit('/admin/webhooks')

    // Step 2: Click Edit button on webhook row
    cy.contains('table tbody tr', webhookTitle, { timeout: 10000 }).within(() => {
      cy.contains('button', 'Edit').click()
    })

    // Step 3: Should navigate to webhook edit page
    cy.url({ timeout: 10000 }).should('include', '/admin/webhooks/')
    cy.contains('Edit webhook', { timeout: 10000 }).should('be.visible')

    // Step 4: Verify webhook data is loaded
    cy.get('input[type="text"]').first().should('have.value', webhookTitle)
    cy.get('input[type="url"]').should('have.value', webhookUrl)

    // Step 5: Update webhook title
    const updatedTitle = webhookTitle + ' (Updated)'
    cy.get('input[type="text"]').first().clear().type(updatedTitle)

    // Step 6: Update description
    cy.get('textarea').clear().type('Updated webhook description')

    // Step 7: Submit form
    cy.get('button[type="submit"]').click()

    // Step 8: Should redirect back to webhooks list
    cy.url({ timeout: 10000 }).should('eq', Cypress.config('baseUrl') + '/admin/webhooks')

    // Step 9: Verify updated title appears in list
    cy.contains(updatedTitle, { timeout: 10000 }).should('be.visible')
  })

  it('should disable webhook', () => {
    // Step 1: Login and navigate to webhooks
    cy.loginAsAdmin()
    cy.visit('/admin/webhooks')

    // Step 2: Find webhook row and disable it
    const updatedTitle = webhookTitle + ' (Updated)'
    cy.contains('table tbody tr', updatedTitle, { timeout: 10000 }).within(() => {
      cy.contains('button', 'Disable').click()
    })

    // Step 3: Verify webhook is now inactive
    cy.contains('table tbody tr', updatedTitle, { timeout: 10000 }).within(() => {
      cy.contains('Inactive').should('be.visible')
    })
  })

  it('should enable webhook', () => {
    // Step 1: Login and navigate to webhooks
    cy.loginAsAdmin()
    cy.visit('/admin/webhooks')

    // Step 2: Find webhook row and enable it
    const updatedTitle = webhookTitle + ' (Updated)'
    cy.contains('table tbody tr', updatedTitle, { timeout: 10000 }).within(() => {
      cy.contains('button', 'Enable').click()
    })

    // Step 3: Verify webhook is now active
    cy.contains('table tbody tr', updatedTitle, { timeout: 10000 }).within(() => {
      cy.contains('Active').should('be.visible')
    })
  })

  it('should delete webhook from list', () => {
    // Step 1: Login and navigate to webhooks
    cy.loginAsAdmin()
    cy.visit('/admin/webhooks')

    // Step 2: Find webhook row and click delete button
    const updatedTitle = webhookTitle + ' (Updated)'
    cy.contains('table tbody tr', updatedTitle, { timeout: 10000 }).within(() => {
      cy.contains('button', 'Delete').click()
    })

    // Step 3: Confirm deletion in browser confirm dialog (handled by Cypress automatically)
    // The webhook list page uses native confirm() dialog

    // Step 4: Verify webhook is deleted from list
    cy.contains(updatedTitle, { timeout: 10000 }).should('not.exist')
  })

  it('should prevent non-admin from accessing webhooks', () => {
    // Step 1: Login as regular user (not admin)
    cy.loginViaMagicLink('user@test.com')

    // Step 2: Try to navigate to webhooks page
    cy.visit('/admin/webhooks', { failOnStatusCode: false })

    // Step 3: Should redirect to home
    cy.url({ timeout: 10000 }).should('not.include', '/admin/webhooks')
    cy.url().should('match', /\/$/)
  })
})
