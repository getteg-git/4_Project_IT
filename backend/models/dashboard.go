package models

type DashboardOverview struct {
	TotalRepairs int `json:"total_repairs"`
	Pending      int `json:"pending"`
	InProgress   int `json:"in_progress"`
	Completed    int `json:"completed"`
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
	IsExceeded      bool    `json:"is_exceeded"`          // เกิน 50% หรือยัง (true/false)
}
