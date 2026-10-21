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
	t.Setenv("SMTP_SENDER", "sender@example.com")
	t.Setenv("SMTP_PASSWORD", "test-password")
}

func TestSendEmailNotificationRequiresSMTPConfiguration(t *testing.T) {
	for _, name := range []string{"SMTP_HOST", "SMTP_PORT", "SMTP_SENDER", "SMTP_PASSWORD"} {
		t.Setenv(name, "")
	}

	err := SendEmailNotification([]string{"recipient@example.com"}, "subject", "<p>body</p>")
	if err == nil {
		t.Fatal("expected missing configuration to return an error")
	}
	for _, name := range []string{"SMTP_HOST", "SMTP_PORT", "SMTP_SENDER", "SMTP_PASSWORD"} {
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

func TestSendEmailNotificationUsesSMTP(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen for SMTP test server: %v", err)
	}
	defer listener.Close()

	setValidSMTPEnvironment(t)
	t.Setenv("SMTP_PORT", fmt.Sprint(listener.Addr().(*net.TCPAddr).Port))

	type smtpResult struct {
		recipients []string
		message    string
		err        error
	}
	resultChannel := make(chan smtpResult, 1)
	go func() {
		connection, acceptErr := listener.Accept()
		if acceptErr != nil {
			resultChannel <- smtpResult{err: acceptErr}
			return
		}
		defer connection.Close()

		result := smtpResult{}
		reader := bufio.NewReader(connection)
		writeReply := func(reply string) bool {
			_, writeErr := fmt.Fprint(connection, reply+"\r\n")
			if writeErr != nil {
				result.err = writeErr
				return false
			}
			return true
		}

		if !writeReply("220 localhost ESMTP") {
			resultChannel <- result
			return
		}
		for {
			line, readErr := reader.ReadString('\n')
			if readErr != nil {
				result.err = readErr
				resultChannel <- result
				return
			}
			command := strings.TrimSpace(line)
			switch {
			case strings.HasPrefix(command, "EHLO "), strings.HasPrefix(command, "HELO "):
				if !writeReply("250-localhost") || !writeReply("250 AUTH PLAIN") {
					resultChannel <- result
					return
				}
			case strings.HasPrefix(command, "AUTH PLAIN "):
				if !writeReply("235 2.7.0 authenticated") {
					resultChannel <- result
					return
				}
			case strings.HasPrefix(command, "MAIL FROM:"):
				if !writeReply("250 2.1.0 sender accepted") {
					resultChannel <- result
					return
				}
			case strings.HasPrefix(command, "RCPT TO:"):
				result.recipients = append(result.recipients, strings.TrimSuffix(strings.TrimPrefix(command, "RCPT TO:<"), ">"))
				if !writeReply("250 2.1.5 recipient accepted") {
					resultChannel <- result
					return
				}
			case command == "DATA":
				if !writeReply("354 end with dot") {
					resultChannel <- result
					return
				}
				var message strings.Builder
				for {
					messageLine, messageErr := reader.ReadString('\n')
					if messageErr != nil {
						result.err = messageErr
						resultChannel <- result
						return
					}
					if messageLine == ".\r\n" {
						break
					}
					message.WriteString(strings.TrimPrefix(messageLine, ".."))
				}
				result.message = message.String()
				if !writeReply("250 2.0.0 message accepted") {
					resultChannel <- result
					return
				}
			case command == "QUIT":
				writeReply("221 2.0.0 goodbye")
				resultChannel <- result
				return
			default:
				if !writeReply("500 unexpected command") {
					resultChannel <- result
					return
				}
			}
		}
	}()

	err = SendEmailNotification(
		[]string{"first@example.com", "FIRST@example.com", "second@example.com"},
		"แจ้งซ่อม",
		"<p>รายละเอียด</p>",
	)
	if err != nil {
		t.Fatalf("SendEmailNotification returned an error: %v", err)
	}

	result := <-resultChannel
	if result.err != nil {
		t.Fatalf("SMTP test server failed: %v", result.err)
	}
	if got, want := strings.Join(result.recipients, ","), "first@example.com,second@example.com"; got != want {
		t.Errorf("SMTP recipients = %q, want %q", got, want)
	}
	for _, expected := range []string{
		"sender@example.com",
		"Subject: =?",
		"Content-Type: text/html; charset=UTF-8",
		"<p>รายละเอียด</p>",
	} {
		if !strings.Contains(result.message, expected) {
			t.Errorf("SMTP message does not contain %q", expected)
		}
	}
}

func TestValidateSMTPConfiguration(t *testing.T) {
	setValidSMTPEnvironment(t)
	if err := ValidateSMTPConfiguration(); err != nil {
		t.Fatalf("expected valid SMTP configuration, got: %v", err)
	}

	t.Setenv("SMTP_SENDER", "not-an-email")
	if err := ValidateSMTPConfiguration(); err == nil || !strings.Contains(err.Error(), "SMTP_SENDER") {
		t.Fatalf("expected invalid sender to identify SMTP_SENDER, got: %v", err)
	}

	setValidSMTPEnvironment(t)
	t.Setenv("SMTP_PORT", "70000")
	if err := ValidateSMTPConfiguration(); err == nil || !strings.Contains(err.Error(), "SMTP_PORT") {
		t.Fatalf("expected invalid port to identify SMTP_PORT, got: %v", err)
	}
}
