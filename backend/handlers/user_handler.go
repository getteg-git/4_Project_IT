package handlers // หรือ package user ตามโครงสร้างโฟลเดอร์ของเก็ต

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
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Role     string `json:"role" binding:"required"` // 'admin' หรือ 'technician'
}

type UpdateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// ---------------------------------------------------------
// 1. GetUsers: ดึงรายชื่อผู้ใช้งานทั้งหมด (Admin / Tech)
// ---------------------------------------------------------
func GetUsers(c *gin.Context) {
	rows, err := database.DB.Query("SELECT id, username, role FROM users ORDER BY id ASC")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var users []models.User

	for rows.Next() {
		var u models.User
		rows.Scan(&u.ID, &u.Username, &u.Role)
		users = append(users, u)
	}

	c.JSON(http.StatusOK, users)
}

// ---------------------------------------------------------
// 2. CreateUser: Admin สร้างบัญชีให้ช่าง/แอดมินใหม่
// ---------------------------------------------------------
func CreateUser(c *gin.Context) {
	var req CreateUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ข้อมูลไม่ครบถ้วน กรุณากรอกใหม่"})
		return
	}

	// เข้ารหัสผ่าน (Hash) ก่อนบันทึกลงฐานข้อมูล
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "เข้ารหัสผ่านไม่สำเร็จ"})
		return
	}

	_, err = database.DB.Exec(
		"INSERT INTO users (username, password, role) VALUES ($1,$2,$3)",
		req.Username,
		string(hashed),
		req.Role,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถสร้างผู้ใช้งานได้: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "สร้างผู้ใช้งานสำเร็จ"})
}

// ---------------------------------------------------------
// 3. UpdateUser: แก้ไขข้อมูลผู้ใช้งาน
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

	result, err := database.DB.Exec(
		"UPDATE users SET username=$1, password=$2 WHERE id=$3",
		req.Username,
		string(hashed),
		id,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบผู้ใช้งานที่ต้องการแก้ไข"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "อัปเดตข้อมูลผู้ใช้งานสำเร็จ"})
}

// ---------------------------------------------------------
// 4. DeleteUser: ลบผู้ใช้งาน
// ---------------------------------------------------------
func DeleteUser(c *gin.Context) {
	id := c.Param("id")

	result, err := database.DB.Exec("DELETE FROM users WHERE id=$1", id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบผู้ใช้งานที่ต้องการลบ"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ลบผู้ใช้งานสำเร็จ"})
}

// ---------------------------------------------------------
// 5. Login: เข้าสู่ระบบ (ตัดเงื่อนไขเช็ค LoginRole ออกแล้ว)
// ---------------------------------------------------------
func Login(c *gin.Context) {
	// เพิ่ม ExpectedRole เพื่อรับค่าว่ากำลังพยายามล็อกอินผ่านประตูไหน
	type LoginRequest struct {
		Username     string `json:"username" binding:"required"`
		Password     string `json:"password" binding:"required"`
		ExpectedRole string `json:"expected_role" binding:"required"` // รับค่า 'admin' หรือ 'technician'
	}

	var req LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ข้อมูลไม่ครบถ้วน"})
		return
	}

	var dbUser models.User
	var dbHashedPassword string

	// ค้นหาใน Database
	err := database.DB.QueryRow(
		"SELECT id, username, password, role FROM users WHERE username=$1",
		req.Username,
	).Scan(
		&dbUser.ID,
		&dbUser.Username,
		&dbHashedPassword,
		&dbUser.Role,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "ชื่อผู้ใช้งานหรือรหัสผ่านไม่ถูกต้อง"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 🔥 เช็คประตูทางเข้า: ถ้า Role ใน DB ไม่ตรงกับประตูที่เข้า ให้เตะออกทันที
	if dbUser.Role != req.ExpectedRole {
		c.JSON(http.StatusForbidden, gin.H{"error": "คุณไม่มีสิทธิ์เข้าสู่ระบบผ่านช่องทางนี้"})
		return
	}

	// เช็ครหัสผ่าน
	err = bcrypt.CompareHashAndPassword([]byte(dbHashedPassword), []byte(req.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "ชื่อผู้ใช้งานหรือรหัสผ่านไม่ถูกต้อง"})
		return
	}

	// Login ผ่าน ส่งคืนเฉพาะข้อมูลที่จำเป็น
	c.JSON(http.StatusOK, gin.H{
		"message": "เข้าสู่ระบบสำเร็จ",
		"user": gin.H{
			"id":       dbUser.ID,
			"username": dbUser.Username,
			"role":     dbUser.Role,
		},
	})
}

// ---------------------------------------------------------
// 6. SearchUsers: ค้นหาผู้ใช้งาน (อ้างอิงฟังก์ชัน SearchUsers ในแพ็กเกจ database)
// ---------------------------------------------------------
func SearchUsers(c *gin.Context) {
	keyword := c.Query("q")

	if keyword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "กรุณาระบุคำค้นหา"})
		return
	}

	users, err := database.SearchUsers(keyword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
}
