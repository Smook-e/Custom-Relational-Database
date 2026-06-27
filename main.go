package main

import (
	"fmt"
	"os"
	"log"
	"encoding/binary"
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

	// write := make([]byte, 4096)

	// for i := range write {
	// 	write[i] = 'A'
	// }
	page := 0
	
	err = filehandler.ReadFromFile(file, page, buffer[page][:])

	if err != nil {
		fmt.Println(err)
		return
	}
	offset := 0

	slice := buffer[page][:]


	num1 := binary.BigEndian.Uint32(slice[offset: offset+4])

	offset += 4

	num2 := binary.BigEndian.Uint64(slice[offset:offset+8])

	offset += 8

	fmt.Printf("num1: %d\n num2 : %d", num1,num2)
	// binary.BigEndian.PutUint32(slice[offset:offset+4], 1234)

	// offset += 4
	// binary.BigEndian.PutUint64(slice[offset:offset+8], 9999999999)
	// offset += 8


	err =  filehandler.WriteToFile(file, page, buffer[page][:])

	if err != nil {
		fmt.Println(err) 
		return
	}
	

	// fmt.Println(buffer[0][:100])
}