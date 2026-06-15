package user

import (
	"backend/database"
	"backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// 1. ดึงข้อมูลสถานที่ทั้งหมด
func GetLocations(c *gin.Context) {
	rows, err := database.DB.Query(
		"SELECT id, name FROM locations ORDER BY id",
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	defer rows.Close()

	var locations []models.Location

	for rows.Next() {
		var loc models.Location
		err := rows.Scan(
			&loc.ID,
			&loc.Name,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		locations = append(locations, loc)
	}

	c.JSON(http.StatusOK, locations)
}

// 2. [เพิ่มใหม่] ดึงข้อมูลชั้นตาม ID ของสถานที่
func GetFloorsByLocation(c *gin.Context) {
	// รับค่า ID ตึกจาก URL Parameter (เช่น /api/locations/1/floors)
	locationID := c.Param("id")

	rows, err := database.DB.Query(
		"SELECT id, location_id, floor_name FROM floors WHERE location_id = $1 ORDER BY id",
		locationID,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	defer rows.Close()

	// ประกาศแบบ make() เพื่อให้ถ้าไม่มีข้อมูล มันจะส่งกลับเป็น [] ไม่ใช่ null
	floors := make([]models.Floor, 0)

	for rows.Next() {
		var f models.Floor
		err := rows.Scan(
			&f.ID,
			&f.LocationID,
			&f.FloorName,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		floors = append(floors, f)
	}

	c.JSON(http.StatusOK, floors)
}
