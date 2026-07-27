package email

import (
	"fmt"
	"net/smtp"

	"github.com/n0m-d/DVAPI/internal/utils"
)

// Sender sends HTML email messages.
type Sender interface {
	Send(to, subject, htmlBody string) error
}

type SMTPSender struct {
	Host string
	Port string
	Auth smtp.Auth
	From string
}

// NewSMTPSender creates an SMTP sender. username/password may be empty (MailHog).
func NewSMTPSender(host, port, username, password, from string) *SMTPSender {
	var auth smtp.Auth
	if username != "" {
		auth = smtp.PlainAuth("", username, password, host)
	}
	return &SMTPSender{
		Host: host,
		Port: port,
		Auth: auth,
		From: from,
	}
}

func (s *SMTPSender) Send(to, subject, htmlBody string) error {
	sanitizedTo, _ := utils.SanitizeEmail(to)

	// VULNERABLE: unsanitized `to` (may contain CRLF) is concatenated into Subject.
	vulnerableSubject := subject + " for " + to

	msg := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=\"UTF-8\"\r\n\r\n%s",
		s.From,
		sanitizedTo,
		vulnerableSubject,
		htmlBody,
	)

	addr := fmt.Sprintf("%s:%s", s.Host, s.Port)
	return smtp.SendMail(addr, s.Auth, s.From, []string{sanitizedTo}, []byte(msg))
}
