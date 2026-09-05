# Custom Relational Database Engine

A relational database engine built from scratch in Go, implementing the storage, indexing, and query execution mechanisms behind modern databases — no ORM, no third-party SQL library, no borrowed storage layer.

**[Download latest release](https://github.com/Smook-e/Custom-Relational-Database/releases/latest)**

---

## Contents

- [Features](#features)
- [Installation](#installation)
- [Usage](#usage)
- [SQL Support](#sql-support)
- [Benchmarks](#benchmarks)
- [Roadmap](#roadmap)
- [Architecture](ARCHITECTURE.md)

---

## Features

- **Hand-written SQL engine** — tokenizer, parser, and executor with no external dependencies
- **B+Tree indexing** — O(log n) primary and secondary key lookups, verified at 1,000,000 rows
- **LRU buffer pool** — page caching with dirty-page tracking, minimizing disk I/O
- **Slotted page storage** — 4KB pages with O(1) record access and efficient variable-length row packing
- **Full CRUD** — `CREATE TABLE`, `INSERT`, `SELECT`, `UPDATE`, `DELETE`
- **Column projection** — `SELECT id, name FROM ...` reads only the requested columns
- **WHERE clause** — single conditions and compound expressions with `AND`, `OR`, and parentheses
- **Secondary indexes** — any column can have its own B+Tree for fast lookups
- **Constraints** — `NOT NULL`, `UNIQUE`, `DEFAULT`, `SERIAL` (auto-increment), foreign keys
- **NULL handling** — compact null bitmap per row, no sentinel values
- **Multiple data types** — `TinyInt`, `SmallInt`, `Int`, `BigInt`, `Serial`, `VarChar(N)`
- **Persistent storage** — all data survives restarts, stored in a single `.bin` file
- **Interactive CLI** — multiline SQL input, meta-commands, formatted table output

---

## Installation

### Download a pre-built binary

Download the binary for your platform from the [latest release](https://github.com/Smook-e/Custom-Relational-Database/releases/latest).

**Linux**
```bash
chmod +x sql_engine_linux
./sql_engine_linux mydb.bin
```

**macOS**
```bash
chmod +x sql_engine_mac
./sql_engine_mac mydb.bin
```

**Windows**
```powershell
sql_engine_windows.exe mydb.bin
```

> The `.bin` file is your database — it will be created automatically on first run if it doesn't exist. You can name it anything you like.

### Build from source

Requires [Go 1.21+](https://go.dev/dl/).

```bash
git clone https://github.com/Smook-e/Custom-Relational-Database.git
cd Custom-Relational-Database
go run main.go
```

Or build a binary yourself:

```bash
go build -o sql_engine main.go
./sql_engine
```

---

## Usage

Start the engine by running the binary with a database file path:

```bash
./sql_engine mydb.bin
```

The file will be created automatically on first run. You will see an interactive prompt:

```
> 
```

SQL statements are executed when a semicolon is reached. Multiline input is supported — the prompt changes to `->` on continuation lines:

```sql
> SELECT * FROM users
-> WHERE age > 18;
```

### Meta-commands

| Command | Description |
|---|---|
| `\t` | List all tables |
| `\d <table>` | Describe a table's columns and constraints |
| `\w` | Flush all dirty pages to disk |
| `\q` | Quit |

---

## SQL Support

### CREATE TABLE

```sql
CREATE TABLE test_users (
    id     SERIAL PRIMARY KEY,
    name   VARCHAR(50) NOT NULL DEFAULT 'anonymous',
    email  VARCHAR(30) NOT NULL UNIQUE,
    age    INT DEFAULT 18,
    job    VARCHAR(50) NOT NULL
);
```

With a foreign key:

```sql
CREATE TABLE orders (
    id       SERIAL PRIMARY KEY,
    user_id  INT,
    FOREIGN KEY (user_id) REFERENCES test_users(id)
);
```

Supported column types: `TINYINT`, `SMALLINT`, `INT`, `BIGINT`, `SERIAL`, `VARCHAR(N)`

Supported constraints: `PRIMARY KEY`, `NOT NULL`, `UNIQUE`, `DEFAULT <value>`, `FOREIGN KEY (<col>) REFERENCES <table>(<col>)`

---

### INSERT

Single row:
```sql
INSERT INTO test_users (name, email, age, job)
VALUES ('alice', 'alice@example.com', 25, 'engineer');
```

Multiple rows:
```sql
INSERT INTO test_users (name, email, age, job) VALUES
    ('alice',   'alice@example.com',   25, 'engineer'),
    ('bob',     'bob@example.com',     30, 'manager'),
    ('charlie', 'charlie@example.com', 35, 'developer');
```

---

### SELECT

```sql
-- All columns
SELECT * FROM test_users;

-- Column projection
SELECT id, name, email FROM test_users;

-- WHERE clause
SELECT * FROM test_users WHERE age > 30;

-- Compound conditions
SELECT * FROM test_users WHERE age > 25 AND job = 'engineer';

-- Parenthesized grouping
SELECT * FROM test_users WHERE (age > 30 OR job = 'manager') AND name = 'bob';
```

Example output:

```
> SELECT * FROM test_users;
| id | name    | email               | age | job        |
+----+---------+---------------------+-----+------------+
| 1  | alice   | alice@example.com   | 25  | engineer   |
| 2  | bob     | bob@example.com     | 30  | manager    |
| 3  | charlie | charlie@example.com | 35  | developer  |
| 4  | dave    | dave@example.com    | 40  | designer   |
| 5  | eve     | eve@example.com     | 45  | analyst    |
| 6  | frank   | frank@example.com   | 50  | consultant |
(6 rows)
```

---

### UPDATE

```sql
-- Update a single column
UPDATE test_users SET age = 26 WHERE id = 1;

-- Update multiple columns
UPDATE test_users SET name = 'alice_updated', age = 26 WHERE id = 1;
```

---

### DELETE

```sql
DELETE FROM test_users WHERE id = 1;

DELETE FROM test_users WHERE age > 40;
```


## Roadmap

- [x] Slotted Page Layout
- [x] Free Space Management
- [x] Binary Type Serialization
- [x] LRU Buffer Pool & Dirty Page Tracking
- [x] Atomic Commit Logic
- [x] B+Tree Indexing — insertion, page-splitting, multi-level root growth, and search, verified at 1M+ rows for both integer and string keys
- [x] Secondary Indexes
- [x] Constraints — NOT NULL, UNIQUE, DEFAULT, Serial, Foreign Keys, null-bitmap NULL handling
- [x] SQL Query Engine — tokenizer, parser, and executor for full CRUD: `CREATE TABLE`, `INSERT`, `SELECT` (with column projection and `WHERE` clause support), `UPDATE`, `DELETE`
- [x] Unit test suite covering buffer pool, metadata pages, B+Tree, and the storage layer
- [ ] Concurrency Control: thread-safe access and locking 
```
