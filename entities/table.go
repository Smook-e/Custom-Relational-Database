package entities

import (
	"errors"
	"fmt"
	"strings"
	"strconv"
	"regexp"
	"encoding/binary"

)
/*
This file defines the core data structures and constants for the database system,
including tables, rows, columns, and their associated properties.
It also provides utility functions Like serialization, deserialization, comparison, and constraint handling.
*/

// Data type identifiers
const (
	TypeTinyInt uint8 = iota //1 byte
	TypeSmallInt   			 // 2 bytes
	TypeInt      			 // 4 bytes
	TypeBigInt				 // 8 bytes
	TypeVarChar 			 // Variable length, up to 255 bytes
	TypeSerial				 // 4 bytes, auto incrementing integer
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
// ColumnDefinition is used to define a column when creating a table. Every field is a string since it is provided by the user.
type ColumnDefinition struct {
	Name string
	DataType string
	Constraints []string
	Default string
}
// ForeignKeyReference represents a reference to another table's column, used for foreign key constraints.
type ForeignKeyReference struct {
	ReferencedTableName string
	ReferencedColumnIndex uint8
}
// ForeignKeyDefinition is used to define a foreign key when creating a table. Every field is a string since it is provided by the user.
type ForeignKeyDefinition struct {
	ColumnName string
	ReferencedTableName string
	ReferencedColumnName string
}
type NullBitmap struct {
	Bitmap []byte
}
func (nb *NullBitmap) IsNull(index int) bool {
	byteIndex := index / 8
	bitIndex := index % 8
	return (nb.Bitmap[byteIndex] & (1 << bitIndex)) != 0
}
func (nb *NullBitmap) SetNull(index int) {
	byteIndex := index / 8
	bitIndex := index % 8
	nb.Bitmap[byteIndex] |= (1 << bitIndex)
}
func (nb *NullBitmap) ClearNull(index int) {
	byteIndex := index / 8
	bitIndex := index % 8
	nb.Bitmap[byteIndex] &^= (1 << bitIndex)
}
func (table *Table) InitializeNullBitmap() *NullBitmap {
	nullBitmapSize := (len(table.Columns) + 7) / 8 // Calculate the size of the null bitmap in bytes
	return &NullBitmap{
		Bitmap: make([]byte, nullBitmapSize),
	}
}
func (table *Table) WriteNullBitmap(nullBitmap *NullBitmap, buffer []byte) ( error) {
	nullBitmapSize := (len(table.Columns) + 7) / 8 // Calculate the size of the null bitmap in bytes
	if len(buffer) < nullBitmapSize {
		return errors.New("Buffer too small to write null bitmap")
	}
	copy(buffer[:nullBitmapSize], nullBitmap.Bitmap)
	return nil
}

func (table *Table) ReadNullBitmap(buffer []byte) (*NullBitmap, error) { 
	nullBitmapSize := (len(table.Columns) + 7) / 8 // Calculate the size of the null bitmap in bytes
	if len(buffer) < nullBitmapSize {
		return nil, errors.New("Buffer too small to read null bitmap")
	}
	nullBitmap := &NullBitmap{
		Bitmap: make([]byte, nullBitmapSize),
	}
	copy(nullBitmap.Bitmap, buffer[:nullBitmapSize])
	return nullBitmap, nil
}
// returns True if the column has the specified constraint, otherwise returns False
func (col *Column) HasConstraint(constraint uint8) bool {
	return col.Constraints&constraint != 0
}


// GetValues takes a slice of string values and converts them to their respective types based on the column definitions of the table. 
// It returns a slice containing the converted values, the total size in bytes of the serialized data, a null bitmap indicating which columns are null.
// it returns an error if any value cannot be converted to the appropriate type.
func (t *Table) GetValues(vals []string) ([]any,uint16, *NullBitmap, error) {
	values := make([]any, len(vals))
	var col *Column
	var size uint16 = 0
	nullBitmap := t.InitializeNullBitmap()
	size += uint16(len(nullBitmap.Bitmap))
	for i, val := range vals {
		if val == "" {
			values[i] = nil
			nullBitmap.SetNull(i)
			continue
		}
		col = &t.Columns[i]
		switch col.DataType {
		case TypeTinyInt:
			n, err := strconv.ParseInt(val, 10, 8)
			if err != nil {
				return nil, 0,nil, fmt.Errorf("Error converting %s to TinyInt: %w", val, err)
			}
			size += uint16(col.Size)
			values[i] = int8(n)
		case TypeSmallInt:
			n, err := strconv.ParseInt(val, 10, 16)
			if err != nil {
				return nil, 0,nil, fmt.Errorf("Error converting %s to SmallInt: %w", val, err)
			}
			size += uint16(col.Size)
			values[i] = int16(n)
		case TypeInt:
			n, err := strconv.ParseInt(val, 10, 32)
			if err != nil {
				return nil, 0,nil, fmt.Errorf("Error converting %s to Int: %w", val, err)
			}
			size += uint16(col.Size)
			values[i] = int32(n)
		case TypeSerial:// same as TypeInt
			return nil, 0,nil, errors.New("Error: Serial type should not be provided by user. It is auto-incremented.")
		case TypeBigInt:
			n, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				return nil, 0,nil, fmt.Errorf("Error converting %s to BigInt: %w", val, err)
			}
			size += uint16(col.Size)
			values[i] = int64(n)
		case TypeVarChar:
			if col.Size > 0 {
				if len(val) > int(col.Size) {
					return nil, 0, nil, fmt.Errorf("Error: Value %s exceeds maximum length of %d for column %s", val, col.Size, col.Name)
				}
				size += uint16(col.Size) + 1 //string length + 1 byte for length prefix
			}else {
				size += uint16(len(val)) + 1 //string length + 1 byte for length prefix
			}
			values[i] = val
		default:
			return nil, 0, nil, fmt.Errorf("Unsupported data type for column %s", col.Name)
		}
	}
	return values, size, nullBitmap, nil

}
func (t *Table) GetSizeOfValues(values []any) (uint16, error) {
	var size uint16 = 0
	size += uint16((len(t.Columns) + 7) / 8) // Add size of null bitmap
	for i, val := range values {
		col := &t.Columns[i]
		if val == nil {
			continue // Null values do not contribute to size
		}
		switch col.DataType {
		case TypeTinyInt:
			size += 1
		case TypeSmallInt:
			size += 2
		case TypeInt, TypeSerial:
			size += 4
		case TypeBigInt:
			size += 8
		case TypeVarChar:
			strVal, ok := val.(string)
			if !ok {
				return 0, fmt.Errorf("Expected string for TypeVarChar, got %T", val)
			}
			if col.Size > 0 && len(strVal) > int(col.Size) {
				return 0, fmt.Errorf("String length exceeds maximum of %d characters for column %s", col.Size, col.Name)
			}
			if col.Size > 0 {
				size += uint16(col.Size) + 1 // +1 for length prefix
			} else {
				size += uint16(len(strVal)) + 1 // +1 for length prefix
			}
		default:
			return 0, fmt.Errorf("Unsupported data type for column %s", col.Name)
		}
	}
	return size, nil
}
func (t *Table) GetNullBitmapForValues(values []any) *NullBitmap {
	nullBitmap := t.InitializeNullBitmap()
	for i, val := range values {
		if val == nil {
			nullBitmap.SetNull(i)
		}
	}
	return nullBitmap
}

