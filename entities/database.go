package entities

import (
	"os"
	"fmt"
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

func (db *Database) PrintFreePages() {
	for _, freePage := range db.FreePages {
        fmt.Printf(" Page: %d | Free Space: %d\n", freePage.PageID, freePage.FreeSpace)
    }

}