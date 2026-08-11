package entities

import (
	"os"
	"fmt"
	
)
/*
This file contains the definition of the Database struct, along with utility methods.
*/
// FreePage represents a page in the database file that has free space available for new records.
type FreePage struct {
	PageID	uint32
	FreeSpace	uint16
}
// Database represents the Metadata of the database, including the file handle, tables, free pages, and total number of pages.
type Database struct {
	File *os.File
	Tables map[string]*Table
	FreePages	[]FreePage
	TotalPages uint32
}

func (db *Database) PrintFreePages() {
	for _, freePage := range db.FreePages {
        fmt.Printf(" Page: %d | Free Space: %d\n", freePage.PageID, freePage.FreeSpace)
    }
}

func (db *Database) PrintTables() {
	
    if len(db.Tables) == 0 {
        fmt.Println("No tables were found in the database.")
    } else {
        for name, table := range db.Tables {
            fmt.Printf("******************************\nTable: %s | Columns: %d\n", name, len(table.Columns))
            
            for _, col := range table.Columns {
                fmt.Printf(" Column: %s | Type: %s | Constraints: %s | Size: %d\n", col.Name, col.PrintDataType(col.DataType), col.PrintConstraints(col.Constraints), col.Size)
            }
            fmt.Println("==================================")
            fmt.Println(" Indexes:")
            for indexName, indexID := range table.Indexes {
                fmt.Printf("  Index: %s | Page ID: %d\n", indexName, indexID)
            }
            fmt.Println("==================================")
            fmt.Println("Foreign Keys:")
            for fkName, fk := range table.ForeignKeys {
                fmt.Printf("  Foreign Key: %s | Referenced Column: %s.%s\n", fkName, fk.ReferencedTableName, db.Tables[fk.ReferencedTableName].Columns[fk.ReferencedColumnIndex].Name)
            }
        }
    }
}

// PrintDatabase prints high level information about the database,
// all tables and their columns, and free page info.
func (db *Database) PrintDatabase() {
    if db == nil {
        fmt.Println("Database is nil")
        return
    }

    // File information
    fileName := "<unknown>"
    if db.File != nil {
        fileName = db.File.Name()
    }
    fmt.Printf("Database file: %s\n", fileName)
    fmt.Printf("Total pages: %d\n", db.TotalPages)
    fmt.Printf("Number of tables: %d\n", len(db.Tables))

    // Tables and columns
    db.PrintTables()

    // Free pages
    fmt.Println("Free Pages:")
    db.PrintFreePages()
}
