// package handlers

// import (
// 	"fmt"
// 	"net/http"
// 	"path/filepath"
// 	"strconv"
// 	"time"

// 	"github.com/gin-gonic/gin"

// 	"backend/database"
// )

// // ---------------------------------------------------------
// // Helper function สำหรับจัดการค่าว่างให้เป็น NULL ใน Database
// // ---------------------------------------------------------
// func nullIfEmpty(value string) interface{} {
// 	if value == "" || value == "null" || value == "undefined" {
// 		return nil
// 	}
// 	return value
// }

// // ---------------------------------------------------------
// // 1. CreateRepair: สำหรับ User แจ้งซ่อมพร้อมแนบรูป
// // ---------------------------------------------------------
// func CreateRepair(c *gin.Context) {
// 	reporterEmail := c.PostForm("reporter_email")
// 	locationID := nullIfEmpty(c.PostForm("location_id"))
// 	otherLocation := nullIfEmpty(c.PostForm("other_location"))
// 	floorID := nullIfEmpty(c.PostForm("floor_id"))
// 	roomID := nullIfEmpty(c.PostForm("room_id"))
// 	equipmentID := nullIfEmpty(c.PostForm("equipment_id"))
// 	problemTypeID := nullIfEmpty(c.PostForm("problem_type_id"))
// 	otherProblemType := nullIfEmpty(c.PostForm("other_problem_type"))
// 	description := c.PostForm("description")

// 	var repairID int
// 	err := database.DB.QueryRow(
// 		`INSERT INTO repairs (reporter_email, location_id, other_location, floor_id, room_id, equipment_id, problem_type_id, other_problem_type, description, status)
// 		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, 'รอซ่อม') RETURNING id`,
// 		reporterEmail, locationID, otherLocation, floorID, roomID, equipmentID, problemTypeID, otherProblemType, description,
// 	).Scan(&repairID)

// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถบันทึกข้อมูลได้: " + err.Error()})
// 		return
// 	}

// 	form, _ := c.MultipartForm()
// 	if form != nil && form.File != nil {
// 		files := form.File["image"]
// 		for _, file := range files {
// 			filename := fmt.Sprintf("%d_%d%s", repairID, time.Now().UnixNano(), filepath.Ext(file.Filename))
// 			path := "uploads/" + filename

// 			if err := c.SaveUploadedFile(file, path); err != nil {
// 				continue
// 			}

// 			imageURL := "/uploads/" + filename
// 			database.DB.Exec(
// 				`INSERT INTO repair_images (repair_id, image_url, image_type) VALUES ($1, $2, 'before')`,
// 				repairID, imageURL,
// 			)
// 		}
// 	}

// 	c.JSON(http.StatusCreated, gin.H{"message": "สร้างรายการแจ้งซ่อมสำเร็จ"})
// }

// // ---------------------------------------------------------
// // 2. GetAllRepairs: ดึงรายการแจ้งซ่อมทั้งหมด (JOIN ข้อมูลใหม่)
// // ---------------------------------------------------------
// func GetAllRepairs(c *gin.Context) {
// 	// 🔥 [แก้ไข] เปลี่ยนจาก r.repair_cost เป็น r.estimated_cost และ r.actual_cost
// 	rows, err := database.DB.Query(`
// 		SELECT r.id, r.description, r.status, r.created_at, r.technician_id, r.technician_note, r.admin_note, r.reporter_email,
// 			   r.estimated_cost, r.actual_cost, r.accepted_at, r.completed_at, r.other_location, r.other_problem_type,
// 			   l.name as location_name, f.floor_name, ro.room_number,
// 			   eq.name as equipment_name, eq.asset_code, p.name as problem_type_name, u.full_name as technician_name
// 		FROM repairs r
// 		LEFT JOIN locations l ON r.location_id = l.id
// 		LEFT JOIN floors f ON r.floor_id = f.id
// 		LEFT JOIN rooms ro ON r.room_id = ro.id
// 		LEFT JOIN equipments eq ON r.equipment_id = eq.id
// 		LEFT JOIN problem_types p ON r.problem_type_id = p.id
// 		LEFT JOIN users u ON r.technician_id = u.id
// 		ORDER BY r.created_at DESC`)

// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}
// 	defer rows.Close()

// 	repairs := make([]map[string]interface{}, 0)
// 	for rows.Next() {
// 		var id int
// 		var desc, status, reporterEmail string
// 		var locName, otherLoc, floorName, roomNumber, eqName, assetCode, probName, otherProb, techName, techNote, adminNote *string
// 		var estimatedCost, actualCost float64
// 		var createdAt time.Time
// 		var acceptedAt, completedAt *time.Time
// 		var techID *int

// 		err := rows.Scan(&id, &desc, &status, &createdAt, &techID, &techNote, &adminNote, &reporterEmail,
// 			&estimatedCost, &actualCost, &acceptedAt, &completedAt, &otherLoc, &otherProb,
// 			&locName, &floorName, &roomNumber, &eqName, &assetCode, &probName, &techName)
// 		if err != nil {
// 			continue
// 		}

// 		imgRows, _ := database.DB.Query("SELECT image_url, image_type FROM repair_images WHERE repair_id = $1", id)
// 		images := make([]map[string]string, 0)
// 		for imgRows.Next() {
// 			var url, imgType string
// 			imgRows.Scan(&url, &imgType)
// 			images = append(images, map[string]string{"url": url, "type": imgType})
// 		}
// 		imgRows.Close()

// 		getString := func(s *string) string {
// 			if s == nil {
// 				return ""
// 			}
// 			return *s
// 		}

