package pages

import (
	"github.com/Smook-e/Custom-Relational-Database/entities"
	"github.com/Smook-e/Custom-Relational-Database/filehandler"
	"encoding/binary"
	


)


func GetDataPage(db *entities.Database, requiredSpace uint16) ([]byte,uint16,uint16,uint32,  error) {
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