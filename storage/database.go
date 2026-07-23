package storage

import (
	// "errors"
	
	"fmt"
	"os"

	"github.com/Smook-e/Custom-Relational-Database/entities"
	
	"github.com/Smook-e/Custom-Relational-Database/pages"
)


func (engine *StorageEngine) OpenDatabase(filename string) ( error) {
	filep, err :=  os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return  fmt.Errorf("Critical Error: Could not open database file: %w", err)
	}
	fileInfo, err := filep.Stat()
	
	if err != nil {
		return  fmt.Errorf("Failed to retrieve file stats: %w", err)
	}
	engine.db = &entities.Database{
		File: filep,
		Tables: make(map[string]*entities.Table),
		FreePages: make([]entities.FreePage, 0),
		TotalPages: uint32(fileInfo.Size() / bufferSize),
	}
	fmt.Println("Total Pages", engine.db.TotalPages)
	err = pages.ReadMetaPage(engine.db)
	if err != nil {
		return  err
	}
	return nil
}
