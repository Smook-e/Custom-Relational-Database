package main

import (
	"fmt"
	"os"
	"log"
	"github.com/Smook-e/Custom-Relational-Database/filehandler"
)



func main(){
	// fmt.Println("hello", filehandler.ReadFromFile("database.txt"))


	file, err :=  os.OpenFile("database.bin", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Fatalf("Critical Error: Could not open database file: %v", err)
	}
	defer file.Close()

	buffer := make([][4096]byte, 10)

	write := make([]byte, 4096)

	for i := range write {
		write[i] = 'A'
	}
	page := 2


	err =  filehandler.WriteToFile(file, page, write)
	
	if err != nil {
		fmt.Println("an error occured: ", err )
	}

	err = filehandler.ReadFromFile(file, page, buffer[0][:])


	fmt.Println(buffer[0][:100])
}