package pages

import (
	"github.com/Smook-e/Custom-Relational-Database/entities"
	"github.com/Smook-e/Custom-Relational-Database/filehandler"
	"encoding/binary"
	


)
//Function a pageid and slot and reads the page into a buffer and specifies the specific offset of the slot
func GetDataPage(db *entities.Database,pageID uint32, slot uint16) ([]byte,uint16, error) {
	buffer := make([]byte, bufferSize)

	err := filehandler.ReadFromFile(db.File, int(pageID),buffer)
	if err != nil {
		return nil,0, err
	}
	var offset uint16 = 0
	offset += 2; // free space offset 2 bytes
	offset += slot * 2;//each slot has 2 bytes
	tableOffset := binary.BigEndian.Uint16(buffer[offset:offset+2]);// read the table offset from the specified slot
	return buffer, tableOffset, nil

}

//Function receives the required space by a row, and returns a buffer, freeSpace offset, numberOfElements(slot) and pageID
func FindDataPage(db *entities.Database, requiredSpace uint16) ([]byte,uint16,uint16,uint32,  error) {
	pageID, err:= FindFreePage(db, requiredSpace)
	if err != nil {
		return nil, 0,0,0, err
	}
	

	buffer := make([]byte, bufferSize)
	
	//Read the Page 
	err = filehandler.ReadFromFile(db.File, int(pageID), buffer)
	if err != nil {
		return nil,0,0,0, err
	}
	offset := 0

	freeSpaceOffset := binary.BigEndian.Uint16(buffer[offset:offset + 2]);
	freeSpaceOffset -= requiredSpace//update the free space offset
	binary.BigEndian.PutUint16(buffer[offset:offset + 2 ], freeSpaceOffset); offset += 2;

	numberOfElements := binary.BigEndian.Uint16(buffer[offset:offset + 2]);
	binary.BigEndian.PutUint16(buffer[offset: offset + 2], numberOfElements + 1)// update the number of elements
	offset += 2 + (int(numberOfElements) * 2)
	binary.BigEndian.PutUint16(buffer[offset: offset + 2], freeSpaceOffset)//add the new element at the next free slot

	return buffer,freeSpaceOffset,numberOfElements, pageID, nil
}

func InitializeNewDataPage(db *entities.Database) error {
	buffer := bufferPool.Get().([]byte)
	defer bufferPool.Put(buffer)
	offset := 0
	binary.BigEndian.PutUint16(buffer[offset: offset + 2], bufferSize); offset += 2;
	binary.BigEndian.PutUint16(buffer[offset:offset + 2], 0)
	err := filehandler.WriteToFile(db.File, uint32(db.TotalPages), buffer)
	if err != nil {
		return err
	}
	//add the new free page to the database
	db.FreePages = append(db.FreePages, entities.FreePage{PageID:db.TotalPages,FreeSpace: bufferSize - 4})// 4 bytes for the freespaceoffset and the numberofelements
	db.TotalPages++
	return nil
}