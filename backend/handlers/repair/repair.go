package repair

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"html"
	"log"
	"math"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"

	"backend/auth"
	"backend/database"
	"backend/utils" // ใช้ utils.SendEmailNotification
)

// =========================================================
// Helper Functions (ฟังก์ชันผู้ช่วย)
// =========================================================

const (
	statusWaitingAdmin       = "รอแอดมินจัดสรร"
	statusWaitingTechnician  = "รอช่างรับงาน"
	statusReturnedToAdmin    = "ส่งคืนแอดมิน"
	statusWaitingForApproval = "รออนุมัติ"
	statusClosedEarly        = "ปิดงานก่อนครบขั้นตอน"
)

// nullIfEmpty สำหรับจัดการค่าว่างให้เป็น NULL ใน Database
func nullIfEmpty(value string) interface{} {
	if value == "" || value == "null" || value == "undefined" {
		return nil
	}
	return value
}

func parseNullableID(value string) (interface{}, error) {
	value = strings.TrimSpace(value)
	if value == "" || value == "null" || value == "undefined" {
		return nil, nil
	}
	id, err := strconv.Atoi(value)
	if err != nil || id <= 0 {
		return nil, fmt.Errorf("invalid ID")
	}
	return id, nil
}

func validateRepairImageFiles(files []*multipart.FileHeader) string {
	if len(files) > 10 {
		return "แนบรูปภาพได้สูงสุด 10 รูป"
	}

	allowedImageExtensions := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true,
		".webp": true, ".bmp": true, ".avif": true, ".heic": true, ".heif": true,
	}
	for _, file := range files {
		if file.Size > 10<<20 {
			return "รูปภาพแต่ละไฟล์ต้องมีขนาดไม่เกิน 10 MB"
		}
		contentType := strings.ToLower(file.Header.Get("Content-Type"))
		extension := strings.ToLower(filepath.Ext(file.Filename))
		if !strings.HasPrefix(contentType, "image/") || !allowedImageExtensions[extension] {
			return "กรุณาเลือกไฟล์รูปภาพที่รองรับ (JPG, PNG, GIF, WebP, BMP, AVIF หรือ HEIC)"
		}
	}
	return ""
}

// generateTicketNumber creates a collision-resistant ticket while preserving the existing display format.
func generateTicketNumber() (string, error) {
	now := time.Now()
	prefix := fmt.Sprintf("REQ-%s", now.Format("060102"))
	randomBytes := make([]byte, 4)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-%s", prefix, strings.ToUpper(hex.EncodeToString(randomBytes))), nil
}

func insertRepairWithUniqueTicket(values ...interface{}) (int, string, error) {
	for attempt := 0; attempt < 5; attempt++ {
		ticketNumber, err := generateTicketNumber()
		if err != nil {
			return 0, "", err
		}

		tx, err := database.DB.Begin()
		if err != nil {
			return 0, "", err
		}

		var repairID int
		args := append([]interface{}{ticketNumber}, values...)
		args = append(args, statusWaitingAdmin)
		err = tx.QueryRow(
			`INSERT INTO repairs (ticket_number, reporter_email, reporter_type, reporter_degree_level, reporter_department, department_id, location_id, other_location, floor_id, room_id, other_room, equipment_id, other_equipment, problem_type_id, other_problem_type, description, status)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17) RETURNING id`,
			args...,
		).Scan(&repairID)
		if err != nil {
			tx.Rollback()
			var pgErr *pq.Error
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				continue
			}
			return 0, "", err
		}
		if err := tx.Commit(); err != nil {
			return 0, "", err
		}
		return repairID, ticketNumber, nil
	}
	return 0, "", fmt.Errorf("unable to allocate a unique repair ticket")
}

