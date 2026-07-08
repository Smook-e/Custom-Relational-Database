package pages

import (
	"encoding/binary"
)
//LeafPage 
/*
isLeaf 1 byte
nextLeafPage 4 bytes
numberOfEntries 2 bytes
key len(buffer)
pageID 4 bytes
slot 2 bytes
.
.
.
*/
type LeafEntry struct {
    Key    []byte
    PageID uint32
    Slot   uint16
}
//InternalPage
/*
isLeaf 1 byte
numberOfEntries 2 bytes
pageID 4 bytes
key len(buffer)
pageID 4 bytes
.
.
.
*/
type InternalEntry struct {
    Key     []byte
    LeftPtr uint32
	// plus one final RightPtr after all entries
}

const (
	LeafPageHeaderSize     = 1 + 4 + 2 // isLeaf + nextLeafPage + numberOfEntries
	InternalPageHeaderSize = 1 + 2   // isLeaf + numberOfEntries
	EntrySize              = 4 + 2   // pageID + slot
	isLeaf = 1
	isInternal = 0
)

func InitializeLeafPage(entries []LeafEntry, buffer []byte) error {
	if len(entries) == 0 {
		return nil
	}
	offset := 0
	buffer[offset] = uint8(isLeaf)
	offset += 1
	binary.BigEndian.PutUint32(buffer[offset:offset+4], 0) // nextLeafPage is initially 0
	offset += 4

	binary.BigEndian.PutUint16(buffer[offset:offset+2], uint16(len(entries)))
	offset += 2

	for _, entry := range entries {
		binary.BigEndian.PutUint32(buffer[offset:offset+4], entry.PageID)
		offset += 4
		binary.BigEndian.PutUint16(buffer[offset:offset+2], entry.Slot)
		offset += 2
	}
	return nil
}