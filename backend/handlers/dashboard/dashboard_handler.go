package dashboard

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"backend/database"
	"backend/models"
)

// GetDashboardOverview ดึงข้อมูลสถิติภาพรวม
func GetDashboardOverview(c *gin.Context) {
	month := c.Query("month")
	year := c.Query("year")

	var overview models.DashboardOverview

	// 🔥 [แก้ไข] อัปเดต Query ให้ดึงครบ 4 สถานะหลัก (รอซ่อม, กำลังซ่อม, ซ่อมไม่ได้, เสร็จสิ้น)
	// หมายเหตุ: ถ้าใน Database คุณเก็บคำว่า 'เสร็จเรียบร้อย' ให้แก้คำว่า 'เสร็จสิ้น' ในบรรทัดล่างให้ตรงกันนะครับ
	query := `
        SELECT 
            COUNT(*) as total_repairs,
            COALESCE(SUM(CASE WHEN status = 'รอซ่อม' THEN 1 ELSE 0 END), 0) as pending,
            COALESCE(SUM(CASE WHEN status = 'กำลังซ่อม' THEN 1 ELSE 0 END), 0) as in_progress,
            COALESCE(SUM(CASE WHEN status = 'ซ่อมไม่ได้' THEN 1 ELSE 0 END), 0) as cannot_repair,
            COALESCE(SUM(CASE WHEN status = 'เสร็จสิ้น' THEN 1 ELSE 0 END), 0) as completed
        FROM repairs
        WHERE 1=1 
    `

	args := []interface{}{}
	paramIndex := 1

	if year != "" {
		query += fmt.Sprintf(" AND EXTRACT(YEAR FROM created_at) = $%d", paramIndex)
		args = append(args, year)
		paramIndex++
	}
	if month != "" {
		query += fmt.Sprintf(" AND EXTRACT(MONTH FROM created_at) = $%d", paramIndex)
		args = append(args, month)
		paramIndex++
	}

	// 🔥 [แก้ไข] เพิ่ม &overview.CannotRepair เข้าไปใน Scan ให้เรียงลำดับตรงกับ SELECT
	err := database.DB.QueryRow(query, args...).Scan(
		&overview.TotalRepairs,
		&overview.Pending,
		&overview.InProgress,
		&overview.CannotRepair,
		&overview.Completed,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "ไม่สามารถดึงข้อมูลสถิติภาพรวมได้",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, overview)
}

// GetTechnicianPerformance ดึงข้อมูลสถิติและเวลาเฉลี่ยในการทำงานของช่างแต่ละคน
func GetTechnicianPerformance(c *gin.Context) {
	month := c.Query("month")
	year := c.Query("year")

	query := `
        SELECT 
            u.id AS technician_id,
            u.full_name AS technician_name, -- 🔥 [แก้ให้ตรง DB] เปลี่ยนจาก username เป็น full_name เพื่อให้แสดงชื่อจริง
            COUNT(r.id) AS completed_jobs,
            COALESCE(AVG(EXTRACT(EPOCH FROM (r.completed_at - r.accepted_at))) / 60.0, 0) AS avg_repair_time_minutes
        FROM repairs r
        JOIN users u ON r.technician_id = u.id
        WHERE r.status = 'เสร็จเรียบร้อย' 
          AND r.accepted_at IS NOT NULL 
          AND r.completed_at IS NOT NULL
    `

	args := []interface{}{}
	paramIndex := 1

	if year != "" {
		query += fmt.Sprintf(" AND EXTRACT(YEAR FROM r.completed_at) = $%d", paramIndex)
		args = append(args, year)
		paramIndex++
	}
	if month != "" {
		query += fmt.Sprintf(" AND EXTRACT(MONTH FROM r.completed_at) = $%d", paramIndex)
		args = append(args, month)
		paramIndex++
	}

	query += " GROUP BY u.id, u.full_name ORDER BY completed_jobs DESC"

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "ไม่สามารถดึงข้อมูลประสิทธิภาพช่างได้",
			"details": err.Error(),
		})
		return
	}
	defer rows.Close()

	var performances []models.TechnicianPerformance

	for rows.Next() {
		var perf models.TechnicianPerformance
		if err := rows.Scan(
			&perf.TechnicianID,
			&perf.TechnicianName,
			&perf.CompletedJobs,
			&perf.AvgRepairTimeMinutes,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "เกิดข้อผิดพลาดในการอ่านข้อมูล"})
			return
		}
		performances = append(performances, perf)
	}

	if performances == nil {
		performances = []models.TechnicianPerformance{}
	}

	c.JSON(http.StatusOK, performances)
}

