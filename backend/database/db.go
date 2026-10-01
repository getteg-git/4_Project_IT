// package database

// import (
// 	"database/sql"
// 	"fmt"
// 	"os"

// 	_ "github.com/lib/pq"
// )

// var DB *sql.DB

// func ConnectDB() error {
// 	// 🛠️ แก้ไข: เปลี่ยนจาก DB_URL เป็น DATABASE_URL
// 	dsn := os.Getenv("DATABASE_URL")

// 	if dsn == "" {
// 		// (Optional) คุณอาจจะใส่ Default Connection สำหรับทดสอบในเครื่องตัวเองไว้ตรงนี้ได้
// 		// dsn = "postgres://username:password@localhost:5432/yourdb?sslmode=disable"
// 		return fmt.Errorf("DATABASE_URL not found")
// 	}

// 	db, err := sql.Open("postgres", dsn)
// 	if err != nil {
// 		return err
// 	}

// 	if err := db.Ping(); err != nil {
// 		return err
// 	}

// 	DB = db

// 	fmt.Println("DB Connected")

// 	return nil
// }

package database

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func ConnectDB() error {
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
