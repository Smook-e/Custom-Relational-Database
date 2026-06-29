package pages

import (
	// "os"
	// "container/list"
	"encoding/binary"
	"fmt"
	
	"os"

	"github.com/Smook-e/Custom-Relational-Database/entities"
	"github.com/Smook-e/Custom-Relational-Database/filehandler"
	"github.com/Smook-e/Custom-Relational-Database/storage"
)


func ReadFreeSpacePage(db *entities.Database) error {


	return nil
}
func WriteFreeSpacePage(db *entities.Database) error {
	buffer := make([]byte, bufferSize)

	offset := 0
	var nextPagePointer uint16 = 0
	binary.BigEndian.PutUint16(buffer, nextPagePointer); offset += 2 ;
	binary.BigEndian.PutUint16(buffer ,uint16(len(db.FreePages))); offset += 2;
	for _, page := range db.FreePages {
		//PageID
		binary.BigEndian.PutUint16()
	}

	return nil
}