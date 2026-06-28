package entities

import (
	"errors"
	"fmt"
	"strings"

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
	Contstraints uint8
	Size uint8
}
type ColumnDefinition struct {
	Name string
	DataType string
	Constraints []string
}

func GetSize(Type byte) (byte, error) {
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