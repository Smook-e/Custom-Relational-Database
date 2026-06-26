package filehandler



import (
	"fmt"
)


func ReadFromFile(filename string) string{
	return fmt.Sprintf("Reading from file %s", filename)
}