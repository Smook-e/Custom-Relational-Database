package storage

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/Smook-e/Custom-Relational-Database/entities"
	"github.com/Smook-e/Custom-Relational-Database/pages"
)

/*
Provides a linear search implementation for searching through the B+Tree index in the database.
The B+Tree Leaf pages are linked together in a linked list, allowing for efficient linear traversal of the keys in the index.
*/


type SearchCondition struct {
	ColumnName string
	Operator string
	Value []byte
}
type NodeType uint8

const (
	NodeCondition NodeType = 1
	NodeAnd       NodeType = 2
	NodeOr        NodeType = 3
)

type Expression struct {
	Type NodeType
	
	Condition *SearchCondition

	Left  *Expression
	Right *Expression
}

func (engine *StorageEngine) GetFirstLeafPage(rootId uint32) (uint32, error) {
	//load page
	buffer, err := engine.Bp.Get(rootId)
	if err != nil {
		return 0 ,fmt.Errorf("An Error Occured %w", err)
	}
	//check if leaf
	isleaf := buffer[0]
	if isleaf == 1 {
		return rootId, nil
	}
	//get first child
	firstChildId := binary.BigEndian.Uint32(buffer[InternalPageHeaderSize:InternalPageHeaderSize + 4])
	return engine.GetFirstLeafPage(firstChildId)
}

func (engine *StorageEngine) LinearTree(rootId uint32, key []byte, col *entities.Column) (uint32, uint16, error) {
	leafId, err := engine.GetFirstLeafPage(rootId)
	if err != nil {
		return 0,0, fmt.Errorf("An Error Occured %w", err)
	}
	offset := 0
	for leafId != 0 {
		//load leaf page
		buffer, err := engine.Bp.Get(leafId)
		if err != nil {
			return 0 ,0,fmt.Errorf("An Error Occured %w", err)
		}
		
		//get next leaf page
		leafId = binary.BigEndian.Uint32(buffer[nextLeafPageOffset:nextLeafPageOffset + 4])
		//get number of keys
		numKeys := binary.BigEndian.Uint16(buffer[leafPageNumEntriesOffset:leafPageNumEntriesOffset + 2])
		offset = LeafPageHeaderSize
		for range numKeys {
			comp , err := entities.Compare(key, buffer[offset: offset + len(key)], col)
			if err != nil {
				return 0,0, fmt.Errorf("An Error Occured %w", err)
			}
			offset += len(key)
			pageId := binary.BigEndian.Uint32(buffer[offset:offset + 4])
			offset += 4
			slot := binary.BigEndian.Uint16(buffer[offset:offset + 2])
			offset += 2
			if comp == 0 {
				return pageId, slot, nil
			}
			if comp < 0 {
				return 0,0, fmt.Errorf("Key not found")
			}
		}
		
	}
	
	
	return 0,0, fmt.Errorf("Key not found")
}
func (engine *StorageEngine) EvaluateExpression(buffer []byte, colOffsets map[string]int, expr *Expression, table *entities.Table) (bool, error) {
	if expr == nil {
		return true, nil
	}
	switch expr.Type {
	case NodeCondition:
		return VerifyCondition(buffer, colOffsets, expr.Condition, table)
	case NodeAnd:
		leftResult, _ := engine.EvaluateExpression(buffer, colOffsets, expr.Left, table)
		if leftResult == false {
			return false, nil
		}
		return engine.EvaluateExpression(buffer, colOffsets, expr.Right, table)
	case NodeOr:
		leftResult, _ := engine.EvaluateExpression(buffer, colOffsets, expr.Left, table)
		if leftResult == true {
			return true, nil
		}
		return engine.EvaluateExpression(buffer, colOffsets, expr.Right, table)	
	default:
		return false, fmt.Errorf("Unknown expression type: %v", expr.Type)
	}
}
func VerifyCondition(buffer []byte,colOffsets map[string]int, condition *SearchCondition, table *entities.Table) (bool, error) {
	if condition == nil {
		return true, nil
	}
	col, _ := table.GetColumnByName(condition.ColumnName)
	size, err := entities.GetSize(col)
	if err != nil {
		return false, fmt.Errorf("An error occurred while getting size of column %s: %w", col.Name, err)
	}
	
	colOffset, exists := colOffsets[condition.ColumnName]
	if !exists {
		return false, fmt.Errorf("Column %s not found in offsets map", condition.ColumnName)
	}
	// Check if the column is null 
	if colOffset == -1 {
		return false, nil
	}
	if size == 0 {// For variable-length types
		size = buffer[colOffset] + 1 // First byte indicates the length of the variable-length data
	}
	// Compare the value in the buffer with the condition value
	comp, err := entities.Compare(buffer[colOffset:colOffset+int(size)], condition.Value, col)
	if err != nil {
		return false, fmt.Errorf("An error occurred while comparing values: %w", err)
	}
	
	switch condition.Operator {
	case "=" , "==":
		return comp == 0, nil
	case "!=" , "<>":
		return comp != 0, nil
	case "<" :
		return comp < 0, nil
	case ">":
		return comp > 0, nil
	case "<=":
		return comp <= 0, nil
	case ">=":
		return comp >= 0, nil
	default:
		return false, fmt.Errorf("Unsupported operator: %s", condition.Operator)
	}
}


