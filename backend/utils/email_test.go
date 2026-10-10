package utils

import (
	"net/mail"
	"strings"
	"testing"
)

func setValidSMTPEnvironment(t *testing.T) {
	t.Helper()
	t.Setenv("SMTP_HOST", "smtp.example.com")
	t.Setenv("SMTP_PORT", "587")
	t.Setenv("SMTP_SENDER", "sender@example.com")
	t.Setenv("SMTP_PASSWORD", "test-password")
}

func TestSendEmailNotificationRequiresSMTPConfiguration(t *testing.T) {
	t.Setenv("SMTP_HOST", "")
	t.Setenv("SMTP_PORT", "")
	t.Setenv("SMTP_SENDER", "")
	t.Setenv("SMTP_PASSWORD", "")

	err := SendEmailNotification([]string{"recipient@example.com"}, "subject", "<p>body</p>")
	if err == nil {
		t.Fatal("expected missing SMTP configuration to return an error")
	}
	for _, name := range []string{"SMTP_HOST", "SMTP_PORT", "SMTP_SENDER", "SMTP_PASSWORD"} {
		if !strings.Contains(err.Error(), name) {
			t.Errorf("expected error to identify missing %s; got %q", name, err)
		}
	}
}

func TestSendEmailNotificationRejectsInvalidPortAndRecipients(t *testing.T) {
	setValidSMTPEnvironment(t)
	t.Setenv("SMTP_PORT", "not-a-port")
	if err := SendEmailNotification([]string{"recipient@example.com"}, "subject", "<p>body</p>"); err == nil {
		t.Fatal("expected invalid SMTP port to return an error")
	}

	t.Setenv("SMTP_PORT", "587")
	if err := SendEmailNotification([]string{"not-an-email"}, "subject", "<p>body</p>"); err == nil {
		t.Fatal("expected invalid recipient to return an error")
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
}

func TestBuildEmailMessageEncodesUTF8HeadersAndBody(t *testing.T) {
	sender, err := mail.ParseAddress("sender@example.com")
	if err != nil {
		t.Fatal(err)
	}
	message := buildEmailMessage(sender, []string{"recipient@example.com"}, "แจ้งเตือน", "<p>ข้อความภาษาไทย</p>")

	for _, want := range []string{
		"Subject: =?UTF-8?B?",
		"Content-Transfer-Encoding: base64",
		"To: undisclosed-recipients:;",
	} {
		if !strings.Contains(message, want) {
			t.Errorf("expected email message to contain %q", want)
		}
	}
}
