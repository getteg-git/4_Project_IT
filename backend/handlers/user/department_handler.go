package user

import (
	"backend/database"
	"backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GET: ดึงข้อมูลสาขาวิชาทั้งหมด (เอาไปทำ Dropdown หน้าเว็บ)
func GetDepartments(c *gin.Context) {
	query := `SELECT id, name FROM departments ORDER BY id ASC`
	rows, err := database.DB.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถดึงข้อมูลสาขาวิชาได้: " + err.Error()})
		return
	}
	defer rows.Close()

	var departments []models.Department
	for rows.Next() {
		var dept models.Department
		if err := rows.Scan(&dept.ID, &dept.Name); err != nil {
			continue
		}
		departments = append(departments, dept)
	}

	if departments == nil {
		departments = []models.Department{}
	}

	c.JSON(http.StatusOK, departments)
}