// insertRepairLog บันทึกประวัติการทำรายการทุกฝีก้าวลงตาราง repair_logs
func insertRepairLog(repairID interface{}, userID interface{}, action, oldStatus, newStatus, note string) {
	database.DB.Exec(`
		INSERT INTO repair_logs (repair_id, user_id, action, old_status, new_status, note)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, repairID, userID, action, nullIfEmpty(oldStatus), nullIfEmpty(newStatus), nullIfEmpty(note))
}

func adminEmailRecipients() []string {
	rows, err := database.DB.Query(`
		SELECT email FROM users
		WHERE role = 'admin' AND is_active = TRUE
		  AND email IS NOT NULL AND BTRIM(email) <> ''
		ORDER BY id`)
	if err != nil {
		log.Printf("failed to load admin email recipients: %v", err)
		return nil
	}
	defer rows.Close()

	recipients := make([]string, 0)
	for rows.Next() {
		var email string
		if err := rows.Scan(&email); err != nil {
			log.Printf("failed to read admin email recipient: %v", err)
			continue
		}
		recipients = append(recipients, strings.TrimSpace(email))
	}
	if err := rows.Err(); err != nil {
		log.Printf("failed while reading admin email recipients: %v", err)
	}
	return recipients
}

func sendAdminEmail(subject, body string) error {
	recipients := adminEmailRecipients()
	if len(recipients) == 0 {
		return fmt.Errorf("no active admin email addresses found")
	}
	return utils.SendEmailNotification(recipients, subject, body)
}

type computerHeadTechnician struct {
	ID    int
	Name  string
	Email string
}

func assignComputerRepairToHead(repairID int, problemTypeID, departmentID, locationID, floorID, roomID interface{}) (*computerHeadTechnician, error) {

	var lead computerHeadTechnician
	err := database.DB.QueryRow(`

        SELECT u.id, u.full_name, COALESCE(u.email, '')
        FROM users u
        JOIN technician_specialties ts ON ts.user_id = u.id
        JOIN problem_types pt ON pt.id = ts.problem_type_id
        WHERE u.role = 'technician'
          AND u.is_active = TRUE
          AND u.is_head_technician = TRUE
          AND ts.problem_type_id = $1
          AND EXISTS (
              SELECT 1
              FROM technician_responsible_areas ra
              WHERE ra.technician_id = u.id
                AND (
                    ra.location_id = $3
                    OR ra.floor_id = $4
                    OR ra.room_id = $5
                )
          )
        ORDER BY CASE WHEN u.department_id IS NOT DISTINCT FROM $2 THEN 0 ELSE 1 END, u.id
        LIMIT 1
    `, problemTypeID, departmentID, locationID, floorID, roomID).Scan(&lead.ID, &lead.Name, &lead.Email)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find head technician: %w", err)
	}

	tx, err := database.DB.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to start repair assignment: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec("UPDATE repairs SET technician_id = $1, status = $2 WHERE id = $3", lead.ID, statusWaitingTechnician, repairID); err != nil {
		return nil, fmt.Errorf("failed to assign repair to head technician: %w", err)
	}

	if _, err := tx.Exec(`
        INSERT INTO repair_logs (repair_id, user_id, action, old_status, new_status, note)
        VALUES ($1, $2, 'ASSIGNED', $3, $4, $5)
    `, repairID, lead.ID, statusWaitingAdmin, statusWaitingTechnician, "ระบบส่งงานให้หัวหน้าช่างประจำหมวด: "+lead.Name); err != nil {
		return nil, fmt.Errorf("failed to log repair assignment: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to save repair assignment: %w", err)
	}

	return &lead, nil
}

// =========================================================
// ธีมสีกลาง สำหรับอีเมลทุกฉบับในระบบ
// =========================================================
const (
	colorSuccess = "#166534" // เขียวเข้มเพื่อให้อ่านชัดบนพื้นขาว
	colorInfo    = "#2563eb" // ฟ้า: แจ้งเตือนทั่วไป / ข้อมูล
	colorWarning = "#92400e" // ส้มเข้มเพื่อให้อ่านชัดบนพื้นขาว
	colorDanger  = "#dc2626" // แดง: ปัญหา / ยกเลิก / ปฏิเสธ
)

func frontendURL() string {
	if configuredURL := strings.TrimRight(strings.TrimSpace(os.Getenv("FRONTEND_URL")), "/"); configuredURL != "" {
		return configuredURL
	}
	return "http://localhost:5173"
}

func emailHTMLStyle() string {
	return `
		body { font-family: 'Sarabun', Arial, sans-serif; color: #1f2937; background-color: #f4f6f9; margin: 0; padding: 15px; line-height: 1.6; }
		.card { max-width: 480px; color: #1f2937; background-color: #ffffff; margin: 0 auto; border-radius: 10px; padding: 25px; box-shadow: 0 2px 8px rgba(0,0,0,0.08); }
		.box-success { background-color: #dcfce7; color: #14532d; border-left: 4px solid #166534; padding: 12px; margin: 15px 0; border-radius: 4px; }
		.box-info { background-color: #dbeafe; color: #1e40af; border-left: 4px solid #2563eb; padding: 12px; margin: 15px 0; border-radius: 4px; }
		.box-warning { background-color: #fef3c7; color: #92400e; border-left: 4px solid #d97706; padding: 12px; margin: 15px 0; border-radius: 4px; }
		.box-danger { background-color: #fee2e2; color: #991b1b; border-left: 4px solid #dc2626; padding: 12px; margin: 15px 0; border-radius: 4px; }
		.btn { display: inline-block; background-color: #10a7ff; color: #ffffff; text-decoration: none; padding: 10px 18px; border-radius: 6px; margin-top: 15px; font-weight: 600; }
		.btn-success { background-color: #166534; }
		.footer-note { font-size: 12px; color: #888888; margin-top: 20px; }
	`
}

func wrapEmail(headerColor string, headerText string, bodyContent string) string {
	return fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head><meta charset="UTF-8"><style>%s</style></head>
		<body><div class="card"><h2 style="color: %s; margin-top: 0;">%s</h2>%s</div></body>
		</html>
	`, emailHTMLStyle(), headerColor, headerText, bodyContent)
}

// =========================================================
// API Handlers หลัก
// =========================================================

// ---------------------------------------------------------
// 1. CreateRepair: สร้างใบแจ้งซ่อมใหม่ (อัปเกรดรองรับ other_room)
// ---------------------------------------------------------
func CreateRepair(c *gin.Context) {
	reporterEmail := c.PostForm("reporter_email")
	reporterType := strings.TrimSpace(c.PostForm("reporter_type"))
	reporterDepartment := strings.TrimSpace(c.PostForm("reporter_department"))
	reporterDegreeLevel := strings.TrimSpace(c.PostForm("reporter_degree_level"))
	departmentID, err := parseNullableID(c.PostForm("department_id")) // 🔥 รับสาขาของคนแจ้ง
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "รหัสภาควิชาสังกัดไม่ถูกต้อง"})
		return
	}
	locationID, err := parseNullableID(c.PostForm("location_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "รหัสอาคารหรือสถานที่ไม่ถูกต้อง"})
		return
	}
	otherLocation := nullIfEmpty(c.PostForm("other_location"))
	floorID, err := parseNullableID(c.PostForm("floor_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "รหัสชั้นไม่ถูกต้อง"})
		return
	}
	roomID, err := parseNullableID(c.PostForm("room_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "รหัสห้องไม่ถูกต้อง"})
		return
	}
	otherRoom := nullIfEmpty(c.PostForm("other_room")) // 🔥 [เพิ่มใหม่] รับชื่อห้องที่กรอกเองกรณีเลือก "อื่นๆ"
	equipmentID, err := parseNullableID(c.PostForm("equipment_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "รหัสอุปกรณ์ไม่ถูกต้อง"})
		return
	}
	otherEquipment := strings.TrimSpace(c.PostForm("other_equipment"))
	if len(otherEquipment) > 255 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ชื่ออุปกรณ์ยาวเกินไป"})
		return
	}
	problemTypeID, err := parseNullableID(c.PostForm("problem_type_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "รหัสหมวดหมู่งานไม่ถูกต้อง"})
		return
	}
	otherProblemType := nullIfEmpty(c.PostForm("other_problem_type"))
	description := c.PostForm("description")
	reporterEmail = strings.TrimSpace(reporterEmail)
	description = strings.TrimSpace(description)

	if reporterEmail == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "กรุณากรอกอีเมลผู้แจ้ง"})
		return
	}
	if len(reporterEmail) > 254 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "อีเมลผู้แจ้งยาวเกินไป"})
		return
	}
	if description == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "กรุณาอธิบายรายละเอียดของปัญหา"})
		return
	}

	switch reporterType {
	case "บุคลากร":
		reporterType = "เจ้าหน้าที่"
	case "บุคคลภายนอก / อื่น ๆ":
		reporterType = "บุคลากรภายนอก"
	}

	if reporterType != "นักศึกษา" && reporterType != "อาจารย์" &&
		reporterType != "เจ้าหน้าที่" && reporterType != "บุคลากรภายนอก" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "กรุณาระบุประเภทผู้แจ้งให้ถูกต้อง"})
		return
	}

	if reporterType == "บุคลากรภายนอก" {
		if reporterDepartment == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "กรุณาระบุหน่วยงานหรือความเกี่ยวข้อง"})
			return
		}
		departmentID = nil
		reporterDegreeLevel = ""
	} else {
		if departmentID == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "กรุณาเลือกภาควิชาสังกัดจากฐานข้อมูล"})
			return
		}

		var departmentName, departmentDegree string
		err := database.DB.QueryRow(`
			SELECT d.name, COALESCE(dl.name, '')
			FROM departments d
			LEFT JOIN degree_levels dl ON dl.id = d.degree_level_id
			WHERE d.id = $1
		`, departmentID).Scan(&departmentName, &departmentDegree)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ไม่พบภาควิชาที่เลือก"})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถตรวจสอบภาควิชาที่เลือกได้"})
			return
		}
		if reporterType == "นักศึกษา" && (reporterDegreeLevel == "" || reporterDegreeLevel != departmentDegree) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ระดับการศึกษาไม่ตรงกับภาควิชาที่เลือก"})
			return
		}
		reporterDepartment = departmentName
		if reporterType != "นักศึกษา" {
			reporterDegreeLevel = ""
		}
	}

	repairID, ticketNumber, err := insertRepairWithUniqueTicket(
		reporterEmail,
		reporterType,
		nullIfEmpty(reporterDegreeLevel),
		reporterDepartment,
		departmentID,
		locationID,
		otherLocation,
		floorID,
		roomID,
		otherRoom,
		equipmentID,
		nullIfEmpty(otherEquipment),
		problemTypeID,
		otherProblemType,
		description,
	)

	if err != nil {
		log.Printf("failed to insert repair: %v", err)
		var pgErr *pq.Error
		switch {
		case errors.As(err, &pgErr) && pgErr.Code == "23505":
			c.JSON(http.StatusConflict, gin.H{"error": "หมายเลข Ticket ซ้ำ กรุณากดส่งอีกครั้ง ระบบจะสร้างหมายเลขใหม่"})
		case errors.As(err, &pgErr) && pgErr.Code == "23503":
			c.JSON(http.StatusBadRequest, gin.H{"error": "พบข้อมูลอาคาร ห้อง อุปกรณ์ หรือหมวดหมู่ที่ไม่มีอยู่ในระบบ กรุณาเลือกใหม่"})
		case errors.As(err, &pgErr) && pgErr.Code == "23502":
			c.JSON(http.StatusBadRequest, gin.H{"error": "ข้อมูลที่จำเป็นยังไม่ครบ กรุณาตรวจสอบฟอร์มอีกครั้ง"})
		case errors.As(err, &pgErr) && pgErr.Code == "22P02":
			c.JSON(http.StatusBadRequest, gin.H{"error": "รหัสข้อมูลในฟอร์มไม่ถูกต้อง กรุณารีเฟรชหน้าและเลือกใหม่"})
		case errors.As(err, &pgErr) && pgErr.Code == "22001":
			c.JSON(http.StatusBadRequest, gin.H{"error": "ข้อมูลในฟอร์มยาวเกินที่ระบบรองรับ"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถบันทึกข้อมูลได้ กรุณาลองส่งรายการอีกครั้ง"})
		}
		return
	}
	// 🔥 บันทึก Log แรกเริ่ม
	insertRepairLog(repairID, nil, "CREATED", "", statusWaitingAdmin, "สร้างใบแจ้งซ่อมใหม่โดยผู้ใช้งาน")

	var computerLead *computerHeadTechnician
	if problemTypeID != nil {
		var routingErr error
		computerLead, routingErr = assignComputerRepairToHead(repairID, problemTypeID, departmentID, locationID, floorID, roomID)
		if routingErr != nil {
			log.Printf("failed to route computer repair %d to a head technician: %v", repairID, routingErr)
		}
	}
	reporterStatus := statusWaitingAdmin
	if computerLead != nil {
		reporterStatus = statusWaitingTechnician
	}

	form, _ := c.MultipartForm()
	if form != nil && form.File != nil {
		files := form.File["image"]
		for _, file := range files {
			filename := fmt.Sprintf("%d_%d%s", repairID, time.Now().UnixNano(), filepath.Ext(file.Filename))
			path := "uploads/" + filename
			if err := c.SaveUploadedFile(file, path); err == nil {
				database.DB.Exec(`INSERT INTO repair_images (repair_id, image_url, image_type) VALUES ($1, $2, 'before')`, repairID, "/uploads/"+filename)
			}
		}
	}

	// ---------- ส่งอีเมลแจ้งเตือน ----------
	go func() {
		formURL := frontendURL() + "/repair/history"

		// 1. อีเมลสำหรับผู้แจ้ง
		userBodyContent := fmt.Sprintf(`
			<p><strong>หมายเลขอ้างอิง:</strong> %s</p>
			<p><strong>รายละเอียด:</strong> %s</p>
			<div class="box-warning">ขณะนี้อยู่ในสถานะ: <strong>%s</strong></div>
			<p>ทางเราจะแจ้งเตือนอีกครั้งเมื่อช่างเริ่มเข้าดำเนินการครับ</p>
			<a href="%s" class="btn">ตรวจสอบรายการแจ้งซ่อมได้ที่นี่</a>
		`, html.EscapeString(ticketNumber), html.EscapeString(description), html.EscapeString(reporterStatus), formURL)
		userSubject := fmt.Sprintf("✅ รับเรื่องแจ้งซ่อมเรียบร้อยแล้ว (%s)", ticketNumber)
		userBody := wrapEmail(colorSuccess, "ระบบได้รับเรื่องแจ้งซ่อมของคุณเรียบร้อยแล้ว", userBodyContent)

		if reporterEmail != "" {
			if err := utils.SendEmailNotification([]string{reporterEmail}, userSubject, userBody); err != nil {
				log.Printf("failed to email repair receipt to reporter for repair %d: %v", repairID, err)
			}
		} else {
			log.Printf("cannot email repair receipt for repair %d: reporter has no email address", repairID)
		}

		if computerLead != nil {
			if computerLead.Email == "" {
				log.Printf("computer head technician %d has no email for repair %d", computerLead.ID, repairID)
			} else {
				leadBodyContent := fmt.Sprintf(`
					<p><strong>รหัสอ้างอิง:</strong> %s</p>
					<p><strong>รายละเอียดงาน:</strong> %s</p>
					<a href="%s/tech/home" class="btn">เปิดรายการงานช่าง</a>
				`, html.EscapeString(ticketNumber), html.EscapeString(description), frontendURL())
				if err := utils.SendEmailNotification(
					[]string{computerLead.Email},
					fmt.Sprintf("มีงานใหม่ส่งถึงหัวหน้าช่าง (%s)", ticketNumber),
					wrapEmail(colorInfo, "มีงานใหม่ส่งถึงคุณ", leadBodyContent),
				); err != nil {
					log.Printf("failed to email computer head technician for repair %d: %v", repairID, err)
				}
			}
		}

		// 2. อีเมลสำหรับ Admin
		adminLink := frontendURL() + "/admin/home"
		adminInstruction := "กรุณาเข้าสู่ระบบเพื่อพิจารณามอบหมายช่างดำเนินการ"
		if computerLead != nil {
			adminInstruction = "ระบบส่งงานให้หัวหน้าช่าง " + html.EscapeString(computerLead.Name) + " แล้ว"
		}
		adminBodyContent := fmt.Sprintf(`
			<p>%s</p>
			<p><strong>หมายเลขอ้างอิง:</strong> %s</p>
			<p><strong>ผู้แจ้ง:</strong> %s</p>
			<p><strong>รายละเอียด:</strong> %s</p>
			<div class="box-warning">ขณะนี้อยู่ในสถานะ: <strong>%s</strong></div>
			<a href="%s" class="btn">ตรวจสอบรายการได้ที่นี่</a>
		`, adminInstruction, html.EscapeString(ticketNumber), html.EscapeString(reporterEmail), html.EscapeString(description), html.EscapeString(reporterStatus), adminLink)
		adminSubject := fmt.Sprintf("🔔 มีงานแจ้งซ่อมใหม่เข้ามา (%s)", ticketNumber)
		adminBody := wrapEmail(colorInfo, "มีรายการแจ้งซ่อมใหม่เข้ามาในระบบ", adminBodyContent)

		if err := sendAdminEmail(adminSubject, adminBody); err != nil {
			log.Printf("failed to email new repair to admins: %v", err)
		}
	}()

	response := gin.H{"message": "สร้างรายการแจ้งซ่อมสำเร็จ", "ticket_number": ticketNumber, "status": reporterStatus}
	if computerLead != nil {
		response["assigned_technician"] = computerLead.Name
	}
	c.JSON(http.StatusCreated, response)
}

// ---------------------------------------------------------
// 2. GetAllRepairs: ดึงรายการซ่อมทั้งหมด (อัปเกรดดึง other_room)
// ---------------------------------------------------------
func GetAllRepairs(c *gin.Context) {
	query := `
		SELECT r.id, r.ticket_number, r.description, r.status, r.created_at, r.technician_id, r.problem_type_id, r.technician_note, r.admin_note, r.reporter_email,
			   r.reporter_type, r.reporter_degree_level, r.reporter_department,
			   r.location_id, r.floor_id, r.room_id,
			   r.estimated_cost, r.actual_cost, r.accepted_at, r.completed_at, r.other_location, r.other_room, r.other_problem_type,
			   r.other_equipment, r.return_note, r.closure_note,
			   r.department_id, d.name as department_name,
			   l.name as location_name, f.floor_name, ro.room_number,
			   eq.name as equipment_name, eq.asset_code, eq.base_price, p.name as problem_type_name, u.full_name as technician_name
		FROM repairs r
		LEFT JOIN departments d ON r.department_id = d.id
		LEFT JOIN locations l ON r.location_id = l.id
		LEFT JOIN floors f ON r.floor_id = f.id
		LEFT JOIN rooms ro ON r.room_id = ro.id
		LEFT JOIN equipments eq ON r.equipment_id = eq.id
		LEFT JOIN problem_types p ON r.problem_type_id = p.id
		LEFT JOIN users u ON r.technician_id = u.id
		WHERE 1=1
	`

	args := make([]interface{}, 0, 1)
	if auth.Role(c) == "technician" {
		args = append(args, auth.UserID(c))
		query += ` AND (
			r.technician_id = $1
			OR EXISTS (
				SELECT 1
				FROM users head
				JOIN users member ON member.id = r.technician_id
				WHERE head.id = $1
				  AND head.is_head_technician = TRUE
				  AND head.department_id IS NOT NULL
				  AND member.department_id = head.department_id
			)
		)`
	}

	query += ` ORDER BY r.created_at DESC`

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	repairs := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id int
		var desc, status, reporterEmail string
		var reporterType, reporterDegreeLevel, reporterDepartment *string
		var ticketNumber, locName, otherLoc, otherRoom, floorName, roomNumber, eqName, assetCode, probName, otherProb, otherEquipment, returnNote, closureNote, techName, techNote, adminNote, deptName *string
		var basePrice *float64
		var estimatedCost, actualCost float64
		var createdAt time.Time
		var acceptedAt, completedAt *time.Time
		var techID, problemTypeID, dID, locationID, floorID, roomID *int

		err := rows.Scan(&id, &ticketNumber, &desc, &status, &createdAt, &techID, &problemTypeID, &techNote, &adminNote, &reporterEmail,
			&reporterType, &reporterDegreeLevel, &reporterDepartment,
			&locationID, &floorID, &roomID,
			&estimatedCost, &actualCost, &acceptedAt, &completedAt, &otherLoc, &otherRoom, &otherProb,
			&otherEquipment, &returnNote, &closureNote,
			&dID, &deptName, &locName, &floorName, &roomNumber, &eqName, &assetCode, &basePrice, &probName, &techName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถอ่านรายการแจ้งซ่อมได้"})
			return
		}

		imgRows, err := database.DB.Query("SELECT image_url, image_type FROM repair_images WHERE repair_id = $1", id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถอ่านรูปภาพของรายการแจ้งซ่อมได้"})
			return
		}
		images := make([]map[string]string, 0)
		for imgRows.Next() {
			var url, imgType string
			if err := imgRows.Scan(&url, &imgType); err != nil {
				imgRows.Close()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถอ่านรูปภาพของรายการแจ้งซ่อมได้"})
				return
			}
			images = append(images, map[string]string{"url": url, "type": imgType})
		}
		if err := imgRows.Err(); err != nil {
			imgRows.Close()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถอ่านรูปภาพของรายการแจ้งซ่อมได้"})
			return
		}
		imgRows.Close()

		getString := func(s *string) string {
			if s == nil {
				return ""
			}
			return *s
		}
		equipmentName := getString(eqName)
		if customName := getString(otherEquipment); customName != "" {
			equipmentName = customName
		}

		repairs = append(repairs, map[string]interface{}{
			"id":                    id,
			"ticket_number":         getString(ticketNumber),
			"description":           desc,
			"status":                status,
			"created_at":            createdAt,
			"department_id":         dID,
			"problem_type_id":       problemTypeID,
			"department_name":       getString(deptName),
			"location_name":         getString(locName),
			"other_location":        getString(otherLoc),
			"floor_name":            getString(floorName),
			"room_number":           getString(roomNumber),
			"other_room":            getString(otherRoom), // 🔥 [เพิ่มใหม่] คืนค่าชื่อห้องที่พิมพ์เอง
			"equipment_name":        equipmentName,
			"other_equipment":       getString(otherEquipment),
			"asset_code":            getString(assetCode),
			"base_price":            basePrice,
			"problem_type":          getString(probName),
			"other_problem_type":    getString(otherProb),
			"reporter_email":        reporterEmail,
			"reporter_type":         getString(reporterType),
			"reporter_degree_level": getString(reporterDegreeLevel),
			"reporter_department":   getString(reporterDepartment),
			"location_id":           locationID,
			"floor_id":              floorID,
			"room_id":               roomID,
			"technician_id":         techID,
			"technician_note":       getString(techNote),
			"admin_note":            getString(adminNote),
			"return_note":           getString(returnNote),
			"closure_note":          getString(closureNote),
			"technician_name":       getString(techName),
			"estimated_cost":        estimatedCost,
			"actual_cost":           actualCost,
			"accepted_at":           acceptedAt,
			"completed_at":          completedAt,
			"images":                images,
		})
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถอ่านรายการแจ้งซ่อมได้"})
		return
	}
	c.JSON(http.StatusOK, repairs)
}

