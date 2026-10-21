package utils

import (
	"encoding/base64"
	"fmt"
	"net/mail"
	"net/smtp"
	"os"
	"strconv"
	"strings"
)

func SendEmailNotification(toRecipients []string, subject string, bodyHTML string) error {
	host := strings.TrimSpace(os.Getenv("SMTP_HOST"))
	port := strings.TrimSpace(os.Getenv("SMTP_PORT"))
	sender := strings.TrimSpace(os.Getenv("SMTP_SENDER"))
	password := os.Getenv("SMTP_PASSWORD")

	if err := ValidateSMTPConfiguration(); err != nil {
		return err
	}

	senderAddress, err := mail.ParseAddress(sender)
	if err != nil {
		return fmt.Errorf("invalid sender address: %w", err)
	}

	// กรองอีเมลผู้รับที่ซ้ำหรือว่างเปล่า (โค้ดเดิมของเพื่อน)
	recipients := make([]string, 0, len(toRecipients))
	seenRecipients := make(map[string]struct{}, len(toRecipients))
	for _, recipient := range toRecipients {
		recipient = strings.TrimSpace(recipient)
		if recipient == "" {
			continue
		}
		parsed, parseErr := mail.ParseAddress(recipient)
		if parseErr != nil {
			return fmt.Errorf("invalid email recipient %q: %w", recipient, parseErr)
		}
		key := strings.ToLower(parsed.Address)
		if _, exists := seenRecipients[key]; exists {
			continue
		}
		seenRecipients[key] = struct{}{}
		recipients = append(recipients, parsed.Address)
	}
	if len(recipients) == 0 {
		return fmt.Errorf("email has no valid recipients")
	}

	auth := smtp.PlainAuth("", senderAddress.Address, password, host)
	serverAddress := host + ":" + port
	if err := smtp.SendMail(serverAddress, auth, senderAddress.Address, recipients, []byte(
		buildEmailMessage(senderAddress, recipients, subject, bodyHTML),
	)); err != nil {
		return fmt.Errorf("send email via SMTP server %s: %w", serverAddress, err)
	}

	return nil
}

func ValidateSMTPConfiguration() error {
	host := strings.TrimSpace(os.Getenv("SMTP_HOST"))
	port := strings.TrimSpace(os.Getenv("SMTP_PORT"))
	sender := strings.TrimSpace(os.Getenv("SMTP_SENDER"))
	password := os.Getenv("SMTP_PASSWORD")

	missing := make([]string, 0, 4)
	if host == "" {
		missing = append(missing, "SMTP_HOST")
	}
	if port == "" {
		missing = append(missing, "SMTP_PORT")
	}
	if sender == "" {
		missing = append(missing, "SMTP_SENDER")
	}
	if strings.TrimSpace(password) == "" {
		missing = append(missing, "SMTP_PASSWORD")
	}

	if len(missing) > 0 {
		return fmt.Errorf("email is not configured: missing %s", strings.Join(missing, ", "))
	}

	portNumber, err := strconv.Atoi(port)
	if err != nil || portNumber < 1 || portNumber > 65535 {
		return fmt.Errorf("invalid SMTP_PORT: must be a number between 1 and 65535")
	}

	if _, err := mail.ParseAddress(sender); err != nil {
		return fmt.Errorf("invalid SMTP_SENDER: %w", err)
	}
	return nil
}

func buildEmailMessage(sender *mail.Address, recipients []string, subject, bodyHTML string) string {
	displayName := "ระบบแจ้งซ่อม SC-SCI"
	from := mimeHeader(displayName) + " <" + sender.Address + ">"
	encodedSubject := mimeHeader(subject)
	encodedBody := base64.StdEncoding.EncodeToString([]byte(bodyHTML))

	var wrappedBody strings.Builder
	for len(encodedBody) > 76 {
		wrappedBody.WriteString(encodedBody[:76])
		wrappedBody.WriteString("\r\n")
		encodedBody = encodedBody[76:]
	}
	wrappedBody.WriteString(encodedBody)

	return "From: " + from + "\r\n" +
		"To: undisclosed-recipients:;\r\n" +
		"Subject: " + encodedSubject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n" +
		"Content-Transfer-Encoding: base64\r\n\r\n" +
		wrappedBody.String() + "\r\n"
}

func mimeHeader(value string) string {
	return "=?UTF-8?B?" + base64.StdEncoding.EncodeToString([]byte(value)) + "?="
}
