# Architecture

A deep-dive into the internals of the Custom Relational Database Engine — how each layer works, why it was designed the way it was, and what tradeoffs were made deliberately.

---

## Table of Contents

- [Overview](#overview)
- [SQL Query Engine](#1-sql-query-engine)
- [B+Tree Indexing](#2-btree-indexing)
- [Schema, Constraints & Relational Integrity](#3-schema-constraints--relational-integrity)
- [Slotted Page Architecture](#4-slotted-page-architecture)
- [LRU Buffer Pool Manager](#5-lru-buffer-pool-manager)
- [Metadata & Page Management](#6-metadata--page-management)
- [Binary Type System](#7-binary-type-system)
- [Testing & Bugs Found](#8-testing--bugs-found)
- [Benchmarks](#9-benchmarks)
- [Page Layout Reference](#10-page-layout-reference)

---

## Overview

The engine is structured as a strict set of layers, each only communicating with the one directly below it:

```
User input
    ↓
CLI (prompt, multiline input, meta-commands, result formatting)
    ↓
Tokenizer → Parser → Query (Parse + Execute interface)
    ↓
Storage Engine (B+Tree, linear scan, constraint enforcement, CRUD logic)
    ↓
Buffer Pool (LRU cache, dirty-page tracking)
    ↓
Disk (.bin file, slotted pages)
```

No layer skips another. The B+Tree never touches disk directly — it always goes through the buffer pool. The query layer never performs CRUD logic — it parses SQL into structured objects and calls the storage engine with the correct parameters. The CLI never touches the storage engine directly — it hands raw SQL to the query layer and receives a `QueryResult` back. This separation makes each layer independently testable and replaceable.

---

## 0. CLI

The CLI is the topmost layer — the only part of the system the user directly interacts with. It is responsible for input handling, query dispatch, and result formatting. It never touches the storage engine or buffer pool directly.

### Input Handling

Built using Go's `golang.org/x/term` package for raw terminal mode, giving the CLI full control over cursor movement, line editing, and multiline input. The prompt shows `>` for the first line of a new statement and `->` for continuation lines. A query is dispatched to the query layer only when a semicolon (`;`) is encountered — until then, input accumulates across as many lines as needed.

### Meta-commands

Meta-commands (prefixed with `\`) are intercepted before the SQL tokenizer sees the input, so they never interfere with SQL parsing:

| Command | Description |
|---|---|
| `\t` | List all tables |
| `\d <table>` | Describe a table's columns and constraints |
| `\w` | Flush all dirty pages to disk explicitly |
| `\q` | Quit |

### Result Formatting

The CLI receives a `*QueryResult` from the query layer. The result carries both the payload and the query type, so the CLI can format output correctly:

- `SELECT` results are printed as an aligned table with column headers, a separator line, and a row count footer.
- `INSERT`, `UPDATE`, `DELETE` results print the number of rows affected.
- `CREATE TABLE` prints a confirmation message.
- Errors are printed with a clear prefix and the query loop continues — one bad query doesn't terminate the session.

Column widths are computed dynamically from the maximum of each column's header length and its longest value, so output stays aligned regardless of data.

---

## 1. SQL Query Engine

A hand-written tokenizer, recursive-descent parser, and executor with no parser generator or third-party SQL library.

### Tokenizer

Converts raw SQL text into a flat slice of typed tokens. Each token carries a `Type` (keyword, identifier, string literal, number, operator, punctuation) and a `Value` (the raw text). Keywords are normalized to uppercase during tokenization so the parser never has to case-fold. String literals are delimited by single quotes per the SQL standard — double quotes produce a parse error with a clear message rather than silently accepting non-standard syntax.

Multi-character operators (`>=`, `<=`, `!=`) are handled by peeking one character ahead after matching the first character. The semicolon (`;`) is tokenized as its own type so the CLI can detect statement boundaries for multiline input without passing it to the parser.

### Parser

A recursive-descent parser that walks the token slice positionally, consuming tokens one at a time with direct index arithmetic — no helper-function abstraction layer, which keeps the code flat and the control flow obvious.

Each statement type has its own parse function returning a concrete struct (`SelectStatement`, `InsertStatement`, `UpdateStatement`, `DeleteStatement`, `CreateTableStatement`). A top-level dispatcher checks the first token's keyword and routes to the correct parse function.

**WHERE-clause expression parser** builds a binary expression tree from `AND`/`OR`/parenthesized conditions. The algorithm is a single left-to-right scan with two distinct node-attachment strategies:

- `AND` mutates the rightmost leaf node in place, extending it into an `AND` node — giving `AND` tighter binding than `OR` without an explicit precedence table.
- `OR` re-roots the entire tree, placing everything built so far as the right subtree of a new `OR` node.
- Parenthesized sub-expressions recurse into a nested `ParseWhereExpression` call, treating the result as a single atomic unit regardless of its internal complexity.

This approach correctly handles arbitrary `AND`/`OR`/parentheses combinations with standard SQL precedence, and was derived independently rather than via a textbook precedence-climbing algorithm.

### Query Interface

Every statement type implements the `Query` interface:

```go
type Query interface {
    Execute(engine *storage.StorageEngine) (*QueryResult, error)
    Parse(p *Parser) error
}
```

`Parse` is responsible for consuming the token stream and populating the statement's fields — each statement type knows how to parse itself. `Execute` calls into the storage engine with the already-parsed data and returns a `QueryResult`, which carries both the result payload (rows, affected count) and the query type so the CLI knows how to format and print the output correctly. The top-level caller receives a `Query` interface and calls `Parse` then `Execute` without knowing the concrete type. Adding a new statement type requires implementing both methods — the compiler enforces it, so a missing implementation is a build error rather than a runtime panic.

### Query Planning

A `SELECT` query checks whether its `WHERE` clause is a single equality condition on an indexed column. If so, it passes that condition to the storage engine, which routes directly through the B+Tree for an O(log n) lookup. Multi-condition queries (AND/OR expressions) pass the full expression tree to the storage engine, which evaluates it against each row during a linear scan. This is a simple, deterministic planning decision — there is no cost-based optimizer.

### Column Projection

`SELECT id, name FROM users` passes the requested column list to the storage engine. The storage engine computes the byte offset of each requested column within a row (using the null bitmap and column type sizes) and reads only those bytes rather than deserializing the full row. For tables with wide columns (long `VarChar` fields), this avoids loading and copying bytes that will be discarded immediately.

---

## 1b. Storage Engine (CRUD Logic)

The storage engine owns all CRUD logic — insert, search, update, delete. The query layer parses SQL and extracts the relevant parameters (table name, column values, WHERE expression), then calls into the storage engine's functions with those parameters. The storage engine never sees raw SQL.

### UPDATE

Before committing any change, `UPDATE` re-validates every affected value against the column's constraints — `UNIQUE` (via an index search for the new value), foreign keys (the referenced row must still exist), `NOT NULL` (cannot set to null unless a `DEFAULT` exists), and `Serial` immutability (serial columns cannot be manually updated). If any constraint fails, the entire update is rejected before any page is modified.

Two update paths exist:

- **In-place:** fixed-size, non-indexed columns have their bytes overwritten directly in the buffer pool page. The page is marked dirty.
- **Delete + reinsert:** if the new value changes the row's size (variable-length column) or affects an indexed column (whose B+Tree position would become incorrect), the old row is deleted and a new row is inserted with the merged data. Every index is updated atomically as part of this path.

The update loop uses a two-pass strategy — collect matching `RowID`s during the scan, then apply updates — to avoid the correctness problem of mutating the structure being traversed (see [Testing & Bugs Found](#8-testing--bugs-found)).

### DELETE

`DELETE` also uses a two-pass strategy, for the same reason as `UPDATE`: collecting matching rows during the scan, then deleting in a second pass, prevents slot-offset corruption when multiple rows on the same leaf page match the `WHERE` condition.

For indexed deletes (a single equality condition on an indexed column), the engine looks up the row directly via the B+Tree, avoiding a full scan entirely.

Page-level rebalancing (merging underfull leaf pages back together) is deliberately out of scope. After estimating the engineering cost against the realistic benefit at this project's scale — a single internal B+Tree node covers hundreds of leaf pages, so leaf-level deletions rarely meaningfully affect tree depth — the decision was to skip rebalancing. Freed slots within a page are reclaimed by the next insert into that page.

---

## 2. B+Tree Indexing

A recursive, type-agnostic B+Tree supporting O(log n) key lookups across arbitrarily large tables.

### Type Agnosticism

The tree never inspects the bytes it stores. Every operation — insert, search, delete — receives raw `[]byte` keys and delegates all ordering decisions to a pluggable comparison function provided at call time. The same tree implementation handles `TinyInt`, `BigInt`, and fixed-width `VarChar` columns identically, with the comparison function encoding all type-specific logic.

### Insertion

Insertion begins at the root and recurses to the correct leaf. At each level, the tree checks whether the current node has room for a potential split — if so, ancestor locks (in a future concurrent version) can be released early (lock-coupling / crabbing, a pattern derived independently during design). If not, ancestors must be kept locked in case the split propagates upward.

Within a leaf that has room, the new key is inserted in sorted position via in-place byte shifting — no slice allocation. When a leaf overflows, entries are loaded into a temporary slice, split at the median, and written to two pages. The median key is propagated to the parent. For internal nodes, the median moves up and out; for leaf nodes, it is copied up while remaining in the right sibling (matching standard B+Tree semantics for range scan continuity via the linked leaf chain).

When the root itself splits, a new root page is allocated and the two halves become its children — growing the tree by one level.

### Fixed-Size Keys

All keys in a given B+Tree are a fixed, uniform width known at tree construction time. This allows entry positions to be computed by direct multiplication (`LeafPageHeaderSize + i*(keySize+6)`) rather than scanning from the beginning, and makes the capacity of each page a simple division. Variable-length string keys are handled by the fixed-width `VarChar(N)` column type, which pads or truncates strings to a constant size.

### Leaf Chain

Every leaf page stores a `NextLeafPage` pointer to its right sibling. Full table scans and unindexed `WHERE` clauses walk this chain from the first leaf to the last — O(n), but with no tree traversal overhead. The chain is maintained through splits: when a leaf splits, the new right sibling inherits the old leaf's `NextLeafPage`, and the old leaf is updated to point to the new sibling.

### Deletion

B+Tree key deletion removes the entry from the leaf page by compacting remaining entries (shifting left, decrementing the key count). The internal node structure is not updated — internal nodes may retain keys that no longer have corresponding leaf entries. This is safe because the leaf chain is the source of truth for scans, and internal keys serve only as routing hints for point lookups. The routing hint being stale for a deleted key has no correctness impact since the key genuinely no longer exists in any leaf.

### Scale Verification

Tested with 1,000,000 sequential `BigInt` keys and 1,000,000 fixed-width `VarChar` string keys. String tests included case-sensitivity edge cases (`"Apple"` before `"apple"`, byte-order comparison) and prefix relationships (`"Cat"` before `"Catalog"`). Both sequential and interleaved (even-then-odd) insertion patterns were used, the latter specifically to exercise mid-page splits and the internal-node insertion path that only triggers when a child split propagates upward into a non-full parent.

---

## 3. Schema, Constraints & Relational Integrity

### Schema Storage

Table and column definitions are stored on metadata pages, serialized using length-prefixing for variable-length fields (column names, string defaults) and linked page chaining for overflow when the number of tables exceeds one page's capacity. On startup, all metadata is read from disk into an in-memory hashmap (`map[string]*Table`) for O(1) table lookup by name. Column offsets within this map are pre-indexed for fast lookup by column name as well.

### Null Bitmap

Each row begins with a compact null bitmap — one bit per column, packed into the minimum number of bytes. A set bit means the corresponding column is null; the column's bytes are absent from the row entirely (no sentinel values, no zero-padding for null columns). This bitmap must be read before any column offset can be computed, since preceding null columns shift the positions of subsequent ones.

### Constraints

- **NOT NULL:** checked at insert and update time. If the value is null and a `DEFAULT` exists, the default is substituted silently. If no default exists, the insert/update is rejected with an error.
- **UNIQUE:** enforced via an index search before the insert/update commits. If the new value already exists in the column's index, the operation is rejected.
- **DEFAULT:** stored per-column in metadata as a serialized byte value of the column's type. Applied when the column is omitted from an `INSERT` or set to null in an `UPDATE` where `NOT NULL` would otherwise reject it.
- **SERIAL:** an auto-incrementing integer backed by a per-table counter in metadata. The counter increments atomically with each insert and is persisted to disk. Serial columns cannot be manually updated — `UPDATE` rejects any attempt.
- **Foreign Keys:** stored as table-level constraints mapping a local column to a referenced table and column. On insert and update, the referenced table's index is searched for the new value — if not found, the operation is rejected. Foreign key declarations use standard SQL syntax: `FOREIGN KEY (col) REFERENCES other_table(col)`.

### Secondary Indexes

Any column can be declared with an index, giving it its own B+Tree separate from the primary key's tree. The query executor checks available indexes before executing a `SELECT` — if the `WHERE` condition is a single equality on an indexed column, it routes through that column's B+Tree instead of scanning. On insert, every indexed column's value is added to its corresponding tree. On delete, every indexed column's key is removed. On update, indexed columns trigger the delete+reinsert path to keep tree positions correct.

---

## 4. Slotted Page Architecture

All persistent data — row data, metadata, B+Tree nodes — lives in 4KB pages. This size matches the OS memory page size, aligning disk I/O with the operating system's own unit of memory management and avoiding partial-page reads.

### Layout

```
[ Free Space Offset (2B) ] [ Num Elements (2B) ]
[ Slot 0 Offset (2B) ] [ Slot 1 Offset (2B) ] ... [ Slot N Offset (2B) ]
                    ... (Free Space) ...
[ Row N Data ] [ Row N-1 Data ] ... [ Row 0 Data ]
```

The slot directory grows forward from the page header; row data grows backward from the page end. Free space sits between them. A row's physical location is found in O(1) by reading its slot's offset from the slot directory — no scanning.

### Insertion

Before writing, a dry-run pass calculates the exact byte size of the serialized row (accounting for the null bitmap, each column's type width, and any variable-length `VarChar` content) and checks whether it fits in the page's remaining free space. If it doesn't fit, a new page is allocated. This prevents partial writes — a row is either fully written or not written at all.

### Deletion

Deleting a row compacts the slot directory: the deleted slot's entry is removed and subsequent slots are shifted. The row's bytes are reclaimed by updating the free space offset. The freed slot becomes available for the next insert — new rows are allocated from free space at the page end, and previously-freed slots may be reused if they fit the new row's size.

### Symmetric I/O

Every type has a paired serializer and deserializer. The serializer writes a fixed, known byte layout; the deserializer reads it back using the same field-size assumptions. These two functions are the only place type-specific byte layout logic lives — nothing else in the engine makes assumptions about how a particular type is represented on disk.

---

## 5. LRU Buffer Pool Manager

No component above the buffer pool ever reads or writes disk directly. Every page access goes through `Bp.Get(pageID)`, which either returns the page from cache or loads it from disk and caches it.

### Data Structures

The pool maintains three structures in parallel:

- **`[PoolSize][4096]byte`** — a fixed, statically-allocated array of 4KB page buffers. Pages live at known array indices, not heap-allocated per-page. The garbage collector never scans individual page buffers.
- **`map[uint32]*Node`** — maps a page ID to a pointer directly into the LRU linked list node for that page. O(1) lookup.
- **Doubly linked list** — orders pages from most-recently-used (head) to least-recently-used (tail). On access, a page's node is moved to the head. On eviction, the tail is removed.

A free-slot stack tracks which array indices are available. When a page is evicted, its slot index is pushed onto the free stack. When a new page is loaded, a slot is popped from the free stack.

### Eviction

When the pool is full and a new page must be loaded, the tail of the LRU list is evicted. If the evicted page is dirty (has been modified since it was last written to disk), it is flushed to disk before eviction. This is the only moment a non-explicit write hits disk — all other modifications stay in memory until forced out by eviction or an explicit `\w` flush command.

### Dirty Page Tracking

Every function that modifies a page's bytes must call `MarkDirty(pageID)` after writing. This was the source of one of the most subtle bugs found in the project (see [Testing & Bugs Found](#8-testing--bugs-found)).

### Performance

```
BenchmarkWarmRead-12    74507995    15.79 ns/op    0 B/op    0 allocs/op
BenchmarkColdRead-12      628132  1859.00 ns/op   52 B/op    1 allocs/op
```

A warm hit completes in ~16ns with zero allocations. A cold miss costs ~1.9µs and allocates one `Node` struct (52 bytes) to register the newly loaded page in the hashmap and linked list. The warm/cold gap is ~118x, measured with Go's built-in benchmarking tool on an AMD Ryzen 5 5600.

---

## 6. Metadata & Page Management

### Free Space Manager

A global `map[uint32]uint16` maps each page ID to its remaining free bytes, loaded from disk at startup and kept in sync as rows are inserted and deleted. When a row needs to be written, the free space manager finds a page with sufficient room in O(1) — no disk scan, no page-by-page probing.

### Metadata Pages

Table and column definitions are serialized to dedicated metadata pages, separate from data pages. On startup, all metadata pages are read in chain order and their contents loaded into the in-memory table hashmap.

### Metadata Page Binary Layout

```
Meta Page:
  Next Meta Page Pointer    (4 bytes)
  Free Space Offset         (2 bytes)
  Number of Tables          (2 bytes)
  [List of Table Offsets    (2 bytes each)]

At each Table Offset:
  Table Name Length         (1 byte)
  Table Name                (variable)
  Number of Columns         (1 byte)

  For each Column:
    Column Name Length      (1 byte)
    Column Name             (variable)
    Data Type               (1 byte)
    Constraints             (1 byte, bit flags)
    Size                    (1 byte)
    Default Value           (variable, if applicable)

  Number of Indexes         (1 byte)
  For each Index:
    Column index            (1 byte)  — position in Columns array
    Index Root Page ID      (4 bytes)

  Number of Foreign Keys    (1 byte)
  For each Foreign Key:
    Column index            (1 byte)  — position in Columns array
    Referenced Table Name Length  (1 byte)
    Referenced Table Name   (variable)
    Referenced Column index (1 byte)  — position in referenced table's Columns array
```

The metadata page uses the same slotted-page layout as data pages — a slot directory at the front, table definitions growing from the end. If the number of tables exceeds one page's capacity, a linked `Next Meta Page Pointer` chains to an additional metadata page.

---

## 7. Binary Type System

All data is stored in a custom binary format, not as text. Every value is serialized to a fixed-size or length-prefixed byte slice, written directly to a page buffer, and deserialized by reading the same number of bytes at the same offset.

### Types

| Type | Size | Notes |
|---|---|---|
| `TinyInt` | 1 byte | Signed 8-bit integer |
| `SmallInt` | 2 bytes | Signed 16-bit integer, BigEndian |
| `Int` | 4 bytes | Signed 32-bit integer, BigEndian |
| `BigInt` | 8 bytes | Signed 64-bit integer, BigEndian |
| `Serial` | 4 bytes | Auto-incrementing `Int`, managed by engine |
| `VarChar(N)` | N+1 bytes | 1-byte length prefix + up to N bytes of content |
| `VarChar` (unbounded) | len+1 bytes | 1-byte length prefix + actual content bytes |

BigEndian encoding is used throughout for cross-platform binary compatibility — the same `.bin` file produces the same bytes regardless of the host machine's native endianness.

### Type Guardrails

Every value is validated against its declared column type before serialization. An `Int` value that would overflow 32 bits is rejected with an error before a single byte is written. This prevents silent data corruption from type truncation — a common failure mode in systems that defer validation to the storage layer or rely on language-level implicit conversions.

### VarChar Storage

Fixed-width `VarChar(N)` stores exactly `N+1` bytes per value: one byte for the actual string length, followed by up to `N` bytes of content. Short strings are not padded — the length prefix tells the deserializer exactly how many bytes to read. This design keeps B+Tree entry sizes uniform (required for the tree's fixed-offset arithmetic) while still supporting strings shorter than the declared maximum.

Unbounded `VarChar` (no declared length) stores `len(value)+1` bytes — a 1-byte length prefix followed by the string bytes. Unbounded `VarChar` columns cannot be used as B+Tree index keys, since variable-length keys would break the tree's fixed-entry-size assumption.

---

## 8. Testing & Bugs Found

Unit tests are written per component (`_test.go` files in the same package), using isolated per-test databases via Go's `t.TempDir()`. Each test gets a fresh `.bin` file that is cleaned up automatically when the test completes.

### Bug 1 — Missing MarkDirty (Silent Data Loss)

**Symptom:** a large-scale insert test (1,000,000 keys) passed all insertions without error, but a subsequent full search found some keys missing — no error, no crash, just absent data.

**Cause:** one insertion code path wrote bytes into a buffer pool page but forgot to call `MarkDirty`. The write succeeded in memory, and the page looked correct for as long as it stayed in the pool. When the pool eventually evicted that page under memory pressure, it saw the page as clean and discarded it without flushing to disk. The next search loaded a stale version of the page from disk, finding neither the key nor any indication that it had ever existed.

**Why it was hard to find:** the bug only manifested under sustained insertion volume — enough to actually fill the buffer pool and trigger eviction of the specific page. Small-scale tests never evicted the affected page, so the data loss was invisible. The fix was a single `MarkDirty` call; the test confirmed the fix by reinserting 1,000,000 keys and verifying every one could be found.

**Lesson:** a dirty-page tracking system requires every write path to participate. One missing call silently breaks durability for exactly the rows written through that path.

---

### Bug 2 — Traverse-While-Delete (Skipped Rows)

**Symptom:** deleting multiple rows matching a `WHERE` condition that covered several rows on the same leaf page would silently skip some of them.

**Cause:** the original `LinearDelete` deleted each matching row immediately upon finding it, while still iterating the leaf page's entry list. Deletion compacts the page — remaining entries shift left, slot counts decrement. This meant the iteration's position index was now pointing at a different entry than expected, causing alternating entries to be skipped.

**Fix:** switched to a two-pass approach (same pattern used in `LinearUpdate`): scan the entire leaf chain, collecting matching `RowID`s into a slice. Only after the scan is complete, iterate the collected `RowID`s and delete each one. The scan and the mutation are fully separated, so the mutation never corrupts the scan's traversal state.

**Lesson:** modifying a data structure while iterating it is a class of bug that produces no error and leaves no obvious evidence — just silently missing rows. Unit tests that verify row counts after multi-row deletes are the only reliable way to catch this.

---

## 9. Benchmarks

### Indexed Lookup vs. Full Table Scan

Measured on a 1,000,000-row `BigInt`-keyed table (AMD Ryzen 5 5600), searching for the same worst-case key (last in sorted order) via both an indexed B+Tree lookup and a full linear scan of the leaf chain:

```
BenchmarkIndexSearch-12     894442       1340 ns/op
BenchmarkLinearSearch-12        67   16969646 ns/op
```

An indexed lookup is roughly **12,660x faster** than a full scan on this dataset.

With ~292 entries per leaf page and ~341-way fanout on internal nodes, the tree stays at just 3 levels for datasets up to ~34 million entries — every lookup costs at most 3 page accesses. A full scan's cost grows linearly with row count, regardless of tree depth.

### Buffer Pool: Cache Hit vs. Miss

```
BenchmarkWarmRead-12    74507995    15.79 ns/op    0 B/op    0 allocs/op
BenchmarkColdRead-12      628132  1859.00 ns/op   52 B/op    1 allocs/op
```

A warm hit is ~118x faster than a cold read. The cold path's single allocation (52 bytes) is the `Node` struct inserted into the LRU linked list and hashmap when a new page is loaded. Warm hits allocate nothing — the LRU list repoints an existing node.

The cold benchmark measures the cost of a cache miss that still benefits from OS-level disk caching. A genuinely cold disk read with no OS cache would be slower still.

---

## 10. Page Layout Reference

### Data Page (4KB)

```
[ Free Space Offset (2B) ] [ Num Elements (2B) ]
[ Slot 0 Offset (2B) ] [ Slot 1 Offset (2B) ] ... [ Slot N Offset (2B) ]
                    ... (Free Space) ...
[ Row N Data (Variable) ]
[ Row N-1 Data (Variable) ]
[ Row 0 Data (Variable) ]
```

### B+Tree Internal Node

`N` keys, `N+1` child pointers:

```
[ IsLeaf (1B) ] [ NumKeys (2B) ]
[ P0 (4B) ] [ K0 ] [ P1 (4B) ] [ K1 ] ... [ PN (4B) ]
```

### B+Tree Leaf Node

Keys paired with a Record ID (`PageID` + `Slot`), plus a pointer to the next leaf:

```
[ IsLeaf (1B) ] [ NextLeafPage (4B) ] [ NumKeys (2B) ]
[ K0 ] [ PageID0 (4B) ] [ Slot0 (2B) ] [ K1 ] [ PageID1 (4B) ] [ Slot1 (2B) ] ...
```
