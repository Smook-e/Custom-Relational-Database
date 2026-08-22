package storage

import (
	"fmt"
	"os"

	bufferpool "github.com/Smook-e/Custom-Relational-Database/bufferPool"
	"github.com/Smook-e/Custom-Relational-Database/entities"
	"github.com/Smook-e/Custom-Relational-Database/filehandler"
	"github.com/Smook-e/Custom-Relational-Database/pages"
)
/*
This file contains the implementation of the StorageEngine struct, which is responsible for managing the database file, buffer pool, and metadata. 
It provides methods for initializing the storage engine, committing changes to disk, printing metadata, creating new pages, and managing the B+Tree index.
*/

const bufferSize = 4096

type StorageEngine struct {
	db *entities.Database
	Bp *bufferpool.BufferPool
	metaWrite bool // Flag to indicate if the meta page needs to be written to disk
}

// Commit flushes the buffer pool to disk and writes the meta page if necessary.
func (engine *StorageEngine) Commit() error {
	err := engine.Bp.Flush()
	if err != nil {
		return fmt.Errorf("An error occured while commiting to disk: %w", err)
	}
	if engine.metaWrite {
		if err := pages.WriteMetaPage(engine.db); err != nil {
			return fmt.Errorf("An error occured while writing meta page to disk: %w", err)
		}
		engine.metaWrite = false
	}
	fmt.Println("Committed Changes to disk successfully")
	return nil
}


func (engine *StorageEngine) PrintMetaData() error {
	if engine == nil || engine.db == nil {
		return fmt.Errorf("storage engine or database not initialized")
	}
	engine.db.PrintDatabase()
	return nil
}

