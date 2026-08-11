package filehandler



import (
	"fmt"
	"os"
	"errors"
)

// ReadFromFile reads a page from the database file into the provided buffer. It ensures that the buffer is exactly 4096 bytes and reads the specified page from the file.
// It receives the pageId and the buffer to read into.
func ReadFromFile(file *os.File, page uint32, buffer []byte) error{

	if file == nil {
		return errors.New("Error: File pointer Not found")
	}
	if len(buffer) != bufferSize{
		return errors.New("Buffer has to have exactly 4096 bytes")

	}
	offset := int64(bufferSize * page)

	bytes_Read, err := file.ReadAt(buffer, offset)
	if err != nil {
		return	fmt.Errorf("An error occured while reading from file: %w", err)
	}
	if bytes_Read != bufferSize {
		return errors.New("Database file is corrupted, couldn't read 4096 bytes")
	}
	return nil
}