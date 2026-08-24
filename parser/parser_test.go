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
        CREATE TABLE test_users (
            id serial primarykey,
            name varchar(50) notnull,
            email varchar(30) notnull unique,
            age int
        )
    `)
    if err != nil {
        t.Fatalf("CREATE TABLE query failed: %v", err)
    }

    // INSERT via SQL
    _, err = handler.ExecuteQuery(`
        INSERT INTO test_users (name, email, age)
        VALUES ("alice", "alice@example.com", 30)
    `)
    if err != nil {
        t.Fatalf("INSERT query failed: %v", err)
    }

    _, err = handler.ExecuteQuery(`
        INSERT INTO test_users (name, email, age)
        VALUES ("bob", "bob@example.com", 35)
    `)
    if err != nil {
        t.Fatalf("INSERT query failed: %v", err)
    }

    // SELECT via SQL
    result, err := handler.ExecuteQuery(`
        SELECT name, age FROM test_users WHERE age = 30
    `)
    if err != nil {
        t.Fatalf("SELECT query failed: %v", err)
    }

    rows, ok := result.([][]any)
    if !ok {
        t.Fatalf("expected [][]any result, got %T", result)
    }

    if len(rows) != 1 {
        t.Fatalf("expected 1 matching row, got %d", len(rows))
    }

    if len(rows[0]) != 2 {
        t.Fatalf("expected 2 columns, got %d", len(rows[0]))
    }

    if rows[0][0] != "alice" {
        t.Fatalf("expected first row name to be alice, got %v", rows[0][0])
    }

    if rows[0][1] != 30 {
        t.Fatalf("expected first row age to be 30, got %v", rows[0][1])
    }
}
func TestSQL_InsertMultipleRows_ThenSelectAll(t *testing.T) {
    handler := newEmptySQLTestEngine(t)

    _, err := handler.ExecuteQuery(`
        CREATE TABLE test_products (
            id serial primarykey,
            name varchar(50) notnull,
            price int
        )
    `)
    if err != nil {
        t.Fatalf("CREATE TABLE failed: %v", err)
    }

    _, err = handler.ExecuteQuery(`
        INSERT INTO test_products (name, price)
        VALUES ("keyboard", 120)
    `)
    if err != nil {
        t.Fatalf("first INSERT failed: %v", err)
    }

    _, err = handler.ExecuteQuery(`
        INSERT INTO test_products (name, price)
        VALUES ("mouse", 40)
    `)
    if err != nil {
        t.Fatalf("second INSERT failed: %v", err)
    }

    result, err := handler.ExecuteQuery(`
        SELECT name, price
        FROM test_products
        WHERE price > 50
    `)
    if err != nil {
        t.Fatalf("SELECT failed: %v", err)
    }

    rows, ok := result.([][]any)
    if !ok {
        t.Fatalf("expected [][]any result, got %T", result)
    }

    if len(rows) != 1 {
        t.Fatalf("expected 1 matching row, got %d", len(rows))
    }

    if rows[0][0] != "keyboard" {
        t.Fatalf("expected keyboard, got %#v", rows[0][0])
    }
}