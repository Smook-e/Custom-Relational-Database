package storage


import (
	// "errors"
	"encoding/binary"
	"fmt"
	

	"github.com/Smook-e/Custom-Relational-Database/entities"
	"strings"
	"github.com/Smook-e/Custom-Relational-Database/pages"
)

func (engine *StorageEngine) ReadRow(tableName string, pageID uint32, slot uint16) ([]any, error) {
	table, ok := engine.db.Tables[tableName]
	if !ok {
		return nil, fmt.Errorf("Error: Table %s Not Found ", tableName)
	}
	Row := make([]any, len(table.Columns))
	
	buffer, err := engine.Bp.Get(pageID)
	if err != nil {
		return nil,err
	}
	tableOffset, err  := pages.GetDataPageSlotOffset(buffer, slot)
	if err != nil {
		return nil,fmt.Errorf("an error occured while Reading Row: %w", err)
	}
	offset := tableOffset
	for i, col := range table.Columns {
		switch col.DataType {
		case entities.TypeTinyInt://int8
			Row[i] = int8(buffer[offset])
			offset++
		case entities.TypeSmallInt://int16
			Row[i] = int16(binary.BigEndian.Uint16(buffer[offset:offset+2]))
			offset += 2
		case entities.TypeInt:
			
			Row[i] = int32(binary.BigEndian.Uint32(buffer[offset:offset+4]))
			offset += 4
		case entities.TypeBigInt:
			Row[i] = int64(binary.BigEndian.Uint64(buffer[offset:offset+8]))
			offset += 8
		case entities.TypeVarChar:
			length := uint8(buffer[offset])
			offset++
			Row[i] = string(buffer[offset:offset+uint16(length)])
			offset+= uint16(length)
		}
	}
	return Row, nil
}




// Function takesa the array of data as strings, uses a helper function to transform them into their suitable types
// then returns the Pageid and slot the row was inserted at
func  (engine *StorageEngine) InsertRow( data []string, tableName string) (uint32, uint16, error) {
	//Pass 1: Check Validity and calculate size
	table, ok := engine.db.Tables[tableName]
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
	pageID, err := pages.FindFreePage(engine.db, size)
	if err != nil {
		return 0,0,fmt.Errorf("An error occured while finding free page: %w", err)
	}
	buffer, err := engine.Bp.Get(pageID)
	if err != nil {
		return 0,0,fmt.Errorf("An error occured while inserting: %w", err)
	}
	engine.Bp.MarkDirty(pageID)
	freeSpaceOffset,slot, err := pages.FindandUpdateDataPageSlot(buffer, size)

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
	//Get the primary key column name
	primaryKeyColumnName := table.PrimaryKeyColumn
	// Get the index of the primary key column
	primaryKeyColumnIndex, err := table.GetColumnIndexByName(primaryKeyColumnName)
	if err != nil {
		return 0,0, fmt.Errorf("An error occured while inserting: %w", err)
	}
	// Get the actual primary key column
	primaryKeyColumn := table.Columns[primaryKeyColumnIndex]
	// Serialize the primary key value into a byte slice to use for indexing
	serializedKey,err := engine.db.Serialize(vals[primaryKeyColumnIndex], primaryKeyColumn.DataType)
	if err != nil {
		return 0,0, fmt.Errorf("An error occured while inserting: %w", err)
	}
	engine.InsertIntoIndex(table.Indexes[primaryKeyColumnName], serializedKey, pageID, slot, primaryKeyColumn.DataType)
	
	return pageID, slot, nil
}
func (engine *StorageEngine) CreateTable(tableName string, cols []entities.ColumnDefinition) (error) {

	table := &entities.Table{Name: tableName}
	for _, col := range cols{
		cleanst := strings.ToLower(col.DataType)
		dataType, err := entities.GetDataType(cleanst)
		if err != nil {
			return  err
		}
		constraints,isPrimaryKey, err  := entities.GetConstraint(col.Constraints)
		if err != nil {
			return  err
		}
		if isPrimaryKey {
			table.PrimaryKeyColumn = col.Name
		}

		size, err := entities.GetSize(dataType)
		if err != nil {
			return  err
		}
		table.Columns = append(table.Columns, entities.Column{
			Name:        col.Name,
			DataType:    dataType,
			Constraints: constraints,
			Size:        size,
		})
	}
	engine.db.Tables[tableName] = table
	engine.db.Tables[tableName].Indexes = make(map[string]uint32)
	return nil
}