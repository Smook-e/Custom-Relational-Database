package storage

import (
	"fmt"

	"github.com/Smook-e/Custom-Relational-Database/bufferPool"
	"github.com/Smook-e/Custom-Relational-Database/entities"
	"github.com/Smook-e/Custom-Relational-Database/filehandler"
)

const bufferSize = 4096

type StorageEngine struct {
	db *entities.Database
	Bp *bufferpool.BufferPool
}

func (engine *StorageEngine) Commit() error {
	err := engine.Bp.Flush()
	if err != nil {
		return fmt.Errorf("An error occured while commiting to disk: %w", err)
	}
	fmt.Println("Committed Changes to disk successfully")
	return nil
}

func InitializeStorageEngine(filename string) (*StorageEngine, error) {
	engine := &StorageEngine{}
	engine.OpenDatabase(filename)
	engine.Bp = bufferpool.InitializeBufferPool()
	engine.Bp.File = engine.db.File
	return engine, nil
}
func (engine *StorageEngine) NewPage() (uint32, error) {
	newPageID := engine.db.TotalPages
	engine.db.TotalPages++
	buffer := make([]byte, bufferSize)
	filehandler.WriteToFile(engine.db.File, newPageID, buffer)
	engine.Bp.Get(newPageID)
	fmt.Println("created Page :", newPageID)
	return newPageID, nil
}

