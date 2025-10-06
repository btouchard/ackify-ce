// SPDX-License-Identifier: AGPL-3.0-or-later
package email

import (
	"context"
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

func SendSignatureReminderEmail(ctx context.Context, sender Sender, to []string, locale, docID, docURL, signURL string) error {
	data := map[string]any{
		"DocID":   docID,
		"DocURL":  docURL,
		"SignURL": signURL,
	}

	subject := "Reminder: Document signature required"
	if locale == "fr" {
		subject = "Rappel : Signature de document requise"
	}

	return SendEmail(ctx, sender, "signature_reminder", to, locale, subject, data)
}
