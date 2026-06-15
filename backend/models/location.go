package models

// ตาราง locations
type Location struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// [เพิ่มใหม่] ตาราง floors
type Floor struct {
	ID         int    `json:"id"`
	LocationID int    `json:"location_id"`
	FloorName  string `json:"floor_name"`
}
