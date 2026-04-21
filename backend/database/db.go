package database

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func ConnectDB() error {

	dsn := os.Getenv("DB_URL")

	if dsn == "" {
		return fmt.Errorf("DB_URL not found")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return err
	}

	if err := db.Ping(); err != nil {
		return err
	}

	DB = db

	fmt.Println("DB Connected")

	return nil
}
