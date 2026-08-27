package pages

import (
	"encoding/binary"
	
	
	"github.com/Smook-e/Custom-Relational-Database/entities"
	"github.com/Smook-e/Custom-Relational-Database/filehandler"
)
/*
This file contains utility functions for managing data pages in the database.
It is responsible for reading and writing data pages, managing free space, and handling slots within the pages.
Used in the storage layer to Insert and Read rows from the database file.
*/

//Function receives a pageid and slot and returns the specific offset of the slot
func GetDataPageSlotOffset(buffer []byte, slot uint16) (uint16, error) {

	var offset uint16 = 0
	offset += 2; // free space offset 2 bytes
	offset += 2 + slot * 2;//each slot has 2 bytes
	tableOffset := binary.BigEndian.Uint16(buffer[offset:offset+2]);// read the table offset from the specified slot
	return tableOffset, nil

}

//Function receives the required space by a row and the buffer where the row should be inserted.
// It updates the free space offset, number of elements, and adds the new slot to the next available position in the data page.
// it returns the offset where the new row should be inserted and the slot number of the new row.
// used in the storage layer to insert a new row into a data page.
func FindandUpdateDataPageSlot( buffer []byte, requiredSpace uint16) (uint16,uint16, error) {

	offset := 0
	freeSpaceOffset := binary.BigEndian.Uint16(buffer[offset:offset + 2]);
	freeSpaceOffset -= requiredSpace//update the free space offset
	binary.BigEndian.PutUint16(buffer[offset:offset + 2 ], freeSpaceOffset); offset += 2;
	
	numberOfElements := binary.BigEndian.Uint16(buffer[offset:offset + 2]);
	binary.BigEndian.PutUint16(buffer[offset: offset + 2], numberOfElements + 1)// update the number of elements
	offset += 2 + (int(numberOfElements) * 2)
	binary.BigEndian.PutUint16(buffer[offset: offset + 2], freeSpaceOffset)//add the new element at the next free slot
	
	return freeSpaceOffset,numberOfElements, nil
}

func InitializeNewDataPage(db *entities.Database, requiredSpace uint16) error {
	// fmt.Println("initializing new page")
	buffer := bufferPool.Get().([]byte)
	defer bufferPool.Put(buffer)
	offset := 0
	binary.BigEndian.PutUint16(buffer[offset: offset + 2], bufferSize - 1); offset += 2;// free space starts from the end of the file since it's still empty
	binary.BigEndian.PutUint16(buffer[offset:offset + 2], 0)// number of elements is 0 since it's a new page
	err := filehandler.WriteToFile(db.File, db.TotalPages, buffer)
	if err != nil {
		return err
	}
	//add the new free page to the database
	db.FreePages = append(db.FreePages, entities.FreePage{PageID:db.TotalPages,FreeSpace: bufferSize - 2 - 2 - 2 - requiredSpace})// 2 bytes freeSpaceOffset, 2 numberOfElements, 2 slot, -required space by row
	
	db.TotalPages++
	return nil
}