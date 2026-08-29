package parser

import (
	"fmt"
	// "strings"
	"github.com/Smook-e/Custom-Relational-Database/storage"
)
type QueryHandler struct {
	engine *storage.StorageEngine
}


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
	case "DELETE":
		return &DeleteQuery{}
	case "CREATE":
		return &CreateTableQuery{}
	default:
		return nil
	}
}

func (q *CreateTableQuery) Execute(engine *storage.StorageEngine) (any, error) {
	return true, engine.CreateTable(q.TableName, q.Columns, q.ForeignKeys)
}
func (q *SelectQuery) Execute(engine *storage.StorageEngine) (any, error) {
	return engine.Search(q.TableName,q.Columns, q.Where)
}

func (q *InsertQuery) Execute(engine *storage.StorageEngine) (any, error) {
	return engine.Insert(q.TableName, q.Columns, q.Values)
}
func (q *DeleteQuery) Execute(engine *storage.StorageEngine) (any, error) {
	return engine.Delete(q.TableName, q.Where)
}

func (Q *QueryHandler) ExecuteQuery(query string) (any, error) {
	tokens, err := Tokenize(query)
	if err != nil {
		return nil, fmt.Errorf("Syntax Error: %w", err)
		
	}
	p := NewParser(tokens)
	queryType := GetQueryType(p)
	if queryType == nil {
		return nil, fmt.Errorf("Error: Unsupported query type")
	}
	err = queryType.Parse(p)
	if err != nil {
		return nil, fmt.Errorf("Syntax Error: %w", err)
	}
	return queryType.Execute(Q.engine)
}
func InitializeQueryHandler(engine *storage.StorageEngine) *QueryHandler {
	return &QueryHandler{
		engine: engine,
	}
}