// SPDX-License-Identifier: AGPL-3.0-or-later
/// <reference types="cypress" />

describe('Test 16: Admin Settings Configuration', () => {
  beforeEach(() => {
    cy.clearCookies()
  })

  // Helper to save settings and wait for success message
  const saveAndWaitForSuccess = () => {
    cy.get('button').contains('Save').click()
    cy.contains('saved successfully', { timeout: 10000 }).should('be.visible')
  }

  // Helper to navigate to a settings section
  const navigateToSection = (sectionName: string) => {
    cy.contains('button', sectionName).click()
  }

  describe('Basic Navigation and Access', () => {
    it('should allow admin to access settings page', () => {
      cy.loginAsAdmin()
      cy.visit('/admin/settings')
      cy.url({ timeout: 10000 }).should('include', '/admin/settings')

      // Verify page elements
      cy.contains('Settings', { timeout: 10000 }).should('be.visible')

      // Verify sidebar navigation sections are present
      cy.contains('button', 'General').should('be.visible')
      cy.contains('button', 'OAuth / OIDC').should('be.visible')
      cy.contains('button', 'Magic Link').should('be.visible')
      cy.contains('button', 'Email (SMTP)').should('be.visible')
      cy.contains('button', 'Storage').should('be.visible')
    })

    it('should prevent non-admin from accessing settings', () => {
      cy.loginViaMagicLink('user@test.com')
      cy.visit('/admin/settings', { failOnStatusCode: false })

      // Should redirect away from admin settings
      cy.url({ timeout: 10000 }).should('not.include', '/admin/settings')
    })

    it('should navigate to settings from admin dashboard', () => {
      cy.loginAsAdmin()
      cy.visit('/admin')

      // Click on Settings link/button in the dashboard (router-link wrapped button)
      cy.get('a[href="/admin/settings"]').click()

      // Should navigate to settings page
      cy.url({ timeout: 10000 }).should('include', '/admin/settings')
      cy.contains('Settings', { timeout: 10000 }).should('be.visible')
    })
  })

  describe('General Settings Section', () => {
    beforeEach(() => {
      cy.loginAsAdmin()
      cy.visit('/admin/settings')
      cy.contains('General', { timeout: 10000 }).should('be.visible')
    })

    it('should display general settings form', () => {
      // General section should be visible by default
      cy.contains('h2', 'General', { timeout: 10000 }).should('be.visible')

      // Verify form fields are present
      cy.get('[data-testid="organisation"]').should('exist')
      cy.get('[data-testid="only_admin_can_create"]').should('exist')
    })

    it('should update organisation name and persist after reload', () => {
      const newOrg = 'E2E Test Org ' + Date.now()

      // Update organisation name
      cy.get('[data-testid="organisation"]').clear().type(newOrg)

      // Save settings
      saveAndWaitForSuccess()

      // Reload and verify persistence
      cy.reload()
      cy.get('[data-testid="organisation"]', { timeout: 10000 }).should('have.value', newOrg)

      // Restore original value
      cy.get('[data-testid="organisation"]').clear().type('Ackify Test')
      saveAndWaitForSuccess()
    })

    it('should toggle only_admin_can_create setting', () => {
      // Get initial state
      cy.get('[data-testid="only_admin_can_create"]').then(($checkbox) => {
        const wasChecked = $checkbox.is(':checked')

        // Toggle the checkbox
        if (wasChecked) {
          cy.get('[data-testid="only_admin_can_create"]').uncheck()
        } else {
          cy.get('[data-testid="only_admin_can_create"]').check()
        }

        saveAndWaitForSuccess()

        // Reload and verify the change persisted
        cy.reload()
        cy.get('[data-testid="only_admin_can_create"]', { timeout: 10000 }).should(
          wasChecked ? 'not.be.checked' : 'be.checked'
        )

        // Restore original state
        if (wasChecked) {
          cy.get('[data-testid="only_admin_can_create"]').check()
        } else {
          cy.get('[data-testid="only_admin_can_create"]').uncheck()
        }
        saveAndWaitForSuccess()
      })
    })
  })

  describe('OAuth/OIDC Settings Section', () => {
    beforeEach(() => {
      cy.loginAsAdmin()
      cy.visit('/admin/settings')
      navigateToSection('OAuth / OIDC')
      cy.contains('h2', 'OAuth / OIDC', { timeout: 10000 }).should('be.visible')
    })

    it('should display OIDC form fields', () => {
      cy.get('[data-testid="oidc_enabled"]').should('exist')

      // Enable OIDC to see more fields
      cy.get('[data-testid="oidc_enabled"]').check()
      cy.get('[data-testid="oidc_provider"]').should('be.visible')
      cy.get('[data-testid="oidc_client_id"]').should('be.visible')
      cy.get('[data-testid="oidc_client_secret"]').should('be.visible')
    })

    it('should show custom provider URLs when custom is selected', () => {
      cy.get('[data-testid="oidc_enabled"]').check()
      cy.get('[data-testid="oidc_provider"]').select('custom')

      // Custom URL fields should appear
      cy.get('[data-testid="oidc_auth_url"]').should('be.visible')
      cy.get('[data-testid="oidc_token_url"]').should('be.visible')
      cy.get('[data-testid="oidc_userinfo_url"]').should('be.visible')
    })

    it('should mask client secret', () => {
      cy.get('[data-testid="oidc_enabled"]').check()

      // Client secret input should be password type (masked)
      cy.get('[data-testid="oidc_client_secret"]').should('have.attr', 'type', 'password')
    })
  })

  describe('Magic Link Settings Section', () => {
    beforeEach(() => {
      cy.loginAsAdmin()
      cy.visit('/admin/settings')
      navigateToSection('Magic Link')
      cy.contains('h2', 'Magic Link', { timeout: 10000 }).should('be.visible')
    })

    it('should display Magic Link form fields', () => {
      cy.get('[data-testid="magiclink_enabled"]').should('exist')
    })

    it('should toggle Magic Link enabled setting', () => {
      cy.get('[data-testid="magiclink_enabled"]').then(($checkbox) => {
        const wasEnabled = $checkbox.is(':checked')

        // Toggle
        if (wasEnabled) {
          cy.get('[data-testid="magiclink_enabled"]').uncheck()
        } else {
          cy.get('[data-testid="magiclink_enabled"]').check()
        }

        saveAndWaitForSuccess()

        // Reload and verify
        cy.reload()
        navigateToSection('Magic Link')
        cy.get('[data-testid="magiclink_enabled"]', { timeout: 10000 }).should(
          wasEnabled ? 'not.be.checked' : 'be.checked'
        )

        // Restore
        if (wasEnabled) {
          cy.get('[data-testid="magiclink_enabled"]').check()
        } else {
          cy.get('[data-testid="magiclink_enabled"]').uncheck()
        }
        saveAndWaitForSuccess()
      })
    })
  })

  describe('SMTP Settings Section', () => {
    beforeEach(() => {
      cy.loginAsAdmin()
      cy.visit('/admin/settings')
      navigateToSection('Email (SMTP)')
      cy.contains('h2', 'Email (SMTP)', { timeout: 10000 }).should('be.visible')
    })

    it('should display SMTP form fields', () => {
      cy.get('[data-testid="smtp_host"]').should('exist')
      cy.get('[data-testid="smtp_port"]').should('exist')
      cy.get('[data-testid="smtp_from"]').should('exist')
      cy.get('[data-testid="smtp_from_name"]').should('exist')
      cy.get('[data-testid="smtp_tls"]').should('exist')
      cy.get('[data-testid="smtp_starttls"]').should('exist')
    })

    it('should update SMTP port and persist', () => {
      // Get current port value
      cy.get('[data-testid="smtp_port"]').invoke('val').then((originalPort) => {
        const newPort = '2525'

        cy.get('[data-testid="smtp_port"]').clear().type(newPort)
        saveAndWaitForSuccess()

        // Reload and verify
        cy.reload()
        navigateToSection('Email (SMTP)')
        cy.get('[data-testid="smtp_port"]', { timeout: 10000 }).should('have.value', newPort)

        // Restore original
        cy.get('[data-testid="smtp_port"]').clear().type(String(originalPort))
        saveAndWaitForSuccess()
      })
    })
  })

  describe('Storage Settings Section', () => {
    beforeEach(() => {
      cy.loginAsAdmin()
      cy.visit('/admin/settings')
      navigateToSection('Storage')
      cy.contains('h2', 'Storage', { timeout: 10000 }).should('be.visible')
    })

    it('should display storage form fields', () => {
      cy.get('[data-testid="storage_type"]').should('exist')
      cy.get('[data-testid="storage_max_size_mb"]').should('exist')
    })

    it('should show local path when storage type is local', () => {
      cy.get('[data-testid="storage_type"]').select('local')
      cy.get('[data-testid="storage_local_path"]').should('be.visible')
    })

    it('should show S3 fields when storage type is s3', () => {
      cy.get('[data-testid="storage_type"]').select('s3')

      // S3 fields should appear
      cy.get('[data-testid="storage_s3_endpoint"]').should('be.visible')
      cy.get('[data-testid="storage_s3_bucket"]').should('be.visible')
      cy.get('[data-testid="storage_s3_access_key"]').should('be.visible')
      cy.get('[data-testid="storage_s3_secret_key"]').should('be.visible')
      cy.get('[data-testid="storage_s3_region"]').should('be.visible')
      cy.get('[data-testid="s3_use_ssl"]').should('be.visible')
    })

    it('should update max size and persist', () => {
      cy.get('[data-testid="storage_max_size_mb"]').invoke('val').then((originalSize) => {
        const newSize = '100'

        cy.get('[data-testid="storage_max_size_mb"]').clear().type(newSize)
        saveAndWaitForSuccess()

        // Reload and verify
        cy.reload()
        navigateToSection('Storage')
        cy.get('[data-testid="storage_max_size_mb"]', { timeout: 10000 }).should('have.value', newSize)

        // Restore original
        cy.get('[data-testid="storage_max_size_mb"]').clear().type(String(originalSize))
        saveAndWaitForSuccess()
      })
    })

    it('should maintain selected section after save', () => {
      // We're already on Storage section
      cy.get('[data-testid="storage_max_size_mb"]').clear().type('75')
      saveAndWaitForSuccess()

      // Should still be on Storage section
      cy.contains('h2', 'Storage').should('be.visible')
    })
  })

  describe('Reset from ENV', () => {
    beforeEach(() => {
      cy.loginAsAdmin()
      cy.visit('/admin/settings')
    })

    it('should show reset confirmation dialog', () => {
      cy.contains('button', 'Reset from ENV').click()

      // Confirmation modal should appear
      cy.contains('Reset Settings?', { timeout: 10000 }).should('be.visible')
      cy.contains('reset all settings').should('be.visible')

      // Cancel the reset
      cy.contains('button', 'Cancel').click()

      // Modal should be closed
      cy.contains('Reset Settings?').should('not.exist')
    })

    it('should reset settings to ENV values when confirmed', () => {
      // First, modify a setting
      cy.get('[data-testid="organisation"]').clear()
      cy.get('[data-testid="organisation"]').type('Modified Org Name')
      saveAndWaitForSuccess()

      // Now reset from ENV
      cy.contains('button', 'Reset from ENV').click()
      cy.contains('Reset Settings?', { timeout: 10000 }).should('be.visible')

      // Confirm reset - click the amber button inside the modal (not the Reset from ENV button)
      cy.get('.bg-amber-600').click()

      // Wait for reset success
      cy.contains('reset', { matchCase: false, timeout: 10000 }).should('be.visible')

      // Organisation should be back to ENV value
      cy.get('[data-testid="organisation"]', { timeout: 10000 }).should('have.value', 'Ackify Test')
    })
  })

  describe('Full Flow: Auth Settings affect Login Page', () => {
    it('should hide MagicLink option on login page when disabled', () => {
      // Login as admin
      cy.loginAsAdmin()

      // Go to settings and disable MagicLink
      cy.visit('/admin/settings')
      navigateToSection('Magic Link')
      cy.contains('h2', 'Magic Link', { timeout: 10000 }).should('be.visible')

      // Remember current state and disable
      cy.get('[data-testid="magiclink_enabled"]').then(($checkbox) => {
        const wasEnabled = $checkbox.is(':checked')

        if (wasEnabled) {
          cy.get('[data-testid="magiclink_enabled"]').uncheck()
          saveAndWaitForSuccess()

          // Logout
          cy.logout()

          // Visit auth page (fresh load to get new window variables)
          cy.visit('/auth')

          // MagicLink card should NOT be visible
          cy.contains('Send Magic Link').should('not.exist')

          // OAuth should still be visible (if enabled)
          // Note: OAuth button might auto-redirect if only method available

          // Re-enable MagicLink via API for other tests
          cy.loginAsAdmin()
          cy.visit('/admin/settings')
          navigateToSection('Magic Link')
          cy.get('[data-testid="magiclink_enabled"]').check()
          saveAndWaitForSuccess()
        } else {
          // MagicLink was already disabled, enable it first then run test
          cy.get('[data-testid="magiclink_enabled"]').check()
          saveAndWaitForSuccess()

          cy.logout()
          cy.visit('/auth')

          // MagicLink should be visible now
          cy.contains('Send Magic Link', { timeout: 10000 }).should('be.visible')

          // Disable it
          cy.loginAsAdmin()
          cy.visit('/admin/settings')
          navigateToSection('Magic Link')
          cy.get('[data-testid="magiclink_enabled"]').uncheck()
          saveAndWaitForSuccess()

          cy.logout()
          cy.visit('/auth')

          // MagicLink should NOT be visible
          cy.contains('Send Magic Link').should('not.exist')

          // Re-enable for other tests
          cy.loginAsAdmin()
          cy.visit('/admin/settings')
          navigateToSection('Magic Link')
          cy.get('[data-testid="magiclink_enabled"]').check()
          saveAndWaitForSuccess()
        }
      })
    })

    it('should hide OAuth option on login page when disabled', () => {
      cy.loginAsAdmin()

      // Go to settings and check OIDC status
      cy.visit('/admin/settings')
      navigateToSection('OAuth / OIDC')
      cy.contains('h2', 'OAuth / OIDC', { timeout: 10000 }).should('be.visible')

      cy.get('[data-testid="oidc_enabled"]').then(($checkbox) => {
        const wasEnabled = $checkbox.is(':checked')

        if (wasEnabled) {
          // Disable OAuth
          cy.get('[data-testid="oidc_enabled"]').uncheck()
          saveAndWaitForSuccess()

          cy.logout()
          cy.visit('/auth')

          // OAuth login button should NOT be visible
          cy.contains('Continue with OAuth').should('not.exist')

          // MagicLink should still work
          cy.contains('Send Magic Link', { timeout: 10000 }).should('be.visible')

          // Re-enable OAuth
          cy.loginAsAdmin()
          cy.visit('/admin/settings')
          navigateToSection('OAuth / OIDC')
          cy.get('[data-testid="oidc_enabled"]').check()
          saveAndWaitForSuccess()
        } else {
          // OAuth was disabled, enable it first
          cy.get('[data-testid="oidc_enabled"]').check()
          // Fill in required fields for custom provider
          cy.get('[data-testid="oidc_provider"]').select('custom')
          cy.get('[data-testid="oidc_client_id"]').clear().type('test_client_id')
          cy.get('[data-testid="oidc_client_secret"]').clear().type('test_client_secret')
          cy.get('[data-testid="oidc_auth_url"]').clear().type('https://auth.url.com/auth')
          cy.get('[data-testid="oidc_token_url"]').clear().type('https://auth.url.com/token')
          cy.get('[data-testid="oidc_userinfo_url"]').clear().type('https://auth.url.com/userinfo')
          saveAndWaitForSuccess()

          cy.logout()
          cy.visit('/auth')

          // OAuth should be visible
          cy.contains('Continue with OAuth', { timeout: 10000 }).should('be.visible')

          // Disable it
          cy.loginAsAdmin()
          cy.visit('/admin/settings')
          navigateToSection('OAuth / OIDC')
          cy.get('[data-testid="oidc_enabled"]').uncheck()
          saveAndWaitForSuccess()

          cy.logout()
          cy.visit('/auth')

          // OAuth should NOT be visible
          cy.contains('Continue with OAuth').should('not.exist')

          // Don't re-enable as it was originally disabled
        }
      })
    })

    it('should show both auth methods when both are enabled', () => {
      cy.loginAsAdmin()
      cy.visit('/admin/settings')

      // Ensure MagicLink is enabled
      navigateToSection('Magic Link')
      cy.get('[data-testid="magiclink_enabled"]').check()
      saveAndWaitForSuccess()

      // Ensure OAuth is enabled
      navigateToSection('OAuth / OIDC')
      cy.get('[data-testid="oidc_enabled"]').check()
      // If provider not set, select custom and fill required fields
      cy.get('[data-testid="oidc_provider"]').then(($select) => {
        if (!$select.val()) {
          cy.get('[data-testid="oidc_provider"]').select('custom')
          cy.get('[data-testid="oidc_client_id"]').clear().type('test_client_id')
          cy.get('[data-testid="oidc_client_secret"]').clear().type('test_client_secret')
          cy.get('[data-testid="oidc_auth_url"]').clear().type('https://auth.url.com/auth')
          cy.get('[data-testid="oidc_token_url"]').clear().type('https://auth.url.com/token')
          cy.get('[data-testid="oidc_userinfo_url"]').clear().type('https://auth.url.com/userinfo')
        }
      })
      saveAndWaitForSuccess()

      cy.logout()
      cy.visit('/auth')

      // Both auth methods should be visible
      cy.contains('Continue with OAuth', { timeout: 10000 }).should('be.visible')
      cy.contains('Send Magic Link').should('be.visible')
    })

    // Note: Test for "no auth method available" is not feasible in e2e
    // because disabling both auth methods would prevent re-login to restore state.
    // This scenario should be tested at the unit/integration level.
  })

  describe('Validation Errors', () => {
    beforeEach(() => {
      cy.loginAsAdmin()
      cy.visit('/admin/settings')
    })

    it('should show validation error for OIDC without required fields', () => {
      navigateToSection('OAuth / OIDC')
      cy.contains('h2', 'OAuth / OIDC', { timeout: 10000 }).should('be.visible')

      // Enable OIDC
      cy.get('[data-testid="oidc_enabled"]').check()

      // Select custom provider
      cy.get('[data-testid="oidc_provider"]').select('custom')

      // Fill only client_id, leave URLs empty
      cy.get('[data-testid="oidc_client_id"]').clear()
      cy.get('[data-testid="oidc_client_id"]').type('test-id')
      cy.get('[data-testid="oidc_auth_url"]').clear()
      cy.get('[data-testid="oidc_token_url"]').clear()
      cy.get('[data-testid="oidc_userinfo_url"]').clear()

      // Try to save
      cy.get('button').contains('Save').click()

      // Should show validation error (red alert box with error icon)
      cy.get('.bg-red-50, .bg-red-900\\/20', { timeout: 10000 }).should('be.visible')
    })
  })
})