// 		r := map[string]interface{}{
// 			"id":                 id,
// 			"description":        desc,
// 			"status":             status,
// 			"created_at":         createdAt,
// 			"location_name":      getString(locName),
// 			"other_location":     getString(otherLoc),
// 			"floor_name":         getString(floorName),
// 			"room_number":        getString(roomNumber),
// 			"equipment_name":     getString(eqName),
// 			"asset_code":         getString(assetCode),
// 			"problem_type":       getString(probName),
// 			"other_problem_type": getString(otherProb),
// 			"reporter_email":     reporterEmail,
// 			"technician_id":      techID,
// 			"technician_note":    getString(techNote),
// 			"admin_note":         getString(adminNote),
// 			"technician_name":    getString(techName),
// 			"estimated_cost":     estimatedCost,
// 			"actual_cost":        actualCost,
// 			"accepted_at":        acceptedAt,
// 			"completed_at":       completedAt,
// 			"images":             images,
// 		}
// 		repairs = append(repairs, r)
// 	}

// 	c.JSON(http.StatusOK, repairs)
// }

// // ---------------------------------------------------------
// // 3. GetRepairByID: ดูรายละเอียดงานซ่อมเฉพาะเคส
// // ---------------------------------------------------------
// func GetRepairByID(c *gin.Context) {
// 	repairID := c.Param("id")

// 	var id int
// 	var desc, status, reporterEmail string
// 	var locName, otherLoc, floorName, roomNumber, eqName, assetCode, probName, otherProb, techName, techNote, adminNote *string
// 	var estimatedCost, actualCost float64
// 	var createdAt time.Time
// 	var acceptedAt, completedAt *time.Time

// 	// 🔥 [แก้ไข] อัปเดต Query เป็น estimated_cost, actual_cost
// 	err := database.DB.QueryRow(`
// 		SELECT r.id, r.description, r.status, r.reporter_email, r.technician_note, r.admin_note, r.created_at,
// 			   r.estimated_cost, r.actual_cost, r.accepted_at, r.completed_at, r.other_location, r.other_problem_type,
// 			   l.name, f.floor_name, ro.room_number, eq.name, eq.asset_code, p.name, u.full_name
// 		FROM repairs r
// 		LEFT JOIN locations l ON r.location_id = l.id
// 		LEFT JOIN floors f ON r.floor_id = f.id
// 		LEFT JOIN rooms ro ON r.room_id = ro.id
// 		LEFT JOIN equipments eq ON r.equipment_id = eq.id
// 		LEFT JOIN problem_types p ON r.problem_type_id = p.id
// 		LEFT JOIN users u ON r.technician_id = u.id
// 		WHERE r.id = $1`, repairID).Scan(
// 		&id, &desc, &status, &reporterEmail, &techNote, &adminNote, &createdAt,
// 		&estimatedCost, &actualCost, &acceptedAt, &completedAt, &otherLoc, &otherProb,
// 		&locName, &floorName, &roomNumber, &eqName, &assetCode, &probName, &techName)

// 	if err != nil {
// 		c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบรายการแจ้งซ่อม"})
// 		return
// 	}

// 	rows, _ := database.DB.Query("SELECT image_url, image_type FROM repair_images WHERE repair_id = $1", id)
// 	defer rows.Close()

// 	images := make([]map[string]string, 0)
// 	for rows.Next() {
// 		var url, imgType string
// 		rows.Scan(&url, &imgType)
// 		images = append(images, map[string]string{"url": url, "type": imgType})
// 	}

// 	getString := func(s *string) string {
// 		if s == nil {
// 			return ""
// 		}
// 		return *s
// 	}

// 	repair := map[string]interface{}{
// 		"id":                 id,
// 		"description":        desc,
// 		"status":             status,
// 		"reporter_email":     reporterEmail,
// 		"technician_note":    getString(techNote),
// 		"admin_note":         getString(adminNote),
// 		"technician_name":    getString(techName),
// 		"location_name":      getString(locName),
// 		"other_location":     getString(otherLoc),
// 		"floor_name":         getString(floorName),
// 		"room_number":        getString(roomNumber),
// 		"equipment_name":     getString(eqName),
// 		"asset_code":         getString(assetCode),
// 		"problem_type":       getString(probName),
// 		"other_problem_type": getString(otherProb),
// 		"estimated_cost":     estimatedCost,
// 		"actual_cost":        actualCost,
// 		"accepted_at":        acceptedAt,
// 		"completed_at":       completedAt,
// 		"created_at":         createdAt,
// 		"images":             images,
// 	}

// 	c.JSON(http.StatusOK, repair)
// }

// // ---------------------------------------------------------
// // 4. AssignRepair: Admin มอบหมายช่าง
// // ---------------------------------------------------------
// func AssignRepair(c *gin.Context) {
// 	repairID := c.Param("id")
// 	type AssignRequest struct {
// 		TechnicianID int `json:"technician_id" binding:"required"`
// 	}

// 	var req AssignRequest
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "ข้อมูลช่างไม่ถูกต้อง"})
// 		return
// 	}

// 	_, err := database.DB.Exec(
// 		`UPDATE repairs SET status = 'รอซ่อม', technician_id = $1, accepted_at = NULL WHERE id = $2`,
// 		req.TechnicianID, repairID,
// 	)

// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}
// 	c.JSON(http.StatusOK, gin.H{"message": "มอบหมายงานให้ช่างสำเร็จ"})
// }

// // ---------------------------------------------------------
// // 5. RevokeRepair: Admin ดึงงานกลับ
// // ---------------------------------------------------------
// func RevokeRepair(c *gin.Context) {
// 	repairID := c.Param("id")
// 	_, err := database.DB.Exec(`UPDATE repairs SET status = 'รอซ่อม', technician_id = NULL, accepted_at = NULL, completed_at = NULL WHERE id = $1`, repairID)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}
// 	c.JSON(http.StatusOK, gin.H{"message": "ยกเลิกการจ่ายงานสำเร็จ"})
// }

// // ---------------------------------------------------------
// // 6. RejectRepair: Tech ปฏิเสธงาน
// // ---------------------------------------------------------
// func RejectRepair(c *gin.Context) {
// 	repairID := c.Param("id")
// 	_, err := database.DB.Exec(`UPDATE repairs SET status = 'รอซ่อม', technician_id = NULL, accepted_at = NULL, completed_at = NULL WHERE id = $1`, repairID)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}
// 	c.JSON(http.StatusOK, gin.H{"message": "ปฏิเสธงานและส่งคืนระบบสำเร็จ"})
// }

