package pages

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/Smook-e/Custom-Relational-Database/entities"
	"github.com/Smook-e/Custom-Relational-Database/filehandler"
	
)
/*
This file contains utility functions for managing free space pages in the database.
It is responsible for reading and writing free space pages, managing free space information, and finding suitable pages for new records.
Used in the storage layer when inserting new records.
*/

// ReadFreeSpacePage reads the free space page from the database file and populates the FreePages slice in the Database struct.
// Usually called during the initialization of the storage engine to load the free space information into memory.
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

// WriteFreeSpacePage writes the current state of the FreePages slice in the Database struct to the free space page in the database file on disk.
// Usually called during the commit operation to persist the free space information to disk.
func WriteFreeSpacePage(db *entities.Database) error {
	buffer := bufferPool.Get().([]byte)
	defer bufferPool.Put(buffer)

	offset := 0
	var nextPagePointer uint16 = 0
	binary.BigEndian.PutUint16(buffer[offset: offset + 2], nextPagePointer); offset += 2 ;//Write next page ID
	binary.BigEndian.PutUint16(buffer[offset: offset + 2] ,uint16(len(db.FreePages))); offset += 2;// number of elements in the page
	for _, page := range db.FreePages {
		// fmt.Println("Wrote to FreeSpaceFile", page.PageID, page.FreeSpace)
		//PageID
		binary.BigEndian.PutUint32(buffer[offset: offset + 4], page.PageID); offset += 4
		//FreeSpace
		binary.BigEndian.PutUint16(buffer[offset: offset + 2], page.FreeSpace); offset += 2;
	}
	filehandler.WriteToFile(db.File, 1, buffer)

	return nil
}
// FindFreePage searches for a free page in the database that has enough space to accommodate a new record of the specified size.
// If a suitable page is found, it returns the page ID. If no suitable page is found, it initializes a new data page and returns its page ID.
// Usually called during the insert operation to find a page for a new record.
func FindFreePage(db *entities.Database, requiredSpace uint16) (uint32, error) {
	if requiredSpace > bufferSize - 7 {
		return 0, errors.New("No page has more than 4089 free bytes")
	}
	for i := range db.FreePages {
		if db.FreePages[i].FreeSpace >= requiredSpace + 2 { // 2 bytes for the new slot that points to the new row
			db.FreePages[i].FreeSpace -= (requiredSpace) // required space + 2 bytes for the new slot the points to it
			return db.FreePages[i].PageID, nil
		}
	}
	err := InitializeNewDataPage(db, requiredSpace)
	// fmt.Println("From findfreepage: Initializing a new page. Returning page", db.TotalPages - 1)
	if err != nil {
		return 0,err
	}
	return uint32(db.TotalPages - 1), nil
}