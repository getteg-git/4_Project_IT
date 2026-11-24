package repair

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"backend/database"
)

type publicRepairCard struct {
	ID               int       `json:"id"`
	TicketNumber     string    `json:"ticket_number"`
	Description      string    `json:"description"`
	Status           string    `json:"status"`
	CreatedAt        time.Time `json:"created_at"`
	LocationName     string    `json:"location_name"`
	OtherLocation    string    `json:"other_location"`
	FloorName        string    `json:"floor_name"`
	RoomNumber       string    `json:"room_number"`
	OtherRoom        string    `json:"other_room"`
	EquipmentName    string    `json:"equipment_name"`
	HasImages        bool      `json:"has_images"`
	ProblemType      string    `json:"problem_type"`
	OtherProblemType string    `json:"other_problem_type"`
}

func GetPublicRepairBoard(c *gin.Context) {
	rows, err := database.DB.Query(`
		SELECT r.id, COALESCE(r.ticket_number, ''), r.description, r.status, r.created_at,
		       COALESCE(l.name, ''), COALESCE(r.other_location, ''),
		       COALESCE(f.floor_name, ''), COALESCE(ro.room_number, ''),
		       COALESCE(r.other_room, ''), COALESCE(NULLIF(r.other_equipment, ''), eq.name, ''),
		       EXISTS (SELECT 1 FROM repair_images ri WHERE ri.repair_id = r.id),
		       COALESCE(pt.name, ''), COALESCE(r.other_problem_type, '')
		FROM repairs r
		LEFT JOIN locations l ON l.id = r.location_id
		LEFT JOIN floors f ON f.id = r.floor_id
		LEFT JOIN rooms ro ON ro.id = r.room_id
		LEFT JOIN equipments eq ON eq.id = r.equipment_id
		LEFT JOIN problem_types pt ON pt.id = r.problem_type_id
		WHERE r.created_at >= CURRENT_TIMESTAMP - INTERVAL '90 days'
		ORDER BY r.created_at DESC, r.id DESC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถดึงรายการแจ้งซ่อมสาธารณะได้"})
		return
	}
	defer rows.Close()

	repairs := make([]publicRepairCard, 0)
	for rows.Next() {
		var repair publicRepairCard
		if err := rows.Scan(
			&repair.ID,
			&repair.TicketNumber,
			&repair.Description,
			&repair.Status,
			&repair.CreatedAt,
			&repair.LocationName,
			&repair.OtherLocation,
			&repair.FloorName,
			&repair.RoomNumber,
			&repair.OtherRoom,
			&repair.EquipmentName,
			&repair.HasImages,
			&repair.ProblemType,
			&repair.OtherProblemType,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถอ่านรายการแจ้งซ่อมสาธารณะได้"})
			return
		}
		repairs = append(repairs, repair)
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถอ่านรายการแจ้งซ่อมสาธารณะได้"})
		return
	}

	c.JSON(http.StatusOK, repairs)
}

func GetTrackedRepair(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	reporterEmail := strings.TrimSpace(c.Query("email"))
	if reporterEmail == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "กรุณาระบุอีเมลที่ใช้แจ้งซ่อม"})
		return
	}
	if len(reporterEmail) > 254 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ข้อมูลค้นหาไม่ถูกต้อง"})
		return
	}

	rows, err := database.DB.Query(`
		SELECT id
		FROM repairs
		WHERE LOWER(BTRIM(reporter_email)) = LOWER($1)
		ORDER BY created_at DESC, id DESC
	`, reporterEmail)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถค้นหารายการแจ้งซ่อมได้"})
		return
	}
	defer rows.Close()

	repairIDs := make([]int, 0)
	for rows.Next() {
		var repairID int
		if err := rows.Scan(&repairID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถอ่านรายการแจ้งซ่อมได้"})
			return
		}
		repairIDs = append(repairIDs, repairID)
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถอ่านรายการแจ้งซ่อมได้"})
		return
	}
	if len(repairIDs) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบรายการแจ้งซ่อมที่ใช้อีเมลนี้"})
		return
	}

	repairs := make([]map[string]interface{}, 0, len(repairIDs))
	for _, repairID := range repairIDs {
		repair, err := getTrackedRepairByID(strconv.Itoa(repairID), reporterEmail)
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบรายการแจ้งซ่อมที่ใช้อีเมลนี้"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถอ่านรายละเอียดรายการแจ้งซ่อมได้"})
			return
		}
		repairs = append(repairs, repair)
	}
	c.JSON(http.StatusOK, repairs)
}

func getTrackedRepairByID(repairID, email string) (map[string]interface{}, error) {
	var id int
	var description, status, reporterEmail string
	var ticketNumber, locationName, otherLocation, floorName, roomNumber, otherRoom *string
	var equipmentName, otherEquipment, problemType, otherProblemType, technicianName, technicianNote, adminNote, closureNote *string
	var assetCode *string
	var reporterType, reporterDegreeLevel, reporterDepartment *string
	var estimatedCost, actualCost float64
	var createdAt time.Time
	var acceptedAt, completedAt *time.Time
	var departmentID *int

	err := database.DB.QueryRow(`
		SELECT r.id, r.ticket_number, r.description, r.status, r.reporter_email,
		       r.reporter_type, r.reporter_degree_level, r.reporter_department,
		       r.technician_note, r.admin_note, r.created_at, r.estimated_cost, r.actual_cost,
		       r.accepted_at, r.completed_at, r.other_location, r.other_room, r.other_problem_type,
		       r.other_equipment, r.closure_note,
		       r.department_id, l.name, f.floor_name, ro.room_number, eq.name, eq.asset_code,
		       pt.name, u.full_name
		FROM repairs r
		LEFT JOIN locations l ON l.id = r.location_id
		LEFT JOIN floors f ON f.id = r.floor_id
		LEFT JOIN rooms ro ON ro.id = r.room_id
		LEFT JOIN equipments eq ON eq.id = r.equipment_id
		LEFT JOIN problem_types pt ON pt.id = r.problem_type_id
		LEFT JOIN users u ON u.id = r.technician_id
		WHERE r.id = $1 AND LOWER(BTRIM(r.reporter_email)) = LOWER($2)
	`, repairID, email).Scan(
		&id, &ticketNumber, &description, &status, &reporterEmail,
		&reporterType, &reporterDegreeLevel, &reporterDepartment,
		&technicianNote, &adminNote, &createdAt, &estimatedCost, &actualCost,
		&acceptedAt, &completedAt, &otherLocation, &otherRoom, &otherProblemType,
		&otherEquipment, &closureNote,
		&departmentID, &locationName, &floorName, &roomNumber, &equipmentName,
		&assetCode, &problemType, &technicianName,
	)
	if err != nil {
		return nil, err
	}

	imageRows, err := database.DB.Query(`
		SELECT id, image_type FROM repair_images WHERE repair_id = $1 ORDER BY id
	`, id)
	if err != nil {
		return nil, err
	}
	defer imageRows.Close()

	images := make([]map[string]string, 0)
	for imageRows.Next() {
		var imageID int
		var imageType string
		if err := imageRows.Scan(&imageID, &imageType); err != nil {
			return nil, err
		}
		images = append(images, map[string]string{
			"url":  fmt.Sprintf("/api/repairs/track/%d/images/%d", id, imageID),
			"type": imageType,
		})
	}
	if err := imageRows.Err(); err != nil {
		return nil, err
	}

	if customName := stringValue(otherEquipment); customName != "" {
		equipmentName = otherEquipment
	}

	return map[string]interface{}{
		"id":                    id,
		"ticket_number":         stringValue(ticketNumber),
		"description":           description,
		"status":                status,
		"reporter_email":        reporterEmail,
		"reporter_type":         stringValue(reporterType),
		"reporter_degree_level": stringValue(reporterDegreeLevel),
		"reporter_department":   stringValue(reporterDepartment),
		"department_id":         departmentID,
		"technician_note":       stringValue(technicianNote),
		"admin_note":            stringValue(adminNote),
		"closure_note":          stringValue(closureNote),
		"technician_name":       stringValue(technicianName),
		"location_name":         stringValue(locationName),
		"other_location":        stringValue(otherLocation),
		"floor_name":            stringValue(floorName),
		"room_number":           stringValue(roomNumber),
		"other_room":            stringValue(otherRoom),
		"equipment_name":        stringValue(equipmentName),
		"other_equipment":       stringValue(otherEquipment),
		"asset_code":            stringValue(assetCode),
		"problem_type":          stringValue(problemType),
		"other_problem_type":    stringValue(otherProblemType),
		"estimated_cost":        estimatedCost,
		"actual_cost":           actualCost,
		"accepted_at":           acceptedAt,
		"completed_at":          completedAt,
		"created_at":            createdAt,
		"images":                images,
	}, nil
}

func GetTrackedRepairLogs(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	repairID := c.Param("id")
	reporterEmail := strings.TrimSpace(c.GetHeader("X-Reporter-Email"))
	if reporterEmail == "" || len(reporterEmail) > 254 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "กรุณาระบุอีเมลที่ใช้แจ้งซ่อม"})
		return
	}

	var matched bool
	err := database.DB.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM repairs
			WHERE id = $1
			  AND LOWER(BTRIM(reporter_email)) = LOWER($2)
		)
	`, repairID, reporterEmail).Scan(&matched)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถตรวจสอบสิทธิ์ดูประวัติได้"})
		return
	}
	if !matched {
		c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบรายการแจ้งซ่อมที่ใช้อีเมลนี้"})
		return
	}

	rows, err := database.DB.Query(`
		SELECT id, repair_id, new_status, created_at
		FROM repair_logs
		WHERE repair_id = $1
		  AND new_status IN ('รับงานแล้ว', 'กำลังซ่อม', 'เสร็จเรียบร้อย', 'เสร็จสิ้น')
		ORDER BY created_at DESC
	`, repairID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถดึงความคืบหน้างานซ่อมได้"})
		return
	}
	defer rows.Close()

	logs := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, repairLogID int
		var status string
		var createdAt time.Time
		if err := rows.Scan(&id, &repairLogID, &status, &createdAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถอ่านความคืบหน้างานซ่อมได้"})
			return
		}
		logs = append(logs, map[string]interface{}{
			"id":         id,
			"repair_id":  repairLogID,
			"new_status": status,
			"created_at": createdAt,
		})
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถอ่านความคืบหน้างานซ่อมได้"})
		return
	}
	c.JSON(http.StatusOK, logs)
}

