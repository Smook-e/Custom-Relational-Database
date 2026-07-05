package storage

import (
	"github.com/Smook-e/Custom-Relational-Database/entities"
)

const bufferSize = 4096

type StorageEngine struct {
	db *entities.Database

}

