package bufferpool

import (
	"testing"
)


func BenchmarkColdRead(b *testing.B) {
    // ensure pages aren't cached — e.g. fresh pool each iteration, or evict before each read
	bp := InitializeBufferPool()
    for i := 0; i < b.N; i++ {
        bp.Get(uint32(i  % cacheSize)) // Access pages in a round-robin fashion
    }
}