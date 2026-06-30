package pages

import (
	// "errors"
	"encoding/binary"
	"fmt"
	"strings"
	"sync"

	"github.com/Smook-e/Custom-Relational-Database/entities"
)


func InsertRow(db *entities.Database, data []string, tableName string) (uint32, uint16, error) {
	//Pass 1: Check Validity and calculate size
	table, ok := db.Tables[tableName]
	if !ok {
		return 0,0, fmt.Errorf("Error: Table %s Not Found ", tableName)
	}
	if len(table.Columns) != len(data){
		return 0,0, fmt.Errorf("Error: Invalid input size. Please enter %d Fields", len(table.Columns))
	}
	vals, size, err := table.GetValues(data)
	if err != nil {
		return 0,0,fmt.Errorf("An error occured while inserting: %w", err)
	}
	buffer, freeSpaceOffset, err := GetDataPage(db, size)
	if err != nil {
		return 0,0,fmt.Errorf("An error occured while inserting: %w", err)
	}
	offset := freeSpaceOffset
	for i, val := range vals {
		switch v := val.(type) {
		case int8:
			buffer[offset] = byte(v)// doesnt work
		}
	}	
}