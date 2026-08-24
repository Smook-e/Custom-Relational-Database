package parser


import (
	"fmt"
	"github.com/Smook-e/Custom-Relational-Database/storage"
)

type SelectQuery struct {
	Columns []string
	TableName string
	Where *storage.Expression
}