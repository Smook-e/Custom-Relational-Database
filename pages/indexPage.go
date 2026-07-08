package pages

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



