package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"database/sql"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func main() {
	var dbDSN = flag.String("db-dsn", os.Getenv("ACKIFY_DB_DSN"), "Database DSN")
	var migrationsPath = flag.String("migrations-path", "file://migrations", "Path to migrations directory")
	flag.Parse()

	if *dbDSN == "" {
		log.Fatal("DB_DSN environment variable or -db-dsn flag is required")
	}

	args := flag.Args()
	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	command := args[0]

	// Open database connection
	db, err := sql.Open("postgres", *dbDSN)
	if err != nil {
		log.Fatal("Cannot connect to database:", err)
	}
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)

	// Create postgres driver
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatal("Cannot create database driver:", err)
	}

	// Create migrator
	m, err := migrate.NewWithDatabaseInstance(*migrationsPath, "postgres", driver)
	if err != nil {
		log.Fatal("Cannot create migrator:", err)
	}

	switch command {
	case "up":
		err = m.Up()
		if err != nil && !errors.Is(err, migrate.ErrNoChange) {
			log.Fatal("Migration up failed:", err)
		}
		fmt.Println("CE migrations applied successfully")
	case "down":
		steps := 1
		if len(args) > 1 {
			_, _ = fmt.Sscanf(args[1], "%d", &steps)
		}
		err = m.Steps(-steps)
		if err != nil && !errors.Is(err, migrate.ErrNoChange) {
			log.Fatal("Migration down failed:", err)
		}
		fmt.Printf("CE migrations rolled back %d steps\n", steps)
	case "version":
		version, dirty, err := m.Version()
		if err != nil {
			log.Fatal("Cannot get version:", err)
		}
		fmt.Printf("Version: %d, Dirty: %t\n", version, dirty)
	case "drop":
		err = m.Drop()
		if err != nil {
			log.Fatal("Drop failed:", err)
		}
		fmt.Println("All CE migrations dropped")
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: migrate [options] <command>")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  up           Apply all CE migrations")
	fmt.Println("  down [n]     Rollback n CE migrations (default: 1)")
	fmt.Println("  version      Show current migration version")
	fmt.Println("  drop         Drop all migrations (DANGER)")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -db-dsn string         Database DSN (or DB_DSN env var)")
	fmt.Println("  -migrations-path string Path to migrations (default: file://migrations)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  migrate up")
	fmt.Println("  migrate down 2")
	fmt.Println("  migrate version")
}