func (engine *StorageEngine) LinearSearch(tableName string, cols []string, colOffsets map[string]int, expr *Expression) ([][]any, error) {
	// Get the table object from the database
	table, ok := engine.db.Tables[tableName]
	if !ok {
		return nil, fmt.Errorf("Table %s not found", tableName)
	}
	// Get the Primary Key Column Name to start searching
	primaryKeyColumn, _, err := table.GetPrimaryKeyColumn()
	if err != nil {
		return nil, fmt.Errorf("An Error Occured %w", err)
	}
	rootId, ok := table.Indexes[primaryKeyColumn.Name]
	if !ok {
		return nil, fmt.Errorf("Primary key index not found for table %s", tableName)
	}
	// Get the first leaf page of the B+ tree (linked list of leaf pages)
	leafId, err := engine.GetFirstLeafPage(rootId)
	if err != nil {
		return nil, fmt.Errorf("An Error Occured %w", err)
	}
	offset := 0
	var results [][]any
	for leafId != 0 {
		//load leaf page
		buffer, err := engine.Bp.Get(leafId)
		if err != nil {
			return nil, fmt.Errorf("An Error Occured %w", err)
		}
		//get next leaf page
		leafId = binary.BigEndian.Uint32(buffer[nextLeafPageOffset:nextLeafPageOffset + 4])
		//get number of keys
		numKeys := binary.BigEndian.Uint16(buffer[leafPageNumEntriesOffset: leafPageNumEntriesOffset + 2])
		offset = LeafPageHeaderSize
		for range numKeys {
			//skip the key bytes
			offset += int(primaryKeyColumn.Size)
			pageId := binary.BigEndian.Uint32(buffer[offset:offset + 4])
			offset += 4
			slot := binary.BigEndian.Uint16(buffer[offset:offset + 2])
			offset += 2
			// Find the Row in the table using the pageId and slot
			dataBuffer, err := engine.Bp.Get(pageId)
			if err != nil {
				return nil,err
			}
			tableOffset, err  := pages.GetDataPageSlotOffset(dataBuffer, slot)
			if errors.Is(err, pages.ErrRowNotFound){
				continue
			}
			if err != nil {
				return nil,fmt.Errorf("an error occured while Reading Row: %w", err)
			}
			// Get the column offsets
			// Read the null bitmap first
			nullBitmap, err := table.ReadNullBitmap(dataBuffer[tableOffset:])
			if err != nil {
				return nil, fmt.Errorf("An error occurred while reading null bitmap: %w", err)
			}
			tableOffset += uint16(len(nullBitmap.Bitmap)) // Move the offset past the null bitmap
			GetColumnOffsets(table,dataBuffer, tableOffset,nullBitmap, colOffsets)
			//check the condition
			conditionMet, err := engine.EvaluateExpression(dataBuffer, colOffsets, expr, table)
			if err != nil {
				return nil, fmt.Errorf("An Error Occured %w", err)
			}
			if conditionMet {
				// Read the pageId and slot from the buffer
				row, err := engine.ReadRow(tableName, cols, colOffsets, dataBuffer, tableOffset)
				if err != nil {
					return nil, fmt.Errorf("An Error Occured %w", err)
				}
				results = append(results, row)
			}
		}

	}

	return results, nil
}
func (engine *StorageEngine) LinearDelete(table *entities.Table, colOffsets map[string]int, expr *Expression) (int, error) { 
	// Get the Primary Key Column Name to start searching
	primaryKeyColumn, _, err := table.GetPrimaryKeyColumn()
	if err != nil {
		return 0, fmt.Errorf("An Error Occured %w", err)
	}
	rootId, ok := table.Indexes[primaryKeyColumn.Name]
	if !ok {
		return 0, fmt.Errorf("Primary key index not found for table %s", table.Name)
	}
	// Get the first leaf page of the B+ tree (linked list of leaf pages)
	leafId, err := engine.GetFirstLeafPage(rootId)
	if err != nil {
		return 0, fmt.Errorf("An Error Occured %w", err)
	}
	offset := 0
	rowsDeleted := 0
	rowsToDelete := make([]entities.RowID, 0)
	for leafId != 0 {
		//load leaf page
		buffer, err := engine.Bp.Get(leafId)
		if err != nil {
			return rowsDeleted, fmt.Errorf("An Error Occured %w", err)
		}
		//get next leaf page
		leafId = binary.BigEndian.Uint32(buffer[nextLeafPageOffset:nextLeafPageOffset + 4])
		//get number of keys
		numKeys := binary.BigEndian.Uint16(buffer[leafPageNumEntriesOffset: leafPageNumEntriesOffset + 2])
		offset = LeafPageHeaderSize
		for range numKeys {
			//skip the key bytes
			offset += int(primaryKeyColumn.Size)
			pageId := binary.BigEndian.Uint32(buffer[offset:offset + 4])
			offset += 4
			slot := binary.BigEndian.Uint16(buffer[offset:offset + 2])
			offset += 2
			// Find the Row in the table using the pageId and slot
			dataBuffer, err := engine.Bp.Get(pageId)
			if err != nil {
				return rowsDeleted,err
			}
			tableOffset, err  := pages.GetDataPageSlotOffset(dataBuffer, slot)
			if errors.Is(err, pages.ErrRowNotFound){
				continue
			}
			if err != nil {
				return rowsDeleted,fmt.Errorf("an error occured while Reading Row: %w", err)
			}
			// Get the column offsets
			// Read the null bitmap first
			nullBitmap, err := table.ReadNullBitmap(dataBuffer[tableOffset:])
			if err != nil {
				return rowsDeleted, fmt.Errorf("An error occurred while reading null bitmap: %w", err)
			}
			tableOffset += uint16(len(nullBitmap.Bitmap)) // Move the offset past the null bitmap
			// Populate the ColOffsets map with offsets for each column
			GetColumnOffsets(table,dataBuffer, tableOffset,nullBitmap, colOffsets)
			//check the condition
			conditionMet, err := engine.EvaluateExpression(dataBuffer, colOffsets, expr, table)
			if err != nil {
				return rowsDeleted, fmt.Errorf("An Error Occured %w", err)
			}
			if conditionMet {
				rowsToDelete = append(rowsToDelete, entities.RowID{PageID: pageId, Slot: slot})
			}
		}

	}

	for _, row := range rowsToDelete {
		dataBuffer, err := engine.Bp.Get(row.PageID)
		if err != nil {
			return rowsDeleted, fmt.Errorf("An Error Occured %w", err)
		}
		tableOffset, err  := pages.GetDataPageSlotOffset(dataBuffer, row.Slot)
		if errors.Is(err, pages.ErrRowNotFound){
			continue
		}
		if err != nil {
			return rowsDeleted,fmt.Errorf("an error occured while Reading Row: %w", err)
		}
		// Get the column offsets
		// Read the null bitmap first
		nullBitmap, err := table.ReadNullBitmap(dataBuffer[tableOffset:])
		if err != nil {
			return rowsDeleted, fmt.Errorf("An error occurred while reading null bitmap: %w", err)
		}
		tableOffset += uint16(len(nullBitmap.Bitmap)) // Move the offset past the null bitmap
		// Populate the ColOffsets map with offsets for each column
		GetColumnOffsets(table,dataBuffer, tableOffset,nullBitmap, colOffsets)
		err = engine.DeleteRow(table, make(map[string]int), dataBuffer, row.PageID, row.Slot)
		if errors.Is(err, pages.ErrRowNotFound) {
			continue
		}
		if err != nil {
			return rowsDeleted, fmt.Errorf("An Error Occured %w", err)
		}
		rowsDeleted++
	}

	return rowsDeleted, nil

}

