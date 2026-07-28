package entities

import (
	"os"
	"fmt"
	// "errors"
	
)

type FreePage struct {
	PageID	uint32
	FreeSpace	uint16
}
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
            fmt.Printf("Table: %s | Columns: %d\n", name, len(table.Columns))

            for _, col := range table.Columns {
                fmt.Printf(" Column: %s | Type: %d | Constraints: %v\n", col.Name, col.DataType, col.Constraints)
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
