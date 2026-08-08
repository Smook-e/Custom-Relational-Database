package storage

import (
	"encoding/binary"
	"fmt"

	"github.com/Smook-e/Custom-Relational-Database/entities"
	"github.com/Smook-e/Custom-Relational-Database/pages"
)
const (
	nextLeafPageOffset = 1
	leafPageNumEntriesOffset = 5
	LeafPageHeaderSize     = 1 + 4 + 2 // isLeaf + nextLeafPage + numberOfEntries
	InternalPageHeaderSize = 1 + 2   // isLeaf + numberOfEntries
)
func (engine *StorageEngine) InsertIntoIndex(root uint32, key []byte, pageID uint32, slot uint16, col *entities.Column) (uint32, error) {
	newRoot, newKey, err := IndexInsert(engine, root, key, pageID, slot, col)
	if err != nil {
		
		return 0, fmt.Errorf("failed to insert into index: %w", err)
	}
	if newKey != nil {
		// A split occurred at the root, so we need to create a new root
		newRootPageID, err := engine.NewPage()
		if err != nil {
			return 0, fmt.Errorf("failed to allocate new root page: %w", err)
		}
		newRootBuffer, err := engine.Bp.Get(newRootPageID)
		if err != nil {
			return 0, fmt.Errorf("failed to get buffer for new root page %d: %w", newRootPageID, err)
		}
		internalEntries := []pages.InternalEntry{
			{
				Key:     newKey,
				LeftPtr: root,
			},
		}
		err = pages.InitializeInternalPage(internalEntries, newRootBuffer, newRoot)
		if err != nil {
			return 0, fmt.Errorf("failed to initialize new root page: %w", err)
		}
		engine.Bp.MarkDirty(newRootPageID)
		
		return newRootPageID, nil
	}
	
	return newRoot, nil
}