func GetTrackedRepairImage(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	repairID := c.Param("id")
	imageID := c.Param("image_id")
	reporterEmail := strings.TrimSpace(c.GetHeader("X-Reporter-Email"))
	if reporterEmail == "" || len(reporterEmail) > 254 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ไม่พบข้อมูลยืนยันสิทธิ์ดูรูปภาพ"})
		return
	}

	var imagePath string
	err := database.DB.QueryRow(`
		SELECT ri.image_url
		FROM repair_images ri
		JOIN repairs r ON r.id = ri.repair_id
		WHERE r.id = $1
		  AND ri.id = $2
		  AND LOWER(BTRIM(r.reporter_email)) = LOWER($3)
	`, repairID, imageID, reporterEmail).Scan(&imagePath)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบรูปภาพหรือข้อมูลยืนยันไม่ถูกต้อง"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถตรวจสอบสิทธิ์ดูรูปภาพได้"})
		return
	}

	filename := filepath.Base(imagePath)
	if filename == "." || filename == string(filepath.Separator) {
		c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบรูปภาพ"})
		return
	}
	file, err := os.Open(filepath.Join("uploads", filename))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบรูปภาพ"})
		return
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถอ่านรูปภาพได้"})
		return
	}
	contentType := "application/octet-stream"
	header := make([]byte, 512)
	n, err := file.Read(header)
	if err != nil && n == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถอ่านรูปภาพได้"})
		return
	}
	contentType = http.DetectContentType(header[:n])
	if _, err := file.Seek(0, 0); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถอ่านรูปภาพได้"})
		return
	}

	c.Header("Cache-Control", "no-store")
	c.DataFromReader(http.StatusOK, fileInfo.Size(), contentType, file, nil)
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
