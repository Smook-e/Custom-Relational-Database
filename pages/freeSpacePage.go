package pages

import (
	// "os"
	// "container/list"
	"encoding/binary"
	"errors"
	"fmt"

	// "os"

	"github.com/Smook-e/Custom-Relational-Database/entities"
	"github.com/Smook-e/Custom-Relational-Database/filehandler"
	// "github.com/Smook-e/Custom-Relational-Database/storage"
)


func ReadFreeSpacePage(db *entities.Database) error {
	buffer := bufferPool.Get().([]byte)
	defer bufferPool.Put(buffer)
	
	err := filehandler.ReadFromFile(db.File, 1, buffer)
	if err != nil {
		return fmt.Errorf("An error occured while reading the Free space pages: %w", err)
	}

	offset := 0
	// nextPagePointer := binary.BigEndian.Uint16(buffer[offset: offset + 2]);
	offset += 2;// Read Next Page Id

	numberOfElements := binary.BigEndian.Uint16(buffer[offset: offset + 2]); offset += 2;// Read Number of elements

	freeSpaces := make([]entities.FreePage, 0, numberOfElements)
	var pageID uint32; var freeSpace uint16;

	for range numberOfElements {
		pageID = binary.BigEndian.Uint32(buffer[offset:offset + 4]); offset += 4;
		freeSpace = binary.BigEndian.Uint16(buffer[offset: offset+2]);offset += 2;
		freeSpaces = append(freeSpaces, entities.FreePage{PageID: pageID, FreeSpace: freeSpace})
	}
	db.FreePages = freeSpaces


	return nil
}

func WriteFreeSpacePage(db *entities.Database) error {
	buffer := bufferPool.Get().([]byte)
	defer bufferPool.Put(buffer)

	offset := 0
	var nextPagePointer uint16 = 0
	binary.BigEndian.PutUint16(buffer[offset: offset + 2], nextPagePointer); offset += 2 ;//Write next page ID
	binary.BigEndian.PutUint16(buffer[offset: offset + 2] ,uint16(len(db.FreePages))); offset += 2;// number of elements in the page
	for _, page := range db.FreePages {
		//PageID
		binary.BigEndian.PutUint32(buffer[offset: offset + 4], page.PageID); offset += 4
		//FreeSpace
		binary.BigEndian.PutUint16(buffer[offset: offset + 2], page.FreeSpace); offset += 2;
	}
	
	filehandler.WriteToFile(db.File, 1, buffer)

	return nil
}
func FindFreePage(db *entities.Database, requiredSpace uint16) (uint32, error) {
	if requiredSpace > bufferSize - 7 {
		return 0, errors.New("No page has more than 4089 free bytes")
	}
	for _, freePage := range db.FreePages {
		if freePage.FreeSpace >= requiredSpace {
			freePage.FreeSpace -= requiredSpace
			return freePage.PageID, nil
		}
	}
	return uint32(db.TotalPages), nil
}