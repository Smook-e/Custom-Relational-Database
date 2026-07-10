package storage

import (
	"fmt"
	"github.com/Smook-e/Custom-Relational-Database/entities"
)

func (engine *StorageEngine) IndexSearch(root uint32, key []byte, typeOfPage uint8) (uint32, uint16, error) {
	buffer , err := engine.Bp.Get(root)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get buffer for page %d: %w", root, err)
	}
	
	return 0, 0, nil
}