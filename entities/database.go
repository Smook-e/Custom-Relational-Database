package entities

import (
	"os"
	
	// "errors"
	
)

type FreePage struct {
	PageID	uint32
	FreeSpace	uint16
}
type Database struct {
	File *os.File
	Tables map[string]*Table
	FreePages	[]FreePage
	TotalPages uint32

}

