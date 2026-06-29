package entities

import (
	"os"
	
	// "errors"
	
)

type FreePage struct {
	PageID	int
	FreeSpace	int
}
type Database struct {
	File *os.File
	Tables map[string]*Table
	FreePages	[]FreePage
	TotalPages int

}

