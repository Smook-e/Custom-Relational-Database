package bufferpool

import (
	"github.com/Smook-e/Custom-Relational-Database/utilities/linkedlist"
	"github.com/Smook-e/Custom-Relational-Database/filehandler"
	"fmt"
)

const bufferSize = 4096


type BufferPool struct {
	pages [512]Page
	cache map[uint32]*linkedlist.Node
	list linkedlist.LinkedList // Head is most recently used
}

type Page struct {
	buffer [bufferSize]byte
}

func (bp *BufferPool) Get(pageId uint32) ([]byte, error) {
	//Check if page exists in cache
	node, ok := bp.cache[pageId]

	if ok {//Cache Hit
		index, ok := node.Value.(int)
		if !ok {
			return nil, fmt.Errorf("Invalid node value type for pageId %d", pageId)
		}
		bp.list.MoveNodeToHead(node)
		return bp.pages[index].buffer[:], nil
	}
	
	//Cache Miss, read from file then add to cache

	LRUNode := bp.list.RemoveTail()
	index := LRUNode.Value.(int)
	LRUNode.
}