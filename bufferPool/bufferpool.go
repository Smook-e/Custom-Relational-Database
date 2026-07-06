package bufferpool

const bufferSize = 4096


type BufferPool struct {

}

type Page struct {
	buffer [bufferSize]byte
}