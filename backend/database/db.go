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
	dsn := os.Getenv("DB_URL")

	if dsn == "" {
		return fmt.Errorf("DB_URL environment variable is not set")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	_, err = db.Exec(`
		DO $$
		BEGIN
			IF EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_schema = 'public' AND table_name = 'users' AND column_name = 'is_central'
			) AND NOT EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_schema = 'public' AND table_name = 'users' AND column_name = 'is_head_technician'
			) THEN
				ALTER TABLE users RENAME COLUMN is_central TO is_head_technician;
			END IF;
		END $$;
	`)
	if err != nil {
		return fmt.Errorf("failed to migrate technician designation: %w", err)
	}

	_, err = db.Exec(`
		INSERT INTO problem_types (name)
		SELECT 'งานคอมพิวเตอร์'
		WHERE NOT EXISTS (
			SELECT 1 FROM problem_types WHERE name ILIKE '%คอมพิวเตอร์%' OR name ILIKE '%computer%'
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to ensure computer problem type exists: %w", err)
	}

	_, err = db.Exec(`
		UPDATE users u
		SET department_id = d.id
		FROM departments d
		JOIN degree_levels dl ON dl.id = d.degree_level_id
		WHERE u.username = 'tech_it'
		  AND u.role = 'technician'
		  AND d.name = 'เทคโนโลยีสารสนเทศ (IT)'
		  AND dl.name = 'ปริญญาตรี'
		  AND u.department_id IS DISTINCT FROM d.id
	`)
	if err != nil {
		return fmt.Errorf("failed to align the seeded IT technician department: %w", err)
	}

	_, err = db.Exec(`
		INSERT INTO technician_specialties (user_id, problem_type_id)
		SELECT u.id, pt.id
		FROM users u
		JOIN problem_types pt ON pt.name ILIKE '%คอมพิวเตอร์%' OR pt.name ILIKE '%computer%'
		WHERE u.username = 'tech_it'
		  AND u.role = 'technician'
		ON CONFLICT DO NOTHING
	`)
	if err != nil {
		return fmt.Errorf("failed to set up computer specialty for the seeded technician: %w", err)
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
