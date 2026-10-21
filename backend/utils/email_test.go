package utils

import (
	"net/mail"
	"strings"
	"testing"
)

func setValidResendEnvironment(t *testing.T) {
	t.Helper()
	t.Setenv("RESEND_API_KEY", "re_123456789")
	t.Setenv("SMTP_SENDER", "onboarding@resend.dev")
}

func TestSendEmailNotificationRequiresConfiguration(t *testing.T) {
	t.Setenv("RESEND_API_KEY", "")
	t.Setenv("SMTP_SENDER", "")

	err := SendEmailNotification([]string{"recipient@example.com"}, "subject", "<p>body</p>")
	if err == nil {
		t.Fatal("expected missing email configuration to return an error")
	}
	for _, name := range []string{"RESEND_API_KEY", "SMTP_SENDER"} {
		if !strings.Contains(err.Error(), name) {
			t.Errorf("expected error to identify missing %s; got %q", name, err)
		}
	}
}

func TestSendEmailNotificationRejectsInvalidRecipients(t *testing.T) {
	setValidResendEnvironment(t)
	if err := SendEmailNotification([]string{"not-an-email"}, "subject", "<p>body</p>"); err == nil {
		t.Fatal("expected invalid recipient to return an error")
	}
}

func TestValidateSMTPConfiguration(t *testing.T) {
	setValidResendEnvironment(t)
	if err := ValidateSMTPConfiguration(); err != nil {
		t.Fatalf("expected valid email configuration, got: %v", err)
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
