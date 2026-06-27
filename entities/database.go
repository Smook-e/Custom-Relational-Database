package entities

import (
	"os"
)

type Database struct {
	file *os.File
	tables map[string]Table

	totalPages int

}