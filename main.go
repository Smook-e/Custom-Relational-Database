package main

import (
	"fmt"
	"os"
	"github.com/Smook-e/Custom-Relational-Database/filehandler"
)

const bufferSize int = 4096

func main(){
	fmt.Println("hello", filehandler.ReadFromFile("databse.txt"), filehandler.WriteToFile("database.txt"))


	file, err :=  os.OpenFile("database.bin", os.O_RDWR|os.O_CREATE, 0644)

	if err != nil {
		panic(err)
	}
	defer file.Close()
	buffer := make([]byte, bufferSize)

	for i := range buffer {
		buffer[i] = 'a'
	}
	
}