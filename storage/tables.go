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
		constraints, err  = entities.GetConstraint(col.Constraints)
		if err != nil {
			return nil, err
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
	
	return table,nil
}