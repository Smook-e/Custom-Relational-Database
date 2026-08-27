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
	offset += 2; // number of elements 2 bytes
	offset += slot * 2;//each slot has 2 bytes
	tableOffset := binary.BigEndian.Uint16(buffer[offset:offset+2]);// read the table offset from the specified slot
	return tableOffset, nil

}
// Returns the starting offset of the previous slot in the data page, given the current slot number.
// if the current slot is the first slot, it returns the BufferSize
// Used to know the end point of the row data in the data page, since the row data is stored between the current ofsset and the previous offset.
func GetDataPageSlotOffsetEnd(buffer []byte, slot uint16) (uint16, error) {
	if slot == 0 {
		return bufferSize, nil
	}
	var offset uint16 = 0
	offset += 2; // free space offset 2 bytes
	offset += 2; // number of elements 2 bytes
	offset += (slot - 1) * 2
	tableOffset := binary.BigEndian.Uint16(buffer[offset:offset+2]);// read the table offset from the previous slot
	return tableOffset, nil
}
func UpdateDataPageSlotOffset(buffer []byte, slot uint16, newOffset uint16) error {
	var offset uint16 = 0
	offset += 2; // free space offset 2 bytes
	offset += 2; // number of elements 2 bytes
	offset += slot * 2;//each slot has 2 bytes
	binary.BigEndian.PutUint16(buffer[offset:offset+2], newOffset)
	return nil
}
func UpdateDataPageSlotOffsets(buffer []byte, slot uint16, netChange int16) error {
	var offset uint16 = 0
	offset += 2; // free space offset 2 bytes
	numberOfElements := binary.BigEndian.Uint16(buffer[offset:offset + 2])
	offset += 2; // number of elements 2 bytes
	offset += slot * 2 + 2;//each slot has 2 bytes + skip the current slot since we don't want to update it, we only want to update the slots after it
	for i := slot + 1; i < numberOfElements; i++ {
		currentOffset := binary.BigEndian.Uint16(buffer[offset:offset+2])
		newOffset := int16(currentOffset) + netChange
		binary.BigEndian.PutUint16(buffer[offset:offset+2], uint16(newOffset))
		offset += 2
	}
	return nil
}

func GetDataPageFreeSpace(buffer []byte) (uint16, error) {
	var offset uint16 = 0
	freeSpaceOffset := binary.BigEndian.Uint16(buffer[offset:offset + 2]);
	return freeSpaceOffset, nil
}
func UpdateDataPageFreeSpace(buffer []byte, newFreeSpace uint16) error {
	var offset uint16 = 0
	binary.BigEndian.PutUint16(buffer[offset:offset + 2], newFreeSpace);
	return nil
}

func GetFreeSpace(buffer []byte) (uint16, error) {
	var offset uint16 = 0
	freeSpaceOffset := binary.BigEndian.Uint16(buffer[offset:offset + 2]);
	numberOfElements := binary.BigEndian.Uint16(buffer[offset + 2:offset + 4]);
	usedSpace := numberOfElements * 2 // each slot takes 2 bytes
	freeSpace := freeSpaceOffset - usedSpace - 4 // 4 bytes for the free space offset and number of elements
	return freeSpace, nil
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