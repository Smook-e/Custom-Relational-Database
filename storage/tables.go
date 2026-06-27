package storage

import (
	
	
	
	"fmt"
	"strings"

	"github.com/Smook-e/Custom-Relational-Database/entities"
	
)


func CreateTable(tableName string, cols []entities.ColumnDefinition) (*entities.Table, error) {

	table := &entities.Table{Name: tableName}
	var dataType uint8
	for i, col := range cols{
		cleanst := strings.ToLower(col.DataType)
		switch cleanst {
		case "tinyint":
			dataType = entities.TypeTinyInt
		case "smallint":
			dataType = entities.TypeSmallInt
		case "bigint":
			dataType = entities.TypeBigInt
		case "int":
			dataType = entities.TypeInt
		case "varchar":
			dataType = entities.TypeVarChar
		}

		

	}
	

	return table,nil
}