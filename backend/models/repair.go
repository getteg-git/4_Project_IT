package models

import "time"

// ตาราง repairs
type Repair struct {
	ID               int        `json:"id"`
	ReporterEmail    string     `json:"reporter_email"`
	TechnicianID     *int       `json:"technician_id"`
	TechnicianName   string     `json:"technician_name,omitempty"`
	LocationID       *int       `json:"location_id"`
	LocationName     string     `json:"location_name,omitempty"`
	OtherLocation    *string    `json:"other_location"` // [เพิ่มใหม่] รองรับสถานที่/อาคารอื่นๆ
	FloorID          *int       `json:"floor_id"`
	FloorName        string     `json:"floor_name,omitempty"`
	RoomID           *int       `json:"room_id"`
	RoomNumber       string     `json:"room_number,omitempty"`
	EquipmentID      *int       `json:"equipment_id"`
	EquipmentName    string     `json:"equipment_name,omitempty"`
	AssetCode        string     `json:"asset_code,omitempty"`
	ProblemTypeID    *int       `json:"problem_type_id"`
	ProblemType      string     `json:"problem_type,omitempty"`
	OtherProblemType *string    `json:"other_problem_type"` // [เพิ่มใหม่] รองรับหมวดหมู่งานซ่อมอื่นๆ
	Description      string     `json:"description"`
	TechnicianNote   *string    `json:"technician_note"`
	Status           string     `json:"status"`
	RepairCost       float64    `json:"repair_cost"`
	AcceptedAt       *time.Time `json:"accepted_at"`
	CompletedAt      *time.Time `json:"completed_at"`
	CreatedAt        time.Time  `json:"created_at"`
}

// [ไม่มีการแก้ไข] ตาราง repair_images โครงสร้างถูกต้องแล้วครับ
type RepairImage struct {
	ID        int    `json:"id"`
	RepairID  int    `json:"repair_id"`
	ImageURL  string `json:"image_url"`
	ImageType string `json:"image_type"` // 'before' หรือ 'after'
}
