// package utils

// import (
// 	"fmt"
// 	"net/smtp"
// 	"os"
// )

// // SendEmailNotification ฟังก์ชันกลางส่งอีเมล
// func SendEmailNotification(to string, subject string, bodyHTML string) error {
// 	host := os.Getenv("SMTP_HOST")
// 	port := os.Getenv("SMTP_PORT")
// 	sender := os.Getenv("SMTP_SENDER")
// 	password := os.Getenv("SMTP_PASSWORD")

// 	auth := smtp.PlainAuth("", sender, password, host)

// 	// จัด Header ข้อความให้รองรับ HTML
// 	header := make(map[string]string)
// 	header["From"] = fmt.Sprintf("ระบบแจ้งซ่อม IT <%s>", sender)
// 	header["To"] = to
// 	header["Subject"] = subject
// 	header["MIME-Version"] = "1.0"
// 	header["Content-Type"] = "text/html; charset=UTF-8"

// 	message := ""
// 	for k, v := range header {
// 		message += fmt.Sprintf("%s: %s\r\n", k, v)
// 	}
// 	message += "\r\n" + bodyHTML

// 	addr := fmt.Sprintf("%s:%s", host, port)

// 	// ส่งเมล
// 	err := smtp.SendMail(addr, auth, sender, []string{to}, []byte(message))
// 	if err != nil {
// 		return fmt.Errorf("failed to send email: %w", err)
// 	}

// 	return nil
// }

package utils

import (
	"fmt"
	"net/smtp"
	"os"
	"strings"
)

// SendEmailNotification เปลี่ยนการรับ to จาก string เป็น []string (รับกี่คนก็ได้)
const (
    AdminEmail = "chonpasu01@gmail.com"
    TechEmail  = "getaengja@gmail.com" // 👈 ใส่เมลส่วนตัวของช่างที่นี่
)
func SendEmailNotification(toRecipients []string, subject string, bodyHTML string) error {
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	sender := os.Getenv("SMTP_SENDER")
	password := os.Getenv("SMTP_PASSWORD")

	auth := smtp.PlainAuth("", sender, password, host)

	// แปลง List อีเมลผู้รับทั้งหมด ให้เป็นข้อความคั่นด้วยเครื่องหมายจุลภาค (,) เช่น "user@mail.com, admin@mail.com"
	toHeader := strings.Join(toRecipients, ", ")

	// จัด Header ข้อความ
	header := make(map[string]string)
	header["From"] = fmt.Sprintf("ระบบแจ้งซ่อม SC-SCI <%s>", sender)
	header["To"] = toHeader
	header["Subject"] = subject
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/html; charset=UTF-8"

	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + bodyHTML

	addr := fmt.Sprintf("%s:%s", host, port)

	// ส่งเมลไปยังผู้รับทุกคนใน toRecipients
	err := smtp.SendMail(addr, auth, sender, toRecipients, []byte(message))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}