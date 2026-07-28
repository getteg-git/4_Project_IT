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

// [เพิ่มใหม่] ตาราง equipments (อุปกรณ์/ครุภัณฑ์)
type Equipment struct {
	ID        int     `json:"id"`
	RoomID    int     `json:"room_id"`
	AssetCode string  `json:"asset_code"`
	Name      string  `json:"name"`
	Category  string  `json:"category"`
	BasePrice float64 `json:"base_price"`
	IsActive  bool    `json:"is_active"`
}
