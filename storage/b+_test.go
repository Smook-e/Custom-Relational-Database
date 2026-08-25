package storage_test

import (
	"path/filepath"
	"testing"
	"github.com/Smook-e/Custom-Relational-Database/storage"
)


func newEmptyStorageEngine(t *testing.T) *storage.StorageEngine {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "storage-test.db")

	engine, err := storage.InitializeStorageEngine(dbPath)
	if err != nil {
		t.Fatalf("failed to initialize storage engine: %v", err)
	}
	return engine
}