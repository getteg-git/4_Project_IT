package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/mail"
	"os"
	"strings"
	"time"
)

const brevoAPIEndpoint = "https://api.brevo.com/v3/smtp/email"

type brevoConfiguration struct {
	apiKey string
	sender *mail.Address
}

func loadBrevoConfiguration() (brevoConfiguration, error) {
	apiKey := strings.TrimSpace(os.Getenv("BREVO_API_KEY"))
	sender := strings.TrimSpace(os.Getenv("BREVO_SENDER"))

	missing := make([]string, 0, 2)
	if apiKey == "" {
		missing = append(missing, "BREVO_API_KEY")
	}
	if sender == "" {
		missing = append(missing, "BREVO_SENDER")
	}
	if len(missing) > 0 {
		return brevoConfiguration{}, fmt.Errorf("email is not configured: missing %s", strings.Join(missing, ", "))
	}

	parsedSender, err := mail.ParseAddress(sender)
	if err != nil {
		return brevoConfiguration{}, fmt.Errorf("invalid BREVO_SENDER: %w", err)
	}
	return brevoConfiguration{apiKey: apiKey, sender: parsedSender}, nil
}

func SendEmailNotification(toRecipients []string, subject string, bodyHTML string) error {
	config, err := loadBrevoConfiguration()
	if err != nil {
		return err
	}

	recipients, err := parseEmailRecipients(toRecipients)
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: 15 * time.Second}
	return sendBrevoEmail(client, brevoAPIEndpoint, config, recipients, subject, bodyHTML)
}

func parseEmailRecipients(toRecipients []string) ([]brevoRecipient, error) {
	recipients := make([]brevoRecipient, 0, len(toRecipients))
	seenRecipients := make(map[string]struct{}, len(toRecipients))
	for _, recipient := range toRecipients {
		recipient = strings.TrimSpace(recipient)
		if recipient == "" {
			continue
		}
		parsed, parseErr := mail.ParseAddress(recipient)
		if parseErr != nil {
			return nil, fmt.Errorf("invalid email recipient %q: %w", recipient, parseErr)
		}
		key := strings.ToLower(parsed.Address)
		if _, exists := seenRecipients[key]; exists {
			continue
		}
		seenRecipients[key] = struct{}{}
		recipients = append(recipients, brevoRecipient{Email: parsed.Address})
	}
	if len(recipients) == 0 {
		return nil, fmt.Errorf("email has no valid recipients")
	}
	return recipients, nil
}

type brevoRecipient struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

type brevoEmailRequest struct {
	Sender      brevoRecipient   `json:"sender"`
	To          []brevoRecipient `json:"to"`
	Subject     string           `json:"subject"`
	HTMLContent string           `json:"htmlContent"`
}

func sendBrevoEmail(client *http.Client, endpoint string, config brevoConfiguration, recipients []brevoRecipient, subject string, bodyHTML string) error {
	payload := brevoEmailRequest{
		Sender:      brevoRecipient{Email: config.sender.Address, Name: "ระบบแจ้งซ่อม SC-SCI"},
		To:          recipients,
		Subject:     subject,
		HTMLContent: bodyHTML,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encode Brevo email request: %w", err)
	}

	request, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create Brevo email request: %w", err)
	}
	request.Header.Set("api-key", config.apiKey)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")

	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("send email via Brevo: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		responseBody, readErr := io.ReadAll(io.LimitReader(response.Body, 4096))
		if readErr != nil {
			return fmt.Errorf("send email via Brevo: unexpected HTTP status %d (could not read response: %w)", response.StatusCode, readErr)
		}
		return fmt.Errorf("send email via Brevo: unexpected HTTP status %d: %s", response.StatusCode, strings.TrimSpace(string(responseBody)))
	}
	return nil
}

func ValidateBrevoConfiguration() error {
	_, err := loadBrevoConfiguration()
	return err
}
