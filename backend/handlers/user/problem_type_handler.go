package user

import (
	"backend/database"
	"backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetProblemTypes(c *gin.Context) {
	rows, err := database.DB.Query(
		"SELECT id, name FROM problem_types ORDER BY id",
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	defer rows.Close()

	// ปรับเป็น make() เพื่อให้ส่งกลับเป็น [] แทน null กรณีไม่มีข้อมูล
	types := make([]models.ProblemType, 0)

	for rows.Next() {
		var t models.ProblemType

		err := rows.Scan(
			&t.ID,
			&t.Name,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		types = append(types, t)
	}

	c.JSON(http.StatusOK, types)
}

// ---------------------------------------------------------
// CreateProblemType: เพิ่มหมวดหมู่งานซ่อมใหม่
// ---------------------------------------------------------
func CreateProblemType(c *gin.Context) {
	type ProblemTypeRequest struct {
		Name string `json:"name" binding:"required"`
	}

	var req ProblemTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "กรุณาระบุชื่อหมวดหมู่งานซ่อม"})
		return
	}

	var newID int
	err := database.DB.QueryRow(
		`INSERT INTO problem_types (name) VALUES ($1) RETURNING id`,
		req.Name,
	).Scan(&newID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถบันทึกหมวดหมู่ได้: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "เพิ่มหมวดหมู่งานซ่อมใหม่สำเร็จ",
		"id":      newID,
		"name":    req.Name,
	})
}

// ---------------------------------------------------------
// UpdateProblemType: แก้ไขหมวดหมู่งานซ่อม
// ---------------------------------------------------------
func UpdateProblemType(c *gin.Context) {
	id := c.Param("id")

	type ProblemTypeRequest struct {
		Name string `json:"name" binding:"required"`
	}
	var req ProblemTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "กรุณาระบุชื่อหมวดหมู่ใหม่"})
		return
	}

	result, err := database.DB.Exec(`UPDATE problem_types SET name = $1 WHERE id = $2`, req.Name, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "แก้ไขข้อมูลไม่สำเร็จ: " + err.Error()})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบหมวดหมู่ที่ต้องการแก้ไข"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "แก้ไขหมวดหมู่สำเร็จ"})
}

// ---------------------------------------------------------
// DeleteProblemType: ลบหมวดหมู่งานซ่อม
// ---------------------------------------------------------
func DeleteProblemType(c *gin.Context) {
	id := c.Param("id")

	result, err := database.DB.Exec(`DELETE FROM problem_types WHERE id = $1`, id)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "ไม่สามารถลบได้ เนื่องจากหมวดหมู่นี้มีงานซ่อมค้างอยู่"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบหมวดหมู่ที่ต้องการลบ"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ลบหมวดหมู่สำเร็จ"})
}
