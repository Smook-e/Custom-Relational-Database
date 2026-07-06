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
	File *os.File //original file pointer from database
	pages [cacheSize]Page //Actual buffer pool, each page is 4KB
	cache map[uint32]*linkedlist.Node // maps pageId to the corresponding node in the linked list
	list linkedlist.LinkedList // Head is most recently used
	freeIndex []int // a stack of free indices in the pages array, used to track which pages are available for use
	dirtyPages map[uint32]struct{} // a set of pageIds that have been modified and need to be written back to disk
}

type Page struct {
	buffer [bufferSize]byte
}

func InitializeBufferPool() *BufferPool {
	bp := & BufferPool{}
	bp.cache = make(map[uint32]*linkedlist.Node)
	bp.list = linkedlist.LinkedList{Tail: nil, Head: nil}
	bp.dirtyPages = make(map[uint32]struct{})
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
	var freeindex int;

	if len(bp.freeIndex) > 0{ // Array still has free space
		freeindex, bp.freeIndex = bp.freeIndex[len(bp.freeIndex) - 1], bp.freeIndex[:len(bp.freeIndex) - 1]
	}else  {
		LRUNode := bp.list.RemoveTail()// remove least recently used node
		freeindex = LRUNode.Value.(int)// assign its place for our new node

		if _, isDirty := bp.dirtyPages[LRUNode.PageID]; isDirty { // if the evicted page is dirty, write it back to disk
			// Flush the page to disk before overwriting the buffer
			err := filehandler.WriteToFile(bp.File, LRUNode.PageID, bp.pages[freeindex].buffer[:])
			if err != nil {
				return nil, fmt.Errorf("failed to flush evicted page: %w", err)
			}
			delete(bp.dirtyPages, LRUNode.PageID)
		}

		delete(bp.cache, LRUNode.PageID)// remove its entry from the cache
	}
	err := filehandler.ReadFromFile(bp.File, pageId, bp.pages[freeindex].buffer[:])// read the page from file
	if err != nil {
		return nil, err
	}
	bp.list.AddNodetoHead(freeindex)// add the new node to the head of the list
	bp.cache[pageId] = bp.list.Head// add the new node to the cache
	bp.list.Head.PageID = pageId// set the pageId of the new node
	return bp.pages[freeindex].buffer[:], nil



	

}

func (bp *BufferPool) MarkDirty(pageId uint32) error {
	if _, ok := bp.cache[pageId]; !ok {
		return fmt.Errorf("Page %d not found in cache", pageId)
	}
	bp.dirtyPages[pageId] = struct{}{}
	return nil
}

func (bp *BufferPool) Flush() error {//writes all dirty pages to disk
	for pageId := range bp.dirtyPages {
		node, ok := bp.cache[pageId]
		if !ok {
			return fmt.Errorf("Page %d not found in cache", pageId)
		}
		index, ok := node.Value.(int)
		if !ok {
			return fmt.Errorf("Invalid node value type for pageId %d", pageId)
		}
		err := filehandler.WriteToFile(bp.File, pageId, bp.pages[index].buffer[:])
		if err != nil {
			return err
		}
		
	}
	clear(bp.dirtyPages)
	return nil
		
}