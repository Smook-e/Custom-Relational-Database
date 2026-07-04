

# Custom Relational Database Engine (Go)

A high-performance, from-scratch relational database engine implemented in Go. This project focuses on the **physical storage layer**, implementing a custom binary serialization format and a slotted-page architecture to manage variable-length data on disk.

## 🚀 Core Engine Features

### 1. Slotted Page Architecture
Implemented a **Slotted Page** layout for both table metadata and data rows. Each page consists of 4kb to match RAM and Operating system pages. This allows the engine to handle variable-length records while maintaining minimal access to the disk.
- **Forward-Growing Header:** A slot array at the beginning of each 4KB page that stores the exact byte offsets of records.
- **Backward-Growing Payload:** Data is appended from the end of the page moving toward the header, maximizing space utilization.
- **Symmetric Read/Write:** A perfectly mirrored serialization and deserialization pipeline ensuring data integrity.

### 2. Advanced Metadata Management
The engine manages table schemas (names, columns, types, and constraints) using a specialized meta-page system.
- **Variable-Length Serialization:** Implemented length-prefixing for table and column names to support dynamic schema definitions.
- **Linked Page Chaining:** Metadata pages are linked via `nextPage` pointers, allowing the schema to grow across multiple pages if the number of tables exceeds a single page's capacity.

### 3. Intelligent Page Management
Implemented a custom **Free Space Manager** to optimize data placement and avoid expensive disk scans.
- **Free Space Tracking:** The engine maintains a global map of `PageID` and `RemainingBytes` to instantly locate a page with sufficient space for a new row.
- **On-the-Fly Page Allocation:** Automatically initializes and registers new data pages when existing pages cannot accommodate new insertions.
- **Two-Pass Allocation:** Every row insertion is preceded by a "Dry Run" size calculation to ensure the target page can fit the record before any binary write occurs.

### 4. Strict Binary Type System
To ensure data integrity and prevent silent corruption, the engine implements a strict binary type system with explicit bit-width conversions.
- **Supported Types:** `TinyInt` (8-bit), `SmallInt` (16-bit), `Int` (32-bit), `BigInt` (64-bit), and `VarChar` (Variable length).
- **Sized Validation:** Implemented a two-pass validation system that converts user input strings into concrete Go types and validates them against the schema's bit-limits before writing to disk.
- **Symmetric Binary I/O:** Uses `BigEndian` encoding to ensure cross-platform compatibility of the database files.

## 🛠 Technical Stack
- **Language:** Go (Golang)
- **I/O:** Direct file manipulation via `os.File` and `WriteAt`/`ReadAt` for precise byte-level control.
- **Memory Management:** Optimized using `sync.Pool` for buffer reuse to reduce GC pressure during high-frequency page operations.

## 📂 Internal Page Layout (4KB)
```text
[ Free Space Offset (2B) ] [ Num Elements (2B) ]
[ Slot 0 Offset (2B) ] [ Slot 1 Offset (2B) ] ... [ Slot N Offset (2B) ]
... (Unused Space) ...
[ Row N Data (Variable) ]
[ Row N-1 Data (Variable) ]
[ Row 0 Data (Variable) ]
```

## 📈 Future Roadmap
- [ ] **Buffer Pool Manager:** Implement an LRU Cache to reduce disk I/O.
- [ ] **B+ Tree Indexing:** High-speed lookups for primary and secondary keys.
- [ ] **Query Engine:** Implementation of SQl Queries like `SELECT`, `INSERT` and `DELETE` operations.
- [ ] **Concurrency Control:** Implementation of locking mechanisms for multi-threaded access.

***
