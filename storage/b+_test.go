package storage

import (
	"path/filepath"
	"testing"
	"fmt"
	// "github.com/Smook-e/Custom-Relational-Database/storage"
	"github.com/Smook-e/Custom-Relational-Database/entities"
)


func newEmptyStorageEngine(t *testing.T) *StorageEngine {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "storage-test.db")

	engine, err := InitializeStorageEngine(dbPath)
	if err != nil {
		t.Fatalf("failed to initialize storage engine: %v", err)
	}
	return engine
}

func TestIndexInsert(t *testing.T) {
	engine := newEmptyStorageEngine(t)

	// Create a table
	if err := engine.CreateTable("users", []entities.ColumnDefinition{
			{Name: "id", DataType: "serial", Constraints: []string{"primarykey"}},
		}, []entities.ForeignKeyDefinition{}); err != nil {
			t.Fatalf("failed to create users table: %v", err)
	}
	rootID := engine.db.Tables["users"].Indexes["id"]
	col := engine.db.Tables["users"].Columns[0]
	// Insert 1,000,000 numbers into the index
	for i := 1; i <= 1000000; i++ {
		val , err := entities.Serialize(int32(i), &col)
		if err != nil {
			t.Fatalf("failed to serialize value: %v", err)
		}
		rootID,  err = engine.InsertIntoIndex(rootID, val, uint32(i), uint16(i), &col); 
		if err != nil {
			t.Fatalf("failed to insert into users table: %v", err)
		}
	}
	// Search for all numbers and verify they exist
	var pageID uint32
	for i := 1; i <= 1000000; i++ {
		val , err := entities.Serialize(int32(i), &col)
		if err != nil {
			t.Fatalf("failed to serialize value: %v", err)
		}
		pageID, _, err = engine.IndexSearch(rootID, val, &col)
		if err != nil {
			t.Fatalf("failed to search in users table: %v", err)
		}
		if pageID == 0 {
			t.Fatalf("value %d not found in index", i)
		}
	}
	// Search for a non-existent number and verify it does not exist
	val , err := entities.Serialize(int32(1000001), &col)
	if err != nil {
		t.Fatalf("failed to serialize value: %v", err)
	}
	pageID, _, err = engine.IndexSearch(rootID, val, &col)
	if err != nil {
		t.Fatalf("failed to search in users table: %v", err)
	}
	if pageID != 0 {
		t.Fatalf("value 1000001 should not be found in index")
	}
}
func TestIndexInsert_ForStrings(t *testing.T) {
	engine := newEmptyStorageEngine(t)

	// Create a table
	if err := engine.CreateTable("users", []entities.ColumnDefinition{
			{Name: "id", DataType: "VARCHAR(50)", Constraints: []string{"primarykey"}},
		}, []entities.ForeignKeyDefinition{}); err != nil {
			t.Fatalf("failed to create users table: %v", err)
	}
	rootID := engine.db.Tables["users"].Indexes["id"]
	col := engine.db.Tables["users"].Columns[0]
	// Insert 100,000 strings into the index
	for i := 1; i <= 100000; i++ {
		val , err := entities.Serialize(fmt.Sprintf("value_%d", i), &col)
		if err != nil {
			t.Fatalf("failed to serialize value: %v", err)
		}
		rootID,  err = engine.InsertIntoIndex(rootID, val, uint32(i), uint16(i), &col); 
		if err != nil {
			t.Fatalf("failed to insert into users table: %v", err)
		}
	}
	// Search for all strings and verify they exist
	var pageID uint32
	for i := 1; i <= 100000; i++ {
		val , err := entities.Serialize(fmt.Sprintf("value_%d", i), &col)
		if err != nil {
			t.Fatalf("failed to serialize value: %v", err)
		}
		pageID, _, err = engine.IndexSearch(rootID, val, &col)
		if err != nil {
			t.Fatalf("failed to search in users table: %v", err)
		}
		if pageID == 0 {
			t.Fatalf("value %d not found in index", i)
		}
	}
	// Search for a non-existent string and verify it does not exist
	val , err := entities.Serialize(fmt.Sprintf("value_%d", 1000002), &col)
	if err != nil {
		t.Fatalf("failed to serialize value: %v", err)
	}
	pageID, _, err = engine.IndexSearch(rootID, val, &col)
	if err != nil {
		t.Fatalf("failed to search in users table: %v", err)
	}
	if pageID != 0 {
		t.Fatalf("value 1000001 should not be found in index")
	}
}