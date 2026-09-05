package storage_test

import (

	"testing"

	"path/filepath"

	"github.com/Smook-e/Custom-Relational-Database/parser"
)

func newEmptySQLTestEngine(t *testing.T) *parser.QueryHandler {
    t.Helper()

    dbPath := filepath.Join(t.TempDir(), "sql-test.db")
    qh , err := parser.InitializeQueryHandler(dbPath)
    if err != nil {
        t.Fatalf("failed to initialize query handler: %v", err)
    }
    return qh
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
        );
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

    // Invalid insert: name exceeds varchar(50)
    longName := ""
    for i := 0; i < 60; i++ { longName += "x" }
    _, err = handler.ExecuteQuery(`
        INSERT INTO test_users (name, email, age, job) VALUES ('` + longName + `', 'long@example.com', 29, 'tester')
    `)
    if err == nil {
        t.Fatalf("expected insert with too-long name to fail")
    }

    // Invalid insert: non-numeric value into int column
    _, err = handler.ExecuteQuery(`
        INSERT INTO test_users (name, email, age, job) VALUES ('erin', 'erin@example.com', 'twenty', 'intern')
    `)
    if err == nil {
        t.Fatalf("expected insert with non-numeric age to fail")
    }

    // Invalid insert: empty string for NOT NULL 'job' (treated as null)
    _, err = handler.ExecuteQuery(`
        INSERT INTO test_users (name, email, age, job) VALUES ('fred', 'fred@example.com', 28, '')
    `)
    if err == nil {
        t.Fatalf("expected insert with empty NOT NULL 'job' to fail")
    }

    // Valid insert: provide empty name (should use default 'anonymous')
    _, err = handler.ExecuteQuery(`
        INSERT INTO test_users (email, age, job) VALUES ( 'gina@example.com', 27, 'analyst')
    `)
    if err != nil {
        t.Fatalf("expected insert with empty name to use default and succeed, got error: %v", err)
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
	if inserted.Result.(int) != 6 {
		t.Fatalf("expected 6 rows inserted, got %d", inserted.Result.(int))
	}

    // Select specific columns
    res, err := handler.ExecuteQuery(`
        SELECT id, name FROM test_users
    `)
    if err != nil {
        t.Fatalf("failed to execute select query: %v", err)
    }
    rows := res.Result.([][]any)
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
    rows = res.Result.([][]any)
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
    rows = res.Result.([][]any)
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
    rows = res.Result.([][]any)
    if len(rows) != 2 {
        t.Fatalf("expected 2 rows for age > 40, got %d", len(rows))
    }

}
func TestEngineDelete(t *testing.T) {
    handler := newEmptySQLTestEngine(t)

    _, err := handler.ExecuteQuery(`
        CREATE TABLE test_users (
            id serial primary key,
            name varchar(50) not null default 'anonymous',
            age int default 18,
			job varchar(50),
            email varchar(30) unique	
        )
    `)
    if err != nil {
        t.Fatalf("failed to execute query: %v", err)
    }

    // Insert some test data
    _, err = handler.ExecuteQuery(`
        INSERT INTO test_users ( name, age, job, email) VALUES
            ('alice', 25, 'developer', 'alice@example.com'),
            ('bob', 30, 'manager', 'bob@example.com'),
            ('charlie', 35, 'designer', 'charlie@example.com'),
            ('dave', 40, 'analyst', 'dave@example.com'),
            ('eve', 45, 'consultant', ''),
            ('frank', 50, 'engineer', '') 
    `)
    if err != nil {
        t.Fatalf("failed to insert test data: %v", err)
    }

    // Delete a specific user
    _, err = handler.ExecuteQuery(`
        DELETE FROM test_users WHERE id = 2
    `)
    if err != nil {
        t.Fatalf("failed to execute delete query: %v", err)
    }

    // Verify the user was deleted
    res, err := handler.ExecuteQuery(`
        SELECT id, name FROM test_users
    `)
    if err != nil {
        t.Fatalf("failed to execute select query: %v", err)
    }
    rows := res.Result.([][]any)
    if len(rows) != 5 {
        t.Fatalf("expected 5 rows after delete, got %d", len(rows))
    }
    // Ensure that the deleted user (id=2) is not present
    for _, r := range rows {
        if r[0].(int32) == 2 {
            t.Fatalf("expected user with id=2 to be deleted, but found in results")
        }
    }
    // Attempt to delete a non-existent user (should not error, but no rows affected)
    rowsDeleted, err := handler.ExecuteQuery(`
        DELETE FROM test_users WHERE id = 999
    `)
    if err != nil {
        t.Fatalf("failed to execute delete query for non-existent user: %v", err)
    }
    if rowsDeleted.Result.(int) != 0 {
        t.Fatalf("expected 0 rows deleted for non-existent user, got %d", rowsDeleted.Result.(int))
    }

    //Delete multiple users with a condition
    _, err = handler.ExecuteQuery(`
        DELETE FROM test_users WHERE age >= 40
    `)
    if err != nil {
        t.Fatalf("failed to execute delete query for age > 40: %v", err)
    }

    // Verify the users with age > 40 were deleted
    res, err = handler.ExecuteQuery(`
        SELECT id, name, age FROM test_users
    `)
    if err != nil {
        t.Fatalf("failed to execute select query after delete: %v", err)
    }
    rows = res.Result.([][]any)
    for _, r := range rows {
        if r[2].(int32) >= 40 {
            t.Fatalf("expected no users with age >= 40 after delete, but found user with age %d", r[2].(int32))
        }
    }
    // ensure that the remaining users are correct
    expectedRemaining := map[int32]string{
        1: "alice",
        3: "charlie",
    }
    if len(rows) != len(expectedRemaining) {
        t.Fatalf("expected %d remaining users, got %d", len(expectedRemaining), len(rows))
    }
    for _, r := range rows {
        id := r[0].(int32)
        name := r[1].(string)
        if expectedName, ok := expectedRemaining[id]; !ok || expectedName != name {
            t.Fatalf("unexpected remaining user: id=%d, name=%s", id, name)
        }
    }

    // Delete all remaining users
    _, err = handler.ExecuteQuery(`
        DELETE FROM test_users
    `)
    if err != nil {
        t.Fatalf("failed to execute delete all query: %v", err)
    }

    // Verify that the table is now empty
    res, err = handler.ExecuteQuery(`
        SELECT id, name FROM test_users
    `)
    if err != nil {
        t.Fatalf("failed to execute select query after deleting all: %v", err)
    }
    rows = res.Result.([][]any)
    if len(rows) != 0 {
        t.Fatalf("expected 0 rows after deleting all users, got %d", len(rows))
    }

    // Attempt to delete from an empty table (should not error, but no rows affected)
    rowsDeleted, err = handler.ExecuteQuery(`
        DELETE FROM test_users
    `)
    if err != nil {
        t.Fatalf("failed to execute delete on empty table: %v", err)
    }
    if rowsDeleted.Result.(int) != 0 {
        t.Fatalf("expected 0 rows deleted from empty table, got %d", rowsDeleted.Result.(int))
    }
    // Attempt to insert a user after deletion to ensure table is still functional
    _, err = handler.ExecuteQuery(`
        INSERT INTO test_users (name, age, job, email) VALUES ('grace', 28, 'analyst', 'grace@example.com')
    `)
    if err != nil {
        t.Fatalf("failed to insert after deleting all users: %v", err)
    }
    // Verify the new user was inserted
    res, err = handler.ExecuteQuery(`
        SELECT id, name FROM test_users WHERE name = 'grace'
    `)
    if err != nil {
        t.Fatalf("failed to select newly inserted user: %v", err)
    }
    rows = res.Result.([][]any)
    if len(rows) != 1 || rows[0][1].(string) != "grace" {
        t.Fatalf("expected to find newly inserted user 'grace', but did not")
    }

    // Attempt to insert a user with a duplicate email to ensure unique constraint is still enforced
    _, err = handler.ExecuteQuery(`
        INSERT INTO test_users (name, age, job, email) VALUES ('hank', 32, 'developer', 'grace@example.com')
    `)
    if err == nil {
        t.Fatalf("expected insert with duplicate email to fail, but it succeeded")
    }
    // Delete the newly inserted user to clean up
    _, err = handler.ExecuteQuery(`
        DELETE FROM test_users WHERE name = 'grace'
    `)
    if err != nil {
        t.Fatalf("failed to delete newly inserted user 'grace': %v", err)
    }
    // try to insert a user with the same email again 
    _, err = handler.ExecuteQuery(`
        INSERT INTO test_users (name, age, job, email) VALUES ('hank', 32, 'developer', 'grace@example.com')
    `)
    if err != nil {
        t.Fatalf("expected insert with previously used email to succeed after deletion, but got error: %v", err)
    }

}

