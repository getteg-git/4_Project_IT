package main

import (
	"fmt"
	"log"
	"os" // 🛠️ เพิ่มแพ็กเกจ os สำหรับดึง Environment Variable
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"backend/database"
	"backend/routes"
	"backend/utils"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		fmt.Println("No .env file found (using system env)")
	}

	if err := utils.ValidateSMTPConfiguration(); err != nil {
		log.Printf("email notifications are unavailable: %v; configure SMTP variables on the Railway backend service and redeploy", err)
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
		// 🛠️ แก้ไข: เพิ่ม URL ของ Vercel เข้าไปเพื่อให้เว็บดึงข้อมูลจาก API ได้
		AllowOrigins: []string{
			"http://localhost:5173",
			"http://192.168.0.11:5173",
			"https://test-pro-mu.vercel.app",
			"https://4-project-it.vercel.app",
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	routes.SetupRoutes(r)

	// 🛠️ แก้ไข: ดึง PORT จาก Railway (ถ้าหาไม่เจอให้ใช้ 8080 สำหรับรันในเครื่อง)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Println("Server is running on port:", port)
	r.Run(":" + port)
}
