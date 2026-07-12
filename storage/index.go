package storage

import (
	"fmt"
	"github.com/Smook-e/Custom-Relational-Database/entities"
	"encoding/binary"
)
const (
	LeafPageHeaderSize     = 1 + 4 + 2 // isLeaf + nextLeafPage + numberOfEntries
	InternalPageHeaderSize = 1 + 2   // isLeaf + numberOfEntries
)

// Recursive function to insert a key into the index starting from the root page, and split the page if necessary. Returns the new root page ID and the new key to be inserted into the parent page if a split occurs.
func (engine *StorageEngine) IndexInsert(root uint32, key []byte, pageID uint32, slot uint16, dataType uint8) (uint32, []byte, error) {

	buffer , err := engine.Bp.Get(root)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to get buffer for page %d: %w", root, err)
	}
	// Check if the page is a leaf or internal page
	offset := 0
	if buffer[offset] == 0 {
		// Internal page
		offset += 1
		maxEntries := (4096 - InternalPageHeaderSize) / (len(key) + 4) // 4 bytes for pageID
		numberOfEntries := binary.BigEndian.Uint16(buffer[offset : offset+2])
		offset += 2
		found := false
		for range numberOfEntries {
			leftptr := binary.BigEndian.Uint32(buffer[offset : offset+4]) // left PageID
			offset += 4
			comp, err := entities.Compare(key, buffer[offset:offset+len(key)], dataType)
			if err != nil {
				return 0, nil, fmt.Errorf("comparison error: %w", err)
			}
			if comp < 0 {
				// Key is less than the current entry's key, so we should go to the left pointer
				found = true
				newRoot, newKey, err := engine.IndexInsert(leftptr, key, pageID, slot, dataType)
				if err != nil {
					return 0, nil, err
				}
				break
			}
		}
		// If we reach here, it means the key is greater than all entries, so we should go to the right pointer
		if !found {
			rightptr := binary.BigEndian.Uint32(buffer[offset : offset+4])
			newRoot, newKey, err := engine.IndexInsert(rightptr, key, pageID, slot, dataType)
			if err != nil {
				return 0, nil, err
			}
		}
		// Handle the case where a split occurred in the child page
		if newKey != nil {
			if numberOfEntries < uint16(maxEntries) {
				// Insert the new key and pageID into this internal page
				// (Implementation of insertion logic goes here)
				return root, nil, nil
			} else {
				// Split this internal page and return the new root and key to be inserted into the parent
				// (Implementation of split logic goes here)
				return newRoot, newKey, nil
			}
		}
	}
}





func (engine *StorageEngine) IndexSearch(root uint32, key []byte, dataType uint8) (uint32, uint16, error) {
	buffer , err := engine.Bp.Get(root)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get buffer for page %d: %w", root, err)
	}
	// Check if the page is a leaf or internal page
	offset := 0
	if buffer[offset] == 0 {
		// Internal page
		offset += 1 
		// Read the number of entries
		numberOfEntries := binary.BigEndian.Uint16(buffer[offset : offset+2])
		offset += 2
		for range numberOfEntries {
			leftptr := binary.BigEndian.Uint32(buffer[offset : offset+4])// left PageID
			offset += 4
			comp, err := entities.Compare(key, buffer[offset:offset+len(key)], dataType)
			if err != nil {
				return 0, 0, fmt.Errorf("comparison error: %w", err)
			}
			if comp < 0 {
				// Key is less than the current entry's key, so we should go to the left pointer
				return engine.IndexSearch(leftptr, key, dataType)
			}
			offset += len(key)
		}
		// If we reach here, it means the key is greater than all entries, so we should go to the right pointer
		rightptr := binary.BigEndian.Uint32(buffer[offset : offset+4])
		return engine.IndexSearch(rightptr, key, dataType)
	} else {
		// Leaf page
		offset += 1 + 4 // Skip isLeaf and nextLeafPage
		numberOfEntries := binary.BigEndian.Uint16(buffer[offset : offset+2])
		offset += 2
		for range numberOfEntries {
			comp, err := entities.Compare(key, buffer[offset:offset+len(key)], dataType)
			if err != nil {
				return 0, 0, fmt.Errorf("comparison error: %w", err)
			}
			offset += len(key)
			if comp == 0 {
				// Key found, return the pageID and slot
				pageID := binary.BigEndian.Uint32(buffer[offset : offset+4])
				offset += 4
				slot := binary.BigEndian.Uint16(buffer[offset : offset+2])
				return pageID, slot, nil
			}else if comp < 0 {
				// Key is less than the current entry's key, so it doesn't exist in this leaf page
				return 0, 0, fmt.Errorf("key not found")
			}
			offset += 4 + 2 // Skip pageID and slot
		}
		return 0, 0, fmt.Errorf("key not found")

	}
}