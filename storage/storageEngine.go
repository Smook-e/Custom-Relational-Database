package storage

import (
	"fmt"
	"os"

	bufferpool "github.com/Smook-e/Custom-Relational-Database/bufferPool"
	"github.com/Smook-e/Custom-Relational-Database/entities"
	"github.com/Smook-e/Custom-Relational-Database/filehandler"
	"github.com/Smook-e/Custom-Relational-Database/pages"
)

const bufferSize = 4096

type StorageEngine struct {
	db *entities.Database
	Bp *bufferpool.BufferPool
}

func (engine *StorageEngine) Commit() error {
	err := engine.Bp.Flush()
	if err != nil {
		return fmt.Errorf("An error occured while commiting to disk: %w", err)
	}
	fmt.Println("Committed Changes to disk successfully")
	return nil
}

func InitializeStorageEngine(filename string) (*StorageEngine, error) {
	engine := &StorageEngine{}

	// Open or create the file and initialize the Database struct
	filep, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open database file: %w", err)
	}
	fi, err := filep.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat database file: %w", err)
	}

	engine.db = &entities.Database{
		File:      filep,
		Tables:    make(map[string]*entities.Table),
		FreePages: make([]entities.FreePage, 0),
		TotalPages: uint32(fi.Size() / bufferSize),
	}

	// Initialize buffer pool
	engine.Bp = bufferpool.InitializeBufferPool(engine.db.File)

	// If the file is empty, create a minimal, testable database: meta + free-space pages,
	// two example tables and a few rows.
	if engine.db.TotalPages == 0 {
		// create example tables
		if err := engine.db.CreateTable("products", []entities.ColumnDefinition{
			{Name: "id", DataType: "int", Constraints: []string{"primarykey", "notnull"}},
			{Name: "name", DataType: "varchar", Constraints: []string{"notnull"}},
			{Name: "price", DataType: "int", Constraints: []string{"notnull"}},
			{Name: "quantity", DataType: "int", Constraints: []string{"notnull"}},
			{Name: "seller", DataType: "varchar", Constraints: []string{"notnull"}},
		}); err != nil {
			return nil, fmt.Errorf("failed to create products table: %w", err)
		}

		if err := engine.db.CreateTable("users", []entities.ColumnDefinition{
			{Name: "id", DataType: "int", Constraints: []string{"primarykey"}},
			{Name: "name", DataType: "varchar", Constraints: []string{"notnull"}},
			{Name: "age", DataType: "int", Constraints: []string{}},
		}); err != nil {
			return nil, fmt.Errorf("failed to create users table: %w", err)
		}

		// ensure FreePages is empty for meta write
		engine.db.FreePages = []entities.FreePage{}

		// write meta + free-space pages (pages.WriteMetaPage writes page 0 and page 1)
		if err := pages.WriteMetaPage(engine.db); err != nil {
			return nil, fmt.Errorf("failed to write meta pages: %w", err)
		}

		// update total pages to account for meta + free-space pages
		engine.db.TotalPages = 2

		// insert a few sample rows to make the DB testable (FindFreePage will create data pages)
		if _, _, err := engine.InsertRow([]string{"1", "joe", "20"}, "users"); err != nil {
			return nil, fmt.Errorf("failed to insert sample user row: %w", err)
		}
		if _, _, err := engine.InsertRow([]string{"2", "emily", "25"}, "users"); err != nil {
			return nil, fmt.Errorf("failed to insert sample user row: %w", err)
		}
		if _, _, err := engine.InsertRow([]string{"1", "IPhone", "1000", "2", "apple"}, "products"); err != nil {
			return nil, fmt.Errorf("failed to insert sample product row: %w", err)
		}
		if _, _, err := engine.InsertRow([]string{"2", "Macbook", "1200", "3", "apple"}, "products"); err != nil {
			return nil, fmt.Errorf("failed to insert sample product row: %w", err)
		}

		// persist updated meta + free-space pages after inserts
		if err := pages.WriteMetaPage(engine.db); err != nil {
			return nil, fmt.Errorf("failed to write meta pages after inserts: %w", err)
		}
	} else {
		// existing database, attempt to read meta pages to populate tables/free list
		if err := pages.ReadMetaPage(engine.db); err != nil {
			return nil, fmt.Errorf("failed to read meta pages: %w", err)
		}
	}

	return engine, nil
}
func (engine *StorageEngine) NewPage() (uint32, error) {
	newPageID := engine.db.TotalPages
	engine.db.TotalPages++
	buffer := make([]byte, bufferSize)
	filehandler.WriteToFile(engine.db.File, newPageID, buffer)
	engine.Bp.Get(newPageID)
	
	return newPageID, nil
}

