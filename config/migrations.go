// config/migrations.go
package config

import (
	"database/sql"
	"log"
)

type Migration struct {
	Version string
	Query   string
}

func RunMigrations(db *sql.DB) {
	migrations := []Migration{
		{
			Version: "20240620_initial",
			Query: `
			CREATE TABLE IF NOT EXISTS emaginenet_blocked_numbers (
				id SERIAL PRIMARY KEY,
				phone_number VARCHAR(20) NOT NULL UNIQUE,
				reason TEXT,
				blocked_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				blocked_by VARCHAR(100),
				is_active BOOLEAN DEFAULT TRUE,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			);

			CREATE TABLE IF NOT EXISTS schema_migrations (
				version VARCHAR(50) PRIMARY KEY,
				applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			);
			`,
		},
		{
			Version: "20250731_remove_unique_constraint",
			Query: `
			-- Remove the unique constraint to allow multiple records for same phone number
			ALTER TABLE emaginenet_blocked_numbers DROP CONSTRAINT IF EXISTS emaginenet_blocked_numbers_phone_number_key;
			
			-- Add index for better query performance
			CREATE INDEX IF NOT EXISTS idx_phone_number_active ON emaginenet_blocked_numbers(phone_number, is_active);
			CREATE INDEX IF NOT EXISTS idx_blocked_date ON emaginenet_blocked_numbers(blocked_date DESC);
			`,
		},
	}

	// Ensure schema_migrations table exists
	_, _ = db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version VARCHAR(50) PRIMARY KEY,
		applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)

	for _, m := range migrations {
		var exists bool
		err := db.QueryRow("SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version = $1)", m.Version).Scan(&exists)
		if err != nil {
			log.Fatalf("Error checking migration version %s: %v", m.Version, err)
		}

		if exists {
			log.Printf("Migration %s already applied. Skipping.", m.Version)
			continue
		}

		_, err = db.Exec(m.Query)
		if err != nil {
			log.Fatalf("Failed to apply migration %s: %v", m.Version, err)
		}

		_, err = db.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", m.Version)
		if err != nil {
			log.Fatalf("Failed to record migration %s: %v", m.Version, err)
		}

		log.Printf("Migration %s applied successfully.", m.Version)
	}
}
