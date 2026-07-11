package storage

import (
	"fmt"

	"github.com/Smook-e/Custom-Relational-Database/entities"
	"github.com/Smook-e/Custom-Relational-Database/pages"
)
func (engine *StorageEngine) TestIndexPageRoot() {
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
}
func (engine *StorageEngine) TestIndexPage() {
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