func GetValueFromString(val string, col *Column) (any, error) {
	if val == "" {
		return nil, nil
	}
	switch col.DataType {
	case TypeTinyInt:
		n, err := strconv.ParseInt(val, 10, 8)
		if err != nil {
			return nil, fmt.Errorf("Error converting %s to TinyInt: %w", val, err)
		}
		return int8(n), nil
	case TypeSmallInt:
		n, err := strconv.ParseInt(val, 10, 16)
		if err != nil {
			return nil, fmt.Errorf("Error converting %s to SmallInt: %w", val, err)
		}
		return int16(n), nil
	case TypeInt:
		n, err := strconv.ParseInt(val, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("Error converting %s to Int: %w", val, err)
		}
		return int32(n), nil
	case TypeSerial:// same as TypeInt
		return nil, errors.New("Error: Serial type should not be provided by user. It is auto-incremented.")
	case TypeBigInt:
		n, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("Error converting %s to BigInt: %w", val, err)
		}
		return int64(n), nil
	case TypeVarChar:
		if col.Size > 0 {
			if len(val) > int(col.Size) {
				return nil, fmt.Errorf("Error: Value %s exceeds maximum length of %d for column %s", val, col.Size, col.Name)
			}
		}
		return val, nil
	default:
		return nil, fmt.Errorf("Unsupported data type for column %s", col.Name)
	}
}

