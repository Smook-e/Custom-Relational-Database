package entities

import (
	"errors"
	"fmt"
	"strings"
	"strconv"
	"regexp"
	"encoding/binary"

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
	ConstraintSerial     uint8 = 1 << 4 // 16
	ConstraintDefault    uint8 = 1 << 5 // 32
)

type Table struct {
	Name string
	RootIndex uint32
	Columns []Column
	Indexes map[string]uint32 // Map of column name to index page ID
	ForeignKeys map[string]ForeignKeyReference // Map of column name to foreign key reference
}
type Row struct {
	PageID uint32
	Slot uint16
}
type Column struct {
	Name string
	DataType uint8
	Constraints uint8
	Size uint8
	Default any
}
type ColumnDefinition struct {
	Name string
	DataType string
	Constraints []string
	Default string
}
type ForeignKeyReference struct {
	ReferencedTableName string
	ReferencedColumnIndex uint8
}
func (col *Column) HasConstraint(constraint uint8) bool {
	return col.Constraints&constraint != 0
}

func (t *Table) GetValues(vals []string) ([]any,uint16 ,  error) {
	values := make([]any, len(vals))
	var col *Column
	var size uint16 = 0
	for i, val := range vals {
		col = &t.Columns[i]
		switch col.DataType {
		case TypeTinyInt:
			n, err := strconv.ParseInt(val, 10, 8)
			if err != nil {
				return nil, 0, fmt.Errorf("Error converting %s to TinyInt: %w", val, err)
			}
			size += uint16(col.Size)
			values[i] = int8(n)
		case TypeSmallInt:
			n, err := strconv.ParseInt(val, 10, 16)
			if err != nil {
				return nil, 0, fmt.Errorf("Error converting %s to SmallInt: %w", val, err)
			}
			size += uint16(col.Size)
			values[i] = int16(n)
		case TypeInt:
			n, err := strconv.ParseInt(val, 10, 32)
			if err != nil {
				return nil, 0, fmt.Errorf("Error converting %s to Int: %w", val, err)
			}
			size += uint16(col.Size)
			values[i] = int32(n)
		case TypeBigInt:
			n, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				return nil, 0, fmt.Errorf("Error converting %s to BigInt: %w", val, err)
			}
			size += uint16(col.Size)
			values[i] = int64(n)
		case TypeVarChar:
			if col.Size > 0 {
				if len(val) > int(col.Size) {
					return nil, 0, fmt.Errorf("Error: Value %s exceeds maximum length of %d for column %s", val, col.Size, col.Name)
				}
				size += uint16(col.Size) + 1 //string length + 1 byte for length prefix
			}else {
				size += uint16(len(val)) + 1 //string length + 1 byte for length prefix
			}
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
var varcharRegex = regexp.MustCompile(`(?i)^varchar(?:[(](\d+)[)])?$`)
func GetDataTypeAndSize(datatype string) (uint8, uint8, error) {
	switch  {
		case datatype == "tinyint":
			return TypeTinyInt, 1, nil
		case datatype == "smallint":
			return TypeSmallInt, 2, nil
		case datatype == "bigint":
			return TypeBigInt, 8, nil
		case datatype == "int":
			return TypeInt, 4, nil
		case varcharRegex.MatchString(datatype):
			matches := varcharRegex.FindStringSubmatch(datatype)
			 // Index 1 holds the digits if they were provided
			if matches[1] != "" {
				// Parse into uint8 
				length, err := strconv.ParseUint(matches[1], 10, 8)
				if err != nil {
					return 0, 0, fmt.Errorf("Invalid VARCHAR length: %s", matches[1])
				}
				return TypeVarChar, uint8(length), nil
			}
			return TypeVarChar, 0, nil
		default:
			return 0, 0, fmt.Errorf("Data type %s not supported", datatype)
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
			case "serial":
				result |= ConstraintSerial
			case "default":
				result |= ConstraintDefault
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




func (db *Database) Serialize(val any, col *Column) ([]byte, error) {

	switch col.DataType {
	case TypeTinyInt:
		v, ok := val.(int8)
		if !ok {
			return nil, fmt.Errorf("Expected int8 for TypeTinyInt, got %T", val)
		}
		return []byte{byte(v)}, nil
	case TypeSmallInt:
		buf := make([]byte, 2)
		v, ok := val.(int16)
		if !ok {
			return nil, fmt.Errorf("Expected int16 for TypeSmallInt, got %T", val)
		}
		binary.BigEndian.PutUint16(buf, uint16(v))
		return buf, nil
	case TypeInt:
		buf := make([]byte, 4)
		v, ok := val.(int32)
		if !ok {
			return nil, fmt.Errorf("Expected int32 for TypeInt, got %T", val)
		}
		binary.BigEndian.PutUint32(buf, uint32(v))
		return buf, nil
	case TypeBigInt:
		buf := make([]byte, 8)
		v, ok := val.(int64)
		if !ok {
			return nil, fmt.Errorf("Expected int64 for TypeBigInt, got %T", val)
		}
		binary.BigEndian.PutUint64(buf, uint64(v))
		return buf, nil
	case TypeVarChar:
		strVal, ok := val.(string)
		if !ok {
			return nil, fmt.Errorf("Expected string for TypeVarChar, got %T", val)
		}
		if len(strVal) > 255 {
			return nil, fmt.Errorf("String length exceeds maximum of 255 characters")
		}
		var buf []byte
		if col.Size > 0 {// fixed size varchar
			if len(strVal) > int(col.Size) {
				return nil, fmt.Errorf("String length exceeds maximum of %d characters", col.Size)
			}
			buf = make([]byte, col.Size+1)
		}else{
			buf = make([]byte, len(strVal)+1)
		}
		buf[0] = byte(len(strVal))
		copy(buf[1:], strVal)
		return buf, nil
	default:
		return nil, fmt.Errorf("Unsupported data type for serialization")
	}
}
func Deserialize(data []byte, dataType uint8) (any, error) {
	switch dataType {
	case TypeTinyInt:
		return int8(data[0]), nil
	case TypeSmallInt:
		return int16(binary.BigEndian.Uint16(data)), nil
	case TypeInt:
		return int32(binary.BigEndian.Uint32(data)), nil
	case TypeBigInt:
		return int64(binary.BigEndian.Uint64(data)), nil
	case TypeVarChar:
		length := int(data[0])
		if length+1 > len(data) {
			return nil, fmt.Errorf("Data length mismatch for VarChar")
		}
		return string(data[1 : 1+length]), nil
	default:
		return nil, fmt.Errorf("Unsupported data type for deserialization")
	}
}
func Compare(val1, val2 []byte, col *Column) (int, error) {
	switch col.DataType {
	case TypeTinyInt:
		v1 := int8(val1[0])
		v2 := int8(val2[0])
		if v1 < v2 {
			return -1, nil
		} else if v1 > v2 {
			return 1, nil
		}
		return 0, nil
	case TypeSmallInt:
		v1 := int16(binary.BigEndian.Uint16(val1))
		v2 := int16(binary.BigEndian.Uint16(val2))
		if v1 < v2 {
			return -1, nil
		} else if v1 > v2 {
			return 1, nil
		}
		return 0, nil
	case TypeInt:
		v1 := int32(binary.BigEndian.Uint32(val1))
		v2 := int32(binary.BigEndian.Uint32(val2))
		if v1 < v2 {
			return -1, nil
		} else if v1 > v2 {
			return 1, nil
		}
		return 0, nil
	case TypeBigInt:
		v1 := int64(binary.BigEndian.Uint64(val1))
		v2 := int64(binary.BigEndian.Uint64(val2))
		if v1 < v2 {
			return -1, nil
		} else if v1 > v2 {
			return 1, nil
		}
		return 0, nil
	case TypeVarChar:
		if col.Size > 0 {
			val1Size := int(val1[0])
			val2Size := int(val2[0])
			return strings.Compare(string(val1[1:1+val1Size]), string(val2[1:1+val2Size])), nil
		}
		return strings.Compare(string(val1[1:]), string(val2[1:])), nil
	default:
		return 0, fmt.Errorf("Unsupported data type for comparison")
	}
}
func (db *Database) PrintConstraints(constraints uint8) string {
	var result string
	if constraints&ConstraintPrimaryKey != 0 {
		result += "PRIMARY KEY, "
	}
	if constraints&ConstraintUnique != 0 {
		result += "UNIQUE, "
	}
	if constraints&ConstraintNotNull != 0 {
		result += "NOT NULL, "
	}
	if constraints&ConstraintIndex != 0 {
		result += "INDEX, "
	}
	if constraints&ConstraintSerial != 0 {
		result += "SERIAL, "
	}
	return result
}
func (db *Database) PrintDataType(dataType uint8) string {
	switch dataType {
	case TypeTinyInt:
		return "TINYINT"
	case TypeSmallInt:
		return "SMALLINT"
	case TypeInt:
		return "INT"
	case TypeBigInt:
		return "BIGINT"
	case TypeVarChar:
		return "VARCHAR"
	default:
		return "UNKNOWN"
	}
}

func (table *Table) GetPrimaryKeyColumn() (*Column, int, error) {
	for i, col := range table.Columns {
		if col.Constraints&ConstraintPrimaryKey != 0 {
			return &col, i, nil
		}
	}
	return nil, -1, fmt.Errorf("No primary key column found in table %s", table.Name)
}
// SetDefaultValue reads the buffer and sets the default value for the column based on its data type. it returns the number of bytes read from the buffer and an error
func (col *Column) SetDefaultValue(buffer []byte, offset int) (int,error) {
	switch col.DataType {
	case TypeTinyInt:
		col.Default = int8(buffer[offset])
		return 1, nil
	case TypeSmallInt:
		col.Default = int16(binary.BigEndian.Uint16(buffer[offset : offset+2]))
		return 2, nil
	case TypeInt:
		col.Default = int32(binary.BigEndian.Uint32(buffer[offset : offset+4]))
		return 4, nil
	case TypeBigInt:
		col.Default = int64(binary.BigEndian.Uint64(buffer[offset : offset+8]))
		return 8, nil
	case TypeVarChar:
		length := int(buffer[offset])
		col.Default = string(buffer[offset+1 : offset+1+length])
		if col.Size > 0 {// fixed size varchar
			return int(col.Size) + 1, nil // Return the size of the column + 1 for the length byte
		}
		return length + 1, nil // Return the length of the string + 1 for the length byte
	default:
		return 0, fmt.Errorf("Unsupported data type for setting default value")
	}
}
func (col *Column) GetDefaultValue(value string) (any, error) {
	switch col.DataType {
	case TypeTinyInt:
		n, err := strconv.ParseInt(value, 10, 8)
		if err != nil {
			return nil, fmt.Errorf("Error converting %s to TinyInt: %w", value, err)
		}
		return int8(n), nil
	case TypeSmallInt:
		n, err := strconv.ParseInt(value, 10, 16)
		if err != nil {
			return nil, fmt.Errorf("Error converting %s to SmallInt: %w", value, err)
		}
		return int16(n), nil
	case TypeInt:
		n, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("Error converting %s to Int: %w", value, err)
		}
		return int32(n), nil
	case TypeBigInt:
		n, err := strconv.ParseInt(value,	 10, 64)
		if err != nil {
			return nil, fmt.Errorf("Error converting %s to BigInt: %w", value, err)
		}
		return int64(n), nil
	case TypeVarChar:
		if len(value) > 255 {
			return nil, fmt.Errorf("String length exceeds maximum of 255 characters")
		}
		if col.Size > 0 && len(value) > int(col.Size) {
			return nil, fmt.Errorf("String length exceeds maximum of %d characters", col.Size)
		}
		return value, nil
	default:
		return nil, fmt.Errorf("Unsupported data type for default value")
	}
}