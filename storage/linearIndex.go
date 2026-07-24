package storage
import (
	"encoding/binary"
	"fmt"
)


func (engine *StorageEngine) GetFirstLeafPage(rootId uint32) (uint32, error) {
	//load page
	buffer, err := engine.Bp.Get(rootId)
	if err != nil {
		return 0 ,fmt.Errorf("An Error Occured %w", err)
	}
	//check if leaf
	isleaf := buffer[0]
	if isleaf == 1 {
		return rootId, nil
	}
	//get first child
	firstChildId := binary.BigEndian.Uint32(buffer[InternalPageHeaderSize:InternalPageHeaderSize + 4])
	return engine.GetFirstLeafPage(firstChildId)
}