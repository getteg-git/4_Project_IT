package user

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/mail"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"backend/database"
	"backend/models"
)

// ==========================================
// Struct สำหรับรับข้อมูลจาก Frontend (Request)
// ==========================================
type CreateUserRequest struct {
	Username     string `json:"username" binding:"required"`
	Password     string `json:"password" binding:"required"`
	FullName     string `json:"full_name" binding:"required"`
	Email        string `json:"email" binding:"required"` // 🔥 [เพิ่มใหม่]
	Role         string `json:"role" binding:"required"`
	DepartmentID *int   `json:"department_id"` // 🔥 [เพิ่มใหม่] (เป็น null ได้สำหรับ admin)
	IsCentral    bool   `json:"is_central"`    // 🔥 [เพิ่มใหม่]
	Specialties  []int  `json:"specialties"`
}

type UpdateUserRequest struct {
	Username     string `json:"username" binding:"required"`
	Password     string `json:"password"`
	FullName     string `json:"full_name" binding:"required"`
	Email        string `json:"email" binding:"required"` // 🔥 [เพิ่มใหม่]
	Role         string `json:"role" binding:"required"`
	DepartmentID *int   `json:"department_id"` // 🔥 [เพิ่มใหม่]
	IsCentral    bool   `json:"is_central"`    // 🔥 [เพิ่มใหม่]
	Specialties  []int  `json:"specialties"`
}

type LoginRequest struct {
	Username     string `json:"username" binding:"required"`
	Password     string `json:"password" binding:"required"`
	ExpectedRole string `json:"expected_role" binding:"required"`
}

var userUsernamePattern = regexp.MustCompile(`^[A-Za-z0-9._-]{3,30}$`)

func validateUserPayload(username, password, fullName, email, role string, departmentID *int, specialties []int, allowEmptyPassword bool) error {
	username = strings.TrimSpace(username)
	password = strings.TrimSpace(password)
	fullName = strings.TrimSpace(fullName)
	email = strings.TrimSpace(email)
	role = strings.ToLower(strings.TrimSpace(role))

	if username == "" {
		return fmt.Errorf("username ต้องไม่ว่าง")
	}
	if !userUsernamePattern.MatchString(username) {
		return fmt.Errorf("username ต้องมีความยาว 3-30 ตัวอักษร และใช้ได้เฉพาะ a-z, A-Z, 0-9, '.', '_', '-'")
	}
	if password != "" {
		if len(password) < 3 || len(password) > 30 {
			return fmt.Errorf("password ต้องมีความยาว 3-30 ตัวอักษร")
		}
	} else if !allowEmptyPassword {
		return fmt.Errorf("password ต้องมีความยาว 3-30 ตัวอักษร")
	}
	if len(fullName) < 2 {
		return fmt.Errorf("ชื่อ-นามสกุลจริง ต้องมีความยาวอย่างน้อย 2 ตัวอักษร")
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return fmt.Errorf("อีเมลไม่ถูกต้อง")
	}
	if role != "admin" && role != "technician" {
		return fmt.Errorf("สิทธิ์ผู้ใช้ต้องเป็น admin หรือ technician")
	}
	if role == "technician" {
		if departmentID == nil {
			return fmt.Errorf("ช่างเทคนิคต้องเลือกสาขาวิชา")
		}
		if len(specialties) == 0 {
			return fmt.Errorf("ช่างเทคนิคต้องมีความถนัดอย่างน้อย 1 หมวดหมู่")
		}
		if len(specialties) > 3 {
			return fmt.Errorf("ช่างเทคนิคสามารถมีความถนัดได้สูงสุด 3 หมวดหมู่")
		}
		seen := make(map[int]bool)
		for _, specialtyID := range specialties {
			if specialtyID <= 0 || seen[specialtyID] {
				return fmt.Errorf("ข้อมูลความถนัดไม่ถูกต้อง")
			}
			seen[specialtyID] = true
		}
	} else if len(specialties) > 0 {
		return fmt.Errorf("แอดมินไม่จำเป็นต้องระบุความถนัด")
	}

	return nil
}

