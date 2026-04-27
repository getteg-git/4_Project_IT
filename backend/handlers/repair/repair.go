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
// 1. CreateRepair: สำหรับ User แจ้งซ่อมพร้อมแนบรูป (Before)
// ---------------------------------------------------------
func CreateRepair(c *gin.Context) {
	userID := c.PostForm("user_id")
	locationID := c.PostForm("location_id")
	problemTypeID := c.PostForm("problem_type_id")
	description := c.PostForm("description")

	var repairID int
	err := database.DB.QueryRow(
		`INSERT INTO repairs (user_id, location_id, problem_type_id, description)
		VALUES ($1,$2,$3,$4) RETURNING id`,
		userID, locationID, problemTypeID, description,
	).Scan(&repairID)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	form, _ := c.MultipartForm()
	files := form.File["images"]

	for _, file := range files {
		filename := fmt.Sprintf("%d_%d%s", repairID, time.Now().UnixNano(), filepath.Ext(file.Filename))
		path := "uploads/" + filename

		if err := c.SaveUploadedFile(file, path); err != nil {
			continue
		}

		imageURL := "/uploads/" + filename
		database.DB.Exec(
			`INSERT INTO repair_images (repair_id, image_url, image_type) VALUES ($1,$2, 'before')`,
			repairID, imageURL,
		)
	}

	c.JSON(http.StatusCreated, gin.H{"message": "สร้างรายการแจ้งซ่อมสำเร็จ"})
}

// ---------------------------------------------------------
// 2. GetAllRepairs: ดึงรายการแจ้งซ่อมทั้งหมด (พร้อมข้อมูลชื่อตึกและชื่อปัญหา)
// ---------------------------------------------------------
func GetAllRepairs(c *gin.Context) {
	// 1. เพิ่ม r.technician_id ใน SELECT
	rows, err := database.DB.Query(`
		SELECT r.id, r.description, r.status, r.created_at, r.technician_id,
		       l.name as location_name, p.name as problem_type_name, 
		       u.username as requester_name
		FROM repairs r
		JOIN locations l ON r.location_id = l.id
		JOIN problem_types p ON r.problem_type_id = p.id
		JOIN users u ON r.user_id = u.id
		ORDER BY r.created_at DESC`)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var repairs []map[string]interface{}
	for rows.Next() {
		var r map[string]interface{}
		var id, desc, status, locName, probName, reqName string
		var createdAt time.Time

		// 2. ประกาศตัวแปรเป็น Pointer (*int) เพื่อรับค่า NULL จาก Database ได้
		var techID *int

		// 3. เอา &techID ไปใส่ใน rows.Scan ให้ตรงตำแหน่ง
		rows.Scan(&id, &desc, &status, &createdAt, &techID, &locName, &probName, &reqName)

		r = map[string]interface{}{
			"id":            id,
			"description":   desc,
			"status":        status,
			"created_at":    createdAt,
			"location":      locName,
			"problem_type":  probName,
			"requester":     reqName,
			"technician_id": techID, // 4. ใส่ลงไปใน JSON Response ตรงนี้!
		}
		repairs = append(repairs, r)
	}

	c.JSON(http.StatusOK, repairs)
}

// ---------------------------------------------------------
// 3. GetRepairByID: ดูรายละเอียดงานซ่อมรายตัว พร้อมรูปภาพทั้งหมด
// ---------------------------------------------------------
func GetRepairByID(c *gin.Context) {
	repairID := c.Param("id")

	// 1. ดึงข้อมูลงานซ่อม
	var repair map[string]interface{}
	var id, desc, status, locName, probName string
	err := database.DB.QueryRow(`
		SELECT r.id, r.description, r.status, l.name, p.name
		FROM repairs r
		JOIN locations l ON r.location_id = l.id
		JOIN problem_types p ON r.problem_type_id = p.id
		WHERE r.id = $1`, repairID).Scan(&id, &desc, &status, &locName, &probName)

	if err != nil {
		c.JSON(404, gin.H{"error": "ไม่พบรายการแจ้งซ่อม"})
		return
	}

	// 2. ดึงรูปภาพทั้งหมดที่เกี่ยวข้อง
	rows, _ := database.DB.Query("SELECT image_url, image_type FROM repair_images WHERE repair_id = $1", id)
	defer rows.Close()

	var images []map[string]string
	for rows.Next() {
		var url, imgType string
		rows.Scan(&url, &imgType)
		images = append(images, map[string]string{"url": url, "type": imgType})
	}

	repair = map[string]interface{}{
		"id": id, "description": desc, "status": status,
		"location": locName, "problem_type": probName, "images": images,
	}

	c.JSON(http.StatusOK, repair)
}

