package storage


import (
	// "errors"
	
	"fmt"
	

	
	
	"github.com/Smook-e/Custom-Relational-Database/pages"
)



func (engine *StorageEngine) TestOpenDatabase(filename string) error {
	err := engine.OpenDatabase(filename)
    if err != nil {
        return fmt.Errorf("OpenDatabase failed: %v", err)
    }
    defer engine.db.File.Close()
	engine.db.File.Truncate(2 * bufferSize)
    
    
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