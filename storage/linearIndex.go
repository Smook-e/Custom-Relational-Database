package storage
import (
	"encoding/binary"
	"fmt"
	"github.com/Smook-e/Custom-Relational-Database/entities"
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

func (engine *StorageEngine) LinearTree(rootId uint32, key []byte, dataType uint8) (uint32, uint16, error) {
	leafId, err := engine.GetFirstLeafPage(rootId)
	if err != nil {
		return 0,0, fmt.Errorf("An Error Occured %w", err)
	}
	offset := 0
	for leafId != 0 {
		//load leaf page
		buffer, err := engine.Bp.Get(leafId)
		if err != nil {
			return 0 ,0,fmt.Errorf("An Error Occured %w", err)
		}
		
		//get next leaf page
		leafId = binary.BigEndian.Uint32(buffer[nextLeafPageOffset:nextLeafPageOffset + 4])
		//get number of keys
		numKeys := binary.BigEndian.Uint16(buffer[leafPageNumEntriesOffset:leafPageNumEntriesOffset + 2])
		offset = LeafPageHeaderSize
		for range numKeys {
			comp , err := entities.Compare(key, buffer[offset: offset + len(key)], dataType)
			if err != nil {
				return 0,0, fmt.Errorf("An Error Occured %w", err)
			}
			offset += len(key)
			pageId := binary.BigEndian.Uint32(buffer[offset:offset + 4])
			offset += 4
			slot := binary.BigEndian.Uint16(buffer[offset:offset + 2])
			offset += 2
			if comp == 0 {
				return pageId, slot, nil
			}
			if comp < 0 {
				return 0,0, fmt.Errorf("Key not found")
			}
		}
		
	}
	
	
	return 0,0, fmt.Errorf("Key not found")
}

// func (engine *StorageEngine) LinearSearch()