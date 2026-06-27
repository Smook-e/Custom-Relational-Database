package pages

import (
	// "os"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/Smook-e/Custom-Relational-Database/entities"
	"github.com/Smook-e/Custom-Relational-Database/filehandler"
)



func ReadMetaPage(db *entities.Database) error{
	buffer := make([]byte, 4096)

	
	var nextPage uint16 = 0
	
	for{
		err := filehandler.ReadFromFile(db.File,int(nextPage), buffer)
		if err != nil {
			return fmt.Errorf("An Error occured while reading Meta pages: %w", err)
		}
		offset := 0
		nextPage = binary.BigEndian.Uint16(buffer[offset:offset+2]); offset += 2;
		//freeSpaceOffset := binary.BigEndian.Uint16(buffer[offset:offset+2]); 
		offset += 2;
		
		numberOfTables := binary.BigEndian.Uint16(buffer[offset:offset+2]); offset += 2;
		for _ = range numberOfTables {
			table := &entities.Table{}
			tableOffset := binary.BigEndian.Uint16(buffer[offset:offset+2]); offset += 2;
			nameLength := binary.BigEndian.Uint16(buffer[tableOffset:tableOffset+2]); tableOffset += 2;
			tableName := buffer[tableOffset: tableOffset + nameLength]; tableOffset += nameLength;
			table.RootIndex = binary.BigEndian.Uint32(buffer[tableOffset:tableOffset+4]); tableOffset += 4;
			
			numberOfColumns := buffer[tableOffset]; tableOffset++;
			
			for _ = range numberOfColumns {
				columnNameLength := binary.BigEndian.Uint16(buffer[tableOffset:tableOffset+2]); tableOffset += 2;
				columnName := buffer[tableOffset: tableOffset + columnNameLength]; tableOffset += columnNameLength;
				column := &entities.Column{Name: string(columnName)}
				column.DataType = buffer[tableOffset]; tableOffset++;
				column.Contstraints = buffer[tableOffset]; tableOffset++;
				column.Size, _ = entities.GetSize(column.DataType)
				table.Columns = append(table.Columns, *column)
			
			}
			db.Tables[string(tableName)] = *table
	
		}
		if nextPage == 0{
			break
		}
	}

	return nil
}