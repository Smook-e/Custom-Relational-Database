package main

import (
	"fmt"
	"os"
	"log"
	// "encoding/binary"
	// "github.com/Smook-e/Custom-Relational-Database/filehandler"
	// "github.com/Smook-e/Custom-Relational-Database/pages"
	// "github.com/Smook-e/Custom-Relational-Database/storage"
	"github.com/Smook-e/Custom-Relational-Database/entities"
)



func main(){
	// fmt.Println("hello", filehandler.ReadFromFile("database.txt"))
	// db, err := pages.OpenDatabase("database.bin")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	filep, err :=  os.OpenFile("database.bin", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer filep.Close()
	fileInfo, err := filep.Stat()
	
	if err != nil {
		log.Fatal(err)
	}
	db := &entities.Database{
		File: filep,
		Tables: make(map[string]*entities.Table),
		TotalPages: int(fileInfo.Size() / 4096),
	}
	fmt.Printf(`Opened file
	Total pages : %d
	Tables : %v
	`, db.TotalPages, db.Tables)

	
}