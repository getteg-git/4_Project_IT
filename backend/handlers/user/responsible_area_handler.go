package user

import (
	"database/sql"
	"fmt"
	"net/http"

	"backend/database"

	"github.com/gin-gonic/gin"
)

type responsibleAreaInput struct {
	Scope      string `json:"scope" binding:"required"`
	LocationID *int   `json:"location_id"`
	FloorID    *int   `json:"floor_id"`
	RoomID     *int   `json:"room_id"`
}

func GetTechnicianResponsibleAreas(c *gin.Context) {
	rows, err := database.DB.Query(`
		SELECT ra.technician_id, ra.scope, ra.location_id, ra.floor_id, ra.room_id,
		       l.name, f.floor_name, r.room_number
		FROM technician_responsible_areas ra
		LEFT JOIN rooms r ON r.id = ra.room_id
		LEFT JOIN floors f ON f.id = COALESCE(ra.floor_id, r.floor_id)
		LEFT JOIN locations l ON l.id = COALESCE(ra.location_id, f.location_id)
		ORDER BY ra.technician_id, l.name, f.floor_name, r.room_number
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถดึงข้อมูลพื้นที่รับผิดชอบได้"})
		return
	}
	defer rows.Close()

	areas := make([]gin.H, 0)
	for rows.Next() {
		var technicianID int
		var scope string
		var locationID, floorID, roomID sql.NullInt64
		var locationName, floorName, roomNumber sql.NullString
		if err := rows.Scan(
			&technicianID, &scope, &locationID, &floorID, &roomID,
			&locationName, &floorName, &roomNumber,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถอ่านข้อมูลพื้นที่รับผิดชอบได้"})
			return
		}

		area := gin.H{"technician_id": technicianID, "scope": scope}
		if locationID.Valid {
			area["location_id"] = int(locationID.Int64)
			area["location_name"] = locationName.String
		}
		if floorID.Valid {
			area["floor_id"] = int(floorID.Int64)
			area["floor_name"] = floorName.String
		}
		if roomID.Valid {
			area["room_id"] = int(roomID.Int64)
		}
		if locationName.Valid {
			area["location_name"] = locationName.String
		}
		if floorName.Valid {
			area["floor_name"] = floorName.String
		}
		if roomNumber.Valid {
			area["room_number"] = roomNumber.String
		}
		areas = append(areas, area)
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "เกิดข้อผิดพลาดระหว่างอ่านข้อมูลพื้นที่รับผิดชอบ"})
		return
	}

	c.JSON(http.StatusOK, areas)
}

func GetUserResponsibleAreas(c *gin.Context) {
	technicianID := c.Param("id")
	var isTechnician bool
	if err := database.DB.QueryRow(
		"SELECT EXISTS (SELECT 1 FROM users WHERE id = $1 AND role = 'technician')",
		technicianID,
	).Scan(&isTechnician); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถตรวจสอบบัญชีช่างได้"})
		return
	}
	if !isTechnician {
		c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบบัญชีช่าง"})
		return
	}

	rows, err := database.DB.Query(`
		SELECT ra.scope, ra.location_id, ra.floor_id, ra.room_id,
		       l.name, f.floor_name, r.room_number
		FROM technician_responsible_areas ra
		LEFT JOIN rooms r ON r.id = ra.room_id
		LEFT JOIN floors f ON f.id = COALESCE(ra.floor_id, r.floor_id)
		LEFT JOIN locations l ON l.id = COALESCE(ra.location_id, f.location_id)
		WHERE ra.technician_id = $1
		ORDER BY l.name, f.floor_name, r.room_number
	`, technicianID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถดึงข้อมูลพื้นที่รับผิดชอบได้"})
		return
	}
	defer rows.Close()

	areas := make([]gin.H, 0)
	for rows.Next() {
		var scope string
		var locationID, floorID, roomID sql.NullInt64
		var locationName, floorName, roomNumber sql.NullString
		if err := rows.Scan(
			&scope, &locationID, &floorID, &roomID,
			&locationName, &floorName, &roomNumber,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถอ่านข้อมูลพื้นที่รับผิดชอบได้"})
			return
		}

		area := gin.H{"scope": scope}
		if locationID.Valid {
			area["location_id"] = int(locationID.Int64)
			area["location_name"] = locationName.String
		}
		if floorID.Valid {
			area["floor_id"] = int(floorID.Int64)
			area["floor_name"] = floorName.String
		}
		if roomID.Valid {
			area["room_id"] = int(roomID.Int64)
		}
		if locationName.Valid {
			area["location_name"] = locationName.String
		}
		if floorName.Valid {
			area["floor_name"] = floorName.String
		}
		if roomNumber.Valid {
			area["room_number"] = roomNumber.String
		}
		areas = append(areas, area)
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "เกิดข้อผิดพลาดระหว่างอ่านข้อมูลพื้นที่รับผิดชอบ"})
		return
	}

	c.JSON(http.StatusOK, areas)
}

func UpdateUserResponsibleAreas(c *gin.Context) {
	technicianID := c.Param("id")
	var request struct {
		Areas []responsibleAreaInput `json:"areas"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "รูปแบบข้อมูลพื้นที่รับผิดชอบไม่ถูกต้อง"})
		return
	}

	var isTechnician bool
	if err := database.DB.QueryRow(
		"SELECT EXISTS (SELECT 1 FROM users WHERE id = $1 AND role = 'technician' AND is_active = TRUE)",
		technicianID,
	).Scan(&isTechnician); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถตรวจสอบบัญชีช่างได้"})
		return
	}
	if !isTechnician {
		c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบบัญชีช่างที่ใช้งานอยู่"})
		return
	}

	seen := make(map[string]bool, len(request.Areas))
	for _, area := range request.Areas {
		var exists bool
		var key string
		switch area.Scope {
		case "location":
			if area.LocationID == nil || area.FloorID != nil || area.RoomID != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "กรุณาระบุพื้นที่ระดับอาคารให้ถูกต้อง"})
				return
			}
			key = fmt.Sprintf("location:%d", *area.LocationID)
			err := database.DB.QueryRow("SELECT EXISTS (SELECT 1 FROM locations WHERE id = $1)", *area.LocationID).Scan(&exists)
			if err != nil || !exists {
				c.JSON(http.StatusBadRequest, gin.H{"error": "ไม่พบอาคารที่ระบุ"})
				return
			}
		case "floor":
			if area.LocationID != nil || area.FloorID == nil || area.RoomID != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "กรุณาระบุพื้นที่ระดับชั้นให้ถูกต้อง"})
				return
			}
			key = fmt.Sprintf("floor:%d", *area.FloorID)
			err := database.DB.QueryRow("SELECT EXISTS (SELECT 1 FROM floors WHERE id = $1)", *area.FloorID).Scan(&exists)
			if err != nil || !exists {
				c.JSON(http.StatusBadRequest, gin.H{"error": "ไม่พบชั้นที่ระบุ"})
				return
			}
		case "room":
			if area.LocationID != nil || area.FloorID != nil || area.RoomID == nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "กรุณาระบุพื้นที่ระดับห้องให้ถูกต้อง"})
				return
			}
			key = fmt.Sprintf("room:%d", *area.RoomID)
			err := database.DB.QueryRow("SELECT EXISTS (SELECT 1 FROM rooms WHERE id = $1)", *area.RoomID).Scan(&exists)
			if err != nil || !exists {
				c.JSON(http.StatusBadRequest, gin.H{"error": "ไม่พบห้องที่ระบุ"})
				return
			}
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "ระดับพื้นที่ต้องเป็นอาคาร ชั้น หรือห้อง"})
			return
		}
		if seen[key] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "มีการระบุพื้นที่ซ้ำ"})
			return
		}
		seen[key] = true
	}

	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถเริ่มบันทึกพื้นที่รับผิดชอบได้"})
		return
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM technician_responsible_areas WHERE technician_id = $1", technicianID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถแทนที่ข้อมูลพื้นที่รับผิดชอบเดิมได้"})
		return
	}
	for _, area := range request.Areas {
		if _, err := tx.Exec(`
			INSERT INTO technician_responsible_areas (technician_id, scope, location_id, floor_id, room_id)
			VALUES ($1, $2, $3, $4, $5)
		`, technicianID, area.Scope, area.LocationID, area.FloorID, area.RoomID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถบันทึกพื้นที่รับผิดชอบได้"})
			return
		}
	}
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถยืนยันการบันทึกพื้นที่รับผิดชอบได้"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "บันทึกพื้นที่รับผิดชอบสำเร็จ"})
}
