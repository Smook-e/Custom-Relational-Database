# Custom Relational Database Engine

A relational database engine built from scratch in Go, implementing the storage, indexing, and query execution mechanisms.

The engine uses **4KB slotted pages**, a custom **LRU buffer pool**, and a **type-agnostic B+Tree** for persistent, logarithmic-time indexing. On top of that sits a hand-written **SQL tokenizer, parser, and executor** supporting full CRUD — `CREATE TABLE`, `INSERT`, `SELECT` (with column projection and `WHERE` clauses supporting `AND`/`OR`/parentheses), `UPDATE`, and `DELETE` all run as real SQL text, not just direct Go function calls. It also supports secondary indexes, constraints, foreign keys, NULL handling, and multiple binary data types.

## Core Engine Features

### 1. SQL Query Engine

A hand-written tokenizer, recursive-descent parser, and executor sitting above the storage engine — no parser generator, no third-party SQL library.

- **Tokenizer:** converts raw SQL text into typed tokens (keywords, identifiers, literals, operators), with case-insensitive keyword matching and quoted string literal handling.
- **WHERE-clause expression parser:** builds a binary expression tree from `AND`/`OR`/parenthesized conditions with correct operator precedence (`AND` binds tighter than `OR`) and support for nested grouping — derived independently rather than via a standard precedence-climbing algorithm.
- **Statement interface:** each statement type (`SelectStatement`, `InsertStatement`, `CreateTableStatement`, `UpdateStatement`, `DeleteStatement`) implements a common `Execute` method, so adding a new statement type is enforced at compile time rather than requiring a manually-maintained dispatch table.
- **Query planning:** a `SELECT` with a single indexable condition routes directly through the B+Tree; multi-condition queries fall back to walking the expression tree against a linear scan.
- **Column projection:** `SELECT id, name FROM ...` returns only the requested columns, reading and validating rows via direct column-offset seeks rather than deserializing full rows unnecessarily.
- **UPDATE:** re-validates constraints (`UNIQUE`, foreign keys, `NOT NULL`/`DEFAULT`, `Serial` immutability) against the new values before committing. Fixed-size, non-indexed columns are updated in place; a size-changing update or a change to an indexed column falls back to delete-and-reinsert, keeping every index in sync with the row's new location.
- **DELETE:** removes matching rows via in-page entry compaction and slot reclamation (freed slots are reused by the next insert into that page); page-level rebalancing across the B+Tree is deliberately out of scope, a scoping decision made after estimating its cost-to-benefit ratio at this project's scale.

### 2. B+Tree Indexing

A recursive, type-agnostic B+Tree supporting O(log n) key lookups, indexing both primary and secondary columns across arbitrarily large tables without knowing what type of data it's indexing.

- **Type-agnostic by design:** the tree never inspects key contents — it operates entirely on raw byte buffers, delegating all ordering decisions to a pluggable comparison function. The same tree implementation indexes `TinyInt`, `BigInt`, and fixed-width `VarChar` columns identically.
- **In-place insertion with binary search:** keys are inserted directly into existing pages via byte-level shifting when space allows, avoiding unnecessary allocation on the common path.
- **Page-splitting with median propagation:** on overflow, a page splits and its median key propagates to the parent — internal nodes push the median up and out, leaf nodes copy it up while retaining it, matching standard B+Tree semantics.
- **Linked leaf pages:** every leaf holds a pointer to its right sibling, enabling full table scans and range queries to walk the leaf layer directly without re-traversing the tree.
- **Multi-level tree growth:** when a root page overflows, a new root is created above it, growing the tree by a level.
- **Verified at scale:** tested with 1,000,000 sequential `BigInt` keys and 1,000,000 fixed-width `VarChar` string keys, including case-sensitivity and prefix-relationship edge cases (e.g. `"Cat"` vs `"Catalog"`), across sequential and interleaved insertion patterns that force mid-page and multi-level splits.

### 3. Schema, Constraints & Relational Integrity

A full schema layer sitting above the storage engine, enforced directly in the insert and update paths rather than left to the caller.

- **Secondary indexes:** any column, not just the primary key, can have its own B+Tree, letting the engine choose the right index for a given lookup rather than always falling back to a full scan.
- **NULL handling via a null bitmap:** each row carries a compact bitmap marking which columns are null, avoiding the need to reserve space or sentinel values for absent data.
- **NOT NULL and UNIQUE constraints:** validated at insert and update time, before a row is committed to a page.
- **DEFAULT values:** static defaults for any type, plus a distinct `Serial` type backed by a per-table auto-incrementing counter stored in table metadata.
- **Fixed-width `VarChar(N)`:** strings are stored in a constant-size slot (`N + 1` bytes: a 1-byte length prefix plus up to `N` bytes of content), keeping B+Tree entry sizes uniform and making string columns indexable with the same offset math as fixed-width integer types.
- **Foreign keys:** column-level references to other tables, enforced during insert and update.

### 4. Slotted Page Architecture

A **slotted page** layout for both table metadata and data rows, using 4KB pages to align with OS and disk page sizes.