// // ---------------------------------------------------------
// // 7. UpdateRepairStatus: Tech อัปเดตสถานะ (รับงาน / ปิดงาน)
// // ---------------------------------------------------------
// func UpdateRepairStatus(c *gin.Context) {
// 	repairID := c.Param("id")

// 	status := c.PostForm("status")
// 	actualCostStr := c.PostForm("actual_cost") // 🔥 [แก้ไข] เปลี่ยนจาก repair_cost เป็น actual_cost

// 	tNote := nullIfEmpty(c.PostForm("technician_note"))
// 	aNote := nullIfEmpty(c.PostForm("admin_note"))

// 	if status == "" {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "กรุณาระบุสถานะ"})
// 		return
// 	}

// 	var actualCost float64
// 	if actualCostStr != "" && actualCostStr != "null" && actualCostStr != "undefined" {
// 		parsedCost, err := strconv.ParseFloat(actualCostStr, 64)
// 		if err == nil {
// 			actualCost = parsedCost
// 		}
// 	}

// 	var query string

// 	// 🔥 [แก้ไข] อัปเดตฟิลด์ actual_cost แทน repair_cost
// 	if status == "กำลังซ่อม" {
// 		query = `UPDATE repairs SET status = $1, technician_note = COALESCE($2, technician_note), admin_note = COALESCE($3, admin_note), actual_cost = $4, accepted_at = CURRENT_TIMESTAMP WHERE id = $5`
// 	} else if status == "เสร็จเรียบร้อย" || status == "ซ่อมไม่ได้" {
// 		query = `UPDATE repairs SET status = $1, technician_note = COALESCE($2, technician_note), admin_note = COALESCE($3, admin_note), actual_cost = $4, completed_at = CURRENT_TIMESTAMP WHERE id = $5`
// 	} else {
// 		query = `UPDATE repairs SET status = $1, technician_note = COALESCE($2, technician_note), admin_note = COALESCE($3, admin_note), actual_cost = $4 WHERE id = $5`
// 	}

// 	_, err := database.DB.Exec(query, status, tNote, aNote, actualCost, repairID)

// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	form, _ := c.MultipartForm()
// 	if form != nil && form.File != nil {
// 		files := form.File["image"]
// 		for _, file := range files {
// 			filename := fmt.Sprintf("%s_after_%d%s", repairID, time.Now().UnixNano(), filepath.Ext(file.Filename))
// 			path := "uploads/" + filename

// 			if err := c.SaveUploadedFile(file, path); err != nil {
// 				continue
// 			}

// 			imageURL := "/uploads/" + filename
// 			database.DB.Exec(
// 				`INSERT INTO repair_images (repair_id, image_url, image_type) VALUES ($1, $2, 'after')`,
// 				repairID, imageURL,
// 			)
// 		}
// 	}

// 	c.JSON(http.StatusOK, gin.H{"message": "อัปเดตสถานะสำเร็จ"})
// }

// // ---------------------------------------------------------
// // 8. CancelRepairByAdmin: Admin ยกเลิกงาน (ไม่คุ้มทุน)
// // ---------------------------------------------------------
// func CancelRepairByAdmin(c *gin.Context) {
// 	repairID := c.Param("id")
// 	type CancelRequest struct {
// 		AdminNote string `json:"admin_note" binding:"required"`
// 	}

// 	var req CancelRequest
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "กรุณาระบุเหตุผลที่ยกเลิกงาน"})
// 		return
// 	}

// 	_, err := database.DB.Exec(
// 		`UPDATE repairs SET status = 'ซ่อมไม่ได้', admin_note = $1, completed_at = CURRENT_TIMESTAMP WHERE id = $2`,
// 		req.AdminNote, repairID,
// 	)

// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{"message": "ยกเลิกงานสำเร็จ (สถานะ: ซ่อมไม่ได้)"})
// }

// // ---------------------------------------------------------
// // 🔥 9. EstimateRepair: ช่างประเมินราคาและเช็กจุดคุ้มทุน (Hybrid Logic - 4 สถานะ)
// // ---------------------------------------------------------
// func EstimateRepair(c *gin.Context) {
// 	repairID := c.Param("id")

// 	// 🔥 [แก้ไขจุดนี้] เปลี่ยนจาก binding:"required" เป็น binding:"gte=0" เพื่อให้รับค่า 0 ได้
// 	type EstimateRequest struct {
// 		EstimatedCost float64 `json:"estimated_cost" binding:"gte=0"`
// 	}

// 	var req EstimateRequest
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "กรุณาระบุราคาประเมินให้ถูกต้อง"})
// 		return
// 	}

// 	// 1. ดึง Equipment ID และ Base Price ของงานซ่อมนี้
// 	var eqID *int
// 	var basePrice float64
// 	err := database.DB.QueryRow(`
// 		SELECT r.equipment_id, e.base_price
// 		FROM repairs r
// 		JOIN equipments e ON r.equipment_id = e.id
// 		WHERE r.id = $1
// 	`, repairID).Scan(&eqID, &basePrice)

// 	if err != nil || eqID == nil {
// 		// กรณีไม่มีข้อมูล Equipment ให้ผ่านไปเลยโดยไม่อิง Break-even
// 		database.DB.Exec(`UPDATE repairs SET estimated_cost = $1, status = 'กำลังซ่อม', accepted_at = CURRENT_TIMESTAMP WHERE id = $2`, req.EstimatedCost, repairID)
// 		c.JSON(http.StatusOK, gin.H{"message": "บันทึกราคาประเมินสำเร็จ", "status": "กำลังซ่อม"})
// 		return
// 	}

// 	// 2. คำนวณยอดซ่อมสะสม (Accumulated Cost) เฉพาะงานที่เสร็จแล้ว
// 	var accumulatedCost float64
// 	database.DB.QueryRow(`
// 		SELECT COALESCE(SUM(actual_cost), 0)
// 		FROM repairs
// 		WHERE equipment_id = $1 AND status = 'เสร็จเรียบร้อย'
// 	`, *eqID).Scan(&accumulatedCost)