func requireRepairAccess(c *gin.Context, repairID string, allowTeamHead bool) bool {
	if auth.Role(c) == "admin" {
		return true
	}
	if auth.Role(c) != "technician" {
		c.JSON(http.StatusForbidden, gin.H{"error": "คุณไม่มีสิทธิ์ดูรายการแจ้งซ่อมนี้"})
		return false
	}

	teamCondition := ""
	if allowTeamHead {
		teamCondition = `
			OR EXISTS (
				SELECT 1
				FROM users head
				JOIN users member ON member.id = r.technician_id
				WHERE head.id = $2
				  AND head.is_head_technician = TRUE
				  AND head.department_id IS NOT NULL
				  AND member.department_id = head.department_id
			)
		`
	}
	query := `
		SELECT EXISTS (
			SELECT 1 FROM repairs r
			WHERE r.id = $1 AND (r.technician_id = $2
	`
	query += teamCondition + `))`

	var allowed bool
	if err := database.DB.QueryRow(query, repairID, auth.UserID(c)).Scan(&allowed); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถตรวจสอบสิทธิ์เข้าถึงรายการแจ้งซ่อมได้"})
		return false
	}
	if !allowed {
		c.JSON(http.StatusForbidden, gin.H{"error": "คุณไม่มีสิทธิ์ดูรายการแจ้งซ่อมนี้"})
		return false
	}
	return true
}

// ---------------------------------------------------------
// 3. GetRepairByID: ดึงรายละเอียดใบแจ้งซ่อมเฉพาะใบ (อัปเกรดดึง other_room)
// ---------------------------------------------------------
func GetRepairByID(c *gin.Context) {
	repairID := c.Param("id")
	if !requireRepairAccess(c, repairID, true) {
		return
	}

	var id int
	var desc, status, reporterEmail string
	var reporterType, reporterDegreeLevel, reporterDepartment *string
	var ticketNumber, locName, otherLoc, otherRoom, floorName, roomNumber, eqName, assetCode, probName, otherProb, otherEquipment, returnNote, closureNote, techName, techNote, adminNote, deptName *string
	var estimatedCost, actualCost float64
	var createdAt time.Time
	var acceptedAt, completedAt *time.Time
	var dID *int

	err := database.DB.QueryRow(`
		SELECT r.id, r.ticket_number, r.description, r.status, r.reporter_email,
			   r.reporter_type, r.reporter_degree_level, r.reporter_department,
			   r.technician_note, r.admin_note, r.created_at,
			   r.estimated_cost, r.actual_cost, r.accepted_at, r.completed_at, r.other_location, r.other_room, r.other_problem_type,
			   r.other_equipment, r.return_note, r.closure_note,
			   r.department_id, d.name as department_name,
			   l.name, f.floor_name, ro.room_number, eq.name, eq.asset_code, p.name, u.full_name
		FROM repairs r
		LEFT JOIN departments d ON r.department_id = d.id
		LEFT JOIN locations l ON r.location_id = l.id
		LEFT JOIN floors f ON r.floor_id = f.id
		LEFT JOIN rooms ro ON r.room_id = ro.id
		LEFT JOIN equipments eq ON r.equipment_id = eq.id
		LEFT JOIN problem_types p ON r.problem_type_id = p.id
		LEFT JOIN users u ON r.technician_id = u.id
		WHERE r.id = $1`, repairID).Scan(
		&id, &ticketNumber, &desc, &status, &reporterEmail,
		&reporterType, &reporterDegreeLevel, &reporterDepartment,
		&techNote, &adminNote, &createdAt,
		&estimatedCost, &actualCost, &acceptedAt, &completedAt, &otherLoc, &otherRoom, &otherProb,
		&otherEquipment, &returnNote, &closureNote,
		&dID, &deptName, &locName, &floorName, &roomNumber, &eqName, &assetCode, &probName, &techName)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบรายการแจ้งซ่อม"})
		return
	}

	rows, err := database.DB.Query("SELECT image_url, image_type FROM repair_images WHERE repair_id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถอ่านรูปภาพของรายการแจ้งซ่อมได้"})
		return
	}
	defer rows.Close()

	images := make([]map[string]string, 0)
	for rows.Next() {
		var url, imgType string
		if err := rows.Scan(&url, &imgType); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถอ่านรูปภาพของรายการแจ้งซ่อมได้"})
			return
		}
		images = append(images, map[string]string{"url": url, "type": imgType})
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถอ่านรูปภาพของรายการแจ้งซ่อมได้"})
		return
	}

	getString := func(s *string) string {
		if s == nil {
			return ""
		}
		return *s
	}

	equipmentName := getString(eqName)
	if customName := getString(otherEquipment); customName != "" {
		equipmentName = customName
	}

	repair := map[string]interface{}{
		"id":                    id,
		"ticket_number":         getString(ticketNumber),
		"description":           desc,
		"status":                status,
		"reporter_email":        reporterEmail,
		"reporter_type":         getString(reporterType),
		"reporter_degree_level": getString(reporterDegreeLevel),
		"reporter_department":   getString(reporterDepartment),
		"department_id":         dID,
		"department_name":       getString(deptName),
		"technician_note":       getString(techNote),
		"admin_note":            getString(adminNote),
		"return_note":           getString(returnNote),
		"closure_note":          getString(closureNote),
		"technician_name":       getString(techName),
		"location_name":         getString(locName),
		"other_location":        getString(otherLoc),
		"floor_name":            getString(floorName),
		"room_number":           getString(roomNumber),
		"other_room":            getString(otherRoom), // 🔥 [เพิ่มใหม่] คืนค่าชื่อห้องที่พิมพ์เอง
		"equipment_name":        equipmentName,
		"other_equipment":       getString(otherEquipment),
		"asset_code":            getString(assetCode),
		"problem_type":          getString(probName),
		"other_problem_type":    getString(otherProb),
		"estimated_cost":        estimatedCost,
		"actual_cost":           actualCost,
		"accepted_at":           acceptedAt,
		"completed_at":          completedAt,
		"created_at":            createdAt,
		"images":                images,
	}
	c.JSON(http.StatusOK, repair)
}

