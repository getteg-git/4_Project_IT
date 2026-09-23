package models

import "time"

type User struct {
	ID             int       `json:"id"`
	Username       string    `json:"username"`
	Password       string    `json:"password,omitempty"`
	FullName       string    `json:"full_name"`
	Email          string    `json:"email"` // 🔥 [เพิ่มใหม่] อีเมล
	Role           string    `json:"role"`
	DepartmentID   *int      `json:"department_id"`             // 🔥 [เพิ่มใหม่] ใช้ *int เพื่อรองรับค่า null (สำหรับแอดมิน)
	DepartmentName string    `json:"department_name,omitempty"` // 🔥 [แถมให้] เอาไว้เก็บชื่อสาขาตอน JOIN ตารางส่งให้ Frontend
	IsCentral      bool      `json:"is_central"`                // 🔥 [เพิ่มใหม่] สถานะรับจบ (ช่างส่วนกลาง)
	IsActive       bool      `json:"is_active"`                 // 🔥 [เพิ่มใหม่] สถานะการใช้งาน
	Specialties    []int     `json:"specialties,omitempty"`
	SpecialtyNames []string  `json:"specialty_names,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}
