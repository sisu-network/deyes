package database_test

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/sisu-network/deyes/database"
)

func TestMigrationsTempDir(t *testing.T) {
	t.Parallel()

	tmpDir, err := database.MigrationsTempDir()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Running with go test sets the working directory to the parent of the test source file,
	// so we can just directly read from disk at this path.
	onDiskFS := os.DirFS("./migrations")

	// Walk the on-disk FS and ensure the FS returned by db.MigrationsTempDir
	// matches exactly.
	if err := fs.WalkDir(onDiskFS, ".", func(path string, d fs.DirEntry, _ error) error {
		if d.IsDir() {
			// Nothing to validate.
			return nil
		}

		// Read the walked path from disk.
		onDiskContent, err := os.ReadFile(filepath.Join(".", "migrations", path))
		if err != nil {
			t.Fatalf("failed to source migration file at path %q: %v", path, err)
		}

		// Compare with the file inside the temporary directory.
		tmpContent, err := os.ReadFile(filepath.Join(tmpDir, path))
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(onDiskContent, tmpContent) {
			t.Fatalf("contents differed for path %q", path)
		}

		return nil
	}); err != nil {
		t.Fatal(err)
	}
}
