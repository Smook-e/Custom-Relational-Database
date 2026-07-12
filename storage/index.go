package storage

import (
	"encoding/binary"
	"fmt"

	"github.com/Smook-e/Custom-Relational-Database/entities"
	"github.com/Smook-e/Custom-Relational-Database/pages"
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
		found := -1
		var newRoot uint32
		var newKey []byte
		for i := range numberOfEntries {
			leftptr := binary.BigEndian.Uint32(buffer[offset : offset+4]) // left PageID
			offset += 4
			comp, err := entities.Compare(key, buffer[offset:offset+len(key)], dataType)
			if err != nil {
				return 0, nil, fmt.Errorf("comparison error: %w", err)
			}
			if comp < 0 {
				// Key is less than the current entry's key, so we should go to the left pointer
				found = int(i)
				newRoot, newKey, err = engine.IndexInsert(leftptr, key, pageID, slot, dataType)
				if err != nil {
					return 0, nil, err
				}
				break
			}
			offset += len(key)
		}
		// If we reach here, it means the key is greater than all entries, so we should go to the right pointer
		if found == -1 {
			found = int(numberOfEntries)
			rightptr := binary.BigEndian.Uint32(buffer[offset : offset+4])
			newRoot, newKey, err = engine.IndexInsert(rightptr, key, pageID, slot, dataType)
			if err != nil {
				return 0, nil, err
			}
		}
		// Handle the case where a split occurred in the child page
		if newKey != nil {
			if numberOfEntries < uint16(maxEntries) {
				// Insert the new key and pageID into this internal page
				// find the correct position to insert the new key
				offset = InternalPageHeaderSize + int(found)*(len(key) + 4) - len(key) // move back right after the left pointer of the found entry
				// Shift entries to make space for the new entry
				dataEnd := InternalPageHeaderSize + int(numberOfEntries)*(len(key) + 4)
				copy(buffer[offset + len(newKey) + 4:dataEnd + len(newKey) + 4], buffer[offset:dataEnd]) // Shift the rest of the entries
				// Insert the new entry
				copy(buffer[offset:offset+len(newKey)], newKey)
				offset += len(newKey)
				binary.BigEndian.PutUint32(buffer[offset:offset + 4], newRoot)//right pointer of the new key
				// Update the number of entries
				binary.BigEndian.PutUint16(buffer[1:1+2], numberOfEntries+1)
				engine.Bp.MarkDirty(root)
				return root, nil, nil
			}else {
				// Split this internal page and return the new root and key to be inserted into the parent
				internalEntries := make([]pages.InternalEntry, numberOfEntries + 1)
				//Insert all entries into internalEntries slice
				offset = InternalPageHeaderSize
				for i := 0; i <= int(numberOfEntries) ; i++ {
					leftptr := binary.BigEndian.Uint32(buffer[offset : offset+4]) // left PageID
					offset += 4
					if i == found {
						internalEntries[i] = pages.InternalEntry{
							Key:     newKey,
							LeftPtr: leftptr,
						}
						offset -= 4 // Move back to overwrite the left pointer of the found entry
						binary.BigEndian.PutUint32(buffer[offset:offset + 4], newRoot) // right pointer of the new key
					}else {
						internalEntries[i] = pages.InternalEntry{
							Key:     buffer[offset:offset+len(key)],
							LeftPtr: leftptr,
						}
						offset += len(key)
					}
				}
				rightPtr := binary.BigEndian.Uint32(buffer[offset : offset+4]) // right pointer of the last entry
				// Split the internalEntries into two halves
				mid := len(internalEntries) / 2
				leftEntries := internalEntries[:mid]
				rightEntries := internalEntries[mid+1:]
				midKey := internalEntries[mid].Key
				firstRightPtr := internalEntries[mid].LeftPtr

				// Create a new internal page for the right half
				rightPage, err := engine.NewPage()
				if err != nil {
					return 0, nil, fmt.Errorf("failed to allocate new page: %w", err)
				}
				// Initialize the right page
				rightBuffer, err := engine.Bp.Get(rightPage)
				if err != nil {
					return 0, nil, fmt.Errorf("failed to get buffer for new page %d: %w", rightPage, err)
				}
				err = pages.InitializeInternalPage(rightEntries, rightBuffer, rightPtr)
				if err != nil {
					return 0, nil, fmt.Errorf("failed to initialize right page: %w", err)
				}
				// Initialize the left page (current page) with the left half of the entries
				err = pages.InitializeInternalPage(leftEntries, buffer, firstRightPtr)
				if err != nil {
					return 0, nil, fmt.Errorf("failed to initialize left page: %w", err)
				}
				engine.Bp.MarkDirty(rightPage)
				engine.Bp.MarkDirty(root)
				// Return the right page ID and the midKey to be inserted into the parent internal page
				return rightPage, midKey, nil
			}
	} else {
	// Leaf page
		offset += 1 + 4 // Skip isLeaf and nextLeafPage
		numberOfEntries := binary.BigEndian.Uint16(buffer[offset : offset+2])
		offset += 2
		maxEntries := (4096 - LeafPageHeaderSize) / (len(key) + 6) // 6 bytes for pageID and slot
		if numberOfEntries < uint16(maxEntries) {
			// Insert the new key, pageID, and slot into this leaf page
			for range numberOfEntries {
				comp, err := entities.Compare(key, buffer[offset:offset+len(key)], dataType)
				if err != nil {
					return 0, nil, fmt.Errorf("comparison error: %w", err)
				}
				if comp < 0 {
					// Shift entries to make space for the new entry
					dataEnd := LeafPageHeaderSize + int(numberOfEntries)*(len(key) + 6)
					copy(buffer[offset + len(key) + 6:dataEnd + len(key) + 6], buffer[offset:dataEnd]) // Shift the rest of the entries
					// Insert the new entry
					copy(buffer[offset:offset+len(key)], key)
					offset += len(key)
					binary.BigEndian.PutUint32(buffer[offset:offset + 4], pageID)
					offset += 4
					binary.BigEndian.PutUint16(buffer[offset:offset + 2], slot)
					// Update the number of entries
					binary.BigEndian.PutUint16(buffer[1+4:1+4+2], numberOfEntries+1)
					return root, nil, nil
				}
				offset += len(key) + 6 // Move to the next entry
			}
			engine.Bp.MarkDirty(root)
			return root, nil, nil
		} else {
			// Split this leaf page and return the new root and key to be inserted into the parent
			leafEntries := make([]pages.LeafEntry, numberOfEntries + 1)
			//Insert all entries into leafEntries slice
			for i := 0; i <= int(numberOfEntries) ; i++ {

				comp, err := entities.Compare(key, buffer[offset:offset+len(key)], dataType)
				if err != nil {
					return 0, nil, fmt.Errorf("comparison error: %w", err)
				}
				if comp < 0 {
					// Insert the new entry here
					leafEntries[i] = pages.LeafEntry{
						Key:    key,
						PageID: pageID,
						Slot:   slot,
					}
				}else{
					offset += len(key)
					page := binary.BigEndian.Uint32(buffer[offset : offset+4])
					offset += 4
					slot := binary.BigEndian.Uint16(buffer[offset : offset+2])
					offset += 2
					leafEntries[i] = pages.LeafEntry{
						Key:    buffer[offset-len(key)-6 : offset-6], // Extract the key from the buffer
						PageID: page,
						Slot:   slot,
					}
				}
				
			}
			// Split the leafEntries into two halves
			mid := len(leafEntries) / 2
			leftEntries := leafEntries[:mid]
			rightEntries := leafEntries[mid:]

			// Create a new leaf page for the right half
			rightPage, err := engine.NewPage()
			if err != nil {
				return 0, nil, fmt.Errorf("failed to allocate new page: %w", err)
			}
			// Initialize the right page
			rightBuffer, err := engine.Bp.Get(rightPage)
			if err != nil {
				return 0, nil, fmt.Errorf("failed to get buffer for new page %d: %w", rightPage, err)
			}
			err = pages.InitializeLeafPage(rightEntries, rightBuffer)
			if err != nil {
				return 0, nil, fmt.Errorf("failed to initialize right page: %w", err)
			}
			// Initialize the left page (current page) with the left half of the entries
			err = pages.InitializeLeafPage(leftEntries, buffer)
			if err != nil {
				return 0, nil, fmt.Errorf("failed to initialize left page: %w", err)
			}
			//update the nextLeafPage pointer of the left page to point to the new right page
			binary.BigEndian.PutUint32(buffer[1:1+4], rightPage)
			// Return the right page ID and the first key of the right page to be inserted into the parent internal page
			engine.Bp.MarkDirty(rightPage)
			engine.Bp.MarkDirty(root)
			return rightPage, rightEntries[0].Key, nil
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