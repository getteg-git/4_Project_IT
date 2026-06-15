package repair

import (
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"

	"backend/database"
)

// ---------------------------------------------------------
// 1. CreateRepair: สำหรับ User แจ้งซ่อมพร้อมแนบรูป (เพิ่ม floor_id)
// ---------------------------------------------------------
func CreateRepair(c *gin.Context) {
	reporterEmail := c.PostForm("reporter_email")
	locationID := c.PostForm("location_id")
	floorID := c.PostForm("floor_id") // รับค่า floor_id เพิ่มเข้ามา
	problemTypeID := c.PostForm("problem_type_id")
	description := c.PostForm("description")

	var repairID int
	// อัปเดต SQL ให้บันทึก floor_id ลงตาราง repairs
	err := database.DB.QueryRow(
		`INSERT INTO repairs (reporter_email, location_id, floor_id, problem_type_id, description, status)
		VALUES ($1, $2, $3, $4, $5, 'รอซ่อม') RETURNING id`,
		reporterEmail, locationID, floorID, problemTypeID, description,
	).Scan(&repairID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถบันทึกข้อมูลได้: " + err.Error()})
		return
	}

	form, _ := c.MultipartForm()
	if form != nil && form.File != nil {
		files := form.File["image"] // บังคับ Key ชื่อ "image" ให้ตรงกันกับฝั่ง React
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
// 2. GetAllRepairs: ดึงรายการแจ้งซ่อมทั้งหมด (JOIN เอาชื่อชั้นและชื่อช่างจริงมาแสดง)
// ---------------------------------------------------------
func GetAllRepairs(c *gin.Context) {
	rows, err := database.DB.Query(`
		SELECT r.id, r.description, r.status, r.created_at, r.technician_id, r.technician_note, r.reporter_email,
		       l.name as location_name, f.floor_name, p.name as problem_type_name, u.username as technician_name
		FROM repairs r
		JOIN locations l ON r.location_id = l.id
		JOIN floors f ON r.floor_id = f.id
		JOIN problem_types p ON r.problem_type_id = p.id
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
		var desc, status, locName, floorName, probName, reporterEmail string
		var createdAt time.Time

		var techID *int
		var techNote, techName *string

		err := rows.Scan(&id, &desc, &status, &createdAt, &techID, &techNote, &reporterEmail, &locName, &floorName, &probName, &techName)
		if err != nil {
			continue
		}

		// 🔥 [เพิ่มใหม่] ค้นหารูปภาพของงานซ่อม ID นี้
		imgRows, _ := database.DB.Query("SELECT image_url, image_type FROM repair_images WHERE repair_id = $1", id)
		images := make([]map[string]string, 0)
		for imgRows.Next() {
			var url, imgType string
			imgRows.Scan(&url, &imgType)
			images = append(images, map[string]string{"url": url, "type": imgType})
		}
		imgRows.Close() // อย่าลืมปิด connection รูปภาพ

		r := map[string]interface{}{
			"id":              id,
			"description":     desc,
			"status":          status,
			"created_at":      createdAt,
			"location":        locName,
			"floor_name":      floorName,
			"problem_type":    probName,
			"reporter_email":  reporterEmail,
			"technician_id":   techID,
			"technician_note": techNote,
			"technician_name": techName,
			"images":          images,
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
	var desc, status, locName, floorName, probName, reporterEmail string
	var techNote, techName *string

	err := database.DB.QueryRow(`
		SELECT r.id, r.description, r.status, r.reporter_email, r.technician_note, 
		       l.name, f.floor_name, p.name, u.username
		FROM repairs r
		JOIN locations l ON r.location_id = l.id
		JOIN floors f ON r.floor_id = f.id
		JOIN problem_types p ON r.problem_type_id = p.id
		LEFT JOIN users u ON r.technician_id = u.id
		WHERE r.id = $1`, repairID).Scan(&id, &desc, &status, &reporterEmail, &techNote, &locName, &floorName, &probName, &techName)

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

	repair := map[string]interface{}{
		"id":              id,
		"description":     desc,
		"status":          status,
		"reporter_email":  reporterEmail,
		"technician_note": techNote,
		"technician_name": techName,
		"location":        locName,
		"floor_name":      floorName,
		"problem_type":    probName,
		"images":          images,
	}

	c.JSON(http.StatusOK, repair)
}

// ---------------------------------------------------------
// 4. AssignRepair: Admin มอบหมายช่าง (เปลี่ยนเป็น 'กำลังซ่อม')
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
		`UPDATE repairs SET status = 'กำลังซ่อม', technician_id = $1 WHERE id = $2`,
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

	_, err := database.DB.Exec(`UPDATE repairs SET status = 'รอซ่อม', technician_id = NULL WHERE id = $1`, repairID)

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

	_, err := database.DB.Exec(`UPDATE repairs SET status = 'รอซ่อม', technician_id = NULL WHERE id = $1`, repairID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ปฏิเสธงานและส่งคืนระบบสำเร็จ"})
}

// ---------------------------------------------------------
// 7. UpdateRepairStatus: Tech ปิดงาน (เสร็จเรียบร้อย / ซ่อมไม่ได้)
// ---------------------------------------------------------
func UpdateRepairStatus(c *gin.Context) {
	repairID := c.Param("id")

	type UpdateStatusRequest struct {
		Status         string `json:"status" binding:"required"`
		TechnicianNote string `json:"technician_note" binding:"required"`
	}

	var req UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ข้อมูลไม่ถูกต้อง"})
		return
	}

	_, err := database.DB.Exec(
		`UPDATE repairs SET status = $1, technician_note = $2 WHERE id = $3`,
		req.Status, req.TechnicianNote, repairID,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "บันทึกผลการปฏิบัติงานสำเร็จ"})
}
