package filehandler


import (
	"os"
	"errors"
	"fmt"
)
const bufferSize = 4096

// WriteToFile writes a page to the database file from the provided buffer. It ensures that the buffer is exactly 4096 bytes and writes the specified page to the file.
func WriteToFile(file *os.File, page uint32, buffer []byte)  error{
	if file == nil{
		return errors.New("No file pointer found")
	}
	if len(buffer) != bufferSize {
		return  errors.New("Buffer has to have exactly 4096 bytes")
	}


	_, err := file.WriteAt(buffer, int64(page * bufferSize))
	if err != nil {
		return fmt.Errorf("An error occured while reading from file: %w", err)
	}
	return nil

}