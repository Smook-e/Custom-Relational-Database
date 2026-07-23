# Custom Relational Database Engine (Go)

A relational database engine built from scratch in Go. Every byte on disk is placed there by code in this repository: a slotted-page storage engine, an LRU-backed buffer pool, and a Multi-type B+Tree index supporting O(log n) lookups over millions of rows.

This is not a key-value store with SQL syntax bolted on. It's an attempt to reproduce, and understand, the actual mechanics that PostgreSQL and SQLite are built on — page layout, buffer management, and tree-structured indexing — from scratch.

## Core Engine Features

### 1. B+Tree Indexing

A recursive, type-agnostic B+Tree supporting O(log n) key lookups, built to index primary keys across arbitrarily large tables without knowing what type of data it's indexing.

- **Type-agnostic by design:** the tree never inspects key contents — it operates entirely on raw byte buffers, delegating all ordering decisions to a pluggable comparison function. The same tree implementation indexes `TinyInt` and `BigInt` columns identically.
- **In-place insertion with binary search:** keys are inserted directly into existing pages via byte-level shifting when space allows, avoiding unnecessary allocation on the common path.
- **Page-splitting with median propagation:** on overflow, a page splits and its median key propagates to the parent — internal nodes push the median up and out, leaf nodes copy it up while retaining it, matching standard B+Tree semantics.
- **Linked leaf pages:** every leaf holds a pointer to its right sibling, enabling full table scans and range queries to walk the leaf layer directly without re-traversing the tree.
- **Multi-level tree growth:** when a root page overflows, a new root is created above it, growing the tree by a level — confirmed by the 100,000-key test below, which produces a multi-level tree.
- **Verified at scale:** tested with 100,000 sequential `BigInt` keys across 690 total pages (~291 entries per leaf page), and with interleaved insertion patterns that exercise mid-page splits and boundary conditions.

### 2. Slotted Page Architecture

A **slotted page** layout for both table metadata and data rows, using 4KB pages to align with OS and disk page sizes.

- **Forward-growing slot directory:** a slot array at the page start storing exact byte offsets of records for O(1) access by index.
- **Backward-growing payload:** row data is appended from the page end toward the header, so variable-length rows pack tightly without wasted space.
- **Symmetric I/O:** a mirrored serialization/deserialization pipeline guarantees that what gets written is exactly what gets read back — no drift between the two paths.

### 3. LRU Buffer Pool Manager

A custom buffer pool sits between every higher layer (B+Tree, page logic) and disk, so no component ever performs raw disk I/O directly.

- **LRU eviction in O(1):** a doubly linked list tracks access order, a hash map gives direct pointers into that list — both lookup and eviction are constant time.
- **Dirty page tracking:** modified pages are marked in memory and only flushed to disk on eviction or explicit commit, avoiding a syscall on every write.
- **GC-friendly fixed allocation:** the pool holds pages in a static, fixed-size array (`[PoolSize][4096]byte`) rather than heap-allocating a struct per page, so the garbage collector never has to scan or manage individual page buffers.

### 4. Advanced Metadata & Page Management

- **Schema serialization:** table and column definitions are stored using length-prefixing and linked page chaining (`nextPage` pointers), supporting an unbounded number of tables without a fixed-size metadata region.
- **Free space manager:** a global map of `PageID → RemainingBytes` allows the engine to locate a page with sufficient room instantly, without scanning disk.
- **Sized allocation:** a dry-run pass validates row size and type constraints before a single byte is committed to a page, preventing partial writes on invalid input.
- **Zero-allocation buffering:** metadata page operations draw scratch buffers from a `sync.Pool` rather than allocating fresh ones, reducing GC pressure during schema reads and writes.

### 5. Strict Binary Type System

- **Supported types:** `TinyInt` (8-bit), `SmallInt` (16-bit), `Int` (32-bit), `BigInt` (64-bit), and `VarChar`.
- **Type guardrails:** every value is validated against its column's bit-width during conversion, catching silent overflow before it corrupts a page.
- **BigEndian encoding:** used throughout for cross-platform binary compatibility.

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

## Roadmap

- [x] Slotted Page Layout
- [x] Free Space Management
- [x] Binary Type Serialization
- [x] LRU Buffer Pool & Dirty Page Tracking
- [x] Atomic Commit Logic
- [x] B+Tree Indexing — insertion, page-splitting, multi-level root growth, and search, verified at 100K+ rows
- [ ] Query Engine: `SELECT`, `INSERT`, `UPDATE`, `DELETE` via a hand-written SQL parser
- [ ] Concurrency Control: thread-safe access and locking
