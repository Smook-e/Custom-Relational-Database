package pages


type LeafEntry struct {
    Key    []byte
    PageID uint32
    Slot   uint16
}

type InternalEntry struct {
    Key     []byte
    LeftPtr uint32
	// plus one final RightPtr after all entries
}