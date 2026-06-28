package main

import (
	"fmt"
	"log"
	"os"

	// "encoding/binary"
	// "github.com/Smook-e/Custom-Relational-Database/filehandler"
	
	
	"github.com/Smook-e/Custom-Relational-Database/entities"
	"github.com/Smook-e/Custom-Relational-Database/pages"
	"github.com/Smook-e/Custom-Relational-Database/storage"
)



func main(){
filename := "database.bin"

    // 1. CLEAN START: Wipe the file to ensure no "garbage" from previous crashes
    err := os.Remove(filename)
    if err != nil && !os.IsNotExist(err) {
        log.Fatalf("Failed to clean old database file: %v", err)
    }
    fmt.Println("--- Step 1: File cleaned. Starting fresh. ---")

    // 2. SETUP: Create the initial DB state
    filep, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
    if err != nil {
        log.Fatal(err)
    }

    db := &entities.Database{
        File:   filep,
        Tables: make(map[string]*entities.Table),
    }

    // 3. CREATE DATA: Add tables to the map
    fmt.Println("--- Step 2: Creating tables in memory ---")
    
    // Table 1
    t1, err := storage.CreateTable("products", []entities.ColumnDefinition{
        {Name: "id", DataType: "int", Constraints: []string{"primarykey"}},
        {Name: "name", DataType: "varchar", Constraints: []string{"notnull"}},
        {Name: "price", DataType: "int", Constraints: []string{"notnull"}},
    })
    if err != nil {
        log.Fatal(err)
    }
    db.Tables[t1.Name] = t1

    // Table 2
    t2, err := storage.CreateTable("users", []entities.ColumnDefinition{
        {Name: "id", DataType: "int", Constraints: []string{"primarykey"}},
        {Name: "name", DataType: "varchar", Constraints: []string{"notnull"}},
        {Name: "age", DataType: "int", Constraints: []string{}},
    })
    if err != nil {
        log.Fatal(err)
    }
    db.Tables[t2.Name] = t2

    // 4. WRITE: Serialize the map to the meta page
    fmt.Println("--- Step 3: Writing Meta Page to disk ---")
    err = pages.WriteMetaPage(db)
    if err != nil {
        log.Fatalf("WriteMetaPage failed: %v", err)
    }

    // IMPORTANT: Close the file handle here! 
    // This flushes the OS cache and ensures OpenDatabase starts fresh.
    db.File.Close() 
    fmt.Println("File closed and flushed.")

    // 5. READ: Open a brand new database instance from the file
    fmt.Println("--- Step 4: Opening database and reading Meta Page ---")
    db2, err := pages.OpenDatabase(filename)
    if err != nil {
        log.Fatalf("OpenDatabase failed: %v", err)
    }
    defer db2.File.Close()

    // 6. VERIFY: Check if the map was reconstructed correctly
    fmt.Println("--- Final Results ---")
    if len(db2.Tables) == 0 {
        fmt.Println("Error: No tables were recovered!")
    } else {
        for name, table := range db2.Tables {
            fmt.Printf("Table: %s | Columns: %d\n", name, len(table.Columns))
            for _, col := range table.Columns {
                fmt.Printf("  -> Column: %s (Type: %d)\n", col.Name, col.DataType)
            }
        }
    }

}