package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/rs/zerolog/log"
)

// MigrationRecord tracks applied database migrations
type MigrationRecord struct {
	ID        int    `db:"id"`
	Filename  string `db:"filename"`
	AppliedAt string `db:"applied_at"`
}

// createMigrationsTable creates the migrations tracking table if it doesn't exist
func (m *MariaDB) createMigrationsTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			id INT AUTO_INCREMENT PRIMARY KEY,
			filename VARCHAR(255) NOT NULL UNIQUE,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_applied_at (applied_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
	`
	_, err := m.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}
	return nil
}

// getAppliedMigrations returns a list of already applied migration filenames
func (m *MariaDB) getAppliedMigrations() (map[string]bool, error) {
	var records []MigrationRecord
	query := `SELECT filename FROM schema_migrations ORDER BY id`
	err := m.db.Select(&records, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get applied migrations: %w", err)
	}

	applied := make(map[string]bool)
	for _, record := range records {
		applied[record.Filename] = true
	}
	return applied, nil
}

// recordMigration marks a migration as applied
func (m *MariaDB) recordMigration(filename string) error {
	query := `INSERT INTO schema_migrations (filename) VALUES (?)`
	_, err := m.db.Exec(query, filename)
	if err != nil {
		return fmt.Errorf("failed to record migration %s: %w", filename, err)
	}
	return nil
}

// RunMigrations automatically applies all pending SQL migrations
func (m *MariaDB) RunMigrations(migrationsDir string) error {
	log.Info().Str("dir", migrationsDir).Msg("Starting auto-migration")

	// Create migrations tracking table
	if err := m.createMigrationsTable(); err != nil {
		return err
	}

	// Get list of applied migrations
	applied, err := m.getAppliedMigrations()
	if err != nil {
		return err
	}

	// Read all .sql files from migrations directory
	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.sql"))
	if err != nil {
		return fmt.Errorf("failed to read migration files: %w", err)
	}

	// Sort files alphabetically (ensures order: 001, 002, 003, etc.)
	sort.Strings(files)

	migrationsRun := 0
	for _, file := range files {
		filename := filepath.Base(file)

		// Skip if already applied
		if applied[filename] {
			log.Debug().Str("file", filename).Msg("Migration already applied, skipping")
			continue
		}

		log.Info().Str("file", filename).Msg("Applying migration")

		// Read migration file
		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", filename, err)
		}

		// Execute migration SQL
		if err := m.executeMigration(string(content)); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", filename, err)
		}

		// Record as applied
		if err := m.recordMigration(filename); err != nil {
			return err
		}

		migrationsRun++
		log.Info().Str("file", filename).Msg("Migration applied successfully")
	}

	if migrationsRun == 0 {
		log.Info().Msg("No pending migrations")
	} else {
		log.Info().Int("count", migrationsRun).Msg("Migrations completed")
	}

	return nil
}

// executeMigration runs a migration SQL file
func (m *MariaDB) executeMigration(sql string) error {
	// Split by semicolon to handle multiple statements
	statements := strings.Split(sql, ";")

	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" || strings.HasPrefix(stmt, "--") {
			continue
		}

		_, err := m.db.Exec(stmt)
		if err != nil {
			return fmt.Errorf("failed to execute statement: %w\nSQL: %s", err, stmt)
		}
	}

	return nil
}
