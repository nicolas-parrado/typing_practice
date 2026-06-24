package db

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

// InitDB initializes the SQLite database, creating tables if they do not exist
func InitDB() *sql.DB {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data/typing.db"
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Fatalf("Error creating database directory: %v", err)
	}

	var err error
	DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("Error opening SQLite database: %v", err)
	}

	// Set connection limits
	DB.SetMaxOpenConns(1) // SQLite works best with 1 writer connection

	createTables()
	log.Printf("Database initialized at: %s", dbPath)
	return DB
}

func createTables() {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS profiles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE NOT NULL,
			xp INTEGER DEFAULT 0,
			level INTEGER DEFAULT 1,
			streak INTEGER DEFAULT 0,
			last_active TEXT DEFAULT '',
			theme TEXT DEFAULT 'glass',
			sound_enabled INTEGER DEFAULT 1,
			switch_type TEXT DEFAULT 'blue',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,

		`CREATE TABLE IF NOT EXISTS exercises (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			category TEXT NOT NULL,
			difficulty TEXT NOT NULL,
			is_endurance INTEGER DEFAULT 0
		);`,

		`CREATE TABLE IF NOT EXISTS sessions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			profile_id INTEGER NOT NULL,
			exercise_id TEXT NOT NULL,
			mode TEXT NOT NULL,
			wpm REAL NOT NULL,
			accuracy REAL NOT NULL,
			duration_seconds INTEGER NOT NULL,
			raw_data_json TEXT DEFAULT '',
			completed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(profile_id) REFERENCES profiles(id) ON DELETE CASCADE
		);`,

		`CREATE TABLE IF NOT EXISTS failed_attempts (
			profile_id INTEGER NOT NULL,
			exercise_id TEXT NOT NULL,
			wpm REAL NOT NULL,
			accuracy REAL NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY(profile_id, exercise_id),
			FOREIGN KEY(profile_id) REFERENCES profiles(id) ON DELETE CASCADE
		);`,

		`CREATE TABLE IF NOT EXISTS key_metrics (
			profile_id INTEGER NOT NULL,
			char TEXT NOT NULL,
			attempts INTEGER DEFAULT 0,
			errors INTEGER DEFAULT 0,
			avg_latency_ms INTEGER DEFAULT 0,
			PRIMARY KEY(profile_id, char),
			FOREIGN KEY(profile_id) REFERENCES profiles(id) ON DELETE CASCADE
		);`,

		`CREATE TABLE IF NOT EXISTS achievements (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			profile_id INTEGER NOT NULL,
			code TEXT NOT NULL,
			unlocked_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(profile_id, code),
			FOREIGN KEY(profile_id) REFERENCES profiles(id) ON DELETE CASCADE
		);`,
	}

	for _, query := range queries {
		if _, err := DB.Exec(query); err != nil {
			log.Fatalf("Error creating table: %v\nQuery: %s", err, query)
		}
	}

	// Migrations for existing databases (ignore error if columns already exist)
	_, _ = DB.Exec("ALTER TABLE profiles ADD COLUMN theme TEXT DEFAULT 'glass'")
	_, _ = DB.Exec("ALTER TABLE profiles ADD COLUMN sound_enabled INTEGER DEFAULT 1")
	_, _ = DB.Exec("ALTER TABLE profiles ADD COLUMN switch_type TEXT DEFAULT 'blue'")
}
