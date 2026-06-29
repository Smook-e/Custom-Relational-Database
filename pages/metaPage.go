package pages

import (
	// "os"
	// "container/list"
	"encoding/binary"
	"fmt"
	
	"os"

	"github.com/Smook-e/Custom-Relational-Database/entities"
	"github.com/Smook-e/Custom-Relational-Database/filehandler"
	"github.com/Smook-e/Custom-Relational-Database/storage"
)

const bufferSize = 4096

func ReadMetaPage(db *entities.Database) error{
	buffer := make([]byte, bufferSize)

	
	var nextPage uint16 = 0
	
	for{
		err := filehandler.ReadFromFile(db.File,int(nextPage) * bufferSize, buffer)
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
			tableOffset := int(binary.BigEndian.Uint16(buffer[offset:offset+2])); offset += 2;
			nameLength := int(buffer[tableOffset]); tableOffset++;
			tableName := buffer[tableOffset: tableOffset + nameLength]; tableOffset += nameLength;
			table.RootIndex = binary.BigEndian.Uint32(buffer[tableOffset:tableOffset+4]); tableOffset += 4;
			
			numberOfColumns := buffer[tableOffset]; tableOffset++;
			
			for range numberOfColumns {
				columnNameLength := buffer[tableOffset]; tableOffset++;
				
				columnName :=  buffer[tableOffset: tableOffset + int(columnNameLength)]; tableOffset += int(columnNameLength);
				column := &entities.Column{Name: string(columnName)}
				column.DataType = buffer[tableOffset]; tableOffset++;
				column.Constraints = buffer[tableOffset]; tableOffset++;
				column.Size, _ = entities.GetSize(column.DataType); tableOffset++;
				table.Columns = append(table.Columns, *column)
			
			}
			table.Name = string(tableName)
			db.Tables[string(tableName)] = table
	
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
	keys :=  make([]string,0, len(db.Tables))
	for name := range db.Tables {
		keys = append(keys, name)
	}
	
	var cols []entities.Column
	var table *entities.Table
	for _, name := range keys {
		size := 0
		table = db.Tables[name]
		cols = table.Columns
		//Pass 1 : Calculate the size of the columns
		//length of name + name + root index
		size += 1 + len(table.Name) + 4 +1
		for _, col := range cols {
			// length of name + name + datatype + constraints + size
			size += 1 + len(col.Name) + 1 + 1 + 1
		}
		tableOffset := freeSpaceOffset - size
		freeSpaceOffset = tableOffset
		numberOfTables++
		
		binary.BigEndian.PutUint16(buffer[offset:offset+2], uint16(tableOffset)); offset += 2 //add the table offset slot
		//Pass 2: write the actual content
		buffer[tableOffset] = uint8(len(table.Name)); tableOffset++;
		copy(buffer[tableOffset: tableOffset + len(table.Name)], table.Name); tableOffset+= len(table.Name)
		binary.BigEndian.PutUint32(buffer[tableOffset:tableOffset+4], table.RootIndex); tableOffset += 4
		buffer[tableOffset] = uint8(len(cols)); tableOffset++;
		for _, col := range cols {
			buffer[tableOffset] = uint8(len(col.Name));tableOffset++;
			copy(buffer[tableOffset: tableOffset + len(col.Name)], col.Name); tableOffset+= len(col.Name);
			buffer[tableOffset] = col.DataType; tableOffset++;
			buffer[tableOffset] = col.Constraints; tableOffset++;
			buffer[tableOffset] = col.Size; tableOffset++;
		}





	}
	
	binary.BigEndian.PutUint16(buffer[freeSpaceOffsetOffset: freeSpaceOffsetOffset + 2], uint16(freeSpaceOffset))// assign the final Free space offset
	binary.BigEndian.PutUint16(buffer[numberOfTablesOffset: numberOfTablesOffset + 2], uint16(numberOfTables))// assign the final Number of tables
	db.File.WriteAt(buffer, 0)
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
		Tables: make(map[string]*entities.Table),
		TotalPages: int(fileInfo.Size() / bufferSize),
	}
	err = ReadMetaPage(db)
	if err != nil {
		return nil, err
	}
	return db, nil
}
func TestOpenDatabase(filename string) error {
	filep, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
    if err != nil {
        return err
    }

    db := &entities.Database{
        File:   filep,
        Tables: make(map[string]*entities.Table),
    }

    
    // initialize tables
    
    t1, err := storage.CreateTable("products", []entities.ColumnDefinition{
        {Name: "id", DataType: "int", Constraints: []string{"primarykey", "notnull"}},
        {Name: "name", DataType: "varchar", Constraints: []string{"notnull"}},
        {Name: "price", DataType: "int", Constraints: []string{"notnull"}},
    })
    if err != nil {
        return err
    }
    db.Tables[t1.Name] = t1

    // Table 2
    t2, err := storage.CreateTable("users", []entities.ColumnDefinition{
        {Name: "id", DataType: "int", Constraints: []string{"primarykey"}},
        {Name: "name", DataType: "varchar", Constraints: []string{"notnull"}},
        {Name: "age", DataType: "int", Constraints: []string{}},
    })
    if err != nil {
        return err
    }
    db.Tables[t2.Name] = t2

    
    // Write the meta page to the file
    err = WriteMetaPage(db)
    if err != nil {
        return fmt.Errorf("WriteMetaPage failed: %v", err)
    }

    // Close the file to ensure all data is flushed
    db.File.Close() 
    fmt.Println("File closed and flushed.")

    
    // Reopen the database to test recovery
    db2, err := OpenDatabase(filename)
    if err != nil {
        return fmt.Errorf("OpenDatabase failed: %v", err)
    }
    defer db2.File.Close()

    
    
    if len(db2.Tables) == 0 {
        fmt.Println("Error: No tables were recovered!")
    } else {
        for name, table := range db2.Tables {
            fmt.Printf("Table: %s | Columns: %d\n", name, len(table.Columns))
            for _, col := range table.Columns {
                fmt.Printf(" Column: %s | Type: %d | Constraints: %v\n", col.Name, col.DataType, col.Constraints)
            }
        }
    }

}