package pages

import (
	// "errors"
	"encoding/binary"
	"fmt"
	

	"github.com/Smook-e/Custom-Relational-Database/entities"
	"github.com/Smook-e/Custom-Relational-Database/filehandler"
)

func ReadRow(db *entities.Database,tableName string, pageID uint32, slot uint16) ([]any, error) {
	table, ok := db.Tables[tableName]
	if !ok {
		return nil, fmt.Errorf("Error: Table %s Not Found ", tableName)
	}

}




// Function takesa the array of data as strings, uses a helper function to transform them into their suitable types
// then returns the Pageid and slot the row was inserted at
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
	//Get a suitable data page and slot to insert into
	buffer, freeSpaceOffset,slot,pageID, err := FindDataPage(db, size)
	if err != nil {
		return 0,0,fmt.Errorf("An error occured while inserting: %w", err)
	}
	offset := freeSpaceOffset
	//Pass 2: write the values into the page
	for _, val := range vals {
		//cast the value into its type first
		switch v := val.(type) {
		case int8:
			buffer[offset] = byte(v)
			offset++
		case int16:
			binary.BigEndian.PutUint16(buffer[offset: offset+2], uint16(v))
			offset+=2
		case int32:
			binary.BigEndian.PutUint32(buffer[offset: offset+4], uint32(v))
			offset+=4
		case int64:
			binary.BigEndian.PutUint64(buffer[offset: offset+8], uint64(v))
			offset+=8
		case string:
			buffer[offset] = uint8(len(v))
			offset++
			copy(buffer[offset:], v)
			offset += uint16(len(v))
		}
	}
	err = filehandler.WriteToFile(db.File,uint32(pageID), buffer)
	if err != nil {
		return 0,0,fmt.Errorf("Failed to Commit Row to disk:%w",err)
	}
	return pageID, slot, nil
}