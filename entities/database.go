package entities

import (
	"os"
	"fmt"
	// "errors"
	
)


type Database struct {
	File *os.File
	Tables map[string]Table

	TotalPages int

}

