package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func setValidBrevoEnvironment(t *testing.T) {
	t.Helper()
	t.Setenv("BREVO_API_KEY", "test-api-key")
	t.Setenv("BREVO_SENDER", "sender@example.com")
}

func TestSendEmailNotificationRequiresBrevoConfiguration(t *testing.T) {
	t.Setenv("BREVO_API_KEY", "")
	t.Setenv("BREVO_SENDER", "")

	err := SendEmailNotification([]string{"recipient@example.com"}, "subject", "<p>body</p>")
	if err == nil {
		t.Fatal("expected missing configuration to return an error")
	}
	for _, name := range []string{"BREVO_API_KEY", "BREVO_SENDER"} {
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
	if recipients[0].Email != "first@example.com" || recipients[1].Email != "second@example.com" {
		t.Errorf("unexpected recipients: %+v", recipients)
	}
}

func TestSendBrevoEmailUsesAPI(t *testing.T) {
	setValidBrevoEnvironment(t)
	config, err := loadBrevoConfiguration()
	if err != nil {
		t.Fatalf("load Brevo configuration: %v", err)
	}

	requestReceived := make(chan brevoEmailRequest, 1)
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			t.Errorf("request method = %q, want POST", request.Method)
		}
		if got := request.Header.Get("api-key"); got != "test-api-key" {
			t.Errorf("api-key header = %q, want configured key", got)
		}
		if got := request.Header.Get("Content-Type"); got != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", got)
		}

		var payload brevoEmailRequest
		if decodeErr := json.NewDecoder(request.Body).Decode(&payload); decodeErr != nil {
			t.Errorf("decode request body: %v", decodeErr)
		}
		requestReceived <- payload
		writer.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	err = sendBrevoEmail(
		server.Client(),
		server.URL,
		config,
		[]brevoRecipient{{Email: "first@example.com"}, {Email: "second@example.com"}},
		"แจ้งซ่อม",
		"<p>รายละเอียด</p>",
	)
	if err != nil {
		t.Fatalf("sendBrevoEmail returned an error: %v", err)
	}

	payload := <-requestReceived
	if payload.Sender.Email != "sender@example.com" {
		t.Errorf("sender = %q, want sender@example.com", payload.Sender.Email)
	}
	if got := len(payload.To); got != 2 {
		t.Fatalf("recipient count = %d, want 2", got)
	}
	if payload.Subject != "แจ้งซ่อม" || payload.HTMLContent != "<p>รายละเอียด</p>" {
		t.Errorf("unexpected email content: subject=%q html=%q", payload.Subject, payload.HTMLContent)
	}
}

func TestSendBrevoEmailReportsAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		http.Error(writer, "sender not verified", http.StatusBadRequest)
	}))
	defer server.Close()

	setValidBrevoEnvironment(t)
	config, err := loadBrevoConfiguration()
	if err != nil {
		t.Fatalf("load Brevo configuration: %v", err)
	}
	err = sendBrevoEmail(
		server.Client(),
		server.URL,
		config,
		[]brevoRecipient{{Email: "recipient@example.com"}},
		"subject",
		"<p>body</p>",
	)
	if err == nil || !strings.Contains(err.Error(), fmt.Sprint(http.StatusBadRequest)) {
		t.Fatalf("expected HTTP error status to be reported, got: %v", err)
	}
}

func TestValidateBrevoConfiguration(t *testing.T) {
	setValidBrevoEnvironment(t)
	if err := ValidateBrevoConfiguration(); err != nil {
		t.Fatalf("expected valid Brevo configuration, got: %v", err)
	}

	t.Setenv("BREVO_SENDER", "not-an-email")
	if err := ValidateBrevoConfiguration(); err == nil || !strings.Contains(err.Error(), "BREVO_SENDER") {
		t.Fatalf("expected invalid sender to identify BREVO_SENDER, got: %v", err)
	}
}
