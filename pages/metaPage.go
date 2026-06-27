package pages

import (
	// "os"
	"errors"
	"github.com/Smook-e/Custom-Relational-Database/entities"
	"github.com/Smook-e/Custom-Relational-Database/filehandler"
	"encoding/binary"
)



func ReadMetaPage(db *entities.Database) error{
	buffer := make([]byte, 4096)

	filehandler.ReadFromFile(db.File,0, buffer)
	offset := 0
	nextPage := binary.BigEndian.Uint16(buffer[offset:offset+2]); offset += 2;
	freeSpaceOffset := binary.BigEndian.Uint16(buffer[offset:offset+2]); offset += 2;
	numberOfTables := binary.BigEndian.Uint16(buffer[offset:offset+2]); offset += 2;
	

	return nil
}