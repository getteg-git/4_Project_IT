package repair

import (
	"database/sql"
	"fmt"
	"html"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"backend/database"
	"backend/utils" // ใช้ utils.SendEmailNotification
)

// =========================================================
// Helper Functions (ฟังก์ชันผู้ช่วย)
// =========================================================

// nullIfEmpty สำหรับจัดการค่าว่างให้เป็น NULL ใน Database
func nullIfEmpty(value string) interface{} {
	if value == "" || value == "null" || value == "undefined" {
		return nil
	}
	return value
}

// generateTicketNumber สร้างรหัสตั๋วอัตโนมัติ (เช่น REQ-260901-001)
func generateTicketNumber() string {
	now := time.Now()
	prefix := fmt.Sprintf("REQ-%s", now.Format("060102")) // รูปแบบ: REQ-YYMMDD
	var count int
	// นับจำนวนงานซ่อมของวันนี้เพื่อรันเลขต่อ
	database.DB.QueryRow("SELECT COUNT(*) FROM repairs WHERE DATE(created_at) = CURRENT_DATE").Scan(&count)
	return fmt.Sprintf("%s-%03d", prefix, count+1)
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

// =========================================================
// ธีมสีกลาง สำหรับอีเมลทุกฉบับในระบบ
// =========================================================
const (
	colorSuccess = "#166534" // เขียวเข้มเพื่อให้อ่านชัดบนพื้นขาว
	colorInfo    = "#2563eb" // ฟ้า: แจ้งเตือนทั่วไป / ข้อมูล
	colorWarning = "#92400e" // ส้มเข้มเพื่อให้อ่านชัดบนพื้นขาว
	colorDanger  = "#dc2626" // แดง: ปัญหา / ยกเลิก / ปฏิเสธ
	frontendURL  = "https://4-project-it.vercel.app"
)

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
// 1. CreateRepair: สร้างใบแจ้งซ่อมใหม่ (อัปเกรด Ticket Number + Log)
// ---------------------------------------------------------
func CreateRepair(c *gin.Context) {
	ticketNumber := generateTicketNumber() // 🔥 Gen รหัสอัตโนมัติ
	reporterEmail := c.PostForm("reporter_email")
	departmentID := nullIfEmpty(c.PostForm("department_id")) // 🔥 รับสาขาของคนแจ้ง
	locationID := nullIfEmpty(c.PostForm("location_id"))
	otherLocation := nullIfEmpty(c.PostForm("other_location"))
	floorID := nullIfEmpty(c.PostForm("floor_id"))
	roomID := nullIfEmpty(c.PostForm("room_id"))
	equipmentID := nullIfEmpty(c.PostForm("equipment_id"))
	problemTypeID := nullIfEmpty(c.PostForm("problem_type_id"))
	otherProblemType := nullIfEmpty(c.PostForm("other_problem_type"))
	description := c.PostForm("description")

	var repairID int
	err := database.DB.QueryRow(
		`INSERT INTO repairs (ticket_number, reporter_email, department_id, location_id, other_location, floor_id, room_id, equipment_id, problem_type_id, other_problem_type, description, status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, 'รอซ่อม') RETURNING id`,
		ticketNumber, reporterEmail, departmentID, locationID, otherLocation, floorID, roomID, equipmentID, problemTypeID, otherProblemType, description,
	).Scan(&repairID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถบันทึกข้อมูลได้: " + err.Error()})
		return
	}

	// 🔥 บันทึก Log แรกเริ่ม
	insertRepairLog(repairID, nil, "CREATED", "", "รอซ่อม", "สร้างใบแจ้งซ่อมใหม่โดยผู้ใช้งาน")

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
		formURL := frontendURL + "/repair/history"

		// 1. อีเมลสำหรับผู้แจ้ง
		userBodyContent := fmt.Sprintf(`
			<p><strong>หมายเลขอ้างอิง:</strong> %s</p>
			<p><strong>รายละเอียด:</strong> %s</p>
			<div class="box-warning">ขณะนี้อยู่ในสถานะ: <strong>รอซ่อม</strong></div>
			<p>ทางเราจะแจ้งเตือนอีกครั้งเมื่อช่างเริ่มเข้าดำเนินการครับ</p>
			<a href="%s" class="btn">ตรวจสอบรายการแจ้งซ่อมได้ที่นี่</a>
		`, html.EscapeString(ticketNumber), html.EscapeString(description), formURL)
		userSubject := fmt.Sprintf("✅ รับเรื่องแจ้งซ่อมเรียบร้อยแล้ว (%s)", ticketNumber)
		userBody := wrapEmail(colorSuccess, "ระบบได้รับเรื่องแจ้งซ่อมของคุณเรียบร้อยแล้ว", userBodyContent)

		if reporterEmail != "" {
			if err := utils.SendEmailNotification([]string{reporterEmail}, userSubject, userBody); err != nil {
				log.Printf("failed to email repair receipt to reporter for repair %d: %v", repairID, err)
			}
		} else {
			log.Printf("cannot email repair receipt for repair %d: reporter has no email address", repairID)
		}

		// 2. อีเมลสำหรับ Admin
		adminLink := frontendURL + "/admin/manage"
		adminBodyContent := fmt.Sprintf(`
			<p>กรุณาเข้าสู่ระบบเพื่อพิจารณามอบหมายช่างดำเนินการ</p>
			<p><strong>หมายเลขอ้างอิง:</strong> %s</p>
			<p><strong>ผู้แจ้ง:</strong> %s</p>
			<p><strong>รายละเอียด:</strong> %s</p>
			<div class="box-warning">ขณะนี้อยู่ในสถานะ: <strong>รอซ่อม</strong></div>
			<a href="%s" class="btn">ตรวจสอบรายการได้ที่นี่</a>
		`, html.EscapeString(ticketNumber), html.EscapeString(reporterEmail), html.EscapeString(description), adminLink)
		adminSubject := fmt.Sprintf("🔔 มีงานแจ้งซ่อมใหม่เข้ามา (%s)", ticketNumber)
		adminBody := wrapEmail(colorInfo, "มีรายการแจ้งซ่อมใหม่เข้ามาในระบบ", adminBodyContent)

		if err := sendAdminEmail(adminSubject, adminBody); err != nil {
			log.Printf("failed to email new repair to admins: %v", err)
		}
	}()

	c.JSON(http.StatusCreated, gin.H{"message": "สร้างรายการแจ้งซ่อมสำเร็จ", "ticket_number": ticketNumber})
}

