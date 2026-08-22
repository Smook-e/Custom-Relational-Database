package storage
import (
	"encoding/binary"
	"fmt"
	"github.com/Smook-e/Custom-Relational-Database/entities"
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
func (engine *StorageEngine) VerifyCondition(buffer []byte, condition *SearchCondition, table *entities.Table) (bool, error) {
	if condition == nil {
		return true, nil
	}
	return false, fmt.Errorf("Condition checking not implemented yet")
}



func (engine *StorageEngine) LinearSearch(tableName string, condition *SearchCondition) ([][]any, error) {
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
			//check the condition
			if condition == nil {
				// read all rows if no condition is provided
				pageId := binary.BigEndian.Uint32(buffer[offset:offset + 4])
				offset += 4
				slot := binary.BigEndian.Uint16(buffer[offset:offset + 2])
				offset += 2
				row, err := engine.ReadRow(tableName, pageId, slot)
				if err != nil {
					return nil, fmt.Errorf("An Error Occured %w", err)
				}
				results = append(results, row)
			}
		}

	}

	return results, nil
}