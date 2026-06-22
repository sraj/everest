package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	// Load .env file if exists
	_ = godotenv.Load()

	// Parse flags
	var (
		migrationsPath string
		seedsPath      string
		databaseURL    string
		steps          int
	)

	flag.StringVar(&migrationsPath, "migrations", "db/migrations", "Path to migrations directory")
	flag.StringVar(&seedsPath, "seeds", "db/seeds", "Path to seeds directory")
	flag.StringVar(&databaseURL, "database", os.Getenv("DATABASE_URL"), "Database URL")
	flag.IntVar(&steps, "steps", 0, "Number of steps for up/down (0 = all)")
	flag.Parse()

	command := flag.Arg(0)
	if command == "" {
		printUsage()
		os.Exit(1)
	}

	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required. Set via -database flag or DATABASE_URL env var")
	}

	switch command {
	// Migration commands
	case "up":
		runMigrate(databaseURL, migrationsPath, "up", steps)
	case "down":
		runMigrate(databaseURL, migrationsPath, "down", steps)
	case "version":
		runMigrate(databaseURL, migrationsPath, "version", 0)
	case "force":
		runMigrate(databaseURL, migrationsPath, "force", steps)
	case "drop":
		runMigrate(databaseURL, migrationsPath, "drop", 0)

	// Seed commands
	case "seed":
		runSeeds(databaseURL, seedsPath)
	case "seed:status":
		showSeedStatus(databaseURL, seedsPath)
	case "seed:reset":
		resetSeeds(databaseURL)

	// Combined commands
	case "setup":
		// Run migrations then seeds
		runMigrate(databaseURL, migrationsPath, "up", 0)
		runSeeds(databaseURL, seedsPath)

	case "fresh":
		// Drop, migrate, and seed
		runMigrate(databaseURL, migrationsPath, "drop", 0)
		runMigrate(databaseURL, migrationsPath, "up", 0)
		runSeeds(databaseURL, seedsPath)

	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Database migration and seeding tool")
	fmt.Println("")
	fmt.Println("Usage: migrate [flags] <command>")
	fmt.Println("")
	fmt.Println("Migration Commands:")
	fmt.Println("  up           Apply all or N up migrations")
	fmt.Println("  down         Roll back all or N down migrations")
	fmt.Println("  version      Print current migration version")
	fmt.Println("  force        Set version V without running migrations (use -steps)")
	fmt.Println("  drop         Drop everything in the database")
	fmt.Println("")
	fmt.Println("Seed Commands:")
	fmt.Println("  seed         Run all pending seeds")
	fmt.Println("  seed:status  Show seed status")
	fmt.Println("  seed:reset   Reset seeds tracking")
	fmt.Println("")
	fmt.Println("Combined Commands:")
	fmt.Println("  setup        Run migrations then seeds")
	fmt.Println("  fresh        Drop, migrate, and seed (fresh start)")
	fmt.Println("")
	fmt.Println("Flags:")
	flag.PrintDefaults()
}

func runMigrate(databaseURL, migrationsPath, command string, steps int) {
	m, err := migrate.New(
		fmt.Sprintf("file://%s", migrationsPath),
		databaseURL,
	)
	if err != nil {
		log.Fatalf("Failed to create migrate instance: %v", err)
	}
	defer m.Close()

	switch command {
	case "up":
		if steps > 0 {
			err = m.Steps(steps)
		} else {
			err = m.Up()
		}
		if err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Migration up failed: %v", err)
		}
		fmt.Println("Migrations applied successfully")

	case "down":
		if steps > 0 {
			err = m.Steps(-steps)
		} else {
			err = m.Down()
		}
		if err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Migration down failed: %v", err)
		}
		fmt.Println("Migrations rolled back successfully")

	case "version":
		version, dirty, err := m.Version()
		if errors.Is(err, migrate.ErrNilVersion) {
			fmt.Println("Version: 0, Dirty: false (no migrations applied)")
		} else if err != nil {
			log.Fatalf("Failed to get version: %v", err)
		} else {
			fmt.Printf("Version: %d, Dirty: %v\n", version, dirty)
		}

	case "force":
		if steps == 0 {
			log.Fatal("Version required for force command. Use -steps flag")
		}
		err = m.Force(steps)
		if err != nil {
			log.Fatalf("Force failed: %v", err)
		}
		fmt.Printf("Forced version to %d\n", steps)

	case "drop":
		err = m.Drop()
		if err != nil {
			log.Fatalf("Drop failed: %v", err)
		}
		fmt.Println("Database dropped successfully")
	}
}

func runSeeds(databaseURL, seedsPath string) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Create seeds tracking table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS seeds (
			name VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)
	`)
	if err != nil {
		log.Fatalf("Failed to create seeds table: %v", err)
	}

	// Get list of seed files
	files, err := filepath.Glob(filepath.Join(seedsPath, "*.sql"))
	if err != nil {
		log.Fatalf("Failed to list seed files: %v", err)
	}
	sort.Strings(files)

	if len(files) == 0 {
		fmt.Println("No seed files found")
		return
	}

	// Get applied seeds
	applied := make(map[string]bool)
	rows, err := db.Query("SELECT name FROM seeds")
	if err != nil {
		log.Fatalf("Failed to query applied seeds: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			log.Fatal(err)
		}
		applied[name] = true
	}

	// Run pending seeds
	for _, file := range files {
		name := filepath.Base(file)
		if applied[name] {
			fmt.Printf("Skipping %s (already applied)\n", name)
			continue
		}

		fmt.Printf("Applying %s...\n", name)

		content, err := os.ReadFile(file)
		if err != nil {
			log.Fatalf("Failed to read %s: %v", name, err)
		}

		if _, err := db.Exec(string(content)); err != nil {
			log.Fatalf("Failed to execute %s: %v", name, err)
		}

		if _, err := db.Exec("INSERT INTO seeds (name) VALUES ($1)", name); err != nil {
			log.Fatalf("Failed to mark %s as applied: %v", name, err)
		}
	}

	fmt.Println("Seeds applied successfully")
}

func showSeedStatus(databaseURL, seedsPath string) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Get list of seed files
	files, err := filepath.Glob(filepath.Join(seedsPath, "*.sql"))
	if err != nil {
		log.Fatalf("Failed to list seed files: %v", err)
	}
	sort.Strings(files)

	// Get applied seeds
	applied := make(map[string]string)
	rows, err := db.Query("SELECT name, applied_at FROM seeds")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var name, appliedAt string
			if err := rows.Scan(&name, &appliedAt); err == nil {
				applied[name] = appliedAt
			}
		}
	}

	fmt.Println("Seed Status:")
	fmt.Println(strings.Repeat("-", 60))

	for _, file := range files {
		name := filepath.Base(file)
		if appliedAt, ok := applied[name]; ok {
			fmt.Printf("✓ %s (applied: %s)\n", name, appliedAt[:19])
		} else {
			fmt.Printf("○ %s (pending)\n", name)
		}
	}
}

func resetSeeds(databaseURL string) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	_, err = db.Exec("DELETE FROM seeds")
	if err != nil {
		log.Fatalf("Failed to reset seeds: %v", err)
	}
	fmt.Println("Seeds reset successfully")
}
