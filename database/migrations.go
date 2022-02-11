package database

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

//go:embed migrations/*
var migrationsFS embed.FS

// MigrationsTempDir creates a temporary directory, populates it with the migration files,
// and returns the path to that directory.
// This is useful to run database migrations with only the dheart binary,
// without having to ship around the migration files separately.
//
// It is the caller's repsonsibility to remove the directory when it is no longer needed.
func MigrationsTempDir() (string, error) {
	tmpDir, err := os.MkdirTemp("", "dheart-migrations-*")
	if err != nil {
		return "", err
	}

	mFS, err := fs.Sub(migrationsFS, "migrations")
	if err != nil {
		return "", err
	}

	if err := fs.WalkDir(mFS, ".", func(path string, d fs.DirEntry, _ error) error {
		dst := filepath.Join(tmpDir, path)
		if dst == tmpDir {
			return nil
		}

		if d.IsDir() {
			if err := os.Mkdir(dst, 0700); err != nil {
				return fmt.Errorf("failed to mkdir %q: %w", dst, err)
			}
			return nil
		}

		content, err := migrationsFS.ReadFile(filepath.Join("migrations", path))
		if err != nil {
			return err
		}

		return os.WriteFile(dst, content, 0600)
	}); err != nil {
		return "", err
	}

	return tmpDir, nil
}
