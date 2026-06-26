package main

import (
	"fmt"
	"github.com/Smook-e/Custom-Relational-Database/filehandler"
)


func main(){
	fmt.Println("hello", filehandler.ReadFromFile("databse.txt"), filehandler.WriteToFile("database.txt"))

}