// ---------------------------------------------------------
// 4. AssignRepair: Admin มอบหมายช่าง (🔥 อัปเกรดใส่ Log)
// ---------------------------------------------------------
func AssignRepair(c *gin.Context) {
	repairID := c.Param("id")
	type AssignRequest struct {
		TechnicianID     int `json:"technician_id" binding:"required"`
		HeadTechnicianID int `json:"head_technician_id"`
	}

	var req AssignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ข้อมูลไม่ถูกต้อง"})
		return
	}
	if auth.Role(c) == "technician" {
		req.HeadTechnicianID = auth.UserID(c)
	} else {
		req.HeadTechnicianID = 0
	}

	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถเริ่มบันทึกการมอบหมายงานได้"})
		return
	}
	defer tx.Rollback()

	var currentStatus string
	var currentTechnicianID, repairProblemTypeID, locationID, floorID, roomID *int
	var otherLocation string
	err = tx.QueryRow(`
		SELECT status, technician_id, problem_type_id, location_id, floor_id, room_id, COALESCE(other_location, '')
		FROM repairs WHERE id = $1 FOR UPDATE
	`, repairID).Scan(&currentStatus, &currentTechnicianID, &repairProblemTypeID, &locationID, &floorID, &roomID, &otherLocation)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบรายการแจ้งซ่อม"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if auth.Role(c) == "technician" && currentStatus != statusWaitingTechnician {
		c.JSON(http.StatusConflict, gin.H{"error": "หัวหน้าช่างมอบหมายต่อได้เฉพาะงานที่รอช่างรับงาน"})
		return
	}
	if auth.Role(c) == "admin" &&
		currentStatus != statusWaitingAdmin &&
		currentStatus != statusReturnedToAdmin {
		c.JSON(http.StatusConflict, gin.H{"error": "มอบหมายได้เฉพาะงานที่รอจัดสรรหรือถูกส่งคืนจากช่าง"})
		return
	}

	actorID := auth.UserID(c)
	assignmentNote := fmt.Sprintf("Admin มอบหมายงานให้ช่าง ID: %d", req.TechnicianID)
	if repairProblemTypeID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ไม่สามารถตรวจสอบความถนัดของช่างสำหรับงานนี้ได้"})
		return
	}
	var eligible bool
	err = tx.QueryRow(`
		SELECT EXISTS (
			SELECT 1
			FROM users u
			JOIN technician_specialties ts ON ts.user_id = u.id
			WHERE u.id = $1
			  AND u.role = 'technician'
			  AND u.is_active = TRUE
			  AND ($3::boolean = FALSE OR (
				u.is_head_technician = FALSE
				AND u.department_id = (
					SELECT department_id FROM users
					WHERE id = $4 AND role = 'technician' AND is_head_technician = TRUE
				)
			  ))
			  AND ts.problem_type_id = $2
		)
	`, req.TechnicianID, *repairProblemTypeID, req.HeadTechnicianID > 0, req.HeadTechnicianID).Scan(&eligible)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !eligible {
		c.JSON(http.StatusBadRequest, gin.H{"error": "เลือกได้เฉพาะช่างที่ใช้งานอยู่และมีความถนัดตรงกับประเภทงานนี้"})
		return
	}

	if req.HeadTechnicianID > 0 {
		var isHeadTechnician bool
		err = tx.QueryRow(`
			SELECT is_head_technician
			FROM users
			WHERE id = $1 AND role = 'technician' AND is_active = TRUE
		`, req.HeadTechnicianID).Scan(&isHeadTechnician)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusForbidden, gin.H{"error": "เฉพาะหัวหน้าช่างที่ใช้งานอยู่เท่านั้นที่มอบหมายงานได้"})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if !isHeadTechnician {
			c.JSON(http.StatusForbidden, gin.H{"error": "เฉพาะหัวหน้าช่างที่ใช้งานอยู่เท่านั้นที่มอบหมายงานได้"})
			return
		}
		if currentStatus != statusWaitingTechnician || currentTechnicianID == nil || *currentTechnicianID != req.HeadTechnicianID {
			c.JSON(http.StatusForbidden, gin.H{"error": "หัวหน้าช่างมอบหมายได้เฉพาะงานที่ได้รับมอบหมายและยังไม่เริ่มดำเนินการ"})
			return
		}

		assignmentNote = fmt.Sprintf("หัวหน้าช่าง ID: %d มอบหมายงานให้ช่าง ID: %d", req.HeadTechnicianID, req.TechnicianID)
	}

	areaMatch := strings.TrimSpace(otherLocation) != ""
	if !areaMatch {
		err = tx.QueryRow(`
		SELECT EXISTS (
			SELECT 1
			FROM technician_responsible_areas ra
			WHERE ra.technician_id = $1
			  AND (
				(ra.scope = 'location' AND ra.location_id = $2)
				OR (ra.scope = 'floor' AND ra.floor_id = $3)
				OR (ra.scope = 'room' AND ra.room_id = $4)
			  )
		)
		`, req.TechnicianID, locationID, floorID, roomID).Scan(&areaMatch)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถตรวจสอบพื้นที่รับผิดชอบของช่างได้"})
			return
		}
	}
	if !areaMatch {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ช่างที่เลือกไม่ได้รับผิดชอบพื้นที่ของงานนี้"})
		return
	}

	_, err = tx.Exec(
		`UPDATE repairs SET status = $1, technician_id = $2, admin_note = NULL, return_note = NULL, closure_note = NULL, accepted_at = NULL, completed_at = NULL WHERE id = $3`,
		statusWaitingTechnician, req.TechnicianID, repairID,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	_, err = tx.Exec(`
		INSERT INTO repair_logs (repair_id, user_id, action, old_status, new_status, note)
		VALUES ($1, $2, 'ASSIGNED', $3, $4, $5)`,
		repairID, actorID, currentStatus, statusWaitingTechnician, assignmentNote)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถบันทึกประวัติการมอบหมายงานได้"})
		return
	}
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถบันทึกการมอบหมายงานได้"})
		return
	}

	// ---------- ส่งอีเมลแจ้งเตือน ----------
	go func() {
		var reporterEmail, description, tNumber, techName, techEmail string
		err := database.DB.QueryRow(`
			SELECT r.reporter_email, r.description, r.ticket_number,
			       COALESCE(u.full_name, ''), COALESCE(u.email, '')
			FROM repairs r LEFT JOIN users u ON u.id = $1 WHERE r.id = $2`, req.TechnicianID, repairID).
			Scan(&reporterEmail, &description, &tNumber, &techName, &techEmail)
		if err != nil {
			log.Printf("failed to load assignment email details for repair %s: %v", repairID, err)
			return
		}

		// แจ้งผู้แจ้ง
		if reporterEmail != "" {
			userBodyContent := fmt.Sprintf(`
				<p><strong>รายการ:</strong> %s</p>
				<div class="box-info">ช่างผู้รับผิดชอบ: <strong>%s</strong></div>
				<p>ช่างจะเข้าประเมินและดำเนินการต่อไปครับ</p>
			`, html.EscapeString(description), html.EscapeString(techName))
			if err := utils.SendEmailNotification([]string{reporterEmail}, fmt.Sprintf("⚙️ มอบหมายช่างซ่อมแล้ว (%s)", tNumber), wrapEmail(colorInfo, "มอบหมายช่างเข้าดูแลงานแล้ว", userBodyContent)); err != nil {
				log.Printf("failed to email repair assignment to reporter for repair %s: %v", repairID, err)
			}
		}

		// แจ้งช่าง
		if techEmail != "" {
			techBodyContent := fmt.Sprintf(`
				<p><strong>รหัสอ้างอิง:</strong> %s</p>
				<p><strong>รายละเอียดงาน:</strong> %s</p>
				<a href="%s/tech/home" class="btn">เปิดรายการงานช่าง</a>
			`, html.EscapeString(tNumber), html.EscapeString(description), frontendURL())
			if err := utils.SendEmailNotification([]string{techEmail}, fmt.Sprintf("🔧 คุณได้รับมอบหมายงานซ่อมใหม่ (%s)", tNumber), wrapEmail(colorInfo, "คุณได้รับมอบหมายงานซ่อมใหม่", techBodyContent)); err != nil {
				log.Printf("failed to email repair assignment to technician for repair %s: %v", repairID, err)
			}
		} else {
			log.Printf("technician %d has no email for repair assignment %s", req.TechnicianID, repairID)
		}
	}()

	c.JSON(http.StatusOK, gin.H{"message": "มอบหมายงานให้ช่างสำเร็จ", "status": statusWaitingTechnician})
}

