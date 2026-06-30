package storage

import (
	// "errors"
	"fmt"
	"strings"
	"sync"

	"github.com/Smook-e/Custom-Relational-Database/entities"
	"github.com/Smook-e/Custom-Relational-Database/pages"
)
const bufferSize = 4096

var bufferPool = sync.Pool{
    New: func() interface{} {
        // This ensures every buffer produced by the pool is 4KB
        return make([]byte, bufferSize)
    },
}


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
	pageID, err := pages.FindFreePage(db,size)
	if err != nil {
		return 0,0,fmt.Errorf("An error occured while inserting: %w", err)
	}
	buffer := bufferPool.Get().([]byte)
	defer bufferPool.Put(buffer)
}