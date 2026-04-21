package handlers

import (
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"backend/database"
	"backend/models"
)

func GetUsers(c *gin.Context) {

	rows, err := database.DB.Query("SELECT id, username, password FROM users")
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var users []models.User

	for rows.Next() {
		var u models.User
		rows.Scan(&u.ID, &u.Username, &u.Password)
		users = append(users, u)
	}

	c.JSON(200, users)
}

func CreateUser(c *gin.Context) {

	var user models.User

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(400, gin.H{"error": "bad request"})
		return
	}

	_, err := database.DB.Exec(
		"INSERT INTO users (username, password) VALUES ($1,$2)",
		user.Username,
		user.Password,
	)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, gin.H{"message": "user created"})
}

func UpdateUser(c *gin.Context) {

	id := c.Param("id")

	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "bad request"})
		return
	}

	// 🔥 hash password (สำคัญมาก)
	hashed, err := bcrypt.GenerateFromPassword(
		[]byte(body.Password),
		bcrypt.DefaultCost,
	)

	if err != nil {
		c.JSON(500, gin.H{"error": "password hash failed"})
		return
	}

	result, err := database.DB.Exec(
		"UPDATE users SET username=$1, password=$2 WHERE id=$3",
		body.Username,
		string(hashed),
		id,
	)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	rows, _ := result.RowsAffected()

	if rows == 0 {
		c.JSON(404, gin.H{"error": "user not found"})
		return
	}

	c.JSON(200, gin.H{"message": "user updated"})
}

func DeleteUser(c *gin.Context) {

	id := c.Param("id")

	result, err := database.DB.Exec(
		"DELETE FROM users WHERE id=$1",
		id,
	)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	rows, _ := result.RowsAffected()

	if rows == 0 {
		c.JSON(404, gin.H{"error": "user not found"})
		return
	}

	c.JSON(200, gin.H{"message": "user deleted"})
}

func Login(c *gin.Context) {

	var user models.User

	// รับ JSON
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(400, gin.H{"error": "bad request"})
		return
	}

	var dbUser models.User

	// ดึงข้อมูลจาก DB (เพิ่ม role)
	err := database.DB.QueryRow(
		"SELECT id, username, password, role FROM users WHERE username=$1",
		user.Username,
	).Scan(
		&dbUser.ID,
		&dbUser.Username,
		&dbUser.Password,
		&dbUser.Role,
	)

	// ถ้าไม่เจอ user
	if err != nil {
		c.JSON(401, gin.H{"error": "user not found"})
		return
	}

	// เช็ก password
	if user.Password != dbUser.Password {
		c.JSON(401, gin.H{"error": "wrong password"})
		return
	}

	// Login ผ่าน
	c.JSON(200, gin.H{
		"message": "login success",
		"user": gin.H{
			"id":       dbUser.ID,
			"username": dbUser.Username,
			"role":     dbUser.Role,
		},
	})
}

func Register(c *gin.Context) {

	var user models.User

	// รับ JSON
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(400, gin.H{"error": "bad request"})
		return
	}

	// เช็กซ้ำ username
	var count int
	err := database.DB.QueryRow(
		"SELECT COUNT(*) FROM users WHERE username=$1",
		user.Username,
	).Scan(&count)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	if count > 0 {
		c.JSON(400, gin.H{"error": "username already exists"})
		return
	}

	// Insert
	_, err = database.DB.Exec(
		"INSERT INTO users (username, password) VALUES ($1,$2)",
		user.Username,
		user.Password,
	)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, gin.H{
		"message": "register success",
	})
}

func SearchUsers(c *gin.Context) {
	keyword := c.Query("q") // รับค่าจาก ?q=

	if keyword == "" {
		c.JSON(400, gin.H{
			"error": "query is required",
		})
		return
	}

	users, err := database.SearchUsers(keyword)
	if err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(200, users)
}