// Initializes the storage engine based on the given filename. 
// If the file does not exist, it creates a new database with example tables and sample rows. 
// If the file exists, it reads the metadata from the file to populate the database structure.
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

	// If the file is empty, create a Database with a meta page, free-space page, and
	// example tables and a few rows.
	if engine.db.TotalPages == 0 {
		// update total pages to account for meta + free-space pages
		engine.db.TotalPages = 2
		// create example tables
		if err := engine.CreateTable("products", []entities.ColumnDefinition{
			{Name: "id", DataType: "serial", Constraints: []string{"primarykey"}},
			{Name: "name", DataType: "varchar(50)", Constraints: []string{"notnull"}},
			{Name: "price", DataType: "bigint", Constraints: []string{"notnull"}},
			{Name: "quantity", DataType: "smallint", Constraints: []string{"notnull", "default"}, Default: "1"},
			{Name: "seller", DataType: "varchar", Constraints: []string{}},
		}, []entities.ForeignKeyDefinition{}); err != nil {
			return nil, fmt.Errorf("failed to create products table: %w", err)
		}
		
		
		if err := engine.CreateTable("users", []entities.ColumnDefinition{
			{Name: "id", DataType: "serial", Constraints: []string{"primarykey"}},
			{Name: "name", DataType: "varchar(50)", Constraints: []string{"notnull", "default"}, Default: "anonymous"},
			{Name: "email", DataType: "varchar(30)", Constraints: []string{"notnull","unique", "default"}, Default: "unknown"},
			{Name: "phone_number", DataType: "varchar(15)", Constraints: []string{"notnull","unique", "default"}, Default: "unknown"},
			{Name: "age", DataType: "int", Constraints: []string{}},
		}, []entities.ForeignKeyDefinition{}); err != nil {
			return nil, fmt.Errorf("failed to create users table: %w", err)
		}
		
		if err := engine.CreateTable("orders", []entities.ColumnDefinition{
			{Name: "user_id", DataType: "int", Constraints: []string{"primarykey", "notnull"}},
			{Name: "product_id", DataType: "int", Constraints: []string{"primarykey", "notnull"}},
			{Name: "quantity", DataType: "int", Constraints: []string{"notnull"}},
		}, []entities.ForeignKeyDefinition{
			{ColumnName: "user_id", ReferencedTableName: "users", ReferencedColumnName: "id"},
			{ColumnName: "product_id", ReferencedTableName: "products", ReferencedColumnName: "id"},
		}); err != nil {
			return nil, fmt.Errorf("failed to create orders table: %w", err)
		}
		
		// ensure FreePages is empty for meta write
		engine.db.FreePages = []entities.FreePage{}

		

		

		// insert a few sample rows to make the DB testable (FindFreePage will create data pages)
		if _, _, err := engine.InsertRow([]string{"", "","email@example.com","123-456-7890", "20"}, "users"); err != nil {
			fmt.Printf("failed to insert sample user row: %v", err)
		}
		if _, _, err := engine.InsertRow([]string{"", "emily", "emily@example.com", "098-765-4321", "25"}, "users"); err != nil {
			return engine, fmt.Errorf("failed to insert sample user row: %w", err)
		}
		if _, _, err := engine.InsertRow([]string{"", "alice", "alice@example.com", "098-765-4320", "30"}, "users"); err != nil {
			return engine, fmt.Errorf("failed to insert sample user row: %w", err)
		}
		if _, _, err := engine.InsertRow([]string{"", "bob", "bob@example.com", "098-761-4321", "35"}, "users"); err != nil {
			return engine, fmt.Errorf("failed to insert sample user row: %w", err)
		}


		if _, _, err := engine.InsertRow([]string{"", "IPhone", "1000", "2", "apple"}, "products"); err != nil {
			fmt.Printf("failed to insert sample product row: %v", err)
		}
		if _, _, err := engine.InsertRow([]string{"", "Macbook", "1200", "", "apple"}, "products"); err != nil {
			fmt.Printf("failed to insert sample product row: %v", err)
		}
		if _, _, err := engine.InsertRow([]string{"", "Samsung Galaxy", "800", "1", "samsung"}, "products"); err != nil {
			return engine, fmt.Errorf("failed to insert sample product row: %w", err)
		}
		if _, _, err := engine.InsertRow([]string{"", "Google Pixel", "600", "2", "google"}, "products"); err != nil {
			return engine, fmt.Errorf("failed to insert sample product row: %w", err)
		}
		if _, _, err := engine.InsertRow([]string{"", "OnePlus", "400", "", "oneplus"}, "products"); err != nil {
			return engine, fmt.Errorf("failed to insert sample product row: %w", err)
		}
		if _, _, err := engine.InsertRow([]string{"", "Xiaomi", "600", "3", ""}, "products"); err != nil {
			return engine, fmt.Errorf("failed to insert sample product row: %w", err)
		}


		if _,_, err := engine.InsertRow([]string{"1", "1", "2"}, "orders"); err != nil {
			fmt.Printf("failed to insert sample order row: %v", err)
		}
		if _,_, err := engine.InsertRow([]string{"2", "2", "2"}, "orders"); err != nil {
			fmt.Printf("failed to insert sample order row: %v", err)
		}
		if _,_, err := engine.InsertRow([]string{"3", "3", "2"}, "orders"); err != nil {
			fmt.Printf("failed to insert sample order row: %v", err)
		}
		if _,_, err := engine.InsertRow([]string{"4", "4", "2"}, "orders"); err != nil {
			fmt.Printf("failed to insert sample order row: %v", err)
		}
		// for _, table := range engine.db.Tables {
		// 	var buf []byte
		// 	for colName, root := range table.Indexes {
		// 		col, err := table.GetColumnByName(colName)
		// 		buf, err = engine.Bp.Get(root)
		// 		if err != nil {
		// 			return engine, fmt.Errorf("failed to get buffer for index root: %w", err)
		// 		}
		// 		PrintLeafPageEntries(buf, col)
		// 	}
		// }
		engine.metaWrite = true
		if err := engine.Commit(); err != nil {
			return engine, fmt.Errorf("failed to commit initial database state: %w", err)
		}

	} else {
		// existing database, attempt to read meta pages to populate tables/free list
		if err := pages.ReadMetaPage(engine.db); err != nil {
			return engine, fmt.Errorf("failed to read meta pages: %w", err)
		}
		// Test Linear Search
		targetColumn := &engine.db.Tables["products"].Columns[4]
		
		val, err := targetColumn.GetDefaultValue("apple")
		if err != nil {
			return engine, fmt.Errorf("failed to get default value for column %s: %w", targetColumn.Name, err)
		}
		// Serialize the value to match the format stored in the database
		serializedValue, err := entities.Serialize(val, targetColumn)
		if err != nil {
			return engine, fmt.Errorf("failed to serialize value for column %s: %w", targetColumn.Name, err)
		}
		searchCondition := &SearchCondition{
			ColumnName: targetColumn.Name,
			Operator: "=",
			Value: serializedValue,
		}
		result, err := engine.LinearSearch("products", searchCondition)
		if err != nil {
			return engine, fmt.Errorf("failed to perform linear search: %w", err)
		}
		fmt.Println("Linear Search results:")
		for _, row := range result {
			fmt.Println(row)
		}
	}

	return engine, nil
}
// NewPage creates a new page in the database file and returns its page ID.
func (engine *StorageEngine) NewPage() (uint32, error) {
	newPageID := engine.db.TotalPages
	engine.db.TotalPages++
	buffer := make([]byte, bufferSize)
	filehandler.WriteToFile(engine.db.File, newPageID, buffer)
	engine.Bp.Get(newPageID)
	
	return newPageID, nil
}

