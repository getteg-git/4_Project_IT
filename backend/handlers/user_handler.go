package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"backend/database"
	"backend/models"
)

// ==========================================
// Struct สำหรับรับข้อมูลจาก Frontend (Request)
// ==========================================
type CreateUserRequest struct {
	Username    string `json:"username" binding:"required"`
	Password    string `json:"password" binding:"required"`
	FullName    string `json:"full_name" binding:"required"` // ชื่อแสดงผลจริงภาษาไทย/อังกฤษ
	Role        string `json:"role" binding:"required"`      // 'admin' หรือ 'technician'
	Specialties []int  `json:"specialties"`                  // รายการ ID หมวดหมู่งานที่ถนัด (ส่งมาเป็น Array เช่น [1, 2])
}

type UpdateUserRequest struct {
	Username    string `json:"username" binding:"required"`
	Password    string `json:"password" binding:"required"`  // รหัสผ่านใหม่เพื่อยืนยันการบันทึกข้อมูล
	FullName    string `json:"full_name" binding:"required"` // ชื่อแสดงผลจริงที่ต้องการแก้ไข
	Specialties []int  `json:"specialties"`                  // รายการ ID หมวดหมู่งานที่ถนัดชุดใหม่
}

type LoginRequest struct {
	Username     string `json:"username" binding:"required"`
	Password     string `json:"password" binding:"required"`
	ExpectedRole string `json:"expected_role" binding:"required"` // ประตูทางเข้า 'admin' หรือ 'technician'
}

// ---------------------------------------------------------
// 1. GetUsers: ดึงรายชื่อผู้ใช้งานทั้งหมด พร้อมข้อมูลความถนัดจริง
// ---------------------------------------------------------
func GetUsers(c *gin.Context) {
	// ดึงข้อมูลหลักจากตาราง users
	rows, err := database.DB.Query("SELECT id, username, full_name, role, created_at FROM users ORDER BY id ASC")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถดึงข้อมูลผู้ใช้ได้: " + err.Error()})
		return
	}
	defer rows.Close()

	var users []models.User

	for rows.Next() {
		var u models.User
		// กำหนดค่าเริ่มต้นเป็น Slice เปล่า ป้องกันการส่งค่า null กลับไปที่ React
		u.Specialties = []int{}
		u.SpecialtyNames = []string{}

		err := rows.Scan(&u.ID, &u.Username, &u.FullName, &u.Role, &u.CreatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "เกิดข้อผิดพลาดในการ Scan ข้อมูล: " + err.Error()})
			return
		}

		// ถ้าเป็นช่างเทคนิค ให้ไป Query ดึงข้อมูลความถนัดจริงจาก Junction Table ออกมาด้วย
		if u.Role == "technician" {
			specRows, err := database.DB.Query(`
				SELECT ts.problem_type_id, pt.name 
				FROM technician_specialties ts
				JOIN problem_types pt ON ts.problem_type_id = pt.id
				WHERE ts.user_id = $1`, u.ID)

			if err == nil {
				for specRows.Next() {
					var specID int
					var specName string
					if err := specRows.Scan(&specID, &specName); err == nil {
						u.Specialties = append(u.Specialties, specID)
						u.SpecialtyNames = append(u.SpecialtyNames, specName)
					}
				}
				specRows.Close()
			}
		}

		users = append(users, u)
	}

	c.JSON(http.StatusOK, users)
}

