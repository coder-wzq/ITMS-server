package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	dsn := flag.String("dsn", "postgres://itms:itms@2026@localhost:5432/itms?sslmode=disable", "PostgreSQL DSN")
	action := flag.String("action", "up", "action: up, down, version, force")
	version := flag.Int("version", 0, "target version for 'force' action")
	flag.Parse()

	m, err := migrate.New("file://migrations", *dsn)
	if err != nil {
		log.Fatalf("migrate init failed: %v", err)
	}

	switch *action {
	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("migrate up failed: %v", err)
		}
		ver, dirty, _ := m.Version()
		fmt.Printf("Migration up complete. Version=%d Dirty=%v\n", ver, dirty)

	case "down":
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("migrate down failed: %v", err)
		}
		ver, dirty, _ := m.Version()
		fmt.Printf("Migration down complete. Version=%d Dirty=%v\n", ver, dirty)

	case "version":
		ver, dirty, err := m.Version()
		if err != nil {
			log.Fatalf("version query failed: %v", err)
		}
		fmt.Printf("Version=%d Dirty=%v\n", ver, dirty)

	case "force":
		if err := m.Force(*version); err != nil {
			log.Fatalf("force version %d failed", *version)
		}
		fmt.Printf("Forced version to %d\n", *version)

	default:
		fmt.Fprintf(os.Stderr, "unknown action: %s\n", *action)
		os.Exit(1)
	}
}
