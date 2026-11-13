package utils

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"mime"
	"net"
	"net/mail"
	"net/smtp"
	"os"
	"strconv"
	"strings"
	"time"
)

type smtpConfiguration struct {
	host     string
	port     string
	username string
	password string
	sender   *mail.Address
	startTLS bool
}

func loadSMTPConfiguration() (smtpConfiguration, error) {
	host := strings.TrimSpace(os.Getenv("SMTP_HOST"))
	port := strings.TrimSpace(os.Getenv("SMTP_PORT"))
	username := strings.TrimSpace(os.Getenv("SMTP_USERNAME"))
	password := os.Getenv("SMTP_PASSWORD")
	sender := strings.TrimSpace(os.Getenv("SMTP_SENDER"))
	startTLS := true

	missing := make([]string, 0, 3)
	if host == "" {
		missing = append(missing, "SMTP_HOST")
	}
	if port == "" {
		missing = append(missing, "SMTP_PORT")
	}
	if sender == "" {
		missing = append(missing, "SMTP_SENDER")
	}
	if len(missing) > 0 {
		return smtpConfiguration{}, fmt.Errorf("email is not configured: missing %s", strings.Join(missing, ", "))
	}

	portNumber, err := strconv.Atoi(port)
	if err != nil || portNumber < 1 || portNumber > 65535 {
		return smtpConfiguration{}, fmt.Errorf("invalid SMTP_PORT: expected a port between 1 and 65535")
	}
	if value := strings.TrimSpace(os.Getenv("SMTP_STARTTLS")); value != "" {
		startTLS, err = strconv.ParseBool(value)
		if err != nil {
			return smtpConfiguration{}, fmt.Errorf("invalid SMTP_STARTTLS: expected true or false")
		}
	}

	parsedSender, err := mail.ParseAddress(sender)
	if err != nil {
		return smtpConfiguration{}, fmt.Errorf("invalid SMTP_SENDER: %w", err)
	}
	if username == "" && password != "" {
		username = parsedSender.Address
	}
	if (username == "") != (password == "") {
		return smtpConfiguration{}, fmt.Errorf("SMTP_USERNAME and SMTP_PASSWORD must both be configured or both be empty")
	}
	return smtpConfiguration{
		host:     host,
		port:     port,
		username: username,
		password: password,
		sender:   parsedSender,
		startTLS: startTLS,
	}, nil
}

func SendEmailNotification(toRecipients []string, subject string, bodyHTML string) error {
	config, err := loadSMTPConfiguration()
	if err != nil {
		return err
	}

	recipients, err := parseEmailRecipients(toRecipients)
	if err != nil {
		return err
	}
	if strings.ContainsAny(subject, "\r\n") {
		return fmt.Errorf("email subject must not contain line breaks")
	}

	message := buildSMTPMessage(config.sender, recipients, subject, bodyHTML)
	return sendSMTPEmail(config, recipients, message)
}

func parseEmailRecipients(toRecipients []string) ([]*mail.Address, error) {
	recipients := make([]*mail.Address, 0, len(toRecipients))
	seenRecipients := make(map[string]struct{}, len(toRecipients))
	for _, recipient := range toRecipients {
		recipient = strings.TrimSpace(recipient)
		if recipient == "" {
			continue
		}
		parsed, err := mail.ParseAddress(recipient)
		if err != nil {
			return nil, fmt.Errorf("invalid email recipient %q: %w", recipient, err)
		}
		key := strings.ToLower(parsed.Address)
		if _, exists := seenRecipients[key]; exists {
			continue
		}
		seenRecipients[key] = struct{}{}
		recipients = append(recipients, parsed)
	}
	if len(recipients) == 0 {
		return nil, fmt.Errorf("email has no valid recipients")
	}
	return recipients, nil
}

func buildSMTPMessage(sender *mail.Address, recipients []*mail.Address, subject string, bodyHTML string) []byte {
	recipientHeaders := make([]string, 0, len(recipients))
	for _, recipient := range recipients {
		recipientHeaders = append(recipientHeaders, recipient.String())
	}
	bodyHTML = strings.ReplaceAll(strings.ReplaceAll(bodyHTML, "\r\n", "\n"), "\n", "\r\n")

	return []byte(fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\nContent-Transfer-Encoding: 8bit\r\n\r\n%s",
		sender.String(),
		strings.Join(recipientHeaders, ", "),
		mime.QEncoding.Encode("UTF-8", subject),
		bodyHTML,
	))
}

func sendSMTPEmail(config smtpConfiguration, recipients []*mail.Address, message []byte) error {
	address := net.JoinHostPort(config.host, config.port)
	connection, err := net.DialTimeout("tcp", address, 15*time.Second)
	if err != nil {
		return fmt.Errorf("connect to SMTP server: %w", err)
	}
	defer connection.Close()
	if err := connection.SetDeadline(time.Now().Add(30 * time.Second)); err != nil {
		return fmt.Errorf("set SMTP connection deadline: %w", err)
	}

	client, err := smtp.NewClient(connection, config.host)
	if err != nil {
		return fmt.Errorf("create SMTP client: %w", err)
	}
	defer client.Close()

	if config.startTLS {
		if ok, _ := client.Extension("STARTTLS"); !ok {
			return fmt.Errorf("SMTP server does not support STARTTLS")
		}
		if err := client.StartTLS(&tls.Config{
			ServerName: config.host,
			MinVersion: tls.VersionTLS12,
		}); err != nil {
			return fmt.Errorf("start SMTP TLS: %w", err)
		}
	}

	if config.username != "" {
		if err := client.Auth(smtp.PlainAuth("", config.username, config.password, config.host)); err != nil {
			return fmt.Errorf("authenticate with SMTP server: %w", err)
		}
	}
	if err := client.Mail(config.sender.Address); err != nil {
		return fmt.Errorf("set SMTP sender: %w", err)
	}
	for _, recipient := range recipients {
		if err := client.Rcpt(recipient.Address); err != nil {
			return fmt.Errorf("set SMTP recipient %q: %w", recipient.Address, err)
		}
	}

	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("begin SMTP message: %w", err)
	}
	if _, err := io.Copy(writer, bytes.NewReader(message)); err != nil {
		writer.Close()
		return fmt.Errorf("write SMTP message: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("finish SMTP message: %w", err)
	}
	if err := client.Quit(); err != nil {
		return fmt.Errorf("close SMTP session: %w", err)
	}
	return nil
}

func ValidateSMTPConfiguration() error {
	_, err := loadSMTPConfiguration()
	return err
}
