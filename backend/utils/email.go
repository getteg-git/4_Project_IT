package utils

import (
	"bytes"
	"fmt"
	"mime"
	"net"
	"net/mail"
	"net/smtp"
	"os"
	"strconv"
	"strings"
)

type smtpConfiguration struct {
	host     string
	port     int
	sender   *mail.Address
	password string
}

func loadSMTPConfiguration() (smtpConfiguration, error) {
	config := smtpConfiguration{
		host:     strings.TrimSpace(os.Getenv("SMTP_HOST")),
		password: os.Getenv("SMTP_PASSWORD"),
	}
	sender := strings.TrimSpace(os.Getenv("SMTP_SENDER"))
	port := strings.TrimSpace(os.Getenv("SMTP_PORT"))

	missing := make([]string, 0, 4)
	if config.host == "" {
		missing = append(missing, "SMTP_HOST")
	}
	if port == "" {
		missing = append(missing, "SMTP_PORT")
	}
	if sender == "" {
		missing = append(missing, "SMTP_SENDER")
	}
	if config.password == "" {
		missing = append(missing, "SMTP_PASSWORD")
	}
	if len(missing) > 0 {
		return smtpConfiguration{}, fmt.Errorf("email is not configured: missing %s", strings.Join(missing, ", "))
	}

	if strings.Contains(config.host, "://") {
		return smtpConfiguration{}, fmt.Errorf("invalid SMTP_HOST: expected a hostname, not a URL")
	}
	config.port, _ = strconv.Atoi(port)
	if config.port < 1 || config.port > 65535 {
		return smtpConfiguration{}, fmt.Errorf("invalid SMTP_PORT: expected a port between 1 and 65535")
	}

	parsedSender, err := mail.ParseAddress(sender)
	if err != nil {
		return smtpConfiguration{}, fmt.Errorf("invalid SMTP_SENDER: %w", err)
	}
	config.sender = parsedSender
	return config, nil
}

func SendEmailNotification(toRecipients []string, subject string, bodyHTML string) error {
	config, err := loadSMTPConfiguration()
	if err != nil {
		return err
	}

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

	message := buildSMTPMessage(config.sender, subject, bodyHTML)
	address := net.JoinHostPort(config.host, strconv.Itoa(config.port))
	auth := smtp.PlainAuth("", config.sender.Address, config.password, config.host)
	if err := smtp.SendMail(address, auth, config.sender.Address, recipients, message); err != nil {
		return fmt.Errorf("send email via SMTP: %w", err)
	}
	return nil
}

func buildSMTPMessage(sender *mail.Address, subject string, bodyHTML string) []byte {
	var message bytes.Buffer
	message.WriteString("From: ")
	message.WriteString((&mail.Address{Name: "ระบบแจ้งซ่อม SC-SCI", Address: sender.Address}).String())
	message.WriteString("\r\nTo: undisclosed-recipients:;\r\nSubject: ")
	message.WriteString(mime.QEncoding.Encode("UTF-8", subject))
	message.WriteString("\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\nContent-Transfer-Encoding: 8bit\r\n\r\n")
	message.WriteString(bodyHTML)
	return message.Bytes()
}

func ValidateSMTPConfiguration() error {
	_, err := loadSMTPConfiguration()
	return err
}
