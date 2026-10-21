package utils

import (
	"strings"
	"testing"
)

func setValidBrevoEnvironment(t *testing.T) {
	t.Helper()
	t.Setenv("BREVO_API_KEY", "xkeysib-test-key")
	t.Setenv("SMTP_SENDER", "sender@example.com")
}

func TestSendEmailNotificationRequiresConfiguration(t *testing.T) {
	t.Setenv("BREVO_API_KEY", "")
	t.Setenv("SMTP_SENDER", "")

	err := SendEmailNotification([]string{"recipient@example.com"}, "subject", "<p>body</p>")
	if err == nil {
		t.Fatal("expected missing configuration to return an error")
	}
	for _, name := range []string{"BREVO_API_KEY", "SMTP_SENDER"} {
		if !strings.Contains(err.Error(), name) {
			t.Errorf("expected error to identify missing %s; got %q", name, err)
		}
	}
}

func TestSendEmailNotificationRejectsInvalidRecipients(t *testing.T) {
	setValidBrevoEnvironment(t)
	if err := SendEmailNotification([]string{"not-an-email"}, "subject", "<p>body</p>"); err == nil {
		t.Fatal("expected invalid recipient to return an error")
	}
}

func TestValidateSMTPConfiguration(t *testing.T) {
	setValidBrevoEnvironment(t)
	if err := ValidateSMTPConfiguration(); err != nil {
		t.Fatalf("expected valid email configuration, got: %v", err)
	}

	t.Setenv("SMTP_SENDER", "not-an-email")
	if err := ValidateSMTPConfiguration(); err == nil || !strings.Contains(err.Error(), "SMTP_SENDER") {
		t.Fatalf("expected invalid sender to identify SMTP_SENDER, got: %v", err)
	}
}
