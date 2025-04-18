package email

import (
	"fmt"
	"net/smtp"
)

type Mailer interface {
	SendVerificationEmail(email, username, verificationLink string) error
}

type EmailConfig struct {
	SMTPServer   string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string
	FromEmail    string
}

func NewMailer(server, port, username, password, from string) Mailer {
	return &EmailConfig{
		SMTPServer:   server,
		SMTPPort:     port,
		SMTPUsername: username,
		SMTPPassword: password,
		FromEmail:    from,
	}
}

// Implement the SendVerificationEmail method
func (ec *EmailConfig) SendVerificationEmail(email, username, verificationLink string) error {
	// Email message
	subject := "Verify your account"
	body := fmt.Sprintf(`
<html>
	<body>
		<p>Hello %s,</p>
		<p>Please verify your account by clicking the link below:</p>
		<p><a href="%s">%s</a></p>
		<p>The link will expire in 24 hours.</p>
		<p>If you didn't register, please ignore this email.</p>
		<br>
		<p>— The InsightForge Team</p>
	</body>
</html>
`, username, verificationLink, verificationLink)
	message := []byte("To: " + email + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-version: 1.0;\r\n" +
		"Content-Type: text/html; charset=\"UTF-8\";\r\n" +
		"\r\n" +
		body)

	// Connect to SMTP server
	auth := smtp.PlainAuth("", ec.SMTPUsername, ec.SMTPPassword, ec.SMTPServer)
	addr := ec.SMTPServer + ":" + ec.SMTPPort

	// Send email
	err := smtp.SendMail(addr, auth, ec.FromEmail, []string{email}, message)
	if err != nil {
		return err
	}

	return nil
}
