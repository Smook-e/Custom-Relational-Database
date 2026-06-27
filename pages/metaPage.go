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

	err := filehandler.ReadFromFile(db.File,0, buffer)
	if err != nil {
		return fmt.Errorf("An Error occured while reading Meta pages: %w", err)
	}
	offset := 0
	nextPage := binary.BigEndian.Uint16(buffer[offset:offset+2]); offset += 2;
	freeSpaceOffset := binary.BigEndian.Uint16(buffer[offset:offset+2]); offset += 2;
	numberOfTables := binary.BigEndian.Uint16(buffer[offset:offset+2]); offset += 2;

	for _ = range numberOfTables {
		table := &entities.Table{}
		tableStart := binary.BigEndian.Uint16(buffer[offset:offset+2]); offset += 2;
		tableOffset := tableStart
		nameLength := binary.BigEndian.Uint16(buffer[tableOffset:tableOffset+2]); tableOffset += 2;
		tableName := buffer[tableOffset: tableOffset + nameLength]; tableOffset += nameLength;
		
		numberOfColumns := buffer[tableOffset]; tableOffset++;
		
		for _ = range numberOfColumns {
			columnNameLength := binary.BigEndian.Uint16(buffer[tableOffset:tableOffset+2]); tableOffset += 2;
			columnName := buffer[tableOffset: tableOffset + nameLength]; tableOffset += nameLength;
			column := &entities.Column{Name: string(columnName)}
			column.DataType = buffer[tableOffset]; tableOffset++;
			column.Contstraints = buffer[tableOffset]; tableOffset++;
		}
	}

	return nil
}