package storage

import (
	"github.com/Smook-e/Custom-Relational-Database/entities"
	"github.com/Smook-e/Custom-Relational-Database/bufferPool"
)

const bufferSize = 4096

type StorageEngine struct {
	db *entities.Database
	bp *bufferpool.BufferPool
}