// 	// 3. ลอจิกตรวจสอบเงื่อนไข Hybrid
// 	newStatus := "กำลังซ่อม"
// 	adminSystemNote := ""

// 	isSingleExceed := req.EstimatedCost > (basePrice * 0.5)
// 	isAccumulatedExceed := (accumulatedCost + req.EstimatedCost) > (basePrice * 0.7)

// 	if isSingleExceed || isAccumulatedExceed {
// 		// 🔥 [แก้ไขจุดสำคัญ] ตีกลับเป็น "รอซ่อม" เพื่อให้แอดมินตัดสินใจ โดยไม่เพิ่มสถานะใหม่
// 		newStatus = "รอซ่อม"

// 		if isSingleExceed && isAccumulatedExceed {
// 			adminSystemNote = fmt.Sprintf("ระบบระงับอัตโนมัติ: ประเมินครั้งนี้เกิน 50%% และยอดสะสมรวมเกิน 70%% ของราคาต้นทุน (%.2f บาท)", basePrice)
// 		} else if isSingleExceed {
// 			adminSystemNote = fmt.Sprintf("ระบบระงับอัตโนมัติ: ประเมินครั้งนี้เกิน 50%% ของราคาต้นทุน (%.2f บาท)", basePrice)
// 		} else {
// 			adminSystemNote = fmt.Sprintf("ระบบระงับอัตโนมัติ: ยอดสะสมรวมกับครั้งนี้เกิน 70%% ของราคาต้นทุน (%.2f บาท)", basePrice)
// 		}
// 	}

// 	// 4. บันทึกลง Database
// 	if newStatus == "รอซ่อม" {
// 		// ถ้าระบบระงับ จะอัปเดตราคาประเมิน ฝังโน้ตเตือนแอดมิน และเคลียร์เวลา accepted_at ออกชั่วคราว
// 		_, err = database.DB.Exec(`
// 			UPDATE repairs
// 			SET estimated_cost = $1, status = $2, admin_note = $3, accepted_at = NULL
// 			WHERE id = $4
// 		`, req.EstimatedCost, newStatus, adminSystemNote, repairID)
// 	} else {
// 		// ถ้าไม่เกินเกณฑ์ ให้ช่างเริ่มงานได้เลย (เปลี่ยนเป็น กำลังซ่อม)
// 		_, err = database.DB.Exec(`
// 			UPDATE repairs
// 			SET estimated_cost = $1, status = $2, accepted_at = CURRENT_TIMESTAMP
// 			WHERE id = $3
// 		`, req.EstimatedCost, newStatus, repairID)
// 	}

// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "อัปเดตข้อมูลล้มเหลว: " + err.Error()})
// 		return
// 	}

// 	// ส่งข้อความกลับไปบอก Frontend ให้ช่างรู้ตัวด้วยว่าถูกระงับ
// 	responseMsg := "ประเมินราคาสำเร็จและเริ่มซ่อมได้"
// 	if newStatus == "รอซ่อม" {
// 		responseMsg = "ประเมินราคาเกินจุดคุ้มทุน ระบบได้ส่งเรื่องกลับไปให้แอดมินพิจารณาแล้ว"
// 	}

// 	c.JSON(http.StatusOK, gin.H{
// 		"message":          responseMsg,
// 		"status":           newStatus,
// 		"estimated_cost":   req.EstimatedCost,
// 		"accumulated_cost": accumulatedCost,
// 	})
// }

// // ---------------------------------------------------------
// // 10. ApproveRepairThreshold: Admin อนุมัติงานที่ติดเงื่อนไข Break-even
// // ---------------------------------------------------------
// func ApproveRepairThreshold(c *gin.Context) {
// 	repairID := c.Param("id")

// 	// เปลี่ยนสถานะเป็น 'กำลังซ่อม' (ถือว่าเริ่มงานเลย)
// 	// ประทับเวลา accepted_at เป็นปัจจุบัน
// 	// และต่อท้าย admin_note ว่าอนุมัติแล้ว เพื่อเป็นประวัติ
// 	_, err := database.DB.Exec(`
// 		UPDATE repairs
// 		SET status = 'กำลังซ่อม',
// 		    accepted_at = CURRENT_TIMESTAMP,
// 		    admin_note = CONCAT(admin_note, ' -> (Admin อนุมัติให้ดำเนินการต่อ)')
// 		WHERE id = $1
// 	`, repairID)

// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถอนุมัติงานได้: " + err.Error()})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{"message": "อนุมัติงานซ่อมสำเร็จ"})
// }

package handlers

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
// ธีมสีกลาง สำหรับอีเมลทุกฉบับในระบบ
// ---------------------------------------------------------
const (
	colorSuccess = "#16a34a" // เขียว: สำเร็จ / ยืนยัน
	colorInfo    = "#2563eb" // ฟ้า: แจ้งเตือนทั่วไป / ข้อมูล
	colorWarning = "#d97706" // ส้ม: รอดำเนินการ / ต้องพิจารณา
	colorDanger  = "#dc2626" // แดง: ปัญหา / ยกเลิก / ปฏิเสธ
)

// emailHTMLStyle คืนค่า CSS กลางที่ใช้ร่วมกันทุกอีเมล (การ์ด, กล่องสี, ปุ่ม)
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

