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
	Execute(engine *storage.StorageEngine) (QueryResult, error)
	Parse(p *Parser) ( error)
}
type QueryResult struct {
	result any
	QueryType string
	columns []string
}

func GetQueryType(p *Parser) Query {
	switch p.Peek().Value {
	case "SELECT":
		return &SelectQuery{}
	case "INSERT":
		return &InsertQuery{}
	case "DELETE":
		return &DeleteQuery{}
	case "UPDATE":
		return &UpdateQuery{}
	case "CREATE":
		return &CreateTableQuery{}
	default:
		return nil
	}
}

func (q *CreateTableQuery) Execute(engine *storage.StorageEngine) (QueryResult, error) {
	err :=engine.CreateTable(q.TableName, q.Columns, q.ForeignKeys)
	if err != nil {
		return QueryResult{result: "Failed to create table"}, err
	}
	qr := QueryResult{
		result: fmt.Sprintf("Table %s created successfully", q.TableName),
		QueryType: "CREATE TABLE",
	}
	return qr, nil
}

func (q *SelectQuery) Execute(engine *storage.StorageEngine) (QueryResult, error) {
	result, cols, err := engine.Search(q.TableName,q.Columns, q.Where)
	if err != nil {
		return QueryResult{result: "Failed to execute SELECT query"}, err
	}
	qr := QueryResult{
		result: result,
		QueryType: "SELECT",
		columns: cols,
	}
	return qr, nil
}

func (q *InsertQuery) Execute(engine *storage.StorageEngine) (QueryResult, error) {
	result, err := engine.Insert(q.TableName, q.Columns, q.Values)
	if err != nil {
		return QueryResult{result: "Failed to execute INSERT query"}, err
	}
	qr := QueryResult{
		result: result,
		QueryType: "INSERT",
	}
	return qr, nil
}
func (q *DeleteQuery) Execute(engine *storage.StorageEngine) (QueryResult, error) {
	result, err := engine.Delete(q.TableName, q.Where)
	if err != nil {
		return QueryResult{result: "Failed to execute DELETE query"}, err
	}
	qr := QueryResult{
		result: result,
		QueryType: "DELETE",
	}
	return  qr, nil
} 
func (q *UpdateQuery) Execute(engine *storage.StorageEngine) (QueryResult, error) {
	result, err := engine.Update(q.TableName,  q.Where, q.SetClauses)
	if err != nil {
		return QueryResult{result: "Failed to execute UPDATE query"}, err
	}
	qr := QueryResult{
		result: result,
		QueryType: "UPDATE",
	}
	return  qr, nil
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
func InitializeQueryHandler(fileName string) *QueryHandler {
	engine, err := storage.InitializeStorageEngine(fileName)
	if err != nil {
		fmt.Print(err)
	}
	return &QueryHandler{
		engine: engine,
	}
}