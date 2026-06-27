package main

import (
	"fmt"
	"os"
	"log"
	"github.com/Smook-e/Custom-Relational-Database/filehandler"
)

const bufferSize int = 4096

func main(){
	// fmt.Println("hello", filehandler.ReadFromFile("database.txt"))


	file, err :=  os.OpenFile("database.bin", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Fatalf("Critical Error: Could not open database file: %v", err)
	}
	defer file.Close()

	buffer := new([10][4096]byte)

	for i := range buffer[0] {
		buffer[0][i] = 'c'
	}
	page := 1

	err =  filehandler.WriteToFile(file, page, buffer[0][:])
	if err != nil {
		fmt.Println("an error occured: ", err )
	}

	err = filehandler.ReadFromFile(file, page, buffer[0][:])


	fmt.Println(buffer[0][:100])
}