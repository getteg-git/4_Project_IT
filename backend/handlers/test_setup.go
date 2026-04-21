package handlers

import (
	"log"

	"backend/database"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func setupTest() *gin.Engine {

	// โหลด env
	err := godotenv.Load("../.env")
	if err != nil {
		log.Println("No .env file found")
	}

	// connect DB (test DB)
	err = database.InitTestDB()
	if err != nil {
		panic(err)
	}

	gin.SetMode(gin.TestMode)

	r := gin.Default()

	// Auth
	r.POST("/login", Login)
	r.POST("/register", Register)

	// Users
	r.GET("/users", GetUsers)
	r.GET("/users/search", SearchUsers)
	r.POST("/users", CreateUser)
	r.PUT("/users/:id", UpdateUser)
	r.DELETE("/users/:id", DeleteUser)

	return r
}
