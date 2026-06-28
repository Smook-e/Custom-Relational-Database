package pages

import (
	// "os"
	"encoding/binary"
	"os"
	"fmt"

	"github.com/Smook-e/Custom-Relational-Database/entities"
	"github.com/Smook-e/Custom-Relational-Database/filehandler"
)

const bufferSize = 4096

func ReadMetaPage(db *entities.Database) error{
	buffer := make([]byte, bufferSize)

	
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
		for range numberOfTables {
			table := &entities.Table{}
			tableOffset := binary.BigEndian.Uint16(buffer[offset:offset+2]); offset += 2;
			nameLength := buffer[tableOffset]; tableOffset++;
			tableName := buffer[tableOffset: tableOffset + uint16(nameLength)]; tableOffset += uint16(nameLength);
			table.RootIndex = binary.BigEndian.Uint32(buffer[tableOffset:tableOffset+4]); tableOffset += 4;
			
			numberOfColumns := buffer[tableOffset]; tableOffset++;
			
			for range numberOfColumns {
				columnNameLength := buffer[tableOffset]; tableOffset++;
				columnName :=  buffer[tableOffset: tableOffset + uint16(columnNameLength)]; tableOffset += uint16(columnNameLength);
				column := &entities.Column{Name: string(columnName)}
				column.DataType = buffer[tableOffset]; tableOffset++;
				column.Contstraints = buffer[tableOffset]; tableOffset++;
				column.Size, _ = entities.GetSize(column.DataType)
				table.Columns = append(table.Columns, *column)
			
			}
			table.Name = string(tableName)
			db.Tables[string(tableName)] = *table
	
		}
		if nextPage == 0{
			break
		}
	}

	return nil
}

func WriteMetaPage(db *entities.Database) error {
	buffer := make([]byte, bufferSize)
	offset := 0
	binary.BigEndian.PutUint16(buffer,0); offset += 2;
	freeSpaceOffset := bufferSize; freeSpaceOffsetOffset := offset
	offset += 2
	numberOfTables := 0; numberOfTablesOffset := offset
	offset += 2
	keys :=  make([]string, len(db.Tables))
	for name := range db.Tables {
		keys = append(keys, name)
	}
	tabeBuffer := make([]byte, 128)
	var cols []entities.Column
	var col  *entities.Column
	for _, name := range keys {
		cols = db.Tables[name].Columns
		
	}

	return nil
}

func OpenDatabase(filename string) (*entities.Database, error) {
	filep, err :=  os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, fmt.Errorf("Critical Error: Could not open database file: %w", err)
	}
	fileInfo, err := filep.Stat()
	
	if err != nil {
		return nil, fmt.Errorf("Failed to retrieve file stats: %w", err)
	}
	db := &entities.Database{
		File: filep,
		Tables: make(map[string]entities.Table),
		TotalPages: int(fileInfo.Size() / bufferSize),
	}
	ReadMetaPage(db)
	return db, nil
}