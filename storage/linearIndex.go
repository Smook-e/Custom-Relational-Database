package storage
import (
	"encoding/binary"
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
func (engine *StorageEngine) EvaluateExpression(buffer []byte, expr *Expression, table *entities.Table) (bool, error) {
	if expr == nil {
		return true, nil
	}
	switch expr.Type {
	case NodeCondition:
		return VerifyCondition(buffer, expr.Condition, table)
	case NodeAnd:
		leftResult, _ := engine.EvaluateExpression(buffer, expr.Left, table)
		if leftResult == false {
			return false, nil
		}
		return engine.EvaluateExpression(buffer, expr.Right, table)
	case NodeOr:
		leftResult, _ := engine.EvaluateExpression(buffer, expr.Left, table)
		if leftResult == true {
			return true, nil
		}
		return engine.EvaluateExpression(buffer, expr.Right, table)	
	default:
		return false, fmt.Errorf("Unknown expression type: %v", expr.Type)
	}
}
func VerifyCondition(buffer []byte, condition *SearchCondition, table *entities.Table) (bool, error) {
	if condition == nil {
		return true, nil
	}
	offset := 0
	// Read the null bitmap first
	nullBitmap, err := table.ReadNullBitmap(buffer)
	if err != nil {
		return false, fmt.Errorf("An error occurred while reading null bitmap: %w", err)
	}
	offset += len(nullBitmap.Bitmap)
	for i, col := range table.Columns {
		size, err := entities.GetSize(&col)
		if err != nil {
			return false, fmt.Errorf("An error occurred while getting size of column %s: %w", col.Name, err)
		}
		if size == 0 {// For variable-length types
			size = buffer[offset] + 1 // First byte indicates the length of the variable-length data
		}
		
		if col.Name == condition.ColumnName {
			// Check if the column is null using the null bitmap
			if nullBitmap.IsNull(i) {
				return false, nil
			}
			// Compare the value in the buffer with the condition value
			comp, err := entities.Compare(buffer[offset:offset+int(size)], condition.Value, &col)
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
		if !nullBitmap.IsNull(i) {
			offset += int(size)
		}
	}
	return false, fmt.Errorf("Condition checking not implemented yet")
}



func (engine *StorageEngine) LinearSearch(tableName string, expr *Expression) ([][]any, error) {
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
			if err != nil {
				return nil,fmt.Errorf("an error occured while Reading Row: %w", err)
			}
			//check the condition
			conditionMet, err := engine.EvaluateExpression(dataBuffer[tableOffset:], expr, table)
			if err != nil {
				return nil, fmt.Errorf("An Error Occured %w", err)
			}
			if conditionMet {
				// Read the pageId and slot from the buffer
				row, err := engine.ReadRow(tableName, dataBuffer, tableOffset)
				if err != nil {
					return nil, fmt.Errorf("An Error Occured %w", err)
				}
				results = append(results, row)
			}
		}

	}

	return results, nil
}


func (engine *StorageEngine) Search(tableName string, expr *Expression) ([][]any, error) {
	// Get the table object from the database
	table, ok := engine.db.Tables[tableName]
	if !ok {
		return nil, fmt.Errorf("Table %s not found", tableName)
	}
	
	// If there's only one condition, check if it has an index and use the indexed search
	if expr != nil && expr.Type == NodeCondition {
		condition := expr.Condition
		if condition != nil {
			rootID, exists := table.Indexes[condition.ColumnName]// Check if the column has an index
			if exists {
				// If the column has an index, use the indexed search
				column, _ := table.GetColumnByName(condition.ColumnName)
				pageId, slot, _ := engine.IndexSearch(rootID, condition.Value, column)
				
				if pageId == 0 && slot == 0 {
					return [][]any{}, nil
				}else {
					// If the key is found, read the row
					dataBuffer, err := engine.Bp.Get(pageId)
					if err != nil {
						return nil, fmt.Errorf("An Error Occured %w", err)
					}
					tableOffset, err  := pages.GetDataPageSlotOffset(dataBuffer, slot)
					if err != nil {
						return nil,fmt.Errorf("an error occured while Reading Row: %w", err)
					}
					row, err := engine.ReadRow(tableName, dataBuffer, tableOffset)
					if err != nil {
						return nil, fmt.Errorf("An Error Occured %w", err)
					}
					return [][]any{row}, nil
				}

			}
			
		}
	}
	// If the condition is more complex or the column doesn't have an index, use linear search
	results, err := engine.LinearSearch(tableName, expr)
	if err != nil {
		return nil, fmt.Errorf("An Error Occured %w", err)
	}

	return results, nil
}