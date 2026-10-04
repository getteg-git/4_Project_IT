package database

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func ConnectDB() error {
	// 🛠️ แก้ไข: เปลี่ยนชื่อตัวแปรให้ตรงกับใน Railway
	dsn := os.Getenv("DATABASE_URL")

	if dsn == "" {
		return fmt.Errorf("DATABASE_URL environment variable is not set")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	DB = db
	fmt.Println("DB Connected Successfully!")

	return nil
}
