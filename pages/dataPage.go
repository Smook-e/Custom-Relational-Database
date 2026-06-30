package pages

import (
	"github.com/Smook-e/Custom-Relational-Database/entities"
	"github.com/Smook-e/Custom-Relational-Database/filehandler"
	"encoding/binary"
	


)


func GetDataPage(db *entities.Database, requiredSpace uint16) ([]byte,uint16, error) {
	pageID, err:= FindFreePage(db, requiredSpace)
	if err != nil {
		return nil, 0, err
	}
	// if no Freepage was found

	//Read the Page and determine the slot 
	buffer := make([]byte, bufferSize)

	err = filehandler.ReadFromFile(db.File, int(pageID), buffer)
	if err != nil {
		return nil,0, err
	}
	offset := 0
	freeSpaceOffset := binary.BigEndian.Uint16(buffer[offset:offset + 2]);
	freeSpaceOffset -= requiredSpace
	binary.BigEndian.PutUint16(buffer[offset:offset + 2 ], freeSpaceOffset); offset += 2;

	numberOfElements := binary.BigEndian.Uint16(buffer[offset:offset + 2]);
	numberOfElements += 1;
	binary.BigEndian.PutUint16(buffer, numberOfElements)
	offset += 2 + (int(numberOfElements) * 2)
	binary.BigEndian.PutUint16(buffer, freeSpaceOffset)



	return buffer,freeSpaceOffset, nil
}