// ---------------------------------------------------------
// 2. CreateUser: Admin สร้างบัญชีผู้ใช้ใหม่ พร้อมบันทึกความถนัดลง Database
// ---------------------------------------------------------
func CreateUser(c *gin.Context) {
	var req CreateUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ข้อมูลไม่ครบถ้วน กรุณากรอกใหม่"})
		return
	}

	// เข้ารหัสผ่าน (Hash)
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "เข้ารหัสผ่านไม่สำเร็จ"})
		return
	}

	// ใช้ระบบ Transaction เพื่อป้องกันข้อมูลตกค้างหรือแหว่งในระบบ
	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถเริ่มกระบวนการฐานข้อมูลได้"})
		return
	}

	var newUserID int
	// บันทึกลงตาราง users และนำ ID ล่าสุดกลับมาใช้ผูกตารางเชื่อม
	err = tx.QueryRow(
		"INSERT INTO users (username, password, full_name, role) VALUES ($1, $2, $3, $4) RETURNING id",
		req.Username, string(hashed), req.FullName, req.Role,
	).Scan(&newUserID)

	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ชื่อผู้ใช้งานนี้มีอยู่ในระบบแล้ว ไม่สามารถสร้างซ้ำได้"})
		return
	}

	// บันทึกความถนัดลงตารางเชื่อมเฉพาะสิทธิ์ช่างเทคนิคเท่านั้น
	if req.Role == "technician" && len(req.Specialties) > 0 {
		// ควบคุมโควตาความถนัดสูงสุดไม่เกิน 3 หมวดหมู่ตามเงื่อนไขที่กำหนดไว้
		if len(req.Specialties) > 3 {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "ช่าง 1 คนสามารถมีความถนัดได้สูงสุด 3 หมวดหมู่เท่านั้น"})
			return
		}

		for _, problemTypeID := range req.Specialties {
			_, err = tx.Exec(
				"INSERT INTO technician_specialties (user_id, problem_type_id) VALUES ($1, $2)",
				newUserID, problemTypeID,
			)
			if err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถบันทึกข้อมูลความถนัดได้: " + err.Error()})
				return
			}
		}
	}

	// ยืนยันกระบวนการบันทึกข้อมูลทั้งหมดพร้อมกัน
	err = tx.Commit()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถยืนยันการบันทึกข้อมูลลงระบบได้"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "สร้างบัญชีผู้ใช้งานใหม่สำเร็จเรียบร้อย"})
}

