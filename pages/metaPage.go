package pages

import (
	// "os"
	
	"encoding/binary"
	"fmt"
	// "sort"

	
	"sync"
	"sort"
	"github.com/Smook-e/Custom-Relational-Database/entities"
	"github.com/Smook-e/Custom-Relational-Database/filehandler"

	
)

const bufferSize = 4096

var bufferPool = sync.Pool{
    New: func() any{
        // This returns a reusable buffer of the specified size. 
        return make([]byte, bufferSize)
    },
}
/*
Meta Page Structure:
- Next Meta Page Pointer (4 bytes)
- Free Space Offset (2 bytes)
- Number of Tables (2 bytes)
- [List of Table offsets (2 bytes each)]
At each Table offset:
	- Table Name Length (1 byte)
	- Table Name (variable length)
	- Number of Columns (1 byte)
	- For each Column:
		- Column Name Length (1 byte)
		- Column Name (variable length)
		- Data Type (1 byte)
		- Constraints (1 byte)
		- Default Value (variable length, if applicable)
		- Size (1 byte)
	- Number of Indexes (1 byte)
	- For each Index:
		- Column index in Columns array (1 byte)
		- Index Page ID (4 bytes)
	- Number of Foreign Keys (1 byte)
	- For each Foreign Key:
		- Column index in Columns array (1 byte)
		- Referenced Table Name Length (1 byte)
		- Referenced Table Name (variable length)
		- Referenced Column index in Referenced Table's Columns array (1 byte)
*/

func ReadMetaPage(db *entities.Database) error{
	buffer := bufferPool.Get().([]byte)
	defer bufferPool.Put(buffer)

	
	var nextPage uint32 = 0
	
	for{
		err := filehandler.ReadFromFile(db.File,nextPage, buffer)
		if err != nil {
			return fmt.Errorf("An Error occured while reading Meta pages: %w", err)
		}
		offset := 0
		nextPage = binary.BigEndian.Uint32(buffer[offset:offset+4]);
		offset += 4;
		
		//freeSpaceOffset := binary.BigEndian.Uint16(buffer[offset:offset+2]); 
		offset += 2;
		
		numberOfTables := binary.BigEndian.Uint16(buffer[offset:offset+2]); offset += 2;
		// Loop through each table and read its metadata
		for range numberOfTables {
			table := &entities.Table{}
			tableOffset := int(binary.BigEndian.Uint16(buffer[offset:offset+2])); offset += 2;// read the table offset then jump to that offset to read the table data
			nameLength := int(buffer[tableOffset]); tableOffset++;
			tableName := string(buffer[tableOffset: tableOffset + nameLength]); tableOffset += nameLength;
			table.Name = tableName
			db.Tables[tableName] = table
			numberOfColumns := buffer[tableOffset]; tableOffset++;
			// Loop through each column and read its metadata
			for range numberOfColumns {
				columnNameLength := buffer[tableOffset]; tableOffset++;
				columnName :=  buffer[tableOffset: tableOffset + int(columnNameLength)]; tableOffset += int(columnNameLength);
				column := &entities.Column{Name: string(columnName)}
				column.DataType = buffer[tableOffset]; tableOffset++;
				column.Constraints = buffer[tableOffset]; tableOffset++;
				column.Size, _ = entities.GetSize(column.DataType); tableOffset++;
				table.Columns = append(table.Columns, *column)
				
			}
			
			numberOfIndexes := int(buffer[tableOffset]); tableOffset++;
			table.Indexes = make(map[string]uint32, numberOfIndexes)
			// Loop through each index and read its metadata
			for range numberOfIndexes {
				columnIndex := int(buffer[tableOffset]); tableOffset++;
				indexPageID := binary.BigEndian.Uint32(buffer[tableOffset:tableOffset+4]); tableOffset += 4;
				if columnIndex < 0 || columnIndex >= len(table.Columns) {
					return fmt.Errorf("Invalid column index %d for table %s", columnIndex, string(tableName))
				}
				columnName := table.Columns[columnIndex].Name
				table.Indexes[columnName] = indexPageID
			}

			numberOfForeignKeys := int(buffer[tableOffset]); tableOffset++;
			table.ForeignKeys = make(map[string]entities.ForeignKeyReference, numberOfForeignKeys)
			// Loop through each foreign key and read its metadata
			for range numberOfForeignKeys {
				columnIndex := int(buffer[tableOffset]); tableOffset++;
				referencedTableNameLength := int(buffer[tableOffset]); tableOffset++;
				referencedTableName := string(buffer[tableOffset: tableOffset + referencedTableNameLength]); tableOffset += referencedTableNameLength;
				referencedColumnIndex := uint8(buffer[tableOffset]); tableOffset++;
				table.ForeignKeys[table.Columns[columnIndex].Name] = entities.ForeignKeyReference{
					ReferencedTableName: referencedTableName,
					ReferencedColumnIndex: referencedColumnIndex,
				}
			}
	
		}
		err = ReadFreeSpacePage(db)
		if err != nil {
			return err
		}
		if nextPage == 0{
			break
		}
	}

	return nil
}

