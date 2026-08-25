package storage

import (
	// "errors"
	"encoding/binary"
	"fmt"
	// "slices"

	"github.com/Smook-e/Custom-Relational-Database/entities"
	"github.com/Smook-e/Custom-Relational-Database/pages"
)

/*
This file contains functions for inserting and reading rows in the database, as well as creating new tables.
*/

// ReadRow reads a row from the specified table at the given page ID and slot, returning the row data as a slice of any type.
func (engine *StorageEngine) ReadRow(tableName string, cols []string, colIndexes []int, colOffsets map[string]int, buffer []byte, offset uint16) ([]any, error) {
	table, ok := engine.db.Tables[tableName]
	if !ok {
		return nil, fmt.Errorf("Error: Table %q Not Found ", tableName)
	}
	Row := make([]any, len(cols))
	
	
	// read the null bitmap first
	nullBitmap, err := table.ReadNullBitmap(buffer[offset:])
	if err != nil {
		return nil,fmt.Errorf("an error occured while Reading Row: %w", err)
	}
	nullBitmap , err = table.ReadNullBitmap(buffer[offset:])
	if err != nil {
		return nil,fmt.Errorf("an error occured while Reading Row: %w", err)
	}
	offset += uint16(len(nullBitmap.Bitmap))
	// First pass: assign the offsets for the columns we want to read
	if len(colOffsets) == 0 {
		GetColumnOffsets(table, buffer, offset, nullBitmap, colOffsets)
	}
	var colOffset int
	// Second pass: read the values for the specified columns
	for i, col := range cols {
		colOffset = colOffsets[col]
		if colOffset == -1 {
			Row[i] = nil // Column is null
			continue
		}
		switch table.Columns[colIndexes[i]].DataType {
		case entities.TypeTinyInt://int8
			Row[i] = int8(buffer[colOffset])

		case entities.TypeSmallInt://int16
			Row[i] = int16(binary.BigEndian.Uint16(buffer[colOffset:colOffset+2]))

		case entities.TypeInt:
			Row[i] = int32(binary.BigEndian.Uint32(buffer[colOffset:colOffset+4]))

		case entities.TypeSerial:// same as TypeInt
			Row[i] = int32(binary.BigEndian.Uint32(buffer[colOffset:colOffset+4]))

		case entities.TypeBigInt:
			Row[i] = int64(binary.BigEndian.Uint64(buffer[colOffset:colOffset+8]))
		case entities.TypeVarChar:
			length := uint8(buffer[colOffset])
			Row[i] = string(buffer[colOffset+1:colOffset+1+int(length)])
		}
		
	}
	return Row, nil
}


