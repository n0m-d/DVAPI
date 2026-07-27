package database

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5"
)

var dbNameRE = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// EnsureDatabase connects to the server's default database and creates the
// target database from databaseURL when it does not already exist.
// Returns true when a new database was created.
func EnsureDatabase(ctx context.Context, databaseURL string) (created bool, err error) {
	adminURL, dbName, err := adminURLAndDBName(databaseURL)
	if err != nil {
		return false, err
	}
	if !dbNameRE.MatchString(dbName) {
		return false, fmt.Errorf("invalid database name %q", dbName)
	}

	conn, err := pgx.Connect(ctx, adminURL)
	if err != nil {
		return false, fmt.Errorf("connect to postgres: %w", err)
	}
	defer conn.Close(ctx)

	var exists bool
	if err := conn.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", dbName).Scan(&exists); err != nil {
		return false, fmt.Errorf("check database %s: %w", dbName, err)
	}
	if exists {
		return false, nil
	}

	if _, err := conn.Exec(ctx, "CREATE DATABASE "+dbName); err != nil {
		return false, fmt.Errorf("create database %s: %w", dbName, err)
	}
	return true, nil
}

func adminURLAndDBName(databaseURL string) (adminURL, dbName string, err error) {
	u, err := url.Parse(databaseURL)
	if err != nil {
		return "", "", fmt.Errorf("parse database url: %w", err)
	}

	dbName = strings.TrimPrefix(u.Path, "/")
	if dbName == "" {
		return "", "", fmt.Errorf("database name missing from url")
	}
	if i := strings.IndexByte(dbName, '/'); i >= 0 {
		dbName = dbName[:i]
	}

	admin := *u
	admin.Path = "/postgres"
	return admin.String(), dbName, nil
}
