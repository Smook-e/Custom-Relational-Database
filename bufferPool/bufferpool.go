package bufferpool

const bufferSize = 4096


type BufferPool struct {
	pages [512]Page
}

type Page struct {
	buffer [bufferSize]byte
}