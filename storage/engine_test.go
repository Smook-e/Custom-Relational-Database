package storage_test

import (
	"testing"

	"github.com/Smook-e/Custom-Relational-Database/storage"

	// "github.com/Smook-e/Custom-Relational-Database/entities"
	"path/filepath"

	"github.com/Smook-e/Custom-Relational-Database/parser"
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

func TestEngineSelect(t *testing.T) {
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

	// Insert some rows
	inserted, err := handler.ExecuteQuery(`
		INSERT INTO test_users (name, email, age, job) VALUES 
		('alice', 'alice@example.com', 25, 'engineer'),
		('bob', 'bob@example.com', 30, 'manager'),
		('charlie', 'charlie@example.com', 35, 'developer'),
		('dave', 'dave@example.com', 40, 'designer'),
		('eve', 'eve@example.com', 45, 'analyst'),
		('frank', 'frank@example.com', 50, 'consultant')
	`)
	if err != nil {
		t.Fatalf("failed to execute query: %v", err)
	}
	if inserted.(int) != 6 {
		t.Fatalf("expected 6 rows inserted, got %d", inserted.(int))
	}

    // Select specific columns
    res, err := handler.ExecuteQuery(`
        SELECT id, name FROM test_users
    `)
    if err != nil {
        t.Fatalf("failed to execute select query: %v", err)
    }
    rows := res.([][]any)
    if len(rows) != 6 {
        t.Fatalf("expected 6 rows from select, got %d", len(rows))
    }
    // first row should be alice
    if rows[0][1] != "alice" {
        t.Fatalf("expected first name alice, got %v", rows[0][1])
    }

    // Complex where: (id = 2 or age > 18) and (age < 30 or job = 'developer')
    res, err = handler.ExecuteQuery(`
        SELECT id, name FROM test_users WHERE (id = 2 OR age > 18) AND (age < 30 OR job = 'developer')
    `)
    if err != nil {
        t.Fatalf("failed to execute complex where select: %v", err)
    }
    rows = res.([][]any)
    // Expect alice (id 1) and charlie (id 3)
    if len(rows) != 2 {
        t.Fatalf("expected 2 rows from complex where, got %d", len(rows))
    }
    found := map[string]bool{}
    for _, r := range rows {
        name := r[1].(string)
        found[name] = true
    }
    if !found["alice"] || !found["charlie"] {
        t.Fatalf("expected alice and charlie in complex where results, got %v", found)
    }

    // Or combination on job
    res, err = handler.ExecuteQuery(`
        SELECT id, name, job FROM test_users WHERE job = 'manager' OR job = 'designer'
    `)
    if err != nil {
        t.Fatalf("failed to execute job OR select: %v", err)
    }
    rows = res.([][]any)
    if len(rows) != 2 {
        t.Fatalf("expected 2 rows for job filter, got %d", len(rows))
    }

    // Age greater than 40
    res, err = handler.ExecuteQuery(`
        SELECT id, name, age FROM test_users WHERE age > 40
    `)
    if err != nil {
        t.Fatalf("failed to execute age filter select: %v", err)
    }
    rows = res.([][]any)
    if len(rows) != 2 {
        t.Fatalf("expected 2 rows for age > 40, got %d", len(rows))
    }

}