package storage

import (
	"fmt"

	"github.com/Smook-e/Custom-Relational-Database/entities"
	"github.com/Smook-e/Custom-Relational-Database/pages"
)

func (engine *StorageEngine) TestIndexPage() {
	pageID, err := engine.NewPage()
	if err != nil {
		fmt.Println("Error creating new page:", err)
		return
	}
	fmt.Println("Created new page with ID:", pageID)
	keys := []int{10,20,30,40,50}
	entries := make([]pages.LeafEntry, len(keys))
	for i, key := range keys {
		key,err := engine.db.Serialize(key, entities.TypeInt)
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
	
}