func (engine *StorageEngine) LinearUpdate(table *entities.Table, colOffsets map[string]int, expr *Expression, updates map[string]any) (int, error) {
	// Get the Primary Key Column Name to start searching
	primaryKeyColumn, _, err := table.GetPrimaryKeyColumn()
	if err != nil {
		return 0, fmt.Errorf("An Error Occured %w", err)
	}
	rootId, ok := table.Indexes[primaryKeyColumn.Name]
	if !ok {
		return 0, fmt.Errorf("Primary key index not found for table %s", table.Name)
	}
	// Get the first leaf page of the B+ tree (linked list of leaf pages)
	leafId, err := engine.GetFirstLeafPage(rootId)
	if err != nil {
		return 0, fmt.Errorf("An Error Occured %w", err)
	}
	offset := 0
	rowsUpdated := 0
	rowsToUpdate := make([]entities.RowID, 0)
	for leafId != 0 {
		//load leaf page
		buffer, err := engine.Bp.Get(leafId)
		if err != nil {
			return rowsUpdated, fmt.Errorf("An Error Occured %w", err)
		}
		//get next leaf page
		leafId = binary.BigEndian.Uint32(buffer[nextLeafPageOffset:nextLeafPageOffset + 4])
		//get number of keys
		numKeys := binary.BigEndian.Uint16(buffer[leafPageNumEntriesOffset: leafPageNumEntriesOffset + 2])
		offset = LeafPageHeaderSize
		for range numKeys {
			//skip the key bytes
			offset += int(primaryKeyColumn.Size)
			pageId := binary.BigEndian.Uint32(buffer[offset:offset + 4])
			offset += 4
			slot := binary.BigEndian.Uint16(buffer[offset:offset + 2])
			offset += 2
			// Find the Row in the table using the pageId and slot
			dataBuffer, err := engine.Bp.Get(pageId)
			if err != nil {
				return rowsUpdated,err
			}
			tableOffset, err  := pages.GetDataPageSlotOffset(dataBuffer, slot)
			if errors.Is(err, pages.ErrRowNotFound){
				continue
			}
			if err != nil {
				return rowsUpdated,fmt.Errorf("an error occured while Reading Row: %w", err)
			}
			// Get the column offsets
			// Read the null bitmap first
			nullBitmap, err := table.ReadNullBitmap(dataBuffer[tableOffset:])
			if err != nil {
				return rowsUpdated, fmt.Errorf("An error occurred while reading null bitmap: %w", err)
			}
			tableOffset += uint16(len(nullBitmap.Bitmap)) // Move the offset past the null bitmap
			// Populate the ColOffsets map with offsets for each column
			GetColumnOffsets(table,dataBuffer, tableOffset,nullBitmap, colOffsets)
			//check the condition
			conditionMet, err := engine.EvaluateExpression(dataBuffer, colOffsets, expr, table)
			if err != nil {
				return rowsUpdated, fmt.Errorf("An Error Occured %w", err)
			}
			if conditionMet {
				rowsToUpdate = append(rowsToUpdate, entities.RowID{PageID: pageId, Slot: slot})
			}
		}

	}

	for _, row := range rowsToUpdate {
		dataBuffer, err := engine.Bp.Get(row.PageID)
		if err != nil {
			return rowsUpdated, fmt.Errorf("An Error Occured %w", err)
		}
		tableOffset, err  := pages.GetDataPageSlotOffset(dataBuffer, row.Slot)
		if errors.Is(err, pages.ErrRowNotFound){
			continue
		}
		if err != nil {
			return rowsUpdated,fmt.Errorf("an error occured while Reading Row: %w", err)
		}
		// Get the column offsets
		// Read the null bitmap first
		nullBitmap, err := table.ReadNullBitmap(dataBuffer[tableOffset:])
		if err != nil {
			return rowsUpdated, fmt.Errorf("An error occurred while reading null bitmap: %w", err)
		}
		tableOffset += uint16(len(nullBitmap.Bitmap)) // Move the offset past the null bitmap
		// Populate the ColOffsets map with offsets for each column
		GetColumnOffsets(table,dataBuffer, tableOffset,nullBitmap, colOffsets)
		err = engine.UpdateRow(table, colOffsets, dataBuffer, row.PageID, row.Slot, updates)
		if errors.Is(err, pages.ErrRowNotFound) {
			continue
		}
		if err != nil {
			return rowsUpdated, fmt.Errorf("An Error Occured %w", err)
		}
		rowsUpdated++
	}

	return rowsUpdated, nil


}


func SerializeSearchExpression(ex *Expression, table *entities.Table) ( error) {

	if ex.Type == NodeCondition {
		col, err := table.GetColumnByName(ex.Condition.ColumnName)
		if err != nil {
			return fmt.Errorf("An Error Occured %w", err)
		}
		val , err := col.GetDefaultValue(string(ex.Condition.Value))
		if err != nil {
			return fmt.Errorf("An Error Occured %w", err)
		}
		value, err := entities.Serialize(val, col)
		if err != nil {
			return fmt.Errorf("An Error Occured %w", err)
		}
		ex.Condition.Value = value
	}
	if ex.Left != nil {
		err := SerializeSearchExpression(ex.Left, table)
		if err != nil {
			return fmt.Errorf("An Error Occured %w", err)
		}
	}
	if ex.Right != nil {
		err := SerializeSearchExpression(ex.Right, table)
		if err != nil {
			return fmt.Errorf("An Error Occured %w", err)
		}
	}
	return nil
}