// ---------------------------------------------------------
// 2. GetAllRepairs: ดึงรายการซ่อมทั้งหมด (🔥 อัปเกรด Smart Routing)
// ---------------------------------------------------------
func GetAllRepairs(c *gin.Context) {
	role := c.Query("role")
	techID := c.Query("tech_id")
	deptID := c.Query("department_id")
	isCentral := c.Query("is_central")

	query := `
		SELECT r.id, r.ticket_number, r.description, r.status, r.created_at, r.technician_id, r.technician_note, r.admin_note, r.reporter_email,
			   r.estimated_cost, r.actual_cost, r.accepted_at, r.completed_at, r.other_location, r.other_problem_type,
			   r.department_id, d.name as department_name,
			   l.name as location_name, f.floor_name, ro.room_number,
			   eq.name as equipment_name, eq.asset_code, p.name as problem_type_name, u.full_name as technician_name
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

	// 🔥 กฎ Smart Routing: ถ้าคนล็อกอินเป็นช่าง ให้กรองงาน
	if role == "technician" && techID != "" {
		// กรองให้เห็นเฉพาะงานที่ "ตรงกับความถนัด" ของช่างคนนั้น
		query += fmt.Sprintf(` AND (r.problem_type_id IN (SELECT problem_type_id FROM technician_specialties WHERE user_id = %s) OR r.problem_type_id IS NULL)`, techID)

		// ถ้าไม่ใช่ช่างส่วนกลาง (is_central = false) ให้เห็นแค่ "งานของสาขาตัวเอง" และ "งานที่ไม่ได้ระบุสาขา"
		if isCentral != "true" && deptID != "" {
			query += fmt.Sprintf(` AND (r.department_id = %s OR r.department_id IS NULL)`, deptID)
		}
	}

	query += ` ORDER BY r.created_at DESC`

	rows, err := database.DB.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	repairs := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id int
		var desc, status, reporterEmail string
		var ticketNumber, locName, otherLoc, floorName, roomNumber, eqName, assetCode, probName, otherProb, techName, techNote, adminNote, deptName *string
		var estimatedCost, actualCost float64
		var createdAt time.Time
		var acceptedAt, completedAt *time.Time
		var techID, dID *int

		err := rows.Scan(&id, &ticketNumber, &desc, &status, &createdAt, &techID, &techNote, &adminNote, &reporterEmail,
			&estimatedCost, &actualCost, &acceptedAt, &completedAt, &otherLoc, &otherProb,
			&dID, &deptName, &locName, &floorName, &roomNumber, &eqName, &assetCode, &probName, &techName)
		if err != nil {
			continue
		}

		imgRows, _ := database.DB.Query("SELECT image_url, image_type FROM repair_images WHERE repair_id = $1", id)
		images := make([]map[string]string, 0)
		for imgRows.Next() {
			var url, imgType string
			imgRows.Scan(&url, &imgType)
			images = append(images, map[string]string{"url": url, "type": imgType})
		}
		imgRows.Close()

		getString := func(s *string) string {
			if s == nil {
				return ""
			}
			return *s
		}

		repairs = append(repairs, map[string]interface{}{
			"id":                 id,
			"ticket_number":      getString(ticketNumber),
			"description":        desc,
			"status":             status,
			"created_at":         createdAt,
			"department_id":      dID,
			"department_name":    getString(deptName),
			"location_name":      getString(locName),
			"other_location":     getString(otherLoc),
			"floor_name":         getString(floorName),
			"room_number":        getString(roomNumber),
			"equipment_name":     getString(eqName),
			"asset_code":         getString(assetCode),
			"problem_type":       getString(probName),
			"other_problem_type": getString(otherProb),
			"reporter_email":     reporterEmail,
			"technician_id":      techID,
			"technician_note":    getString(techNote),
			"admin_note":         getString(adminNote),
			"technician_name":    getString(techName),
			"estimated_cost":     estimatedCost,
			"actual_cost":        actualCost,
			"accepted_at":        acceptedAt,
			"completed_at":       completedAt,
			"images":             images,
		})
	}
	c.JSON(http.StatusOK, repairs)
}

// ---------------------------------------------------------
// 3. GetRepairByID: ดึงรายละเอียดใบแจ้งซ่อมเฉพาะใบ
// ---------------------------------------------------------
func GetRepairByID(c *gin.Context) {
	repairID := c.Param("id")

	var id int
	var desc, status, reporterEmail string
	var ticketNumber, locName, otherLoc, floorName, roomNumber, eqName, assetCode, probName, otherProb, techName, techNote, adminNote, deptName *string
	var estimatedCost, actualCost float64
	var createdAt time.Time
	var acceptedAt, completedAt *time.Time
	var dID *int

	err := database.DB.QueryRow(`
		SELECT r.id, r.ticket_number, r.description, r.status, r.reporter_email, r.technician_note, r.admin_note, r.created_at,
			   r.estimated_cost, r.actual_cost, r.accepted_at, r.completed_at, r.other_location, r.other_problem_type,
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
		&id, &ticketNumber, &desc, &status, &reporterEmail, &techNote, &adminNote, &createdAt,
		&estimatedCost, &actualCost, &acceptedAt, &completedAt, &otherLoc, &otherProb,
		&dID, &deptName, &locName, &floorName, &roomNumber, &eqName, &assetCode, &probName, &techName)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบรายการแจ้งซ่อม"})
		return
	}

	rows, _ := database.DB.Query("SELECT image_url, image_type FROM repair_images WHERE repair_id = $1", id)
	defer rows.Close()

	images := make([]map[string]string, 0)
	for rows.Next() {
		var url, imgType string
		rows.Scan(&url, &imgType)
		images = append(images, map[string]string{"url": url, "type": imgType})
	}

	getString := func(s *string) string {
		if s == nil {
			return ""
		}
		return *s
	}

	repair := map[string]interface{}{
		"id":                 id,
		"ticket_number":      getString(ticketNumber),
		"description":        desc,
		"status":             status,
		"reporter_email":     reporterEmail,
		"department_id":      dID,
		"department_name":    getString(deptName),
		"technician_note":    getString(techNote),
		"admin_note":         getString(adminNote),
		"technician_name":    getString(techName),
		"location_name":      getString(locName),
		"other_location":     getString(otherLoc),
		"floor_name":         getString(floorName),
		"room_number":        getString(roomNumber),
		"equipment_name":     getString(eqName),
		"asset_code":         getString(assetCode),
		"problem_type":       getString(probName),
		"other_problem_type": getString(otherProb),
		"estimated_cost":     estimatedCost,
		"actual_cost":        actualCost,
		"accepted_at":        acceptedAt,
		"completed_at":       completedAt,
		"created_at":         createdAt,
		"images":             images,
	}
	c.JSON(http.StatusOK, repair)
}