// ---------------------------------------------------------
// 5. RevokeRepair: Admin ดึงงานกลับ (🔥 อัปเกรดใส่ Log)
// ---------------------------------------------------------
func RevokeRepair(c *gin.Context) {
	repairID := c.Param("id")
	revokeReason := strings.TrimSpace(c.PostForm("rejection_reason"))
	adminID := auth.UserID(c)

	if revokeReason == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "กรุณาระบุเหตุผลในการดึงงานกลับ"})
		return
	}

	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถเริ่มบันทึกการดึงงานกลับได้"})
		return
	}
	defer tx.Rollback()

	var techName, techEmail, description, tNumber, currentStatus string
	var technicianID int
	err = tx.QueryRow(`
		SELECT u.id, u.full_name, COALESCE(u.email, ''), r.description, COALESCE(r.ticket_number, ''), r.status
		FROM repairs r JOIN users u ON r.technician_id = u.id
		WHERE r.id = $1 FOR UPDATE OF r`, repairID).
		Scan(&technicianID, &techName, &techEmail, &description, &tNumber, &currentStatus)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบงานที่มอบหมายให้ช่าง"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	_, err = tx.Exec(`UPDATE repairs SET status = $1, technician_id = NULL, admin_note = NULL, return_note = NULL, accepted_at = NULL, completed_at = NULL WHERE id = $2`, statusWaitingAdmin, repairID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	_, err = tx.Exec(`
		INSERT INTO repair_logs (repair_id, user_id, action, old_status, new_status, note)
		VALUES ($1, $2, 'REVOKED', $3, $4, $5)`,
		repairID, adminID, currentStatus, statusWaitingAdmin, "Admin ดึงงานกลับจากช่าง. เหตุผล: "+revokeReason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถบันทึกประวัติการดึงงานกลับได้"})
		return
	}
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถบันทึกการดึงงานกลับได้"})
		return
	}

	go func() {
		if techEmail != "" {
			bodyContent := fmt.Sprintf(`
				<p>เรียนคุณ %s</p>
				<p>ระบบขอแจ้งให้ทราบว่ารายการแจ้งซ่อมที่คุณรับผิดชอบ ได้ถูกดึงงานกลับ/ยกเลิกการมอบหมายแล้วครับ</p>
				<p><strong>รหัสตั๋ว:</strong> %s</p>
				<p><strong>รายละเอียดงาน:</strong> %s</p>
				<div class="box-warning"><strong>📌 เหตุผล:</strong><br>%s</div>
				<a href="%s/tech/home" class="btn">เปิดหน้ารายการงานช่าง</a>
			`, html.EscapeString(techName), html.EscapeString(tNumber), html.EscapeString(description), html.EscapeString(revokeReason), frontendURL())
			if err := utils.SendEmailNotification([]string{techEmail}, fmt.Sprintf("⚠️ แจ้งเตือนการดึงงานซ่อมกลับ (%s)", tNumber), wrapEmail(colorDanger, "แจ้งเตือนการดึงงานซ่อมกลับ", bodyContent)); err != nil {
				log.Printf("failed to email revoked repair assignment to technician for repair %s: %v", repairID, err)
			}
		} else {
			log.Printf("cannot email revoked repair assignment for repair %s: technician %d has no email address", repairID, technicianID)
		}
	}()

	c.JSON(http.StatusOK, gin.H{"message": "ยกเลิกการจ่ายงานสำเร็จ"})
}

func technicianReturnReason(category, details string) (string, bool) {
	switch strings.TrimSpace(category) {
	case "ไม่ถนัด", "ส่งงานผิดประเภท", "ทำไม่ได้", "งานซ้ำ":
		if details = strings.TrimSpace(details); details != "" {
			return strings.TrimSpace(category) + ": " + details, true
		}
		return strings.TrimSpace(category), true
	default:
		return "", false
	}
}

func repairReturnAction(currentStatus string, isAssignedTechnician bool) (string, bool) {
	switch currentStatus {
	case statusWaitingTechnician:
		if !isAssignedTechnician {
			return "RETURNED_TO_ADMIN", true
		}
		return "REJECTED_BEFORE_ACCEPT", true
	case "รับงานแล้ว", "กำลังซ่อม":
		return "RETURNED_TO_ADMIN", true
	default:
		return "", false
	}
}

// RejectRepair returns an unaccepted or unworkable job to the administrator with a reason.
func RejectRepair(c *gin.Context) {
	repairID := c.Param("id")
	technicianID := auth.UserID(c)
	if technicianID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ไม่พบข้อมูลช่างผู้ส่งคืนงาน"})
		return
	}
	rejectionReason, validReason := technicianReturnReason(c.PostForm("rejection_reason"), c.PostForm("rejection_note"))
	if !validReason {
		c.JSON(http.StatusBadRequest, gin.H{"error": "กรุณาเลือกเหตุผลในการคืนงาน"})
		return
	}

	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถเริ่มบันทึกการคืนงานได้"})
		return
	}
	defer tx.Rollback()

	var currentStatus string
	var assignedTechnicianID sql.NullInt64
	err = tx.QueryRow(`
		SELECT status, technician_id
		FROM repairs WHERE id = $1 FOR UPDATE
	`, repairID).Scan(&currentStatus, &assignedTechnicianID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบรายการแจ้งซ่อม"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถตรวจสอบงานที่ต้องการคืนได้"})
		return
	}
	if !assignedTechnicianID.Valid {
		c.JSON(http.StatusConflict, gin.H{"error": "งานนี้ยังไม่มีช่างรับผิดชอบ กรุณารีเฟรชรายการ"})
		return
	}
	action, canReturn := repairReturnAction(currentStatus, assignedTechnicianID.Int64 == int64(technicianID))
	if !canReturn {
		c.JSON(http.StatusConflict, gin.H{"error": "ไม่สามารถส่งคืนงานในสถานะปัจจุบันได้"})
		return
	}

	var authorized bool
	if assignedTechnicianID.Int64 == int64(technicianID) {
		err = tx.QueryRow(`
			SELECT EXISTS (
				SELECT 1 FROM users
				WHERE id = $1 AND role = 'technician' AND is_active = TRUE
			)
		`, technicianID).Scan(&authorized)
	} else {
		err = tx.QueryRow(`
			SELECT EXISTS (
				SELECT 1
				FROM users owner
				JOIN users head ON head.id = $2
				WHERE owner.id = $1
				  AND owner.role = 'technician'
				  AND head.role = 'technician'
				  AND head.is_active = TRUE
				  AND head.is_head_technician = TRUE
				  AND owner.department_id IS NOT NULL
				  AND owner.department_id = head.department_id
			)
		`, assignedTechnicianID.Int64, technicianID).Scan(&authorized)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถตรวจสอบสิทธิ์ในการคืนงานได้"})
		return
	}
	if !authorized {
		c.JSON(http.StatusForbidden, gin.H{"error": "เฉพาะผู้รับผิดชอบงานหรือหัวหน้าช่างในทีมเดียวกันเท่านั้นที่ส่งคืนงานได้"})
		return
	}
	var assignedTechnicianName, assignedTechnicianEmail string
	err = tx.QueryRow(`
		SELECT full_name, COALESCE(email, '')
		FROM users WHERE id = $1
	`, assignedTechnicianID.Int64).Scan(&assignedTechnicianName, &assignedTechnicianEmail)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถอ่านข้อมูลช่างผู้รับผิดชอบได้"})
		return
	}

	result, err := tx.Exec(`
		UPDATE repairs
		SET status = $1, technician_id = NULL, accepted_at = NULL, completed_at = NULL, return_note = $5
		WHERE id = $2 AND technician_id = $3 AND status = $4
	`, statusReturnedToAdmin, repairID, assignedTechnicianID.Int64, currentStatus, rejectionReason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถส่งงานคืนให้แอดมินได้"})
		return
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "งานเปลี่ยนสถานะแล้ว กรุณารีเฟรชรายการ"})
		return
	}

	actionNote := "ช่างส่งคืนงานให้แอดมิน"
	if action == "REJECTED_BEFORE_ACCEPT" {
		actionNote = "ช่างปฏิเสธรับงานก่อนเริ่มดำเนินการ"
	}
	if _, err := tx.Exec(`
		INSERT INTO repair_logs (repair_id, user_id, action, old_status, new_status, note)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, repairID, technicianID, action, currentStatus, statusReturnedToAdmin, actionNote+": "+rejectionReason); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถบันทึกประวัติการคืนงานได้"})
		return
	}
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถบันทึกการคืนงานได้"})
		return
	}

	go func() {
		var reporterEmail, description, ticketNumber string
		if err := database.DB.QueryRow(`SELECT reporter_email, description, ticket_number FROM repairs WHERE id = $1`, repairID).
			Scan(&reporterEmail, &description, &ticketNumber); err != nil {
			log.Printf("failed to load returned repair notification details for %s: %v", repairID, err)
			return
		}

		bodyContent := fmt.Sprintf(`
			<p><strong>รหัสตั๋ว:</strong> %s</p>
			<p><strong>รายการ:</strong> %s</p>
			<p><strong>ผู้รับผิดชอบเดิม:</strong> %s</p>
			<div class="box-warning"><strong>เหตุผลที่ช่างส่งงานคืน:</strong><br>%s</div>
		`, html.EscapeString(ticketNumber), html.EscapeString(description), html.EscapeString(assignedTechnicianName), html.EscapeString(rejectionReason))
		recipients := adminEmailRecipients()
		if reporterEmail != "" {
			recipients = append(recipients, reporterEmail)
		}
		if assignedTechnicianEmail != "" {
			recipients = append(recipients, assignedTechnicianEmail)
		}
		if err := utils.SendEmailNotification(recipients, fmt.Sprintf("ช่างส่งงานคืนให้แอดมิน (%s)", ticketNumber), wrapEmail(colorWarning, "งานซ่อมถูกส่งคืนให้ผู้ดูแลพิจารณา", bodyContent)); err != nil {
			log.Printf("failed to email returned repair notification for %s: %v", repairID, err)
		}
	}()

	message := "ส่งคืนงานให้แอดมินแล้ว"
	if action == "REJECTED_BEFORE_ACCEPT" {
		message = "ปฏิเสธรับงานและส่งกลับให้แอดมินแล้ว"
	}
	c.JSON(http.StatusOK, gin.H{"message": message, "status": statusReturnedToAdmin})
}

// ---------------------------------------------------------
// 7. UpdateRepairStatus: อัปเดตสถานะ (🔥 ล็อกสเต็ป & บังคับ Note)
// ---------------------------------------------------------
func UpdateRepairStatus(c *gin.Context) {
	repairID := c.Param("id")
	status := c.PostForm("status")
	actionByUserID := auth.UserID(c)
	actualCostStr := c.PostForm("actual_cost")
	tNote := nullIfEmpty(c.PostForm("technician_note"))
	techNoteStr := c.PostForm("technician_note")

	if !requireRepairAccess(c, repairID, false) {
		return
	}

	if status == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "กรุณาระบุสถานะ"})
		return
	}
	if status != "รับงานแล้ว" && status != "เสร็จเรียบร้อย" && status != "ซ่อมไม่ได้" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ไม่อนุญาตให้เปลี่ยนเป็นสถานะนี้ผ่านรายการช่าง"})
		return
	}

	var currentStatus string
	if err := database.DB.QueryRow("SELECT status FROM repairs WHERE id = $1", repairID).Scan(&currentStatus); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบรายการแจ้งซ่อม"})
		return
	}
	if (status == "รับงานแล้ว" && currentStatus != statusWaitingTechnician) ||
		((status == "เสร็จเรียบร้อย" || status == "ซ่อมไม่ได้") && currentStatus != "กำลังซ่อม") {
		c.JSON(http.StatusConflict, gin.H{"error": "ไม่สามารถเปลี่ยนสถานะจากสถานะปัจจุบันได้"})
		return
	}

	var completionFiles []*multipart.FileHeader
	if status == "เสร็จเรียบร้อย" || status == "ซ่อมไม่ได้" {
		form, err := c.MultipartForm()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ไม่สามารถอ่านไฟล์รูปภาพได้"})
			return
		}
		completionFiles = form.File["image"]
		if validationError := validateRepairImageFiles(completionFiles); validationError != "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationError})
			return
		}
	}

	// 🔥 Guardrail: บังคับสรุปงานก่อนปิดจ๊อบเพื่อลด Human Error
	if status == "เสร็จเรียบร้อย" && techNoteStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ระบบป้องกัน: กรุณากรอก 'บันทึกจากช่าง' เพื่อสรุปงานก่อนทำการปิดงานซ่อม"})
		return
	}

	shouldNotifyCompletion := status == "เสร็จเรียบร้อย" && currentStatus != "เสร็จเรียบร้อย"
	shouldUpdateActualCost := status == "เสร็จเรียบร้อย" || status == "ซ่อมไม่ได้"
	shouldSetAcceptedAt := status == "รับงานแล้ว"

	var actualCost float64
	if actualCostStr != "" && actualCostStr != "null" {
		parsedCost, err := strconv.ParseFloat(actualCostStr, 64)
		if err != nil || math.IsNaN(parsedCost) || math.IsInf(parsedCost, 0) || parsedCost < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "กรุณาระบุค่าใช้จ่ายจริงให้ถูกต้อง"})
			return
		}
		actualCost = parsedCost
	}

	query := `UPDATE repairs
		SET status = $1,
		    technician_note = COALESCE($2, technician_note),
		    actual_cost = CASE WHEN $7 THEN $3 ELSE actual_cost END,
		    accepted_at = CASE WHEN $8 THEN CURRENT_TIMESTAMP ELSE accepted_at END,
		    completed_at = CASE WHEN $7 THEN CURRENT_TIMESTAMP ELSE completed_at END
		WHERE id = $4 AND status = $5 AND technician_id = $6`

	uploadedImagePaths := make([]string, 0, len(completionFiles))
	uploadedImageURLs := make([]string, 0, len(completionFiles))
	cleanupUploadedImages := func() {
		for _, path := range uploadedImagePaths {
			if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
				log.Printf("failed to remove uploaded repair image %s: %v", path, err)
			}
		}
	}
	for _, file := range completionFiles {
		filename := fmt.Sprintf("%s_after_%d%s", repairID, time.Now().UnixNano(), filepath.Ext(file.Filename))
		path := filepath.Join("uploads", filename)
		uploadedImagePaths = append(uploadedImagePaths, path)
		if err := c.SaveUploadedFile(file, path); err != nil {
			cleanupUploadedImages()
			log.Printf("failed to save completion image for repair %s: %v", repairID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "บันทึกรูปภาพไม่สำเร็จ กรุณาลองใหม่"})
			return
		}
		uploadedImageURLs = append(uploadedImageURLs, "/uploads/"+filename)
	}

	tx, err := database.DB.Begin()
	if err != nil {
		cleanupUploadedImages()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	result, err := tx.Exec(query, status, tNote, actualCost, repairID, currentStatus, auth.UserID(c), shouldUpdateActualCost, shouldSetAcceptedAt)
	if err != nil {
		tx.Rollback()
		cleanupUploadedImages()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		tx.Rollback()
		cleanupUploadedImages()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถตรวจสอบผลการอัปเดตสถานะได้"})
		return
	}
	if rowsAffected == 0 {
		tx.Rollback()
		cleanupUploadedImages()
		c.JSON(http.StatusConflict, gin.H{"error": "งานเปลี่ยนสถานะหรือผู้รับผิดชอบแล้ว กรุณารีเฟรชรายการ"})
		return
	}
	for _, imageURL := range uploadedImageURLs {
		if _, err = tx.Exec(`INSERT INTO repair_images (repair_id, image_url, image_type) VALUES ($1, $2, 'after')`, repairID, imageURL); err != nil {
			tx.Rollback()
			cleanupUploadedImages()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "บันทึกข้อมูลรูปภาพไม่สำเร็จ"})
			return
		}
	}
	if err = tx.Commit(); err != nil {
		cleanupUploadedImages()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 🔥 บันทึก Log การเปลี่ยนสถานะ
	insertRepairLog(repairID, actionByUserID, "STATUS_CHANGED", currentStatus, status, techNoteStr)

	// ส่งอีเมล
	go func() {
		var reporterEmail, description, tNumber, techEmail string
		if err := database.DB.QueryRow(`
			SELECT r.reporter_email, r.description, r.ticket_number, COALESCE(u.email, '')
			FROM repairs r LEFT JOIN users u ON r.technician_id = u.id WHERE r.id = $1`, repairID).
			Scan(&reporterEmail, &description, &tNumber, &techEmail); err != nil {
			log.Printf("failed to load repair notification details for repair %s: %v", repairID, err)
			return
		}

		if shouldNotifyCompletion {
			if reporterEmail != "" {
				userBodyContent := fmt.Sprintf(`
					<p><strong>รายการ:</strong> %s</p>
					<div class="box-success"><strong>สถานะ:</strong> งานซ่อมเสร็จเรียบร้อยแล้ว</div>
					<div class="box-success"><strong>บันทึกจากช่าง:</strong><br>%s</div>
					<a href="https://rb.gy/bjiuqq"
   class="btn btn-success"
   style="display:inline-block; background-color:#007A53; color:#ffffff !important; text-decoration:none; padding:12px 20px; border-radius:6px; font-weight:bold;">
   ให้คะแนนความพึงพอใจ
</a>
				`, html.EscapeString(description), html.EscapeString(techNoteStr))
				if err := utils.SendEmailNotification([]string{reporterEmail}, fmt.Sprintf("🎉 งานซ่อม %s เรียบร้อยแล้ว", tNumber), wrapEmail(colorSuccess, "🎉 งานซ่อมของคุณเสร็จเรียบร้อยแล้ว", userBodyContent)); err != nil {
					log.Printf("failed to email repair completion to reporter for repair %s: %v", repairID, err)
				}
			} else {
				log.Printf("cannot email repair completion for repair %s: reporter has no email address", repairID)
			}

			internalRecipients := adminEmailRecipients()
			if techEmail != "" {
				internalRecipients = append(internalRecipients, techEmail)
			}
			if len(internalRecipients) > 0 {
				internalBodyContent := fmt.Sprintf(`
					<p>งานซ่อมรายการนี้เสร็จเรียบร้อยแล้ว</p>
					<p><strong>หมายเลขใบงาน:</strong> %s</p>
					<p><strong>รายละเอียดงาน:</strong> %s</p>
					<div class="box-success"><strong>สถานะ:</strong> เสร็จเรียบร้อย</div>
					<div class="box-success"><strong>บันทึกจากช่าง:</strong><br>%s</div>
				`, html.EscapeString(tNumber), html.EscapeString(description), html.EscapeString(techNoteStr))
				if err := utils.SendEmailNotification(internalRecipients, fmt.Sprintf("✅ งานซ่อมเสร็จเรียบร้อยแล้ว (%s)", tNumber), wrapEmail(colorSuccess, "งานซ่อมเสร็จเรียบร้อยแล้ว", internalBodyContent)); err != nil {
					log.Printf("failed to email repair completion to admins and technician for repair %s: %v", repairID, err)
				}
			}
		} else if status == "ซ่อมไม่ได้" {
			bodyContent := fmt.Sprintf(`
				<p><strong>รายการ:</strong> %s</p>
				<div class="box-danger"><strong>เหตุผล:</strong><br>%s</div>
			`, html.EscapeString(description), html.EscapeString(techNoteStr))
			recipients := adminEmailRecipients()
			if techEmail != "" {
				recipients = append(recipients, techEmail)
			}
			if reporterEmail != "" {
				recipients = append(recipients, reporterEmail)
			}
			if err := utils.SendEmailNotification(recipients, fmt.Sprintf("⚠️ รายงานปัญหาการซ่อม (%s)", tNumber), wrapEmail(colorDanger, "ช่างไม่สามารถดำเนินการซ่อมได้", bodyContent)); err != nil {
				log.Printf("failed to email non-repairable notification for repair %s: %v", repairID, err)
			}
		}
	}()

	c.JSON(http.StatusOK, gin.H{"message": "อัปเดตสถานะสำเร็จ"})
}

