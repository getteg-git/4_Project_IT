package repair

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"backend/database"
	"backend/models"
)

// GetRepairLogs: ดึงประวัติการทำรายการ (Timeline) ของใบแจ้งซ่อมนั้นๆ
func GetRepairLogs(c *gin.Context) {
	repairID := c.Param("id")

	// ดึงข้อมูล Log พร้อม JOIN ไปหาตาราง users เพื่อเอาชื่อคนที่ทำรายการมาแสดงด้วย
	query := `
		SELECT rl.id, rl.repair_id, rl.user_id, u.full_name as user_name,
		       rl.action, rl.old_status, rl.new_status, rl.note, rl.created_at
		FROM repair_logs rl
		LEFT JOIN users u ON rl.user_id = u.id
		WHERE rl.repair_id = $1
		ORDER BY rl.created_at DESC
	`

	rows, err := database.DB.Query(query, repairID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถดึงประวัติการทำงานได้: " + err.Error()})
		return
	}
	defer rows.Close()

	// ใช้ make เพื่อให้แน่ใจว่าถ้าไม่มีข้อมูล จะส่ง [] (Array ว่าง) กลับไป แทนที่จะส่ง null ให้ React
	logs := make([]models.RepairLog, 0)

	for rows.Next() {
		var log models.RepairLog
		var userName *string // ใช้ pointer รับค่าเผื่อกรณีที่ระบบ (ไม่มี user) เป็นคนทำรายการ

		err := rows.Scan(
			&log.ID,
			&log.RepairID,
			&log.UserID,
			&userName,
			&log.Action,
			&log.OldStatus,
			&log.NewStatus,
			&log.Note,
			&log.CreatedAt,
		)
		if err != nil {
			continue
		}

		if userName != nil {
			log.UserName = *userName
		}

		logs = append(logs, log)
	}

	c.JSON(http.StatusOK, logs)
}
