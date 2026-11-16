package models

type DashboardOverview struct {
	TotalRepairs int `json:"total_repairs"` // ยอดแจ้งซ่อมทั้งหมด
	Pending      int `json:"pending"`       // รอซ่อม
	InProgress   int `json:"in_progress"`   // กำลังซ่อม
	CannotRepair int `json:"cannot_repair"` // 🔥 ซ่อมไม่ได้ (เพิ่มให้ครบ 4 สถานะเป๊ะๆ ตามลอจิก)
	Completed    int `json:"completed"`     // เสร็จสิ้น / เสร็จเรียบร้อย
}

type TechnicianPerformance struct {
	TechnicianID         int     `json:"technician_id"`
	TechnicianName       string  `json:"technician_name"`
	CompletedJobs        int     `json:"completed_jobs"`
	AvgRepairTimeMinutes float64 `json:"avg_repair_time_minutes"` // เวลาเฉลี่ยเป็นนาที
}

type BreakEvenAnalysis struct {
	EquipmentID     int     `json:"equipment_id"`
	AssetCode       string  `json:"asset_code"`
	EquipmentName   string  `json:"equipment_name"`
	Category        string  `json:"category"`
	BasePrice       float64 `json:"base_price"`
	TotalRepairCost float64 `json:"total_repair_cost"`
	RepairRatio     float64 `json:"repair_ratio_percent"` // เปอร์เซ็นต์เทียบกับราคาต้นทุน
	Threshold30     float64 `json:"threshold_30"`
	Threshold50     float64 `json:"threshold_50"`
	IsWarning       bool    `json:"is_warning"`
	IsExceeded      bool    `json:"is_exceeded"`
}

type ProblemTypeStat struct {
	CategoryName string `json:"category_name"`
	RepairCount  int    `json:"repair_count"`
}
