package filehandler


import (
	"fmt"
)


func WriteToFile(filename string) string{
	return fmt.Sprintf("Writing to file %s", filename)
}