// GetBreakEvenAnalysis วิเคราะห์จุดคุ้มทุนของอุปกรณ์
func GetBreakEvenAnalysis(c *gin.Context) {
	// 🔥 [แก้ไข] เปลี่ยน r.repair_cost เป็น r.actual_cost
	query := `
        SELECT 
            e.id AS equipment_id,
            e.asset_code,
            e.name AS equipment_name,
            e.category,
            e.base_price,
            COALESCE(SUM(r.actual_cost), 0) AS total_repair_cost,
            CASE 
                WHEN e.base_price > 0 THEN (COALESCE(SUM(r.actual_cost), 0) / e.base_price) * 100 
                ELSE 0 
            END AS repair_ratio
        FROM equipments e
        LEFT JOIN repairs r ON e.id = r.equipment_id AND r.status = 'เสร็จเรียบร้อย'
        WHERE e.is_active = TRUE
        GROUP BY e.id, e.asset_code, e.name, e.category, e.base_price
        ORDER BY total_repair_cost DESC
    `

	rows, err := database.DB.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "ไม่สามารถดึงข้อมูลจุดคุ้มทุนได้",
			"details": err.Error(),
		})
		return
	}
	defer rows.Close()

	var analysisList []models.BreakEvenAnalysis

	for rows.Next() {
		var item models.BreakEvenAnalysis
		if err := rows.Scan(
			&item.EquipmentID,
			&item.AssetCode,
			&item.EquipmentName,
			&item.Category,
			&item.BasePrice,
			&item.TotalRepairCost,
			&item.RepairRatio,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "เกิดข้อผิดพลาดในการอ่านข้อมูล"})
			return
		}

		// 🔥 [แก้ไข] ปรับเป็น 70% ตามลอจิก Accumulated Cost ของระบบ Hybrid
		item.IsExceeded = item.RepairRatio >= 70.0

		analysisList = append(analysisList, item)
	}

	if analysisList == nil {
		analysisList = []models.BreakEvenAnalysis{}
	}

	c.JSON(http.StatusOK, analysisList)
}

// GetProblemTypeStats ดึงสัดส่วนหมวดหมู่ปัญหา/งานซ่อม (สำหรับทำ Doughnut Chart)
func GetProblemTypeStats(c *gin.Context) {
	query := `
        SELECT 
            COALESCE(e.category, 'ไม่ระบุหมวดหมู่') AS category_name,
            COUNT(r.id) AS repair_count
        FROM repairs r
        LEFT JOIN equipments e ON r.equipment_id = e.id
        GROUP BY e.category
        ORDER BY repair_count DESC
    `

	rows, err := database.DB.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "ไม่สามารถดึงข้อมูลสัดส่วนหมวดหมู่ได้",
			"details": err.Error(),
		})
		return
	}
	defer rows.Close()

	var statsList []models.ProblemTypeStat

	for rows.Next() {
		var stat models.ProblemTypeStat
		if err := rows.Scan(&stat.CategoryName, &stat.RepairCount); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "เกิดข้อผิดพลาดในการอ่านข้อมูล"})
			return
		}
		statsList = append(statsList, stat)
	}

	if statsList == nil {
		statsList = []models.ProblemTypeStat{}
	}

	c.JSON(http.StatusOK, statsList)
}