// ---------------------------------------------------------
// 4. AssignRepair: Admin มอบหมายช่าง (🔥 อัปเกรดใส่ Log)
// ---------------------------------------------------------
func AssignRepair(c *gin.Context) {
	repairID := c.Param("id")
	type AssignRequest struct {
		TechnicianID int `json:"technician_id" binding:"required"`
		AdminID      int `json:"admin_id"` // รับ ID แอดมินมาเพื่อเก็บ Log
	}

	var req AssignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ข้อมูลไม่ถูกต้อง"})
		return
	}

	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถเริ่มบันทึกการมอบหมายงานได้"})
		return
	}
	defer tx.Rollback()

	var currentStatus string
	err = tx.QueryRow("SELECT status FROM repairs WHERE id = $1 FOR UPDATE", repairID).Scan(&currentStatus)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบรายการแจ้งซ่อม"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	_, err = tx.Exec(
		`UPDATE repairs SET status = 'รอซ่อม', technician_id = $1, admin_note = NULL, accepted_at = NULL, completed_at = NULL WHERE id = $2`,
		req.TechnicianID, repairID,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	_, err = tx.Exec(`
		INSERT INTO repair_logs (repair_id, user_id, action, old_status, new_status, note)
		VALUES ($1, $2, 'ASSIGNED', $3, 'รอซ่อม', $4)`,
		repairID, req.AdminID, currentStatus, fmt.Sprintf("Admin มอบหมายงานให้ช่าง ID: %d", req.TechnicianID))
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
			`, html.EscapeString(tNumber), html.EscapeString(description), frontendURL)
			if err := utils.SendEmailNotification([]string{techEmail}, fmt.Sprintf("🔧 คุณได้รับมอบหมายงานซ่อมใหม่ (%s)", tNumber), wrapEmail(colorInfo, "คุณได้รับมอบหมายงานซ่อมใหม่", techBodyContent)); err != nil {
				log.Printf("failed to email repair assignment to technician for repair %s: %v", repairID, err)
			}
		} else {
			log.Printf("technician %d has no email for repair assignment %s", req.TechnicianID, repairID)
		}
	}()

	c.JSON(http.StatusOK, gin.H{"message": "มอบหมายงานให้ช่างสำเร็จ"})
}

// ---------------------------------------------------------
// 5. RevokeRepair: Admin ดึงงานกลับ (🔥 อัปเกรดใส่ Log)
// ---------------------------------------------------------
func RevokeRepair(c *gin.Context) {
	repairID := c.Param("id")
	revokeReason := strings.TrimSpace(c.PostForm("rejection_reason"))
	adminID := nullIfEmpty(c.PostForm("admin_id"))

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

	_, err = tx.Exec(`UPDATE repairs SET status = 'รอซ่อม', technician_id = NULL, admin_note = NULL, accepted_at = NULL, completed_at = NULL WHERE id = $1`, repairID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	_, err = tx.Exec(`
		INSERT INTO repair_logs (repair_id, user_id, action, old_status, new_status, note)
		VALUES ($1, $2, 'REVOKED', $3, 'รอซ่อม', $4)`,
		repairID, adminID, currentStatus, "Admin ดึงงานกลับจากช่าง. เหตุผล: "+revokeReason)
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
			`, html.EscapeString(techName), html.EscapeString(tNumber), html.EscapeString(description), html.EscapeString(revokeReason), frontendURL)
			if err := utils.SendEmailNotification([]string{techEmail}, fmt.Sprintf("⚠️ แจ้งเตือนการดึงงานซ่อมกลับ (%s)", tNumber), wrapEmail(colorDanger, "แจ้งเตือนการดึงงานซ่อมกลับ", bodyContent)); err != nil {
				log.Printf("failed to email revoked repair assignment to technician for repair %s: %v", repairID, err)
			}
		} else {
			log.Printf("cannot email revoked repair assignment for repair %s: technician %d has no email address", repairID, technicianID)
		}
	}()

	c.JSON(http.StatusOK, gin.H{"message": "ยกเลิกการจ่ายงานสำเร็จ"})
}

