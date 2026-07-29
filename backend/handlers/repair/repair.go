package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"backend/database"
)

// ---------------------------------------------------------
// Helper function สำหรับจัดการค่าว่างให้เป็น NULL ใน Database
// ---------------------------------------------------------
func nullIfEmpty(value string) interface{} {
	if value == "" || value == "null" || value == "undefined" {
		return nil
	}
	return value
}

// ---------------------------------------------------------
// 1. CreateRepair: สำหรับ User แจ้งซ่อมพร้อมแนบรูป
// ---------------------------------------------------------
func CreateRepair(c *gin.Context) {
	reporterEmail := c.PostForm("reporter_email")
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
		`INSERT INTO repairs (reporter_email, location_id, other_location, floor_id, room_id, equipment_id, problem_type_id, other_problem_type, description, status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, 'รอซ่อม') RETURNING id`,
		reporterEmail, locationID, otherLocation, floorID, roomID, equipmentID, problemTypeID, otherProblemType, description,
	).Scan(&repairID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถบันทึกข้อมูลได้: " + err.Error()})
		return
	}

	form, _ := c.MultipartForm()
	if form != nil && form.File != nil {
		files := form.File["image"]
		for _, file := range files {
			filename := fmt.Sprintf("%d_%d%s", repairID, time.Now().UnixNano(), filepath.Ext(file.Filename))
			path := "uploads/" + filename

			if err := c.SaveUploadedFile(file, path); err != nil {
				continue
			}

			imageURL := "/uploads/" + filename
			database.DB.Exec(
				`INSERT INTO repair_images (repair_id, image_url, image_type) VALUES ($1, $2, 'before')`,
				repairID, imageURL,
			)
		}
	}

	c.JSON(http.StatusCreated, gin.H{"message": "สร้างรายการแจ้งซ่อมสำเร็จ"})
}

// ---------------------------------------------------------
// 2. GetAllRepairs: ดึงรายการแจ้งซ่อมทั้งหมด (JOIN ข้อมูลใหม่)
// ---------------------------------------------------------
func GetAllRepairs(c *gin.Context) {
	// 🔥 [แก้ไข] เปลี่ยนจาก r.repair_cost เป็น r.estimated_cost และ r.actual_cost
	rows, err := database.DB.Query(`
		SELECT r.id, r.description, r.status, r.created_at, r.technician_id, r.technician_note, r.admin_note, r.reporter_email,
			   r.estimated_cost, r.actual_cost, r.accepted_at, r.completed_at, r.other_location, r.other_problem_type,
			   l.name as location_name, f.floor_name, ro.room_number,
			   eq.name as equipment_name, eq.asset_code, p.name as problem_type_name, u.full_name as technician_name
		FROM repairs r
		LEFT JOIN locations l ON r.location_id = l.id
		LEFT JOIN floors f ON r.floor_id = f.id
		LEFT JOIN rooms ro ON r.room_id = ro.id
		LEFT JOIN equipments eq ON r.equipment_id = eq.id
		LEFT JOIN problem_types p ON r.problem_type_id = p.id
		LEFT JOIN users u ON r.technician_id = u.id
		ORDER BY r.created_at DESC`)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	repairs := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id int
		var desc, status, reporterEmail string
		var locName, otherLoc, floorName, roomNumber, eqName, assetCode, probName, otherProb, techName, techNote, adminNote *string
		var estimatedCost, actualCost float64
		var createdAt time.Time
		var acceptedAt, completedAt *time.Time
		var techID *int

		err := rows.Scan(&id, &desc, &status, &createdAt, &techID, &techNote, &adminNote, &reporterEmail,
			&estimatedCost, &actualCost, &acceptedAt, &completedAt, &otherLoc, &otherProb,
			&locName, &floorName, &roomNumber, &eqName, &assetCode, &probName, &techName)
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

		r := map[string]interface{}{
			"id":                 id,
			"description":        desc,
			"status":             status,
			"created_at":         createdAt,
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
		}
		repairs = append(repairs, r)
	}

	c.JSON(http.StatusOK, repairs)
}

