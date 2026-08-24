package parser

import (
	// "fmt"
	// "strings"
	"github.com/Smook-e/Custom-Relational-Database/storage"
)


type Query interface {
	Execute(engine *storage.StorageEngine) (any, error)
}