// ---------------------------------------------------------
// 6. RejectRepair: Tech ปฏิเสธงาน (🔥 อัปเกรดใส่ Log)
// ---------------------------------------------------------
func RejectRepair(c *gin.Context) {
	repairID := c.Param("id")
	rejectionReason := c.PostForm("rejection_reason")
	techID := nullIfEmpty(c.PostForm("tech_id"))

	var currentStatus string
	database.DB.QueryRow("SELECT status FROM repairs WHERE id = $1", repairID).Scan(&currentStatus)

	_, err := database.DB.Exec(`UPDATE repairs SET status = 'รอซ่อม', technician_id = NULL, accepted_at = NULL, completed_at = NULL WHERE id = $1`, repairID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 🔥 บันทึก Log ช่างปฏิเสธงาน
	insertRepairLog(repairID, techID, "REJECTED", currentStatus, "รอซ่อม", "ช่างปฏิเสธงาน. เหตุผล: "+rejectionReason)

	go func() {
		var reporterEmail, description, tNumber string
		database.DB.QueryRow(`SELECT reporter_email, description, ticket_number FROM repairs WHERE id = $1`, repairID).Scan(&reporterEmail, &description, &tNumber)

		// แจ้ง Admin + ผู้แจ้ง
		bodyContent := fmt.Sprintf(`
			<p><strong>รหัสตั๋ว:</strong> %s</p>
			<p><strong>รายการ:</strong> %s</p>
			<div class="box-danger"><strong>📌 เหตุผลในการปฏิเสธ:</strong><br>%s</div>
		`, html.EscapeString(tNumber), html.EscapeString(description), html.EscapeString(rejectionReason))

		recipients := adminEmailRecipients()
		if reporterEmail != "" {
			recipients = append(recipients, reporterEmail)
		}
		if err := utils.SendEmailNotification(recipients, fmt.Sprintf("⚠️ ช่างปฏิเสธงานซ่อม (%s)", tNumber), wrapEmail(colorDanger, "ช่างปฏิเสธงาน / ส่งกลับเข้าระบบ", bodyContent)); err != nil {
			log.Printf("failed to email rejected repair notification for repair %s: %v", repairID, err)
		}
	}()

	c.JSON(http.StatusOK, gin.H{"message": "ปฏิเสธงานและส่งคืนระบบสำเร็จ"})
}

// ---------------------------------------------------------
// 7. UpdateRepairStatus: อัปเดตสถานะ (🔥 ล็อกสเต็ป & บังคับ Note)
// ---------------------------------------------------------
func UpdateRepairStatus(c *gin.Context) {
	repairID := c.Param("id")
	status := c.PostForm("status")
	actionByUserID := nullIfEmpty(c.PostForm("action_by"))
	actualCostStr := c.PostForm("actual_cost")
	tNote := nullIfEmpty(c.PostForm("technician_note"))
	aNote := nullIfEmpty(c.PostForm("admin_note"))
	techNoteStr := c.PostForm("technician_note")

	if status == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "กรุณาระบุสถานะ"})
		return
	}

	// 🔥 Guardrail: บังคับสรุปงานก่อนปิดจ๊อบเพื่อลด Human Error
	if status == "เสร็จเรียบร้อย" && techNoteStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ระบบป้องกัน: กรุณากรอก 'บันทึกจากช่าง' เพื่อสรุปงานก่อนทำการปิดงานซ่อม"})
		return
	}

	var currentStatus string
	database.DB.QueryRow("SELECT status FROM repairs WHERE id = $1", repairID).Scan(&currentStatus)
	shouldNotifyStart := status == "กำลังซ่อม" && currentStatus != "กำลังซ่อม"
	shouldNotifyCompletion := status == "เสร็จเรียบร้อย" && currentStatus != "เสร็จเรียบร้อย"

	var actualCost float64
	if actualCostStr != "" && actualCostStr != "null" {
		actualCost, _ = strconv.ParseFloat(actualCostStr, 64)
	}

	var query string
	if status == "กำลังซ่อม" {
		query = `UPDATE repairs SET status = $1, technician_note = COALESCE($2, technician_note), admin_note = COALESCE($3, admin_note), actual_cost = $4, accepted_at = CURRENT_TIMESTAMP WHERE id = $5`
	} else if status == "เสร็จเรียบร้อย" || status == "ซ่อมไม่ได้" {
		query = `UPDATE repairs SET status = $1, technician_note = COALESCE($2, technician_note), admin_note = COALESCE($3, admin_note), actual_cost = $4, completed_at = CURRENT_TIMESTAMP WHERE id = $5`
	} else {
		query = `UPDATE repairs SET status = $1, technician_note = COALESCE($2, technician_note), admin_note = COALESCE($3, admin_note), actual_cost = $4 WHERE id = $5`
	}

	_, err := database.DB.Exec(query, status, tNote, aNote, actualCost, repairID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 🔥 บันทึก Log การเปลี่ยนสถานะ
	insertRepairLog(repairID, actionByUserID, "STATUS_CHANGED", currentStatus, status, techNoteStr)

	// เซฟรูปภาพ
	form, _ := c.MultipartForm()
	if form != nil && form.File != nil {
		files := form.File["image"]
		for _, file := range files {
			filename := fmt.Sprintf("%s_after_%d%s", repairID, time.Now().UnixNano(), filepath.Ext(file.Filename))
			path := "uploads/" + filename
			if err := c.SaveUploadedFile(file, path); err == nil {
				database.DB.Exec(`INSERT INTO repair_images (repair_id, image_url, image_type) VALUES ($1, $2, 'after')`, repairID, "/uploads/"+filename)
			}
		}
	}

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

		if shouldNotifyStart {
			bodyContent := fmt.Sprintf(`
				<p>ช่างเทคนิคเริ่มดำเนินการซ่อมแล้ว</p>
				<p><strong>หมายเลขใบงาน:</strong> %s</p>
				<p><strong>รายละเอียดงาน:</strong> %s</p>
				<div class="box-info">สถานะปัจจุบัน: <strong>กำลังซ่อม</strong></div>
			`, html.EscapeString(tNumber), html.EscapeString(description))
			recipients := adminEmailRecipients()
			if reporterEmail != "" {
				recipients = append(recipients, reporterEmail)
			}
			if techEmail != "" {
				recipients = append(recipients, techEmail)
			}
			if len(recipients) > 0 {
				if err := utils.SendEmailNotification(recipients, fmt.Sprintf("🔧 ช่างเริ่มดำเนินการซ่อมแล้ว (%s)", tNumber), wrapEmail(colorInfo, "ช่างเริ่มดำเนินการซ่อมแล้ว", bodyContent)); err != nil {
					log.Printf("failed to email repair start notification for repair %s: %v", repairID, err)
				}
			} else {
				log.Printf("cannot email repair start notification for repair %s: no recipients have email addresses", repairID)
			}
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
		AdminID   int    `json:"admin_id"`
	}

	var req CancelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "กรุณาระบุเหตุผลที่ยกเลิกงาน"})
		return
	}

	var currentStatus string
	database.DB.QueryRow("SELECT status FROM repairs WHERE id = $1", repairID).Scan(&currentStatus)

	_, err := database.DB.Exec(
		`UPDATE repairs SET status = 'ซ่อมไม่ได้', admin_note = $1, completed_at = CURRENT_TIMESTAMP WHERE id = $2`,
		req.AdminNote, repairID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 🔥 บันทึก Log งานถูกแอดมินยกเลิก
	insertRepairLog(repairID, req.AdminID, "CANCELLED", currentStatus, "ซ่อมไม่ได้", "Admin ยกเลิกงาน: "+req.AdminNote)

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

// OutsourceRepair บันทึกการส่งงานที่ช่างซ่อมไม่ได้ไปให้ผู้รับจ้างภายนอก
func OutsourceRepair(c *gin.Context) {
	repairID := c.Param("id")
	type OutsourceRequest struct {
		Details string `json:"details" binding:"required"`
		AdminID int    `json:"admin_id"`
	}

	var req OutsourceRequest
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Details) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "กรุณาระบุผู้รับจ้างหรือรายละเอียดการส่งซ่อมภายนอก"})
		return
	}

	result, err := database.DB.Exec(`
		UPDATE repairs
		SET status = 'ส่งซ่อมภายนอก', technician_id = NULL, accepted_at = NULL,
		    completed_at = NULL, admin_note = $1
		WHERE id = $2 AND status = 'ซ่อมไม่ได้'`,
		"ส่งซ่อมภายนอก: "+strings.TrimSpace(req.Details), repairID)
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
		c.JSON(http.StatusConflict, gin.H{"error": "ส่งซ่อมภายนอกได้เฉพาะงานที่ช่างแจ้งว่าซ่อมไม่ได้"})
		return
	}

	insertRepairLog(repairID, req.AdminID, "OUTSOURCED", "ซ่อมไม่ได้", "ส่งซ่อมภายนอก", req.Details)
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
			<div class="box-info"><strong>สถานะ:</strong> ส่งซ่อมภายนอก</div>
			<div class="box-info"><strong>รายละเอียด:</strong><br>%s</div>
		`, html.EscapeString(ticketNumber), html.EscapeString(description), html.EscapeString(strings.TrimSpace(req.Details)))
		if err := utils.SendEmailNotification(
			[]string{reporterEmail},
			fmt.Sprintf("อัปเดตงานซ่อมเป็นส่งซ่อมภายนอก (%s)", ticketNumber),
			wrapEmail(colorInfo, "อัปเดตสถานะงานซ่อม", bodyContent),
		); err != nil {
			log.Printf("failed to email outsourced repair update for %s: %v", repairID, err)
		}
	}()
	c.JSON(http.StatusOK, gin.H{"message": "บันทึกการส่งซ่อมภายนอกสำเร็จ"})
}

// ---------------------------------------------------------
// 9. EstimateRepair: ประเมินราคา & เบิกจุดคุ้มทุน (🔥 อัปเกรดใส่ Log)
// ---------------------------------------------------------
func EstimateRepair(c *gin.Context) {
	repairID := c.Param("id")
	type EstimateRequest struct {
		EstimatedCost float64 `json:"estimated_cost" binding:"gte=0"`
		TechID        int     `json:"tech_id"`
	}

	var req EstimateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "กรุณาระบุราคาประเมินให้ถูกต้อง"})
		return
	}

	var currentStatus string
	if err := database.DB.QueryRow("SELECT status FROM repairs WHERE id = $1", repairID).Scan(&currentStatus); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบรายการแจ้งซ่อม"})
		return
	}

	var eqID *int
	var basePrice float64
	err := database.DB.QueryRow(`SELECT r.equipment_id, e.base_price FROM repairs r JOIN equipments e ON r.equipment_id = e.id WHERE r.id = $1`, repairID).Scan(&eqID, &basePrice)

	if err != nil && err != sql.ErrNoRows {
		log.Printf("failed to load equipment for repair estimate %s: %v", repairID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถตรวจสอบข้อมูลอุปกรณ์เพื่อประเมินราคาได้"})
		return
	}

	if err == sql.ErrNoRows || eqID == nil {
		result, err := database.DB.Exec(`UPDATE repairs SET estimated_cost = $1, status = 'กำลังซ่อม', accepted_at = CURRENT_TIMESTAMP WHERE id = $2`, req.EstimatedCost, repairID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถบันทึกราคาประเมินได้"})
			return
		}
		if rows, err := result.RowsAffected(); err != nil || rows == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบรายการแจ้งซ่อม"})
			return
		}
		insertRepairLog(repairID, req.TechID, "ESTIMATED", currentStatus, "กำลังซ่อม", fmt.Sprintf("ประเมินราคา: %.2f บาท (เริ่มซ่อมได้เลยไม่มีเงื่อนไข)", req.EstimatedCost))
		go sendEstimateEmail(repairID, "กำลังซ่อม", "")
		c.JSON(http.StatusOK, gin.H{"message": "บันทึกราคาประเมินสำเร็จ", "status": "กำลังซ่อม"})
		return
	}

	var accumulatedCost float64
	if err := database.DB.QueryRow(`SELECT COALESCE(SUM(actual_cost), 0) FROM repairs WHERE equipment_id = $1 AND status = 'เสร็จเรียบร้อย'`, *eqID).Scan(&accumulatedCost); err != nil {
		log.Printf("failed to calculate accumulated repair cost for repair %s: %v", repairID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถตรวจสอบต้นทุนสะสมของอุปกรณ์ได้"})
		return
	}

	newStatus := "กำลังซ่อม"
	adminSystemNote := ""

	isSingleExceed := req.EstimatedCost > (basePrice * 0.5)
	isAccumulatedExceed := (accumulatedCost + req.EstimatedCost) > (basePrice * 0.7)

	if isSingleExceed || isAccumulatedExceed {
		newStatus = "รอซ่อม"
		adminSystemNote = fmt.Sprintf("แจ้งเตือน: ราคาประเมินรวมเกินจุดคุ้มทุน (ฐาน %v บาท)", basePrice)
		if _, err := database.DB.Exec(`UPDATE repairs SET estimated_cost = $1, status = $2, admin_note = $3, accepted_at = NULL WHERE id = $4`, req.EstimatedCost, newStatus, adminSystemNote, repairID); err != nil {
			log.Printf("failed to save threshold repair estimate %s: %v", repairID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถบันทึกราคาประเมินได้"})
			return
		}
	} else {
		if _, err := database.DB.Exec(`UPDATE repairs SET estimated_cost = $1, status = $2, accepted_at = CURRENT_TIMESTAMP WHERE id = $3`, req.EstimatedCost, newStatus, repairID); err != nil {
			log.Printf("failed to save repair estimate %s: %v", repairID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถบันทึกราคาประเมินได้"})
			return
		}
	}

	// 🔥 บันทึก Log การประเมินราคา
	insertRepairLog(repairID, req.TechID, "ESTIMATED", currentStatus, newStatus, fmt.Sprintf("ประเมินราคา: %.2f บาท | หมายเหตุระบบ: %s", req.EstimatedCost, adminSystemNote))

	go sendEstimateEmail(repairID, newStatus, adminSystemNote)

	responseMsg := "ประเมินราคาสำเร็จและเริ่มซ่อมได้"
	if newStatus == "รอซ่อม" {
		responseMsg = "ประเมินราคาเกินจุดคุ้มทุน ส่งเรื่องกลับไปให้แอดมินพิจารณาแล้ว"
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

	if newStatus == "รอซ่อม" {
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
			body := wrapEmail(colorInfo, "ช่างได้ประเมินราคาและเริ่มดำเนินการซ่อมแล้ว", bodyContent)

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
	type ApproveReq struct {
		AdminID int `json:"admin_id"`
	}

	var req ApproveReq
	c.ShouldBindJSON(&req) // ดึง admin_id ถ้าหน้าเว็บส่งมา

	var currentStatus string
	if err := database.DB.QueryRow("SELECT status FROM repairs WHERE id = $1", repairID).Scan(&currentStatus); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบรายการแจ้งซ่อม"})
		return
	}

	_, err := database.DB.Exec(`
		UPDATE repairs 
		SET status = 'กำลังซ่อม', accepted_at = CURRENT_TIMESTAMP, admin_note = CONCAT(admin_note, ' -> (Admin อนุมัติให้ดำเนินการต่อ)')
		WHERE id = $1
	`, repairID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถอนุมัติงานได้: " + err.Error()})
		return
	}

	// 🔥 บันทึก Log การอนุมัติ
	insertRepairLog(repairID, req.AdminID, "APPROVED", currentStatus, "กำลังซ่อม", "Admin อนุมัติให้ซ่อมงานที่เกินจุดคุ้มทุนได้")

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
