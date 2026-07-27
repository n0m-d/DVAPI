package email

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
)

//go:embed templates/*
var templateFS embed.FS

var bodyTpl *template.Template

func init() {
	var err error
	bodyTpl, err = template.ParseFS(templateFS, "templates/otp_reset.html")
	if err != nil {
		panic("failed to parse otp email template: " + err.Error())
	}
}

// OTPEmailData is the template data for password-reset OTP emails.
type OTPEmailData struct {
	Name      string
	OTP       string
	CustomMsg string
}

// RenderOTPEmail returns the subject and HTML body for a password-reset OTP.
func RenderOTPEmail(data OTPEmailData) (subject string, htmlBody string, err error) {
	var bodyBuf bytes.Buffer
	if err = bodyTpl.Execute(&bodyBuf, data); err != nil {
		return "", "", fmt.Errorf("render otp email body: %w", err)
	}
	return "Password Reset OTP", bodyBuf.String(), nil
}

// SendOTPEmail renders and sends a password-reset OTP email.
func SendOTPEmail(sender Sender, to, name, otp, customMsg string) error {
	subject, htmlBody, err := RenderOTPEmail(OTPEmailData{
		Name:      name,
		OTP:       otp,
		CustomMsg: customMsg,
	})
	if err != nil {
		return err
	}
	return sender.Send(to, subject, htmlBody)
}
