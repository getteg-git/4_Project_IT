package user

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
