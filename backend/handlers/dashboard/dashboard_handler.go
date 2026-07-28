package dashboard

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"backend/database" // ปรับ import path ให้ตรงกับโปรเจกต์คุณ
	"backend/models"
)

// GetDashboardOverview ดึงข้อมูลสถิติภาพรวม
func GetDashboardOverview(c *gin.Context) {
	// 1. รับค่า Filter (เดือน/ปี) จาก Query String
	month := c.Query("month")
	year := c.Query("year")

	var overview models.DashboardOverview

	// 2. เขียน SQL Query โดยใช้ SUM + CASE WHEN และจับสถานะให้ตรงกับ DB
	query := `
		SELECT 
			COUNT(*) as total_repairs,
			COALESCE(SUM(CASE WHEN status = 'รอซ่อม' THEN 1 ELSE 0 END), 0) as pending,
			COALESCE(SUM(CASE WHEN status = 'กำลังซ่อม' THEN 1 ELSE 0 END), 0) as in_progress,
			COALESCE(SUM(CASE WHEN status = 'เสร็จเรียบร้อย' THEN 1 ELSE 0 END), 0) as completed
		FROM repairs
		WHERE 1=1 
	`

	args := []interface{}{}
	paramIndex := 1 // เอาไว้นับเลข $1, $2 ของ PostgreSQL

	// 3. เพิ่มเงื่อนไข Filter ตามเดือน/ปี แบบ PostgreSQL
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

	// 4. สั่งรัน Query
	err := database.DB.QueryRow(query, args...).Scan(
		&overview.TotalRepairs,
		&overview.Pending,
		&overview.InProgress,
		&overview.Completed,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "ไม่สามารถดึงข้อมูลสถิติภาพรวมได้",
			"details": err.Error(),
		})
		return
	}

	// 5. ส่งข้อมูลกลับไปให้ Frontend
	c.JSON(http.StatusOK, overview)
}

// GetTechnicianPerformance ดึงข้อมูลสถิติและเวลาเฉลี่ยในการทำงานของช่างแต่ละคน
func GetTechnicianPerformance(c *gin.Context) {
	month := c.Query("month")
	year := c.Query("year")

	// ใช้ JOIN กับตาราง users เพื่อดึงชื่อช่างออกมาด้วย
	// หมายเหตุ: ตรง u.username ให้เปลี่ยนเป็นชื่อคอลัมน์ที่เก็บชื่อในตาราง users ของคุณ (เช่น u.full_name หรือ u.name)
	query := `
		SELECT 
			u.id AS technician_id,
			u.username AS technician_name, 
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

	// Filter ตามเดือน/ปี (แนะนำให้กรองจากวันที่ซ่อมเสร็จ completed_at)
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

	// GROUP BY ช่างแต่ละคน
	query += " GROUP BY u.id, u.username ORDER BY completed_jobs DESC"

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

	// ถ้าไม่มีข้อมูล ให้ส่ง Array ว่างกลับไป (Frontend จะได้ไม่พังเพราะได้ค่า null)
	if performances == nil {
		performances = []models.TechnicianPerformance{}
	}

	c.JSON(http.StatusOK, performances)
}

// GetBreakEvenAnalysis วิเคราะห์จุดคุ้มทุนของอุปกรณ์ (เทียบค่าซ่อมสะสมกับ base_price 50%)
func GetBreakEvenAnalysis(c *gin.Context) {
	// ใช้ LEFT JOIN เพื่อดึงอุปกรณ์ทุกชิ้น แม้ว่าจะยังไม่เคยซ่อมเลยก็ตาม (ค่าซ่อมสะสมจะเป็น 0)
	query := `
		SELECT 
			e.id AS equipment_id,
			e.asset_code,
			e.name AS equipment_name,
			e.category,
			e.base_price,
			COALESCE(SUM(r.repair_cost), 0) AS total_repair_cost,
			CASE 
				WHEN e.base_price > 0 THEN (COALESCE(SUM(r.repair_cost), 0) / e.base_price) * 100 
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

		// เช็คว่าเกินเกณฑ์ 50% หรือยัง
		item.IsExceeded = item.RepairRatio >= 50.0

		analysisList = append(analysisList, item)
	}

	if analysisList == nil {
		analysisList = []models.BreakEvenAnalysis{}
	}

	c.JSON(http.StatusOK, analysisList)
}