func TestEngineUpdate(t *testing.T) {
    handler := newEmptySQLTestEngine(t)

    _, err := handler.ExecuteQuery(`
        CREATE TABLE test_users (
            id serial primary key,
            name varchar(50) not null default 'anonymous',
            age int default 18,
            job varchar(50) not null,
            email varchar(30) unique
        )
    `)
    if err != nil {
        t.Fatalf("failed to create test table: %v", err)
    }

    _, err = handler.ExecuteQuery(`
        INSERT INTO test_users (name, email, age, job) VALUES
            ('alice', 'alice@example.com', 25, 'engineer'),
            ('bob', 'bob@example.com', 30, 'manager'),
            ('charlie', 'charlie@example.com', 35, 'designer'),
            ('dana', 'dana@example.com', 40, 'analyst')
    `)
    if err != nil {
        t.Fatalf("failed to insert seed rows: %v", err)
    }

    // Valid update: one row, one column.
    rowsUpdated, err := handler.ExecuteQuery(`
        UPDATE test_users SET age = 26 WHERE id = 1
    `)
    if err != nil {
        t.Fatalf("expected valid single-row update to succeed, got error: %v", err)
    }
    if rowsUpdated.Result.(int) != 1 {
        t.Fatalf("expected 1 row updated, got %d", rowsUpdated.Result.(int))
    }

    res, err := handler.ExecuteQuery(`
        SELECT id, age FROM test_users WHERE id = 1
    `)
    if err != nil {
        t.Fatalf("failed to verify updated row: %v", err)
    }
    rows := res.Result.([][]any)
    if len(rows) != 1 || rows[0][1].(int32) != 26 {
        t.Fatalf("expected id=1 age to be 26 after update, got %#v", rows)
    }

    // Valid update: multiple columns and a complex WHERE clause.
    rowsUpdated, err = handler.ExecuteQuery(`
        UPDATE test_users SET name = 'alice-smith', job = 'lead engineer' WHERE (id = 1 OR id = 3) AND age >= 25
    `)
    if err != nil {
        t.Fatalf("expected multi-column update to succeed, got error: %v", err)
    }
    if rowsUpdated.Result.(int) != 2 {
        t.Fatalf("expected 2 rows updated by complex condition, got %d", rowsUpdated.Result.(int))
    }

    res, err = handler.ExecuteQuery(`
        SELECT id, name, job FROM test_users WHERE id = 1 OR id = 3 
    `)
    if err != nil {
        t.Fatalf("failed to read updated rows: %v", err)
    }
    rows = res.Result.([][]any)
    if len(rows) != 2 {
        t.Fatalf("expected 2 matching updated rows, got %d", len(rows))
    }
    found := map[int32]map[string]string{}
    for _, row := range rows {
        id := row[0].(int32)
        found[id] = map[string]string{
            "name": row[1].(string),
            "job": row[2].(string),
        }
    }
    if found[1]["name"] != "alice-smith" || found[1]["job"] != "lead engineer" {
        t.Fatalf("expected row 1 updated to alice-smith / lead engineer, got %#v", found[1])
    }
    if found[3]["name"] != "alice-smith" || found[3]["job"] != "lead engineer" {
        t.Fatalf("expected row 3 to also be updated by the matching OR condition, got %#v", found[3])
    }

    // Update multiple rows matching a simple condition.
    rowsUpdated, err = handler.ExecuteQuery(`
        UPDATE test_users SET job = 'senior' WHERE age >= 30
    `)
    if err != nil {
        t.Fatalf("expected bulk update to succeed, got error: %v", err)
    }
    if rowsUpdated.Result.(int) != 3 {
        t.Fatalf("expected 3 rows updated for age >= 30, got %d", rowsUpdated.Result.(int))
    }

    res, err = handler.ExecuteQuery(`
        SELECT id, job FROM test_users WHERE age >= 30
    `)
    if err != nil {
        t.Fatalf("failed to verify bulk updated rows: %v", err)
    }
    rows = res.Result.([][]any)
    if len(rows) != 3 {
        t.Fatalf("expected 3 rows with age >= 30, got %d", len(rows))
    }
    for _, row := range rows {
        if row[1].(string) != "senior" {
            t.Fatalf("expected all matching rows to have job='senior', got row %#v", row)
        }
    }

    // No rows match: update should be a no-op and return 0.
    rowsUpdated, err = handler.ExecuteQuery(`
        UPDATE test_users SET age = 99 WHERE id = 999
    `)
    if err != nil {
        t.Fatalf("expected no-op update to succeed, got error: %v", err)
    }
    if rowsUpdated.Result.(int) != 0 {
        t.Fatalf("expected 0 rows updated for missing id, got %d", rowsUpdated.Result.(int))
    }

    // Unique constraint violation should fail.
    _, err = handler.ExecuteQuery(`
        UPDATE test_users SET email = 'alice@example.com' WHERE id = 2
    `)
    if err == nil {
        t.Fatalf("expected duplicate-email update to fail due to unique constraint")
    }

    // Invalid type should fail.
    _, err = handler.ExecuteQuery(`
        UPDATE test_users SET age = 'not-a-number' WHERE id = 1
    `)
    if err == nil {
        t.Fatalf("expected non-numeric age update to fail")
    }

    // Invalid varchar length should fail.
    longName := ""
    for i := 0; i < 60; i++ { longName += "x" }
    _, err = handler.ExecuteQuery(`
        UPDATE test_users SET name = '` + longName + `' WHERE id = 1
    `)
    if err == nil {
        t.Fatalf("expected oversized name update to fail")
    }

    // Serial primary key cannot be manually updated.
    _, err = handler.ExecuteQuery(`
        UPDATE test_users SET id = 10 WHERE id = 1
    `)
    if err == nil {
        t.Fatalf("expected serial primary key update to fail")
    }

    // Set a column to a duplicate value for a unique field should fail even when row itself is being updated.
    _, err = handler.ExecuteQuery(`
        UPDATE test_users SET email = 'dana@example.com' WHERE id = 1
    `)
    if err == nil {
        t.Fatalf("expected update to duplicate existing unique email to fail")
    }

    // Valid multi-column update should apply all assignments together.
    rowsUpdated, err = handler.ExecuteQuery(`
        UPDATE test_users SET age = 101, job = 'director' WHERE id = 2
    `)
    if err != nil {
        t.Fatalf("expected valid multi-column update to succeed, got error: %v", err)
    }
    if rowsUpdated.Result.(int) != 1 {
        t.Fatalf("expected 1 row updated by valid multi-column update, got %d", rowsUpdated.Result.(int))
    }

    res, err = handler.ExecuteQuery(`
        SELECT id, age, job FROM test_users WHERE id = 2
    `)
    if err != nil {
        t.Fatalf("failed to verify valid multi-column update: %v", err)
    }
    rows = res.Result.([][]any)
    if len(rows) != 1 || rows[0][1].(int32) != 101 || rows[0][2].(string) != "director" {
        t.Fatalf("expected row 2 to be updated to age=101, job='director', got %#v", rows)
    }
}