// ---------------------------------------------------------
// 8. CancelRepairByAdmin: Admin ยกเลิกงาน (🔥 อัปเกรดใส่ Log)
// ---------------------------------------------------------
func CancelRepairByAdmin(c *gin.Context) {
	repairID := c.Param("id")
	type CancelRequest struct {
		AdminNote string `json:"admin_note" binding:"required"`
	}

	var req CancelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "กรุณาระบุเหตุผลที่ยกเลิกงาน"})
		return
	}

	var currentStatus string
	if err := database.DB.QueryRow("SELECT status FROM repairs WHERE id = $1", repairID).Scan(&currentStatus); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบรายการแจ้งซ่อม"})
		return
	}

	_, err := database.DB.Exec(
		`UPDATE repairs SET status = 'ซ่อมไม่ได้', admin_note = $1, completed_at = CURRENT_TIMESTAMP WHERE id = $2`,
		req.AdminNote, repairID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 🔥 บันทึก Log งานถูกแอดมินยกเลิก
	insertRepairLog(repairID, auth.UserID(c), "CANCELLED", currentStatus, "ซ่อมไม่ได้", "Admin ยกเลิกงาน: "+req.AdminNote)

	go func() {
		var reporterEmail, description, ticketNumber, technicianEmail string
		err := database.DB.QueryRow(`
			SELECT r.reporter_email, r.description, r.ticket_number, COALESCE(u.email, '')
			FROM repairs r LEFT JOIN users u ON r.technician_id = u.id WHERE r.id = $1`, repairID).
			Scan(&reporterEmail, &description, &ticketNumber, &technicianEmail)
		if err != nil {
			log.Printf("failed to load repair rejection email details for %s: %v", repairID, err)
			return
		}

		recipients := make([]string, 0, 2)
		if reporterEmail != "" {
			recipients = append(recipients, reporterEmail)
		}
		if technicianEmail != "" && technicianEmail != reporterEmail {
			recipients = append(recipients, technicianEmail)
		}
		if len(recipients) == 0 {
			log.Printf("cannot email rejected repair notification for repair %s: no recipients have email addresses", repairID)
			return
		}

		bodyContent := fmt.Sprintf(`
			<p><strong>หมายเลขใบแจ้ง:</strong> %s</p>
			<p><strong>รายการ:</strong> %s</p>
			<div class="box-danger"><strong>ผลการพิจารณา:</strong> ไม่อนุมัติการซ่อม</div>
			<div class="box-danger"><strong>เหตุผล:</strong><br>%s</div>
		`, html.EscapeString(ticketNumber), html.EscapeString(description), html.EscapeString(req.AdminNote))
		if err := utils.SendEmailNotification(recipients,
			fmt.Sprintf("ผลการพิจารณางานซ่อม: ไม่อนุมัติ (Ticket %s)", ticketNumber),
			wrapEmail(colorDanger, "ไม่อนุมัติการซ่อม", bodyContent)); err != nil {
			log.Printf("failed to email repair rejection for %s: %v", ticketNumber, err)
		}
	}()

	c.JSON(http.StatusOK, gin.H{"message": "ยกเลิกงานสำเร็จ (สถานะ: ซ่อมไม่ได้)"})
}

