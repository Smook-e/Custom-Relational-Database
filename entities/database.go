package entities

import (
	"os"
	
	// "errors"
	
)


type Database struct {
	File *os.File
	Tables map[string]Table

	TotalPages int

}

