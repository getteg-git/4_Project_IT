package equipment

import (
	"backend/database"
	"backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ---------------------------------------------------------
// 1. ดึงข้อมูลสถานที่ทั้งหมด
// ---------------------------------------------------------
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

// ---------------------------------------------------------
// 2. ดึงข้อมูลชั้นตาม ID ของสถานที่
// ---------------------------------------------------------
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

// ---------------------------------------------------------
// 3. [เพิ่มใหม่] ดึงข้อมูลห้องตาม ID ของชั้น
// ---------------------------------------------------------
func GetRoomsByFloor(c *gin.Context) {
	// รับค่า ID ชั้นจาก URL Parameter (เช่น /api/floors/1/rooms)
	floorID := c.Param("id")

	rows, err := database.DB.Query(
		"SELECT id, floor_id, room_number FROM rooms WHERE floor_id = $1 ORDER BY room_number",
		floorID,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	defer rows.Close()

	rooms := make([]models.Room, 0)

	for rows.Next() {
		var r models.Room
		err := rows.Scan(
			&r.ID,
			&r.FloorID,
			&r.RoomNumber,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		rooms = append(rooms, r)
	}

	c.JSON(http.StatusOK, rooms)
}

// ---------------------------------------------------------
// 4. [เพิ่มใหม่] ดึงข้อมูลอุปกรณ์ตาม ID ของห้อง
// ---------------------------------------------------------
func GetEquipmentsByRoom(c *gin.Context) {
	// รับค่า ID ห้องจาก URL Parameter (เช่น /api/rooms/1/equipments)
	roomID := c.Param("id")

	rows, err := database.DB.Query(
		"SELECT id, room_id, name, asset_code, base_price FROM equipments WHERE room_id = $1 ORDER BY name",
		roomID,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	defer rows.Close()

	equipments := make([]models.Equipment, 0)

	for rows.Next() {
		var eq models.Equipment
		err := rows.Scan(
			&eq.ID,
			&eq.RoomID,
			&eq.Name,
			&eq.AssetCode,
			&eq.BasePrice,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		equipments = append(equipments, eq)
	}

	c.JSON(http.StatusOK, equipments)
}

// ---------------------------------------------------------
// CreateLocation: เพิ่มข้อมูลอาคาร/สถานที่ใหม่
// ---------------------------------------------------------
func CreateLocation(c *gin.Context) {
	type LocationRequest struct {
		Name string `json:"name" binding:"required"`
	}

	var req LocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "กรุณาระบุชื่ออาคาร/สถานที่"})
		return
	}

	var newID int
	// Insert และขอให้ Return ID ที่ถูกสร้างขึ้นมากลับมาด้วย
	err := database.DB.QueryRow(
		`INSERT INTO locations (name) VALUES ($1) RETURNING id`,
		req.Name,
	).Scan(&newID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถบันทึกสถานที่ได้: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "เพิ่มอาคาร/สถานที่ใหม่สำเร็จ",
		"id":      newID,
		"name":    req.Name,
	})
}

// ---------------------------------------------------------
// UpdateLocation: แก้ไขชื่อสถานที่
// ---------------------------------------------------------
func UpdateLocation(c *gin.Context) {
	id := c.Param("id") // รับ ID จาก URL

	type LocationRequest struct {
		Name string `json:"name" binding:"required"`
	}
	var req LocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "กรุณาระบุชื่ออาคาร/สถานที่ใหม่"})
		return
	}

	result, err := database.DB.Exec(`UPDATE locations SET name = $1 WHERE id = $2`, req.Name, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "แก้ไขข้อมูลไม่สำเร็จ: " + err.Error()})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบสถานที่ที่ต้องการแก้ไข"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "แก้ไขสถานที่สำเร็จ"})
}

// ---------------------------------------------------------
// DeleteLocation: ลบสถานที่
// ---------------------------------------------------------
func DeleteLocation(c *gin.Context) {
	id := c.Param("id")

	result, err := database.DB.Exec(`DELETE FROM locations WHERE id = $1`, id)
	if err != nil {
		// ดักจับ Error กรณีสถานที่นี้ถูกใช้งานอยู่ในตารางแจ้งซ่อม
		c.JSON(http.StatusConflict, gin.H{"error": "ไม่สามารถลบได้ เนื่องจากสถานที่นี้ถูกอ้างอิงอยู่ในรายการแจ้งซ่อม"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบสถานที่ที่ต้องการลบ"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ลบสถานที่สำเร็จ"})
}
