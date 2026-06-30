package pages

import "github.com/Smook-e/Custom-Relational-Database/entities"


func GetDataPage(db *entities.Database, requiredSpace uint16) (uint16, error) {
	pageID, err:= FindFreePage(db, requiredSpace)
	if err != nil {
		return 0, err
	}
	// if no Freepage was found

	//Read the Page and determine the slot 
	buffer := bufferPool.Get().([]byte)
	defer bufferPool.Put(buffer)


	return 0, nil
}