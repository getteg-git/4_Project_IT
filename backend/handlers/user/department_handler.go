package user

import (
	"backend/database"
	"backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GET: ดึงข้อมูลสาขาวิชาทั้งหมด (เอาไปทำ Dropdown หน้าเว็บ)
func GetDepartments(c *gin.Context) {
	query := `
		SELECT d.id, COALESCE(d.degree_level_id, 0), d.name, COALESCE(dl.name, '')
		FROM departments d
		LEFT JOIN degree_levels dl ON dl.id = d.degree_level_id
		ORDER BY d.id ASC
	`
	rows, err := database.DB.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถดึงข้อมูลสาขาวิชาได้: " + err.Error()})
		return
	}
	defer rows.Close()

	var departments []models.Department
	for rows.Next() {
		var dept models.Department
		if err := rows.Scan(&dept.ID, &dept.DegreeLevelID, &dept.Name, &dept.DegreeLevelName); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถอ่านข้อมูลสาขาวิชาได้: " + err.Error()})
			return
		}
		departments = append(departments, dept)
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถอ่านข้อมูลสาขาวิชาได้: " + err.Error()})
		return
	}

	if departments == nil {
		departments = []models.Department{}
	}

	c.JSON(http.StatusOK, departments)
}
