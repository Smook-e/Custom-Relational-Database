package bufferpool

import (
	"github.com/Smook-e/Custom-Relational-Database/utilities/linkedlist"
)

const bufferSize = 4096


type BufferPool struct {
	pages [512]Page
	cache map[uint32]*linkedlist.Node
	
}

type Page struct {
	buffer [bufferSize]byte
}