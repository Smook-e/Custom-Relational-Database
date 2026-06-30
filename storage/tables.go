package storage

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Smook-e/Custom-Relational-Database/entities"
)


func CreateTable(tableName string, cols []entities.ColumnDefinition) (*entities.Table, error) {

	table := &entities.Table{Name: tableName}
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
			Constraints: constraints,
			Size:        size,
		})
	}
	
	return table,nil
}
func InsertRow(db *entities.Database, data []string, tableName string) (uint32, uint16, error) {
	//Pass 1: Check Validity and calculate size
	table, ok := db.Tables[tableName]
	if !ok {
		return 0,0, fmt.Errorf("Error: Table %s Not Found ", tableName)
	}
	if len(table.Columns) != len(data){
		return 0,0, fmt.Errorf("Error: Invalid input size. Please enter %d Fields", len(table.Columns))
	}
	size := 0
	
}