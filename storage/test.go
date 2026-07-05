package storage


import (
	// "errors"
	
	"fmt"
	
	"os"
	
	
	"github.com/Smook-e/Custom-Relational-Database/pages"
	"github.com/Smook-e/Custom-Relational-Database/entities"
)



func (engine *StorageEngine) TestOpenDatabase(filename string) error {
	err := engine.OpenDatabase(filename)
    if err != nil {
        return fmt.Errorf("OpenDatabase failed: %v", err)
    }
    defer engine.db.File.Close()
	// engine.db.File.Truncate(2 * bufferSize)
    
    
	pageID, slot, err := engine.InsertRow([]string{"1", "joe", "20"}, "users")
	if err != nil {
		return err
	}
	Row, err := engine.ReadRow( "users", pageID, slot)
	if err != nil {
		return err
	}
	fmt.Println(Row)
	pageID, slot, err = engine.InsertRow( []string{"2", "emily", "25"}, "users")
	if err != nil {
		return err
	}
	Row, err = engine.ReadRow("users", pageID, slot)
	if err != nil {
		return err
	}
	fmt.Println(Row)
	pageID, slot, err = engine.InsertRow([]string{"1", "Phone", "1000"}, "products")
	if err != nil {
		return err
	}
	Row, err = engine.ReadRow("products", pageID, slot)
	if err != nil {
		return err
	}
	fmt.Println(Row)
	fmt.Println("Free Pages:")
    engine.db.PrintFreePages()
	pages.WriteMetaPage(engine.db)
	return nil
}


func (engine *StorageEngine) TestWriteandReadDatabase(filename string) error {
	filep, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
	filep.Truncate(0)
    if err != nil {
        return err
    }

    engine.db = &entities.Database{
		File:   filep,
		Tables: make(map[string]*entities.Table),
		FreePages: make([]entities.FreePage, 0),
	}

    
    // initialize tables
    
    t1, err := entities.CreateTable("products", []entities.ColumnDefinition{
        {Name: "id", DataType: "int", Constraints: []string{"primarykey", "notnull"}},
        {Name: "name", DataType: "varchar", Constraints: []string{"notnull"}},
        {Name: "price", DataType: "int", Constraints: []string{"notnull"}},
    })
    if err != nil {
        return err
    }
    engine.db.Tables[t1.Name] = t1

    // Table 2
    t2, err := entities.CreateTable("users", []entities.ColumnDefinition{
        {Name: "id", DataType: "int", Constraints: []string{"primarykey"}},
        {Name: "name", DataType: "varchar", Constraints: []string{"notnull"}},
        {Name: "age", DataType: "int", Constraints: []string{}},
    })
    if err != nil {
        return err
    }
    engine.db.Tables[t2.Name] = t2
	
	
    
    // Write the meta page to the file
    err = pages.WriteMetaPage(engine.db)
    if err != nil {
        return fmt.Errorf("WriteMetaPage failed: %v", err)
    }

    // Close the file to ensure all data is flushed
    engine.db.File.Close() 
    

    
    // Reopen the database to test recovery
    err = engine.OpenDatabase(filename)
    if err != nil {
        return fmt.Errorf("OpenDatabase failed: %v", err)
    }
    defer engine.db.File.Close()

    
    
    if len(engine.db.Tables) == 0 {
        fmt.Println("Error: No tables were recovered!")
    } else {
        for name, table := range engine.db.Tables {
            fmt.Printf("Table: %s | Columns: %d\n", name, len(table.Columns))

            for _, col := range table.Columns {
                fmt.Printf(" Column: %s | Type: %d | Constraints: %v\n", col.Name, col.DataType, col.Constraints)
            }
        }
    }
	
	fmt.Println("Free Pages:")
    engine.db.PrintFreePages()
	pages.WriteMetaPage(engine.db)
	return nil
}