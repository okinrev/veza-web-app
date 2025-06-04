package database

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/lib/pq"
)

// NewConnection creates a new database connection
func NewConnection(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)

	log.Println("✅ Database connection established")
	return db, nil
}

// RunMigrations runs all SQL migration files in the migrations directory
func RunMigrations(db *sql.DB) error {
	// Create migrations table if it doesn't exist
	if err := createMigrationsTable(db); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get all migration files
	migrationFiles, err := getMigrationFiles("./internal/database/migrations")
	if err != nil {
		return fmt.Errorf("failed to get migration files: %w", err)
	}

	// Get applied migrations
	appliedMigrations, err := getAppliedMigrations(db)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Run pending migrations
	for _, file := range migrationFiles {
		if _, applied := appliedMigrations[file]; !applied {
			log.Printf("Running migration: %s", file)
			
			if err := runMigrationFile(db, filepath.Join("./internal/database/migrations", file)); err != nil {
				return fmt.Errorf("failed to run migration %s: %w", file, err)
			}

			if err := recordMigration(db, file); err != nil {
				return fmt.Errorf("failed to record migration %s: %w", file, err)
			}

			log.Printf("✅ Migration completed: %s", file)
		}
	}

	log.Println("✅ All migrations completed")
	return nil
}

// createMigrationsTable creates the migrations tracking table
func createMigrationsTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS migrations (
			id SERIAL PRIMARY KEY,
			filename VARCHAR(255) NOT NULL UNIQUE,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`
	_, err := db.Exec(query)
	return err
}

// getMigrationFiles returns all .sql files in the migrations directory, sorted
func getMigrationFiles(dir string) ([]string, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var migrationFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".sql") {
			migrationFiles = append(migrationFiles, file.Name())
		}
	}

	// Sort files to ensure consistent migration order
	sort.Strings(migrationFiles)
	return migrationFiles, nil
}

// getAppliedMigrations returns a map of applied migration filenames
func getAppliedMigrations(db *sql.DB) (map[string]bool, error) {
	rows, err := db.Query("SELECT filename FROM migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	appliedMigrations := make(map[string]bool)
	for rows.Next() {
		var filename string
		if err := rows.Scan(&filename); err != nil {
			return nil, err
		}
		appliedMigrations[filename] = true
	}

	return appliedMigrations, rows.Err()
}

// runMigrationFile executes a single migration file
func runMigrationFile(db *sql.DB, filePath string) error {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	// Split by semicolons and execute each statement
	statements := strings.Split(string(content), ";")
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}

		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("failed to execute statement: %s, error: %w", stmt, err)
		}
	}

	return nil
}

// recordMigration records that a migration has been applied
func recordMigration(db *sql.DB, filename string) error {
	_, err := db.Exec("INSERT INTO migrations (filename) VALUES ($1)", filename)
	return err
}