package bufferpool

import (
	"testing"
)


func BenchmarkColdRead(b *testing.B) {
    // ensure pages aren't cached — e.g. fresh pool each iteration, or evict before each read
	bp := InitializeBufferPool()
    for i := 0; b.Loop(); i++ {
		/*
		1450 is the number of pages in the database after inserting 200,000 BigInt keys,
		we use i % 1450 to loop through the pages ensuring every page is read from disk because cachesize is 512 and we have 1450 pages,
		so every page will be evicted from the cache before it is read again
		*/
        bp.Get(uint32(i  % 1450)) 
    }
}

