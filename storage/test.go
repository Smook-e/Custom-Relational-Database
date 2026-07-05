package storage



func TestOpenDatabase(filename string) error {
	db, err := OpenDatabase(filename)
    if err != nil {
        return fmt.Errorf("OpenDatabase failed: %v", err)
    }
    defer db.File.Close()
	// db.File.Truncate(2 * bufferSize)
    
    
	pageID, slot, err := InsertRow(db, []string{"1", "joe", "20"}, "users")
	if err != nil {
		return err
	}
	Row, err := ReadRow(db, "users", pageID, slot)
	if err != nil {
		return err
	}
	fmt.Println(Row)
	pageID, slot, err = InsertRow(db, []string{"2", "emily", "25"}, "users")
	if err != nil {
		return err
	}
	Row, err = ReadRow(db, "users", pageID, slot)
	if err != nil {
		return err
	}
	fmt.Println(Row)
	pageID, slot, err = InsertRow(db, []string{"1", "Phone", "1000"}, "products")
	if err != nil {
		return err
	}
	Row, err = ReadRow(db, "products", pageID, slot)
	if err != nil {
		return err
	}
	fmt.Println(Row)
	fmt.Println("Free Pages:")
    for _, freePage := range db.FreePages {
        fmt.Printf(" Page: %d | Free Space: %d\n", freePage.PageID, freePage.FreeSpace)
    }
	WriteMetaPage(db)
	return nil
}