// ---------------------------------------------------------
// 4. ApproveRepair: Admin อนุมัติและมอบหมายช่าง
// ---------------------------------------------------------
func ApproveRepair(c *gin.Context) {
	repairID := c.Param("id")
	type ApprovalRequest struct {
		TechnicianID int `json:"technician_id"`
	}

	var req ApprovalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ข้อมูลไม่ถูกต้อง"})
		return
	}

	_, err := database.DB.Exec(
		`UPDATE repairs SET status = 'approved', technician_id = $1 WHERE id = $2`,
		req.TechnicianID, repairID,
	)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "อนุมัติและมอบหมายงานสำเร็จ"})
}

// ---------------------------------------------------------
// 5. UpdateProgress: Technician รับงานและแนบรูประหว่างซ่อม (During)
// ---------------------------------------------------------
func UpdateProgress(c *gin.Context) {
	repairID := c.Param("id")

	_, err := database.DB.Exec(`UPDATE repairs SET status = 'in_progress' WHERE id = $1`, repairID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	form, err := c.MultipartForm()
	if err == nil {
		files := form.File["tech_images"]
		for _, file := range files {
			filename := fmt.Sprintf("tech_%s_%d%s", repairID, time.Now().UnixNano(), filepath.Ext(file.Filename))
			path := "uploads/" + filename
			if err := c.SaveUploadedFile(file, path); err == nil {
				imageURL := "/uploads/" + filename
				database.DB.Exec(
					`INSERT INTO repair_images (repair_id, image_url, image_type) VALUES ($1, $2, 'during')`,
					repairID, imageURL,
				)
			}
		}
	}
	c.JSON(http.StatusOK, gin.H{"message": "อัปเดตสถานะกำลังซ่อมสำเร็จ"})
}

// ---------------------------------------------------------
// 6. CompleteRepair: ปิดงานซ่อม (Done)
// ---------------------------------------------------------
// ---------------------------------------------------------
// 6. CompleteRepair: ปิดงานซ่อม (Done) พร้อมบันทึกรูปจากช่าง
// ---------------------------------------------------------
func CompleteRepair(c *gin.Context) {
	repairID := c.Param("id")

	// 1. อัปเดตสถานะงานเป็น 'done'
	_, err := database.DB.Exec(`UPDATE repairs SET status = 'done' WHERE id = $1`, repairID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// 2. รับรูปภาพที่ช่างแนบมาตอนปิดงาน (ใช้ชื่อฟิลด์ "images" ตามที่ React ส่งมา)
	form, err := c.MultipartForm()
	if err == nil {
		files := form.File["images"]
		for _, file := range files {
			// สร้างชื่อไฟล์ไม่ให้ซ้ำกัน
			filename := fmt.Sprintf("tech_done_%s_%d%s", repairID, time.Now().UnixNano(), filepath.Ext(file.Filename))
			path := "uploads/" + filename

			// ถ้าเซฟไฟล์ลงโฟลเดอร์ uploads สำเร็จ
			if err := c.SaveUploadedFile(file, path); err == nil {
				imageURL := "/uploads/" + filename

				// บันทึก Path รูปลง Database โดยระบุว่าเป็นรูปประเภท 'during' (รูปจากช่าง)
				database.DB.Exec(
					`INSERT INTO repair_images (repair_id, image_url, image_type) VALUES ($1, $2, 'during')`,
					repairID, imageURL,
				)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "ปิดงานซ่อมและอัปโหลดรูปหลักฐานสำเร็จ"})
}
