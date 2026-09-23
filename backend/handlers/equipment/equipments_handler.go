package equipment

import (
	"backend/database"
	"backend/models"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// 1. GET: ดึงข้อมูลอุปกรณ์ทั้งหมด
func GetEquipments(c *gin.Context) {
	// 🛠️ แก้ไข: เพิ่มการ Select คอลัมน์ asset_code และ status เพื่อให้ Frontend นำไปแสดงผลในตารางได้
	query := `
		SELECT 
			id, 
			asset_code, 
			name, 
			category, 
			status, 
			base_price 
		FROM equipments 
		WHERE is_active = true 
		ORDER BY id DESC
	`
	rows, err := database.DB.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถดึงข้อมูลอุปกรณ์ได้: " + err.Error()})
		return
	}
	defer rows.Close()

	var equipments []models.Equipment
	for rows.Next() {
		var eq models.Equipment
		// 🛠️ แก้ไข: เรียงลำดับตัวแปรใน Scan ให้ตรงกับลำดับคอลัมน์ใน SELECT ด้านบน
		if err := rows.Scan(&eq.ID, &eq.AssetCode, &eq.Name, &eq.Category, &eq.Status, &eq.BasePrice); err != nil {
			continue
		}
		equipments = append(equipments, eq)
	}

	if equipments == nil {
		equipments = []models.Equipment{}
	}

	c.JSON(http.StatusOK, equipments)
}

// 2. POST: สร้างอุปกรณ์ใหม่
func CreateEquipment(c *gin.Context) {
	var req models.Equipment
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ข้อมูลไม่ถูกต้อง"})
		return
	}

	tempAssetCode := fmt.Sprintf("TEMP-%d", time.Now().Unix())

	// 🛠️ แก้ไข: เปลี่ยน ? เป็น $1, $2, $3, $4 สำหรับ PostgreSQL และกำหนด default status
	query := `
		INSERT INTO equipments (name, category, base_price, asset_code, status, is_active) 
		VALUES ($1, $2, $3, $4, 'พร้อมใช้งาน', true)
	`
	_, err := database.DB.Exec(query, req.Name, req.Category, req.BasePrice, tempAssetCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถเพิ่มอุปกรณ์ได้: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "เพิ่มอุปกรณ์สำเร็จ"})
}

// 3. PUT: แก้ไขข้อมูลอุปกรณ์
func UpdateEquipment(c *gin.Context) {
	id := c.Param("id")
	var req models.Equipment

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ข้อมูลไม่ถูกต้อง"})
		return
	}

	// 🛠️ แก้ไข: เปลี่ยน ? เป็น $1, $2, $3, $4
	query := `UPDATE equipments SET name = $1, category = $2, base_price = $3 WHERE id = $4`
	result, err := database.DB.Exec(query, req.Name, req.Category, req.BasePrice, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถอัปเดตอุปกรณ์ได้: " + err.Error()})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบอุปกรณ์ที่ต้องการแก้ไข"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "อัปเดตข้อมูลสำเร็จ"})
}

// 4. DELETE: ลบอุปกรณ์ (ใช้เทคนิค Soft Delete)
func DeleteEquipment(c *gin.Context) {
	id := c.Param("id")

	// 🛠️ แก้ไข: เปลี่ยน ? เป็น $1
	query := `UPDATE equipments SET is_active = false WHERE id = $1`
	result, err := database.DB.Exec(query, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถลบอุปกรณ์ได้: " + err.Error()})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบอุปกรณ์ที่ต้องการลบ"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ลบอุปกรณ์สำเร็จ"})
}
