package entities

import (
	"os"
	"fmt"
	// "errors"
)
const bufferSize = 4096

type Database struct {
	File *os.File
	Tables map[string]Table

	TotalPages int

}

func InitializeDatabase(filename string) (*Database, error) {
	filep, err :=  os.OpenFile("database.bin", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, fmt.Errorf("Critical Error: Could not open database file: %w", err)
	}
	fileInfo, err := filep.Stat()
	
	if err != nil {
		return nil, fmt.Errorf("Failed to retrieve file stats: %w", err)
	}
	db := &Database{
		File: filep,
		Tables: make(map[string]Table),
		TotalPages: int(fileInfo.Size() / bufferSize),
	}

	return db, nil
}