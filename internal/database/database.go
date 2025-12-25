package database

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

var DB *sql.DB

func Init(dbPath string) error {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	var err error
	DB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}

	if err = DB.Ping(); err != nil {
		return err
	}
	log.Printf("Database connected: %s", dbPath)

	return createTables()
}

func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

func createTables() error {
	schema := `
    CREATE TABLE IF NOT EXISTS event (
        event_id    INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
        description TEXT NOT NULL,
        token       TEXT NOT NULL,
        event_date  TEXT NOT NULL,
        is_open     INTEGER DEFAULT 0,
        creator     TEXT
    );

    CREATE TABLE IF NOT EXISTS system_log (
        id          INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
        timestamp   DATETIME DEFAULT CURRENT_TIMESTAMP,
        level       TEXT NOT NULL,
        module      TEXT NOT NULL,
        action      TEXT NOT NULL,
        user_id     TEXT,
        event_id    TEXT,
        ip_address  TEXT,
        details     TEXT
    );
    `
	_, err := DB.Exec(schema)
	return err
}
