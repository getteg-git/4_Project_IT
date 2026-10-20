package utils

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/mail"
	"os"
	"strings"
	"time"
)

const httpTimeout = 15 * time.Second

type ResendEmailRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
}

func SendEmailNotification(toRecipients []string, subject string, bodyHTML string) error {
	apiKey := strings.TrimSpace(os.Getenv("RESEND_API_KEY"))
	sender := strings.TrimSpace(os.Getenv("SMTP_SENDER"))

	if err := ValidateSMTPConfiguration(); err != nil {
		return err
	}

	senderAddress, err := mail.ParseAddress(sender)
	if err != nil {
		return fmt.Errorf("invalid sender address: %w", err)
	}

	// กรองอีเมลผู้รับที่ซ้ำหรือว่างเปล่า (โค้ดเดิมของเพื่อน)
	recipients := make([]string, 0, len(toRecipients))
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
		recipients = append(recipients, parsed.Address)
	}
	if len(recipients) == 0 {
		return fmt.Errorf("email has no valid recipients")
	}

	displayName := "ระบบแจ้งซ่อม SC-SCI"
	fromField := fmt.Sprintf("%s <%s>", displayName, senderAddress.Address)

	// สร้าง Request Body สำหรับ Resend API
	emailReq := ResendEmailRequest{
		From:    fromField,
		To:      recipients,
		Subject: subject,
		HTML:    bodyHTML,
	}

	jsonBytes, err := json.Marshal(emailReq)
	if err != nil {
		return fmt.Errorf("marshal email request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return fmt.Errorf("create http request to resend: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	nutshellClient := &http.Client{Timeout: httpTimeout}
	resp, err := nutshellClient.Do(req)
	if err != nil {
		return fmt.Errorf("send http request to resend: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("resend api returned error status: %d", resp.StatusCode)
	}

	return nil
}

func ValidateSMTPConfiguration() error {
	apiKey := strings.TrimSpace(os.Getenv("RESEND_API_KEY"))
	sender := strings.TrimSpace(os.Getenv("SMTP_SENDER"))

	missing := make([]string, 0, 2)
	if apiKey == "" {
		missing = append(missing, "RESEND_API_KEY")
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

// เก็บฟังก์ชันช่วยตกแต่ง Header เผื่อส่วนอื่นเรียกใช้ แต่ใน Resend จะส่งผ่าน JSON ให้อัตโนมัติ
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
		"To: undisclosed-recipients:;\r\n" +
		"Subject: " + encodedSubject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n" +
		"Content-Transfer-Encoding: base64\r\n\r\n" +
		wrappedBody.String() + "\r\n"
}

func mimeHeader(value string) string {
	return "=?UTF-8?B?" + base64.StdEncoding.EncodeToString([]byte(value)) + "?="
}