func CloseRepairEarly(c *gin.Context) {
	repairID := c.Param("id")
	var req struct {
		Note string `json:"note" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "กรุณาระบุหมายเหตุการปิดงาน"})
		return
	}
	req.Note = strings.TrimSpace(req.Note)
	if req.Note == "" || len(req.Note) > 2000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "หมายเหตุต้องมี 1-2000 ตัวอักษร"})
		return
	}

	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถเริ่มปิดงานได้"})
		return
	}
	defer tx.Rollback()

	var currentStatus string
	if err := tx.QueryRow("SELECT status FROM repairs WHERE id = $1 FOR UPDATE", repairID).Scan(&currentStatus); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบรายการแจ้งซ่อม"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถตรวจสอบสถานะงานได้"})
		return
	}
	if !canCloseRepairEarly(currentStatus) {
		c.JSON(http.StatusConflict, gin.H{"error": "ปิดงานก่อนครบขั้นตอนได้เฉพาะงานรอมอบหมาย งานที่ส่งคืน หรือรายการซ่อมไม่ได้"})
		return
	}

	if _, err := tx.Exec(`
		UPDATE repairs
		SET status = $1, closure_note = $2, completed_at = CURRENT_TIMESTAMP
		WHERE id = $3 AND status = $4
	`, statusClosedEarly, req.Note, repairID, currentStatus); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถบันทึกสถานะปิดงานได้"})
		return
	}
	if _, err := tx.Exec(`
		INSERT INTO repair_logs (repair_id, user_id, action, old_status, new_status, note)
		VALUES ($1, $2, 'CLOSED_EARLY', $3, $4, $5)
	`, repairID, auth.UserID(c), currentStatus, statusClosedEarly, req.Note); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถบันทึกประวัติการปิดงานได้"})
		return
	}
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถบันทึกการปิดงานได้"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ปิดงานก่อนครบขั้นตอนสำเร็จ", "status": statusClosedEarly})
}

func canCloseRepairEarly(status string) bool {
	return status == statusWaitingAdmin ||
		status == statusReturnedToAdmin ||
		status == "ซ่อมไม่ได้"
}

// OutsourceRepair บันทึกการส่งงานที่ช่างซ่อมไม่ได้ให้คณะหรือผู้รับจ้างภายนอก
func outsourceStatus(destination string) (string, string, bool) {
	switch strings.TrimSpace(destination) {
	case "faculty":
		return "ส่งซ่อมให้คณะ", "ให้คณะ", true
	case "external":
		return "ส่งซ่อมภายนอก", "ภายนอก", true
	default:
		return "", "", false
	}
}

func OutsourceRepair(c *gin.Context) {
	repairID := c.Param("id")
	type OutsourceRequest struct {
		Details     string `json:"details" binding:"required"`
		Destination string `json:"destination" binding:"required"`
	}

	var req OutsourceRequest
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Details) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "กรุณาระบุรายละเอียดการส่งซ่อม"})
		return
	}

	statusLabel, destinationLabel, validDestination := outsourceStatus(req.Destination)
	if !validDestination {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ประเภทการส่งซ่อมไม่ถูกต้อง"})
		return
	}

	result, err := database.DB.Exec(`
		UPDATE repairs SET status = $1, technician_id = NULL, accepted_at = NULL,
		    completed_at = NULL, admin_note = $2
		WHERE id = $3 AND status = 'ซ่อมไม่ได้'`,
		statusLabel, "ส่งซ่อม"+destinationLabel+": "+strings.TrimSpace(req.Details), repairID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if rowsAffected == 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "ส่งซ่อมต่อได้เฉพาะงานที่ช่างแจ้งว่าซ่อมไม่ได้"})
		return
	}

	insertRepairLog(repairID, auth.UserID(c), "OUTSOURCED", "ซ่อมไม่ได้", statusLabel, req.Details)
	go func() {
		var reporterEmail, ticketNumber, description string
		if err := database.DB.QueryRow(`
			SELECT reporter_email, ticket_number, description FROM repairs WHERE id = $1`,
			repairID,
		).Scan(&reporterEmail, &ticketNumber, &description); err != nil {
			log.Printf("failed to load outsourced repair email details for %s: %v", repairID, err)
			return
		}
		if strings.TrimSpace(reporterEmail) == "" {
			log.Printf("cannot email outsourced repair update for %s: reporter has no email address", repairID)
			return
		}

		bodyContent := fmt.Sprintf(`
			<p><strong>หมายเลขใบแจ้ง:</strong> %s</p>
			<p><strong>รายการ:</strong> %s</p>
			<div class="box-info"><strong>สถานะ:</strong> %s</div>
			<div class="box-info"><strong>รายละเอียด:</strong><br>%s</div>
		`, html.EscapeString(ticketNumber), html.EscapeString(description), html.EscapeString(statusLabel), html.EscapeString(strings.TrimSpace(req.Details)))
		if err := utils.SendEmailNotification(
			[]string{reporterEmail},
			fmt.Sprintf("อัปเดตงานซ่อม: %s (%s)", statusLabel, ticketNumber),
			wrapEmail(colorInfo, "อัปเดตสถานะงานซ่อม", bodyContent),
		); err != nil {
			log.Printf("failed to email outsourced repair update for %s: %v", repairID, err)
		}
	}()
	c.JSON(http.StatusOK, gin.H{"message": "บันทึกการส่งซ่อมต่อสำเร็จ", "status": statusLabel})
}

// ---------------------------------------------------------
// 9. EstimateRepair: ประเมินราคา & เบิกจุดคุ้มทุน (🔥 อัปเกรดใส่ Log)
// ---------------------------------------------------------
type breakEvenSummary struct {
	HasEquipment          bool    `json:"has_equipment"`
	EquipmentBasePrice    float64 `json:"equipment_base_price"`
	AccumulatedRepairCost float64 `json:"accumulated_repair_cost"`
	Threshold30           float64 `json:"threshold_30"`
	Threshold50           float64 `json:"threshold_50"`
	UsagePercent          float64 `json:"usage_percent"`
	AlertLevel            string  `json:"alert_level"`
	RiskMessage           string  `json:"risk_message"`
	SuggestedAction       string  `json:"suggested_action"`
}

func estimateNeedsAdminApproval(hasEquipment bool, basePrice, accumulatedCost, estimatedCost float64, forceReview bool) bool {
	if forceReview {
		return true
	}
	if !hasEquipment {
		return false
	}
	return estimatedCost > basePrice*0.5 || accumulatedCost+estimatedCost > basePrice*0.7
}

func buildBreakEvenSummary(hasEquipment bool, basePrice, accumulatedCost float64) breakEvenSummary {
	summary := breakEvenSummary{
		HasEquipment:          hasEquipment,
		EquipmentBasePrice:    basePrice,
		AccumulatedRepairCost: accumulatedCost,
		Threshold30:           basePrice * 0.3,
		Threshold50:           basePrice * 0.5,
		AlertLevel:            "normal",
	}
	if !hasEquipment {
		summary.AlertLevel = "no_equipment"
		summary.RiskMessage = "งานนี้ไม่ได้ผูกกับครุภัณฑ์ จึงประเมินจุดคุ้มทุนไม่ได้"
		summary.SuggestedAction = "ผู้ดูแลอาจผูกงานกับครุภัณฑ์ หากต้องการใช้ข้อมูลนี้ประกอบการพิจารณา"
		return summary
	}
	if basePrice <= 0 {
		summary.AlertLevel = "no_base_price"
		summary.RiskMessage = "ยังไม่มีราคาทุนของครุภัณฑ์ จึงคำนวณจุดคุ้มทุนไม่ได้"
		summary.SuggestedAction = "ผู้ดูแลอาจปรับปรุงราคาทุนในทะเบียนอุปกรณ์; ระบบจะไม่เปลี่ยนราคาให้อัตโนมัติ"
		return summary
	}

	summary.UsagePercent = accumulatedCost / basePrice * 100
	if accumulatedCost >= summary.Threshold50 {
		summary.AlertLevel = "high"
		summary.RiskMessage = "ค่าซ่อมสะสมถึงหรือเกินจุดเตือน 50% ของราคาทุนแล้ว"
		summary.SuggestedAction = "ช่างและผู้ดูแลอาจเปรียบเทียบการซ่อมต่อกับราคาอุปกรณ์ใหม่หรือใบเสนอราคาภายนอก; หากต้องปรับราคาค่าบริการหรือวงเงิน ให้ผู้ดูแลตัดสินใจเอง"
	} else if accumulatedCost >= summary.Threshold30 {
		summary.AlertLevel = "warning"
		summary.RiskMessage = "ค่าซ่อมสะสมเข้าใกล้จุดเตือน 50% ของราคาทุนแล้ว"
		summary.SuggestedAction = "ช่างอาจตรวจสภาพและขอราคาประเมินอะไหล่เพิ่มเติม เพื่อให้ผู้ดูแลใช้ประกอบการพิจารณา"
	} else {
		summary.RiskMessage = "ค่าซ่อมสะสมยังต่ำกว่าจุดเตือน 30% ของราคาทุน"
		summary.SuggestedAction = "ดำเนินงานตามขั้นตอนปกติ และบันทึกค่าใช้จ่ายจริงเมื่อปิดงาน"
	}
	return summary
}

func GetBreakEvenSummary(c *gin.Context) {
	repairID := c.Param("id")
	if !requireRepairAccess(c, repairID, false) {
		return
	}
	var equipmentID sql.NullInt64
	var basePrice sql.NullFloat64
	if err := database.DB.QueryRow(`
		SELECT r.equipment_id, e.base_price
		FROM repairs r LEFT JOIN equipments e ON e.id = r.equipment_id
		WHERE r.id = $1`, repairID).Scan(&equipmentID, &basePrice); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบรายการแจ้งซ่อม"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถอ่านข้อมูลอุปกรณ์ได้"})
		return
	}

	accumulatedCost := float64(0)
	if equipmentID.Valid {
		if err := database.DB.QueryRow(`
			SELECT COALESCE(SUM(actual_cost), 0)
			FROM repairs
			WHERE equipment_id = $1 AND status IN ('เสร็จเรียบร้อย', 'เสร็จสิ้น')`, equipmentID.Int64).
			Scan(&accumulatedCost); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถคำนวณยอดซ่อมสะสมได้"})
			return
		}
	}

	c.JSON(http.StatusOK, buildBreakEvenSummary(
		equipmentID.Valid,
		basePrice.Float64,
		accumulatedCost,
	))
}

func EstimateRepair(c *gin.Context) {
	repairID := c.Param("id")
	if !requireRepairAccess(c, repairID, false) {
		return
	}
	type EstimateRequest struct {
		EstimatedCost    float64 `json:"estimated_cost"`
		TechID           int     `json:"-"`
		CheckOnly        bool    `json:"check_only"`
		ForceAdminReview bool    `json:"force_admin_review"`
	}

	var req EstimateRequest
	req.TechID = auth.UserID(c)
	var estimateFiles []*multipart.FileHeader
	if strings.HasPrefix(c.ContentType(), "multipart/form-data") {
		if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ไม่สามารถอ่านข้อมูลประเมินราคาและรูปภาพได้"})
			return
		}
		var err error
		req.EstimatedCost, err = strconv.ParseFloat(c.PostForm("estimated_cost"), 64)
		req.CheckOnly = c.PostForm("check_only") == "true"
		req.ForceAdminReview = c.PostForm("force_admin_review") == "true"
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "กรุณาระบุราคาประเมินให้ถูกต้อง"})
			return
		}
		estimateFiles = c.Request.MultipartForm.File["image"]
	} else if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "กรุณาระบุราคาประเมินให้ถูกต้อง"})
		return
	}
	if math.IsNaN(req.EstimatedCost) || math.IsInf(req.EstimatedCost, 0) || req.EstimatedCost < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "กรุณาระบุราคาประเมินให้ถูกต้อง"})
		return
	}
	if validationError := validateRepairImageFiles(estimateFiles); validationError != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": validationError})
		return
	}

	var currentStatus string
	if err := database.DB.QueryRow("SELECT status FROM repairs WHERE id = $1", repairID).Scan(&currentStatus); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบรายการแจ้งซ่อม"})
		return
	}
	if !req.CheckOnly && currentStatus != "รับงานแล้ว" && currentStatus != "กำลังซ่อม" {
		c.JSON(http.StatusConflict, gin.H{"error": "ประเมินหรือเริ่มซ่อมได้เฉพาะงานที่รับไว้แล้ว"})
		return
	}

	var eqID *int
	var basePrice float64
	err := database.DB.QueryRow(`
		SELECT r.equipment_id, COALESCE(e.base_price, 0)
		FROM repairs r JOIN equipments e ON r.equipment_id = e.id
		WHERE r.id = $1
	`, repairID).Scan(&eqID, &basePrice)

	if err != nil && err != sql.ErrNoRows {
		log.Printf("failed to load equipment for repair estimate %s: %v", repairID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถตรวจสอบข้อมูลอุปกรณ์เพื่อประเมินราคาได้"})
		return
	}

	var accumulatedCost float64
	if err == nil && eqID != nil {
		if err := database.DB.QueryRow(`SELECT COALESCE(SUM(actual_cost), 0) FROM repairs WHERE equipment_id = $1 AND status IN ('เสร็จเรียบร้อย', 'เสร็จสิ้น')`, *eqID).Scan(&accumulatedCost); err != nil {
			log.Printf("failed to calculate accumulated repair cost for repair %s: %v", repairID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถตรวจสอบต้นทุนสะสมของอุปกรณ์ได้"})
			return
		}
	}

	newStatus := "กำลังซ่อม"
	adminSystemNote := ""

	requiresApproval := currentStatus == "รับงานแล้ว" &&
		estimateNeedsAdminApproval(eqID != nil, basePrice, accumulatedCost, req.EstimatedCost, req.ForceAdminReview)
	if requiresApproval {
		newStatus = statusWaitingForApproval
		adminSystemNote = fmt.Sprintf(
			"แจ้งเตือน: ราคาประเมินหรือยอดซ่อมสะสมเกินเกณฑ์ (ราคาทุน %.2f บาท)",
			basePrice,
		)
	}

	if req.CheckOnly {
		breakEven := buildBreakEvenSummary(eqID != nil, basePrice, accumulatedCost)
		c.JSON(http.StatusOK, gin.H{
			"status":                     newStatus,
			"estimated_cost":             req.EstimatedCost,
			"accumulated_cost":           accumulatedCost,
			"estimated_accumulated_cost": accumulatedCost + req.EstimatedCost,
			"single_repair_limit":        basePrice * 0.5,
			"accumulated_repair_limit":   basePrice * 0.7,
			"requires_approval":          requiresApproval,
			"break_even_summary":         breakEven,
		})
		return
	}
	uploadedImagePaths := make([]string, 0, len(estimateFiles))
	uploadedImageURLs := make([]string, 0, len(estimateFiles))
	cleanupUploadedImages := func() {
		for _, path := range uploadedImagePaths {
			if removeErr := os.Remove(path); removeErr != nil && !os.IsNotExist(removeErr) {
				log.Printf("failed to remove uploaded estimate image %s: %v", path, removeErr)
			}
		}
	}
	for _, file := range estimateFiles {
		filename := fmt.Sprintf("%s_before_%d%s", repairID, time.Now().UnixNano(), filepath.Ext(file.Filename))
		path := filepath.Join("uploads", filename)
		uploadedImagePaths = append(uploadedImagePaths, path)
		if err := c.SaveUploadedFile(file, path); err != nil {
			cleanupUploadedImages()
			log.Printf("failed to save estimate image for repair %s: %v", repairID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "บันทึกรูปภาพไม่สำเร็จ กรุณาลองใหม่"})
			return
		}
		uploadedImageURLs = append(uploadedImageURLs, "/uploads/"+filename)
	}

	tx, err := database.DB.Begin()
	if err != nil {
		cleanupUploadedImages()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถบันทึกราคาประเมินได้"})
		return
	}
	result, err := tx.Exec(`UPDATE repairs SET estimated_cost = $1, status = $2,
		admin_note = CASE WHEN $3 = '' THEN admin_note ELSE $3 END,
		accepted_at = CASE WHEN $4 THEN NULL ELSE CURRENT_TIMESTAMP END
		WHERE id = $5 AND status = $6 AND technician_id = $7`,
		req.EstimatedCost, newStatus, adminSystemNote, newStatus == statusWaitingForApproval,
		repairID, currentStatus, req.TechID)
	if err != nil {
		tx.Rollback()
		cleanupUploadedImages()
		log.Printf("failed to save repair estimate %s: %v", repairID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถบันทึกราคาประเมินได้"})
		return
	}
	if rows, rowsErr := result.RowsAffected(); rowsErr != nil || rows == 0 {
		tx.Rollback()
		cleanupUploadedImages()
		c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบรายการแจ้งซ่อม"})
		return
	}
	for _, imageURL := range uploadedImageURLs {
		if _, err := tx.Exec(`INSERT INTO repair_images (repair_id, image_url, image_type) VALUES ($1, $2, 'before')`, repairID, imageURL); err != nil {
			tx.Rollback()
			cleanupUploadedImages()
			log.Printf("failed to save estimate image record for repair %s: %v", repairID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "บันทึกข้อมูลรูปภาพไม่สำเร็จ"})
			return
		}
	}
	if err := tx.Commit(); err != nil {
		cleanupUploadedImages()
		log.Printf("failed to commit repair estimate %s: %v", repairID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถบันทึกราคาประเมินได้"})
		return
	}

	if req.TechID > 0 && req.EstimatedCost == 0 {
		insertRepairLog(repairID, req.TechID, "STATUS_CHANGED", currentStatus, newStatus, "ช่างเริ่มดำเนินการซ่อม")
	} else {
		insertRepairLog(repairID, req.TechID, "ESTIMATED", currentStatus, newStatus, fmt.Sprintf("ประเมินราคา: %.2f บาท", req.EstimatedCost))
	}

	go sendEstimateEmail(repairID, newStatus, adminSystemNote)

	responseMsg := "บันทึกข้อมูลและเริ่มดำเนินการซ่อมแล้ว"
	if req.TechID == 0 {
		responseMsg = "บันทึกราคาประเมินเรียบร้อย"
	}

	c.JSON(http.StatusOK, gin.H{
		"message":          responseMsg,
		"status":           newStatus,
		"estimated_cost":   req.EstimatedCost,
		"accumulated_cost": accumulatedCost,
	})
}

// ---------------------------------------------------------
// 10. sendEstimateEmail: แจ้งเตือนหลัง EstimateRepair
// ---------------------------------------------------------

func sendEstimateEmail(repairID string, newStatus string, adminSystemNote string) {
	var reporterEmail, description, ticketNumber, techEmail string
	var estimatedCost float64
	err := database.DB.QueryRow(`
		SELECT r.reporter_email, r.description, r.ticket_number, r.estimated_cost, COALESCE(u.email, '')
		FROM repairs r LEFT JOIN users u ON r.technician_id = u.id WHERE r.id = $1`, repairID).
		Scan(&reporterEmail, &description, &ticketNumber, &estimatedCost, &techEmail)
	if err != nil {
		log.Printf("failed to load estimate email details for repair %s: %v", repairID, err)
		return
	}

	if newStatus == statusWaitingForApproval {
		// 🔴 เกินจุดคุ้มทุน -> แจ้ง Admin ให้เข้ามาพิจารณาอนุมัติ
		bodyContent := fmt.Sprintf(`
			<p><strong>หมายเลขใบแจ้ง:</strong> %s</p>
			<p><strong>รายการ:</strong> %s</p>
			<p><strong>ราคาประเมินจากช่าง:</strong> %.2f บาท</p>
			<div class="box-danger"><strong>แจ้งเตือนจุดคุ้มทุน:</strong><br>%s</div>
			<p>กรุณาเข้าสู่ระบบเพื่อพิจารณาอนุมัติหรือไม่อนุมัติงานนี้</p>
		`, html.EscapeString(ticketNumber), html.EscapeString(description), estimatedCost, html.EscapeString(adminSystemNote))
		subject := fmt.Sprintf("แจ้งเตือนจุดคุ้มทุน: งานรอพิจารณา (Ticket %s)", ticketNumber)
		body := wrapEmail(colorDanger, "งานซ่อมเกินเกณฑ์จุดคุ้มทุน", bodyContent)

		if err := sendAdminEmail(subject, body); err != nil {
			log.Printf("failed to email threshold approval request for repair %s: %v", repairID, err)
		}
		recipients := make([]string, 0, 2)
		if reporterEmail != "" {
			recipients = append(recipients, reporterEmail)
		}
		if techEmail != "" {
			recipients = append(recipients, techEmail)
		}
		statusBody := fmt.Sprintf(`
			<p><strong>หมายเลขใบแจ้ง:</strong> %s</p>
			<p><strong>รายการ:</strong> %s</p>
			<p><strong>ราคาประเมิน:</strong> %.2f บาท</p>
			<div class="box-warning">งานอยู่ระหว่างรอผู้ดูแลระบบพิจารณาจุดคุ้มทุน กรุณารอผลอนุมัติก่อนดำเนินการต่อ</div>
		`, html.EscapeString(ticketNumber), html.EscapeString(description), estimatedCost)
		if err := utils.SendEmailNotification(recipients, fmt.Sprintf("งานซ่อมรออนุมัติงบ (%s)", ticketNumber), wrapEmail(colorWarning, "งานซ่อมรอผู้ดูแลระบบพิจารณา", statusBody)); err != nil {
			log.Printf("failed to email threshold pending status to reporter and technician for repair %s: %v", repairID, err)
		}
	} else {
		// 🟢 ไม่เกินจุดคุ้มทุน -> แจ้งผู้แจ้ง + แจ้งกลุ่มภายใน (Admin + ช่าง)

		// 1. ส่งหาผู้แจ้งซ่อม
		if reporterEmail != "" {
			bodyContent := fmt.Sprintf(`
				<p><strong>หมายเลขใบแจ้ง:</strong> %s</p>
				<p><strong>รายการ:</strong> %s</p>
			`, html.EscapeString(ticketNumber), html.EscapeString(description))
			subject := fmt.Sprintf("ช่างเริ่มดำเนินการซ่อมแล้ว (Ticket %s)", ticketNumber)
			body := wrapEmail(colorInfo, "ช่างได้เริ่มดำเนินการซ่อมแล้ว", bodyContent)

			if err := utils.SendEmailNotification([]string{reporterEmail}, subject, body); err != nil {
				log.Printf("failed to email estimate update to reporter for repair %s: %v", repairID, err)
			}
		} else {
			log.Printf("cannot email estimate update for repair %s: reporter has no email address", repairID)
		}

		// 2. ✨ [เพิ่มใหม่] ส่งหา Admin + ช่าง ยืนยันการรับงานและเริ่มซ่อม
		internalRecipients := adminEmailRecipients()
		if techEmail != "" {
			internalRecipients = append(internalRecipients, techEmail)
		}
		internalBodyContent := fmt.Sprintf(`
			<p><strong>หมายเลขใบแจ้ง:</strong> %s</p>
			<p><strong>รายการ:</strong> %s</p>
			<p>สถานะปัจจุบัน: <b>กำลังซ่อม</b> (ราคาประเมินผ่านเกณฑ์จุดคุ้มทุนเรียบร้อยแล้ว)</p>
		`, html.EscapeString(ticketNumber), html.EscapeString(description))
		internalSubject := fmt.Sprintf("ยืนยันการเริ่มซ่อม (Ticket %s)", ticketNumber)
		internalBody := wrapEmail(colorInfo, "ช่างรับงานและเริ่มดำเนินการซ่อมแล้ว", internalBodyContent)

		if err := utils.SendEmailNotification(internalRecipients, internalSubject, internalBody); err != nil {
			log.Printf("failed to email estimate update to admins and technician for repair %s: %v", repairID, err)
		}
	}
}

// ---------------------------------------------------------
// 11. ApproveRepairThreshold: Admin อนุมัติซ่อมเกินงบ (🔥 อัปเกรดใส่ Log)
// ---------------------------------------------------------
func ApproveRepairThreshold(c *gin.Context) {
	repairID := c.Param("id")

	var currentStatus string
	if err := database.DB.QueryRow("SELECT status FROM repairs WHERE id = $1", repairID).Scan(&currentStatus); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบรายการแจ้งซ่อม"})
		return
	}

	result, err := database.DB.Exec(`
		UPDATE repairs 
		SET status = 'กำลังซ่อม', accepted_at = CURRENT_TIMESTAMP, admin_note = CONCAT(admin_note, ' -> (Admin อนุมัติให้ดำเนินการต่อ)')
		WHERE id = $1 AND status = $2 AND technician_id IS NOT NULL
	`, repairID, statusWaitingForApproval)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถอนุมัติงานได้: " + err.Error()})
		return
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถตรวจสอบผลการอนุมัติงานได้"})
		return
	}
	if rowsAffected == 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "อนุมัติได้เฉพาะงานที่รอพิจารณาและมีช่างรับผิดชอบ"})
		return
	}

	// 🔥 บันทึก Log การอนุมัติ
	insertRepairLog(repairID, auth.UserID(c), "APPROVED", currentStatus, "กำลังซ่อม", "Admin อนุมัติให้ซ่อมงานที่เกินจุดคุ้มทุนได้")

	go func() {
		var description, ticketNumber, techEmail, reporterEmail string
		var estimatedCost float64
		if err := database.DB.QueryRow(`
			SELECT r.description, r.ticket_number, r.estimated_cost, COALESCE(u.email, ''), r.reporter_email
			FROM repairs r LEFT JOIN users u ON r.technician_id = u.id WHERE r.id = $1`, repairID).
			Scan(&description, &ticketNumber, &estimatedCost, &techEmail, &reporterEmail); err != nil {
			log.Printf("failed to load repair approval email details for %s: %v", repairID, err)
			return
		}
		recipients := make([]string, 0, 2)
		if techEmail != "" {
			recipients = append(recipients, techEmail)
		}
		if reporterEmail != "" {
			recipients = append(recipients, reporterEmail)
		}
		bodyContent := fmt.Sprintf(`<p><strong>หมายเลขใบแจ้ง:</strong> %s</p><p><strong>รายการ:</strong> %s</p><p><strong>ราคาประเมินที่อนุมัติ:</strong> %.2f บาท</p><p>ผู้ดูแลระบบอนุมัติงบแล้ว ช่างสามารถดำเนินการซ่อมต่อได้</p>`, html.EscapeString(ticketNumber), html.EscapeString(description), estimatedCost)
		if err := utils.SendEmailNotification(recipients, fmt.Sprintf("อนุมัติงานซ่อมแล้ว (Ticket %s)", ticketNumber), wrapEmail(colorSuccess, "อนุมัติให้ดำเนินการซ่อมต่อ", bodyContent)); err != nil {
			log.Printf("failed to email repair approval update to technician and reporter for repair %s: %v", repairID, err)
		}
	}()

	c.JSON(http.StatusOK, gin.H{"message": "อนุมัติงานซ่อมสำเร็จ"})
}
