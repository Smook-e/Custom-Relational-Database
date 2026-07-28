package storage

import (
	"fmt"
	

	"encoding/binary"

	"github.com/Smook-e/Custom-Relational-Database/entities"
	"github.com/Smook-e/Custom-Relational-Database/pages"
)
func ( engine *StorageEngine) TestSearch(RootPageID uint32) {
	for i := range 10000 {
		key, err := engine.db.Serialize(int64(i), entities.TypeBigInt)
		if err != nil {
			fmt.Println("Error serializing key:", err)
			return
		}
		_, _, err = engine.IndexSearch(RootPageID, key, entities.TypeBigInt)
		if err != nil {
			fmt.Println("Error searching for key:", i, "Error:", err)
			continue
		}
		
	}
}
func (engine *StorageEngine) TestIndexSearchPageRoot() {
	pageID, err := engine.NewPage()
	if err != nil {
		fmt.Println("Error creating new page:", err)
		return
	}
	buffer, err := engine.Bp.Get(pageID)
	if err != nil {
		fmt.Println("Error retrieving page from buffer pool:", err)
		return
	}
	// create first leaf page
	leaf1_keys := []int32{10, 20, 30}
	leaf1_entries := make([]pages.LeafEntry, len(leaf1_keys))
	for i, val := range leaf1_keys {
		key, err := engine.db.Serialize(val, entities.TypeInt)
		if err != nil {
			fmt.Println("Error serializing key:", err)
			return
		}
		leaf1_entries[i] = pages.LeafEntry{
			Key:    key,
			PageID: uint32(i),
			Slot:   uint16(i),
		}
	}
	pages.InitializeLeafPage(leaf1_entries, buffer)
	fmt.Println("Initialized first leaf page with entries:", leaf1_keys)
	// create second leaf page
	pageID2, err := engine.NewPage()
	if err != nil {
		fmt.Println("Error creating new page:", err)
		return
	}
	buffer2, err := engine.Bp.Get(pageID2)
	if err != nil {
		fmt.Println("Error retrieving page from buffer pool:", err)
		return
	}
	leaf2_keys := []int32{40, 50, 60}
	leaf2_entries := make([]pages.LeafEntry, len(leaf2_keys))
	for i, val := range leaf2_keys {
		key, err := engine.db.Serialize(val, entities.TypeInt)
		if err != nil {
			fmt.Println("Error serializing key:", err)
			return
		}
		leaf2_entries[i] = pages.LeafEntry{
			Key:    key,
			PageID: uint32(i + len(leaf1_keys)), // continue page IDs
			Slot:   uint16(i),
		}
	}
	pages.InitializeLeafPage(leaf2_entries, buffer2)
	fmt.Println("Initialized second leaf page with entries:", leaf2_keys)
	// create root internal page
	pageID3, err := engine.NewPage()
	if err != nil {
		fmt.Println("Error creating new page:", err)
		return
	}
	buffer3, err := engine.Bp.Get(pageID3)
	if err != nil {
		fmt.Println("Error retrieving page from buffer pool:", err)
		return
	}
	root_entries := []pages.InternalEntry{
		{
			Key: leaf2_entries[0].Key, // first key of second leaf page
			LeftPtr: pageID,           // first leaf page
		},
	}
	
	pages.InitializeInternalPage(root_entries, buffer3, pageID2) // right pointer to second leaf page
	fmt.Println("Initialized root internal page with entries:", root_entries)
	// search for keys in the index
	for _, val := range append(leaf1_keys, leaf2_keys...) {
		key, err := engine.db.Serialize(val, entities.TypeInt)
		if err != nil {
			fmt.Println("Error serializing key:", err)
			return
		}
		pageID, slot, err := engine.IndexSearch(pageID3, key, entities.TypeInt)
		if err != nil {
			fmt.Println("Error searching for key:", val, "Error:", err)
			continue
		}
		fmt.Printf("Found key %d at PageID: %d, Slot: %d\n", val, pageID, slot)
	}
}
func (engine *StorageEngine) TestIndexSearchPage() {
	pageID, err := engine.NewPage()
	if err != nil {
		fmt.Println("Error creating new page:", err)
		return
	}
	buffer, err := engine.Bp.Get(pageID)
	if err != nil {
		fmt.Println("Error retrieving page from buffer pool:", err)
		return
	}
	fmt.Println("Created new page with ID:", pageID)
	keys := []int32{10,20,30,40,50}
	entries := make([]pages.LeafEntry, len(keys))
	for i, val := range keys {
		key,err := engine.db.Serialize(val, entities.TypeInt)
		if err != nil {
			fmt.Println("Error serializing key:", err)
			return
		}
		entries[i] = pages.LeafEntry{
			Key:    key,
			PageID: uint32(i),
			Slot:   uint16(i),
		}
	}
	pages.InitializeLeafPage(entries, buffer)
	fmt.Println("Initialized leaf page with entries:", keys)
	for _, val := range keys {
		key,err := engine.db.Serialize(val, entities.TypeInt)
		if err != nil {
			fmt.Println("Error serializing key:", err)
			return
		}
		pageID, slot, err := engine.IndexSearch(pageID, key, entities.TypeInt)
		if err != nil {
			fmt.Println("Error searching for key:", val, "Error:", err)
			continue
		}
		fmt.Printf("Found key %d at PageID: %d, Slot: %d\n", val, pageID, slot)
	}
	
}
func (engine *StorageEngine) TestIndexInsertMiddleRoot(rootPageID uint32){
	// create a new page to act as the root of the index
	var root uint32
	var err error
	if rootPageID == 0 {
		root, err = engine.NewPage()
		if err != nil {
			fmt.Println("Error creating new page:", err)
			return
		}
	} else {
		root = rootPageID
	}
	buffer, err := engine.Bp.Get(root)
	if err != nil {
		fmt.Println("Error retrieving page from buffer pool:", err)
		return
	}
	err =pages.InitializeLeafPage([]pages.LeafEntry{

	}, buffer)
	if err != nil {
		fmt.Println("Error initializing leaf page:", err)
		return
	}

	for i:= 0; i <= 100000; i += 2 {
		key, err := engine.db.Serialize(int64(i), entities.TypeBigInt)
		if err != nil {
			fmt.Println("Error serializing key:", err)
			return
		}
		root, err = engine.InsertIntoIndex(root, key, uint32(i), uint16(i), entities.TypeBigInt)
		
		if err != nil {
			fmt.Println("Error inserting key:", i, "Error:", err)
			return
		}
		if  i % 10000 == 0 {
			fmt.Print( i, " ")
		}
		if i % 20000 == 0 {
			fmt.Println(root)
		}
		
	}
	for i:= 1; i <= 100000; i += 2 {
		key, err := engine.db.Serialize(int64(i), entities.TypeBigInt)
		if err != nil {
			fmt.Println("Error serializing key:", err)
			return
		}
		root, err = engine.InsertIntoIndex(root, key, uint32(i), uint16(i), entities.TypeBigInt)
		
		if err != nil {
			fmt.Println("Error inserting key:", i, "Error:", err)
			return
		}
		if  i % 1001 == 0 {
			fmt.Print( i, " ")
		}
		if i % 10001 == 0 {
			fmt.Println("root:", root)
		}
		
	}
	//search
	for i := range 100000 {
		key, err := engine.db.Serialize(int64(i), entities.TypeBigInt)
		if err != nil {
			fmt.Println("Error serializing key:", err)
			return
		}
		_, _, err = engine.IndexSearch(root, key, entities.TypeBigInt)
		if err != nil {
			fmt.Println("Error searching for key:", i, "Error:", err)
			continue
		}
		
	}
	// buffer, _ = engine.Bp.Get(3)
	// PrintLeafPageEntries(buffer, 8, entities.TypeBigInt)
}
func PrintLeafPageEntries(buffer []byte,keySize int, dataType uint8) {
	if buffer[0] != uint8(pages.IsLeaf) {
		fmt.Println("Not a leaf page")
		return
	}
	nextLeafPage := binary.BigEndian.Uint32(buffer[1:5])
	fmt.Printf("Next Leaf Page ID: %d\n", nextLeafPage)
	numberOfEntries := binary.BigEndian.Uint16(buffer[5:7])
	offset := 7
	fmt.Printf("Leaf Page Entries (Total: %d):\n", numberOfEntries)
	for i := 0; i < int(numberOfEntries); i++ {
		keyData := buffer[offset : offset+keySize]
		key, err := entities.Deserialize(keyData, dataType)
		if err != nil {
			fmt.Println("Error deserializing key:", err)
			return
		}
		fmt.Printf("  Key %d: %v\n", i, key)
		offset += keySize + 6 // Move to the next entry
	}
}