// ---------------------------------------------------------
// 3. UpdateUser: แก้ไขข้อมูลผู้ใช้และล้าง/อัปเดตความถนัดใหม่ทั้งหมด
// ---------------------------------------------------------
func UpdateUser(c *gin.Context) {
	id := c.Param("id")
	var req UpdateUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ข้อมูลไม่ครบถ้วน"})
		return
	}

	// เข้ารหัสผ่านใหม่
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "เข้ารหัสผ่านไม่สำเร็จ"})
		return
	}

	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถเริ่มกระบวนการฐานข้อมูลได้"})
		return
	}

	// ตรวจสอบสิทธิ์สิทธิ์ดั้งเดิมจาก ID เพื่อนำไปประมวลผลตารางเชื่อมความถนัด
	var role string
	err = tx.QueryRow("SELECT role FROM users WHERE id = $1", id).Scan(&role)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบผู้ใช้งานที่ต้องการแก้ไข"})
		return
	}

	// 1. อัปเดตข้อมูลหลักในตาราง users (อัปเดต username, password, และ full_name)
	_, err = tx.Exec(
		"UPDATE users SET username=$1, password=$2, full_name=$3 WHERE id=$4",
		req.Username, string(hashed), req.FullName, id,
	)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถอัปเดตข้อมูลผู้ใช้งานได้: " + err.Error()})
		return
	}

	// 2. จัดการข้อมูลความถนัดใหม่หากผู้ใช้งานรายนั้นเป็นช่างเทคนิค
	if role == "technician" {
		// เคลียร์ข้อมูลความถนัดเก่าออกให้หมดก่อนเพื่อรออัปเดตชุดใหม่
		_, err = tx.Exec("DELETE FROM technician_specialties WHERE user_id = $1", id)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถรีเซ็ตข้อมูลความถนัดดั้งเดิมได้"})
			return
		}

		// บันทึกความถนัดชุดใหม่เข้าไปแทนที่
		if len(req.Specialties) > 0 {
			if len(req.Specialties) > 3 {
				tx.Rollback()
				c.JSON(http.StatusBadRequest, gin.H{"error": "ช่าง 1 คนสามารถมีความถนัดได้สูงสุด 3 หมวดหมู่เท่านั้น"})
				return
			}

			for _, problemTypeID := range req.Specialties {
				_, err = tx.Exec(
					"INSERT INTO technician_specialties (user_id, problem_type_id) VALUES ($1, $2)",
					id, problemTypeID,
				)
				if err != nil {
					tx.Rollback()
					c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถอัปเดตข้อมูลความถนัดชุดใหม่ได้"})
					return
				}
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถยืนยันการอัปเดตข้อมูลได้"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "อัปเดตข้อมูลผู้ใช้งานและจัดระเบียบความถนัดสำเร็จ"})
}

// ---------------------------------------------------------
// 4. DeleteUser: ลบบัญชีผู้ใช้งาน (ตารางเชื่อมจะลบอัตโนมัติด้วย ON DELETE CASCADE)
// ---------------------------------------------------------
func DeleteUser(c *gin.Context) {
	id := c.Param("id")

	result, err := database.DB.Exec("DELETE FROM users WHERE id=$1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "เกิดข้อผิดพลาดในการลบ: " + err.Error()})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบผู้ใช้งานที่ต้องการลบ"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ลบผู้ใช้งานออกจากระบบสำเร็จ"})
}

// ---------------------------------------------------------
// 5. Login: เข้าสู่ระบบพร้อมส่ง FullName กลับไปแสดงผลแทน Username
// ---------------------------------------------------------
func Login(c *gin.Context) {
	var req LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ข้อมูลไม่ครบถ้วน"})
		return
	}

	var dbUser models.User
	var dbHashedPassword string

	// ค้นหาข้อมูลใน Database โดยดึงฟิลด์ full_name ออกมาด้วย
	err := database.DB.QueryRow(
		"SELECT id, username, password, full_name, role FROM users WHERE username=$1",
		req.Username,
	).Scan(
		&dbUser.ID,
		&dbUser.Username,
		&dbHashedPassword,
		&dbUser.FullName,
		&dbUser.Role,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "ชื่อผู้ใช้งานหรือรหัสผ่านไม่ถูกต้อง"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// ตรวจสอบความถูกต้องของสิทธิ์ทางเข้า
	if dbUser.Role != req.ExpectedRole {
		c.JSON(http.StatusForbidden, gin.H{"error": "คุณไม่มีสิทธิ์เข้าสู่ระบบผ่านช่องทางนี้"})
		return
	}

	// ตรวจสอบความถูกต้องของรหัสผ่าน
	err = bcrypt.CompareHashAndPassword([]byte(dbHashedPassword), []byte(req.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "ชื่อผู้ใช้งานหรือรหัสผ่านไม่ถูกต้อง"})
		return
	}

	// ส่งข้อมูลเฉพาะที่จำเป็นกลับหน้าบ้าน โดยมี full_name ไปโชว์ต้อนรับอย่างสวยงาม
	c.JSON(http.StatusOK, gin.H{
		"message": "เข้าสู่ระบบสำเร็จ",
		"user": gin.H{
			"id":        dbUser.ID,
			"username":  dbUser.Username,
			"full_name": dbUser.FullName, // หน้า React จะนำค่านี่ไปแสดงผลเป็น "ยินดีต้อนรับ, สมชาย"
			"role":      dbUser.Role,
		},
	})
}

// ---------------------------------------------------------
// 6. SearchUsers: ค้นหาผู้ใช้งานจาก Username หรือ ชื่อจริง (ดึงความถนัดจริง ไม่ใช้ Mockup)
// ---------------------------------------------------------
func SearchUsers(c *gin.Context) {
	keyword := c.Query("q")

	if keyword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "กรุณาระบุคำค้นหา"})
		return
	}

	// ค้นหาแบบยืดหยุ่นโดยรองรับทั้งการหาด้วย username หรือชื่อจริง full_name ภาษาไทย
	rows, err := database.DB.Query(
		"SELECT id, username, full_name, role, created_at FROM users WHERE username ILIKE $1 OR full_name ILIKE $1 ORDER BY id ASC",
		"%"+keyword+"%",
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var users []models.User

	for rows.Next() {
		var u models.User
		u.Specialties = []int{}
		u.SpecialtyNames = []string{}

		err := rows.Scan(&u.ID, &u.Username, &u.FullName, &u.Role, &u.CreatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if u.Role == "technician" {
			specRows, err := database.DB.Query(`
				SELECT ts.problem_type_id, pt.name 
				FROM technician_specialties ts
				JOIN problem_types pt ON ts.problem_type_id = pt.id
				WHERE ts.user_id = $1`, u.ID)

			if err == nil {
				for specRows.Next() {
					var specID int
					var specName string
					if err := specRows.Scan(&specID, &specName); err == nil {
						u.Specialties = append(u.Specialties, specID)
						u.SpecialtyNames = append(u.SpecialtyNames, specName)
					}
				}
				specRows.Close()
			}
		}
		users = append(users, u)
	}

	c.JSON(http.StatusOK, users)
}