// ---------------------------------------------------------
// 3. GetRepairByID: ดูรายละเอียดงานซ่อมเฉพาะเคส
// ---------------------------------------------------------
func GetRepairByID(c *gin.Context) {
	repairID := c.Param("id")

	var id int
	var desc, status, reporterEmail string
	var locName, otherLoc, floorName, roomNumber, eqName, assetCode, probName, otherProb, techName, techNote, adminNote *string
	var estimatedCost, actualCost float64
	var createdAt time.Time
	var acceptedAt, completedAt *time.Time

	// 🔥 [แก้ไข] อัปเดต Query เป็น estimated_cost, actual_cost
	err := database.DB.QueryRow(`
		SELECT r.id, r.description, r.status, r.reporter_email, r.technician_note, r.admin_note, r.created_at,
			   r.estimated_cost, r.actual_cost, r.accepted_at, r.completed_at, r.other_location, r.other_problem_type,
			   l.name, f.floor_name, ro.room_number, eq.name, eq.asset_code, p.name, u.full_name
		FROM repairs r
		LEFT JOIN locations l ON r.location_id = l.id
		LEFT JOIN floors f ON r.floor_id = f.id
		LEFT JOIN rooms ro ON r.room_id = ro.id
		LEFT JOIN equipments eq ON r.equipment_id = eq.id
		LEFT JOIN problem_types p ON r.problem_type_id = p.id
		LEFT JOIN users u ON r.technician_id = u.id
		WHERE r.id = $1`, repairID).Scan(
		&id, &desc, &status, &reporterEmail, &techNote, &adminNote, &createdAt,
		&estimatedCost, &actualCost, &acceptedAt, &completedAt, &otherLoc, &otherProb,
		&locName, &floorName, &roomNumber, &eqName, &assetCode, &probName, &techName)

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
		"description":        desc,
		"status":             status,
		"reporter_email":     reporterEmail,
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
// 4. AssignRepair: Admin มอบหมายช่าง
// ---------------------------------------------------------
func AssignRepair(c *gin.Context) {
	repairID := c.Param("id")
	type AssignRequest struct {
		TechnicianID int `json:"technician_id" binding:"required"`
	}

	var req AssignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ข้อมูลช่างไม่ถูกต้อง"})
		return
	}

	_, err := database.DB.Exec(
		`UPDATE repairs SET status = 'รอซ่อม', technician_id = $1, accepted_at = NULL WHERE id = $2`,
		req.TechnicianID, repairID,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "มอบหมายงานให้ช่างสำเร็จ"})
}

// ---------------------------------------------------------
// 5. RevokeRepair: Admin ดึงงานกลับ
// ---------------------------------------------------------
func RevokeRepair(c *gin.Context) {
	repairID := c.Param("id")
	_, err := database.DB.Exec(`UPDATE repairs SET status = 'รอซ่อม', technician_id = NULL, accepted_at = NULL, completed_at = NULL WHERE id = $1`, repairID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ยกเลิกการจ่ายงานสำเร็จ"})
}

// ---------------------------------------------------------
// 6. RejectRepair: Tech ปฏิเสธงาน
// ---------------------------------------------------------
func RejectRepair(c *gin.Context) {
	repairID := c.Param("id")
	_, err := database.DB.Exec(`UPDATE repairs SET status = 'รอซ่อม', technician_id = NULL, accepted_at = NULL, completed_at = NULL WHERE id = $1`, repairID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ปฏิเสธงานและส่งคืนระบบสำเร็จ"})
}

// ---------------------------------------------------------
// 7. UpdateRepairStatus: Tech อัปเดตสถานะ (รับงาน / ปิดงาน)
// ---------------------------------------------------------
func UpdateRepairStatus(c *gin.Context) {
	repairID := c.Param("id")

	status := c.PostForm("status")
	actualCostStr := c.PostForm("actual_cost") // 🔥 [แก้ไข] เปลี่ยนจาก repair_cost เป็น actual_cost

	tNote := nullIfEmpty(c.PostForm("technician_note"))
	aNote := nullIfEmpty(c.PostForm("admin_note"))

	if status == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "กรุณาระบุสถานะ"})
		return
	}

	var actualCost float64
	if actualCostStr != "" && actualCostStr != "null" && actualCostStr != "undefined" {
		parsedCost, err := strconv.ParseFloat(actualCostStr, 64)
		if err == nil {
			actualCost = parsedCost
		}
	}

	var query string

	// 🔥 [แก้ไข] อัปเดตฟิลด์ actual_cost แทน repair_cost
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

	form, _ := c.MultipartForm()
	if form != nil && form.File != nil {
		files := form.File["image"]
		for _, file := range files {
			filename := fmt.Sprintf("%s_after_%d%s", repairID, time.Now().UnixNano(), filepath.Ext(file.Filename))
			path := "uploads/" + filename

			if err := c.SaveUploadedFile(file, path); err != nil {
				continue
			}

			imageURL := "/uploads/" + filename
			database.DB.Exec(
				`INSERT INTO repair_images (repair_id, image_url, image_type) VALUES ($1, $2, 'after')`,
				repairID, imageURL,
			)
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "อัปเดตสถานะสำเร็จ"})
}

