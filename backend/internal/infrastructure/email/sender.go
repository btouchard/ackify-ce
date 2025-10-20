// SPDX-License-Identifier: AGPL-3.0-or-later
package email

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	mail "github.com/go-mail/mail/v2"

	"github.com/btouchard/ackify-ce/backend/internal/infrastructure/config"
	"github.com/btouchard/ackify-ce/backend/pkg/logger"
)

type Sender interface {
	Send(ctx context.Context, msg Message) error
}

type Message struct {
	To       []string
	Cc       []string
	Bcc      []string
	Subject  string
	Template string
	Locale   string
	Data     map[string]any
	Headers  map[string]string
}

type SMTPSender struct {
	config   config.MailConfig
	renderer *Renderer
}

func NewSMTPSender(cfg config.MailConfig, renderer *Renderer) *SMTPSender {
	return &SMTPSender{
		config:   cfg,
		renderer: renderer,
	}
}

func (s *SMTPSender) Send(ctx context.Context, msg Message) error {
	if s.config.Host == "" {
		logger.Logger.Info("SMTP not configured, email not sent", "template", msg.Template)
		return nil
	}

	htmlBody, textBody, err := s.renderer.Render(msg.Template, msg.Locale, msg.Data)
	if err != nil {
		return fmt.Errorf("failed to render email template: %w", err)
	}

	m := mail.NewMessage()

	from := s.config.From
	if from == "" {
		return fmt.Errorf("ACKIFY_MAIL_FROM not set")
	}
	m.SetHeader("From", m.FormatAddress(from, s.config.FromName))

	if len(msg.To) == 0 {
		return fmt.Errorf("no recipients specified")
	}
	m.SetHeader("To", msg.To...)

	if len(msg.Cc) > 0 {
		m.SetHeader("Cc", msg.Cc...)
	}

	if len(msg.Bcc) > 0 {
		m.SetHeader("Bcc", msg.Bcc...)
	}

	subject := msg.Subject
	if s.config.SubjectPrefix != "" {
		subject = s.config.SubjectPrefix + subject
	}
	m.SetHeader("Subject", subject)

	for key, value := range msg.Headers {
		m.SetHeader(key, value)
	}

	m.SetBody("text/plain", textBody)
	m.AddAlternative("text/html", htmlBody)

	timeout, err := time.ParseDuration(s.config.Timeout)
	if err != nil {
		timeout = 10 * time.Second
	}

	d := mail.NewDialer(s.config.Host, s.config.Port, s.config.Username, s.config.Password)

	if s.config.TLS {
		d.SSL = true
	}

	if s.config.StartTLS {
		d.TLSConfig = &tls.Config{ServerName: s.config.Host}
	}

	d.Timeout = timeout

	logger.Logger.Info("Sending email", "to", msg.To, "template", msg.Template, "locale", msg.Locale)

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	logger.Logger.Info("Email sent successfully", "to", msg.To)
	return nil
}
