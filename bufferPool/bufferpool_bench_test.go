package bufferpool

import (
	"testing"
	"os"
	"fmt"
	"syscall"
)
/*
This file contains benchmarks for the BufferPool implementation. It benchmarks the performance of cold reads (where pages are not cached) and warm reads (where pages are cached) from the database file.
You can find the benchmark results in the README.md file
*/

func BenchmarkColdRead(b *testing.B) {
    // ensure pages aren't cached — e.g. fresh pool each iteration, or evict before each read
	// Use syscall.O_DIRECT to bypass the page cache
	fd, err := syscall.Open("../database.bin", syscall.O_RDWR|syscall.O_DIRECT, 0644)
	filep := os.NewFile(uintptr(fd), "../database.bin")
	// filep, err :=  os.OpenFile("../database.bin", os.O_RDWR|os.O_CREATE | os.O, 0644)
	// if err != nil {
	// 	fmt.Printf("Critical Error: Could not open database file: %v", err)
	// }
	fileInfo, err := filep.Stat()
	
	if err != nil {
		fmt.Printf("Failed to retrieve file stats: %v", err)
		return
	}
	numPages := fileInfo.Size() / 4096 
	fmt.Println(numPages)
	bp := InitializeBufferPool(filep)
    for i := 0; b.Loop(); i++ {
		/*
		1381 is the number of pages in the database after inserting 200,000 BigInt keys,
		we use i % 1381 to loop through the pages ensuring every page is read from disk because cachesize is 512 and we have 1381 pages,
		so every page will be evicted from the cache before it is read again
		*/
        bp.Get(uint32(i  % int(numPages))) 
    }
}
func BenchmarkWarmRead(b *testing.B) {
	// ensure pages are cached — e.g. read once before benchmark loop
	filep, err :=  os.OpenFile("../database.bin", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		fmt.Printf("Critical Error: Could not open database file: %v", err)
	}
	
	bp := InitializeBufferPool(filep)
	//warm up the cache by reading all pages once
	for i := range 512 {
		bp.Get(uint32(i))
	}
	b.ResetTimer()
	for i := 0; b.Loop(); i++ {
		bp.Get(uint32(i  % 512)) 
	}

}