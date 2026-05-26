package main

import (
	"fmt"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"backend/database"
	"backend/routes"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		fmt.Println("No .env file found (using system env)")
	}

	for {
		err := database.ConnectDB()
		if err == nil {
			break
		}

		fmt.Println("Waiting for DB...")
		time.Sleep(3 * time.Second)
	}

	r := gin.Default()

	r.Static("/uploads", "./uploads")

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://192.168.0.11:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	routes.SetupRoutes(r)

	r.Run(":8080")
}
