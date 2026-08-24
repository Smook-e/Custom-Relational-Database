package parser

import (
    "path/filepath"
    "testing"

    "github.com/Smook-e/Custom-Relational-Database/storage"
)

func newEmptySQLTestEngine(t *testing.T) *QueryHandler {
    t.Helper()

    dbPath := filepath.Join(t.TempDir(), "sql-test.db")

    engine, err := storage.InitializeStorageEngine(dbPath)
    if err != nil {
        t.Fatalf("failed to initialize storage engine: %v", err)
    }

    return InitializeQueryHandler(engine)
}

func TestSQL_CreateTable_ThenInsert_ThenSelect(t *testing.T) {
    handler := newEmptySQLTestEngine(t)

    _, err := handler.ExecuteQuery(`
        CREATE TABLE users (
            id serial primarykey,
            name varchar(50) notnull,
            email varchar(100),
            age int
        )
    `)
    if err != nil {
        t.Fatalf("CREATE TABLE via SQL failed: %v", err)
    }

    _, err = handler.ExecuteQuery(`
        INSERT INTO users (name, email, age)
        VALUES ("alice", "alice@example.com", 30)
    `)
    if err != nil {
        t.Fatalf("INSERT via SQL failed: %v", err)
    }

    result, err := handler.ExecuteQuery(`
        SELECT name, age
        FROM users
        WHERE age = 30
    `)
    if err != nil {
        t.Fatalf("SELECT via SQL failed: %v", err)
    }

    rows, ok := result.([][]any)
    if !ok {
        t.Fatalf("expected [][]any result, got %T", result)
    }

    if len(rows) != 1 {
        t.Fatalf("expected 1 row, got %d", len(rows))
    }

    if rows[0][0] != "alice" {
        t.Fatalf("expected alice, got %#v", rows[0][0])
    }

    if rows[0][1] != 30 {
        t.Fatalf("expected age 30, got %#v", rows[0][1])
    }
}