// createTable is a small helper that executes a CREATE TABLE SQL and returns the error (if any).
func createTable(handler *parser.QueryHandler, t *testing.T, sql string) error {
    _, err := handler.ExecuteQuery(sql)
    return err
}

func TestCreateTable(t *testing.T) {
    handler := newEmptySQLTestEngine(t)

    // Valid table creation should succeed
    err := createTable(handler, t, `
        CREATE TABLE ct_valid (
            id serial primary key,
            name varchar(50) not null,
            val int
        )
    `)
    if err != nil {
        t.Fatalf("expected valid create table to succeed, got: %v", err)
    }

    // Missing primary key should fail
    err = createTable(handler, t, `
        CREATE TABLE ct_no_pk (
            id int,
            name varchar(10)
        )
    `)
    if err == nil {
        t.Fatalf("expected create without primary key to fail")
    }

    // Two primary keys should fail (parser or storage should reject)
    err = createTable(handler, t, `
        CREATE TABLE ct_two_pk (
            a int primary key,
            b int primary key,
            c varchar(10)
        )
    `)
    if err == nil {
        t.Fatalf("expected create with two primary keys to fail")
    }

    // Duplicate column names should fail
    err = createTable(handler, t, `
        CREATE TABLE ct_dup_col (
            id serial primary key,
            name varchar(10),
            name varchar(20)
        )
    `)
    if err == nil {
        t.Fatalf("expected create with duplicate column names to fail")
    }

    // Invalid data type should fail
    err = createTable(handler, t, `
        CREATE TABLE ct_bad_type (
            id serial primary key,
            foo spamtype
        )
    `)
    if err == nil {
        t.Fatalf("expected create with invalid data type to fail")
    }

    // Varchar length too large should fail (e.g., varchar(300))
    err = createTable(handler, t, `
        CREATE TABLE ct_varchar_large (
            id serial primary key,
            big varchar(300)
        )
    `)
    if err == nil {
        t.Fatalf("expected create with oversized varchar to fail")
    }

    // Foreign key referencing non-existent table should fail
    err = createTable(handler, t, `
        CREATE TABLE ct_fk_bad_ref (
            id serial primary key,
            other_id int,
            FOREIGN KEY (other_id) REFERENCES no_table(id)
        )
    `)
    if err == nil {
        t.Fatalf("expected create with foreign key referencing non-existent table to fail")
    }

    // Create referenced table then foreign key referencing non-primary column should fail
    err = createTable(handler, t, `
        CREATE TABLE ct_ref (
            pk serial primary key,
            notpk int
        )
    `)
    if err != nil {
        t.Fatalf("failed to create referenced table: %v", err)
    }

    err = createTable(handler, t, `
        CREATE TABLE ct_fk_notpk (
            id serial primary key,
            other int,
            FOREIGN KEY (other) REFERENCES ct_ref(notpk)
        )
    `)
    if err == nil {
        t.Fatalf("expected foreign key referencing non-primary column to fail")
    }

	// Foreign key referencing same table should fail
    err = createTable(handler, t, `
        CREATE TABLE ct_fk_notpk (
            id serial primary key,
            other int,
            FOREIGN KEY (other) REFERENCES ct_fk_notpk(notpk)
        )
    `)
    if err == nil {
        t.Fatalf("expected foreign key referencing non-primary column to fail")
    }

	// Valid foreign key referencing primary key should succeed
    err = createTable(handler, t, `
        CREATE TABLE ct_fk_notpk (
            id serial primary key,
            other int,
            FOREIGN KEY (other) REFERENCES ct_ref(pk)
        )
    `)
    if err != nil {
        t.Fatalf("expected foreign key referencing non-primary column to fail")
    }
}