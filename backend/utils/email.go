package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/mail"
	"os"
	"strings"
	"time"
)

const httpTimeout = 15 * time.Second

// โครงสร้าง Request สำหรับ Brevo HTTP API
type BrevoSender struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type BrevoRecipient struct {
	Email string `json:"email"`
}

type BrevoEmailRequest struct {
	Sender      BrevoSender      `json:"sender"`
	To          []BrevoRecipient `json:"to"`
	Subject     string           `json:"subject"`
	HtmlContent string           `json:"htmlContent"`
}

func SendEmailNotification(toRecipients []string, subject string, bodyHTML string) error {
	apiKey := strings.TrimSpace(os.Getenv("BREVO_API_KEY"))
	senderEmail := strings.TrimSpace(os.Getenv("SMTP_SENDER"))

	if err := ValidateSMTPConfiguration(); err != nil {
		return err
	}

	senderAddress, err := mail.ParseAddress(senderEmail)
	if err != nil {
		return fmt.Errorf("invalid sender address: %w", err)
	}

	// กรองอีเมลผู้รับที่ซ้ำหรือว่างเปล่า
	var recipients []BrevoRecipient
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
		recipients = append(recipients, BrevoRecipient{Email: parsed.Address})
	}

	if len(recipients) == 0 {
		return fmt.Errorf("email has no valid recipients")
	}

	// สร้าง Request Body สำหรับ Brevo API
	emailReq := BrevoEmailRequest{
		Sender:      BrevoSender{Name: "ระบบแจ้งซ่อม SC-SCI", Email: senderAddress.Address},
		To:          recipients,
		Subject:     subject,
		HtmlContent: bodyHTML,
	}

	jsonBytes, err := json.Marshal(emailReq)
	if err != nil {
		return fmt.Errorf("marshal email request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.brevo.com/v3/smtp/email", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return fmt.Errorf("create http request to brevo: %w", err)
	}

	req.Header.Set("api-key", apiKey)
	req.Header.Set("Content-Type", "application/json")

	httpClient := &http.Client{Timeout: httpTimeout}
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send http request to brevo: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("brevo api returned error status: %d", resp.StatusCode)
	}

	return nil
}

func ValidateSMTPConfiguration() error {
	apiKey := strings.TrimSpace(os.Getenv("BREVO_API_KEY"))
	sender := strings.TrimSpace(os.Getenv("SMTP_SENDER"))

	missing := make([]string, 0, 2)
	if apiKey == "" {
		missing = append(missing, "BREVO_API_KEY")
	}
	if sender == "" {
		missing = append(missing, "SMTP_SENDER")
	}

	if len(missing) > 0 {
		return fmt.Errorf("email is not configured: missing %s", strings.Join(missing, ", "))
	}

	if _, err := mail.ParseAddress(sender); err != nil {
		return fmt.Errorf("invalid SMTP_SENDER: %w", err)
	}
	return nil
}
