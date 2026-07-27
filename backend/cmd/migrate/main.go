package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/n0m-d/DVAPI/internal/database"
	"github.com/pressly/goose/v3"
)

var identRE = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

func main() {
	_ = godotenv.Load(".env.local", ".env")

	flags := flag.NewFlagSet("migrate", flag.ExitOnError)
	dir := flags.String("dir", getEnv("GOOSE_MIGRATION_DIR", "./migrations"), "migrations directory")
	flags.Parse(os.Args[1:])

	args := flags.Args()
	if len(args) < 1 {
		printUsage()
		os.Exit(1)
	}

	command := args[0]

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = os.Getenv("GOOSE_DBSTRING")
	}
	if dbURL == "" {
		fmt.Fprintln(os.Stderr, "DATABASE_URL or GOOSE_DBSTRING is required")
		os.Exit(1)
	}

	ctx := context.Background()

	if command == "up" || command == "status" || command == "version" {
		created, err := database.EnsureDatabase(ctx, dbURL)
		if err != nil {
			fmt.Fprintln(os.Stderr, "ensure database:", err)
			os.Exit(1)
		}
		if created {
			fmt.Println("created database from DATABASE_URL")
		}
	}

	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		fmt.Fprintln(os.Stderr, "open database:", err)
		os.Exit(1)
	}
	defer db.Close()

	if table := os.Getenv("GOOSE_TABLE"); table != "" {
		goose.SetTableName(table)
		if schema := schemaFromTable(table); schema != "" {
			if !identRE.MatchString(schema) {
				fmt.Fprintln(os.Stderr, "invalid schema in GOOSE_TABLE:", schema)
				os.Exit(1)
			}
			if _, err := db.ExecContext(ctx, "CREATE SCHEMA IF NOT EXISTS "+schema); err != nil {
				fmt.Fprintln(os.Stderr, "create goose schema:", err)
				os.Exit(1)
			}
		}
	}

	if err := goose.SetDialect("postgres"); err != nil {
		fmt.Fprintln(os.Stderr, "set dialect:", err)
		os.Exit(1)
	}

	switch command {
	case "up":
		err = goose.UpContext(ctx, db, *dir)
	case "down":
		err = goose.DownContext(ctx, db, *dir)
	case "status":
		err = goose.StatusContext(ctx, db, *dir)
	case "version":
		err = goose.VersionContext(ctx, db, *dir)
	default:
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func schemaFromTable(table string) string {
	parts := strings.SplitN(table, ".", 2)
	if len(parts) == 2 {
		return parts[0]
	}
	return ""
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func printUsage() {
	fmt.Println(`Usage: migrate <command>

Commands:
  up       Apply all pending migrations
  down     Roll back the most recent migration
  status   Print migration status
  version  Print current migration version`)
}
