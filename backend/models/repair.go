package models

import "time"

// ตาราง repairs
type Repair struct {
	ID             int       `json:"id"`
	ReporterEmail  string    `json:"reporter_email"`
	TechnicianID   *int      `json:"technician_id"`
	TechnicianName string    `json:"technician_name,omitempty"` // [เพิ่มใหม่] สำหรับเก็บ full_name ของช่างเวลา JOIN ตาราง users
	LocationID     int       `json:"location_id"`
	LocationName   string    `json:"location_name,omitempty"` // [เพิ่มใหม่] สำหรับเก็บชื่อสถานที่เวลา JOIN ตาราง locations
	FloorID        int       `json:"floor_id"`
	FloorName      string    `json:"floor_name,omitempty"` // [เพิ่มใหม่] สำหรับเก็บชื่อชั้นเวลา JOIN ตาราง floors
	Room           string    `json:"room"`                 // [เพิ่มใหม่] เก็บข้อมูลเลขห้อง / พิกัดจุดเกิดเหตุแยกเฉพาะ
	ProblemTypeID  int       `json:"problem_type_id"`
	ProblemType    string    `json:"problem_type,omitempty"` // [เพิ่มใหม่] สำหรับเก็บชื่อหมวดหมู่ปัญหาเวลา JOIN ตาราง problem_types
	Description    string    `json:"description"`
	TechnicianNote string    `json:"technician_note"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
}

// [เพิ่มใหม่] ตาราง repair_images
type RepairImage struct {
	ID        int    `json:"id"`
	RepairID  int    `json:"repair_id"`
	ImageURL  string `json:"image_url"`
	ImageType string `json:"image_type"` // 'before' หรือ 'after'
}
