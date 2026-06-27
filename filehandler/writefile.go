package filehandler


import (
	"os"
	"errors"
	"fmt"
)
const bufferSize int = 4096

func WriteToFile(file *os.File, page int, buffer []byte)  error{
	if file == nil{
		return errors.New("No file pointer found")
	}
	if len(buffer) != bufferSize {
		return  errors.New("Buffer has to have exactly 4096 bytes")
	}


	_, err := file.WriteAt(buffer, int64(page * bufferSize))
	if err != nil {
		fmt.Errorf("An error occured while reading from file: %w", err)
	}
	return nil

}