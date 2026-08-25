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
            age int default 18,
			job varchar(50) not null	
        )
    `)
    if err != nil {
        t.Fatalf("failed to execute query: %v", err)
    }
    // Valid insert: all columns provided
    _, err = handler.ExecuteQuery(`
        INSERT INTO test_users (name, email, age, job) VALUES ('alice', 'alice@example.com', 25, 'engineer')
    `)
    if err != nil {
        t.Fatalf("expected valid insert to succeed, got error: %v", err)
    }

    // Valid insert: omit name to use default
    _, err = handler.ExecuteQuery(`
        INSERT INTO test_users (email, age, job) VALUES ('bob@example.com', 30, 'manager')
    `)
    if err != nil {
        t.Fatalf("expected insert with default name to succeed, got error: %v", err)
    }

    // Invalid insert: duplicate email (unique constraint)
    _, err = handler.ExecuteQuery(`
        INSERT INTO test_users (name, email, age, job) VALUES ('alice-dup', 'alice@example.com', 28, 'tester')
    `)
    if err == nil {
        t.Fatalf("expected duplicate-email insert to fail due to unique constraint")
    }

    // Invalid insert: missing NOT NULL column 'job'
    _, err = handler.ExecuteQuery(`
        INSERT INTO test_users (name, email, age) VALUES ('charlie', 'charlie@example.com', 22)
    `)
    if err == nil {
        t.Fatalf("expected insert missing NOT NULL 'job' to fail")
    }

    // Invalid insert: providing value for serial 'id' should be rejected
    _, err = handler.ExecuteQuery(`
        INSERT INTO test_users (id, name, email, age, job) VALUES (1, 'dave', 'dave@example.com', 20, 'dev')
    `)
    if err == nil {
        t.Fatalf("expected insert providing serial 'id' to fail")
    }

}