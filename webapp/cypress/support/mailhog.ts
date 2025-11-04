// SPDX-License-Identifier: AGPL-3.0-or-later
/// <reference types="cypress" />

/**
 * Mailhog API Helper
 * Documentation: https://github.com/mailhog/MailHog/blob/master/docs/APIv2.md
 */

interface MailhogMessage {
  ID: string
  From: {
    Mailbox: string
    Domain: string
  }
  To: Array<{
    Mailbox: string
    Domain: string
  }>
  Content: {
    Headers: Record<string, string[]>
    Body: string
  }
  Raw: {
    From: string
    To: string[]
    Data: string
  }
  MIME: {
    Parts: Array<{
      Headers: Record<string, string[]>
      Body: string
    }>
  } | null
}

interface MailhogResponse {
  total: number
  count: number
  start: number
  items: MailhogMessage[]
}

declare global {
  namespace Cypress {
    interface Chainable {
      /**
       * Get the latest email from Mailhog for a specific recipient
       * @param email - Recipient email address
       * @param timeout - Maximum time to wait for email (ms)
       */
      getLatestEmail(email: string, timeout?: number): Chainable<MailhogMessage>

      /**
       * Extract magic link from email body
       * @param message - Mailhog message
       */
      extractMagicLink(message: MailhogMessage): Chainable<string>

      /**
       * Clear all emails from Mailhog
       */
      clearMailbox(): Chainable<void>

      /**
       * Wait for email to arrive in Mailhog
       * @param email - Recipient email address
       * @param subject - Email subject (optional)
       * @param timeout - Maximum time to wait (ms)
       */
      waitForEmail(email: string, subject?: string, timeout?: number): Chainable<MailhogMessage>
    }
  }
}

Cypress.Commands.add('clearMailbox', () => {
  const mailhogUrl = Cypress.env('mailhogUrl') || 'http://localhost:8025'
  cy.request('DELETE', `${mailhogUrl}/api/v1/messages`).then((response) => {
    expect(response.status).to.eq(200)
  })
})

Cypress.Commands.add('getLatestEmail', (email: string, timeout = 10000) => {
  const mailhogUrl = Cypress.env('mailhogUrl') || 'http://localhost:8025'
  const startTime = Date.now()

  const checkForEmail = (): Cypress.Chainable<MailhogMessage> => {
    return cy.request<MailhogResponse>(`${mailhogUrl}/api/v2/messages?limit=50`).then((response) => {
      expect(response.status).to.eq(200)

      const messages = response.body.items || []
      const targetEmail = messages.find((msg) => {
        const recipients = msg.To || []
        return recipients.some((to) => `${to.Mailbox}@${to.Domain}` === email)
      })

      if (targetEmail) {
        return cy.wrap(targetEmail)
      }

      // Retry if timeout not reached
      if (Date.now() - startTime < timeout) {
        cy.wait(500)
        return checkForEmail()
      }

      throw new Error(`No email found for ${email} after ${timeout}ms`)
    })
  }

  return checkForEmail()
})

Cypress.Commands.add('waitForEmail', (email: string, subject?: string, timeout = 10000) => {
  const mailhogUrl = Cypress.env('mailhogUrl') || 'http://localhost:8025'
  const startTime = Date.now()

  const checkForEmail = (): Cypress.Chainable<MailhogMessage> => {
    return cy.request<MailhogResponse>(`${mailhogUrl}/api/v2/messages?limit=50`).then((response) => {
      expect(response.status).to.eq(200)

      const messages = response.body.items || []
      const targetEmail = messages.find((msg) => {
        const recipients = msg.To || []
        const matchesRecipient = recipients.some((to) => `${to.Mailbox}@${to.Domain}` === email)

        if (!matchesRecipient) return false

        if (subject) {
          const emailSubject = msg.Content?.Headers?.Subject?.[0] || ''
          return emailSubject.includes(subject)
        }

        return true
      })

      if (targetEmail) {
        return cy.wrap(targetEmail)
      }

      // Retry if timeout not reached
      if (Date.now() - startTime < timeout) {
        cy.wait(500)
        return checkForEmail()
      }

      throw new Error(`No email found for ${email}${subject ? ` with subject "${subject}"` : ''} after ${timeout}ms`)
    })
  }

  return checkForEmail()
})

Cypress.Commands.add('extractMagicLink', (message: MailhogMessage) => {
  let body = message.Content?.Body || ''

  // Try to get HTML part from MIME if available
  if (message.MIME?.Parts && message.MIME.Parts.length > 0) {
    const htmlPart = message.MIME.Parts.find((part) => {
      const contentType = part.Headers['Content-Type']?.[0] || ''
      return contentType.includes('text/html')
    })

    if (htmlPart) {
      body = htmlPart.Body
    } else {
      // Fallback to plain text
      const textPart = message.MIME.Parts.find((part) => {
        const contentType = part.Headers['Content-Type']?.[0] || ''
        return contentType.includes('text/plain')
      })
      if (textPart) {
        body = textPart.Body
      }
    }
  }

  // Decode quoted-printable encoding (=3D -> =, =\n -> remove)
  body = body
    .replace(/=\r?\n/g, '') // Remove soft line breaks
    .replace(/=([0-9A-F]{2})/g, (_, hex) => String.fromCharCode(parseInt(hex, 16)))

  // Decode HTML entities (&amp; -> &, etc.)
  body = body
    .replace(/&amp;/g, '&')
    .replace(/&lt;/g, '<')
    .replace(/&gt;/g, '>')
    .replace(/&quot;/g, '"')
    .replace(/&#39;/g, "'")

  // Extract magic link URL
  // Pattern: http(s)://domain/api/v1/auth/magic-link/verify?token=xxx&redirect=xxx
  const linkRegex = /(https?:\/\/[^\s]+\/api\/v1\/auth\/magic-link\/verify\?[^\s"<]+)/g
  const matches = body.match(linkRegex)

  if (matches && matches.length > 0) {
    return cy.wrap(matches[0])
  }

  throw new Error('No magic link found in email body')
})

export {}
