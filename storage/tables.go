package storage


import (
	// "errors"
	"encoding/binary"
	"fmt"
	

	"github.com/Smook-e/Custom-Relational-Database/entities"
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
	// read the null bitmap first
	nullBitmapSize := (len(table.Columns) + 7) / 8 // Calculate the size of the null bitmap in bytes
	nullBitmap := buffer[offset : offset+uint16(nullBitmapSize)]
	offset += uint16(nullBitmapSize)
	for i, col := range table.Columns {
		// Check if the column is null using the null bitmap
		byteIndex := i / 8
		bitIndex := uint(i % 8)
		if nullBitmap[byteIndex]&(1<<bitIndex) != 0 {
			Row[i] = nil
			continue
		}
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
			if col.Size > 0 {
				offset += uint16(col.Size)
			}else {
				offset+= uint16(length)
			}
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
	vals, size, nullBitmap, err := table.GetValues(data)
	if err != nil {
		return 0,0,fmt.Errorf("An error occured while inserting: %w", err)
	}
	
	// Validate constraints for each column
	for i, col := range table.Columns {
		if col.HasConstraint(entities.ConstraintNotNull) && (vals[i] == nil || vals[i] == "") {
			if col.HasConstraint(entities.ConstraintDefault) {
				// If the column has a default constraint, use the default value instead of returning an error
				vals[i] = col.Default
				// Clear the null bit from the null bitmap since we are using a default value
				byteIndex := i / 8
				bitIndex := uint(i % 8)
				nullBitmap[byteIndex] &^= (1 << bitIndex) // Clear the bit for this column
				// Update the size to account for the default value
				defaultSize, err := entities.GetSize(&col)
				if err != nil {
					return 0,0, fmt.Errorf("An error occured while inserting: %w", err)
				}
				if defaultSize == 0 {// Varchar with no size specified, use the length of the default value
					size += uint16(len(col.Default.(string))) + 1 // +1 for length prefix
				}
			}else {
				return 0,0, fmt.Errorf("Error: Column '%s' cannot be null", col.Name)
			}
		}
		if col.HasConstraint(entities.ConstraintUnique) || col.HasConstraint(entities.ConstraintPrimaryKey) || col.HasConstraint(entities.ConstraintIndex) {
			// Check for uniqueness in the existing rows
			serializedKey, err := engine.db.Serialize(vals[i], &col)
			if err != nil {
				return 0,0, fmt.Errorf("An error occurred while serializing key: %w", err)
			}
			root := table.Indexes[col.Name]
			if pageID, _, _ := engine.IndexSearch(root, serializedKey, &col); pageID != 0 {
				return 0,0, fmt.Errorf("Error: Column %s must be unique. Value %v already exists.", col.Name, data[i])
			}
		}
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
	// Write the null bitmap first
	copy(buffer[offset:offset+uint16(len(nullBitmap))], nullBitmap)
	offset += uint16(len(nullBitmap))
	for i, val := range vals {
		if val == nil {
			continue
		}
		// Write the value based on its type
		// Use a type switch to determine the type of the value and write it accordingly
		// Use binary.BigEndian to write multi-byte values in big-endian order
		// For strings, write the length first as a single byte, then write the string bytes
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
			if table.Columns[i].Size > 0 {
				offset += uint16(table.Columns[i].Size)
			} else {
				offset += uint16(len(v))
			}
		}
	}
	//Get the primary key column name
	primaryKeyColumn,primaryKeyColumnIndex, err := table.GetPrimaryKeyColumn()
	if err != nil {
		return 0,0, fmt.Errorf("An error occured while inserting: %w", err)
	}
	
	// Serialize the primary key value into a byte slice to use for indexing
	serializedKey,err := engine.db.Serialize(vals[primaryKeyColumnIndex], primaryKeyColumn)
	if err != nil {
		return 0,0, fmt.Errorf("An error occured while inserting: %w", err)
	}
	engine.InsertIntoIndex(table.Indexes[primaryKeyColumn.Name], serializedKey, pageID, slot, primaryKeyColumn)

	return pageID, slot, nil
}

func (engine *StorageEngine) CreateTable(tableName string, cols []entities.ColumnDefinition) (error) {

	table := &entities.Table{Name: tableName}
	engine.db.Tables[tableName] = table
	engine.db.Tables[tableName].Indexes = make(map[string]uint32)
	for _, col := range cols{
		
		dataType,size, err := entities.GetDataTypeAndSize(col.DataType)
		if err != nil {
			return  fmt.Errorf("Error getting data type:%w", err)
		}
		constraints, err  := entities.GetConstraint(col.Constraints)
		if err != nil {
			return  fmt.Errorf("Error getting constraint:%w", err)
		}
		newCol := entities.Column{
			Name:        col.Name,
			DataType:    dataType,
			Constraints: constraints,
			Size:        size,
		}
		// If the column has a primary key, unique, or index constraint, create an index for it
		if constraints & entities.ConstraintPrimaryKey != 0 || constraints & entities.ConstraintUnique != 0 || constraints & entities.ConstraintIndex != 0 {
			// Create an index for the primary key column
			// Create a new page for the index root
			root, err := engine.NewPage()
			if err != nil {
				return fmt.Errorf("Error creating new page:%w", err)
			}
			buffer , err := engine.Bp.Get(root)
			if err != nil {
				return fmt.Errorf("Error getting buffer for new page:%w", err)
			}
			// Initialize the index root page as a leaf page
			err =pages.InitializeLeafPage([]pages.LeafEntry{}, buffer)
			if err != nil {
				return fmt.Errorf("Error initializing leaf page:%w", err)
			}
			// Mark the page as dirty so it will be written to disk
			engine.Bp.MarkDirty(root)
			// Store the root page ID in the table's index map
			table.Indexes[col.Name] = root
		}
		// Set the default value for the column if it has a default constraint
		if constraints & entities.ConstraintDefault != 0 {
			defaultValue, err := newCol.GetDefaultValue(col.Default)
			if err != nil {
				return fmt.Errorf("Error getting default value for column %s: %w", col.Name, err)
			}
			newCol.Default = defaultValue
		}
		

		// size, err := entities.GetSize(dataType)
		// if err != nil {
		// 	return  err
		// }
		table.Columns = append(table.Columns, newCol)
	}
	
	return nil
}