package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// Migrate applies the initial schema (BUILD_SPEC §2) if it has not been applied
// yet. The migration SQL is read from disk so 0001_init.sql stays the single
// source of truth (also used by `ohctl` and humans). The directory is resolved
// from MIGRATIONS_DIR, defaulting to /app/migrations (where the Dockerfile
// copies it) and falling back to ./migrations for local runs.
//
// The frozen migration uses bare `create table` statements, so running it twice
// would error. We guard on the presence of the `users` table: if it already
// exists the schema is assumed in place and we skip. This keeps the frozen SQL
// untouched while making boot idempotent.
func (d *DB) Migrate(ctx context.Context) error {
	if d == nil || d.Pool == nil {
		return fmt.Errorf("db: migrate called with nil pool")
	}

	var exists bool
	err := d.Pool.QueryRow(ctx,
		`select exists (
			select 1 from information_schema.tables
			where table_schema = 'public' and table_name = 'users'
		)`).Scan(&exists)
	if err != nil {
		return fmt.Errorf("db: migrate check: %w", err)
	}
	if exists {
		log.Printf("db: schema already present, skipping migration")
		return nil
	}

	sqlBytes, path, err := readInitSQL()
	if err != nil {
		return err
	}
	if _, err := d.Pool.Exec(ctx, string(sqlBytes)); err != nil {
		return fmt.Errorf("db: apply migration %s: %w", path, err)
	}
	log.Printf("db: applied migration %s", path)
	return nil
}

// readInitSQL locates and reads 0001_init.sql, trying MIGRATIONS_DIR, the
// container path, then common local paths.
func readInitSQL() ([]byte, string, error) {
	candidates := []string{}
	if d := os.Getenv("MIGRATIONS_DIR"); d != "" {
		candidates = append(candidates, filepath.Join(d, "0001_init.sql"))
	}
	candidates = append(candidates,
		"/app/migrations/0001_init.sql",
		"migrations/0001_init.sql",
		"backend/migrations/0001_init.sql",
	)
	for _, p := range candidates {
		if b, err := os.ReadFile(p); err == nil {
			return b, p, nil
		}
	}
	return nil, "", fmt.Errorf("db: could not find 0001_init.sql (set MIGRATIONS_DIR); tried %v", candidates)
}
