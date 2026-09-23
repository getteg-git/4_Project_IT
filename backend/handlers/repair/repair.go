package repair

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"backend/database"
	"backend/utils" // ใช้ utils.AdminEmail, utils.TechEmail, utils.SendEmailNotification
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

// =========================================================
// ธีมสีกลาง สำหรับอีเมลทุกฉบับในระบบ
// =========================================================
const (
	colorSuccess = "#16a34a" // เขียว: สำเร็จ / ยืนยัน
	colorInfo    = "#2563eb" // ฟ้า: แจ้งเตือนทั่วไป / ข้อมูล
	colorWarning = "#d97706" // ส้ม: รอดำเนินการ / ต้องพิจารณา
	colorDanger  = "#dc2626" // แดง: ปัญหา / ยกเลิก / ปฏิเสธ
)

func emailHTMLStyle() string {
	return `
		body { font-family: 'Sarabun', Arial, sans-serif; background-color: #f4f6f9; margin: 0; padding: 15px; }
		.card { max-width: 480px; background-color: #ffffff; margin: 0 auto; border-radius: 10px; padding: 25px; box-shadow: 0 2px 8px rgba(0,0,0,0.08); }
		.box-success { background-color: #dcfce7; color: #166534; border-left: 4px solid #16a34a; padding: 12px; margin: 15px 0; border-radius: 4px; }
		.box-info { background-color: #dbeafe; color: #1e40af; border-left: 4px solid #2563eb; padding: 12px; margin: 15px 0; border-radius: 4px; }
		.box-warning { background-color: #fef3c7; color: #92400e; border-left: 4px solid #d97706; padding: 12px; margin: 15px 0; border-radius: 4px; }
		.box-danger { background-color: #fee2e2; color: #991b1b; border-left: 4px solid #dc2626; padding: 12px; margin: 15px 0; border-radius: 4px; }
		.btn { display: inline-block; background-color: #10a7ff; color: #ffffff; text-decoration: none; padding: 10px 18px; border-radius: 6px; margin-top: 15px; font-weight: 600; }
		.btn-success { background-color: #16a34a; }
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
		formURL := "http://localhost:5173/repair/history"

		// 1. อีเมลสำหรับผู้แจ้ง
		userBodyContent := fmt.Sprintf(`
			<p><strong>หมายเลขอ้างอิง:</strong> %s</p>
			<p><strong>รายละเอียด:</strong> %s</p>
			<div class="box-warning">ขณะนี้อยู่ในสถานะ: <strong>รอซ่อม</strong></div>
			<p>ทางเราจะแจ้งเตือนอีกครั้งเมื่อช่างเริ่มเข้าดำเนินการครับ</p>
			<a href="%s" class="btn">ตรวจสอบรายการแจ้งซ่อมได้ที่นี่</a>
		`, ticketNumber, description, formURL)
		userSubject := fmt.Sprintf("✅ รับเรื่องแจ้งซ่อมเรียบร้อยแล้ว (%s)", ticketNumber)
		userBody := wrapEmail(colorSuccess, "ระบบได้รับเรื่องแจ้งซ่อมของคุณเรียบร้อยแล้ว", userBodyContent)

		if reporterEmail != "" {
			utils.SendEmailNotification([]string{reporterEmail}, userSubject, userBody)
		}

		// 2. อีเมลสำหรับ Admin
		AdminLink := "http://localhost:5173"
		adminBodyContent := fmt.Sprintf(`
			<p>กรุณาเข้าสู่ระบบเพื่อพิจารณามอบหมายช่างดำเนินการ</p>
			<p><strong>หมายเลขอ้างอิง:</strong> %s</p>
			<p><strong>ผู้แจ้ง:</strong> %s</p>
			<p><strong>รายละเอียด:</strong> %s</p>
			<div class="box-warning">ขณะนี้อยู่ในสถานะ: <strong>รอซ่อม</strong></div>
			<a href="%s" class="btn">ตรวจสอบรายการได้ที่นี่</a>
		`, ticketNumber, reporterEmail, description, AdminLink)
		adminSubject := fmt.Sprintf("🔔 มีงานแจ้งซ่อมใหม่เข้ามา (%s)", ticketNumber)
		adminBody := wrapEmail(colorInfo, "มีรายการแจ้งซ่อมใหม่เข้ามาในระบบ", adminBodyContent)

		utils.SendEmailNotification([]string{utils.AdminEmail}, adminSubject, adminBody)
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

	var currentStatus string
	database.DB.QueryRow("SELECT status FROM repairs WHERE id = $1", repairID).Scan(&currentStatus)

	_, err := database.DB.Exec(
		`UPDATE repairs SET status = 'รอซ่อม', technician_id = $1, accepted_at = NULL WHERE id = $2`,
		req.TechnicianID, repairID,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 🔥 บันทึก Log การมอบหมายงาน
	insertRepairLog(repairID, req.AdminID, "ASSIGNED", currentStatus, "รอซ่อม", fmt.Sprintf("Admin มอบหมายงานให้ช่าง ID: %d", req.TechnicianID))

	// ---------- ส่งอีเมลแจ้งเตือน ----------
	go func() {
		var reporterEmail, description, tNumber string
		var techName *string
		database.DB.QueryRow(`
			SELECT r.reporter_email, r.description, r.ticket_number, u.full_name
			FROM repairs r LEFT JOIN users u ON u.id = $1 WHERE r.id = $2`, req.TechnicianID, repairID).Scan(&reporterEmail, &description, &tNumber, &techName)

		safeTechName := ""
		if techName != nil {
			safeTechName = *techName
		}

		// แจ้งผู้แจ้ง
		if reporterEmail != "" {
			userBodyContent := fmt.Sprintf(`
				<p><strong>รายการ:</strong> %s</p>
				<div class="box-info">ช่างผู้รับผิดชอบ: <strong>%s</strong></div>
				<p>ช่างจะเข้าประเมินและดำเนินการต่อไปครับ</p>
			`, description, safeTechName)
			utils.SendEmailNotification([]string{reporterEmail}, fmt.Sprintf("⚙️ มอบหมายช่างซ่อมแล้ว (%s)", tNumber), wrapEmail(colorInfo, "มอบหมายช่างเข้าดูแลงานแล้ว", userBodyContent))
		}

		// แจ้งช่าง
		if utils.TechEmail != "" {
			techBodyContent := fmt.Sprintf(`
				<p><strong>รหัสอ้างอิง:</strong> %s</p>
				<p><strong>รายละเอียดงาน:</strong> %s</p>
				<a href="http://localhost:5173" class="btn">ตรวจสอบรายการมอบหมายงานได้ที่นี่</a>
			`, tNumber, description)
			utils.SendEmailNotification([]string{utils.TechEmail}, fmt.Sprintf("🔧 คุณได้รับมอบหมายงานซ่อมใหม่ (%s)", tNumber), wrapEmail(colorInfo, "คุณได้รับมอบหมายงานซ่อมใหม่", techBodyContent))
		}
	}()

	c.JSON(http.StatusOK, gin.H{"message": "มอบหมายงานให้ช่างสำเร็จ"})
}

// ---------------------------------------------------------
// 5. RevokeRepair: Admin ดึงงานกลับ (🔥 อัปเกรดใส่ Log)
// ---------------------------------------------------------
func RevokeRepair(c *gin.Context) {
	repairID := c.Param("id")
	revokeReason := c.PostForm("rejection_reason")
	adminID := nullIfEmpty(c.PostForm("admin_id"))

	var techName, description, tNumber, currentStatus string
	database.DB.QueryRow(`
		SELECT u.full_name, r.description, r.ticket_number, r.status
		FROM repairs r JOIN users u ON r.technician_id = u.id WHERE r.id = $1`, repairID).Scan(&techName, &description, &tNumber, &currentStatus)

	_, err := database.DB.Exec(`UPDATE repairs SET status = 'รอซ่อม', technician_id = NULL, accepted_at = NULL, completed_at = NULL WHERE id = $1`, repairID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 🔥 บันทึก Log ดึงงานกลับ
	insertRepairLog(repairID, adminID, "REVOKED", currentStatus, "รอซ่อม", "Admin ดึงงานกลับจากช่าง. เหตุผล: "+revokeReason)

	go func() {
		if utils.TechEmail != "" {
			bodyContent := fmt.Sprintf(`
				<p>เรียนคุณ %s</p>
				<p>ระบบขอแจ้งให้ทราบว่ารายการแจ้งซ่อมที่คุณรับผิดชอบ ได้ถูกดึงงานกลับ/ยกเลิกการมอบหมายแล้วครับ</p>
				<p><strong>รหัสตั๋ว:</strong> %s</p>
				<p><strong>รายละเอียดงาน:</strong> %s</p>
				<div class="box-warning"><strong>📌 เหตุผล:</strong><br>%s</div>
			`, techName, tNumber, description, revokeReason)
			utils.SendEmailNotification([]string{utils.TechEmail}, fmt.Sprintf("⚠️ แจ้งเตือนการดึงงานซ่อมกลับ (%s)", tNumber), wrapEmail(colorDanger, "แจ้งเตือนการดึงงานซ่อมกลับ", bodyContent))
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
		`, tNumber, description, rejectionReason)

		recipients := []string{utils.AdminEmail}
		if reporterEmail != "" {
			recipients = append(recipients, reporterEmail)
		}
		utils.SendEmailNotification(recipients, fmt.Sprintf("⚠️ ช่างปฏิเสธงานซ่อม (%s)", tNumber), wrapEmail(colorDanger, "ช่างปฏิเสธงาน / ส่งกลับเข้าระบบ", bodyContent))
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
		var reporterEmail, description, tNumber string
		database.DB.QueryRow(`SELECT reporter_email, description, ticket_number FROM repairs WHERE id = $1`, repairID).Scan(&reporterEmail, &description, &tNumber)

		if status == "เสร็จเรียบร้อย" {
			if reporterEmail != "" {
				userBodyContent := fmt.Sprintf(`
					<p><strong>รายการ:</strong> %s</p>
					<div class="box-success"><strong>บันทึกจากช่าง:</strong><br>%s</div>
					<a href="https://rb.gy/bjiuqq" class="btn btn-success">ให้คะแนนความพึงพอใจ</a>
				`, description, techNoteStr)
				utils.SendEmailNotification([]string{reporterEmail}, fmt.Sprintf("🎉 งานซ่อม %s เรียบร้อยแล้ว", tNumber), wrapEmail(colorSuccess, "🎉 งานซ่อมของคุณเสร็จเรียบร้อยแล้ว", userBodyContent))
			}
		} else if status == "ซ่อมไม่ได้" {
			bodyContent := fmt.Sprintf(`
				<p><strong>รายการ:</strong> %s</p>
				<div class="box-danger"><strong>เหตุผล:</strong><br>%s</div>
			`, description, techNoteStr)
			recipients := []string{utils.AdminEmail, utils.TechEmail}
			if reporterEmail != "" {
				recipients = append(recipients, reporterEmail)
			}
			utils.SendEmailNotification(recipients, fmt.Sprintf("⚠️ รายงานปัญหาการซ่อม (%s)", tNumber), wrapEmail(colorDanger, "ช่างไม่สามารถดำเนินการซ่อมได้", bodyContent))
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

	c.JSON(http.StatusOK, gin.H{"message": "ยกเลิกงานสำเร็จ (สถานะ: ซ่อมไม่ได้)"})
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
	database.DB.QueryRow("SELECT status FROM repairs WHERE id = $1", repairID).Scan(&currentStatus)

	var eqID *int
	var basePrice float64
	err := database.DB.QueryRow(`SELECT r.equipment_id, e.base_price FROM repairs r JOIN equipments e ON r.equipment_id = e.id WHERE r.id = $1`, repairID).Scan(&eqID, &basePrice)

	if err != nil || eqID == nil {
		database.DB.Exec(`UPDATE repairs SET estimated_cost = $1, status = 'กำลังซ่อม', accepted_at = CURRENT_TIMESTAMP WHERE id = $2`, req.EstimatedCost, repairID)
		insertRepairLog(repairID, req.TechID, "ESTIMATED", currentStatus, "กำลังซ่อม", fmt.Sprintf("ประเมินราคา: %.2f บาท (เริ่มซ่อมได้เลยไม่มีเงื่อนไข)", req.EstimatedCost))
		c.JSON(http.StatusOK, gin.H{"message": "บันทึกราคาประเมินสำเร็จ", "status": "กำลังซ่อม"})
		return
	}

	var accumulatedCost float64
	database.DB.QueryRow(`SELECT COALESCE(SUM(actual_cost), 0) FROM repairs WHERE equipment_id = $1 AND status = 'เสร็จเรียบร้อย'`, *eqID).Scan(&accumulatedCost)

	newStatus := "กำลังซ่อม"
	adminSystemNote := ""

	isSingleExceed := req.EstimatedCost > (basePrice * 0.5)
	isAccumulatedExceed := (accumulatedCost + req.EstimatedCost) > (basePrice * 0.7)

	if isSingleExceed || isAccumulatedExceed {
		newStatus = "รอซ่อม"
		adminSystemNote = fmt.Sprintf("แจ้งเตือน: ราคาประเมินรวมเกินจุดคุ้มทุน (ฐาน %v บาท)", basePrice)
		database.DB.Exec(`UPDATE repairs SET estimated_cost = $1, status = $2, admin_note = $3, accepted_at = NULL WHERE id = $4`, req.EstimatedCost, newStatus, adminSystemNote, repairID)
	} else {
		database.DB.Exec(`UPDATE repairs SET estimated_cost = $1, status = $2, accepted_at = CURRENT_TIMESTAMP WHERE id = $3`, req.EstimatedCost, newStatus, repairID)
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
	var reporterEmail, description string
	err := database.DB.QueryRow(`SELECT reporter_email, description FROM repairs WHERE id = $1`, repairID).Scan(&reporterEmail, &description)
	if err != nil {
		fmt.Println("❌ sendEstimateEmail query error:", err)
		return
	}

	if newStatus == "รอซ่อม" {
		// 🔴 เกินจุดคุ้มทุน -> แจ้ง Admin ให้เข้ามาพิจารณาอนุมัติ
		bodyContent := fmt.Sprintf(`
			<p><strong>หมายเลขใบแจ้ง:</strong> #%s</p>
			<p><strong>รายการ:</strong> %s</p>
			<div class="box-warning"><strong>หมายเหตุระบบ:</strong><br>%s</div>
			<p>กรุณาเข้าสู่ระบบเพื่อพิจารณาอนุมัติหรือยกเลิกงานนี้ครับ</p>
		`, repairID, description, adminSystemNote)
		subject := fmt.Sprintf("🔔 งานซ่อมรอการอนุมัติราคา (ใบงาน #%s)", repairID)
		body := wrapEmail(colorWarning, "งานซ่อมนี้เกินจุดคุ้มทุน รอ Admin พิจารณา", bodyContent)

		if err := utils.SendEmailNotification([]string{utils.AdminEmail}, subject, body); err != nil {
			fmt.Println("❌ ส่งอีเมล sendEstimateEmail (รอซ่อม) ไม่สำเร็จ:", err)
		}
	} else {
		// 🟢 ไม่เกินจุดคุ้มทุน -> แจ้งผู้แจ้ง + แจ้งกลุ่มภายใน (Admin + ช่าง)

		// 1. ส่งหาผู้แจ้งซ่อม
		if reporterEmail != "" {
			bodyContent := fmt.Sprintf(`
				<p><strong>หมายเลขใบแจ้ง:</strong> #%s</p>
				<p><strong>รายการ:</strong> %s</p>
			`, repairID, description)
			subject := fmt.Sprintf("🔧 ช่างเริ่มดำเนินการซ่อมแล้ว (ใบงาน #%s)", repairID)
			body := wrapEmail(colorInfo, "ช่างได้ประเมินราคาและเริ่มดำเนินการซ่อมแล้ว", bodyContent)

			if err := utils.SendEmailNotification([]string{reporterEmail}, subject, body); err != nil {
				fmt.Println("❌ ส่งอีเมลผู้แจ้ง (กำลังซ่อม) ไม่สำเร็จ:", err)
			}
		}

		// 2. ✨ [เพิ่มใหม่] ส่งหา Admin + ช่าง ยืนยันการรับงานและเริ่มซ่อม
		internalRecipients := []string{utils.AdminEmail, utils.TechEmail}
		internalBodyContent := fmt.Sprintf(`
			<p><strong>หมายเลขใบแจ้ง:</strong> #%s</p>
			<p><strong>รายการ:</strong> %s</p>
			<p>สถานะปัจจุบัน: <b>กำลังซ่อม</b> (ราคาประเมินผ่านเกณฑ์จุดคุ้มทุนเรียบร้อยแล้ว)</p>
		`, repairID, description)
		internalSubject := fmt.Sprintf("⚙️ ยืนยันการรับงานและเริ่มซ่อม (ใบงาน #%s)", repairID)
		internalBody := wrapEmail(colorInfo, "ช่างรับงานและเริ่มดำเนินการซ่อมแล้ว", internalBodyContent)

		if err := utils.SendEmailNotification(internalRecipients, internalSubject, internalBody); err != nil {
			fmt.Println("❌ ส่งอีเมล admin+ช่าง (เริ่มซ่อม) ไม่สำเร็จ:", err)
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
	database.DB.QueryRow("SELECT status FROM repairs WHERE id = $1", repairID).Scan(&currentStatus)

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
		var description string
		database.DB.QueryRow(`SELECT description FROM repairs WHERE id = $1`, repairID).Scan(&description)
		if utils.TechEmail != "" {
			bodyContent := fmt.Sprintf(`<p><strong>รายการ:</strong> %s</p><p>กรุณาเข้าดำเนินการซ่อมต่อได้ทันทีครับ</p>`, description)
			utils.SendEmailNotification([]string{utils.TechEmail}, fmt.Sprintf("✅ Admin อนุมัติงานซ่อมแล้ว (ใบงาน #%s)", repairID), wrapEmail(colorSuccess, "Admin อนุมัติให้ดำเนินการซ่อมต่อได้แล้ว", bodyContent))
		}
	}()

	c.JSON(http.StatusOK, gin.H{"message": "อนุมัติงานซ่อมสำเร็จ"})
}
