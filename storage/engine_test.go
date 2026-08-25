package storage_test

import (
	"testing"
	"github.com/Smook-e/Custom-Relational-Database/storage"
	// "github.com/Smook-e/Custom-Relational-Database/entities"
	"github.com/Smook-e/Custom-Relational-Database/parser"
	"path/filepath"
	// "fmt"
)

func newEmptySQLTestEngine(t *testing.T) *parser.QueryHandler {
    t.Helper()

    dbPath := filepath.Join(t.TempDir(), "sql-test.db")

    engine, err := storage.InitializeStorageEngine(dbPath)
    if err != nil {
        t.Fatalf("failed to initialize storage engine: %v", err)
    }

    return parser.InitializeQueryHandler(engine)
}


func TestEngineInsert(t *testing.T) {
	handler := newEmptySQLTestEngine(t)

    _, err := handler.ExecuteQuery(`
        CREATE TABLE test_users (
            id serial primary key,
            name varchar(50) not null default 'anonymous',
            email varchar(30) not null unique,
            age int default 18
			job varchar(50) not null
			
        )
    `)
    if err != nil {
        t.Fatalf("failed to execute query: %v", err)
    }

}