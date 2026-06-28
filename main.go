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

    
    
    

    
    filep, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
    if err != nil {
        log.Fatal(err)
    }

    db := &entities.Database{
        File:   filep,
        Tables: make(map[string]*entities.Table),
    }

    
    
    
    t1, err := storage.CreateTable("products", []entities.ColumnDefinition{
        {Name: "id", DataType: "int", Constraints: []string{"primarykey", "notnull"}},
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

    
    
    err = pages.WriteMetaPage(db)
    if err != nil {
        log.Fatalf("WriteMetaPage failed: %v", err)
    }

    
    db.File.Close() 
    fmt.Println("File closed and flushed.")

    
    
    db2, err := pages.OpenDatabase(filename)
    if err != nil {
        log.Fatalf("OpenDatabase failed: %v", err)
    }
    defer db2.File.Close()

    
    
    if len(db2.Tables) == 0 {
        fmt.Println("Error: No tables were recovered!")
    } else {
        for name, table := range db2.Tables {
            fmt.Printf("Table: %s | Columns: %d\n", name, len(table.Columns))
            for _, col := range table.Columns {
                fmt.Printf(" Column: %s | Type: %d | Constraints: %v\n", col.Name, col.DataType, col.Constraints)
            }
        }
    }

}