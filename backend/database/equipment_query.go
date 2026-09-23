package database

import (
	"database/sql"
	"fmt"
	"strings"

	"backend/models"
	"backend/utils" // นำเข้า ParsedLocation
)

// DB เป็นตัวแปร Global ที่เชื่อมต่อฐานข้อมูล (คุณต้องมั่นใจว่าเชื่อมต่อแล้ว)
// var DB *sql.DB

// ProcessEquipmentImport จัดการนำเข้าข้อมูล 1 แถวลง Database
func ProcessEquipmentImport(raw models.RawExcelData, loc utils.ParsedLocation) error {
	// เริ่ม Transaction เพื่อความปลอดภัย
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // ถ้า Error กลางคัน จะย้อนกลับข้อมูลทั้งหมด

	var roomID *int // ถ้าหาสถานที่ไม่เจอ จะปล่อยเป็น nil

	// ==========================================
	// 1. จัดการสถานที่ (Find or Create: ตึก -> ชั้น -> ห้อง)
	// ==========================================
	if loc.Building != "" && loc.Floor != "" && loc.Room != "" {
		// 1.1 ตึก (Location)
		var locationID int
		err = tx.QueryRow("SELECT id FROM locations WHERE name = $1", loc.Building).Scan(&locationID)
		if err == sql.ErrNoRows {
			err = tx.QueryRow("INSERT INTO locations (name) VALUES ($1) RETURNING id", loc.Building).Scan(&locationID)
		}
		if err != nil {
			return fmt.Errorf("จัดการข้อมูลตึกผิดพลาด: %v", err)
		}

		// 1.2 ชั้น (Floor)
		var floorID int
		err = tx.QueryRow("SELECT id FROM floors WHERE floor_name = $1 AND location_id = $2", loc.Floor, locationID).Scan(&floorID)
		if err == sql.ErrNoRows {
			err = tx.QueryRow("INSERT INTO floors (floor_name, location_id) VALUES ($1, $2) RETURNING id", loc.Floor, locationID).Scan(&floorID)
		}
		if err != nil {
			return fmt.Errorf("จัดการข้อมูลชั้นผิดพลาด: %v", err)
		}

		// 1.3 ห้อง (Room)
		var rID int
		err = tx.QueryRow("SELECT id FROM rooms WHERE room_number = $1 AND floor_id = $2", loc.Room, floorID).Scan(&rID)
		if err == sql.ErrNoRows {
			err = tx.QueryRow("INSERT INTO rooms (room_number, floor_id) VALUES ($1, $2) RETURNING id", loc.Room, floorID).Scan(&rID)
		}
		if err != nil {
			return fmt.Errorf("จัดการข้อมูลห้องผิดพลาด: %v", err)
		}

		roomID = &rID
	}

	// ==========================================
	// 2. จัดการอุปกรณ์ (Upsert: ซ้ำอัปเดต / ไม่ซ้ำเพิ่ม)
	// ==========================================

	query := `
		INSERT INTO equipments (
			asset_code, name, category, status, serial_number,
			base_price, accumulated_depreciation, book_value, room_id,
			acquired_date, expected_life_years, is_active
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, true
		)
		ON CONFLICT (asset_code) 
		DO UPDATE SET 
			name = EXCLUDED.name,
			category = EXCLUDED.category,
			status = EXCLUDED.status,
			serial_number = EXCLUDED.serial_number,
			base_price = EXCLUDED.base_price,
			accumulated_depreciation = EXCLUDED.accumulated_depreciation,
			book_value = EXCLUDED.book_value,
			room_id = EXCLUDED.room_id,
			acquired_date = EXCLUDED.acquired_date,
			expected_life_years = EXCLUDED.expected_life_years,
			is_active = true
	`

	// ป้องกันกรณี Name เป็นค่าว่าง (บังคับ NOT NULL ใน DB)
	finalName := strings.TrimSpace(raw.Name)
	if finalName == "" {
		finalName = "ไม่ระบุชื่ออุปกรณ์" // ใส่ Default กัน DB พัง
	}

	_, err = tx.Exec(query,
		raw.AssetCode,
		finalName,
		raw.Category,
		raw.Status,
		raw.SerialNumber,
		raw.BasePrice,
		raw.AccumulatedDepreciation,
		raw.BookValue,
		roomID,
		raw.AcquiredDate,      // รับค่า AcquiredDate จาก Excel
		raw.ExpectedLifeYears, // รับค่า ExpectedLifeYears จาก Excel
	)

	if err != nil {
		return fmt.Errorf("ไม่สามารถบันทึกข้อมูลอุปกรณ์ได้: %v", err)
	}

	// บันทึกสำเร็จ ยืนยัน Transaction
	return tx.Commit()
}
