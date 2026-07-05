package main

import (
	"fmt"
	// "log"
	// "os"

	// "encoding/binary"
	// "github.com/Smook-e/Custom-Relational-Database/filehandler"
	
	
	// "github.com/Smook-e/Custom-Relational-Database/entities"
	// "github.com/Smook-e/Custom-Relational-Database/pages"
	"github.com/Smook-e/Custom-Relational-Database/storage"
)



func main(){

	filename := "database.bin"
	// pages.TestWriteandReadDatabase(filename)
    // err := pages.TestOpenDatabase(filename)
	engine := storage.StorageEngine{}
	engine.TestWriteandReadDatabase(filename)
    err := engine.TestOpenDatabase(filename)
    if err != nil {
        fmt.Print(err)
    }
    
    
    
    
    
}