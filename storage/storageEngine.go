package storage

import (
	"fmt"

	"github.com/Smook-e/Custom-Relational-Database/bufferPool"
	"github.com/Smook-e/Custom-Relational-Database/entities"
)

const bufferSize = 4096

type StorageEngine struct {
	db *entities.Database
	bp *bufferpool.BufferPool
}

func (engine *StorageEngine) Commit() error {
	err := engine.bp.Flush()
	if err != nil {
		return fmt.Errorf("An error occured while commiting to disk: %w", err)
	}
	fmt.Println("Committed Changes to disk successfully")
	return nil
}