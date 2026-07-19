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
	engine, err := storage.InitializeStorageEngine(filename)
	if err != nil {
		fmt.Print(err)
		return
	}
	defer engine.Bp.File.Close()
	engine.TestWriteandReadDatabase()
	engine, err = storage.InitializeStorageEngine(filename)
	if err != nil {
		fmt.Print(err)
		return
	}
    err = engine.TestOpenDatabase()
    if err != nil {
        fmt.Print(err)
    }
	// engine.TestIndexSearchPageRoot()
	engine.TestIndexInsertMiddleRoot()

    
}