func WriteMetaPage(db *entities.Database) error {
	buffer := bufferPool.Get().([]byte)
	defer bufferPool.Put(buffer)
	offset := 0
	binary.BigEndian.PutUint32(buffer,0); offset += 4;//write next page
	freeSpaceOffset := bufferSize; freeSpaceOffsetOffset := offset
	offset += 2
	numberOfTables := 0; numberOfTablesOffset := offset
	offset += 2
	keys :=  make([]string,0, len(db.Tables))
	for name := range db.Tables {
		keys = append(keys, name)
	}
	// Sort the keys to ensure deterministic order
	sort.Strings(keys)
	
	var cols []entities.Column
	var table *entities.Table

	for _, name := range keys {
		size := 0
		table = db.Tables[name]
		cols = table.Columns
		//Pass 1 : Calculate the size of the table entry to determine where to write it in the buffer
		//length of name + name + number of columns + number of ForeignKeys + number of indexes +  indexes * 5 
		size += 1 + len(table.Name) + 1 + 1 + 1 + len(table.Indexes) * 5
		// calculate the size of each column entry
		for _, col := range cols {
			// length of name + name + datatype + constraints + size
			size += 1 + len(col.Name) + 1 + 1 + 1
		}
		// calculate the size of each foreign key entry
		for _, fk := range table.ForeignKeys {
			// column index + referenced table name length + referenced table name + referenced column index
			size += 1 + 1 + len(fk.ReferencedTableName) + 1
		}
		tableOffset := freeSpaceOffset - size
		freeSpaceOffset = tableOffset
		numberOfTables++
		
		binary.BigEndian.PutUint16(buffer[offset:offset+2], uint16(tableOffset)); offset += 2 //add the table offset slot
		//Pass 2: write the actual content
		buffer[tableOffset] = uint8(len(table.Name)); tableOffset++;
		copy(buffer[tableOffset: tableOffset + len(table.Name)], table.Name); tableOffset+= len(table.Name)


		buffer[tableOffset] = uint8(len(table.Columns)); tableOffset++;
		// Sort the column names to ensure consistent order
		sortedColumnNames := make([]string, 0, len(table.Columns))
		
		for _, col := range cols {
			sortedColumnNames = append(sortedColumnNames, col.Name)
		}
		// sort.Strings(sortedColumnNames)
		// write each column's data
		for _, colName := range sortedColumnNames {
			col, err := table.GetColumnByName(colName)
			if err != nil {
				return fmt.Errorf("Error getting column for %s: %v", colName, err)
			}
			buffer[tableOffset] = uint8(len(col.Name));tableOffset++;
			copy(buffer[tableOffset: tableOffset + len(col.Name)], col.Name); tableOffset+= len(col.Name);
			buffer[tableOffset] = col.DataType; tableOffset++;
			buffer[tableOffset] = col.Constraints; tableOffset++;
			buffer[tableOffset] = col.Size; tableOffset++;
		}

		
		buffer[tableOffset] = uint8(len(table.Indexes)); tableOffset++;
		// Extract and sort index keys deterministically
		var indexCols []string
		for colName := range table.Indexes {
			indexCols = append(indexCols, colName)
		}
		sort.Strings(indexCols)
		for _, colName := range indexCols {
			indexPageID := table.Indexes[colName]
			colIndex, err := table.GetColumnIndexByName(colName)
			if err != nil {
				return fmt.Errorf("Error getting column index for %s: %v", colName, err)
			}
			buffer[tableOffset] = uint8(colIndex); tableOffset++
			binary.BigEndian.PutUint32(buffer[tableOffset:tableOffset+4], indexPageID); tableOffset += 4
		}

		buffer[tableOffset] = uint8(len(table.ForeignKeys)); tableOffset++;
		// Extract and sort foreign key columns deterministically
		var fkCols []string
		for colName := range table.ForeignKeys {
			fkCols = append(fkCols, colName)
		}
		sort.Strings(fkCols)
		
		for _, colName := range fkCols {
			fk := table.ForeignKeys[colName]
			colIndex, err := table.GetColumnIndexByName(colName)
			if err != nil {
				return fmt.Errorf("Error getting column index for %s: %v", colName, err)
			}
			buffer[tableOffset] = uint8(colIndex); tableOffset++
			buffer[tableOffset] = uint8(len(fk.ReferencedTableName)); tableOffset++
			copy(buffer[tableOffset: tableOffset + len(fk.ReferencedTableName)], fk.ReferencedTableName); tableOffset+= len(fk.ReferencedTableName);
			buffer[tableOffset] = fk.ReferencedColumnIndex; tableOffset++
		}
	}
	
	binary.BigEndian.PutUint16(buffer[freeSpaceOffsetOffset: freeSpaceOffsetOffset + 2], uint16(freeSpaceOffset))// assign the final Free space offset
	binary.BigEndian.PutUint16(buffer[numberOfTablesOffset: numberOfTablesOffset + 2], uint16(numberOfTables))// assign the final Number of tables
	db.File.WriteAt(buffer, 0)
	err := WriteFreeSpacePage(db)
	if err != nil {
		return fmt.Errorf("WriteFreeSpacePage failed: %v", err)
	}

	return nil
}



































