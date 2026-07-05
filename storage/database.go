package storage

import (
	// "errors"
	"encoding/binary"
	"fmt"
	

	"github.com/Smook-e/Custom-Relational-Database/entities"
	"github.com/Smook-e/Custom-Relational-Database/filehandler"
	"github.com/Smook-e/Custom-Relational-Database/pages"
)


func OpenDatabase(filename string) (*entities.Database, error) {
	filep, err :=  os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, fmt.Errorf("Critical Error: Could not open database file: %w", err)
	}
	fileInfo, err := filep.Stat()
	
	if err != nil {
		return nil, fmt.Errorf("Failed to retrieve file stats: %w", err)
	}
	db := &entities.Database{
		File: filep,
		Tables: make(map[string]*entities.Table),
		TotalPages: uint32(fileInfo.Size() / bufferSize),
	}
	err = ReadMetaPage(db)
	if err != nil {
		return nil, err
	}
	return db, nil
}