func (engine *StorageEngine) TestIndexInsertRoot(rootPageID uint32) {
	// create a new page to act as the root of the index
	var root uint32
	var err error
	if rootPageID == 0 {
		root, err = engine.NewPage()
		if err != nil {
			fmt.Println("Error creating new page:", err)
			return
		}
	} else {
		root = rootPageID
	}
	buffer, err := engine.Bp.Get(root)
	if err != nil {
		fmt.Println("Error retrieving page from buffer pool:", err)
		return
	}
	err =pages.InitializeLeafPage([]pages.LeafEntry{

	}, buffer)
	if err != nil {
		fmt.Println("Error initializing leaf page:", err)
		return
	}

	for i := range 1000001 {
		key, err := engine.db.Serialize(int64(i), entities.TypeBigInt)
		if err != nil {
			fmt.Println("Error serializing key:", err)
			return
		}
		
		root, err = engine.InsertIntoIndex(root, key, uint32(i), uint16(i), entities.TypeBigInt)
		
		if err != nil {
			fmt.Println("Error inserting key:", i, "Error:", err)
			return
		}
		
		if i% 100000 == 0 {
			fmt.Println("key:", i, "root:", root)
		}
		
	}
	// page, err := engine.Bp.Get(541)
	// if err != nil {
	// 	fmt.Println("Error retrieving page from buffer pool:", err)
	// 	return
	// }
	// PrintLeafPageEntries(page, 8, entities.TypeBigInt)
	// search
	engine.Commit()
	for i := 0; i <= 1000000; i++ {
		key, err := engine.db.Serialize(int64(i), entities.TypeBigInt)
		if err != nil {
			fmt.Println("Error serializing key:", err)
			return
		}
		_, _, err = engine.IndexSearch(root, key, entities.TypeBigInt)
		if err != nil {
			fmt.Println("Error searching for key:", i, "Error:", err)
			continue
		}
		
	}
}


