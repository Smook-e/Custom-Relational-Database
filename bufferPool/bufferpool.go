package bufferpool

import (
	"github.com/Smook-e/Custom-Relational-Database/utilities/linkedlist"
	"github.com/Smook-e/Custom-Relational-Database/filehandler"
	"fmt"
	"os"
)

const bufferSize = 4096
const cacheSize = 512

type BufferPool struct {
	File *os.File
	pages [cacheSize]Page
	cache map[uint32]*linkedlist.Node
	list linkedlist.LinkedList // Head is most recently used
	freeIndex []int
	dirtyPages map[uint16]struct{}
}

type Page struct {
	buffer [bufferSize]byte
}

func InitializeBufferPool() *BufferPool {
	bp := & BufferPool{}
	bp.cache = make(map[uint32]*linkedlist.Node)
	bp.list = linkedlist.LinkedList{Tail: nil, Head: nil}
	bp.dirtyPages = make(map[uint16]struct{})
	bp.freeIndex = make([]int, cacheSize)
	for i := range len(bp.freeIndex) {
		bp.freeIndex[i] = len(bp.freeIndex) - 1 - i
	}
	return bp
}

func (bp *BufferPool) Get(pageId uint32) ([]byte, error) {
	//Check if page exists in cache
	
	if node, ok := bp.cache[pageId]; ok {//Cache Hit
		index, ok := node.Value.(int)
		if !ok {
			return nil, fmt.Errorf("Invalid node value type for pageId %d", pageId)
		}
		bp.list.MoveNodeToHead(node)
		return bp.pages[index].buffer[:], nil
	}
	
	//Cache Miss, read from file then add to cache
	

}