// //
// func OpenDatabase(filename string) (*entities.Database, error) {
// 	filep, err :=  os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
// 	if err != nil {
// 		return nil, fmt.Errorf("Critical Error: Could not open database file: %w", err)
// 	}
// 	fileInfo, err := filep.Stat()
	
// 	if err != nil {
// 		return nil, fmt.Errorf("Failed to retrieve file stats: %w", err)
// 	}
// 	db := &entities.Database{
// 		File: filep,
// 		Tables: make(map[string]*entities.Table),
// 		TotalPages: uint32(fileInfo.Size() / bufferSize),
// 	}
// 	err = ReadMetaPage(db)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return db, nil
// }

// func TestOpenDatabase(filename string) error {
// 	db, err := OpenDatabase(filename)
//     if err != nil {
//         return fmt.Errorf("OpenDatabase failed: %v", err)
//     }
//     defer db.File.Close()
// 	// db.File.Truncate(2 * bufferSize)
    
    
// 	pageID, slot, err := InsertRow(db, []string{"1", "joe", "20"}, "users")
// 	if err != nil {
// 		return err
// 	}
// 	Row, err := ReadRow(db, "users", pageID, slot)
// 	if err != nil {
// 		return err
// 	}
// 	fmt.Println(Row)
// 	pageID, slot, err = InsertRow(db, []string{"2", "emily", "25"}, "users")
// 	if err != nil {
// 		return err
// 	}
// 	Row, err = ReadRow(db, "users", pageID, slot)
// 	if err != nil {
// 		return err
// 	}
// 	fmt.Println(Row)
// 	pageID, slot, err = InsertRow(db, []string{"1", "Phone", "1000"}, "products")
// 	if err != nil {
// 		return err
// 	}
// 	Row, err = ReadRow(db, "products", pageID, slot)
// 	if err != nil {
// 		return err
// 	}
// 	fmt.Println(Row)
// 	fmt.Println("Free Pages:")
//     for _, freePage := range db.FreePages {
//         fmt.Printf(" Page: %d | Free Space: %d\n", freePage.PageID, freePage.FreeSpace)
//     }
// 	WriteMetaPage(db)
// 	return nil
// }
// func TestWriteandReadDatabase(filename string) error {
// 	filep, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
// 	filep.Truncate(0)
//     if err != nil {
//         return err
//     }

//     db := &entities.Database{
//         File:   filep,
//         Tables: make(map[string]*entities.Table),
//     }

    
//     // initialize tables
    
//     t1, err := db.CreateTable("products", []entities.ColumnDefinition{
//         {Name: "id", DataType: "int", Constraints: []string{"primarykey", "notnull"}},
//         {Name: "name", DataType: "varchar", Constraints: []string{"notnull"}},
//         {Name: "price", DataType: "int", Constraints: []string{"notnull"}},
//     })
//     if err != nil {
//         return err
//     }
//     db.Tables[t1.Name] = t1

//     // Table 2
//     t2, err := db.CreateTable("users", []entities.ColumnDefinition{
//         {Name: "id", DataType: "int", Constraints: []string{"primarykey"}},
//         {Name: "name", DataType: "varchar", Constraints: []string{"notnull"}},
//         {Name: "age", DataType: "int", Constraints: []string{}},
//     })
//     if err != nil {
//         return err
//     }
//     db.Tables[t2.Name] = t2
// 	db.FreePages = []entities.FreePage{}
	
    
//     // Write the meta page to the file
//     err = WriteMetaPage(db)
//     if err != nil {
//         return fmt.Errorf("WriteMetaPage failed: %v", err)
//     }

//     // Close the file to ensure all data is flushed
//     db.File.Close() 
    

    
//     // Reopen the database to test recovery
//     db2, err := OpenDatabase(filename)
//     if err != nil {
//         return fmt.Errorf("OpenDatabase failed: %v", err)
//     }
//     defer db2.File.Close()

    
    
//     if len(db2.Tables) == 0 {
//         fmt.Println("Error: No tables were recovered!")
//     } else {
//         for name, table := range db2.Tables {
//             fmt.Printf("Table: %s | Columns: %d\n", name, len(table.Columns))

//             for _, col := range table.Columns {
//                 fmt.Printf(" Column: %s | Type: %d | Constraints: %v\n", col.Name, col.DataType, col.Constraints)
//             }
//         }
//     }
// 	// pageID, slot, err := InsertRow(db2, []string{"1", "joe", "20"}, "users")
// 	// if err != nil {
// 	// 	return err
// 	// }
// 	// Row, err := ReadRow(db2, "users", pageID, slot)
// 	// if err != nil {
// 	// 	return err
// 	// }
// 	// fmt.Println(Row)
// 	fmt.Println("Free Pages:")
//     for _, freePage := range db2.FreePages {
//         fmt.Printf(" Page: %d | Free Space: %d\n", freePage.PageID, freePage.FreeSpace)
//     }
// 	WriteMetaPage(db2)
// 	return nil
// }