package main

import (
	"fmt"
	"os"
	"github.com/Smook-e/Custom-Relational-Database/cli"
)



func main(){
	if len(os.Args) < 2 {
        fmt.Println("usage: sql_engine <database_file>")
        os.Exit(1)
    }
    path := os.Args[1]
	cli.MainLoop(path)   
}