// wrapEmail ห่อเนื้อหาอีเมลด้วยการ์ด + หัวข้อสี ให้ทุกฉบับหน้าตาเป็นชุดเดียวกัน
func wrapEmail(headerColor string, headerText string, bodyContent string) string {
	return fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
			<meta charset="UTF-8">
			<style>%s</style>
		</head>
		<body>
			<div class="card">
				<h2 style="color: %s; margin-top: 0;">%s</h2>
				%s
			</div>
		</body>
		</html>
	`, emailHTMLStyle(), headerColor, headerText, bodyContent)
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

	// ---------- ส่งอีเมลแจ้งเตือน ----------
	go func() {
		formURL := "http://localhost:5173/repair/history"

		// 1. อีเมลสำหรับผู้แจ้ง
		userBodyContent := fmt.Sprintf(`
			<p><strong>หมายเลขใบแจ้ง:</strong> #%d</p>
			<p><strong>รายละเอียด:</strong> %s</p>
			<div class="box-warning">ขณะนี้อยู่ในสถานะ: <strong>รอซ่อม</strong></div>
			<p>ทางเราจะแจ้งเตือนอีกครั้งเมื่อช่างเริ่มเข้าดำเนินการครับ</p>
			<a href="%s" class="btn">ตรวจสอบรายการแจ้งซ่อมได้ที่นี่</a>
		`, repairID, description, formURL)
		userSubject := fmt.Sprintf("✅ รับเรื่องแจ้งซ่อมเรียบร้อยแล้ว (รหัสใบแจ้ง #%d)", repairID)
		userBody := wrapEmail(colorSuccess, "ระบบได้รับเรื่องแจ้งซ่อมของคุณเรียบร้อยแล้ว", userBodyContent)

		if reporterEmail != "" {
			if err := utils.SendEmailNotification([]string{reporterEmail}, userSubject, userBody); err != nil {
				fmt.Println("❌ ส่งอีเมลผู้แจ้ง (CreateRepair) ไม่สำเร็จ:", err)
			}
		}

		// 2. อีเมลสำหรับ Admin
		AdminLink := "http://localhost:5173"
		adminBodyContent := fmt.Sprintf(`
			<p>กรุณาเข้าสู่ระบบเพื่อพิจารณามอบหมายช่างดำเนินการ</p>
			<p><strong>หมายเลขใบแจ้ง:</strong> #%d</p>
			<p><strong>ผู้แจ้ง:</strong> %s</p>
			<p><strong>รายละเอียด:</strong> %s</p>
			<div class="box-warning">ขณะนี้อยู่ในสถานะ: <strong>รอซ่อม</strong></div>
			<a href="%s" class="btn">ตรวจสอบรายการแจ้งซ่อมได้ที่นี่</a>
		`, repairID, reporterEmail, description, AdminLink)
		adminSubject := fmt.Sprintf("🔔 มีงานแจ้งซ่อมใหม่เข้ามา (ใบงาน #%d)", repairID)
		adminBody := wrapEmail(colorInfo, "มีรายการแจ้งซ่อมใหม่เข้ามาในระบบ", adminBodyContent)

		if err := utils.SendEmailNotification([]string{utils.AdminEmail}, adminSubject, adminBody); err != nil {
			fmt.Println("❌ ส่งอีเมล admin (CreateRepair) ไม่สำเร็จ:", err)
		}
	}()

	c.JSON(http.StatusCreated, gin.H{"message": "สร้างรายการแจ้งซ่อมสำเร็จ"})
}

// ---------------------------------------------------------
// 2. GetAllRepairs: ดึงรายการแจ้งซ่อมทั้งหมด (JOIN ข้อมูลใหม่)
// ---------------------------------------------------------
func GetAllRepairs(c *gin.Context) {
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

	// ---------- ส่งอีเมลแจ้งเตือน ----------
	go func() {
		var reporterEmail, description string
		var techName *string

		err := database.DB.QueryRow(`
			SELECT r.reporter_email, r.description, u.full_name
			FROM repairs r
			LEFT JOIN users u ON u.id = $1
			WHERE r.id = $2`, req.TechnicianID, repairID).Scan(&reporterEmail, &description, &techName)

		if err != nil {
			fmt.Println("❌ AssignRepair query error:", err)
			return
		}

		safeTechName := ""
		if techName != nil {
			safeTechName = *techName
		}
		techEmail := utils.TechEmail

		// 1. แจ้งผู้แจ้ง
		if reporterEmail != "" {
			userBodyContent := fmt.Sprintf(`
				<p><strong>รายการ:</strong> %s</p>
				<div class="box-info">ช่างผู้รับผิดชอบ: <strong>%s</strong></div>
				<p>ช่างจะเข้าประเมินและดำเนินการต่อไปครับ</p>
			`, description, safeTechName)
			userSubject := fmt.Sprintf("⚙️ มอบหมายช่างซ่อมแล้ว (ใบงาน #%s)", repairID)
			userBody := wrapEmail(colorInfo, "ระบบได้มอบหมายช่างเข้าดูแลงานซ่อมของคุณแล้ว", userBodyContent)

			if err := utils.SendEmailNotification([]string{reporterEmail}, userSubject, userBody); err != nil {
				fmt.Println("❌ ส่งอีเมลผู้แจ้ง (AssignRepair) ไม่สำเร็จ:", err)
			}
		}

		// 2. แจ้งช่าง
		TechLink := "http://localhost:5173"
		if techEmail != "" {
			techBodyContent := fmt.Sprintf(`
				<p><strong>หมายเลขใบแจ้ง:</strong> #%s</p>
				<p><strong>รายละเอียดงาน:</strong> %s</p>
				<p>กรุณาเข้าประเมินราคาและดำเนินการซ่อมครับ</p>
				<a href="%s" class="btn">ตรวจสอบรายการมอบหมายงานได้ที่นี่</a>
			`, repairID, description, TechLink)
			techSubject := fmt.Sprintf("🔧 คุณได้รับมอบหมายงานซ่อมใหม่ (ใบงาน #%s)", repairID)
			techBody := wrapEmail(colorInfo, "คุณได้รับมอบหมายงานซ่อมใหม่", techBodyContent)

			if err := utils.SendEmailNotification([]string{techEmail}, techSubject, techBody); err != nil {
				fmt.Println("❌ ส่งอีเมลช่าง (AssignRepair) ไม่สำเร็จ:", err)
			}
		}

		// 3. แจ้ง Admin ยืนยันมอบหมายสำเร็จ
		AdminLink := "http://localhost:5173"
		adminBodyContent := fmt.Sprintf(`
			<p><strong>รายการ:</strong> %s</p>
			<div class="box-success">ช่างผู้รับผิดชอบ: <strong>%s</strong></div>
			<a href="%s" class="btn">ตรวจสอบรายการมอบหมายงานได้ที่นี่</a>
		`, description, safeTechName, AdminLink)
		adminSubject := fmt.Sprintf("✅ มอบหมายงานเสร็จสิ้น (ใบงาน #%s)", repairID)
		adminBody := wrapEmail(colorSuccess, "มอบหมายงานซ่อมเรียบร้อยแล้ว", adminBodyContent)

		if err := utils.SendEmailNotification([]string{utils.AdminEmail}, adminSubject, adminBody); err != nil {
			fmt.Println("❌ ส่งอีเมล admin (AssignRepair) ไม่สำเร็จ:", err)
		}
	}()

	c.JSON(http.StatusOK, gin.H{"message": "มอบหมายงานให้ช่างสำเร็จ"})
}

// ---------------------------------------------------------
// 5. RevokeRepair: Admin ดึงงานกลับ
// ---------------------------------------------------------
func RevokeRepair(c *gin.Context) {
	repairID := c.Param("id")
	revokeReason := c.PostForm("rejection_reason") // อ่านค่าก่อนเข้า goroutine

	// ดึงข้อมูลช่างปัจจุบันไว้ก่อน จะได้เอาไปแจ้งอีเมล (ก่อนที่ technician_id จะถูกล้าง)
	var techName, description string
	_ = database.DB.QueryRow(`
		SELECT u.full_name, r.description
		FROM repairs r
		JOIN users u ON r.technician_id = u.id
		WHERE r.id = $1`, repairID).Scan(&techName, &description)

	_, err := database.DB.Exec(`UPDATE repairs SET status = 'รอซ่อม', technician_id = NULL, accepted_at = NULL, completed_at = NULL WHERE id = $1`, repairID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// ---------- ส่งอีเมลแจ้งเตือนช่าง ----------
	go func() {
		techEmail := utils.TechEmail
		if techEmail == "" {
			return
		}

		bodyContent := fmt.Sprintf(`
			<p>เรียนคุณ %s</p>
			<p>ระบบขอแจ้งให้ทราบว่ารายการแจ้งซ่อมที่คุณรับผิดชอบ ได้ถูก Admin ดึงงานกลับ/ยกเลิกการมอบหมายแล้วครับ</p>
			<p><strong>หมายเลขใบแจ้ง:</strong> #%s</p>
			<p><strong>รายละเอียดงาน:</strong> %s</p>
			<div class="box-warning"><strong>📌 เหตุผลในการดึงงานกลับ:</strong><br>%s</div>
			<p class="footer-note">งานนี้จะถูกส่งกลับเข้าสู่ระบบกลางเพื่อรอดำเนินการต่อไป คุณไม่ต้องเข้าดำเนินการรายการนี้แล้วครับ</p>
		`, techName, repairID, description, revokeReason)
		subject := fmt.Sprintf("⚠️ แจ้งเตือนการดึงงานซ่อมกลับ (ใบงาน #%s)", repairID)
		body := wrapEmail(colorDanger, "แจ้งเตือนการดึงงานซ่อมกลับ", bodyContent)

		if err := utils.SendEmailNotification([]string{techEmail}, subject, body); err != nil {
			fmt.Println("❌ ส่งอีเมล RevokeRepair ไม่สำเร็จ:", err)
		}
	}()

	c.JSON(http.StatusOK, gin.H{"message": "ยกเลิกการจ่ายงานสำเร็จ"})
}

// ---------------------------------------------------------
// 6. RejectRepair: Tech ปฏิเสธงาน
// ---------------------------------------------------------
func RejectRepair(c *gin.Context) {
	repairID := c.Param("id")
	rejectionReason := c.PostForm("rejection_reason") // อ่านค่าก่อนเข้า goroutine

	_, err := database.DB.Exec(`UPDATE repairs SET status = 'รอซ่อม', technician_id = NULL, accepted_at = NULL, completed_at = NULL WHERE id = $1`, repairID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// ---------- ส่งอีเมลแจ้งผู้แจ้ง + Admin + ช่าง ----------
	go func() {
		var reporterEmail, description string
		err := database.DB.QueryRow(`SELECT reporter_email, description FROM repairs WHERE id = $1`, repairID).Scan(&reporterEmail, &description)
		if err != nil {
			fmt.Println("❌ RejectRepair query error:", err)
			return
		}

		// 1. แจ้งผู้แจ้ง + Admin
		bodyContent := fmt.Sprintf(`
			<p><strong>หมายเลขใบแจ้ง:</strong> #%s</p>
			<p><strong>รายการ:</strong> %s</p>
			<div class="box-danger"><strong>📌 เหตุผลในการปฏิเสธ:</strong><br>%s</div>
			<p class="footer-note">งานนี้ถูกส่งกลับเข้าสถานะ "รอซ่อม" เพื่อรอมอบหมายช่างคนใหม่</p>
		`, repairID, description, rejectionReason)
		subject := fmt.Sprintf("⚠️ ช่างปฏิเสธงานซ่อม (ใบงาน #%s)", repairID)
		body := wrapEmail(colorDanger, "ช่างปฏิเสธงาน / ส่งกลับเข้าระบบ", bodyContent)

		recipients := []string{utils.AdminEmail}
		if reporterEmail != "" {
			recipients = append(recipients, reporterEmail)
		}

		if err := utils.SendEmailNotification(recipients, subject, body); err != nil {
			fmt.Println("❌ ส่งอีเมล RejectRepair (ผู้แจ้ง+admin) ไม่สำเร็จ:", err)
		}

		// 2. แจ้งช่าง (คนที่เพิ่งปฏิเสธ) ยืนยันว่าปฏิเสธงานสำเร็จแล้ว
		techEmail := utils.TechEmail
		if techEmail != "" {
			techBodyContent := fmt.Sprintf(`
				<p><strong>หมายเลขใบแจ้ง:</strong> #%s</p>
				<p><strong>รายการ:</strong> %s</p>
				<p>งานนี้ถูกส่งกลับเข้าระบบกลาง ไม่ต้องดำเนินการต่อครับ</p>
			`, repairID, description)
			techSubject := fmt.Sprintf("✅ ปฏิเสธงานเสร็จสิ้น (ใบงาน #%s)", repairID)
			techBody := wrapEmail(colorSuccess, "คุณได้ปฏิเสธงานซ่อมนี้เรียบร้อยแล้ว", techBodyContent)

			if err := utils.SendEmailNotification([]string{techEmail}, techSubject, techBody); err != nil {
				fmt.Println("❌ ส่งอีเมล RejectRepair (ยืนยันช่าง) ไม่สำเร็จ:", err)
			}
		}
	}()

	c.JSON(http.StatusOK, gin.H{"message": "ปฏิเสธงานและส่งคืนระบบสำเร็จ"})
}

// ---------------------------------------------------------
// 7. UpdateRepairStatus: Tech อัปเดตสถานะ (รับงาน / ปิดงาน)
// ---------------------------------------------------------
func UpdateRepairStatus(c *gin.Context) {
	repairID := c.Param("id")

	status := c.PostForm("status")
	actualCostStr := c.PostForm("actual_cost")

	tNote := nullIfEmpty(c.PostForm("technician_note"))
	aNote := nullIfEmpty(c.PostForm("admin_note"))
	techNoteStr := c.PostForm("technician_note") // อ่านไว้ก่อนเข้า goroutine

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

	// ---------- ส่งอีเมลแจ้งเตือนตามสถานะ ----------
	go func() {
		var reporterEmail, description string
		qErr := database.DB.QueryRow(`SELECT reporter_email, description FROM repairs WHERE id = $1`, repairID).Scan(&reporterEmail, &description)
		if qErr != nil {
			fmt.Println("❌ UpdateRepairStatus query error:", qErr)
			return
		}

		HistoryLink := "http://localhost:5173/repair/history"
		switch status {
		case "เสร็จเรียบร้อย":
			evaluationFormLink := "https://rb.gy/bjiuqq"
			if reporterEmail != "" {
				userBodyContent := fmt.Sprintf(`
					<p><strong>รายการ:</strong> %s</p>
					<div class="box-success"><strong>บันทึกจากช่าง:</strong><br>%s</div>
					<p>ขอบคุณที่ใช้บริการครับ รบกวนให้คะแนนความพึงพอใจด้วยครับ</p>
					<a href="%s" class="btn btn-success">ให้คะแนนความพึงพอใจ</a>
					<a href="%s" class="btn btn-info">ตรวจสอบประวัติการแจ้งซ่อม</a>
				`, description, techNoteStr, evaluationFormLink, HistoryLink)
				userSubject := fmt.Sprintf("🎉 งานซ่อม #%s ดำเนินการเรียบร้อยแล้ว", repairID)
				userBody := wrapEmail(colorSuccess, "🎉 งานซ่อมของคุณเสร็จเรียบร้อยแล้ว", userBodyContent)

				if err := utils.SendEmailNotification([]string{reporterEmail}, userSubject, userBody); err != nil {
					fmt.Println("❌ ส่งอีเมลผู้แจ้ง (UpdateRepairStatus) ไม่สำเร็จ:", err)
				}
			}

			internalBodyContent := fmt.Sprintf(`
				<p><strong>หมายเลขใบแจ้ง:</strong> #%s</p>
				<p><strong>รายการ:</strong> %s</p>
				<a href="%s" class="btn btn-info">ตรวจสอบประวัติการแจ้งซ่อม</a>
			`, repairID, description, HistoryLink)
			internalSubject := fmt.Sprintf("✅ ปิดงานซ่อมแล้ว (ใบงาน #%s)", repairID)
			internalBody := wrapEmail(colorSuccess, "งานซ่อมถูกปิดเรียบร้อยแล้ว", internalBodyContent)

			internalRecipients := []string{utils.AdminEmail, utils.TechEmail}
			if err := utils.SendEmailNotification(internalRecipients, internalSubject, internalBody); err != nil {
				fmt.Println("❌ ส่งอีเมล admin+ช่าง (UpdateRepairStatus) ไม่สำเร็จ:", err)
			}

		case "ซ่อมไม่ได้":
			bodyContent := fmt.Sprintf(`
        <p><strong>หมายเลขใบแจ้ง:</strong> #%s</p>
        <p><strong>รายการ:</strong> %s</p>
        <div class="box-danger"><strong>เหตุผล:</strong><br>%s</div>
    `, repairID, description, techNoteStr)
			subject := fmt.Sprintf("⚠️ รายงานปัญหาการซ่อม (ใบงาน #%s)", repairID)
			body := wrapEmail(colorDanger, "ช่างไม่สามารถดำเนินการซ่อมได้", bodyContent)

			// 🎯 รวม Admin + ช่าง เป็นหลักไว้ก่อน
			recipients := []string{utils.AdminEmail, utils.TechEmail}
			if reporterEmail != "" {
				recipients = append(recipients, reporterEmail)
			}

			if err := utils.SendEmailNotification(recipients, subject, body); err != nil {
				fmt.Println("❌ ส่งอีเมล ซ่อมไม่ได้ (UpdateRepairStatus) ไม่สำเร็จ:", err)
			}
		}
	}()

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

	// ---------- ส่งอีเมลแจ้งผู้แจ้ง + ช่าง (ถ้ามี) ----------
	go func() {
		var reporterEmail, description string
		var techID *int
		qErr := database.DB.QueryRow(`SELECT reporter_email, description, technician_id FROM repairs WHERE id = $1`, repairID).Scan(&reporterEmail, &description, &techID)
		if qErr != nil {
			fmt.Println("❌ CancelRepairByAdmin query error:", qErr)
			return
		}

		bodyContent := fmt.Sprintf(`
			<p><strong>หมายเลขใบแจ้ง:</strong> #%s</p>
			<p><strong>รายการ:</strong> %s</p>
			<div class="box-danger"><strong>เหตุผล:</strong><br>%s</div>
		`, repairID, description, req.AdminNote)
		subject := fmt.Sprintf("⚠️ Admin ยกเลิกงานซ่อม (ใบงาน #%s)", repairID)
		body := wrapEmail(colorDanger, "งานซ่อมนี้ถูกยกเลิกโดยผู้ดูแลระบบ", bodyContent)

		recipients := []string{}
		if reporterEmail != "" {
			recipients = append(recipients, reporterEmail)
		}
		if techID != nil {
			recipients = append(recipients, utils.TechEmail)
		}

		if len(recipients) > 0 {
			if err := utils.SendEmailNotification(recipients, subject, body); err != nil {
				fmt.Println("❌ ส่งอีเมล CancelRepairByAdmin ไม่สำเร็จ:", err)
			}
		}
	}()

	c.JSON(http.StatusOK, gin.H{"message": "ยกเลิกงานสำเร็จ (สถานะ: ซ่อมไม่ได้)"})
}

// ---------------------------------------------------------
// 9. EstimateRepair: ช่างประเมินราคาและเช็กจุดคุ้มทุน (Hybrid Logic - 4 สถานะ)
// ---------------------------------------------------------
func EstimateRepair(c *gin.Context) {
	repairID := c.Param("id")

	type EstimateRequest struct {
		EstimatedCost float64 `json:"estimated_cost" binding:"gte=0"`
	}

	var req EstimateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "กรุณาระบุราคาประเมินให้ถูกต้อง"})
		return
	}

	var eqID *int
	var basePrice float64
	err := database.DB.QueryRow(`
		SELECT r.equipment_id, e.base_price
		FROM repairs r
		JOIN equipments e ON r.equipment_id = e.id
		WHERE r.id = $1
	`, repairID).Scan(&eqID, &basePrice)

	if err != nil || eqID == nil {
		database.DB.Exec(`UPDATE repairs SET estimated_cost = $1, status = 'กำลังซ่อม', accepted_at = CURRENT_TIMESTAMP WHERE id = $2`, req.EstimatedCost, repairID)

		go sendEstimateEmail(repairID, "กำลังซ่อม", "")

		c.JSON(http.StatusOK, gin.H{"message": "บันทึกราคาประเมินสำเร็จ", "status": "กำลังซ่อม"})
		return
	}

	var accumulatedCost float64
	database.DB.QueryRow(`
		SELECT COALESCE(SUM(actual_cost), 0)
		FROM repairs
		WHERE equipment_id = $1 AND status = 'เสร็จเรียบร้อย'
	`, *eqID).Scan(&accumulatedCost)

	newStatus := "กำลังซ่อม"
	adminSystemNote := ""

	isSingleExceed := req.EstimatedCost > (basePrice * 0.5)
	isAccumulatedExceed := (accumulatedCost + req.EstimatedCost) > (basePrice * 0.7)

	if isSingleExceed || isAccumulatedExceed {
		newStatus = "รอซ่อม"

		if isSingleExceed && isAccumulatedExceed {
			adminSystemNote = fmt.Sprintf("ระบบระงับอัตโนมัติ: ประเมินครั้งนี้เกิน 50%% และยอดสะสมรวมเกิน 70%% ของราคาต้นทุน (%.2f บาท)", basePrice)
		} else if isSingleExceed {
			adminSystemNote = fmt.Sprintf("ระบบระงับอัตโนมัติ: ประเมินครั้งนี้เกิน 50%% ของราคาต้นทุน (%.2f บาท)", basePrice)
		} else {
			adminSystemNote = fmt.Sprintf("ระบบระงับอัตโนมัติ: ยอดสะสมรวมกับครั้งนี้เกิน 70%% ของราคาต้นทุน (%.2f บาท)", basePrice)
		}
	}

	if newStatus == "รอซ่อม" {
		_, err = database.DB.Exec(`
			UPDATE repairs
			SET estimated_cost = $1, status = $2, admin_note = $3, accepted_at = NULL
			WHERE id = $4
		`, req.EstimatedCost, newStatus, adminSystemNote, repairID)
	} else {
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

	// ---------- ส่งอีเมลแจ้งเตือนตามผลการประเมิน ----------
	go sendEstimateEmail(repairID, newStatus, adminSystemNote)

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

// sendEstimateEmail แจ้งเตือนหลัง EstimateRepair
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
// 10. ApproveRepairThreshold: Admin อนุมัติงานที่ติดเงื่อนไข Break-even
// ---------------------------------------------------------
func ApproveRepairThreshold(c *gin.Context) {
	repairID := c.Param("id")

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

	// ---------- ส่งอีเมลแจ้งช่างว่า Admin อนุมัติแล้ว ให้ไปดำเนินการต่อ ----------
	go func() {
		var description string
		qErr := database.DB.QueryRow(`SELECT description FROM repairs WHERE id = $1`, repairID).Scan(&description)
		if qErr != nil {
			fmt.Println("❌ ApproveRepairThreshold query error:", qErr)
			return
		}

		techEmail := utils.TechEmail
		if techEmail != "" {
			bodyContent := fmt.Sprintf(`
				<p><strong>หมายเลขใบแจ้ง:</strong> #%s</p>
				<p><strong>รายการ:</strong> %s</p>
				<p>กรุณาเข้าดำเนินการซ่อมต่อได้ทันทีครับ</p>
			`, repairID, description)
			subject := fmt.Sprintf("✅ Admin อนุมัติงานซ่อมแล้ว (ใบงาน #%s)", repairID)
			body := wrapEmail(colorSuccess, "Admin อนุมัติให้ดำเนินการซ่อมต่อได้แล้ว", bodyContent)

			if err := utils.SendEmailNotification([]string{techEmail}, subject, body); err != nil {
				fmt.Println("❌ ส่งอีเมล ApproveRepairThreshold ไม่สำเร็จ:", err)
			}
		}
	}()

	c.JSON(http.StatusOK, gin.H{"message": "อนุมัติงานซ่อมสำเร็จ"})
}