// Recursive function to insert a key into the index starting from the root page, and split the page if necessary. Returns the new root page ID and the new key to be inserted into the parent page if a split occurs.
func IndexInsert(engine *StorageEngine, root uint32, key []byte, pageID uint32, slot uint16, col *entities.Column) (uint32, []byte, error) {

	buffer , err := engine.Bp.Get(root)
	
	if err != nil {
		return 0, nil, fmt.Errorf("failed to get buffer for page %d: %w", root, err)
	}
	// Check if the page is a leaf or internal page
	offset := 0
	if buffer[offset] == 0 {
		
		// Internal page
		offset += 1
		maxEntries := (4096 - InternalPageHeaderSize - 4) / (len(key) + 4) // 4 bytes for pageID
		numberOfEntries := binary.BigEndian.Uint16(buffer[offset : offset+2])
		offset += 2
		found := -1
		var newRoot uint32
		var newKey []byte
		for i := range numberOfEntries {
			leftptr := binary.BigEndian.Uint32(buffer[offset : offset+4]) // left PageID
			offset += 4
			comp, err := entities.Compare(key, buffer[offset:offset+len(key)], col)
			if err != nil {
				return 0, nil, fmt.Errorf("comparison error: %w", err)
			}
			if comp < 0 {
				// Key is less than the current entry's key, so we should go to the left pointer
				found = int(i)
				newRoot, newKey, err = IndexInsert(engine, leftptr, key, pageID, slot, col)
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
			newRoot, newKey, err = IndexInsert(engine, rightptr, key, pageID, slot, col)
			if err != nil {
				return 0, nil, err
			}
		}
		// Handle the case where a split occurred in the child page
		if newKey != nil {
			if numberOfEntries < uint16(maxEntries) {
				// Insert the new key and pageID into this internal page
				// find the correct position to insert the new key
				if found == int(numberOfEntries) {
					// Insert at the end
					
					offset = InternalPageHeaderSize + int(numberOfEntries)*(len(key) + 4) + 4 // 4 bytes for the right pointer
				} else {
					offset = InternalPageHeaderSize + int(found)*(len(key) + 4) + 4// 4 bytes for the left pointer of the found entry
				}
				
				// Shift entries to make space for the new entry
				dataEnd := InternalPageHeaderSize + int(numberOfEntries)*(len(key) + 4) + 4 // 4 bytes for the right pointer
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
						extractedKey := make([]byte, len(key))
						copy(extractedKey, buffer[offset:offset+len(key)])
						internalEntries[i] = pages.InternalEntry{
							Key:     extractedKey,
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
		}
		// If no split occurred in the child page, just return the current root
		return root, nil, nil
	} else {
	// Leaf page
		offset += 1 + 4 // Skip isLeaf and nextLeafPage
		numberOfEntries := binary.BigEndian.Uint16(buffer[offset : offset+2])
		offset += 2
		maxEntries := (4096 - LeafPageHeaderSize) / (len(key) + 6) // 6 bytes for pageID and slot
		inserted := false
		if numberOfEntries < uint16(maxEntries) {
			// Insert the new key, pageID, and slot into this leaf page
			for range numberOfEntries {
				comp, err := entities.Compare(key, buffer[offset:offset+len(key)], col)
				if err != nil {
					return 0, nil, fmt.Errorf("comparison error: %w", err)
				}
				if comp < 0 {
					inserted = true
					// Shift entries to make space for the new entry
					dataEnd := LeafPageHeaderSize + int(numberOfEntries)*(len(key) + 6)
					copy(buffer[offset + len(key) + 6:dataEnd + len(key) + 6], buffer[offset:dataEnd]) // Shift the rest of the entries
					// Insert the new entry
					copy(buffer[offset:offset+len(key)], key)
					offset += len(key)
					binary.BigEndian.PutUint32(buffer[offset:offset + 4], pageID)
					offset += 4
					binary.BigEndian.PutUint16(buffer[offset:offset + 2], slot)
					offset += 2
					// Update the number of entries
					binary.BigEndian.PutUint16(buffer[1+4:1+4+2], numberOfEntries+1)
					return root, nil, nil
				}else{

					offset += len(key) + 6 // Move to the next entry
				}
			}
			if !inserted {
				// If the new entry is greater than all existing entries, insert it at the end
				copy(buffer[offset:offset+len(key)], key)
				offset += len(key)
				binary.BigEndian.PutUint32(buffer[offset:offset + 4], pageID)
				offset += 4
				binary.BigEndian.PutUint16(buffer[offset:offset + 2], slot)
				// Update the number of entries
				binary.BigEndian.PutUint16(buffer[1+4:1+4+2], numberOfEntries+1)
			}
			engine.Bp.MarkDirty(root)
			return root, nil, nil
		} else {
			// Split this leaf page and return the new root and key to be inserted into the parent
			
			leafEntries := make([]pages.LeafEntry, numberOfEntries+1)
			writeIdx := 0
			offset := LeafPageHeaderSize

			for readIdx := 0; readIdx < int(numberOfEntries); readIdx++ {
				existingKey := buffer[offset : offset+len(key)]

				if !inserted {
					comp, err := entities.Compare(key, existingKey, col)
					if err != nil {
						return 0, nil, fmt.Errorf("comparison error: %w", err)
					}
					if comp < 0 {
						leafEntries[writeIdx] = pages.LeafEntry{Key: key, PageID: pageID, Slot: slot}
						writeIdx++
						inserted = true
					}
				}

				// always read and copy through the existing entry, regardless of whether we just inserted
				extractedKey := make([]byte, len(key))
				copy(extractedKey, existingKey)
				offset += len(key)
				page := binary.BigEndian.Uint32(buffer[offset : offset+4])
				offset += 4
				entrySlot := binary.BigEndian.Uint16(buffer[offset : offset+2])
				offset += 2

				leafEntries[writeIdx] = pages.LeafEntry{Key: extractedKey, PageID: page, Slot: entrySlot}
				writeIdx++
			}

			if !inserted {
				leafEntries[writeIdx] = pages.LeafEntry{Key: key, PageID: pageID, Slot: slot}
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
			// fmt.Println("splitting leaf page", root, "into", rightPage)
			// Initialize the right page
			rightBuffer, err := engine.Bp.Get(rightPage)
			if err != nil {
				return 0, nil, fmt.Errorf("failed to get buffer for new page %d: %w", rightPage, err)
			}
			err = pages.InitializeLeafPage(rightEntries, rightBuffer)
			if err != nil {
				return 0, nil, fmt.Errorf("failed to initialize right page: %w", err)
			}
			// Update the nextLeafPage pointer of the new right page to point to the next leaf page of the current page
			previousNextLeafPage := binary.BigEndian.Uint32(buffer[1:1+4])
			binary.BigEndian.PutUint32(rightBuffer[1:1+4], previousNextLeafPage)
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





func (engine *StorageEngine) IndexSearch(root uint32, key []byte, col *entities.Column) (uint32, uint16, error) {
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
			comp, err := entities.Compare(key, buffer[offset:offset+len(key)], col)
			if err != nil {
				return 0, 0, fmt.Errorf("comparison error: %w", err)
			}
			if comp < 0 {
				// Key is less than the current entry's key, so we should go to the left pointer
				return engine.IndexSearch(leftptr, key, col)
			}
			offset += len(key)
		}
		// If we reach here, it means the key is greater than all entries, so we should go to the right pointer
		rightptr := binary.BigEndian.Uint32(buffer[offset : offset+4])
		return engine.IndexSearch(rightptr, key, col)
	} else {
		// Leaf page
		offset += 1 + 4 // Skip isLeaf and nextLeafPage
		numberOfEntries := binary.BigEndian.Uint16(buffer[offset : offset+2])
		offset += 2
		for range numberOfEntries {
			comp, err := entities.Compare(key, buffer[offset:offset+len(key)], col)
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
				return 0, 0, fmt.Errorf("key not found, %d ", binary.BigEndian.Uint64(key))
			}
			offset += 4 + 2 // Skip pageID and slot
		}
		return 0, 0, fmt.Errorf("key not found")

	}
}