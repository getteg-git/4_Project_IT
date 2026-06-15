package models

import "time"

// ตาราง repairs
type Repair struct {
	ID             int       `json:"id"`
	ReporterEmail  string    `json:"reporter_email"` // เปลี่ยนจาก user_id เป็น email
	TechnicianID   *int      `json:"technician_id"`  // ใช้ Pointer (*) เพราะค่าอาจเป็น Null (ยังไม่จ่ายงาน)
	LocationID     int       `json:"location_id"`
	FloorID        int       `json:"floor_id"` // [เพิ่มใหม่]
	ProblemTypeID  int       `json:"problem_type_id"`
	Description    string    `json:"description"`
	TechnicianNote string    `json:"technician_note"` // [เพิ่มใหม่]
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
