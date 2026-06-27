package filehandler



import (
	"fmt"
	"os"
	"errors"
)


func ReadFromFile(file *os.File, page int, buffer []byte) error{

	if file == nil {
		return errors.New("Error: File pointer Not found")
	}
	if len(buffer) != bufferSize{
		return errors.New("Buffer has to have exactly 4096 bytes")

	}
	offset := int64(bufferSize * page)

	_, err := file.ReadAt(buffer, offset)
	if err != nil {
		return	fmt.Errorf("An error occured while reading from file: %w", err)
	}

	return nil
}