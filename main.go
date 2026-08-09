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
	
	engine, err := storage.InitializeStorageEngine(filename)
	if err != nil {
		fmt.Print(err)
	}
	defer engine.Bp.File.Close()
	engine.PrintMetaData()
	rows , err := engine.LinearSearch("products", nil)
	if err != nil {
		fmt.Printf("failed to perform linear search on products table: %v", err)
	}
	fmt.Println("Sample rows in products table:")
	for _, row := range rows {
		fmt.Println(row)
	}
	rows , err = engine.LinearSearch("users", nil)
	if err != nil {
		fmt.Printf("failed to perform linear search on users table: %v", err)
	}
	fmt.Println("Sample rows in users table:")
	for _, row := range rows {
		fmt.Println(row)
	}
	// engine.TestIndexInsertMiddleRoot(0)
	// engine.TestIndexInsertStringMiddleRoot()
	

    
}