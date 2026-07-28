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
	// 🔥 [แก้ Bug 2] ดักจับค่า "null" และ "undefined" ที่อาจหลุดมาจาก Frontend (React)
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
	otherLocation := nullIfEmpty(c.PostForm("other_location")) // 🔥 [เพิ่มใหม่]
	floorID := nullIfEmpty(c.PostForm("floor_id"))
	roomID := nullIfEmpty(c.PostForm("room_id"))
	equipmentID := nullIfEmpty(c.PostForm("equipment_id"))
	problemTypeID := nullIfEmpty(c.PostForm("problem_type_id"))
	otherProblemType := nullIfEmpty(c.PostForm("other_problem_type")) // 🔥 [เพิ่มใหม่]
	description := c.PostForm("description")

	var repairID int
	// 🔥 [แก้ไข] อัปเดต SQL ให้ INSERT ฟิลด์ other_location และ other_problem_type
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
	// 🔥 [แก้ไข] เพิ่มการ Select r.other_location, r.other_problem_type
	rows, err := database.DB.Query(`
        SELECT r.id, r.description, r.status, r.created_at, r.technician_id, r.technician_note, r.reporter_email,
               r.repair_cost, r.accepted_at, r.completed_at, r.other_location, r.other_problem_type,
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
		// 🔥 [แก้ไข] เพิ่มตัวแปรมารับค่า other_location, other_problem_type
		var locName, otherLoc, floorName, roomNumber, eqName, assetCode, probName, otherProb, techName, techNote *string
		var repairCost float64
		var createdAt time.Time
		var acceptedAt, completedAt *time.Time
		var techID *int

		err := rows.Scan(&id, &desc, &status, &createdAt, &techID, &techNote, &reporterEmail,
			&repairCost, &acceptedAt, &completedAt, &otherLoc, &otherProb,
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
			"other_location":     getString(otherLoc), // 🔥
			"floor_name":         getString(floorName),
			"room_number":        getString(roomNumber),
			"equipment_name":     getString(eqName),
			"asset_code":         getString(assetCode),
			"problem_type":       getString(probName),
			"other_problem_type": getString(otherProb), // 🔥
			"reporter_email":     reporterEmail,
			"technician_id":      techID,
			"technician_note":    getString(techNote),
			"technician_name":    getString(techName),
			"repair_cost":        repairCost,
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
	// 🔥 เพิ่มตัวแปร
	var locName, otherLoc, floorName, roomNumber, eqName, assetCode, probName, otherProb, techName, techNote *string
	var repairCost float64
	var createdAt time.Time
	var acceptedAt, completedAt *time.Time

	// 🔥 แก้ Query
	err := database.DB.QueryRow(`
        SELECT r.id, r.description, r.status, r.reporter_email, r.technician_note, r.created_at,
               r.repair_cost, r.accepted_at, r.completed_at, r.other_location, r.other_problem_type,
               l.name, f.floor_name, ro.room_number, eq.name, eq.asset_code, p.name, u.full_name
        FROM repairs r
        LEFT JOIN locations l ON r.location_id = l.id
        LEFT JOIN floors f ON r.floor_id = f.id
        LEFT JOIN rooms ro ON r.room_id = ro.id
        LEFT JOIN equipments eq ON r.equipment_id = eq.id
        LEFT JOIN problem_types p ON r.problem_type_id = p.id
        LEFT JOIN users u ON r.technician_id = u.id
        WHERE r.id = $1`, repairID).Scan(
		&id, &desc, &status, &reporterEmail, &techNote, &createdAt,
		&repairCost, &acceptedAt, &completedAt, &otherLoc, &otherProb,
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
		"technician_name":    getString(techName),
		"location_name":      getString(locName),
		"other_location":     getString(otherLoc), // 🔥
		"floor_name":         getString(floorName),
		"room_number":        getString(roomNumber),
		"equipment_name":     getString(eqName),
		"asset_code":         getString(assetCode),
		"problem_type":       getString(probName),
		"other_problem_type": getString(otherProb), // 🔥
		"repair_cost":        repairCost,
		"accepted_at":        acceptedAt,
		"completed_at":       completedAt,
		"created_at":         createdAt,
		"images":             images,
	}

	c.JSON(http.StatusOK, repair)
}

// ---------------------------------------------------------
// 4. AssignRepair: Admin มอบหมายช่าง (แค่ระบุตัวช่าง ยังไม่เริ่มจับเวลา)
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

	// [แก้ไข] เปลี่ยนแค่ technician_id แต่ให้สถานะเป็น 'รอซ่อม' เหมือนเดิม และเคลียร์เวลา accepted_at เผื่อไว้
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

	// [แก้ไข] รีเซ็ต accepted_at และ completed_at กลับเป็น NULL
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

	// [แก้ไข] รีเซ็ตเวลาเช่นเดียวกับการดึงงานกลับ
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
	technicianNote := c.PostForm("technician_note")
	repairCostStr := c.PostForm("repair_cost")

	if status == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "กรุณาระบุสถานะ"})
		return
	}

	// จัดการแปลงค่าซ่อมจาก String เป็น Float64
	var repairCost float64
	if repairCostStr != "" && repairCostStr != "null" && repairCostStr != "undefined" {
		parsedCost, err := strconv.ParseFloat(repairCostStr, 64)
		if err == nil {
			repairCost = parsedCost
		}
	}

	var query string

	// 🔥 [แก้ไข] ดักสถานะเพื่อสแตมป์เวลาให้ถูกต้อง
	if status == "กำลังซ่อม" {
		// ช่างกด "รับงาน" -> สแตมป์เวลา accepted_at เพื่อเริ่มจับเวลาซ่อม
		query = `UPDATE repairs SET status = $1, technician_note = $2, repair_cost = $3, accepted_at = CURRENT_TIMESTAMP WHERE id = $4`
	} else if status == "เสร็จเรียบร้อย" || status == "ซ่อมไม่ได้" {
		// ช่างกด "ปิดงาน" -> สแตมป์เวลา completed_at เพื่อจบการจับเวลา
		query = `UPDATE repairs SET status = $1, technician_note = $2, repair_cost = $3, completed_at = CURRENT_TIMESTAMP WHERE id = $4`
	} else {
		// กรณีอื่นๆ
		query = `UPDATE repairs SET status = $1, technician_note = $2, repair_cost = $3 WHERE id = $4`
	}

	_, err := database.DB.Exec(query, status, technicianNote, repairCost, repairID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// ส่วนจัดการอัปโหลดรูปภาพ (หลังซ่อม) ใช้งานได้ตามเดิมเลยครับ
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
