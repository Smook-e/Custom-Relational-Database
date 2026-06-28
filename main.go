package main

import (
	"fmt"
	"log"
	"os"

	// "encoding/binary"
	// "github.com/Smook-e/Custom-Relational-Database/filehandler"
	// "github.com/Smook-e/Custom-Relational-Database/pages"
	// "github.com/Smook-e/Custom-Relational-Database/storage"
	"github.com/Smook-e/Custom-Relational-Database/entities"
	"github.com/Smook-e/Custom-Relational-Database/pages"
	"github.com/Smook-e/Custom-Relational-Database/storage"
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
	fileInfo, err := filep.Stat()
	
	if err != nil {
		log.Fatal(err)
	}
	db := &entities.Database{
		File: filep,
		Tables: make(map[string]*entities.Table),
		TotalPages: int(fileInfo.Size() / 4096),
	}
	defer db.File.Close()
	fmt.Printf(`Opened file
	Total pages : %d
	Tables : %v
`, db.TotalPages, db.Tables)

	db.Tables["products"], err = storage.CreateTable("products", []entities.ColumnDefinition{
		{
			Name: "id",
			DataType: "int",
			Constraints: []string{"primarykey", "notnull"},
		},
		{
			Name: "name",
			DataType: "varchar",
			Constraints: []string{"notnull"},
		},
		{
			Name: "price",
			DataType: "int",
			Constraints: []string{ "notnull"},
		},
	})
	db.Tables["users"], err = storage.CreateTable("users", []entities.ColumnDefinition{
		{
			Name: "id",
			DataType: "int",
			Constraints: []string{"primarykey", "notnull"},
		},
		{
			Name: "name",
			DataType: "varchar",
			Constraints: []string{"notnull"},
		},
		{
			Name: "age",
			DataType: "int",
			Constraints: []string{ "notnull"},
		},
	})
	
	err= pages.WriteMetaPage(db)
	if err != nil {
		log.Fatal(err)
	}
	
}