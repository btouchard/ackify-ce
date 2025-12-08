// SPDX-License-Identifier: AGPL-3.0-or-later
package email

import (
	"context"

	"github.com/btouchard/ackify-ce/backend/internal/infrastructure/i18n"
)

func SendEmail(ctx context.Context, sender Sender, template string, to []string, locale string, subject string, data map[string]any) error {
	msg := Message{
		To:       to,
		Subject:  subject,
		Template: template,
		Locale:   locale,
		Data:     data,
	}

	return sender.Send(ctx, msg)
}

func SendSignatureReminderEmail(ctx context.Context, sender Sender, i18nService *i18n.I18n, to []string, locale, docID, docURL, signURL, recipientName string) error {
	data := map[string]any{
		"DocID":         docID,
		"DocURL":        docURL,
		"SignURL":       signURL,
		"RecipientName": recipientName,
	}

	// Get translated subject using i18n
	subject := "Document Reading Confirmation Reminder" // Fallback
	if i18nService != nil {
		subject = i18nService.T(locale, "email.reminder.subject")
	}

	return SendEmail(ctx, sender, "signature_reminder", to, locale, subject, data)
}
