package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func ConnectDB() error {
	dsn, err := databaseDSN()
	if err != nil {
		return err
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}

	connected := false
	defer func() {
		if !connected {
			db.Close()
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
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
	connected = true
	fmt.Println("DB Connected Successfully!")

	return nil
}

func databaseDSN() (string, error) {
	for _, name := range []string{"DATABASE_URL", "DB_URL"} {
		if dsn := strings.TrimSpace(os.Getenv(name)); dsn != "" {
			return dsn, nil
		}
	}

	host := strings.TrimSpace(os.Getenv("PGHOST"))
	port := strings.TrimSpace(os.Getenv("PGPORT"))
	user := strings.TrimSpace(os.Getenv("PGUSER"))
	password := os.Getenv("PGPASSWORD")
	databaseName := strings.TrimSpace(os.Getenv("PGDATABASE"))
	missing := make([]string, 0, 5)
	for _, variable := range []struct {
		name  string
		value string
	}{
		{name: "PGHOST", value: host},
		{name: "PGPORT", value: port},
		{name: "PGUSER", value: user},
		{name: "PGPASSWORD", value: password},
		{name: "PGDATABASE", value: databaseName},
	} {
		if variable.value == "" {
			missing = append(missing, variable.name)
		}
	}
	if len(missing) > 0 {
		return "", fmt.Errorf("database is not configured: set DATABASE_URL or DB_URL, or provide PostgreSQL variables (missing %s)", strings.Join(missing, ", "))
	}

	portNumber, err := strconv.Atoi(port)
	if err != nil || portNumber < 1 || portNumber > 65535 {
		return "", fmt.Errorf("invalid PGPORT: expected a port between 1 and 65535")
	}

	query := url.Values{}
	sslMode := strings.TrimSpace(os.Getenv("PGSSLMODE"))
	if sslMode == "" {
		sslMode = "require"
	}
	query.Set("sslmode", sslMode)
	dsn := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(user, password),
		Host:     net.JoinHostPort(host, port),
		Path:     "/" + databaseName,
		RawQuery: query.Encode(),
	}
	return dsn.String(), nil
}
