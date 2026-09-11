// Package mail sends the magic-link login email. It defines a small
// Mailer interface so the SMTP implementation can be swapped for a
// stdout-logging one in local development.
package mail

type Mailer interface {
	SendMagicLink(to, verifyURL string) error
}

const magicLinkSubject = "Your Trackdate sign-in link"

func magicLinkBody(verifyURL string) string {
	return "Click the link below to sign in to Trackdate.\r\n\r\n" +
		verifyURL + "\r\n\r\n" +
		"This link expires in 15 minutes and can only be used once.\r\n" +
		"If you didn't request this, you can ignore this email.\r\n"
}
