package database

import (
	"database/sql"
	"fmt"
	"log"
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

	result, err := db.Exec(`
		UPDATE repairs r
		SET admin_note = NULL
		WHERE r.status = 'รอซ่อม'
		  AND r.technician_id IS NULL
		  AND r.admin_note IS NOT NULL
		  AND EXISTS (
			SELECT 1
			FROM repair_logs latest
			WHERE latest.repair_id = r.id
			  AND latest.action = 'REVOKED'
			  AND NOT EXISTS (
				SELECT 1
				FROM repair_logs newer
				WHERE newer.repair_id = latest.repair_id
				  AND (newer.created_at > latest.created_at
					OR (newer.created_at = latest.created_at AND newer.id > latest.id))
			  )
		  )`)
	if err != nil {
		return fmt.Errorf("failed to clear stale notes from revoked repairs: %w", err)
	}
	if affected, err := result.RowsAffected(); err != nil {
		return fmt.Errorf("failed to count repairs with stale revoke notes: %w", err)
	} else if affected > 0 {
		log.Printf("cleared stale review notes from %d revoked repairs", affected)
	}

	DB = db
	fmt.Println("DB Connected Successfully!")

	return nil
}
