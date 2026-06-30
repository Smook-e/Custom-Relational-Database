package entities

import (
	"errors"
	"fmt"
	"strings"
	"strconv"

)

// Data type identifiers
const (
	TypeTinyInt uint8 = 0 //1 byte
	TypeSmallInt uint8 = 1 // 2 bytes
	TypeInt     uint8 = 2 // 4 bytes
	TypeBigInt  uint8 = 3 // 8 bytes
	TypeVarChar uint8 = 4 
)

// Constraint identifiers 
const (
	ConstraintNone       uint8 = 0
	ConstraintPrimaryKey uint8 = 1 << 0 // 1
	ConstraintNotNull    uint8 = 1 << 1 // 2
	ConstraintUnique     uint8 = 1 << 2 // 4
	ConstraintIndex      uint8 = 1 << 3 // 8
)

type Table struct {
	Name string
	RootIndex uint32
	Columns []Column
}

type Column struct {
	Name string
	DataType uint8
	Constraints uint8
	Size uint8
}
type ColumnDefinition struct {
	Name string
	DataType string
	Constraints []string
}

func (t *Table) GetValues(vals []string) ([]any,uint16 ,  error) {
	values := make([]any, len(vals))
	var col *Column
	var size uint16 = 0
	for i, val := range vals {
		col = &t.Columns[i]
		switch col.DataType {
		case TypeTinyInt:
			n, err := strconv.Atoi(val)
			if err != nil {
				return nil, 0, fmt.Errorf("Error converting %s to TinyInt", val)
			}
			size += uint16(col.Size)
			values[i] = int8(n)
		case TypeSmallInt:
			n, err := strconv.Atoi(val)
			if err != nil {
				return nil, 0, fmt.Errorf("Error converting %s to SmallInt", val)
			}
			size += uint16(col.Size)
			values[i] = int16(n)
		case TypeInt:
			n, err := strconv.Atoi(val)
			if err != nil {
				return nil, 0, fmt.Errorf("Error converting %s to Int", val)
			}
			size += uint16(col.Size)
			values[i] = int32(n)
		case TypeBigInt:
			n, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				return nil, 0, fmt.Errorf("Error converting %s to BigInt", val)
			}
			size += uint16(col.Size)
			values[i] = int64(n)
		case TypeVarChar:
			size += uint16(len(val)) + 1 //string length + 1 byte for length prefix
			values[i] = val
		default:
			return nil, 0, fmt.Errorf("Unsupported data type for column %s", col.Name)
		}
	}
	return values, size, nil

}

func GetSize(Type uint8) (uint8, error) {
	switch Type {
	case TypeTinyInt:
		return 1, nil
	case TypeSmallInt:
		return 2, nil
	case TypeInt:
		return 4, nil
	case TypeBigInt:
		return 8, nil
	case TypeVarChar:
		return 0, nil
	default:
		return 0, errors.New("unknown DataType")
	}

}

func GetDataType(datatype string) (uint8, error) {
	switch datatype {
		case "tinyint":
			return TypeTinyInt, nil
		case "smallint":
			return TypeSmallInt, nil
		case "bigint":
			return TypeBigInt, nil
		case "int":
			return TypeInt, nil
		case "varchar":
			return TypeVarChar, nil
		default:
			return 0, fmt.Errorf("Data type %s not supported", datatype)
		}
}
func GetConstraint(Constraints []string) (uint8, error) {
	result := uint8(0)
	for _, constraint := range Constraints {
			switch strings.ToLower(constraint) {
			case "primarykey":
				result |= ConstraintPrimaryKey
			case "notnull":
				result |= ConstraintNotNull
			case "unique":
				result |= ConstraintUnique
			case "index":
				result |= ConstraintIndex
			default:
				return 0, fmt.Errorf("Constraint %s not supported", constraint)
			}
	}
	return result, nil
}

func (t *Table) GetColumnByName(name string) (*Column, error) {
	for _, col := range t.Columns {
		if col.Name == name {
			return &col, nil
		}
	}
	return nil, fmt.Errorf("Column %s not found in table %s", name, t.Name)
}

func (t *Table) GetColumnIndexByName(name string) (int, error) {
	for i, col := range t.Columns {
		if col.Name == name {
			return i, nil
		}
	}
	return -1, fmt.Errorf("Column %s not found in table %s", name, t.Name)
}

func (t *Table) GetColumnByIndex(index int) (*Column, error) {
	if index < 0 || index >= len(t.Columns) {
		return nil, fmt.Errorf("Index %d out of bounds for table %s", index, t.Name)
	}
	return &t.Columns[index], nil
}