package repair

import (
	"backend/database"
	"backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ---------------------------------------------------------
// 1. ดึงข้อมูลระดับการศึกษาทั้งหมด (ปริญญาตรี, โท, เอก)
// ---------------------------------------------------------
func GetDegreeLevels(c *gin.Context) {
	rows, err := database.DB.Query(
		"SELECT id, name FROM degree_levels ORDER BY id",
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	defer rows.Close()

	degrees := make([]models.DegreeLevel, 0)

	for rows.Next() {
		var deg models.DegreeLevel
		err := rows.Scan(
			&deg.ID,
			&deg.Name,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		degrees = append(degrees, deg)
	}

	c.JSON(http.StatusOK, degrees)
}

// ---------------------------------------------------------
// 2. ดึงข้อมูลภาควิชาตาม ID ของระดับการศึกษา (Cascading Dropdown)
// ---------------------------------------------------------
func GetDepartmentsByDegree(c *gin.Context) {
	// รับค่า ID ระดับการศึกษาจาก URL Parameter (เช่น /api/degrees/1/departments)
	degreeID := c.Param("id")

	rows, err := database.DB.Query(
		"SELECT id, degree_level_id, name FROM departments WHERE degree_level_id = $1 ORDER BY name",
		degreeID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	defer rows.Close()

	departments := make([]models.Department, 0)

	for rows.Next() {
		var dept models.Department
		err := rows.Scan(
			&dept.ID,
			&dept.DegreeLevelID,
			&dept.Name,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		departments = append(departments, dept)
	}

	c.JSON(http.StatusOK, departments)
}
