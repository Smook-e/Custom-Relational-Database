package storage

import (
	"fmt"
	"os"
	"errors"
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

// Closes the database file and flushes the buffer pool to disk.
func (engine *StorageEngine) Close() error {
	if engine.Bp != nil {
		if err := engine.Bp.Flush(); err != nil {
			return fmt.Errorf("failed to flush buffer pool: %w", err)
		}
	}
	if engine.db != nil && engine.db.File != nil {
		if err := engine.db.File.Close(); err != nil {
			return fmt.Errorf("failed to close database file: %w", err)
		}
	}
	return nil
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
		// if err := engine.CreateTable("products", []entities.ColumnDefinition{
		// 	{Name: "id", DataType: "serial", Constraints: []string{"primarykey"}},
		// 	{Name: "name", DataType: "varchar(50)", Constraints: []string{"notnull"}},
		// 	{Name: "price", DataType: "bigint", Constraints: []string{"notnull"}},
		// 	{Name: "quantity", DataType: "smallint", Constraints: []string{"notnull", "default"}, Default: "1"},
		// 	{Name: "seller", DataType: "varchar", Constraints: []string{}},
		// }, []entities.ForeignKeyDefinition{}); err != nil {
		// 	return nil, fmt.Errorf("failed to create products table: %w", err)
		// }
		
		
		if err := engine.CreateTable("users", []entities.ColumnDefinition{
			{Name: "id", DataType: "serial", Constraints: []string{"primarykey"}},
			{Name: "name", DataType: "varchar(50)", Constraints: []string{"notnull", "default"}, Default: "anonymous"},
			{Name: "email", DataType: "varchar(30)", Constraints: []string{"notnull","unique", "default"}, Default: "unknown"},
			{Name: "phone_number", DataType: "varchar(15)", Constraints: []string{"notnull","unique", "default"}, Default: "unknown"},
			{Name: "age", DataType: "int", Constraints: []string{}},
		}, []entities.ForeignKeyDefinition{}); err != nil {
			return nil, fmt.Errorf("failed to create users table: %w", err)
		}
		
		// if err := engine.CreateTable("orders", []entities.ColumnDefinition{
		// 	{Name: "user_id", DataType: "int", Constraints: []string{"primarykey", "notnull"}},
		// 	{Name: "product_id", DataType: "int", Constraints: []string{"primarykey", "notnull"}},
		// 	{Name: "quantity", DataType: "int", Constraints: []string{"notnull"}},
		// }, []entities.ForeignKeyDefinition{
		// 	{ColumnName: "user_id", ReferencedTableName: "users", ReferencedColumnName: "id"},
		// 	{ColumnName: "product_id", ReferencedTableName: "products", ReferencedColumnName: "id"},
		// }); err != nil {
		// 	return nil, fmt.Errorf("failed to create orders table: %w", err)
		// }
		
		// ensure FreePages is empty for meta write
		// engine.db.FreePages = []entities.FreePage{}

		

		Rows  := []entities.RowID{}

		
		// insert a few sample rows to make the DB testable (FindFreePage will create data pages)
		pageID, slot, err := engine.InsertRow([]string{"", "anon","email@example.com","123-456-7890", "20"}, "users");if err != nil {
			fmt.Printf("failed to insert sample user row: %v", err)
		}
		Rows = append(Rows, entities.RowID{PageID: pageID, Slot: slot})
		// engine.DeleteRow(Rows[0].PageID, Rows[0].Slot)
		
		pageID, slot, err = engine.InsertRow([]string{"", "emily", "emily@example.com", "098-765-4321", "25"}, "users"); if err != nil {
			return engine, fmt.Errorf("failed to insert sample user row: %w", err)
		}
		Rows = append(Rows, entities.RowID{PageID: pageID, Slot: slot})
		// engine.DeleteRow(Rows[0].PageID, Rows[0].Slot)
		pageID, slot, err = engine.InsertRow([]string{"", "alice", "alice@example.com", "098-765-4320", "30"}, "users"); if err != nil {
			return engine, fmt.Errorf("failed to insert sample user row: %w", err)
		}
		Rows = append(Rows, entities.RowID{PageID: pageID, Slot: slot})
		// engine.DeleteRow(Rows[0].PageID, Rows[0].Slot)
		pageID, slot, err = engine.InsertRow([]string{"", "bob", "bob@example.com", "098-761-4321", "35"}, "users"); if err != nil {
			return engine, fmt.Errorf("failed to insert sample user row: %w", err)
		}
		Rows = append(Rows, entities.RowID{PageID: pageID, Slot: slot})
		result, err := engine.Search("users", []string{"*"}, nil)
		if err != nil {
			return engine, fmt.Errorf("failed to search users: %w", err)
		}
		for _, row := range result {
			fmt.Println(row)
		}
		// table := engine.db.Tables["users"]
		// colOffsets := make(map[string]int)
		
		// engine.DeleteRow(table, colOffsets, Rows[0].PageID, Rows[0].Slot)
		// engine.DeleteRow(table, colOffsets, Rows[1].PageID, Rows[1].Slot)
		// clear(colOffsets)
		// engine.DeleteRow(table,colOffsets,Rows[2].PageID, Rows[2].Slot)
		// engine.DeleteRow(table, colOffsets,Rows[3].PageID, Rows[3].Slot)

		result, err = engine.Search("users", []string{"*"}, nil)
		if err != nil {
			return engine, fmt.Errorf("failed to search users: %w", err)
		}
		for _, row := range result {
			fmt.Println(row)
		}

		// if _, _, err := engine.InsertRow([]string{"", "IPhone", "1000", "2", "apple"}, "products"); err != nil {
		// 	fmt.Printf("failed to insert sample product row: %v", err)
		// }
		// if _, _, err := engine.InsertRow([]string{"", "Macbook", "1200", "", "apple"}, "products"); err != nil {
		// 	fmt.Printf("failed to insert sample product row: %v", err)
		// }
		// if _, _, err := engine.InsertRow([]string{"", "Samsung Galaxy", "800", "1", "samsung"}, "products"); err != nil {
		// 	return engine, fmt.Errorf("failed to insert sample product row: %w", err)
		// }
		// if _, _, err := engine.InsertRow([]string{"", "Google Pixel", "600", "2", "google"}, "products"); err != nil {
		// 	return engine, fmt.Errorf("failed to insert sample product row: %w", err)
		// }
		// if _, _, err := engine.InsertRow([]string{"", "OnePlus", "400", "", "oneplus"}, "products"); err != nil {
		// 	return engine, fmt.Errorf("failed to insert sample product row: %w", err)
		// }
		// if _, _, err := engine.InsertRow([]string{"", "Xiaomi", "600", "3", ""}, "products"); err != nil {
		// 	return engine, fmt.Errorf("failed to insert sample product row: %w", err)
		// }


		// if _,_, err := engine.InsertRow([]string{"1", "1", "2"}, "orders"); err != nil {
		// 	fmt.Printf("failed to insert sample order row: %v", err)
		// }
		// if _,_, err := engine.InsertRow([]string{"2", "2", "2"}, "orders"); err != nil {
		// 	fmt.Printf("failed to insert sample order row: %v", err)
		// }
		// if _,_, err := engine.InsertRow([]string{"3", "3", "2"}, "orders"); err != nil {
		// 	fmt.Printf("failed to insert sample order row: %v", err)
		// }
		// if _,_, err := engine.InsertRow([]string{"4", "4", "2"}, "orders"); err != nil {
		// 	fmt.Printf("failed to insert sample order row: %v", err)
		// }
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
		// targetColumn1 := &engine.db.Tables["users"].Columns[3]
		// targetColumn2 := &engine.db.Tables["users"].Columns[4]

		// val1, err := targetColumn1.GetDefaultValue("123-456-7890")
		// if err != nil {
		// 	return engine, fmt.Errorf("failed to get default value for column %s: %w", targetColumn1.Name, err)
		// }
		// val2, err := targetColumn2.GetDefaultValue("30")
		// if err != nil {
		// 	return engine, fmt.Errorf("failed to get default value for column %s: %w", targetColumn2.Name, err)
		// }
		// // Serialize the value to match the format stored in the database
		// serializedValue1, err := entities.Serialize(val1, targetColumn1)
		// if err != nil {
		// 	return engine, fmt.Errorf("failed to serialize value for column %s: %w", targetColumn1.Name, err)
		// }
		// serializedValue2, err := entities.Serialize(val2, targetColumn2)
		// if err != nil {
		// 	return engine, fmt.Errorf("failed to serialize value for column %s: %w", targetColumn2.Name, err)
		// }

		// searchCondition1 := &SearchCondition{
		// 	ColumnName: targetColumn1.Name,
		// 	Operator: "=",
		// 	Value: serializedValue1,
		// }
		// searchCondition2 := &SearchCondition{
		// 	ColumnName: targetColumn2.Name,
		// 	Operator: "=",
		// 	Value: serializedValue2,
		// }
		// expression := &Expression{
		// 	Type: NodeOr,
		// 	Condition: &SearchCondition{
		// 		ColumnName: targetColumn1.Name,
		// 		Operator: "=",
		// 		Value: serializedValue1,
		// 	},
		// 	Left: &Expression{
		// 		Type: NodeCondition,
		// 		Condition: searchCondition1,
		// 	},
		// 	Right: &Expression{
		// 		Type: NodeCondition,
		// 		Condition: searchCondition2,
		// 	},
		// }
		// result, err := engine.Search("users", []string{"email","id", "name"}, expression)
		// if err != nil {
		// 	return engine, fmt.Errorf("failed to perform search: %w", err)
		// }
		// fmt.Println("Search results:")
		// for _, row := range result {
		// 	fmt.Println(row)
		// }
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


func (engine *StorageEngine) Search(tableName string, cols []string, expr *Expression) ([][]any, error) {
	// Get the table object from the database
	table, ok := engine.db.Tables[tableName]
	if !ok {
		return nil, fmt.Errorf("Table %s not found", tableName)
	}
	if cols[0] == "*" {
		cols = make([]string, len(table.Columns))
		for i, col := range table.Columns {
			cols[i] = col.Name
		}
	}else {
		// Validate that the specified columns exist in the table
		for _, colName := range cols {
			if _, err := table.GetColumnByName(colName); err != nil {
				return nil, fmt.Errorf("Column %s not found in table %s", colName, tableName)
			}
		}
	}
	columnOffsets := make(map[string]int)
	if expr != nil { 
		err := SerializeSearchExpression(expr, table)
		if err != nil {
			return nil, fmt.Errorf("An Error Occured %w", err)
		}
	}

	// If there's only one condition, check if it has an index and use the indexed search
	if expr != nil && expr.Type == NodeCondition && (expr.Condition.Operator == "=" || expr.Condition.Operator == "==") {
		condition := expr.Condition
		if condition != nil {
			rootID, exists := table.Indexes[condition.ColumnName]// Check if the column has an index
			if exists {
				// If the column has an index, use the indexed search
				column, _ := table.GetColumnByName(condition.ColumnName)
				pageId, slot, _ := engine.IndexSearch(rootID, condition.Value, column)
				
				if pageId == 0 && slot == 0 {
					return [][]any{}, nil
				}else {
					// If the key is found, read the row
					dataBuffer, err := engine.Bp.Get(pageId)
					if err != nil {
						return nil, fmt.Errorf("An Error Occured %w", err)
					}
					tableOffset, err  := pages.GetDataPageSlotOffset(dataBuffer, slot)
					if errors.Is(err, pages.ErrRowNotFound){
						return [][]any{}, nil
					}
					if err != nil {
						return nil,fmt.Errorf("an error occured while Reading Row: %w", err)
					}
					row, err := engine.ReadRow(tableName, cols, columnOffsets,dataBuffer, tableOffset)
					if err != nil {
						return nil, fmt.Errorf("An Error Occured %w", err)
					}
					return [][]any{row}, nil
				}

			}
			
		}
	}
	// If the condition is more complex or the column doesn't have an index, use linear search
	results, err := engine.LinearSearch(tableName,cols,columnOffsets, expr)
	if err != nil {
		return nil, fmt.Errorf("An Error Occured %w", err)
	}

	return results, nil
}

func (engine *StorageEngine) Insert(tableName string, cols []string, values [][]string) (int, error) {
	// Get the table object from the database
	table, ok := engine.db.Tables[tableName]
	if !ok {
		return 0, fmt.Errorf("Table %s not found", tableName)
	}
	// Validate that the specified columns exist in the table
	for _, colName := range cols {
		if _, err := table.GetColumnByName(colName); err != nil {
			return 0, fmt.Errorf("Column %s not found in table %s", colName, tableName)
		}
	}
	insertedCount := 0
	for _, rowValues := range values {
		row := make([]string, len(table.Columns))
		for i, colName := range cols {
			columnIndex, _ := table.GetColumnIndexByName(colName)
			row[columnIndex] = rowValues[i]
		}
		if _, _, err := engine.InsertRow(row, tableName); err != nil {
			return insertedCount, fmt.Errorf("Failed to insert row: %w", err)
		}
		insertedCount++
	}
	return insertedCount, nil
}

func (engine *StorageEngine) Delete(tableName string, expr *Expression) (int, error) {
	table, ok := engine.db.Tables[tableName]
	if !ok {
		return 0, fmt.Errorf("Table %s not found", tableName)
	}
	columnOffsets := make(map[string]int)
	// Make sure the Search expression is ready for use by other functions
	if expr != nil { 
		err := SerializeSearchExpression(expr, table)
		if err != nil {
			return 0, fmt.Errorf("An Error Occured %w", err)
		}
	}
	// If there's only one condition, check if it has an index and use the indexed search
	if expr != nil && expr.Type == NodeCondition && (expr.Condition.Operator == "=" || expr.Condition.Operator == "==") {
		condition := expr.Condition
		if condition != nil {
			rootID, exists := table.Indexes[condition.ColumnName]// Check if the column has an index
			// If the column has an index, use the indexed search
			if exists {
				column, _ := table.GetColumnByName(condition.ColumnName)
				pageId, slot, err := engine.IndexSearch(rootID, condition.Value, column)
				
				if errors.Is(err, errKeyNotFound) {
					return 0, fmt.Errorf("Key not found")
				}else {
					// If the key is found, read the row
					dataBuffer, err := engine.Bp.Get(pageId)
					if err != nil {
						return 0, fmt.Errorf("An Error Occured %w", err)
					}
					err = engine.DeleteRow(table,columnOffsets,dataBuffer, pageId, slot)
					if err != nil {
						return 0, fmt.Errorf("An Error Occured %w", err)
					}
					return 1, nil
				}

			}
			
		}
	}
	result , err := engine.LinearDelete(table,columnOffsets, expr)
	if err != nil {
		return 0, fmt.Errorf("An Error Occured %w", err)
	}
	return result, nil
}

func (engine *StorageEngine)  UpdateFreePage(pageID uint32, freeSpace uint16) {
	engine.db.UpdateFreePage(pageID, freeSpace)
	engine.metaWrite = true
}

func (engine *StorageEngine)  UpdateFreePageChange(pageID uint32, netChange int16) {
	engine.db.UpdateFreePageChange(pageID, netChange)
	engine.metaWrite = true
}