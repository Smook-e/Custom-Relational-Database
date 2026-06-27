package entities


type Table struct {

	rootIndex int
	columns []Column
}

type Column struct {
	name string
	dataType uint8
	contstraints uint8
	size uint8
}