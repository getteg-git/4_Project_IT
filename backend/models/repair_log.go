package models

import "time"

type RepairLog struct {
	ID        int       `json:"id"`
	RepairID  int       `json:"repair_id"`
	UserID    *int      `json:"user_id"`
	UserName  string    `json:"user_name,omitempty"` // แถมชื่อคนทำรายการ
	Action    string    `json:"action"`              // เช่น 'CREATED', 'STATUS_CHANGED'
	OldStatus *string   `json:"old_status"`
	NewStatus *string   `json:"new_status"`
	Note      *string   `json:"note"` // ข้อความอธิบาย
	CreatedAt time.Time `json:"created_at"`
}
