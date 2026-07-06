package storage


import (
	// "errors"
	
	"fmt"
	
	
	
	
	"github.com/Smook-e/Custom-Relational-Database/pages"
	"github.com/Smook-e/Custom-Relational-Database/entities"
)



func (engine *StorageEngine) TestOpenDatabase() error {
	
    
    
	// engine.db.File.Truncate(2 * bufferSize)
    
    
	pageID, slot, err := engine.InsertRow([]string{"1", "joe", "20"}, "users")
	if err != nil {
		return err
	}
	engine.Commit()
	Row, err := engine.ReadRow( "users", pageID, slot)
	if err != nil {
		return err
	}
	fmt.Println(Row)
	pageID, slot, err = engine.InsertRow( []string{"2", "emily", "25"}, "users")
	if err != nil {
		return err
	}
	engine.Commit()
	Row, err = engine.ReadRow("users", pageID, slot)
	if err != nil {
		return err
	}
	fmt.Println(Row)
	pageID, slot, err = engine.InsertRow([]string{"1", "IPhone", "1000", "2", "apple"}, "products")
	if err != nil {
		return err
	}
	engine.Commit()
	Row, err = engine.ReadRow("products", pageID, slot)
	fmt.Println(Row)
	pageID, slot, err = engine.InsertRow([]string{"2", "Macbook", "1200", "3", "apple"}, "products")
	if err != nil {
		return err
	}
	engine.Commit()
	Row, err = engine.ReadRow("products", pageID, slot)
	if err != nil {
		return err
	}
	engine.Commit()
	fmt.Println(Row)
	fmt.Println("Free Pages:")
    engine.db.PrintFreePages()
	pages.WriteMetaPage(engine.db)
	return nil
}


func (engine *StorageEngine) TestWriteandReadDatabase() error {
	

    
    engine.Bp.File.Truncate(0) // Clear the file before testing
    // initialize tables
    
    err := engine.db.CreateTable("products", []entities.ColumnDefinition{
        {Name: "id", DataType: "int", Constraints: []string{"primarykey", "notnull"}},
        {Name: "name", DataType: "varchar", Constraints: []string{"notnull"}},
        {Name: "price", DataType: "int", Constraints: []string{"notnull"}},
        {Name: "quantity", DataType: "int", Constraints: []string{"notnull"}},
        {Name: "seller", DataType: "varchar", Constraints: []string{"notnull"}},
    })
    if err != nil {
        return err
    }
    // engine.db.Tables[t1.Name] = t1

    // Table 2
    err = engine.db.CreateTable("users", []entities.ColumnDefinition{
        {Name: "id", DataType: "int", Constraints: []string{"primarykey"}},
        {Name: "name", DataType: "varchar", Constraints: []string{"notnull"}},
        {Name: "age", DataType: "int", Constraints: []string{}},
    })
    if err != nil {
        return err
    }
    engine.db.FreePages = []entities.FreePage{}
	
	
    
    // Write the meta page to the file
    err = pages.WriteMetaPage(engine.db)
    if err != nil {
        return fmt.Errorf("WriteMetaPage failed: %v", err)
    }

    // Close the file to ensure all data is flushed
    engine.db.File.Close() 
    

    
    // Reopen the database to test recovery
    engine, err = InitializeStorageEngine(engine.db.File.Name())
    if err != nil {
        return fmt.Errorf("OpenDatabase failed: %v", err)
    }
    defer engine.db.File.Close()

    
    engine.db.PrintTables()
	
	fmt.Println("Free Pages:")
    engine.db.PrintFreePages()
	pages.WriteMetaPage(engine.db)
	return nil
}