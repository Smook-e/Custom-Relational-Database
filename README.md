# Custom Relational Database Engine (Go)

A high-performance, from-scratch relational database engine implemented in Go. This project implements a professional-grade **physical storage layer**, featuring a custom binary serialization format, a slotted-page architecture, and a sophisticated memory-buffering system to minimize disk I/O.

## Core Engine Features

### 1. Slotted Page Architecture
Implemented a **Slotted Page** layout for both table metadata and data rows, using 4KB pages to align with OS and RAM page sizes.
- **Forward-Growing Header:** A slot array at the page start storing exact byte offsets of records for $O(1)$ access.
- **Backward-Growing Payload:** Data is appended from the page end toward the header, maximizing space utilization.
- **Symmetric I/O:** A mirrored serialization/deserialization pipeline that ensures bit-perfect data integrity.

### 2. Advanced Metadata & Page Management
The engine handles dynamic schemas and optimizes data placement via a specialized management system.
- **Schema Serialization:** Uses length-prefixing and linked page chaining (`nextPage` pointers) to support an unlimited number of tables.
- **Free Space Manager:** Maintains a global map of `PageID` and `RemainingBytes` to instantly locate target pages without scanning the disk.
- **Sized Allocation:** Implements a "Dry Run" pass to validate row size and type-constraints before committing bytes to a page.

### 3. LRU Buffer Pool Manager
To eliminate the bottleneck of direct disk I/O, the engine implements a custom **Buffer Pool Manager**.
- **LRU Eviction:** Uses a Doubly Linked List and a Hash Map to maintain a cache of "hot" pages, evicting the Least Recently Used pages when the pool is full.
- **Dirty Page Tracking:** Tracks modified pages in memory and only writes them to disk upon eviction or explicit commit, drastically reducing syscall overhead.
- **Zero-Allocation Buffering:** Integrated with `sync.Pool` to reuse 4KB slices, minimizing Garbage Collector (GC) pressure.

### 4. Strict Binary Type System
Ensures data integrity through explicit bit-width conversions and validation.
- **Supported Types:** `TinyInt` (8-bit), `SmallInt` (16-bit), `Int` (32-bit), `BigInt` (64-bit), and `VarChar`.
- **Type Guardrails:** Validates user input against schema bit-limits during the conversion phase to prevent silent overflows.
- **Endianness:** Uses `BigEndian` encoding for cross-platform binary compatibility.

## 📂 Page Layout (4KB)
**Header (Forward) $\rightarrow$ $\leftarrow$ Payload (Backward)**
`[ FreeSpaceOffset (2B) | NumElements (2B) | Slot Array (N * 2B) ... ]` $\rightarrow$ $\leftarrow$ `[ Row Data ]`

## 📈 Roadmap
- [x] **Slotted Page Layout**
- [x] **Free Space Management**
- [x] **Binary Type Serialization**
- [x] **LRU Buffer Pool & Dirty Page Tracking**
- [x] **Atomic Commit Logic**
- [ ] **B+ Tree Indexing:** High-speed primary and secondary key lookups.
- [ ] **Query Engine:** Implementation of `SELECT`, `INSERT`, and `DELETE` operations.
- [ ] **Concurrency Control:** Thread-safe access and locking mechanisms.

***

### Key Changes made:
1.  **Buffer Pool:** Moved from "Future Roadmap" to "Core Features," highlighting the LRU and Dirty Page logic.
2.  **Commit Logic:** Added the concept of "Dirty Page Tracking" and "Symmetric I/O."
3.  **Page Layout:** Simplified the diagram into a concise "Forward/Backward" representation.
4.  **Technical Stack:** Updated to include the Buffer Pool and memory management strategies.
