package models

import "time"

// ตาราง equipments (อุปกรณ์/ครุภัณฑ์)
type Equipment struct {
	ID                      int        `json:"id"`
	RoomID                  *int       `json:"room_id"` // ใช้ *int เพื่อรองรับค่า NULL ได้
	AssetCode               string     `json:"asset_code"`
	SerialNumber            *string    `json:"serial_number"` // เพิ่มใหม่
	Name                    string     `json:"name"`
	Category                *string    `json:"category"`            // ใช้ *string เพราะอาจจะไม่มีข้อมูล
	Status                  *string    `json:"status"`              // เพิ่มใหม่
	AcquiredDate            *time.Time `json:"acquired_date"`       // เพิ่มใหม่
	ExpectedLifeYears       *int       `json:"expected_life_years"` // เพิ่มใหม่
	BasePrice               float64    `json:"base_price"`
	AccumulatedDepreciation float64    `json:"accumulated_depreciation"` // เพิ่มใหม่
	BookValue               float64    `json:"book_value"`               // เพิ่มใหม่
	IsActive                bool       `json:"is_active"`
}

type RawExcelData struct {
	AssetCode               string
	SerialNumber            *string
	Name                    string
	Category                *string
	Status                  *string
	BasePrice               float64
	AccumulatedDepreciation float64
	BookValue               float64
	RawLocationText         string
	ExpectedLifeYears       *int    // เพิ่มบรรทัดนี้
	AcquiredDate            *string // เพิ่มบรรทัดนี้
}
