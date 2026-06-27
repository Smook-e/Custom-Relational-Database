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


	
	// offset += 4
	
	
	// offset += 8
	
	
	
	// binary.BigEndian.PutUint32(slice[offset:offset+4], 1234)
	
	num1 := binary.BigEndian.Uint32(slice[offset: offset+4])
	offset += 4
	// binary.BigEndian.PutUint64(slice[offset:offset+8], 9999999999)
	num2 := binary.BigEndian.Uint64(slice[offset:offset+8])
	offset += 8
	
	// st := "Hello from database"
	
	// st_Len := len(st)
	
	// binary.BigEndian.PutUint16(slice[offset:offset + 2], uint16(st_Len))
	st_len := binary.BigEndian.Uint16(slice[offset: offset+2])
	offset += 2
	var st string = string(slice[offset: offset + int(st_len)])
	offset += int(st_len)
	
	// copy(slice[offset: offset + st_Len], st)
	
	// offset += st_Len
	
	fmt.Printf("num1: %d\nnum2 : %d\nstring: %s", num1,num2, st)

	// err =  filehandler.WriteToFile(file, page, buffer[page][:])

	if err != nil {
		fmt.Println(err) 
		return
	}
	

	// fmt.Println(buffer[0][:100])
}