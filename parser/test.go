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
