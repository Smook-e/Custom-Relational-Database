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
	binary.BigEndian.PutUint16(buffer[offset: offset + 2], nextPagePointer); offset += 2 ;
	binary.BigEndian.PutUint16(buffer[offset: offset + 2] ,uint16(len(db.FreePages))); offset += 2;// number of elements in the page
	for _, page := range db.FreePages {
		//PageID
		binary.BigEndian.PutUint32(buffer[offset: offset + 4], page.PageID); offset += 4
		//FreeSpace
		binary.BigEndian.PutUint16(buffer[offset: offset + 2], page.FreeSpace); offset += 2;
	}
	db.File.WriteAt(buffer, 1 * bufferSize)

	return nil
}