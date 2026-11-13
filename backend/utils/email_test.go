package utils

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"testing"
)

func setValidSMTPEnvironment(t *testing.T) {
	t.Helper()
	t.Setenv("SMTP_HOST", "localhost")
	t.Setenv("SMTP_PORT", "587")
	t.Setenv("SMTP_USERNAME", "")
	t.Setenv("SMTP_PASSWORD", "")
	t.Setenv("SMTP_SENDER", "sender@example.com")
	t.Setenv("SMTP_STARTTLS", "false")
}

func TestSendEmailNotificationRequiresSMTPConfiguration(t *testing.T) {
	t.Setenv("SMTP_HOST", "")
	t.Setenv("SMTP_PORT", "")
	t.Setenv("SMTP_SENDER", "")

	err := SendEmailNotification([]string{"recipient@example.com"}, "subject", "<p>body</p>")
	if err == nil {
		t.Fatal("expected missing configuration to return an error")
	}
	for _, name := range []string{"SMTP_HOST", "SMTP_PORT", "SMTP_SENDER"} {
		if !strings.Contains(err.Error(), name) {
			t.Errorf("expected error to identify missing %s; got %q", name, err)
		}
	}
}

func TestSendEmailNotificationRejectsInvalidRecipients(t *testing.T) {
	setValidSMTPEnvironment(t)
	if err := SendEmailNotification([]string{"not-an-email"}, "subject", "<p>body</p>"); err == nil {
		t.Fatal("expected invalid recipient to return an error")
	}
}

func TestParseEmailRecipientsDeduplicatesAddresses(t *testing.T) {
	recipients, err := parseEmailRecipients([]string{
		"first@example.com",
		" FIRST@example.com ",
		"second@example.com",
	})
	if err != nil {
		t.Fatalf("parseEmailRecipients returned an error: %v", err)
	}
	if got := len(recipients); got != 2 {
		t.Fatalf("recipient count = %d, want 2", got)
	}
	if recipients[0].Address != "first@example.com" || recipients[1].Address != "second@example.com" {
		t.Errorf("unexpected recipients: %+v", recipients)
	}
}

func TestSendEmailNotificationUsesSMTP(t *testing.T) {
	setValidSMTPEnvironment(t)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen for SMTP test server: %v", err)
	}
	defer listener.Close()

	host, port, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		t.Fatalf("split SMTP test server address: %v", err)
	}
	t.Setenv("SMTP_HOST", host)
	t.Setenv("SMTP_PORT", port)

	messageReceived := make(chan string, 1)
	serverError := make(chan error, 1)
	go func() {
		connection, acceptErr := listener.Accept()
		if acceptErr != nil {
			serverError <- acceptErr
			return
		}
		defer connection.Close()

		reader := bufio.NewReader(connection)
		writer := bufio.NewWriter(connection)
		writeResponse := func(response string) error {
			if _, writeErr := fmt.Fprint(writer, response+"\r\n"); writeErr != nil {
				return writeErr
			}
			return writer.Flush()
		}
		if err := writeResponse("220 localhost ESMTP"); err != nil {
			serverError <- err
			return
		}

		var message strings.Builder
		for {
			line, readErr := reader.ReadString('\n')
			if readErr != nil {
				serverError <- readErr
				return
			}
			command := strings.TrimSpace(line)
			switch {
			case strings.HasPrefix(command, "EHLO "):
				if err := writeResponse("250-localhost"); err != nil {
					serverError <- err
					return
				}
				if err := writeResponse("250 OK"); err != nil {
					serverError <- err
					return
				}
			case strings.HasPrefix(command, "MAIL FROM:"), strings.HasPrefix(command, "RCPT TO:"):
				if err := writeResponse("250 OK"); err != nil {
					serverError <- err
					return
				}
			case command == "DATA":
				if err := writeResponse("354 End data with <CR><LF>.<CR><LF>"); err != nil {
					serverError <- err
					return
				}
				for {
					line, readErr := reader.ReadString('\n')
					if readErr != nil {
						serverError <- readErr
						return
					}
					if strings.TrimSpace(line) == "." {
						break
					}
					message.WriteString(line)
				}
				messageReceived <- message.String()
				if err := writeResponse("250 Message accepted"); err != nil {
					serverError <- err
					return
				}
			case command == "QUIT":
				serverError <- nil
				writeResponse("221 Bye")
				return
			default:
				serverError <- fmt.Errorf("unexpected SMTP command: %s", command)
				return
			}
		}
	}()

	if err := SendEmailNotification(
		[]string{"first@example.com", " FIRST@example.com "},
		"แจ้งซ่อม",
		"<p>รายละเอียด</p>",
	); err != nil {
		t.Fatalf("SendEmailNotification returned an error: %v", err)
	}
	if err := <-serverError; err != nil {
		t.Fatalf("SMTP test server failed: %v", err)
	}

	message := <-messageReceived
	if !strings.Contains(message, "From: <sender@example.com>") {
		t.Errorf("SMTP message missing sender: %q", message)
	}
	if !strings.Contains(message, "To: <first@example.com>") || strings.Contains(message, "FIRST@example.com") {
		t.Errorf("SMTP message has unexpected recipient headers: %q", message)
	}
	if !strings.Contains(message, "Subject: =?UTF-8?q?") || !strings.Contains(message, "<p>รายละเอียด</p>") {
		t.Errorf("SMTP message has unexpected subject or body: %q", message)
	}
}

func TestValidateSMTPConfiguration(t *testing.T) {
	setValidSMTPEnvironment(t)
	if err := ValidateSMTPConfiguration(); err != nil {
		t.Fatalf("expected valid SMTP configuration, got: %v", err)
	}

	t.Setenv("SMTP_PORT", "invalid")
	if err := ValidateSMTPConfiguration(); err == nil || !strings.Contains(err.Error(), "SMTP_PORT") {
		t.Fatalf("expected invalid port to identify SMTP_PORT, got: %v", err)
	}
}

func TestSMTPConfigurationUsesSenderAsDefaultUsername(t *testing.T) {
	setValidSMTPEnvironment(t)
	t.Setenv("SMTP_PASSWORD", "test-password")

	config, err := loadSMTPConfiguration()
	if err != nil {
		t.Fatalf("loadSMTPConfiguration returned an error: %v", err)
	}
	if config.username != "sender@example.com" {
		t.Errorf("username = %q, want sender address", config.username)
	}
}