// ---------------------------------------------------------
// 1. GetUsers: ดึงรายชื่อผู้ใช้งานทั้งหมด (ที่ยัง Active)
// ---------------------------------------------------------
func GetUsers(c *gin.Context) {
	// 🔥 [ปรับปรุง] ใช้ LEFT JOIN เพื่อดึงชื่อสาขามาด้วย และกรองเฉพาะคนที่ is_active = true
	query := `
		SELECT u.id, u.username, u.full_name, u.email, u.role, 
		       u.department_id, d.name as department_name, 
		       u.is_central, u.is_active, u.created_at 
		FROM users u
		LEFT JOIN departments d ON u.department_id = d.id
		WHERE u.is_active = true
		ORDER BY u.id ASC
	`
	rows, err := database.DB.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถดึงข้อมูลผู้ใช้ได้: " + err.Error()})
		return
	}
	defer rows.Close()

	users := make([]models.User, 0)

	for rows.Next() {
		var u models.User
		u.Specialties = []int{}
		u.SpecialtyNames = []string{}

		// ใช้ sql.NullInt64 และ sql.NullString มารับค่าที่อาจเป็น NULL จาก Database ป้องกัน Go Error
		var deptID sql.NullInt64
		var deptName sql.NullString

		err := rows.Scan(
			&u.ID, &u.Username, &u.FullName, &u.Email, &u.Role,
			&deptID, &deptName, &u.IsCentral, &u.IsActive, &u.CreatedAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "เกิดข้อผิดพลาดในการ Scan ข้อมูล: " + err.Error()})
			return
		}

		// แปลงค่า Null กลับเป็น Pointer ให้ Struct
		if deptID.Valid {
			idVal := int(deptID.Int64)
			u.DepartmentID = &idVal
		}
		if deptName.Valid {
			u.DepartmentName = deptName.String
		}

		// ดึงข้อมูลความถนัดจริงจาก Junction Table
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
// 2. CreateUser: สร้างบัญชีผู้ใช้ใหม่
// ---------------------------------------------------------
func CreateUser(c *gin.Context) {
	var req CreateUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ข้อมูลไม่ครบถ้วน กรุณากรอกใหม่"})
		return
	}

	if err := validateUserPayload(req.Username, req.Password, req.FullName, req.Email, req.Role, req.DepartmentID, req.Specialties, false); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

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

	var newUserID int
	// 🔥 [ปรับปรุง] เพิ่ม email, department_id, is_central ลงในคำสั่ง INSERT
	query := `
		INSERT INTO users (username, password, full_name, email, role, department_id, is_central) 
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id
	`
	err = tx.QueryRow(query, req.Username, string(hashed), req.FullName, req.Email, req.Role, req.DepartmentID, req.IsCentral).Scan(&newUserID)

	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ชื่อผู้ใช้งานหรืออีเมลนี้มีอยู่ในระบบแล้ว"})
		return
	}

	if req.Role == "technician" && len(req.Specialties) > 0 {
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
				c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถบันทึกข้อมูลความถนัดได้"})
				return
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถยืนยันการบันทึกข้อมูลลงระบบได้"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "สร้างบัญชีผู้ใช้งานใหม่สำเร็จเรียบร้อย"})
}

// ---------------------------------------------------------
// 3. UpdateUser: แก้ไขข้อมูลผู้ใช้
// ---------------------------------------------------------
func UpdateUser(c *gin.Context) {
	id := c.Param("id")
	var req UpdateUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ข้อมูลไม่ครบถ้วน"})
		return
	}

	if err := validateUserPayload(req.Username, req.Password, req.FullName, req.Email, req.Role, req.DepartmentID, req.Specialties, true); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	passwordHash := req.Password
	if req.Password == "" {
		if err := database.DB.QueryRow("SELECT password FROM users WHERE id = $1", id).Scan(&passwordHash); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถดึงรหัสผ่านปัจจุบันได้"})
			return
		}
	} else {
		hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "เข้ารหัสผ่านไม่สำเร็จ"})
			return
		}
		passwordHash = string(hashed)
	}

	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถเริ่มกระบวนการฐานข้อมูลได้"})
		return
	}

	var exists int
	err = tx.QueryRow("SELECT 1 FROM users WHERE id = $1", id).Scan(&exists)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบผู้ใช้งานที่ต้องการแก้ไข"})
		return
	}

	if req.Role == "admin" {
		req.DepartmentID = nil
		req.IsCentral = false
		req.Specialties = nil
	}

	updateQuery := `
		UPDATE users 
		SET username=$1, password=$2, full_name=$3, email=$4, role=$5, department_id=$6, is_central=$7 
		WHERE id=$8
	`
	_, err = tx.Exec(updateQuery, req.Username, passwordHash, req.FullName, req.Email, req.Role, req.DepartmentID, req.IsCentral, id)

	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถอัปเดตข้อมูลผู้ใช้งานได้"})
		return
	}

	_, err = tx.Exec("DELETE FROM technician_specialties WHERE user_id = $1", id)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถรีเซ็ตข้อมูลความถนัดดั้งเดิมได้"})
		return
	}

	if req.Role == "technician" {
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

	err = tx.Commit()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถยืนยันการอัปเดตข้อมูลได้"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "อัปเดตข้อมูลผู้ใช้งานสำเร็จ"})
}

