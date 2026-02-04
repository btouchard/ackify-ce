package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

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

	db, err := sql.Open("postgres", *dbDSN)
	if err != nil {
		log.Fatal("Cannot connect to database:", err)
	}
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatal("Cannot create database driver:", err)
	}

	m, err := migrate.NewWithDatabaseInstance(*migrationsPath, "postgres", driver)
	if err != nil {
		log.Fatal("Cannot create migrator:", err)
	}

	switch command {
	case "up":
		// Ensure ackify_app role exists before running migrations (for RLS support)
		if err := ensureAppRole(db); err != nil {
			log.Fatal("Failed to ensure ackify_app role:", err)
		}

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
	case "goto":
		if len(args) < 2 {
			log.Fatal("goto requires a version number")
		}
		var version uint
		_, err := fmt.Sscanf(args[1], "%d", &version)
		if err != nil {
			log.Fatal("Invalid version number:", err)
		}
		err = m.Migrate(version)
		if err != nil && !errors.Is(err, migrate.ErrNoChange) {
			log.Fatal("Migration goto failed:", err)
		}
		fmt.Printf("Migrated to version %d\n", version)
	case "force":
		if len(args) < 2 {
			log.Fatal("force requires a version number")
		}
		var version int
		_, err := fmt.Sscanf(args[1], "%d", &version)
		if err != nil {
			log.Fatal("Invalid version number:", err)
		}
		err = m.Force(version)
		if err != nil {
			log.Fatal("Force version failed:", err)
		}
		fmt.Printf("Forced version to %d (no migrations executed)\n", version)
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
	fmt.Println("  goto <v>     Migrate to specific version (up or down)")
	fmt.Println("  force <v>    Force version without running migrations (for existing DBs)")
	fmt.Println("  version      Show current migration version")
	fmt.Println("  drop         Drop all migrations (DANGER)")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -db-dsn string         Database DSN (or DB_DSN env var)")
	fmt.Println("  -migrations-path string Path to migrations (default: file://migrations)")
	fmt.Println()
	fmt.Println("Environment:")
	fmt.Println("  ACKIFY_APP_PASSWORD    Password for the ackify_app role (required for RLS)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  migrate up")
	fmt.Println("  migrate down 2")
	fmt.Println("  migrate goto 5")
	fmt.Println("  migrate force 1        # For existing DB with only signatures table")
	fmt.Println("  migrate version")
}

// ensureAppRole creates or updates the ackify_app role used for RLS.
// The password is read from ACKIFY_APP_PASSWORD environment variable.
// If not set, the function logs a warning and continues (for backward compatibility).
// If set, the role is created (or password updated) before migrations run.
func ensureAppRole(db *sql.DB) error {
	password := strings.TrimSpace(os.Getenv("ACKIFY_APP_PASSWORD"))
	if password == "" {
		log.Println("WARNING: ACKIFY_APP_PASSWORD not set. ackify_app role will not be created.")
		log.Println("         RLS migrations will fail if the role doesn't exist.")
		log.Println("         Set ACKIFY_APP_PASSWORD to enable RLS support.")
		return nil
	}

	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_roles WHERE rolname = 'ackify_app')").Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check if ackify_app role exists: %w", err)
	}

	if exists {
		_, err = db.Exec(fmt.Sprintf("ALTER ROLE ackify_app WITH PASSWORD '%s'", escapePassword(password)))
		if err != nil {
			return fmt.Errorf("failed to update ackify_app password: %w", err)
		}
		log.Println("ackify_app role exists, password updated")
	} else {
		createSQL := fmt.Sprintf(`
			CREATE ROLE ackify_app WITH
				LOGIN
				PASSWORD '%s'
				NOCREATEDB
				NOCREATEROLE
				NOINHERIT
				NOREPLICATION
				CONNECTION LIMIT -1
		`, escapePassword(password))

		_, err = db.Exec(createSQL)
		if err != nil {
			return fmt.Errorf("failed to create ackify_app role: %w", err)
		}
		log.Println("ackify_app role created successfully")
	}

	// Grant CONNECT on database (idempotent)
	var dbName string
	err = db.QueryRow("SELECT current_database()").Scan(&dbName)
	if err != nil {
		return fmt.Errorf("failed to get current database name: %w", err)
	}

	_, err = db.Exec(fmt.Sprintf("GRANT CONNECT ON DATABASE %s TO ackify_app", quoteIdentifier(dbName)))
	if err != nil {
		return fmt.Errorf("failed to grant CONNECT to ackify_app: %w", err)
	}

	// Grant USAGE on public schema (idempotent)
	_, err = db.Exec("GRANT USAGE ON SCHEMA public TO ackify_app")
	if err != nil {
		return fmt.Errorf("failed to grant USAGE on public schema: %w", err)
	}

	return nil
}

// escapePassword escapes single quotes in password for SQL
func escapePassword(password string) string {
	return strings.ReplaceAll(password, "'", "''")
}

// quoteIdentifier quotes a PostgreSQL identifier (table name, database name, etc.)
// to safely handle names containing special characters like hyphens.
func quoteIdentifier(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}
