package database

import (
	"fmt"
	"os"
)

func InitTestDB() error {

	dsn := os.Getenv("DB_URL")

	if dsn == "" {
		return fmt.Errorf("DB_URL not found")
	}

	return ConnectDB()
}
