package models

// ตาราง problem_types (ประเภทปัญหา)
type ProblemType struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// [เพิ่มใหม่] ตาราง technician_specialties (ความถนัดของช่าง)
// เก็บไว้เผื่อใช้ทำ API ดึงข้อมูลช่างตามประเภทงาน
type TechnicianSpecialty struct {
	UserID        int `json:"user_id"`
	ProblemTypeID int `json:"problem_type_id"`
}
