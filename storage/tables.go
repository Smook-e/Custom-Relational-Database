package storage

import (
	"errors"
	"fmt"
	"go/build/constraint"
	"strings"

	"github.com/Smook-e/Custom-Relational-Database/entities"
)


func CreateTable(tableName string, cols []entities.ColumnDefinition) (*entities.Table, error) {

	table := &entities.Table{Name: tableName}
	var dataType uint8
	var constraints uint8
	for _, col := range cols{
		cleanst := strings.ToLower(col.DataType)
		dataType, err := entities.GetDataType(cleanst)
		if err != nil {
			return nil, err
		}
		constraints &= 0
		for _, constraint := range col.Constraints {
			switch strings.ToLower(constraint) {
			case "primarykey":
				constraints |= entities.ConstraintPrimaryKey
			case "notnull":
				constraints |= entities.ConstraintNotNull
			case "unique":
				constraints |= entities.ConstraintUnique
			case "index":
				constraints |= entities.ConstraintIndex
			default:
				return nil, fmt.Errorf("Constraint %s not supported", constraint)
		}
		size, err := entities.GetSize(dataType)
		if err != nil {
			return nil, err
		}
		table.Columns = append(table.Columns, entities.Column{
			Name:        col.Name,
			DataType:    dataType,
			Contstraints: constraints,
			Size:        size,
		})
		}
	}
	return table,nil
}