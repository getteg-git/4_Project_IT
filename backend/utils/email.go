package utils

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/mail"
	"net/smtp"
	"os"
	"strconv"
	"strings"
	"time"
)

const smtpTimeout = 15 * time.Second

func SendEmailNotification(toRecipients []string, subject string, bodyHTML string) error {
	host := strings.TrimSpace(os.Getenv("SMTP_HOST"))
	port := strings.TrimSpace(os.Getenv("SMTP_PORT"))
	sender := strings.TrimSpace(os.Getenv("SMTP_SENDER"))
	password := strings.TrimSpace(os.Getenv("SMTP_PASSWORD"))

	missing := make([]string, 0, 4)
	for name, value := range map[string]string{
		"SMTP_HOST":     host,
		"SMTP_PORT":     port,
		"SMTP_SENDER":   sender,
		"SMTP_PASSWORD": password,
	} {
		if value == "" {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("email is not configured: missing %s", strings.Join(missing, ", "))
	}

	portNumber, err := strconv.Atoi(port)
	if err != nil || portNumber < 1 || portNumber > 65535 {
		return fmt.Errorf("invalid SMTP_PORT: expected a port number from 1 to 65535")
	}

	senderAddress, err := mail.ParseAddress(sender)
	if err != nil {
		return fmt.Errorf("invalid SMTP_SENDER: %w", err)
	}
	recipients := make([]string, 0, len(toRecipients))
	for _, recipient := range toRecipients {
		recipient = strings.TrimSpace(recipient)
		if recipient == "" {
			continue
		}
		parsed, parseErr := mail.ParseAddress(recipient)
		if parseErr != nil {
			return fmt.Errorf("invalid email recipient %q: %w", recipient, parseErr)
		}
		recipients = append(recipients, parsed.Address)
	}
	if len(recipients) == 0 {
		return fmt.Errorf("email has no valid recipients")
	}

	connection, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), smtpTimeout)
	if err != nil {
		return fmt.Errorf("connect to SMTP server: %w", err)
	}
	defer connection.Close()
	if err := connection.SetDeadline(time.Now().Add(smtpTimeout)); err != nil {
		return fmt.Errorf("set SMTP connection deadline: %w", err)
	}

	tlsConfig := &tls.Config{ServerName: host, MinVersion: tls.VersionTLS12}
	if portNumber == 465 {
		tlsConnection := tls.Client(connection, tlsConfig)
		if err := tlsConnection.Handshake(); err != nil {
			return fmt.Errorf("start implicit TLS with SMTP server: %w", err)
		}
		connection = tlsConnection
	}

	client, err := smtp.NewClient(connection, host)
	if err != nil {
		return fmt.Errorf("start SMTP session: %w", err)
	}
	defer client.Close()

	if portNumber != 465 {
		if supported, _ := client.Extension("STARTTLS"); supported {
			if err := client.StartTLS(tlsConfig); err != nil {
				return fmt.Errorf("start TLS with SMTP server: %w", err)
			}
		}
	}

	auth := smtp.PlainAuth("", senderAddress.Address, password, host)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("authenticate with SMTP server: %w", err)
	}
	if err := client.Mail(senderAddress.Address); err != nil {
		return fmt.Errorf("set SMTP sender: %w", err)
	}
	for _, recipient := range recipients {
		if err := client.Rcpt(recipient); err != nil {
			return fmt.Errorf("set SMTP recipient %q: %w", recipient, err)
		}
	}

	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("start SMTP message: %w", err)
	}
	message := buildEmailMessage(senderAddress, recipients, subject, bodyHTML)
	if _, err := io.WriteString(writer, message); err != nil {
		_ = writer.Close()
		return fmt.Errorf("write SMTP message: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("submit SMTP message: %w", err)
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
		"To: " + strings.Join(recipients, ", ") + "\r\n" +
		"Subject: " + encodedSubject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n" +
		"Content-Transfer-Encoding: base64\r\n\r\n" +
		wrappedBody.String() + "\r\n"
}

func mimeHeader(value string) string {
	return "=?UTF-8?B?" + base64.StdEncoding.EncodeToString([]byte(value)) + "?="
}