// ---------------------------------------------------------
// 8. CancelRepairByAdmin: Admin ยกเลิกงาน (ไม่คุ้มทุน)
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

	_, err := database.DB.Exec(
		`UPDATE repairs SET status = 'ซ่อมไม่ได้', admin_note = $1, completed_at = CURRENT_TIMESTAMP WHERE id = $2`,
		req.AdminNote, repairID,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ยกเลิกงานสำเร็จ (สถานะ: ซ่อมไม่ได้)"})
}

// ---------------------------------------------------------
// 🔥 9. EstimateRepair: ช่างประเมินราคาและเช็กจุดคุ้มทุน (Hybrid Logic - 4 สถานะ)
// ---------------------------------------------------------
func EstimateRepair(c *gin.Context) {
	repairID := c.Param("id")

	// 🔥 [แก้ไขจุดนี้] เปลี่ยนจาก binding:"required" เป็น binding:"gte=0" เพื่อให้รับค่า 0 ได้
	type EstimateRequest struct {
		EstimatedCost float64 `json:"estimated_cost" binding:"gte=0"`
	}

	var req EstimateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "กรุณาระบุราคาประเมินให้ถูกต้อง"})
		return
	}

	// 1. ดึง Equipment ID และ Base Price ของงานซ่อมนี้
	var eqID *int
	var basePrice float64
	err := database.DB.QueryRow(`
		SELECT r.equipment_id, e.base_price
		FROM repairs r
		JOIN equipments e ON r.equipment_id = e.id
		WHERE r.id = $1
	`, repairID).Scan(&eqID, &basePrice)

	if err != nil || eqID == nil {
		// กรณีไม่มีข้อมูล Equipment ให้ผ่านไปเลยโดยไม่อิง Break-even
		database.DB.Exec(`UPDATE repairs SET estimated_cost = $1, status = 'กำลังซ่อม', accepted_at = CURRENT_TIMESTAMP WHERE id = $2`, req.EstimatedCost, repairID)
		c.JSON(http.StatusOK, gin.H{"message": "บันทึกราคาประเมินสำเร็จ", "status": "กำลังซ่อม"})
		return
	}

	// 2. คำนวณยอดซ่อมสะสม (Accumulated Cost) เฉพาะงานที่เสร็จแล้ว
	var accumulatedCost float64
	database.DB.QueryRow(`
		SELECT COALESCE(SUM(actual_cost), 0)
		FROM repairs
		WHERE equipment_id = $1 AND status = 'เสร็จเรียบร้อย'
	`, *eqID).Scan(&accumulatedCost)

	// 3. ลอจิกตรวจสอบเงื่อนไข Hybrid
	newStatus := "กำลังซ่อม"
	adminSystemNote := ""

	isSingleExceed := req.EstimatedCost > (basePrice * 0.5)
	isAccumulatedExceed := (accumulatedCost + req.EstimatedCost) > (basePrice * 0.7)

	if isSingleExceed || isAccumulatedExceed {
		// 🔥 [แก้ไขจุดสำคัญ] ตีกลับเป็น "รอซ่อม" เพื่อให้แอดมินตัดสินใจ โดยไม่เพิ่มสถานะใหม่
		newStatus = "รอซ่อม"

		if isSingleExceed && isAccumulatedExceed {
			adminSystemNote = fmt.Sprintf("ระบบระงับอัตโนมัติ: ประเมินครั้งนี้เกิน 50%% และยอดสะสมรวมเกิน 70%% ของราคาต้นทุน (%.2f บาท)", basePrice)
		} else if isSingleExceed {
			adminSystemNote = fmt.Sprintf("ระบบระงับอัตโนมัติ: ประเมินครั้งนี้เกิน 50%% ของราคาต้นทุน (%.2f บาท)", basePrice)
		} else {
			adminSystemNote = fmt.Sprintf("ระบบระงับอัตโนมัติ: ยอดสะสมรวมกับครั้งนี้เกิน 70%% ของราคาต้นทุน (%.2f บาท)", basePrice)
		}
	}

	// 4. บันทึกลง Database
	if newStatus == "รอซ่อม" {
		// ถ้าระบบระงับ จะอัปเดตราคาประเมิน ฝังโน้ตเตือนแอดมิน และเคลียร์เวลา accepted_at ออกชั่วคราว
		_, err = database.DB.Exec(`
			UPDATE repairs
			SET estimated_cost = $1, status = $2, admin_note = $3, accepted_at = NULL
			WHERE id = $4
		`, req.EstimatedCost, newStatus, adminSystemNote, repairID)
	} else {
		// ถ้าไม่เกินเกณฑ์ ให้ช่างเริ่มงานได้เลย (เปลี่ยนเป็น กำลังซ่อม)
		_, err = database.DB.Exec(`
			UPDATE repairs
			SET estimated_cost = $1, status = $2, accepted_at = CURRENT_TIMESTAMP
			WHERE id = $3
		`, req.EstimatedCost, newStatus, repairID)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "อัปเดตข้อมูลล้มเหลว: " + err.Error()})
		return
	}

	// ส่งข้อความกลับไปบอก Frontend ให้ช่างรู้ตัวด้วยว่าถูกระงับ
	responseMsg := "ประเมินราคาสำเร็จและเริ่มซ่อมได้"
	if newStatus == "รอซ่อม" {
		responseMsg = "ประเมินราคาเกินจุดคุ้มทุน ระบบได้ส่งเรื่องกลับไปให้แอดมินพิจารณาแล้ว"
	}

	c.JSON(http.StatusOK, gin.H{
		"message":          responseMsg,
		"status":           newStatus,
		"estimated_cost":   req.EstimatedCost,
		"accumulated_cost": accumulatedCost,
	})
}

// ---------------------------------------------------------
// 10. ApproveRepairThreshold: Admin อนุมัติงานที่ติดเงื่อนไข Break-even
// ---------------------------------------------------------
func ApproveRepairThreshold(c *gin.Context) {
	repairID := c.Param("id")

	// เปลี่ยนสถานะเป็น 'กำลังซ่อม' (ถือว่าเริ่มงานเลย)
	// ประทับเวลา accepted_at เป็นปัจจุบัน
	// และต่อท้าย admin_note ว่าอนุมัติแล้ว เพื่อเป็นประวัติ
	_, err := database.DB.Exec(`
		UPDATE repairs 
		SET status = 'กำลังซ่อม', 
		    accepted_at = CURRENT_TIMESTAMP,
		    admin_note = CONCAT(admin_note, ' -> (Admin อนุมัติให้ดำเนินการต่อ)')
		WHERE id = $1
	`, repairID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถอนุมัติงานได้: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "อนุมัติงานซ่อมสำเร็จ"})
}
