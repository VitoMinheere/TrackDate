package mail

import (
	"fmt"
	"net/smtp"
)

// SMTPMailer sends mail through a standard SMTP relay. smtp.SendMail
// negotiates STARTTLS automatically when the server advertises it, so this
// works unmodified against most providers (Mailgun, SES, Fastmail, a
// self-hosted relay, ...) on port 587.
type SMTPMailer struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

func (m *SMTPMailer) SendMagicLink(to, verifyURL string) error {
	addr := fmt.Sprintf("%s:%d", m.Host, m.Port)

	var auth smtp.Auth
	if m.Username != "" {
		auth = smtp.PlainAuth("", m.Username, m.Password, m.Host)
	}

	msg := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		m.From, to, magicLinkSubject, magicLinkBody(verifyURL),
	)

	return smtp.SendMail(addr, auth, m.From, []string{to}, []byte(msg))
}
