package utils

import (
	"net/mail"
	"strings"
)

func SanitizeEmail(s string) (string, error) {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	s, _, _ = strings.Cut(s, "\n")
	s = strings.TrimSpace(s)

	if _, err := mail.ParseAddress(s); err != nil {
		return "", err
	}

	return s, nil
}