func (engine *StorageEngine) TestIndexInsert() {
	// create a new page to act as the root of the index
	pageID, err := engine.NewPage()
	if err != nil {
		fmt.Println("Error creating new page:", err)
		return
	}
	buffer, err := engine.Bp.Get(pageID)
	if err != nil {
		fmt.Println("Error retrieving page from buffer pool:", err)
		return
	}
	err =pages.InitializeLeafPage([]pages.LeafEntry{

	}, buffer)
	if err != nil {
		fmt.Println("Error initializing leaf page:", err)
		return
	}
	for i := 1; i <= 10; i++ {
		key, err := engine.db.Serialize(int32(i), entities.TypeInt)
		if err != nil {
			fmt.Println("Error serializing key:", err)
			return
		}
		_, err = engine.InsertIntoIndex(pageID, key, uint32(i), uint16(i), entities.TypeInt)
		
		if err != nil {
			fmt.Println("Error inserting key:", i, "Error:", err)
			return
		}
		fmt.Printf("Inserted key %d at PageID: %d\n", i, pageID)
	}
	// search for keys in the index
	for i := 1; i <= 10; i++ {
		key, err := engine.db.Serialize(int32(i), entities.TypeInt)
		if err != nil {
			fmt.Println("Error serializing key:", err)
			return
		}
		pageID, slot, err := engine.IndexSearch(pageID, key, entities.TypeInt)
		if err != nil {
			fmt.Println("Error searching for key:", i, "Error:", err)
			continue
		}
		fmt.Printf("Found key %d at PageID: %d, Slot: %d\n", i, pageID, slot)
	}
}