// ---------------------------------------------------------
// 4. DeleteUser: ลบบัญชีผู้ใช้งาน (ปรับเป็น Soft Delete)
// ---------------------------------------------------------
func DeleteUser(c *gin.Context) {
	id := c.Param("id")

	// 🔥 [ปรับปรุง] ใช้เทคนิค Soft Delete เปลี่ยน is_active = false แทนการลบทิ้ง เพื่อรักษาประวัติงานซ่อม
	result, err := database.DB.Exec("UPDATE users SET is_active = false WHERE id=$1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "เกิดข้อผิดพลาดในการระงับบัญชี: " + err.Error()})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบผู้ใช้งานที่ต้องการระงับ"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ระงับบัญชีผู้ใช้งานสำเร็จ"})
}

// ---------------------------------------------------------
// 5. Login: เข้าสู่ระบบ
// ---------------------------------------------------------
func Login(c *gin.Context) {
	var req LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ข้อมูลไม่ครบถ้วน"})
		return
	}

	var dbUser models.User
	var dbHashedPassword string
	var deptID sql.NullInt64

	// 🔥 [ปรับปรุง] ดึง email, department_id, is_central ออกมาด้วย (ต้อง check is_active ด้วย)
	query := `
		SELECT id, username, password, full_name, email, role, department_id, is_central 
		FROM users 
		WHERE username=$1 AND is_active=true
	`
	err := database.DB.QueryRow(query, req.Username).Scan(
		&dbUser.ID,
		&dbUser.Username,
		&dbHashedPassword,
		&dbUser.FullName,
		&dbUser.Email,
		&dbUser.Role,
		&deptID,
		&dbUser.IsCentral,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "ชื่อผู้ใช้งานหรือรหัสผ่านไม่ถูกต้อง หรือบัญชีถูกระงับ"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if deptID.Valid {
		idVal := int(deptID.Int64)
		dbUser.DepartmentID = &idVal
	}

	if dbUser.Role != req.ExpectedRole {
		c.JSON(http.StatusForbidden, gin.H{"error": "คุณไม่มีสิทธิ์เข้าสู่ระบบผ่านช่องทางนี้"})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(dbHashedPassword), []byte(req.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "ชื่อผู้ใช้งานหรือรหัสผ่านไม่ถูกต้อง"})
		return
	}

	// 🔥 [ปรับปรุง] ส่ง department_id และ is_central กลับไปให้ Frontend ใช้ประมวลผลต่อ
	c.JSON(http.StatusOK, gin.H{
		"message": "เข้าสู่ระบบสำเร็จ",
		"user": gin.H{
			"id":            dbUser.ID,
			"username":      dbUser.Username,
			"full_name":     dbUser.FullName,
			"email":         dbUser.Email,
			"role":          dbUser.Role,
			"department_id": dbUser.DepartmentID,
			"is_central":    dbUser.IsCentral,
		},
	})
}

// ---------------------------------------------------------
// 6. SearchUsers: ค้นหาผู้ใช้งาน (อัปเดต Query ให้ตรงกับ GetUsers)
// ---------------------------------------------------------
func SearchUsers(c *gin.Context) {
	keyword := c.Query("q")

	if keyword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "กรุณาระบุคำค้นหา"})
		return
	}

	// 🔥 [ปรับปรุง] อัปเดตโครงสร้างดึงข้อมูลเหมือน GetUsers
	query := `
		SELECT u.id, u.username, u.full_name, u.email, u.role, 
		       u.department_id, d.name as department_name, 
		       u.is_central, u.is_active, u.created_at 
		FROM users u
		LEFT JOIN departments d ON u.department_id = d.id
		WHERE (u.username ILIKE $1 OR u.full_name ILIKE $1) AND u.is_active = true
		ORDER BY u.id ASC
	`
	rows, err := database.DB.Query(query, "%"+keyword+"%")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	users := make([]models.User, 0)

	for rows.Next() {
		var u models.User
		u.Specialties = []int{}
		u.SpecialtyNames = []string{}

		var deptID sql.NullInt64
		var deptName sql.NullString

		err := rows.Scan(
			&u.ID, &u.Username, &u.FullName, &u.Email, &u.Role,
			&deptID, &deptName, &u.IsCentral, &u.IsActive, &u.CreatedAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if deptID.Valid {
			idVal := int(deptID.Int64)
			u.DepartmentID = &idVal
		}
		if deptName.Valid {
			u.DepartmentName = deptName.String
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
