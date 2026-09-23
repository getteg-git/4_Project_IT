package models

// ตาราง locations (อาคาร)
type Location struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// ตาราง floors (ชั้น)
type Floor struct {
	ID         int    `json:"id"`
	LocationID int    `json:"location_id"`
	FloorName  string `json:"floor_name"`
}

// [เพิ่มใหม่] ตาราง rooms (ห้อง)
type Room struct {
	ID         int    `json:"id"`
	FloorID    int    `json:"floor_id"`
	RoomNumber string `json:"room_number"`
}