- **Forward-growing slot directory:** a slot array at the page start storing exact byte offsets of records for O(1) access by index.
- **Backward-growing payload:** row data is appended from the page end toward the header, so variable-length rows pack tightly without wasted space.
- **Symmetric I/O:** a mirrored serialization/deserialization pipeline guarantees that what gets written is exactly what gets read back — no drift between the two paths.

### 5. LRU Buffer Pool Manager

A custom buffer pool sits between every higher layer (B+Tree, page logic) and disk, so no component ever performs raw disk I/O directly.

- **LRU eviction in O(1):** a doubly linked list tracks access order, a hash map gives direct pointers into that list — both lookup and eviction are constant time, not the O(log n) a naive priority-queue-based LRU would cost.
- **Dirty page tracking:** modified pages are marked in memory and only flushed to disk on eviction or explicit commit, avoiding a syscall on every write.
- **GC-friendly fixed allocation:** the pool holds pages in a static, fixed-size array (`[PoolSize][4096]byte`) rather than heap-allocating a struct per page, so the garbage collector never has to scan or manage individual page buffers.

### 6. Advanced Metadata & Page Management

- **Schema serialization:** table and column definitions are stored using length-prefixing and linked page chaining (`nextPage` pointers), supporting an unbounded number of tables without a fixed-size metadata region.
- **Free space manager:** a global map of `PageID → RemainingBytes` allows the engine to locate a page with sufficient room instantly, without scanning disk.
- **Sized allocation:** a dry-run pass validates row size and type constraints before a single byte is committed to a page, preventing partial writes on invalid input.
- **Zero-allocation buffering:** metadata page operations draw scratch buffers from a `sync.Pool` rather than allocating fresh ones, reducing GC pressure during schema reads and writes.

### 7. Strict Binary Type System

- **Supported types:** `TinyInt` (8-bit), `SmallInt` (16-bit), `Int` (32-bit), `BigInt` (64-bit), `Serial`, and `VarChar` (fixed-width or unbounded).
- **Type guardrails:** every value is validated against its column's bit-width and constraints during conversion, catching silent overflow or invalid input before it corrupts a page.
- **BigEndian encoding:** used throughout for cross-platform binary compatibility.

## Testing

Unit tests cover the buffer pool, metadata pages, B+Tree, and WHERE-clause parsing.

- **Notable bug found:** a missing `MarkDirty` call on one insertion path let a write succeed in memory but be silently dropped on buffer pool eviction — no error, no crash. Fixed and reconfirmed at 1M-key scale.

## Benchmarks

**Indexed lookup vs. full table scan**, on a 1,000,000-row `BigInt`-keyed table (AMD Ryzen 5 5600), searching for the same worst-case key via both an indexed B+Tree lookup and a full linear scan of the leaf chain:

```text
BenchmarkIndexSearch-12     894442       1340 ns/op
BenchmarkLinearSearch-12        67   16969646 ns/op
```

An indexed lookup is roughly **12,660x faster** than a full scan on this dataset. With ~292 entries per leaf page and ~341-way fanout on internal nodes, the tree stays at just 3 levels for datasets up to ~34 million entries — every lookup up to that scale costs at most 3 page accesses, while a full scan cost grows linearly with row count.

**Buffer pool cache hit vs. miss**, measured separately (AMD Ryzen 5 5600):

```text
BenchmarkWarmRead-12    74507995      15.79 ns/op      0 B/op    0 allocs/op
BenchmarkColdRead-12      628132    1859.00 ns/op     52 B/op    1 allocs/op
```

A warm hit is roughly **118x faster** than a cold read, and allocates nothing — the LRU list simply repoints an existing node. A cold read allocates one `Node` (52 bytes) to register the newly loaded page in the pool's hash map and linked list. This is a lower bound on the real-world gap: the test machine's OS-level disk cache likely absorbs some of the cold path's I/O cost, so a genuinely cold disk read on a full cache-miss would be slower still.

## Page Layout (4KB)

```text
[ Free Space Offset (2B) ] [ Num Elements (2B) ]
[ Slot 0 Offset (2B) ] [ Slot 1 Offset (2B) ] ... [ Slot N Offset (2B) ]
... (Unused Space) ...
[ Row N Data (Variable) ]
[ Row N-1 Data (Variable) ]
[ Row 0 Data (Variable) ]
```

## B+Tree Node Layout

**Internal node** — `N` keys, `N+1` child pointers:
```text
[ IsLeaf (1B) ] [ NumKeys (2B) ]
[ P0 (4B) ] [ K0 ] [ P1 (4B) ] [ K1 ] ... [ PN (4B) ]
```

**Leaf node** — keys paired with a Record ID (`PageID` + `Slot`), plus a pointer to the next leaf:
```text
[ IsLeaf (1B) ] [ NextLeafPage (4B) ] [ NumKeys (2B) ]
[ K0 ] [ PageID0 (4B) ] [ Slot0 (2B) ] [ K1 ] [ PageID1 (4B) ] [ Slot1 (2B) ] ...
```

