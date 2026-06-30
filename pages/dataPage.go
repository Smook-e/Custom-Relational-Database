package pages

import "github.com/Smook-e/Custom-Relational-Database/entities"


func GetDataPage(db *entities.Database, requiredSpace uint16) (uint16, error) {
	pageID, err:= FindFreePage(db, requiredSpace)
	

}