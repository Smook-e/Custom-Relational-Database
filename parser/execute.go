package parser

import (
	// "fmt"
	// "strings"
	"github.com/Smook-e/Custom-Relational-Database/storage"
)


type Query interface {
	Execute(engine *storage.StorageEngine) (any, error)
	Parse(p *Parser) ( error)
}
func GetQueryType(p *Parser) Query {
	switch p.Peek().Value {
	case "SELECT":
		return &SelectQuery{}
	case "INSERT":
		return &InsertQuery{}
	default:
		return nil
	}
}

func (q *SelectQuery) Execute(engine *storage.StorageEngine) (any, error) {
	return engine.Search(q.TableName, q.Where)
}

func (q *InsertQuery) Execute(engine *storage.StorageEngine) (any, error) {
	engine.InsertRow(q.Values[0], q.TableName)
	return nil, nil
}