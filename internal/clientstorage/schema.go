package clientstorage

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

//go:embed migrations/remote/*.sql
var remoteMigrationFiles embed.FS

// InitRemoteSchema runs all embedded remote Turso/libSQL migrations.
func InitRemoteSchema(ctx context.Context, db *sql.DB) error {
	remoteFS, err := fs.Sub(remoteMigrationFiles, "migrations/remote")
	if err != nil {
		return err
	}
	return runRemoteMigrations(ctx, db, remoteFS)
}

type remoteMigration struct {
	version int64
	name    string
}

func runRemoteMigrations(ctx context.Context, db *sql.DB, migrations fs.FS) error {
	if err := initRemoteVersionTable(ctx, db); err != nil {
		return fmt.Errorf("init remote version table: %w", err)
	}

	files, err := collectRemoteMigrations(migrations)
	if err != nil {
		return err
	}

	applied, err := remoteAppliedVersions(ctx, db)
	if err != nil {
		return err
	}

	for _, migration := range files {
		if applied[migration.version] {
			continue
		}
		statements, err := readGooseUpStatements(migrations, migration.name)
		if err != nil {
			return err
		}
		for _, statement := range statements {
			if _, err := db.ExecContext(ctx, statement); err != nil {
				return fmt.Errorf("apply remote migration %s: %w", migration.name, err)
			}
		}
		if _, err := db.ExecContext(ctx, `INSERT INTO goose_db_version (version_id, is_applied) VALUES (?, ?)`, migration.version, 1); err != nil {
			return fmt.Errorf("record remote migration %s: %w", migration.name, err)
		}
	}

	return nil
}

func initRemoteVersionTable(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS goose_db_version (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			version_id INTEGER NOT NULL,
			is_applied INTEGER NOT NULL,
			tstamp TIMESTAMP DEFAULT (datetime('now'))
		)
	`); err != nil {
		return err
	}

	var count int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM goose_db_version WHERE version_id = 0`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	_, err := db.ExecContext(ctx, `INSERT INTO goose_db_version (version_id, is_applied) VALUES (?, ?)`, 0, 1)
	return err
}

func collectRemoteMigrations(migrations fs.FS) ([]remoteMigration, error) {
	entries, err := fs.ReadDir(migrations, ".")
	if err != nil {
		return nil, err
	}

	files := make([]remoteMigration, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".sql" {
			continue
		}
		prefix := strings.SplitN(entry.Name(), "_", 2)[0]
		version, err := strconv.ParseInt(prefix, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parse remote migration version %q: %w", entry.Name(), err)
		}
		files = append(files, remoteMigration{version: version, name: entry.Name()})
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].version < files[j].version
	})
	return files, nil
}

func remoteAppliedVersions(ctx context.Context, db *sql.DB) (map[int64]bool, error) {
	rows, err := db.QueryContext(ctx, `SELECT version_id, is_applied FROM goose_db_version`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[int64]bool)
	for rows.Next() {
		var version int64
		var isApplied int
		if err := rows.Scan(&version, &isApplied); err != nil {
			return nil, err
		}
		if isApplied != 0 {
			applied[version] = true
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return applied, nil
}

func readGooseUpStatements(migrations fs.FS, name string) ([]string, error) {
	content, err := fs.ReadFile(migrations, name)
	if err != nil {
		return nil, err
	}

	var up strings.Builder
	inUp := false
	for _, line := range strings.Split(string(content), "\n") {
		trimmed := strings.TrimSpace(line)
		switch trimmed {
		case "-- +goose Up":
			inUp = true
			continue
		case "-- +goose Down":
			inUp = false
			continue
		}
		if inUp {
			up.WriteString(line)
			up.WriteByte('\n')
		}
	}

	statements := splitSQLStatements(up.String())
	if len(statements) == 0 {
		return nil, fmt.Errorf("remote migration %s has no up statements", name)
	}
	return statements, nil
}

func splitSQLStatements(sql string) []string {
	parts := strings.Split(sql, ";")
	statements := make([]string, 0, len(parts))
	for _, part := range parts {
		statement := strings.TrimSpace(part)
		if statement != "" {
			statements = append(statements, statement)
		}
	}
	return statements
}
