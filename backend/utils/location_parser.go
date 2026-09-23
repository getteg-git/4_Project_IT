package utils

import (
	"regexp"
	"strings"
)

// ParsedLocation คือผลลัพธ์ที่จะส่งกลับไปให้ Database นำไป Find or Create
type ParsedLocation struct {
	Building string
	Floor    string
	Room     string
}

// ParseSmartLocation คือฟังก์ชันหลักที่ใช้ลอจิก "กรอง 2 ชั้น"
func ParseSmartLocation(rawText string) ParsedLocation {
	var loc ParsedLocation
	rawText = strings.TrimSpace(rawText)

	if rawText == "" {
		return loc
	}

	// ==========================================
	// 🟢 Tier 1: Explicit Match (ดึงจากข้อความตรงๆ ก่อน)
	// ==========================================

	// 1. หา "อาคาร" (รองรับคำว่า "วิทย์" หรือ "วิทยาศาสตร์")
	// คืนค่าให้ตรงกับชื่อใน Database เช่น "อาคารวิทยาศาสตร์ 1"
	reBuilding := regexp.MustCompile(`อาคาร(วิทย์|วิทยาศาสตร์)\s*([1-5])`)
	matchBldg := reBuilding.FindStringSubmatch(rawText)
	if len(matchBldg) > 2 {
		loc.Building = "อาคารวิทยาศาสตร์ " + matchBldg[2] // matchBldg[2] คือตัวเลขตึก
	}

	// 2. หา "ชั้น" (เช่น "ชั้น 2", "ชั้น 6")
	reFloor := regexp.MustCompile(`ชั้น\s*([1-9])`)
	matchFloor := reFloor.FindStringSubmatch(rawText)
	if len(matchFloor) > 1 {
		loc.Floor = "ชั้น " + matchFloor[1]
	}

	// 3. หา "ห้อง" (รองรับแบบมีทับ เช่น 1227/1 หรือ ห้อง201)
	reRoom := regexp.MustCompile(`ห้อง\s*([0-9A-Za-z/.-]+)`)
	matchRoom := reRoom.FindStringSubmatch(rawText)
	if len(matchRoom) > 1 {
		loc.Room = matchRoom[1]
	} else {
		// ถ้าไม่มีคำว่าห้อง ลองหาตัวเลข 3-4 หลักที่โผล่มาโดดๆ
		reNumbersOnly := regexp.MustCompile(`\b([0-9]{3,4}[/0-9]*)\b`)
		matchNum := reNumbersOnly.FindStringSubmatch(rawText)
		if len(matchNum) > 1 {
			loc.Room = matchNum[1]
		}
	}

	// ==========================================
	// 🟡 Tier 2: Smart Decoder (ทำงานเมื่อขาดตึก หรือ ขาดชั้น)
	// ==========================================

	// ดึงเฉพาะตัวเลข 3 หรือ 4 ตัวแรกมาถอดรหัส (เช่น จาก "1227/1" เอาแค่ "1227")
	reExtractDigits := regexp.MustCompile(`^([0-9]{3,4})`)
	matchDigits := reExtractDigits.FindStringSubmatch(loc.Room)

	if (loc.Building == "" || loc.Floor == "") && len(matchDigits) > 1 {
		coreNumber := matchDigits[1] // ได้เลข 3 หรือ 4 หลักล้วนๆ

		// เคสเลข 4 หลัก (เช่น 5304, 1227)
		if len(coreNumber) == 4 {
			firstDigit := string(coreNumber[0])  // ตัวแรก (ตึก)
			secondDigit := string(coreNumber[1]) // ตัวที่สอง (ชั้น)

			// ถอดรหัสตึก (ถ้าของเดิม Tier 1 หาไม่เจอ)
			if loc.Building == "" {
				switch firstDigit {
				case "1":
					loc.Building = "อาคารวิทยาศาสตร์ 1"
				case "2":
					loc.Building = "อาคารวิทยาศาสตร์ 2"
				case "3":
					loc.Building = "อาคารวิทยาศาสตร์ 3"
				case "4", "5":
					loc.Building = "อาคารวิทยาศาสตร์ 4" // เลข 5 คือวิทย์ 4 ตามที่คุณกำหนด
				}
			}

			// ถอดรหัสชั้น (ถ้าของเดิม Tier 1 หาไม่เจอ)
			if loc.Floor == "" {
				loc.Floor = "ชั้น " + secondDigit
			}
		}

		// เคสเลข 3 หลัก (เช่น 201)
		if len(coreNumber) == 3 {
			firstDigit := string(coreNumber[0]) // ตัวแรก (ชั้น)

			// ถอดรหัสชั้น (ถ้าของเดิม Tier 1 หาไม่เจอ)
			if loc.Floor == "" {
				loc.Floor = "ชั้น " + firstDigit
			}
			// (ส่วนตึก เราถือว่า 3 หลัก มักจะมีบอกตึกมาในข้อความแล้ว ระบบ Tier 1 จะดึงมาให้เอง)
		}
	}

	return loc
}
