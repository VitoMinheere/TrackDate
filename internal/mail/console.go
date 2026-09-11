package mail

import "log"

// ConsoleMailer logs the magic link instead of emailing it. Used when
// SMTP_HOST isn't configured, so the login flow works out of the box for
// local development and first-run testing.
type ConsoleMailer struct{}

func (ConsoleMailer) SendMagicLink(to, verifyURL string) error {
	log.Printf("[mail] SMTP not configured; magic link for %s: %s", to, verifyURL)
	return nil
}