// InsertRow inserts a new row into the specified table with the provided data, returning the page ID and slot of the inserted row.
// it also checks for constraints such as not null, unique, primary key, and foreign key before inserting the row.
// and updates the B+Tree index for the table if necessary.
// it returns an error if any constraint is violated or if the insertion fails for any reason.
func  (engine *StorageEngine) InsertRow( data []string, tableName string) (uint32, uint16, error) {
	//Pass 1: Check Validity and calculate size
	table, ok := engine.db.Tables[tableName]
	if !ok {
		return 0,0, fmt.Errorf("Error: Table %q Not Found ", tableName)
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
		if col.DataType == entities.TypeSerial && vals[i] != nil {
			return 0,0, fmt.Errorf("Error: Column '%q' is of type Serial and cannot be manually set.", col.Name)
		}
		if col.HasConstraint(entities.ConstraintNotNull) && (vals[i] == nil || vals[i] == "") {
			if col.HasConstraint(entities.ConstraintDefault) {
				// If the column has a default constraint, use the default value instead of returning an error
				vals[i] = col.Default
				if col.DataType == entities.TypeSerial {
					// If the column is of type Serial, increment the default value for the next insertion
					col.Default = vals[i].(int32) + 1
					table.Columns[i].Default = col.Default // Update the column in the table with the new default value
					engine.metaWrite = true // Mark the meta page for writing since the default value has changed
				}
				// Clear the null bit from the null bitmap since we are using a default value
				nullBitmap.ClearNull(i)
				// Update the size to account for the default value
				defaultSize, err := entities.GetSize(&col)
				if err != nil {
					return 0,0, fmt.Errorf("An error occured while inserting: %w", err)
				}
				if defaultSize == 0 {// Varchar with no size specified, use the length of the default value
					size += uint16(len(col.Default.(string))) + 1 // +1 for length prefix
				}else{
					size += uint16(defaultSize)
				}
			}else {
				return 0,0, fmt.Errorf("Error: Column '%q' cannot be null", col.Name)
			}
		}
		if vals[i] != nil && (col.HasConstraint(entities.ConstraintUnique) || col.HasConstraint(entities.ConstraintPrimaryKey) || col.HasConstraint(entities.ConstraintIndex)) {
			// Check for uniqueness in the existing rows
			serializedKey, err := entities.Serialize(vals[i], &col)
			if err != nil {
				return 0,0, fmt.Errorf("An error occurred while serializing key: %w", err)
			}
			root := table.Indexes[col.Name]
			if pageID, _, _ := engine.IndexSearch(root, serializedKey, &col); pageID != 0 {
				return 0,0, fmt.Errorf("Error: Column %q must be unique. Value %v already exists.", col.Name, data[i])
			}
		}
		// Check for foreign key constraints
		if fk, exists := table.ForeignKeys[col.Name]; exists {
			referencedTable, ok := engine.db.Tables[fk.ReferencedTableName]
			if !ok {
				return 0,0, fmt.Errorf("Error: Referenced table %q not found for foreign key constraint on column %q", fk.ReferencedTableName, col.Name)
			}
			referencedCol := referencedTable.Columns[fk.ReferencedColumnIndex]
			serializedKey, err := entities.Serialize(vals[i], &col)
			if err != nil {
				return 0,0, fmt.Errorf("An error occurred while serializing key for foreign key check: %w", err)
			}
			root := referencedTable.Indexes[referencedCol.Name]
			if pageID, _, _ := engine.IndexSearch(root, serializedKey, &referencedCol); pageID == 0 {
				return 0,0, fmt.Errorf("Error: Foreign key constraint violation on column '%q'. Value %v does not exist in Column '%q' of referenced table '%q'.", col.Name, data[i], referencedCol.Name, fk.ReferencedTableName)
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
	err = table.WriteNullBitmap(nullBitmap, buffer[offset:offset+uint16(len(nullBitmap.Bitmap))])
	if err != nil {
		return 0,0,fmt.Errorf("An error occured while inserting: %w", err)
	}
	offset += uint16(len(nullBitmap.Bitmap))
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
	//Insert indexes
	for colName, root := range table.Indexes {
		colIndex, err := table.GetColumnIndexByName(colName)
		if err != nil {
			return 0,0, fmt.Errorf("An error occured while inserting: %w", err)
		}
		serializedKey, err := entities.Serialize(vals[colIndex], &table.Columns[colIndex])
		if err != nil {
			return 0,0, fmt.Errorf("An error occured while inserting: %w", err)
		}
		table.Indexes[colName] , err = engine.InsertIntoIndex(root, serializedKey, pageID, slot, &table.Columns[colIndex])
		if err != nil {
			return 0,0, fmt.Errorf("An error occured while inserting: %w", err)
		}
	}

	return pageID, slot, nil
}

// CreateTable creates a new table in the database with the specified name, columns, and foreign keys.
// It initializes the table's columns, constraints, and indexes, and writes the meta page to disk.
func (engine *StorageEngine) CreateTable(tableName string, cols []entities.ColumnDefinition, foreignKeys []entities.ForeignKeyDefinition) (error) {

	table := &entities.Table{Name: tableName}

	engine.db.Tables[tableName] = table
	engine.db.Tables[tableName].Indexes = make(map[string]uint32)
	var hasPrimaryKey bool
	createdColumns := make(map[string]bool)
	for _, col := range cols{
		if _, exists := createdColumns[col.Name]; exists {
			return fmt.Errorf("Error: Duplicate column name %q in table %q. Column names must be unique.", col.Name, tableName)
		}
		dataType,size, err := entities.GetDataTypeAndSize(col.DataType)
		if err != nil {
			return  fmt.Errorf("Error getting data type:%w", err)
		}
		constraints, err  := entities.GetConstraint(col.Constraints)
		if err != nil {
			return  fmt.Errorf("Error getting constraint:%w", err)
		}
		if constraints & entities.ConstraintPrimaryKey != 0 {
			if hasPrimaryKey {
				return fmt.Errorf("Error: Table %q already has a primary key defined. Only one primary key is allowed per table.", tableName)
			}
			hasPrimaryKey = true
		}
		newCol := entities.Column{
			Name:        col.Name,
			DataType:    dataType,
			Constraints: constraints,
			Size:        size,
		}
		createdColumns[col.Name] = true
		// If the column has a primary key, unique, or index constraint, create an index for it
		if constraints & entities.ConstraintIndex != 0 {
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
				return fmt.Errorf("Error getting default value for column %q: %w", col.Name, err)
			}
			newCol.Default = defaultValue
		}
		if dataType == entities.TypeSerial {
			// Set the default value for a serial column to 1 and mark it as not null
			newCol.Default = int32(1)
			newCol.Constraints |= entities.ConstraintNotNull
			newCol.Constraints |= entities.ConstraintDefault
		}
		

		// size, err := entities.GetSize(dataType)
		// if err != nil {
		// 	return  err
		// }
		table.Columns = append(table.Columns, newCol)
	}
	if !hasPrimaryKey {
		return fmt.Errorf("Error: Table %q must have a primary key defined. Please specify a primary key column.", tableName)
	}
	table.ForeignKeys = make(map[string]entities.ForeignKeyReference)
	for _, fk := range foreignKeys {
		// Validate that the column exists in the current table
		if _, err := table.GetColumnByName(fk.ColumnName); err != nil {
			return fmt.Errorf("Error: Column %q not found in table %q for foreign key constraint", fk.ColumnName, tableName)
		}
		referencedTable, ok := engine.db.Tables[fk.ReferencedTableName]
		if !ok || referencedTable.Name == table.Name{
			return fmt.Errorf("Error: Referenced table %q not found for foreign key constraint on column %q", fk.ReferencedTableName, fk.ColumnName)
		}
		referencedColIndex, err := referencedTable.GetColumnIndexByName(fk.ReferencedColumnName)
		if err != nil {
			return fmt.Errorf("Error getting referenced column index for foreign key: %w", err)
		}
		if !referencedTable.Columns[referencedColIndex].HasConstraint(entities.ConstraintPrimaryKey) {
			return fmt.Errorf("Error: Column %q in table %q is not a primary key, which is required for a foreign key constraint", fk.ReferencedColumnName, fk.ReferencedTableName)
		}
		table.ForeignKeys[fk.ColumnName] = entities.ForeignKeyReference{
			ReferencedTableName: fk.ReferencedTableName,
			ReferencedColumnIndex: uint8(referencedColIndex),
		}
	}
	engine.metaWrite = true
	
	return nil
}


func GetColumnOffsets(table *entities.Table,buffer []byte, offset uint16, nullBitmap *entities.NullBitmap, colOffsets map[string]int) {
	for i, col := range table.Columns {
		// Check if the column is null using the null bitmap
		if nullBitmap.IsNull(i) {
			colOffsets[col.Name] = -1 // Mark as null
			continue
		}
		switch col.DataType {
		case entities.TypeTinyInt://int8
			colOffsets[col.Name] = int(offset)
			offset++
		case entities.TypeSmallInt://int16
			colOffsets[col.Name] = int(offset)
			offset += 2
		case entities.TypeInt:
			colOffsets[col.Name] = int(offset)
			offset += 4
		case entities.TypeSerial:// same as TypeInt
			colOffsets[col.Name] = int(offset)
			offset += 4
		case entities.TypeBigInt:
			colOffsets[col.Name] = int(offset)
			offset += 8
		case entities.TypeVarChar:
			colOffsets[col.Name] = int(offset)
			length := uint8(buffer[offset])
			offset++
			if col.Size > 0 {
				offset += uint16(col.Size)
			}else {
				offset+= uint16(length)
			}
		}
	}
}