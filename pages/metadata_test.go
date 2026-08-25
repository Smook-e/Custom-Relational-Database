package pages

import (
	"testing"
	"github.com/Smook-e/Custom-Relational-Database/entities"
	"path/filepath"
	// "fmt"
	"os"
)

func newEmptyDatabase(t *testing.T) *entities.Database {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "storage-test.db")

	filep, err := os.OpenFile(dbPath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		t.Fatalf("failed to open database file: %v", err)
	}
	db := &entities.Database{
		File:      filep,
		Tables:    make(map[string]*entities.Table),
		FreePages: make([]entities.FreePage, 0),
		TotalPages: uint32(0),
	}
	if err != nil {
		t.Fatalf("failed to initialize storage engine: %v", err)
	}
	return db
}