// Returns the size in bytes of the serialized value for a given column.
func GetSize(col *Column) (uint8, error) {
	switch col.DataType {
	case TypeTinyInt:
		return 1, nil
	case TypeSmallInt:
		return 2, nil
	case TypeInt:
		return 4, nil
	case TypeBigInt:
		return 8, nil
	case TypeVarChar:
		if col.Size > 0 {
			return col.Size + 1, nil // +1 for length prefix
		}
		return 0, nil
	case TypeSerial:
		return 4, nil // Serial is stored as a 4-byte integer
	default:
		return 0, errors.New("unknown DataType")
	}

}

var varcharRegex = regexp.MustCompile(`(?i)^varchar(?:[(](\d+)[)])?$`)
// Returns the data type and size for a given string representation of a data type. For VARCHAR, it extracts the length if specified.
// Typically used when creating a table to convert user-provided data type strings into internal representations.
func GetDataTypeAndSize(datatype string) (uint8, uint8, error) {
	datatype = strings.ToLower(datatype)
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
		case datatype == "serial":
			return TypeSerial, 4, nil
		default:
			return 0, 0, fmt.Errorf("Data type %s not supported", datatype)
		}
}
// Returns the Constraint bitmask for a given slice of constraint strings. It converts user-provided constraint strings into internal representations.
// Typically used when creating a table to convert user-provided constraint strings into internal representations.
func GetConstraint(Constraints []string) (uint8, error) {
	result := uint8(0)
	for _, constraint := range Constraints {
			switch strings.ToLower(constraint) {
			case "primarykey":
				result |= ConstraintPrimaryKey
				result |= ConstraintNotNull // Primary key implies NOT NULL
				result |= ConstraintUnique // Primary key implies UNIQUE
				result |= ConstraintIndex // Primary key implies an index
			case "notnull":
				result |= ConstraintNotNull
			case "unique":
				result |= ConstraintUnique
				result |= ConstraintIndex // Unique implies an index
			case "index":
				result |= ConstraintIndex
				result |= ConstraintUnique // Index implies unique
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

// returns a pointer to the Column with the specified name in the Table.
func (t *Table) GetColumnByName(name string) (*Column, error) {
	for _, col := range t.Columns {
		if col.Name == name {
			return &col, nil
		}
	}
	return nil, fmt.Errorf("Column %s not found in table %s", name, t.Name)
}

// returns the index of the Column with the specified name in the Table.
func (t *Table) GetColumnIndexByName(name string) (int, error) {
	for i, col := range t.Columns {
		if col.Name == name {
			return i, nil
		}
	}
	return -1, fmt.Errorf("Column %s not found in table %s", name, t.Name)
}
// returns a pointer to the Column at the specified index in the Table.
func (t *Table) GetColumnByIndex(index int) (*Column, error) {
	if index < 0 || index >= len(t.Columns) {
		return nil, fmt.Errorf("Index %d out of bounds for table %s", index, t.Name)
	}
	return &t.Columns[index], nil
}

func (col *Column) IsValidValue(val any) bool {
	if val == nil {
		return (col.HasConstraint(ConstraintNotNull) == true && col.HasConstraint(ConstraintDefault) == false) // If the column is NOT NULL and has no default, nil is invalid
	}
	switch col.DataType {
	case TypeTinyInt:
		_, ok := val.(int8)
		return ok
	case TypeSmallInt:
		_, ok := val.(int16)
		return ok
	case TypeInt:
		_, ok := val.(int32)
		return ok
	case TypeSerial:// same as TypeInt
		_, ok := val.(int32)
		return ok
	case TypeBigInt:
		_, ok := val.(int64)
		return ok
	case TypeVarChar:
		strVal, ok := val.(string)
		if !ok {
			return false
		}
		if col.Size > 0 && len(strVal) > int(col.Size) {
			return false
		}
		return true
	default:
		return false
	}
}

// Serialize takes a value and a column definition, and returns the serialized byte representation of the value based on the column's data type.
// This is used when writing data to the database to ensure that values are stored in a consistent binary format.
func  Serialize(val any, col *Column) ([]byte, error) {

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
	case TypeSerial:// Same as TypeInt
		buf := make([]byte, 4)
		v, ok := val.(int32)
		if !ok {
			return nil, fmt.Errorf("Expected int32 for TypeSerial, got %T", val)
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
// Takes the serialized byte representation of a value and a column definition, and returns the deserialized value based on the column's data type.
func Deserialize(data []byte, dataType uint8) (any, error) {
	switch dataType {
	case TypeTinyInt:
		return int8(data[0]), nil
	case TypeSmallInt:
		return int16(binary.BigEndian.Uint16(data)), nil
	case TypeInt:
		return int32(binary.BigEndian.Uint32(data)), nil
	case TypeSerial://same as TypeInt
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
// Compare takes two serialized byte representations of values and a column definition, and compares the two values based on the column's data type.
// returns -1 if val1 < val2
// returns 0 if val1 == val2
// returns 1 if val1 > val2
// This is used for sorting and searching within the database.
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
	case TypeSerial:// same as TypeInt
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
		val1Size := int(val1[0])
		val2Size := int(val2[0])
		return strings.Compare(string(val1[1:1+val1Size]), string(val2[1:1+val2Size])), nil
	default:
		return 0, fmt.Errorf("Unsupported data type for comparison")
	}
}
// returns the Constraints of a column as a readable string. 
// For example, a column with primary key and not null constraints would return "PRIMARY KEY, NOT NULL".
func (col *Column) PrintConstraints(constraints uint8) string {
	var result []string
	if constraints == 0 {
		return "None"
	}
	if constraints&ConstraintPrimaryKey != 0 {
		result = append(result, "PRIMARY KEY")
	}
	if constraints&ConstraintUnique != 0 {
		result = append(result, "UNIQUE")
	}
	if constraints&ConstraintNotNull != 0 {
		result = append(result, "NOT NULL")
	}
	if constraints&ConstraintIndex != 0 {
		result = append(result, "INDEX")
	}
	if constraints&ConstraintSerial != 0 {
		result = append(result, "SERIAL")
	}
	if constraints&ConstraintDefault != 0 {
		result = append(result, "DEFAULT = "+fmt.Sprintf("%v", col.Default))
	}
	return strings.Join(result, ", ")
}
// returns the string representation of the column's data type, including size for VARCHAR types.
func (col *Column) PrintDataType(dataType uint8) string {
	switch dataType {
	case TypeTinyInt:
		return "TINYINT"
	case TypeSmallInt:
		return "SMALLINT"
	case TypeInt:
		return "INT"
	case TypeSerial:
		return "SERIAL"
	case TypeBigInt:
		return "BIGINT"
	case TypeVarChar:
		if col.Size > 0 {
			return fmt.Sprintf("VARCHAR(%d)", col.Size)
		}
		return "VARCHAR"
	default:
		return "UNKNOWN"
	}
}
// GetPrimaryKeyColumn returns a pointer to the Column that is the primary key of the table, along with its index.
// If no primary key is found, it returns an error.
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
	case TypeSerial:// same as TypeInt
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
// GetDefaultValue takes a string representation of a value and converts it to the appropriate type based on the column's data type.
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
	case TypeSerial:// same as TypeInt
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

func (table *Table) GetColumnNames() []string {
	names := make([]string, len(table.Columns))
	for i, col := range table.Columns {
		names[i] = col.Name
	}
	return names
}