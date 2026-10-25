package database

import (
	"net/url"
	"strings"
	"testing"
)

func clearDatabaseEnvironment(t *testing.T) {
	t.Helper()
	for _, name := range []string{
		"DATABASE_URL", "DB_URL", "PGHOST", "PGPORT", "PGUSER", "PGPASSWORD", "PGDATABASE", "PGSSLMODE",
	} {
		t.Setenv(name, "")
	}
}

func TestDatabaseDSNUsesDatabaseURL(t *testing.T) {
	clearDatabaseEnvironment(t)
	t.Setenv("DATABASE_URL", "postgres://user:secret@db.example:5432/app?sslmode=require")
	t.Setenv("DB_URL", "postgres://ignored:ignored@db.example/ignored")

	dsn, err := databaseDSN()
	if err != nil {
		t.Fatalf("databaseDSN returned an error: %v", err)
	}
	if dsn != "postgres://user:secret@db.example:5432/app?sslmode=require" {
		t.Errorf("databaseDSN = %q", dsn)
	}
}

func TestDatabaseDSNSupportsDBURL(t *testing.T) {
	clearDatabaseEnvironment(t)
	t.Setenv("DB_URL", "postgres://user:secret@db.example:5432/app?sslmode=require")

	dsn, err := databaseDSN()
	if err != nil {
		t.Fatalf("databaseDSN returned an error: %v", err)
	}
	if !strings.Contains(dsn, "/app") {
		t.Errorf("databaseDSN = %q, want database path app", dsn)
	}
}

func TestDatabaseDSNBuildsFromPostgresVariables(t *testing.T) {
	clearDatabaseEnvironment(t)
	t.Setenv("PGHOST", "db.example")
	t.Setenv("PGPORT", "5432")
	t.Setenv("PGUSER", "app user")
	t.Setenv("PGPASSWORD", "pa:ss word")
	t.Setenv("PGDATABASE", "repair db")

	dsn, err := databaseDSN()
	if err != nil {
		t.Fatalf("databaseDSN returned an error: %v", err)
	}
	parsed, err := url.Parse(dsn)
	if err != nil {
		t.Fatalf("parse generated DSN: %v", err)
	}
	if parsed.Host != "db.example:5432" || parsed.Path != "/repair db" {
		t.Errorf("unexpected host or database path: %q %q", parsed.Host, parsed.Path)
	}
	if got, _ := parsed.User.Password(); got != "pa:ss word" {
		t.Errorf("database password was not URL encoded correctly")
	}
	if parsed.Query().Get("sslmode") != "require" {
		t.Errorf("sslmode = %q, want require", parsed.Query().Get("sslmode"))
	}
}

func TestDatabaseDSNRequiresConnectionConfiguration(t *testing.T) {
	clearDatabaseEnvironment(t)
	_, err := databaseDSN()
	if err == nil {
		t.Fatal("expected missing database configuration error")
	}
	if !strings.Contains(err.Error(), "DATABASE_URL or DB_URL") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestDatabaseDSNRejectsInvalidPort(t *testing.T) {
	clearDatabaseEnvironment(t)
	t.Setenv("PGHOST", "db.example")
	t.Setenv("PGPORT", "invalid")
	t.Setenv("PGUSER", "app")
	t.Setenv("PGPASSWORD", "secret")
	t.Setenv("PGDATABASE", "app")

	_, err := databaseDSN()
	if err == nil || !strings.Contains(err.Error(), "PGPORT") {
		t.Fatalf("expected invalid PGPORT error, got: %v", err)
	}
}
