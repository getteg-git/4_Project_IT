package equipment

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"

	"backend/database" // เปลี่ยน path ให้ตรงกับโปรเจกต์ของคุณ
	"backend/models"
	"backend/utils"
)

// ImportExcelHandler จัดการรับไฟล์ Excel และประมวลผล
func ImportExcelHandler(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "กรุณาอัปโหลดไฟล์ Excel (.xlsx)"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถเปิดไฟล์ได้"})
		return
	}
	defer file.Close()

	f, err := excelize.OpenReader(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "รูปแบบไฟล์ Excel ไม่ถูกต้อง"})
		return
	}
	defer f.Close()

	sheetName := f.GetSheetName(0) // ดึงข้อมูลจากชีตแรกเสมอ
	rows, err := f.GetRows(sheetName)
	if err != nil || len(rows) < 2 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่พบข้อมูลในไฟล์ Excel"})
		return
	}

	// ---------------------------------------------------------
	// 🔥 สร้าง Dynamic Mapping: จับคู่ "ชื่อคอลัมน์" กับ "หมายเลข Index"
	// ---------------------------------------------------------
	headerMap := make(map[string]int)
	for i, colName := range rows[0] {
		cleanName := strings.TrimSpace(colName)
		if _, exists := headerMap[cleanName]; !exists && cleanName != "" {
			headerMap[cleanName] = i
		}
	}

	// ฟังก์ชัน Helper 1: สำหรับดึงข้อมูล Text ทั่วไป
	getCellValue := func(row []string, colName string) string {
		idx, exists := headerMap[colName]
		if !exists || idx >= len(row) {
			return "" // ถ้าหาคอลัมน์ไม่เจอ หรือแหว่ง ให้คืนค่าว่าง
		}
		return strings.TrimSpace(row[idx])
	}

	// ฟังก์ชัน Helper 2: สำหรับดึงตัวเลข (เคลียร์เครื่องหมาย , ออกก่อนแปลง)
	parseFloatSafe := func(val string) float64 {
		cleanVal := strings.ReplaceAll(val, ",", "")
		f, _ := strconv.ParseFloat(cleanVal, 64)
		return f
	}

	// ฟังก์ชัน Helper 3: สำหรับดึงจำนวนเต็ม (อายุการใช้งาน)
	parseIntSafe := func(val string) *int {
		if val == "" {
			return nil
		}
		cleanVal := strings.ReplaceAll(val, ",", "")
		i, err := strconv.Atoi(cleanVal)
		if err != nil {
			return nil
		}
		return &i
	}

	// ฟังก์ชัน Helper 4: สำหรับแปลงวันที่จาก Excel (M/DD/YYYY) เป็นรูปแบบ DB (YYYY-MM-DD)
	parseDateSafe := func(val string) *string {
		if val == "" {
			return nil
		}
		// ลอง Parse แบบ M/D/YYYY หรือ MM/DD/YYYY
		if t, err := time.Parse("1/2/2006", val); err == nil {
			formatted := t.Format("2006-01-02")
			return &formatted
		}
		// เผื่อข้อมูลมาเป็น YYYY-MM-DD ตรงๆ อยู่แล้ว
		if t, err := time.Parse("2006-01-02", val); err == nil {
			formatted := t.Format("2006-01-02")
			return &formatted
		}
		return nil
	}

	var successCount int
	var errorList []string

	// ---------------------------------------------------------
	// 🚀 เริ่มวนลูปอ่านข้อมูลทีละบรรทัด (เริ่มที่บรรทัดข้อมูล row[1])
	// ---------------------------------------------------------
	for i := 1; i < len(rows); i++ {
		row := rows[i]

		// 1. ดึงคีย์หลัก (ถ้ารหัสครุภัณฑ์ว่าง ให้ข้ามบรรทัดนี้ไปเลย)
		assetCode := getCellValue(row, "เลขที่สินค้าคงคลัง")
		if assetCode == "" {
			continue
		}

		// 2. แพ็กข้อมูลใส่ Struct โดยใช้ Helper ที่เราสร้างไว้เพื่อความชัวร์
		raw := models.RawExcelData{
			AssetCode:               assetCode,
			Name:                    getCellValue(row, "คำอธิบายของสินทรัพย์"),
			BasePrice:               parseFloatSafe(getCellValue(row, "มูลค่าการได้มา")),
			AccumulatedDepreciation: parseFloatSafe(getCellValue(row, "ค่าเสื่อมสะสม")),
			BookValue:               parseFloatSafe(getCellValue(row, "มูลค่าตามบัญชี")),
			RawLocationText:         getCellValue(row, "ที่ตั้งสินทรัพย์ถาวร (ครุภัณฑ์)"),
			ExpectedLifeYears:       parseIntSafe(getCellValue(row, "อายุการใช้งาน")),
			AcquiredDate:            parseDateSafe(getCellValue(row, "วันที่ได้มาครั้งแรก")),
		}

		// 3. จัดการ Pointer สำหรับฟิลด์ที่เป็น Text และปล่อยว่างได้
		if val := getCellValue(row, "เลขที่ผลิตภัณฑ์"); val != "" {
			raw.SerialNumber = &val
		}
		if val := getCellValue(row, "ชื่อกลุ่มสินทรัพย์"); val != "" {
			raw.Category = &val
		}
		if val := getCellValue(row, "สถานะของสินทรัพย์"); val != "" {
			raw.Status = &val
		}

		// 4. โยนข้อความสถานที่เข้าสมองกล (Data Parser) กรอง 2 ชั้น
		parsedLocation := utils.ParseSmartLocation(raw.RawLocationText)

		// 5. ส่งเข้า Database (Find or Create & Upsert)
		err := database.ProcessEquipmentImport(raw, parsedLocation)
		if err != nil {
			errorList = append(errorList, fmt.Sprintf("รหัส %s: %v", assetCode, err.Error()))
			continue
		}
		successCount++
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "นำเข้าข้อมูลเสร็จสิ้น",
		"success_count": successCount,
		"errors":        errorList,
	})
}
